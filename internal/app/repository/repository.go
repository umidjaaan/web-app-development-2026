// Package repository — слой доступа к данным: таблицы PostgreSQL через
// ORM GORM и файлы услуг в хранилище MinIO.
package repository

import (
	"errors"
	"strings"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

// Ошибки бизнес-логики: обработчики превращают их в HTTP-коды.
var (
	ErrNotFound    = errors.New("услуга не найдена")
	ErrNotYours    = errors.New("услуга не найдена среди услуг текущего пользователя")
	ErrNoDraft     = errors.New("черновик не найден")
	ErrDraftExists = errors.New("у пользователя уже есть черновик")
	ErrLiked       = errors.New("лайк уже поставлен")
	ErrNotLiked    = errors.New("лайк не был поставлен")
	ErrLoginTaken  = errors.New("логин уже занят")
	ErrBadFile     = errors.New("недопустимый файл")
)

// Repository — доступ к таблицам users, import_categories, likes и к MinIO.
type Repository struct {
	db        *gorm.DB
	minio     *minio.Client
	bucket    string
	mediaBase string
}

// New создаёт репозиторий поверх соединений с PostgreSQL и MinIO.
func New(db *gorm.DB, minioClient *minio.Client, bucket, mediaBase string) *Repository {
	return &Repository{
		db:        db,
		minio:     minioClient,
		bucket:    bucket,
		mediaBase: strings.TrimRight(mediaBase, "/"),
	}
}

// notFound переводит «запись не найдена» GORM в ошибку репозитория.
func notFound(err, notFoundErr error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notFoundErr
	}
	return err
}
