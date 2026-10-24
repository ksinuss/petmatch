package service

import (
	"database/sql"
	"errors"
	"testing"

	"petmatch.local/backend/internal/model"
)

type fakePetRepo struct{ pets map[int]model.Pet }

func (f *fakePetRepo) CreatePet(p model.Pet) (int, error) {
	id := len(f.pets) + 1
	p.ID = id
	f.pets[id] = p
	return id, nil
}
func (f *fakePetRepo) GetPetByID(id int) (model.Pet, error) {
	p, ok := f.pets[id]
	if !ok {
		return model.Pet{}, sql.ErrNoRows
	}
	return p, nil
}
func (f *fakePetRepo) GetPetsByOwner(int) ([]model.Pet, error)    { return nil, nil }
func (f *fakePetRepo) GetFeed(int, int, int) ([]model.Pet, error) { return nil, nil }

func TestCreatePetValidation(t *testing.T) {
	s := NewPetService(&fakePetRepo{pets: map[int]model.Pet{}})
	bad := map[string]model.Pet{
		"без клички":            {Type: "dog"},
		"неверный тип":          {Name: "Рич", Type: "dragon"},
		"отрицательный возраст": {Name: "Рич", Type: "dog", Age: -1},
		"слишком старый":        {Name: "Рич", Type: "dog", Age: 99},
		"неверный режим":        {Name: "Рич", Type: "dog", AdoptionMode: "weird"},
	}
	for name, p := range bad {
		if _, err := s.CreatePet(p); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%s: ожидали ErrInvalidInput, получили %v", name, err)
		}
	}
}

func TestCreatePetDefaultsToOpenMode(t *testing.T) {
	s := NewPetService(&fakePetRepo{pets: map[int]model.Pet{}})
	p, err := s.CreatePet(model.Pet{OwnerID: 1, Name: "  Рич ", Type: "dog"})
	if err != nil {
		t.Fatal(err)
	}
	if p.AdoptionMode != model.AdoptionOpen || p.Name != "Рич" {
		t.Fatalf("ожидали режим open и обрезанное имя, получили %+v", p)
	}
}
