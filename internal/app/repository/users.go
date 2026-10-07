package repository

import "rip/lab1/internal/app/ds"

// CreateUser — регистрация. Логин должен быть свободен;
// is_moderator — системное поле, новый пользователь всегда не модератор.
func (r *Repository) CreateUser(user *ds.User) error {
	var count int64
	if err := r.db.Model(&ds.User{}).Where("login = ?", user.Login).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrLoginTaken
	}
	user.IsModerator = false
	return r.db.Create(user).Error
}
