package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// FileKind — вид файла услуги: от него зависят папка в бакете
// и допустимые форматы.
type FileKind struct {
	Folder  string            // папка в бакете MinIO
	Formats map[string]string // MIME-тип → расширение файла
	MaxSize int64             // максимальный размер, байт
}

var (
	// Image — изображение услуги: JPG, PNG или WEBP до 10 МБ.
	Image = FileKind{
		Folder:  "img",
		Formats: map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"},
		MaxSize: 10 << 20,
	}
	// Video — короткое видео услуги: MP4 или WEBM до 50 МБ.
	Video = FileKind{
		Folder:  "video",
		Formats: map[string]string{"video/mp4": ".mp4", "video/webm": ".webm"},
		MaxSize: 50 << 20,
	}
)

// ConnectMinio открывает клиент MinIO (S3-совместимое хранилище файлов).
func ConnectMinio(endpoint, accessKey, secretKey string) (*minio.Client, error) {
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
}

// EnsureBucket создаёт бакет, если его нет, и открывает его на чтение,
// чтобы ссылки на файлы открывались в браузере.
func (r *Repository) EnsureBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := r.minio.BucketExists(ctx, r.bucket)
	if err != nil {
		return fmt.Errorf("MinIO недоступен: %w", err)
	}
	if exists {
		return nil
	}
	if err := r.minio.MakeBucket(ctx, r.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("создание бакета: %w", err)
	}
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow",`+
		`"Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, r.bucket)
	return r.minio.SetBucketPolicy(ctx, r.bucket, policy)
}

// ImageURL и VideoURL — публичные ссылки на файлы по их именам из БД.
// Нет файла (NULL) — нет ссылки: клиент покажет заглушку.
func (r *Repository) ImageURL(name *string) *string { return r.fileURL(Image, name) }
func (r *Repository) VideoURL(name *string) *string { return r.fileURL(Video, name) }

func (r *Repository) fileURL(kind FileKind, name *string) *string {
	if name == nil || *name == "" {
		return nil
	}
	url := r.mediaBase + "/" + kind.Folder + "/" + *name
	return &url
}

// UploadFile проверяет файл, генерирует для него имя на латинице
// и кладёт в бакет MinIO. Возвращает имя файла для поля БД.
// Неподходящий файл — ErrBadFile (ошибка клиента), сбой MinIO — другая ошибка.
//
// Тип определяется по содержимому файла (http.DetectContentType),
// а не по расширению, поэтому переименованный PDF не пройдёт.
func (r *Repository) UploadFile(kind FileKind, title string, header *multipart.FileHeader) (string, error) {
	if header.Size > kind.MaxSize {
		return "", fmt.Errorf("%w: %s больше %d МБ", ErrBadFile, header.Filename, kind.MaxSize>>20)
	}

	file, err := header.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	contentType := http.DetectContentType(buffer[:n])
	ext, ok := kind.Formats[contentType]
	if !ok {
		return "", fmt.Errorf("%w: %s — формат %s не подходит", ErrBadFile, header.Filename, contentType)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	name, err := fileName(title, ext)
	if err != nil {
		return "", err
	}
	_, err = r.minio.PutObject(context.Background(), r.bucket, kind.Folder+"/"+name, file, header.Size,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("загрузка в MinIO: %w", err)
	}
	return name, nil
}

// RemoveFile удаляет файл из MinIO (если запись в БД не создалась).
func (r *Repository) RemoveFile(kind FileKind, name *string) {
	if name == nil {
		return
	}
	_ = r.minio.RemoveObject(context.Background(), r.bucket, kind.Folder+"/"+*name, minio.RemoveObjectOptions{})
}

// fileName — имя файла на латинице: транслит названия услуги
// и случайный суффикс, например «Книдские амфоры» → knidskie-amfory-3f9a0c.jpg.
func fileName(title, ext string) (string, error) {
	random := make([]byte, 3)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	base := translit(title)
	if base == "" {
		base = "file"
	}
	return base + "-" + hex.EncodeToString(random) + ext, nil
}

// translitTable — русские буквы → латиница.
var translitTable = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e",
	'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
	'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
	'ф': "f", 'х': "h", 'ц': "c", 'ч': "ch", 'ш': "sh", 'щ': "sch",
	'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

// translit оставляет латинские буквы и цифры, русские переводит в латиницу,
// остальные символы заменяет дефисом. Длина ограничена 40 символами.
func translit(title string) string {
	var b strings.Builder
	dash := false
	for _, ch := range strings.ToLower(title) {
		part, ok := translitTable[ch]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			part, ok = string(ch), true
		}
		if !ok {
			if b.Len() > 0 && !dash {
				b.WriteByte('-')
				dash = true
			}
			continue
		}
		if part != "" {
			b.WriteString(part)
			dash = false
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) > 40 {
		s = strings.Trim(s[:40], "-")
	}
	return s
}
