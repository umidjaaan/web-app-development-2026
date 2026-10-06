package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"rip/lab1/internal/app/config"
	"rip/lab1/internal/app/dsn"
)

// Connect открывает соединение с PostgreSQL через GORM.
// Параметры подключения берутся из .env (см. internal/app/dsn).
//
// Логгер GORM включён в режиме Info: в консоли виден каждый SQL-запрос —
// это нужно показать на защите.
func Connect() (*gorm.DB, error) {
	config.LoadEnvFile(".env")

	return gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
}
