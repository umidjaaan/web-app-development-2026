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

// LoginResponse — ответ аутентификации: JWT и пользователь.
// Токен передаётся в остальных запросах заголовком
// «Authorization: Bearer <access_token>».
type LoginResponse struct {
	AccessToken string         `json:"access_token"`
	TokenType   string         `json:"token_type"` // всегда Bearer
	ExpiresIn   int            `json:"expires_in"` // срок действия, секунды
	User        UserSerializer `json:"user"`
}
