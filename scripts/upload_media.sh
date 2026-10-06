#!/usr/bin/env sh
# Ручная заливка медиа в Minio клиентом mc.
# Нужна, если Minio поднят не через docker compose, а отдельно.
#
#   ./scripts/upload_media.sh
#
# Переменные окружения (со значениями по умолчанию):
#   MINIO_ENDPOINT=http://localhost:9000
#   MINIO_USER=minioadmin
#   MINIO_PASSWORD=minioadmin
#   MINIO_BUCKET=import-categories

set -e

ENDPOINT="${MINIO_ENDPOINT:-http://localhost:9000}"
USER="${MINIO_USER:-minioadmin}"
PASSWORD="${MINIO_PASSWORD:-minioadmin}"
BUCKET="${MINIO_BUCKET:-import-categories}"

if ! command -v mc > /dev/null 2>&1; then
    echo "Не найден клиент mc. Установка:"
    echo "  https://min.io/docs/minio/linux/reference/minio-mc.html"
    exit 1
fi

echo "Подключение к $ENDPOINT"
mc alias set rip "$ENDPOINT" "$USER" "$PASSWORD"

echo "Создание бакета $BUCKET"
mc mb --ignore-existing "rip/$BUCKET"

echo "Открытие бакета на чтение (изображения и видео отдаются браузеру напрямую)"
mc anonymous set download "rip/$BUCKET"

echo "Загрузка изображений и видео"
mc cp --recursive media/img   "rip/$BUCKET/"
mc cp --recursive media/video "rip/$BUCKET/"

echo
echo "Готово. Проверить объект:"
echo "  $ENDPOINT/$BUCKET/img/attic-black-figure.jpg"
