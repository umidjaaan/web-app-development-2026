package serializer

import "rip/lab1/internal/app/ds"

// URLBuilder собирает ссылки на файлы MinIO по именам из БД
// (реализован репозиторием).
type URLBuilder interface {
	ImageURL(name *string) *string
	VideoURL(name *string) *string
}

// ---------------------------------------------------------------- ответы

// ImportCategorySerializer — услуга в списке (GET /api/import_categories).
type ImportCategorySerializer struct {
	ID        uint    `json:"id"`
	Title     string  `json:"title"`
	ImageURL  *string `json:"image_url"`
	DateStart *string `json:"date_start"`
	DateEnd   *string `json:"date_end"`
	Period    *string `json:"period"`
	Likes     int     `json:"likes"`
	IsCreator int     `json:"is_creator"` // 1 — создатель услуги = текущий пользователь
}

// FeedSerializer — услуга в ленте (GET /api/import_categories/feed).
// Вложенная сериализация: создатель услуги — UserSerializer.
type FeedSerializer struct {
	ID          uint           `json:"id"`
	Title       string         `json:"title"`
	ImageURL    *string        `json:"image_url"`
	VideoURL    *string        `json:"video_url"`
	DateStart   *string        `json:"date_start"`
	DateEnd     *string        `json:"date_end"`
	Period      *string        `json:"period"`
	Description *string        `json:"description"`
	Likes       int            `json:"likes"`
	IsLiked     int            `json:"is_liked"` // 1 — текущий пользователь лайкнул услугу
	Creator     UserSerializer `json:"creator"`
	Position    int            `json:"position"` // номер услуги в ленте
	Total       int            `json:"total"`    // всего услуг в ленте
	NextID      uint           `json:"next_id"`  // ид следующей услуги
}

// DraftSerializer — черновик (POST и GET /api/import_categories/draft).
type DraftSerializer struct {
	ID          uint           `json:"id"`
	Title       string         `json:"title"`
	Image       *string        `json:"image"` // имя файла в MinIO
	ImageURL    *string        `json:"image_url"`
	Video       *string        `json:"video"` // имя файла в MinIO
	VideoURL    *string        `json:"video_url"`
	DateStart   *string        `json:"date_start"`
	DateEnd     *string        `json:"date_end"`
	Description *string        `json:"description"`
	Status      ds.Status      `json:"status"`
	Creator     UserSerializer `json:"creator"`
}

// LikeSerializer — результат лайка.
type LikeSerializer struct {
	ImportCategoryID uint `json:"import_category_id"`
	IsLiked          int  `json:"is_liked"`
	Likes            int  `json:"likes"`
}

// ---------------------------------------------------------------- запросы

// PublishRequest — тело PUT /api/import_categories/:id/publish:
// два поля по теме и описание. Статуса здесь нет — его меняет бэкенд.
type PublishRequest struct {
	DateStart   string `json:"date_start" binding:"required"`
	DateEnd     string `json:"date_end" binding:"required"`
	Description string `json:"description"`
}

// LikeRequest — тело POST /api/import_categories/:id/like: 1 ставит, 0 отменяет.
type LikeRequest struct {
	Like *int `json:"like" binding:"required,oneof=0 1"`
}

// ---------------------------------------------------------------- преобразования

// ToImportCategory — модель → элемент списка.
func ToImportCategory(c ds.ImportCategory, currentUserID uint, urls URLBuilder) ImportCategorySerializer {
	return ImportCategorySerializer{
		ID:        c.ID,
		Title:     c.Title,
		ImageURL:  urls.ImageURL(c.Image),
		DateStart: c.DateStartString(),
		DateEnd:   c.DateEndString(),
		Period:    c.PeriodLabel(),
		Likes:     c.LikesTotal,
		IsCreator: flag(c.CreatorID == currentUserID),
	}
}

// ToFeed — модель → услуга ленты.
func ToFeed(c ds.ImportCategory, isLiked bool, position, total int, nextID uint, urls URLBuilder) FeedSerializer {
	return FeedSerializer{
		ID:          c.ID,
		Title:       c.Title,
		ImageURL:    urls.ImageURL(c.Image),
		VideoURL:    urls.VideoURL(c.Video),
		DateStart:   c.DateStartString(),
		DateEnd:     c.DateEndString(),
		Period:      c.PeriodLabel(),
		Description: c.Description,
		Likes:       c.LikesTotal,
		IsLiked:     flag(isLiked),
		Creator:     toCreator(c),
		Position:    position,
		Total:       total,
		NextID:      nextID,
	}
}

// ToDraft — модель → черновик.
func ToDraft(c ds.ImportCategory, urls URLBuilder) DraftSerializer {
	return DraftSerializer{
		ID:          c.ID,
		Title:       c.Title,
		Image:       c.Image,
		ImageURL:    urls.ImageURL(c.Image),
		Video:       c.Video,
		VideoURL:    urls.VideoURL(c.Video),
		DateStart:   c.DateStartString(),
		DateEnd:     c.DateEndString(),
		Description: c.Description,
		Status:      c.Status,
		Creator:     toCreator(c),
	}
}

func toCreator(c ds.ImportCategory) UserSerializer {
	if c.Creator == nil {
		return UserSerializer{ID: c.CreatorID}
	}
	return ToUser(*c.Creator)
}
