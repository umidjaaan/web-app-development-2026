// Package repository — слой доступа к данным (Model в терминах MVT).
// На этом этапе данные хранятся в оперативной памяти в виде одной
// коллекции-массива; база данных подключается в ЛР2.
package repository

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"rip/lab1/internal/app/ds"
)

// ErrNotFound возвращается, когда услуга отсутствует или недоступна.
var ErrNotFound = errors.New("услуга не найдена")

// Repository инкапсулирует коллекцию услуг.
type Repository struct {
	categories []ds.ImportCategory
}

// New создаёт репозиторий и проставляет ссылки на объекты Minio.
// mediaBase — публичный базовый URL бакета, например
// http://localhost:9000/rip-media
func New(mediaBase string) *Repository {
	base := strings.TrimRight(mediaBase, "/")

	items := make([]ds.ImportCategory, len(seed))
	copy(items, seed)
	for i := range items {
		items[i].ImageURL = fmt.Sprintf("%s/img/%s.jpg", base, items[i].Slug)
		items[i].VideoURL = fmt.Sprintf("%s/video/%s.mp4", base, items[i].Slug)
	}

	return &Repository{categories: items}
}

// Site возвращает название памятника, по которому собрана сводка.
func (r *Repository) Site() string { return Site }

// GetPublished возвращает опубликованные услуги, у которых число находок
// не меньше minFinds. Фильтрация выполняется на сервере.
func (r *Repository) GetPublished(minFinds int) []ds.ImportCategory {
	result := make([]ds.ImportCategory, 0, len(r.categories))
	for _, c := range r.categories {
		if c.Status != ds.StatusPublished {
			continue
		}
		if c.FindsCount < minFinds {
			continue
		}
		result = append(result, c)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].FindsCount > result[j].FindsCount
	})
	return result
}

// GetByID возвращает опубликованную услугу по идентификатору.
// Черновики и удалённые услуги через этот метод недоступны.
func (r *Repository) GetByID(id int) (ds.ImportCategory, error) {
	for _, c := range r.categories {
		if c.ID == id && c.Status == ds.StatusPublished {
			return c, nil
		}
	}
	return ds.ImportCategory{}, ErrNotFound
}

// GetDraft возвращает текущий черновик — заготовку услуги,
// которая ещё не опубликована.
func (r *Repository) GetDraft() (ds.ImportCategory, error) {
	for _, c := range r.categories {
		if c.Status == ds.StatusDraft {
			return c, nil
		}
	}
	return ds.ImportCategory{}, ErrNotFound
}

// FirstPublishedID — идентификатор первой услуги ленты.
func (r *Repository) FirstPublishedID() int {
	published := r.GetPublished(0)
	if len(published) == 0 {
		return 0
	}
	return published[0].ID
}

// NextID возвращает идентификатор следующей услуги ленты (по кругу).
func (r *Repository) NextID(id int) int {
	published := r.GetPublished(0)
	if len(published) == 0 {
		return 0
	}
	for i, c := range published {
		if c.ID == id {
			return published[(i+1)%len(published)].ID
		}
	}
	return published[0].ID
}

// Position возвращает порядковый номер услуги в ленте и общее их число.
func (r *Repository) Position(id int) (int, int) {
	published := r.GetPublished(0)
	for i, c := range published {
		if c.ID == id {
			return i + 1, len(published)
		}
	}
	return 0, len(published)
}

// TotalFinds — общая масса учтённых импортных находок памятника.
func (r *Repository) TotalFinds() int {
	total := 0
	for _, c := range r.GetPublished(0) {
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

	for _, c := range r.GetPublished(0) {
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
	for _, c := range r.categories {
		if c.Status == ds.StatusDeleted || seen[c.Region] {
			continue
		}
		seen[c.Region] = true
		result = append(result, c.Region)
	}
	sort.Strings(result)
	return result
}
