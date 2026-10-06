// Package api собирает приложение: конфигурация, подключение к базе данных,
// репозиторий, обработчики, шаблоны и статика.
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"rip/lab1/internal/app/config"
	"rip/lab1/internal/app/handler"
	"rip/lab1/internal/app/repository"
)

// StartServer поднимает веб-сервер приложения.
func StartServer() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	cfg := config.Load()
	logger.Infof("Медиа отдаются с %s", cfg.MediaBaseURL)

	db, err := repository.Connect()
	if err != nil {
		logger.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	logger.Info("Подключение к PostgreSQL установлено")

	repo := repository.New(db, cfg.MediaBaseURL)
	h := handler.New(repo, logger, cfg.CurrentUserID)

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
