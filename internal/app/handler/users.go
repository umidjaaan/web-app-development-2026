package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rip/lab1/internal/app/ds"
	"rip/lab1/internal/app/serializer"
)

// Register — POST /api/users/register
// Тело: {"login": "...", "password": "..."}. is_moderator с клиента не принимается.
func (h *Handler) Register(c *gin.Context) {
	var req serializer.UserRequest
	if err := bindJSON(c, &req); err != nil {
		h.fail(c, http.StatusBadRequest, err)
		return
	}

	// Пароль хранится как есть до ЛР4, где появится хеширование.
	user := ds.User{Login: strings.TrimSpace(req.Login), Password: req.Password}
	if err := h.Repository.CreateUser(&user); err != nil {
		h.failRepo(c, err)
		return
	}
	h.Logger.Infof("зарегистрирован пользователь %s (id %d)", user.Login, user.ID)
	h.success(c, http.StatusCreated, serializer.ToUser(user), "пользователь зарегистрирован")
}

// Login — POST /api/users/login. Заглушка: аутентификация появится в ЛР4.
func (h *Handler) Login(c *gin.Context) {
	h.success(c, http.StatusOK, nil, "заглушка: аутентификация будет реализована в ЛР4")
}

// Logout — POST /api/users/logout. Заглушка: деавторизация появится в ЛР4.
func (h *Handler) Logout(c *gin.Context) {
	h.success(c, http.StatusOK, nil, "заглушка: деавторизация будет реализована в ЛР4")
}
