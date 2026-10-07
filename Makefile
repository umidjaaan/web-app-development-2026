.PHONY: run deps up down migrate migrate-drop media diagrams fmt

# Запуск приложения
run:
	go run ./cmd/main

# Загрузка зависимостей и генерация go.sum
deps:
	go mod tidy

# Поднять PostgreSQL, Adminer и Minio
up:
	docker compose -f deployments/docker-compose.yml up -d

# Остановить инфраструктуру
down:
	docker compose -f deployments/docker-compose.yml down

# Создать таблицы и наполнить их
migrate:
	go run ./cmd/migrate

# Пересоздать таблицы с нуля
migrate-drop:
	go run ./cmd/migrate -drop

# Перегенерировать изображения и видео (нужны python3, Pillow, numpy, ffmpeg)
media:
	python3 tools/genmedia/gen.py

# Перерисовать ER-диаграмму и диаграмму классов (docs/*.svg, docs/*.drawio)
diagrams:
	python3 tools/gen_er.py
	python3 tools/gen_classes.py

fmt:
	gofmt -w ./cmd ./internal
