package service

import (
	"database/sql"
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

// isUnique — нарушение UNIQUE-ограничения SQLite (одинаково у любых драйверов).
func isUnique(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func isNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }

func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s && strings.Contains(s[strings.Index(s, "@"):], ".")
}

func runeLen(s string) int { return utf8.RuneCountInString(s) }
