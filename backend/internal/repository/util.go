package repository

import "time"

// toRFC3339 переводит CURRENT_TIMESTAMP из SQLite ("2006-01-02 15:04:05", UTC)
// в RFC3339, который одинаково понимают все браузеры.
func toRFC3339(s string) string {
	if s == "" {
		return ""
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t.UTC().Format(time.RFC3339)
	}
	return s
}
