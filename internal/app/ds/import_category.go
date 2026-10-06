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

// ImportCategory — услуга предметной области: характерная категория
// импортных товаров с указанием центра производства.
// Таблица в базе данных — import_categories.
type ImportCategory struct {
	ID   uint   `gorm:"primaryKey"`
	Slug string `gorm:"size:64;uniqueIndex;not null"`

	Title            string `gorm:"size:128;not null"`      // «Аттическая чернофигурная керамика»
	Shape            string `gorm:"size:64"`                // морфологический тип
	ProductionCenter string `gorm:"size:128"`               // центр производства
	Region           string `gorm:"size:64;not null;index"` // регион-поставщик

	// Два поля предметной области: начало и конец бытования типа.
	// По DateStart работает фильтрация.
	DateStart time.Time `gorm:"type:date;not null;index"`
	DateEnd   time.Time `gorm:"type:date;not null"`

	Diagnostics string `gorm:"type:text"` // признаки-маркёры
	Description string `gorm:"type:text"` // развёрнутая справка

	FindsCount int `gorm:"not null;default:0"` // число фрагментов в сводке

	ImageURL string `gorm:"size:256"` // ключ изображения в Minio
	VideoURL string `gorm:"size:256"` // ключ видео в Minio

	Status Status `gorm:"size:16;not null;default:draft;index"`

	// Автор записи. Ограничение «не более одной неопубликованной категории
	// на пользователя» реализовано частичным уникальным индексом
	// (см. cmd/migrate/main.go).
	CreatorID *uint `gorm:"index"`
	Creator   *User `gorm:"foreignKey:CreatorID"`

	Likes []Like `gorm:"foreignKey:ImportCategoryID"`

	CreatedAt time.Time
	UpdatedAt time.Time

	// LikesTotal не хранится в таблице: репозиторий заполняет его
	// отдельным запросом-подсчётом по таблице лайков.
	LikesTotal int `gorm:"-"`
}

// TableName задаёт имя таблицы явно — оно должно совпадать с предметной
// областью и с адресами приложения.
func (ImportCategory) TableName() string { return "import_categories" }

// LikesCount — количество лайков карточки (используется в шаблонах).
func (c ImportCategory) LikesCount() int { return c.LikesTotal }

// FindsLabel — подпись количества находок для карточки.
func (c ImportCategory) FindsLabel() string {
	return fmt.Sprintf("%d фр.", c.FindsCount)
}

// DateStartLabel — дата начала бытования: «620 г. до н. э.».
func (c ImportCategory) DateStartLabel() string { return bceLabel(c.DateStart) }

// DateEndLabel — дата конца бытования: «480 г. до н. э.».
func (c ImportCategory) DateEndLabel() string { return bceLabel(c.DateEnd) }

// PeriodLabel — интервал бытования: «620–480 гг. до н. э.».
func (c ImportCategory) PeriodLabel() string {
	return fmt.Sprintf("%d–%d гг. до н. э.", c.DateStart.Year(), c.DateEnd.Year())
}

// PeriodShort — компактный интервал для карточки плитки.
func (c ImportCategory) PeriodShort() string {
	return fmt.Sprintf("%d–%d до н. э.", c.DateStart.Year(), c.DateEnd.Year())
}

// DateStartInput — значение даты начала для <input type="date">.
func (c ImportCategory) DateStartInput() string { return c.DateStart.Format(DateLayout) }

// DateEndInput — значение даты конца для <input type="date">.
func (c ImportCategory) DateEndInput() string { return c.DateEnd.Format(DateLayout) }

// IsPublished — опубликована ли категория (для шаблонов).
func (c ImportCategory) IsPublished() bool { return c.Status == StatusPublished }

func bceLabel(t time.Time) string {
	return fmt.Sprintf("%d г. до н. э.", t.Year())
}
