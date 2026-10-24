package database

import (
	"database/sql"
	"fmt"
	"log"
)

// Migrate создаёт таблицы, если их ещё нет. Схема описана в docs/ARCHITECTURE.md.
func Migrate(db *sql.DB) error {
	queries := []string{
		// Аккаунты. role: user | shelter
		`CREATE TABLE IF NOT EXISTS accounts (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			email         TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role          TEXT NOT NULL,
			display_name  TEXT NOT NULL,
			bio           TEXT,
			location      TEXT
		);`,

		// Питомцы. adoption_mode: open | strict
		`CREATE TABLE IF NOT EXISTS pets (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_id      INTEGER NOT NULL,
			name          TEXT NOT NULL,
			type          TEXT NOT NULL,
			breed         TEXT,
			age           INTEGER,
			description   TEXT,
			adoption_mode TEXT DEFAULT 'open',
			status        TEXT DEFAULT 'available',
			FOREIGN KEY (owner_id) REFERENCES accounts(id)
		);`,

		`CREATE TABLE IF NOT EXISTS pet_photos (
			id     INTEGER PRIMARY KEY AUTOINCREMENT,
			pet_id INTEGER NOT NULL,
			url    TEXT NOT NULL,
			FOREIGN KEY (pet_id) REFERENCES pets(id)
		);`,

		`CREATE TABLE IF NOT EXISTS account_photos (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			url        TEXT NOT NULL,
			is_main    INTEGER DEFAULT 0,
			FOREIGN KEY (account_id) REFERENCES accounts(id)
		);`,

		// Свайпы пользователя по питомцам.
		`CREATE TABLE IF NOT EXISTS swipes (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id    INTEGER NOT NULL,
			pet_id     INTEGER NOT NULL,
			is_like    INTEGER NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES accounts(id),
			FOREIGN KEY (pet_id) REFERENCES pets(id),
			UNIQUE(user_id, pet_id)
		);`,

		// Свайпы владельца по пользователям (для режима strict).
		`CREATE TABLE IF NOT EXISTS owner_swipes (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_id   INTEGER NOT NULL,
			user_id    INTEGER NOT NULL,
			pet_id     INTEGER NOT NULL,
			is_like    INTEGER NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_id) REFERENCES accounts(id),
			FOREIGN KEY (user_id)  REFERENCES accounts(id),
			FOREIGN KEY (pet_id)   REFERENCES pets(id),
			UNIQUE(owner_id, user_id, pet_id)
		);`,

		// Мэтчи: пара "пользователь — питомец". Уникальна, чтобы не плодить дубли.
		`CREATE TABLE IF NOT EXISTS matches (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id    INTEGER NOT NULL,
			pet_id     INTEGER NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES accounts(id),
			FOREIGN KEY (pet_id)  REFERENCES pets(id)
		);`,

		`CREATE TABLE IF NOT EXISTS user_preferences (
			id                 INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id            INTEGER NOT NULL UNIQUE,
			preferred_type     TEXT,
			min_age            INTEGER,
			max_age            INTEGER,
			preferred_breed    TEXT,
			preferred_location TEXT,
			FOREIGN KEY (user_id) REFERENCES accounts(id)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_pets_owner     ON pets(owner_id);`,
		`CREATE INDEX IF NOT EXISTS idx_pet_photos_pet ON pet_photos(pet_id);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
	}

	// Уникальный индекс может не создаться на старой БД с дублями — это не фатально.
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_matches_user_pet ON matches(user_id, pet_id);`); err != nil {
		log.Println("warning: unique index on matches not created:", err)
	}

	return nil
}
