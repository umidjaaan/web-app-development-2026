// Package ds (data structures) описывает предметную область приложения
// «Амфора» — оценка торговых связей по статистике импортных находок.
//
// Основная сущность предметной области — import_categories: характерные
// категории импортных товаров с указанием центра производства.
package ds

import (
	"fmt"
	"time"
)

// Status — состояние категории импорта.
//
// Статусов два: «действует» и «удалена» (мягкое удаление). Черновик
// относится не к услуге, а к заявке и появится в ЛР3.
type Status string

const (
	StatusActive  Status = "active"  // действует — видна в ленте и в плитке
	StatusDeleted Status = "deleted" // удалена — не отображается нигде
)

// Label — человекочитаемое название статуса для шаблонов.
func (s Status) Label() string {
	switch s {
	case StatusActive:
		return "Действует"
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

// ImportCategory — одна запись коллекции import_categories: характерная
// категория импортных товаров с указанием центра производства.
//
// Поля ImageURL и VideoURL хранят ссылки на объекты в Minio
// (изображение и видео лежат в разных префиксах бакета).
type ImportCategory struct {
	ID   int
	Slug string

	Title            string // «Аттическая чернофигурная керамика»
	Shape            string // морфологический тип: килик, амфора, чаша…
	ProductionCenter string // центр производства: «Афины, квартал Керамик»
	Region           string // регион-поставщик — по нему считается удельный вес

	// DateStart и DateEnd — два поля предметной области: дата начала
	// бытования типа и дата его конца. По DateStart работает фильтрация.
	DateStart time.Time
	DateEnd   time.Time

	Diagnostics string // признаки-маркёры, по которым категория опознаётся
	Description string // развёрнутая справка

	FindsCount int // число фрагментов в сводке по памятнику

	ImageURL string // ключ изображения в Minio
	VideoURL string // ключ видео в Minio

	LikedBy []int // ID пользователей, отметивших категорию
	Status  Status
}

// Likes — количество лайков карточки.
func (c ImportCategory) Likes() int { return len(c.LikedBy) }

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

// PeriodShort — компактный интервал для карточки плитки: «620–480 до н. э.».
func (c ImportCategory) PeriodShort() string {
	return fmt.Sprintf("%d–%d до н. э.", c.DateStart.Year(), c.DateEnd.Year())
}

// DateStartInput — значение даты начала для <input type="date">: «0620-01-01».
func (c ImportCategory) DateStartInput() string { return c.DateStart.Format(DateLayout) }

// DateEndInput — значение даты конца для <input type="date">.
func (c ImportCategory) DateEndInput() string { return c.DateEnd.Format(DateLayout) }

// DateLayout — формат даты HTML-поля <input type="date">.
const DateLayout = "2006-01-02"

func bceLabel(t time.Time) string {
	return fmt.Sprintf("%d г. до н. э.", t.Year())
}

// RegionShare — удельный вес региона в общей массе импортных находок.
// Это исходные данные для «заявки» — реконструкции торговых путей.
type RegionShare struct {
	Region     string
	Finds      int
	Categories int
	Percent    float64
}

// PercentLabel — доля региона с одним знаком после запятой.
func (r RegionShare) PercentLabel() string {
	return fmt.Sprintf("%.1f%%", r.Percent)
}

// BarWidth — ширина полосы диаграммы в процентах ширины контейнера.
// Нормируется по максимальной доле, чтобы полосы читались на узком экране.
func (r RegionShare) BarWidth(max float64) string {
	if max <= 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", r.Percent/max*100.0)
}
