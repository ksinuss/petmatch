package service

import (
	"strings"

	"petmatch.local/backend/internal/model"
)

type PetRepository interface {
	CreatePet(p model.Pet) (int, error)
	GetPetByID(id int) (model.Pet, error)
	GetPetsByOwner(ownerID int) ([]model.Pet, error)
	GetFeed(userID, limit, offset int) ([]model.Pet, error)
}

type PetService struct {
	repo PetRepository
}

func NewPetService(r PetRepository) *PetService {
	return &PetService{repo: r}
}

var petTypes = map[string]bool{"dog": true, "cat": true, "other": true}

func (s *PetService) CreatePet(p model.Pet) (model.Pet, error) {
	p.Name = strings.TrimSpace(p.Name)
	p.Breed = strings.TrimSpace(p.Breed)
	p.Description = strings.TrimSpace(p.Description)
	if p.AdoptionMode == "" {
		p.AdoptionMode = model.AdoptionOpen
	}

	switch {
	case p.Name == "":
		return model.Pet{}, invalid("Укажите кличку питомца")
	case runeLen(p.Name) > 50:
		return model.Pet{}, invalid("Кличка слишком длинная (максимум 50 символов)")
	case !petTypes[p.Type]:
		return model.Pet{}, invalid("Тип животного: dog, cat или other")
	case p.Age < 0 || p.Age > 40:
		return model.Pet{}, invalid("Возраст должен быть от 0 до 40 лет")
	case runeLen(p.Breed) > 60:
		return model.Pet{}, invalid("Порода слишком длинная (максимум 60 символов)")
	case runeLen(p.Description) > 1000:
		return model.Pet{}, invalid("Описание слишком длинное (максимум 1000 символов)")
	case p.AdoptionMode != model.AdoptionOpen && p.AdoptionMode != model.AdoptionStrict:
		return model.Pet{}, invalid("adoption_mode: open или strict")
	}

	id, err := s.repo.CreatePet(p)
	if err != nil {
		return model.Pet{}, err
	}
	return s.GetPet(id)
}

func (s *PetService) GetPet(id int) (model.Pet, error) {
	p, err := s.repo.GetPetByID(id)
	if isNoRows(err) {
		return p, notFound("Питомец не найден")
	}
	return p, err
}

func (s *PetService) GetPetsByOwner(ownerID int) ([]model.Pet, error) {
	return s.repo.GetPetsByOwner(ownerID)
}

const feedPageSize = 10

func (s *PetService) GetFeed(userID int) ([]model.Pet, error) {
	return s.repo.GetFeed(userID, feedPageSize, 0)
}
