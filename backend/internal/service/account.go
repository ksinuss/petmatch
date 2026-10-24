package service

import (
	"strings"

	"petmatch.local/backend/internal/model"
)

type AccountRepository interface {
	GetByID(id int) (model.Account, error)
	Update(id int, displayName, bio, location string) error
}

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(r AccountRepository) *AccountService {
	return &AccountService{repo: r}
}

func (s *AccountService) GetMe(userID int) (model.Account, error) {
	acc, err := s.repo.GetByID(userID)
	if isNoRows(err) {
		return acc, notFound("Пользователь не найден")
	}
	return acc, err
}

// PublicAccount — то, что видят другие пользователи (без email).
type PublicAccount struct {
	ID          int    `json:"id"`
	Role        string `json:"role"`
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	Location    string `json:"location"`
}

func (s *AccountService) GetPublic(id int) (PublicAccount, error) {
	acc, err := s.repo.GetByID(id)
	if err != nil {
		if isNoRows(err) {
			return PublicAccount{}, notFound("Пользователь не найден")
		}
		return PublicAccount{}, err
	}
	return PublicAccount{ID: acc.ID, Role: acc.Role, DisplayName: acc.DisplayName, Bio: acc.Bio, Location: acc.Location}, nil
}

func (s *AccountService) UpdateMe(userID int, displayName, bio, location string) (model.Account, error) {
	displayName, bio, location = strings.TrimSpace(displayName), strings.TrimSpace(bio), strings.TrimSpace(location)
	switch {
	case displayName == "":
		return model.Account{}, invalid("Имя не может быть пустым")
	case runeLen(displayName) > 60:
		return model.Account{}, invalid("Имя слишком длинное (максимум 60 символов)")
	case runeLen(bio) > 500:
		return model.Account{}, invalid("Описание слишком длинное (максимум 500 символов)")
	case runeLen(location) > 80:
		return model.Account{}, invalid("Город слишком длинный (максимум 80 символов)")
	}
	if err := s.repo.Update(userID, displayName, bio, location); err != nil {
		return model.Account{}, err
	}
	return s.GetMe(userID)
}
