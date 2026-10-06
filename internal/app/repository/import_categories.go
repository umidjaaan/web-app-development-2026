// Package repository — слой доступа к данным (Model в терминах MVT).
// Данные лежат в PostgreSQL, обращение к ним идёт через ORM GORM;
// единственное исключение — логическое удаление: оно выполняется
// прямым SQL-запросом UPDATE через курсор (соединение database/sql).
package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"rip/lab1/internal/app/ds"
)

// Site — археологический памятник, по материалам которого собрана сводка.
const Site = "Елизаветовское городище, дельта Дона"

// ErrNotFound возвращается, когда категория отсутствует или недоступна.
var ErrNotFound = errors.New("категория импорта не найдена")

// Repository инкапсулирует работу с таблицами import_categories, users и likes.
type Repository struct {
	db        *gorm.DB
	mediaBase string
}

// New создаёт репозиторий поверх открытого соединения с базой данных.
func New(db *gorm.DB, mediaBase string) *Repository {
	return &Repository{db: db, mediaBase: strings.TrimRight(mediaBase, "/")}
}

// DB отдаёт соединение (нужно для миграций и служебных задач).
func (r *Repository) DB() *gorm.DB { return r.db }

// Site возвращает название памятника, по которому собрана сводка.
func (r *Repository) Site() string { return Site }

// ==========================================================================
// Чтение (GET) — через ORM
// ==========================================================================

// ImportCategories — опубликованные категории для ленты и плитки.
// Фильтр — по дате начала бытования типа (календарь на странице «Плитка»).
// Черновики и удалённые записи сюда не попадают.
func (r *Repository) ImportCategories(dateStart time.Time) ([]ds.ImportCategory, error) {
	var items []ds.ImportCategory

	query := r.db.Where("status = ?", ds.StatusPublished)

	if !dateStart.IsZero() {
		// Год до н. э. хранится положительным числом, поэтому «появилась
		// не позднее выбранной даты» — это date_start >= выбранной даты.
		query = query.Where("date_start >= ?", dateStart)
	}

	// Хронологический порядок: от ранних категорий к поздним
	// (в хранимом виде это убывание даты — см. ds.BCE).
	if err := query.Order("date_start DESC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}

	r.fillLikes(items)
	return items, nil
}

// ImportCategoryByID — одна опубликованная категория (страница ленты).
// Удалённая запись не находится — в приложении это страница 404.
func (r *Repository) ImportCategoryByID(id uint) (ds.ImportCategory, error) {
	var item ds.ImportCategory

	err := r.db.Where("id = ? AND status = ?", id, ds.StatusPublished).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.ImportCategory{}, ErrNotFound
		}
		return ds.ImportCategory{}, err
	}

	single := []ds.ImportCategory{item}
	r.fillLikes(single)
	return single[0], nil
}

// DraftByUser — черновик пользователя, если он есть (у пользователя
// может быть не больше одного черновика — частичный уникальный индекс).
func (r *Repository) DraftByUser(userID uint) (ds.ImportCategory, error) {
	var item ds.ImportCategory

	err := r.db.Where("status = ? AND creator_id = ?", ds.StatusDraft, userID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.ImportCategory{}, ErrNotFound
		}
		return ds.ImportCategory{}, err
	}
	return item, nil
}

// LikesByUser — сколько категорий отметил пользователь (таблица м-м likes).
func (r *Repository) LikesByUser(userID uint) (int, error) {
	var count int64
	err := r.db.Model(&ds.Like{}).Where("user_id = ?", userID).Count(&count).Error
	return int(count), err
}

// UserByID — текущий пользователь приложения.
func (r *Repository) UserByID(id uint) (ds.User, error) {
	var user ds.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, ErrNotFound
		}
		return ds.User{}, err
	}
	return user, nil
}

// ==========================================================================
// Изменение (POST) — через ORM: две кнопки и два статуса
// ==========================================================================

// CreateDraft — кнопка «Далее»: создаёт черновик (INSERT через ORM).
// В черновике только название, фото и видео; даты и описание
// заполняются на следующем шаге, перед публикацией.
func (r *Repository) CreateDraft(item *ds.ImportCategory) error {
	item.Status = ds.StatusDraft
	return r.db.Create(item).Error
}

// PublishDraft — кнопка «Опубликовать»: записывает поля черновика
// и меняет статус на «Опубликована» (UPDATE через ORM).
// Опубликовать можно только свой черновик.
func (r *Repository) PublishDraft(id uint, userID uint, fields map[string]any) error {
	fields["status"] = ds.StatusPublished

	result := r.db.Model(&ds.ImportCategory{}).
		Where("id = ? AND status = ? AND creator_id = ?", id, ds.StatusDraft, userID).
		Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ==========================================================================
// Логическое удаление — прямым SQL UPDATE через курсор (без ORM)
// ==========================================================================

// DeleteImportCategory — логическое удаление категории.
// По заданию выполняется не средствами ORM, а через курсор: берём из GORM
// «голое» соединение database/sql и выполняем SQL-запрос UPDATE.
// Запись остаётся в таблице, меняется только статус.
func (r *Repository) DeleteImportCategory(id uint) error {
	sqlDB, err := r.db.DB() // соединение database/sql — аналог курсора
	if err != nil {
		return err
	}

	res, err := sqlDB.Exec(`
		UPDATE import_categories
		   SET status = 'deleted'
		 WHERE id = $1 AND status <> 'deleted'`, id)
	if err != nil {
		return err
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ==========================================================================
// Навигация по ленте
// ==========================================================================

// FirstImportCategoryID — идентификатор первой категории ленты.
func (r *Repository) FirstImportCategoryID() uint {
	items, err := r.ImportCategories(time.Time{})
	if err != nil || len(items) == 0 {
		return 0
	}
	return items[0].ID
}

// NextImportCategoryID — идентификатор следующей категории ленты (по кругу).
func (r *Repository) NextImportCategoryID(id uint) uint {
	items, err := r.ImportCategories(time.Time{})
	if err != nil || len(items) == 0 {
		return 0
	}
	for i, item := range items {
		if item.ID == id {
			return items[(i+1)%len(items)].ID
		}
	}
	return items[0].ID
}

// Position — порядковый номер категории в ленте и общее их число.
func (r *Repository) Position(id uint) (int, int) {
	items, err := r.ImportCategories(time.Time{})
	if err != nil {
		return 0, 0
	}
	for i, item := range items {
		if item.ID == id {
			return i + 1, len(items)
		}
	}
	return 0, len(items)
}

// ==========================================================================
// Вспомогательное
// ==========================================================================

// fillLikes проставляет количество лайков одним запросом по таблице likes.
func (r *Repository) fillLikes(items []ds.ImportCategory) {
	if len(items) == 0 {
		return
	}

	ids := make([]uint, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}

	var rows []struct {
		ImportCategoryID uint
		Total            int
	}
	err := r.db.Model(&ds.Like{}).
		Select("import_category_id, COUNT(*) AS total").
		Where("import_category_id IN ?", ids).
		Group("import_category_id").
		Scan(&rows).Error
	if err != nil {
		return
	}

	byID := make(map[uint]int, len(rows))
	for _, row := range rows {
		byID[row.ImportCategoryID] = row.Total
	}
	for i := range items {
		items[i].LikesTotal = byID[items[i].ID]
	}
}

func (r *Repository) mediaURL(prefix, slug, ext string) string {
	return fmt.Sprintf("%s/%s/%s.%s", r.mediaBase, prefix, slug, ext)
}
