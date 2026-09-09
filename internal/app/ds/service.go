// Package ds (data structures) описывает предметную область приложения
// «Амфора» — оценка торговых связей по статистике импортных находок.
package ds

import "fmt"

// Status — состояние услуги. По заданию ЛР1 услуга имеет один из трёх статусов.
type Status string

const (
	StatusDraft     Status = "draft"     // черновик — виден только на странице добавления
	StatusPublished Status = "published" // опубликован — виден в ленте и плитке
	StatusDeleted   Status = "deleted"   // удалён — не отображается нигде
)

// Label — человекочитаемое название статуса для шаблонов.
func (s Status) Label() string {
	switch s {
	case StatusDraft:
		return "Черновик"
	case StatusPublished:
		return "Опубликовано"
	case StatusDeleted:
		return "Удалено"
	}
	return string(s)
}

// ImportCategory — «услуга» предметной области: характерная категория
// импортных товаров с указанием центра производства.
//
// Поля ImageURL и VideoURL хранят ключи-ссылки на объекты в Minio
// (изображение и видео лежат в разных префиксах бакета).
type ImportCategory struct {
	ID   int
	Slug string

	Title            string // «Аттическая чернофигурная керамика»
	Shape            string // морфологический тип: килик, амфора, чаша…
	ProductionCenter string // центр производства: «Афины, квартал Керамик»
	Region           string // регион-поставщик — по нему считается удельный вес
	Period           string // датировка бытования
	Diagnostics      string // признаки-маркёры, по которым категория опознаётся
	Description      string // развёрнутая справка

	FindsCount int // число фрагментов в сводке по памятнику (параметр фильтрации)

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
