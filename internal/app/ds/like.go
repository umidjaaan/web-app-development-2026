package ds

import "time"

// Like — связь «многие ко многим» между пользователем и категорией импорта:
// пользователь отметил категорию. Таблица likes.
//
// Пара (user_id, import_category_id) уникальна, поэтому один пользователь
// не может отметить одну и ту же категорию дважды.
type Like struct {
	ID uint `gorm:"primaryKey"`

	UserID           uint `gorm:"not null;uniqueIndex:idx_likes_user_category"`
	ImportCategoryID uint `gorm:"not null;uniqueIndex:idx_likes_user_category"`

	CreatedAt time.Time

	User           *User           `gorm:"foreignKey:UserID"`
	ImportCategory *ImportCategory `gorm:"foreignKey:ImportCategoryID"`
}

// TableName — имя связной таблицы.
func (Like) TableName() string { return "likes" }
