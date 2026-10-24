package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TokenManager выпускает и проверяет подписанные токены вида
// base64url("<userID>.<expiresUnix>") + "." + base64url(HMAC-SHA256).
// Подмена заголовка (как было с X-User-ID) больше не работает:
// без секрета AUTH_SECRET валидную подпись не получить.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (t *TokenManager) Issue(userID int) string {
	payload := fmt.Sprintf("%d.%d", userID, t.now().Add(t.ttl).Unix())
	enc := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return enc + "." + t.sign(enc)
}

func (t *TokenManager) Parse(token string) (int, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, unauthorized("invalid token")
	}
	if !hmac.Equal([]byte(parts[1]), []byte(t.sign(parts[0]))) {
		return 0, unauthorized("invalid token signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, unauthorized("invalid token")
	}
	fields := strings.Split(string(raw), ".")
	if len(fields) != 2 {
		return 0, unauthorized("invalid token")
	}
	id, err1 := strconv.Atoi(fields[0])
	exp, err2 := strconv.ParseInt(fields[1], 10, 64)
	if err1 != nil || err2 != nil || id <= 0 {
		return 0, unauthorized("invalid token")
	}
	if t.now().Unix() > exp {
		return 0, unauthorized("token expired")
	}
	return id, nil
}

func (t *TokenManager) sign(data string) string {
	m := hmac.New(sha256.New, t.secret)
	m.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
