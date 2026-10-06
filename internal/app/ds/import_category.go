// Package ds (data structures) описывает предметную область приложения
// «Амфора» — оценка торговых связей по статистике импортных находок.
//
// Основная сущность предметной области — import_categories: характерные
// категории импортных товаров с указанием центра производства.
// Структуры этого пакета одновременно являются моделями GORM: по ним
// создаются таблицы в PostgreSQL (см. cmd/migrate).
package ds

import (
	"fmt"
	"time"
)

// Status — состояние категории импорта.
//
//	draft     — черновик: создан кнопкой «Далее» на странице добавления;
//	published — опубликована кнопкой «Опубликовать», видна всем;
//	deleted   — логически удалена (SQL UPDATE через курсор).
type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusDeleted   Status = "deleted"
)

// Label — человекочитаемое название статуса для шаблонов.
func (s Status) Label() string {
	switch s {
	case StatusDraft:
		return "Черновик"
	case StatusPublished:
		return "Опубликована"
	case StatusDeleted:
		return "Удалена"
	}
	return string(s)
}

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

// DateLayout — формат даты HTML-поля <input type="date">.
const DateLayout = "2006-01-02"

// ImportCategory — услуга: категория импортных находок.
// Таблица в базе данных — import_categories.
//
// По кнопке «Далее» заполняются название (обязательно), фото и видео
// (необязательно) — так создаётся черновик. Перед публикацией
// заполняются два поля по теме — дата начала и дата конца — и описание.
// Необязательные поля в таблице допускают NULL.
type ImportCategory struct {
	ID uint `gorm:"primaryKey"`

	// «Далее»
	Title    string `gorm:"size:128;not null"` // обязательно
	ImageURL string `gorm:"size:256"`          // необязательно (NULL)
	VideoURL string `gorm:"size:256"`          // необязательно (NULL)

	// Два поля по теме. У черновика пустые (NULL).
	// По DateStart работает фильтрация.
	DateStart *time.Time `gorm:"type:date;index"`
	DateEnd   *time.Time `gorm:"type:date"`

	Description string `gorm:"type:text"` // необязательно (NULL)

	Status Status `gorm:"size:16;not null;default:draft;index"`

	// Создатель услуги — обязателен. Один черновик на пользователя
	// обеспечивает частичный уникальный индекс (cmd/migrate/main.go).
	CreatorID uint  `gorm:"not null;index"`
	Creator   *User `gorm:"foreignKey:CreatorID"`

	Likes []Like `gorm:"foreignKey:ImportCategoryID"`

	// LikesTotal не хранится в таблице: считается запросом по likes.
	LikesTotal int `gorm:"-"`
}

// TableName — имя таблицы услуг.
func (ImportCategory) TableName() string { return "import_categories" }

// LikesCount — количество лайков карточки (используется в шаблонах).
func (c ImportCategory) LikesCount() int { return c.LikesTotal }

// StartYear и EndYear — годы до н. э. (0, если дата не заполнена).
func (c ImportCategory) StartYear() int { return year(c.DateStart) }
func (c ImportCategory) EndYear() int   { return year(c.DateEnd) }

// PeriodLabel — интервал бытования: «620–480 гг. до н. э.».
func (c ImportCategory) PeriodLabel() string {
	if c.DateStart == nil || c.DateEnd == nil {
		return ""
	}
	return fmt.Sprintf("%d–%d гг. до н. э.", c.StartYear(), c.EndYear())
}

// DateStartInput и DateEndInput — значения для <input type="date">.
func (c ImportCategory) DateStartInput() string { return dateInput(c.DateStart) }
func (c ImportCategory) DateEndInput() string   { return dateInput(c.DateEnd) }

func year(t *time.Time) int {
	if t == nil {
		return 0
	}
	return t.Year()
}

func dateInput(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(DateLayout)
}
