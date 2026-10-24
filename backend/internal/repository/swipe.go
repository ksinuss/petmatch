package repository

import (
	"database/sql"

	"petmatch.local/backend/internal/model"
)

// SwipeRepository — свайпы и мэтчи.
type SwipeRepository struct {
	db *sql.DB
}

func NewSwipeRepository(db *sql.DB) *SwipeRepository {
	return &SwipeRepository{db: db}
}

func (r *SwipeRepository) CreateSwipe(userID, petID int, isLike bool) error {
	_, err := r.db.Exec(`INSERT INTO swipes (user_id, pet_id, is_like) VALUES (?, ?, ?)`, userID, petID, isLike)
	return err
}

func (r *SwipeRepository) CreateOwnerSwipe(ownerID, userID, petID int, isLike bool) error {
	_, err := r.db.Exec(
		`INSERT INTO owner_swipes (owner_id, user_id, pet_id, is_like) VALUES (?, ?, ?, ?)`,
		ownerID, userID, petID, isLike)
	return err
}

func (r *SwipeRepository) exists(query string, args ...any) (bool, error) {
	var one int
	err := r.db.QueryRow(query, args...).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *SwipeRepository) HasOwnerLike(userID, petID int) (bool, error) {
	return r.exists(`SELECT 1 FROM owner_swipes WHERE user_id = ? AND pet_id = ? AND is_like = 1 LIMIT 1`, userID, petID)
}

func (r *SwipeRepository) HasUserLike(userID, petID int) (bool, error) {
	return r.exists(`SELECT 1 FROM swipes WHERE user_id = ? AND pet_id = ? AND is_like = 1 LIMIT 1`, userID, petID)
}

// CreateMatch идемпотентно создаёт мэтч и возвращает его id.
func (r *SwipeRepository) CreateMatch(userID, petID int) (int, error) {
	if _, err := r.db.Exec(`INSERT OR IGNORE INTO matches (user_id, pet_id) VALUES (?, ?)`, userID, petID); err != nil {
		return 0, err
	}
	var id int
	err := r.db.QueryRow(`SELECT id FROM matches WHERE user_id = ? AND pet_id = ?`, userID, petID).Scan(&id)
	return id, err
}

// GetPetInfo возвращает владельца и режим усыновления питомца.
func (r *SwipeRepository) GetPetInfo(petID int) (ownerID int, mode string, err error) {
	err = r.db.QueryRow(`SELECT owner_id, COALESCE(adoption_mode, 'open') FROM pets WHERE id = ?`, petID).Scan(&ownerID, &mode)
	return
}

// GetMatches возвращает диалоги, в которых участвует пользователь —
// и как тот, кто лайкнул питомца, и как владелец питомца.
func (r *SwipeRepository) GetMatches(userID int) ([]model.MatchView, error) {
	rows, err := r.db.Query(`
		SELECT m.id, m.created_at, m.user_id, p.owner_id,
		       p.id, p.name, p.type, COALESCE(p.breed, ''),
		       COALESCE((SELECT url FROM pet_photos WHERE pet_id = p.id ORDER BY id LIMIT 1), '')
		FROM matches m
		JOIN pets p ON p.id = m.pet_id
		WHERE m.user_id = ? OR p.owner_id = ?
		ORDER BY m.id DESC`,
		userID, userID)
	if err != nil {
		return nil, err
	}

	type partnerRef struct{ idx, id int }
	var refs []partnerRef
	result := []model.MatchView{}
	for rows.Next() {
		var mv model.MatchView
		var matchUserID, ownerID int
		var createdAt string
		if err := rows.Scan(&mv.ID, &createdAt, &matchUserID, &ownerID,
			&mv.Pet.ID, &mv.Pet.Name, &mv.Pet.Type, &mv.Pet.Breed, &mv.Pet.Photo); err != nil {
			rows.Close()
			return nil, err
		}
		mv.CreatedAt = toRFC3339(createdAt)
		partner := ownerID
		if userID == ownerID {
			partner = matchUserID
		}
		refs = append(refs, partnerRef{idx: len(result), id: partner})
		result = append(result, mv)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	// Соединение в пуле одно, поэтому доп. запросы — только после закрытия rows.
	for _, ref := range refs {
		mv := &result[ref.idx]
		mv.Partner.ID = ref.id
		if err := r.db.QueryRow(`SELECT display_name, role FROM accounts WHERE id = ?`, ref.id).
			Scan(&mv.Partner.DisplayName, &mv.Partner.Role); err != nil && err != sql.ErrNoRows {
			return nil, err
		}

	}
	return result, nil
}
