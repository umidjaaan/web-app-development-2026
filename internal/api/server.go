// Package api собирает приложение: конфигурация, репозиторий,
// обработчики, шаблоны и статика.
package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"rip/lab1/internal/app/handler"
	"rip/lab1/internal/app/repository"
)

// Config — настройки запуска, читаются из переменных окружения.
type Config struct {
	Addr string // адрес HTTP-сервера
	// MediaBaseURL — публичный базовый URL бакета Minio.
	// Для демонстрации без Minio можно указать /media —
	// тогда файлы отдаются из локальной папки media/.
	MediaBaseURL string
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// LoadConfig собирает конфигурацию из окружения.
func LoadConfig() Config {
	return Config{
		Addr:         envOr("APP_ADDR", ":8080"),
		MediaBaseURL: envOr("MEDIA_BASE_URL", "http://localhost:9000/import-categories"),
	}
}

// StartServer поднимает веб-сервер приложения.
func StartServer() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	cfg := LoadConfig()
	logger.Infof("Медиа отдаются с %s", cfg.MediaBaseURL)

	repo := repository.New(cfg.MediaBaseURL)
	h := handler.New(repo, logger)

	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")

	// Статика: стили и изображения интерфейса.
	router.Static("/resources", "./resources")
	// Резервная отдача медиа из локальной папки (когда Minio не поднят).
	router.Static("/media", "./media")

	h.RegisterHandler(router)

	logger.Infof("Сервер запущен на %s", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		logger.Fatalf("Не удалось запустить сервер: %v", err)
	}
	logger.Info("Сервер остановлен")
}
