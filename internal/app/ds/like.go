package ds

// Like — связь «многие ко многим» между пользователем и услугой:
// пользователь лайкнул услугу. Таблица likes.
//
// Пара (user_id, import_category_id) уникальна: один пользователь
// не может лайкнуть одну и ту же услугу дважды.
type Like struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID           uint `gorm:"not null;uniqueIndex:idx_likes_user_category" json:"user_id"`
	ImportCategoryID uint `gorm:"not null;uniqueIndex:idx_likes_user_category" json:"import_category_id"`

	User           *User           `gorm:"foreignKey:UserID" json:"-"`
	ImportCategory *ImportCategory `gorm:"foreignKey:ImportCategoryID" json:"-"`
}

// TableName — имя связной таблицы.
func (Like) TableName() string { return "likes" }
