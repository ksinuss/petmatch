package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"petmatch.local/backend/internal/model"
)

type fakeAuthRepo struct {
	accounts map[string]model.Account
	nextID   int
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{accounts: map[string]model.Account{}, nextID: 1}
}

func (f *fakeAuthRepo) CreateAccount(email, hash, role, name string) (int, error) {
	if _, ok := f.accounts[email]; ok {
		return 0, errors.New("UNIQUE constraint failed: accounts.email")
	}
	id := f.nextID
	f.nextID++
	f.accounts[email] = model.Account{ID: id, Email: email, PasswordHash: hash, Role: role, DisplayName: name}
	return id, nil
}

func (f *fakeAuthRepo) GetAccountByEmail(email string) (model.Account, error) {
	a, ok := f.accounts[email]
	if !ok {
		return model.Account{}, sql.ErrNoRows
	}
	return a, nil
}

func newAuth() *AuthService {
	return NewAuthService(newFakeAuthRepo(), NewTokenManager("s", time.Hour))
}

func TestRegisterValidation(t *testing.T) {
	cases := []struct {
		name                            string
		email, password, role, dispName string
	}{
		{"плохой email", "not-an-email", "secret1", "user", "Анна"},
		{"email без домена", "a@b", "secret1", "user", "Анна"},
		{"короткий пароль", "a@b.ru", "123", "user", "Анна"},
		{"неизвестная роль", "a@b.ru", "secret1", "admin", "Анна"},
		{"пустое имя", "a@b.ru", "secret1", "user", "  "},
	}
	for _, c := range cases {
		_, err := newAuth().Register(c.email, c.password, c.role, c.dispName)
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%s: ожидали ErrInvalidInput, получили %v", c.name, err)
		}
	}
}

func TestRegisterThenLogin(t *testing.T) {
	s := newAuth()
	reg, err := s.Register("  Anna@Example.COM ", "secret1", "user", "Анна")
	if err != nil {
		t.Fatal(err)
	}
	if reg.Token == "" || reg.Account.Email != "anna@example.com" {
		t.Fatalf("email должен нормализоваться, токен выдан: %+v", reg)
	}
	if reg.Account.PasswordHash != "" {
		t.Error("хеш пароля не должен возвращаться из Register")
	}

	login, err := s.Login("anna@example.com", "secret1")
	if err != nil || login.Account.ID != reg.Account.ID {
		t.Fatalf("вход должен работать: %v", err)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	s := newAuth()
	if _, err := s.Register("a@b.ru", "secret1", "user", "A"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Register("a@b.ru", "secret2", "shelter", "B"); !errors.Is(err, ErrConflict) {
		t.Fatalf("ожидали ErrConflict, получили %v", err)
	}
}

func TestLoginWrongCredentials(t *testing.T) {
	s := newAuth()
	_, _ = s.Register("a@b.ru", "secret1", "user", "A")

	for name, creds := range map[string][2]string{
		"неверный пароль":  {"a@b.ru", "wrong-pass"},
		"нет такого email": {"zzz@b.ru", "secret1"},
	} {
		if _, err := s.Login(creds[0], creds[1]); !errors.Is(err, ErrUnauthorized) {
			t.Errorf("%s: ожидали ErrUnauthorized, получили %v", name, err)
		}
	}
	if _, err := s.Login("", ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("пустые поля: ожидали ErrInvalidInput, получили %v", err)
	}
}
