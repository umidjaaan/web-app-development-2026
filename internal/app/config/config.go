// Package config читает настройки приложения из файла .env и переменных
// окружения. Переменная окружения имеет приоритет над файлом.
package config

import (
	"bufio"
	"os"
	"strings"
)

// Config — настройки запуска приложения.
type Config struct {
	Addr string // адрес HTTP-сервера, APP_ADDR

	// MinIO — хранилище изображений и видео услуг.
	MinioEndpoint  string // MINIO_ENDPOINT, например localhost:9000
	MinioAccessKey string // MINIO_ACCESS_KEY
	MinioSecretKey string // MINIO_SECRET_KEY
	MinioBucket    string // MINIO_BUCKET
	// MediaBaseURL — публичный адрес бакета: из него собираются ссылки
	// на файлы в ответах API (в БД хранятся только имена файлов).
	MediaBaseURL string // MEDIA_BASE_URL
}

// Load читает .env (если он есть) и собирает конфигурацию.
func Load() Config {
	LoadEnvFile(".env")

	return Config{
		Addr:           envOr("APP_ADDR", ":8080"),
		MinioEndpoint:  envOr("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: envOr("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: envOr("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    envOr("MINIO_BUCKET", "import-categories"),
		MediaBaseURL:   envOr("MEDIA_BASE_URL", "http://localhost:9000/import-categories"),
	}
}

// LoadEnvFile подхватывает переменные из файла .env.
// Строки вида KEY=value; пустые строки и комментарии (#) пропускаются.
// Уже заданные переменные окружения не перезаписываются.
func LoadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return // .env необязателен: без него берутся значения по умолчанию
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
