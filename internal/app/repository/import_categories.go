package repository

import (
	"time"

	"gorm.io/gorm"

	"rip/lab1/internal/app/ds"
)

// published — базовый запрос: только опубликованные услуги.
// Черновики и удалённые записи в список и ленту не попадают.
func (r *Repository) published() *gorm.DB {
	return r.db.Model(&ds.ImportCategory{}).Where("status = ?", ds.StatusPublished)
}

// feedOrder — порядок ленты и списка: от ранних категорий к поздним.
// Год до н. э. хранится положительным числом, поэтому это убывание даты.
const feedOrder = "date_start DESC, id ASC"

// ==========================================================================
// Чтение
// ==========================================================================

// ImportCategories — опубликованные услуги с фильтром по дате начала.
func (r *Repository) ImportCategories(dateStart time.Time) ([]ds.ImportCategory, error) {
	var items []ds.ImportCategory

	query := r.published()
	if !dateStart.IsZero() {
		// «Тип появился не позднее выбранной даты»: для годов до н. э.,
		// хранимых положительным числом, это date_start >= выбранной даты.
		query = query.Where("date_start >= ?", dateStart)
	}
	if err := query.Order(feedOrder).Find(&items).Error; err != nil {
		return nil, err
	}
	if err := r.fillLikes(items); err != nil {
		return nil, err
	}
	return items, nil
}

// FeedIDs — идентификаторы опубликованных услуг в порядке ленты.
func (r *Repository) FeedIDs() ([]uint, error) {
	var ids []uint
	err := r.published().Order(feedOrder).Pluck("id", &ids).Error
	return ids, err
}

// PublishedByID — одна опубликованная услуга вместе с создателем
// (вложенная сериализация: Preload подтягивает связанную запись users).
func (r *Repository) PublishedByID(id uint) (ds.ImportCategory, error) {
	var item ds.ImportCategory
	err := r.published().Preload("Creator").Where("id = ?", id).First(&item).Error
	if err != nil {
		return ds.ImportCategory{}, notFound(err, ErrNotFound)
	}
	items := []ds.ImportCategory{item}
	if err := r.fillLikes(items); err != nil {
		return ds.ImportCategory{}, err
	}
	return items[0], nil
}

// IsLiked — лайкнул ли пользователь услугу (признак 0/1 в ленте).
func (r *Repository) IsLiked(userID, categoryID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ds.Like{}).
		Where("user_id = ? AND import_category_id = ?", userID, categoryID).
		Count(&count).Error
	return count > 0, err
}

// DraftByUser — черновик пользователя (не больше одного) вместе с создателем.
func (r *Repository) DraftByUser(userID uint) (ds.ImportCategory, error) {
	var item ds.ImportCategory
	err := r.db.Preload("Creator").
		Where("status = ? AND creator_id = ?", ds.StatusDraft, userID).
		First(&item).Error
	if err != nil {
		return ds.ImportCategory{}, notFound(err, ErrNoDraft)
	}
	return item, nil
}

// HasDraft — есть ли у пользователя черновик.
func (r *Repository) HasDraft(userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ds.ImportCategory{}).
		Where("status = ? AND creator_id = ?", ds.StatusDraft, userID).
		Count(&count).Error
	return count > 0, err
}

// ==========================================================================
// Изменение
// ==========================================================================

// CreateDraft — создание услуги: статус «черновик» ставит бэкенд.
// Второй черновик создать нельзя.
func (r *Repository) CreateDraft(item *ds.ImportCategory) error {
	exists, err := r.HasDraft(item.CreatorID)
	if err != nil {
		return err
	}
	if exists {
		return ErrDraftExists
	}
	item.Status = ds.StatusDraft
	return r.db.Create(item).Error
}

// PublishDraft — публикация: черновик → опубликована, вместе с полями,
// которые заполняются перед публикацией. Условие status = 'draft'
// не даёт опубликовать повторно или «вернуть» удалённую услугу.
func (r *Repository) PublishDraft(id, userID uint, dateStart, dateEnd time.Time, description *string) error {
	result := r.db.Model(&ds.ImportCategory{}).
		Where("id = ? AND creator_id = ? AND status = ?", id, userID, ds.StatusDraft).
		Updates(map[string]any{
			"date_start":  dateStart,
			"date_end":    dateEnd,
			"description": description,
			"status":      ds.StatusPublished,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNoDraft
	}
	return nil
}

// DeleteImportCategory — мягкое удаление: статус «удалена», запись остаётся.
// Удалить можно только свою услугу (черновик или опубликованную).
// Чужая, удалённая или несуществующая услуга — ErrNotYours.
func (r *Repository) DeleteImportCategory(id, userID uint) error {
	result := r.db.Model(&ds.ImportCategory{}).
		Where("id = ? AND creator_id = ? AND status <> ?", id, userID, ds.StatusDeleted).
		Update("status", ds.StatusDeleted)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotYours
	}
	return nil
}

// SetLike — like = true ставит лайк, false отменяет. Только для
// опубликованных услуг. Ответ «успешно» возможен лишь при изменении БД,
// поэтому повторный лайк и отмена несуществующего лайка — конфликт.
// Возвращает новое число лайков услуги.
func (r *Repository) SetLike(userID, categoryID uint, like bool) (int, error) {
	if _, err := r.PublishedByID(categoryID); err != nil {
		return 0, err
	}

	if like {
		liked, err := r.IsLiked(userID, categoryID)
		if err != nil {
			return 0, err
		}
		if liked {
			return 0, ErrLiked
		}
		if err := r.db.Create(&ds.Like{UserID: userID, ImportCategoryID: categoryID}).Error; err != nil {
			return 0, err
		}
	} else {
		result := r.db.Where("user_id = ? AND import_category_id = ?", userID, categoryID).
			Delete(&ds.Like{})
		if result.Error != nil {
			return 0, result.Error
		}
		if result.RowsAffected == 0 {
			return 0, ErrNotLiked
		}
	}

	var total int64
	err := r.db.Model(&ds.Like{}).Where("import_category_id = ?", categoryID).Count(&total).Error
	return int(total), err
}

// fillLikes проставляет число лайков одним запросом GROUP BY по likes.
func (r *Repository) fillLikes(items []ds.ImportCategory) error {
	if len(items) == 0 {
		return nil
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
		return err
	}

	byID := make(map[uint]int, len(rows))
	for _, row := range rows {
		byID[row.ImportCategoryID] = row.Total
	}
	for i := range items {
		items[i].LikesTotal = byID[items[i].ID]
	}
	return nil
}
