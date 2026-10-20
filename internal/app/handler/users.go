package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rip/lab1/internal/app/auth"
	"rip/lab1/internal/app/ds"
	"rip/lab1/internal/app/repository"
	"rip/lab1/internal/app/serializer"
)

// errBadCredentials — неверный логин или пароль. Что именно неверно, не
// сообщаем: так нельзя перебором узнать, какие логины есть в системе.
var errBadCredentials = errors.New("неверный логин или пароль")

// Register godoc
//
//	@Summary		Регистрация пользователя
//	@Description	Создаёт пользователя. Пароль хранится в виде хеша bcrypt. is_moderator с клиента не принимается.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		serializer.UserRequest	true	"Логин и пароль"
//	@Success		201		{object}	serializer.Response{data=serializer.UserSerializer}
//	@Failure		400		{object}	serializer.Response
//	@Failure		409		{object}	serializer.Response	"Логин уже занят"
//	@Router			/api/users/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req serializer.UserRequest
	if err := bindJSON(c, &req); err != nil {
		h.fail(c, http.StatusBadRequest, err)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.fail(c, http.StatusInternalServerError, err)
		return
	}
	user := ds.User{Login: strings.TrimSpace(req.Login), Password: hash}
	if err := h.Repository.CreateUser(&user); err != nil {
		h.failRepo(c, err)
		return
	}
	h.Logger.Infof("зарегистрирован пользователь %s (id %d)", user.Login, user.ID)
	h.success(c, http.StatusCreated, serializer.ToUser(user), "пользователь зарегистрирован")
}

// Login godoc
//
//	@Summary		Аутентификация
//	@Description	Проверяет логин и пароль и выдаёт JWT. Токен передаётся в остальных запросах заголовком Authorization: Bearer <токен>.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		serializer.UserRequest	true	"Логин и пароль"
//	@Success		200		{object}	serializer.Response{data=serializer.LoginResponse}
//	@Failure		400		{object}	serializer.Response
//	@Failure		401		{object}	serializer.Response	"Неверный логин или пароль"
//	@Router			/api/users/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req serializer.UserRequest
	if err := bindJSON(c, &req); err != nil {
		h.fail(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.UserByLogin(strings.TrimSpace(req.Login))
	if errors.Is(err, repository.ErrUserNotFound) {
		h.fail(c, http.StatusUnauthorized, errBadCredentials)
		return
	}
	if err != nil {
		h.failRepo(c, err)
		return
	}
	if !auth.CheckPassword(user.Password, req.Password) {
		h.fail(c, http.StatusUnauthorized, errBadCredentials)
		return
	}

	token, _, err := auth.NewToken(user, h.JWTSecret, h.JWTTTL)
	if err != nil {
		h.fail(c, http.StatusInternalServerError, err)
		return
	}
	h.Logger.Infof("вход пользователя %s (id %d)", user.Login, user.ID)
	h.success(c, http.StatusOK, serializer.LoginResponse{
		AccessToken: token,
		TokenType:   strings.TrimSpace(bearerPrefix),
		ExpiresIn:   int(h.JWTTTL.Seconds()),
		User:        serializer.ToUser(user),
	}, "вход выполнен")
}

// Logout godoc
//
//	@Summary		Деавторизация
//	@Description	Кладёт текущий JWT в чёрный список Redis до конца срока его действия: после этого токен не принимается.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	serializer.Response
//	@Failure		401	{object}	serializer.Response	"Нет токена или токен недействителен"
//	@Router			/api/users/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	claims, _ := auth.Current(c) // middleware уже проверил токен
	token := auth.CurrentToken(c)

	if err := h.Redis.WriteJWTToBlacklist(c.Request.Context(), token, claims.Login, claims.ExpiresIn()); err != nil {
		h.fail(c, http.StatusInternalServerError, err)
		return
	}
	h.Logger.Infof("выход пользователя %s: токен в blacklist", claims.Login)
	h.success(c, http.StatusOK, nil, "выход выполнен, токен отозван")
}
