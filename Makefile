.PHONY: run run-local deps minio-up minio-down media fmt

# Запуск приложения (медиа берутся из Minio)
run:
	go run ./cmd/main

# Запуск без Minio: медиа отдаются из локальной папки media/
run-local:
	MEDIA_BASE_URL=/media go run ./cmd/main

# Загрузка зависимостей и генерация go.sum
deps:
	go mod tidy

# Поднять Minio и залить в него медиа
minio-up:
	docker compose -f deployments/docker-compose.yml up -d

# Остановить Minio
minio-down:
	docker compose -f deployments/docker-compose.yml down

# Перегенерировать изображения и видео (нужны python3, Pillow, numpy, ffmpeg)
media:
	python3 tools/genmedia/gen.py

fmt:
	gofmt -w ./cmd ./internal
