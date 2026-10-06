// Package dsn собирает строку подключения к PostgreSQL из переменных окружения.
package dsn

import (
	"fmt"
	"os"
)

// FromEnv возвращает строку подключения вида
//
//	host=localhost port=5433 user=rip password=rip dbname=amphora sslmode=disable
//
// Значения берутся из переменных окружения DB_HOST, DB_PORT, DB_USER,
// DB_PASS, DB_NAME (файл .env в корне проекта).
func FromEnv() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		envOr("DB_HOST", "localhost"),
		envOr("DB_PORT", "5433"),
		envOr("DB_USER", "rip"),
		envOr("DB_PASS", "rip"),
		envOr("DB_NAME", "amphora"),
	)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
