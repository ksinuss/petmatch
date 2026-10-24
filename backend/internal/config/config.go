// Package config читает настройки приложения из переменных окружения.
// Ничего секретного в коде не хранится: значения по умолчанию подходят
// только для локальной разработки.
package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// DefaultAuthSecret — значение для локальной разработки; в окружении его нужно заменить.
const DefaultAuthSecret = "dev-insecure-secret-change-me"

type Config struct {
	Port        string   // PORT
	DBPath      string   // DB_PATH
	CORSOrigins []string // CORS_ORIGINS (через запятую)
	AuthSecret  string   // AUTH_SECRET — ключ подписи токенов
	TokenTTLHrs int      // TOKEN_TTL_HOURS
	UploadDir   string   // UPLOAD_DIR — куда складывать фото, если S3 не настроен
	SeedDemo    bool     // SEED_DEMO — наполнить пустую БД демо-данными

	// Необязательное S3-хранилище (Yandex Object Storage).
	S3AccessKey string // YANDEX_ACCESS_KEY_ID
	S3SecretKey string // YANDEX_SECRET_ACCESS_KEY
	S3Bucket    string // S3_BUCKET
}

// UseS3 — true, если заданы оба ключа Object Storage.
func (c Config) UseS3() bool {
	return c.S3AccessKey != "" && c.S3SecretKey != ""
}

// Load загружает .env (если он есть) и собирает конфигурацию.
func Load() Config {
	_ = godotenv.Load() // .env опционален: в Docker переменные приходят из окружения

	return Config{
		Port:        env("PORT", "8080"),
		DBPath:      env("DB_PATH", "./data.db"),
		CORSOrigins: splitList(env("CORS_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000")),
		AuthSecret:  env("AUTH_SECRET", DefaultAuthSecret),
		TokenTTLHrs: envInt("TOKEN_TTL_HOURS", 72),
		UploadDir:   env("UPLOAD_DIR", "./uploads"),
		SeedDemo:    envBool("SEED_DEMO", false),
		S3AccessKey: os.Getenv("YANDEX_ACCESS_KEY_ID"),
		S3SecretKey: os.Getenv("YANDEX_SECRET_ACCESS_KEY"),
		S3Bucket:    env("S3_BUCKET", "pets-photos"),
	}
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key))); err == nil && v > 0 {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	if v, err := strconv.ParseBool(strings.TrimSpace(os.Getenv(key))); err == nil {
		return v
	}
	return def
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
