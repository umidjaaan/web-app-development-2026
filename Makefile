.PHONY: run run-local deps up down migrate migrate-drop media placeholder figma fmt

# Запуск приложения
run:
	go run ./cmd/main

# Запуск без Minio: медиа отдаются из локальной папки media/
run-local:
	MEDIA_BASE_URL=/media go run ./cmd/main

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

# Перерисовать изображение-заглушку
placeholder:
	python3 tools/genmedia/placeholder.py

# Перерисовать макеты трёх экранов для Figma
figma:
	python3 tools/genfigma/gen_svg.py

fmt:
	gofmt -w ./cmd ./internal
