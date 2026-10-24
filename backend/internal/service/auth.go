package service

import (
	"strings"

	"petmatch.local/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	CreateAccount(email, passwordHash, role, displayName string) (int, error)
	GetAccountByEmail(email string) (model.Account, error)
}

type AuthService struct {
	repo   AuthRepository
	tokens *TokenManager
}

func NewAuthService(repo AuthRepository, tokens *TokenManager) *AuthService {
	return &AuthService{repo: repo, tokens: tokens}
}

// AuthResult — аккаунт и выданный ему токен.
type AuthResult struct {
	Token   string
	Account model.Account
}

const minPasswordLen = 6

func (s *AuthService) Register(email, password, role, displayName string) (AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	displayName = strings.TrimSpace(displayName)

	switch {
	case !validEmail(email):
		return AuthResult{}, invalid("Введите корректный email")
	case len(password) < minPasswordLen:
		return AuthResult{}, invalid("Пароль должен быть не короче 6 символов")
	case len(password) > 72: // ограничение bcrypt
		return AuthResult{}, invalid("Пароль слишком длинный (максимум 72 символа)")
	case role != "user" && role != "shelter":
		return AuthResult{}, invalid("Роль должна быть user или shelter")
	case displayName == "":
		return AuthResult{}, invalid("Укажите имя")
	case runeLen(displayName) > 60:
		return AuthResult{}, invalid("Имя слишком длинное (максимум 60 символов)")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, err
	}

	id, err := s.repo.CreateAccount(email, string(hash), role, displayName)
	if err != nil {
		if isUnique(err) {
			return AuthResult{}, conflict("Пользователь с таким email уже зарегистрирован")
		}
		return AuthResult{}, err
	}

	acc := model.Account{ID: id, Email: email, Role: role, DisplayName: displayName}
	return AuthResult{Token: s.tokens.Issue(id), Account: acc}, nil
}

func (s *AuthService) Login(email, password string) (AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return AuthResult{}, invalid("Введите email и пароль")
	}

	acc, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		if isNoRows(err) {
			return AuthResult{}, unauthorized("Неверный email или пароль")
		}
		return AuthResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(password)) != nil {
		return AuthResult{}, unauthorized("Неверный email или пароль")
	}
	return AuthResult{Token: s.tokens.Issue(acc.ID), Account: acc}, nil
}
