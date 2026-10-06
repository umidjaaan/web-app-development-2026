package ds

import "time"

// User — пользователь приложения. Таблица users.
//
// Авторизации в ЛР2 ещё нет: текущий пользователь задаётся в конфигурации
// (CURRENT_USER_ID). Настоящий вход по логину и паролю появится в ЛР4,
// поэтому пароль пока хранится как есть, без хеширования.
type User struct {
	ID          uint   `gorm:"primaryKey"`
	Login       string `gorm:"size:64;uniqueIndex;not null"`
	Password    string `gorm:"size:128;not null"`
	FullName    string `gorm:"size:128"`
	IsModerator bool   `gorm:"not null;default:false"`

	CreatedAt time.Time

	Likes []Like `gorm:"foreignKey:UserID"`
}

// TableName — имя таблицы пользователей.
func (User) TableName() string { return "users" }

// RoleLabel — подпись роли для шаблонов.
func (u User) RoleLabel() string {
	if u.IsModerator {
		return "Модератор"
	}
	return "Исследователь"
}
