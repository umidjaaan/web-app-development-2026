// Package api собирает веб-сервис: конфигурация, подключение к PostgreSQL,
// MinIO и Redis, репозиторий и маршруты REST API.
package api

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"rip/lab1/internal/app/config"
	"rip/lab1/internal/app/handler"
	"rip/lab1/internal/app/redis"
	"rip/lab1/internal/app/repository"
)

// StartServer поднимает веб-сервис.
func StartServer() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	cfg := config.Load()

	db, err := repository.Connect()
	if err != nil {
		logger.Fatalf("Не удалось подключиться к PostgreSQL: %v", err)
	}
	logger.Info("Подключение к PostgreSQL установлено")

	minioClient, err := repository.ConnectMinio(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey)
	if err != nil {
		logger.Fatalf("Не удалось подключиться к MinIO: %v", err)
	}
	logger.Infof("MinIO: %s, бакет %s", cfg.MinioEndpoint, cfg.MinioBucket)

	repo := repository.New(db, minioClient, cfg.MinioBucket, cfg.MediaBaseURL)
	if err := repo.EnsureBucket(); err != nil {
		// API для чтения работает и без MinIO; загрузка файлов — нет.
		logger.Warnf("MinIO: %v — загрузка фото и видео работать не будет", err)
	}
	redisClient, err := redis.New(context.Background(), cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		logger.Fatalf("Не удалось подключиться к Redis: %v", err)
	}
	defer redisClient.Close()
	logger.Infof("Redis: %s (чёрный список JWT)", cfg.RedisAddr)

	h := handler.New(repo, redisClient, cfg.JWTSecret, cfg.JWTTTL, logger)

	router := gin.Default()
	h.RegisterHandler(router)

	logger.Infof("API запущен на %s, адреса начинаются с /api", cfg.Addr)
	logger.Infof("Swagger: http://localhost%s/swagger/index.html", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		logger.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
