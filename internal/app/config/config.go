// Package config читает настройки приложения из файла .env и переменных
// окружения. Переменная окружения имеет приоритет над файлом.
package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// Config — настройки запуска приложения.
type Config struct {
	Addr string // адрес HTTP-сервера, APP_ADDR
	// MediaBaseURL — публичный базовый URL бакета Minio, MEDIA_BASE_URL.
	// Для демонстрации без Minio можно указать /media — тогда файлы
	// отдаются из локальной папки media/.
	MediaBaseURL string
	// CurrentUserID — пользователь, от имени которого работает приложение.
	// Авторизация появится в ЛР4, до тех пор пользователь задаётся здесь.
	CurrentUserID uint
}

// Load читает .env (если он есть) и собирает конфигурацию.
func Load() Config {
	LoadEnvFile(".env")

	id, err := strconv.ParseUint(envOr("CURRENT_USER_ID", "1"), 10, 64)
	if err != nil {
		id = 1
	}

	return Config{
		Addr:          envOr("APP_ADDR", ":8080"),
		MediaBaseURL:  envOr("MEDIA_BASE_URL", "http://localhost:9000/import-categories"),
		CurrentUserID: uint(id),
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
