// Package ds — модели предметной области «Амфора» (оценка торговых связей
// по статистике импортных находок). Каждая модель — таблица PostgreSQL:
// по тегам gorm таблицы создаются миграцией (cmd/migrate), теги json
// задают имена полей при сериализации (пароль не сериализуется: json:"-").
package ds

import (
	"fmt"
	"time"
)

// Status — состояние услуги. Меняется только на бэкенде:
//
//	draft     — черновик: создан методом POST /api/import_categories;
//	published — опубликована методом PUT /api/import_categories/:id/publish;
//	deleted   — мягко удалена методом DELETE /api/import_categories/:id.
//
// Вернуть услугу в черновик нельзя.
type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusDeleted   Status = "deleted"
)

// BCE собирает дату по году до нашей эры: BCE(620) — 620 г. до н. э.
//
// HTML-поле <input type="date"> не принимает отрицательный год, поэтому год
// до нашей эры хранится положительным числом: 0620-01-01 читается как
// «620 г. до н. э.». Прямое следствие: чем БОЛЬШЕ хранимая дата, тем РАНЬШЕ
// событие на шкале времени, поэтому сравнения дат в фильтре и сортировке
// инвертированы — каждое такое место отдельно помечено комментарием.
func BCE(year int) time.Time {
	return time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
}

// DateLayout — формат дат в API и в поле <input type="date">: ГГГГ-ММ-ДД.
const DateLayout = "2006-01-02"

// ImportCategory — услуга: категория импортных находок. Таблица import_categories.
//
// При создании (POST) заполняются название (обязательно), фото и видео
// (необязательно) — так появляется черновик. При публикации (PUT)
// заполняются два поля по теме — дата начала и дата конца — и описание.
// Необязательные поля в таблице допускают NULL.
//
// Image и Video — имена файлов в MinIO (например knidskie-amfory-3f9a0c.jpg),
// сами файлы лежат в бакете в папках img/ и video/. Необязательные поля —
// указатели: пустое значение попадает в таблицу как NULL.
type ImportCategory struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Title string  `gorm:"size:128;not null" json:"title"` // обязательно
	Image *string `gorm:"size:256" json:"image"`          // имя файла, NULL
	Video *string `gorm:"size:256" json:"video"`          // имя файла, NULL

	// Два поля по теме, у черновика пустые (NULL). Фильтр списка — по DateStart.
	DateStart *time.Time `gorm:"type:date;index" json:"date_start"`
	DateEnd   *time.Time `gorm:"type:date" json:"date_end"`

	Description *string `gorm:"type:text" json:"description"` // NULL

	// Системные поля: задаются только на бэкенде.
	Status    Status `gorm:"size:16;not null;default:draft;index" json:"status"`
	CreatorID uint   `gorm:"not null;index" json:"creator_id"`
	Creator   *User  `gorm:"foreignKey:CreatorID" json:"-"`

	Likes []Like `gorm:"foreignKey:ImportCategoryID" json:"-"`

	// LikesTotal не хранится в таблице: считается запросом по likes.
	LikesTotal int `gorm:"-" json:"-"`
}

// TableName — имя таблицы услуг.
func (ImportCategory) TableName() string { return "import_categories" }

// StartYear и EndYear — годы до н. э. (0, если дата не заполнена).
func (c ImportCategory) StartYear() int { return year(c.DateStart) }
func (c ImportCategory) EndYear() int   { return year(c.DateEnd) }

// PeriodLabel — интервал бытования: «620–480 гг. до н. э.» (nil у черновика).
func (c ImportCategory) PeriodLabel() *string {
	if c.DateStart == nil || c.DateEnd == nil {
		return nil
	}
	s := fmt.Sprintf("%d–%d гг. до н. э.", c.StartYear(), c.EndYear())
	return &s
}

// DateStartString и DateEndString — даты в формате API (nil, если не заполнены).
func (c ImportCategory) DateStartString() *string { return dateString(c.DateStart) }
func (c ImportCategory) DateEndString() *string   { return dateString(c.DateEnd) }

func year(t *time.Time) int {
	if t == nil {
		return 0
	}
	return t.Year()
}

func dateString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(DateLayout)
	return &s
}

// Optional — необязательное строковое поле: пустая строка → NULL.
func Optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
