package service

import (
	"errors"
	"testing"
	"time"
)

func TestTokenIssueAndParse(t *testing.T) {
	tm := NewTokenManager("secret", time.Hour)
	id, err := tm.Parse(tm.Issue(42))
	if err != nil || id != 42 {
		t.Fatalf("ожидали id=42, получили %d, err=%v", id, err)
	}
}

func TestTokenRejectsTampering(t *testing.T) {
	tm := NewTokenManager("secret", time.Hour)
	other := NewTokenManager("another-secret", time.Hour)

	cases := map[string]string{
		"чужой секрет": other.Issue(1),
		"мусор":        "not-a-token",
		"пустой":       "",
		"без подписи":  "MS4xOTk5OTk5OTk5.",
	}
	for name, tok := range cases {
		if _, err := tm.Parse(tok); !errors.Is(err, ErrUnauthorized) {
			t.Errorf("%s: ожидали ErrUnauthorized, получили %v", name, err)
		}
	}
}

func TestTokenExpires(t *testing.T) {
	tm := NewTokenManager("secret", time.Minute)
	now := time.Now()
	tm.now = func() time.Time { return now }
	tok := tm.Issue(7)

	tm.now = func() time.Time { return now.Add(2 * time.Minute) }
	if _, err := tm.Parse(tok); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("просроченный токен должен отклоняться, получили %v", err)
	}
}
