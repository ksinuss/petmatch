package repository

import (
	"database/sql"

	"petmatch.local/backend/internal/model"
)

// PreferencesRepository — предпочтения пользователя для ленты.
type PreferencesRepository struct {
	db *sql.DB
}

func NewPreferencesRepository(db *sql.DB) *PreferencesRepository {
	return &PreferencesRepository{db: db}
}

func (r *PreferencesRepository) SavePreferences(userID int, p model.Preferences) error {
	_, err := r.db.Exec(`
		INSERT INTO user_preferences (user_id, preferred_type, min_age, max_age, preferred_breed, preferred_location)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			preferred_type     = excluded.preferred_type,
			min_age            = excluded.min_age,
			max_age            = excluded.max_age,
			preferred_breed    = excluded.preferred_breed,
			preferred_location = excluded.preferred_location`,
		userID, p.PreferredType, p.MinAge, p.MaxAge, p.PreferredBreed, p.PreferredLocation)
	return err
}

// GetPreferences возвращает sql.ErrNoRows, если пользователь ничего не сохранял.
func (r *PreferencesRepository) GetPreferences(userID int) (model.Preferences, error) {
	var p model.Preferences
	var t, b, l sql.NullString
	var minAge, maxAge sql.NullInt64
	err := r.db.QueryRow(`
		SELECT preferred_type, min_age, max_age, preferred_breed, preferred_location
		FROM user_preferences WHERE user_id = ?`, userID).Scan(&t, &minAge, &maxAge, &b, &l)
	if err != nil {
		return p, err
	}
	p.PreferredType, p.PreferredBreed, p.PreferredLocation = t.String, b.String, l.String
	p.MinAge, p.MaxAge = int(minAge.Int64), int(maxAge.Int64)
	return p, nil
}
