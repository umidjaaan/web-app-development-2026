// Package auth — пользователь, от имени которого работает API.
//
// Авторизации в ЛР3 ещё нет (она появится в ЛР4), поэтому создатель услуг
// зафиксирован константой. Получить его можно только через функцию-singleton
// CurrentUserID: значение вычисляется один раз (sync.Once) и дальше
// переиспользуется во всех методах API.
package auth

import "sync"

// creatorID — пользователь-создатель (users.id = 1, логин umid).
const creatorID uint = 1

var (
	once          sync.Once
	currentUserID uint
)

// CurrentUserID — функция-singleton: возвращает идентификатор текущего
// пользователя. В ЛР4 здесь появится чтение пользователя из сессии/токена.
func CurrentUserID() uint {
	once.Do(func() {
		currentUserID = creatorID
	})
	return currentUserID
}
