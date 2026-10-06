package ds

// Like — связь «многие ко многим» между пользователем и услугой:
// пользователь отметил услугу. Таблица likes.
//
// Пара (user_id, import_category_id) уникальна: один пользователь
// не может отметить одну и ту же услугу дважды.
type Like struct {
	ID uint `gorm:"primaryKey"`

	UserID           uint `gorm:"not null;uniqueIndex:idx_likes_user_category"`
	ImportCategoryID uint `gorm:"not null;uniqueIndex:idx_likes_user_category"`

	User           *User           `gorm:"foreignKey:UserID"`
	ImportCategory *ImportCategory `gorm:"foreignKey:ImportCategoryID"`
}

// TableName — имя связной таблицы.
func (Like) TableName() string { return "likes" }
