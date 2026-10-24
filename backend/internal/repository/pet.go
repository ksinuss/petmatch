package repository

import (
	"database/sql"
	"strings"

	"petmatch.local/backend/internal/model"
)

// PetRepository — питомцы, их фото и лента для свайпов.
type PetRepository struct {
	db *sql.DB
}

func NewPetRepository(db *sql.DB) *PetRepository {
	return &PetRepository{db: db}
}

const petColumns = `p.id, p.owner_id, COALESCE(a.display_name, ''), p.name, p.type,
	COALESCE(p.breed, ''), COALESCE(p.age, 0), COALESCE(p.description, ''),
	COALESCE(p.adoption_mode, 'open'), COALESCE(p.status, 'available')`

const petFrom = ` FROM pets p JOIN accounts a ON a.id = p.owner_id `

type scanner interface {
	Scan(dest ...any) error
}

func scanPet(s scanner) (model.Pet, error) {
	var p model.Pet
	err := s.Scan(&p.ID, &p.OwnerID, &p.OwnerName, &p.Name, &p.Type, &p.Breed,
		&p.Age, &p.Description, &p.AdoptionMode, &p.Status)
	p.Photos = []string{}
	return p, err
}

func (r *PetRepository) CreatePet(p model.Pet) (int, error) {
	res, err := r.db.Exec(
		`INSERT INTO pets (owner_id, name, type, breed, age, description, adoption_mode)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.OwnerID, p.Name, p.Type, p.Breed, p.Age, p.Description, p.AdoptionMode,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (r *PetRepository) GetPetByID(id int) (model.Pet, error) {
	p, err := scanPet(r.db.QueryRow(`SELECT `+petColumns+petFrom+`WHERE p.id = ?`, id))
	if err != nil {
		return model.Pet{}, err
	}
	pets := []model.Pet{p}
	if err := r.attachPhotos(pets); err != nil {
		return model.Pet{}, err
	}
	return pets[0], nil
}

func (r *PetRepository) GetPetsByOwner(ownerID int) ([]model.Pet, error) {
	rows, err := r.db.Query(`SELECT `+petColumns+petFrom+`WHERE p.owner_id = ? ORDER BY p.id DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	return r.collect(rows)
}

// GetFeed возвращает питомцев, которых пользователь ещё не оценивал и которые
// принадлежат не ему. Подходящие под предпочтения идут выше.
func (r *PetRepository) GetFeed(userID, limit, offset int) ([]model.Pet, error) {
	rows, err := r.db.Query(`
		SELECT `+petColumns+petFrom+`
		LEFT JOIN user_preferences up ON up.user_id = ?
		WHERE p.owner_id != ?
		  AND p.status = 'available'
		  AND p.id NOT IN (SELECT pet_id FROM swipes WHERE user_id = ?)
		ORDER BY
			(CASE WHEN up.preferred_type IS NOT NULL AND up.preferred_type != '' AND p.type = up.preferred_type THEN 3 ELSE 0 END +
			 CASE WHEN up.preferred_breed IS NOT NULL AND up.preferred_breed != '' AND p.breed = up.preferred_breed THEN 2 ELSE 0 END +
			 CASE WHEN up.max_age IS NOT NULL AND up.max_age > 0 AND p.age BETWEEN up.min_age AND up.max_age THEN 2 ELSE 0 END) DESC,
			p.id ASC
		LIMIT ? OFFSET ?`,
		userID, userID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return r.collect(rows)
}

func (r *PetRepository) collect(rows *sql.Rows) ([]model.Pet, error) {
	defer rows.Close()
	pets := []model.Pet{}
	for rows.Next() {
		p, err := scanPet(rows)
		if err != nil {
			return nil, err
		}
		pets = append(pets, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if err := r.attachPhotos(pets); err != nil {
		return nil, err
	}
	return pets, nil
}

// attachPhotos одним запросом подгружает фото для списка питомцев.
func (r *PetRepository) attachPhotos(pets []model.Pet) error {
	if len(pets) == 0 {
		return nil
	}
	args := make([]any, len(pets))
	idx := make(map[int]int, len(pets))
	for i, p := range pets {
		args[i] = p.ID
		idx[p.ID] = i
	}
	q := `SELECT pet_id, url FROM pet_photos WHERE pet_id IN (?` +
		strings.Repeat(",?", len(pets)-1) + `) ORDER BY id`
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var petID int
		var url string
		if err := rows.Scan(&petID, &url); err != nil {
			return err
		}
		i := idx[petID]
		pets[i].Photos = append(pets[i].Photos, url)
	}
	return rows.Err()
}

func (r *PetRepository) AddPetPhoto(petID int, url string) error {
	_, err := r.db.Exec(`INSERT INTO pet_photos (pet_id, url) VALUES (?, ?)`, petID, url)
	return err
}
