// Package database открывает SQLite и создаёт схему.
package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // чистый Go-драйвер, CGO не нужен
)

// InitDB открывает базу. SQLite — файловая БД с одним писателем, поэтому
// ограничиваем пул одним соединением: так PRAGMA foreign_keys действует
// на все запросы, а параллельные записи не ловят "database is locked".
func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}
