// Package repository — слой доступа к данным (Model в терминах MVT).
// На этом этапе коллекция import_categories хранится в оперативной памяти
// в виде массива; база данных подключается в ЛР2.
package repository

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"rip/lab1/internal/app/ds"
)

// ErrNotFound возвращается, когда категория отсутствует или недоступна.
var ErrNotFound = errors.New("категория импорта не найдена")

// Repository инкапсулирует коллекцию import_categories.
type Repository struct {
	importCategories []ds.ImportCategory
}

// New создаёт репозиторий и проставляет ссылки на объекты Minio.
// mediaBase — публичный базовый URL бакета, например
// http://localhost:9000/import-categories
func New(mediaBase string) *Repository {
	base := strings.TrimRight(mediaBase, "/")

	items := make([]ds.ImportCategory, len(importCategories))
	copy(items, importCategories)
	for i := range items {
		items[i].ImageURL = fmt.Sprintf("%s/img/%s.jpg", base, items[i].Slug)
		items[i].VideoURL = fmt.Sprintf("%s/video/%s.mp4", base, items[i].Slug)
	}

	return &Repository{importCategories: items}
}

// Site возвращает название памятника, по которому собрана сводка.
func (r *Repository) Site() string { return Site }

// ImportCategories возвращает действующие категории, появившиеся не позднее
// указанной даты начала бытования. Нулевая дата означает «фильтр не задан».
// Фильтрация выполняется на сервере.
func (r *Repository) ImportCategories(notLaterThan time.Time) []ds.ImportCategory {
	result := make([]ds.ImportCategory, 0, len(r.importCategories))
	for _, c := range r.importCategories {
		if c.Status != ds.StatusActive {
			continue
		}
		// Год до н. э. хранится положительным числом, поэтому «появилась
		// не позднее выбранной даты» — это DateStart НЕ РАНЬШЕ этой даты
		// в хранимом виде (620 г. до н. э. предшествует 500 г. до н. э.).
		if !notLaterThan.IsZero() && c.DateStart.Before(notLaterThan) {
			continue
		}
		result = append(result, c)
	}
	// Сортировка хронологическая: от самых ранних категорий к поздним.
	// В хранимом виде это убывание даты — см. комментарий к ds.BCE.
	sort.Slice(result, func(i, j int) bool {
		return result[i].DateStart.After(result[j].DateStart)
	})
	return result
}

// ImportCategoryByID возвращает действующую категорию по идентификатору.
// Удалённые категории через этот метод недоступны.
func (r *Repository) ImportCategoryByID(id int) (ds.ImportCategory, error) {
	for _, c := range r.importCategories {
		if c.ID == id && c.Status == ds.StatusActive {
			return c, nil
		}
	}
	return ds.ImportCategory{}, ErrNotFound
}

// FirstImportCategoryID — идентификатор первой категории ленты.
func (r *Repository) FirstImportCategoryID() int {
	all := r.ImportCategories(time.Time{})
	if len(all) == 0 {
		return 0
	}
	return all[0].ID
}

// NextImportCategoryID возвращает идентификатор следующей категории
// ленты (по кругу).
func (r *Repository) NextImportCategoryID(id int) int {
	all := r.ImportCategories(time.Time{})
	if len(all) == 0 {
		return 0
	}
	for i, c := range all {
		if c.ID == id {
			return all[(i+1)%len(all)].ID
		}
	}
	return all[0].ID
}

// Position возвращает порядковый номер категории в ленте и общее их число.
func (r *Repository) Position(id int) (int, int) {
	all := r.ImportCategories(time.Time{})
	for i, c := range all {
		if c.ID == id {
			return i + 1, len(all)
		}
	}
	return 0, len(all)
}

// TotalFinds — общая масса учтённых импортных находок памятника.
func (r *Repository) TotalFinds() int {
	total := 0
	for _, c := range r.ImportCategories(time.Time{}) {
		total += c.FindsCount
	}
	return total
}

// RegionShares — удельный вес каждого региона-поставщика в общей массе
// находок. Именно этот расчёт лежит в основе будущей «заявки»
// на реконструкцию торговых путей (ЛР3).
func (r *Repository) RegionShares() []ds.RegionShare {
	total := r.TotalFinds()
	byRegion := make(map[string]*ds.RegionShare)

	for _, c := range r.ImportCategories(time.Time{}) {
		share, ok := byRegion[c.Region]
		if !ok {
			share = &ds.RegionShare{Region: c.Region}
			byRegion[c.Region] = share
		}
		share.Finds += c.FindsCount
		share.Categories++
	}

	result := make([]ds.RegionShare, 0, len(byRegion))
	for _, share := range byRegion {
		if total > 0 {
			share.Percent = float64(share.Finds) / float64(total) * 100.0
		}
		result = append(result, *share)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Finds > result[j].Finds
	})
	return result
}

// MaxRegionPercent — наибольшая доля региона, нужна для нормировки диаграммы.
func (r *Repository) MaxRegionPercent() float64 {
	shares := r.RegionShares()
	if len(shares) == 0 {
		return 0
	}
	return shares[0].Percent
}

// ShareOf — доля конкретной категории в общей массе находок, в процентах.
func (r *Repository) ShareOf(c ds.ImportCategory) float64 {
	total := r.TotalFinds()
	if total == 0 {
		return 0
	}
	return float64(c.FindsCount) / float64(total) * 100.0
}

// Regions — список регионов-поставщиков для выпадающего списка формы.
func (r *Repository) Regions() []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, c := range r.importCategories {
		if c.Status == ds.StatusDeleted || seen[c.Region] {
			continue
		}
		seen[c.Region] = true
		result = append(result, c.Region)
	}
	sort.Strings(result)
	return result
}
