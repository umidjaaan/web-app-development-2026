// Команда миграции: создаёт таблицы в PostgreSQL по моделям GORM,
// добавляет ограничения, которые ORM не умеет описывать тегами,
// и наполняет пустые таблицы исходными данными.
//
//	go run ./cmd/migrate           — создать таблицы и наполнить их
//	go run ./cmd/migrate -drop     — пересоздать таблицы с нуля
package main

import (
	"flag"
	"log"

	"rip/lab1/internal/app/config"
	"rip/lab1/internal/app/ds"
	"rip/lab1/internal/app/repository"
)

// uniqueDraftIndex — частичный уникальный индекс: у одного пользователя
// не может быть больше одного черновика.
const uniqueDraftIndex = `
	CREATE UNIQUE INDEX IF NOT EXISTS uniq_draft_per_user
	    ON import_categories (creator_id)
	 WHERE status = 'draft'`

func main() {
	drop := flag.Bool("drop", false, "удалить таблицы перед созданием")
	flag.Parse()

	cfg := config.Load()

	db, err := repository.Connect()
	if err != nil {
		log.Fatalf("подключение к базе данных: %v", err)
	}

	if *drop {
		log.Println("Удаляю таблицы likes, import_categories, users")
		if err := db.Exec(
			"DROP TABLE IF EXISTS likes, import_categories, users CASCADE").Error; err != nil {
			log.Fatalf("удаление таблиц: %v", err)
		}
	}

	log.Println("Создаю таблицы по моделям GORM")
	if err := db.AutoMigrate(&ds.User{}, &ds.ImportCategory{}, &ds.Like{}); err != nil {
		log.Fatalf("миграция: %v", err)
	}

	log.Println("Добавляю частичный уникальный индекс: один черновик на пользователя")
	if err := db.Exec(uniqueDraftIndex).Error; err != nil {
		log.Fatalf("создание индекса: %v", err)
	}

	log.Println("Наполняю таблицы исходными данными")
	repo := repository.New(db, nil, cfg.MinioBucket, cfg.MediaBaseURL)
	if err := repo.Seed(); err != nil {
		log.Fatalf("наполнение таблиц: %v", err)
	}

	log.Println("Готово: таблицы users, import_categories, likes созданы и заполнены")
}
