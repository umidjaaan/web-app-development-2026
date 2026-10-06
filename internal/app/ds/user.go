package ds

// User — пользователь приложения. Таблица users.
//
// Авторизации в ЛР2 ещё нет: текущий пользователь задаётся в конфигурации
// (CURRENT_USER_ID). Вход по логину и паролю появится в ЛР4,
// поэтому пароль пока хранится как есть, без хеширования.
type User struct {
	ID          uint   `gorm:"primaryKey"`
	Login       string `gorm:"size:64;uniqueIndex;not null"`
	Password    string `gorm:"size:128;not null"`
	IsModerator bool   `gorm:"not null;default:false"`

	Likes []Like `gorm:"foreignKey:UserID"`
}

// TableName — имя таблицы пользователей.
func (User) TableName() string { return "users" }
