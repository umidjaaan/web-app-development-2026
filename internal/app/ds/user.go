package ds

// User — пользователь. Таблица users.
//
// Пароль не попадает в JSON (json:"-"). До ЛР4 он хранится как есть,
// хеширование появится вместе с авторизацией.
type User struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Login       string `gorm:"size:64;uniqueIndex;not null" json:"login"`
	Password    string `gorm:"size:128;not null" json:"-"`
	IsModerator bool   `gorm:"not null;default:false" json:"is_moderator"`

	Likes []Like `gorm:"foreignKey:UserID" json:"-"`
}

// TableName — имя таблицы пользователей.
func (User) TableName() string { return "users" }
