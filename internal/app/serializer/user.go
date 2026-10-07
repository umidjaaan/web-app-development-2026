package serializer

import "rip/lab1/internal/app/ds"

// UserSerializer — пользователь без пароля.
type UserSerializer struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

// UserRequest — тело регистрации и входа. is_moderator сюда не входит:
// это системное поле.
type UserRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// ToUser — модель → пользователь для ответа.
func ToUser(u ds.User) UserSerializer {
	return UserSerializer{ID: u.ID, Login: u.Login, IsModerator: u.IsModerator}
}
