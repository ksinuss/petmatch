package repository

import (
	"database/sql"

	"petmatch.local/backend/internal/model"
)

// AccountRepository — доступ к таблицам accounts (профиль и регистрация).
type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

const accountColumns = `id, email, password_hash, role, display_name, COALESCE(bio, ''), COALESCE(location, '')`

func scanAccount(row *sql.Row) (model.Account, error) {
	var a model.Account
	err := row.Scan(&a.ID, &a.Email, &a.PasswordHash, &a.Role, &a.DisplayName, &a.Bio, &a.Location)
	return a, err
}

func (r *AccountRepository) CreateAccount(email, passwordHash, role, displayName string) (int, error) {
	res, err := r.db.Exec(
		`INSERT INTO accounts (email, password_hash, role, display_name) VALUES (?, ?, ?, ?)`,
		email, passwordHash, role, displayName,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (r *AccountRepository) GetAccountByEmail(email string) (model.Account, error) {
	return scanAccount(r.db.QueryRow(`SELECT `+accountColumns+` FROM accounts WHERE email = ?`, email))
}

func (r *AccountRepository) GetByID(id int) (model.Account, error) {
	return scanAccount(r.db.QueryRow(`SELECT `+accountColumns+` FROM accounts WHERE id = ?`, id))
}

func (r *AccountRepository) Update(id int, displayName, bio, location string) error {
	_, err := r.db.Exec(
		`UPDATE accounts SET display_name = ?, bio = ?, location = ? WHERE id = ?`,
		displayName, bio, location, id,
	)
	return err
}

func (r *AccountRepository) Count() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM accounts`).Scan(&n)
	return n, err
}
