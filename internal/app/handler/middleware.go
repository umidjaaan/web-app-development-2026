package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rip/lab1/internal/app/auth"
)

// bearerPrefix — токен передаётся в заголовке «Authorization: Bearer <JWT>».
const bearerPrefix = "Bearer "

// WithAuthCheck — middleware авторизации.
//
//	required = true  — метод создателя: без действующего токена ответ 401;
//	required = false — метод чтения: доступен и гостю, а с токеном
//	                   дополнительно заполняются признаки is_creator / is_liked.
//
// Токен проверяется по подписи и сроку действия, а затем по чёрному списку
// в Redis (туда он попадает после выхода).
func (h *Handler) WithAuthCheck(required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenFromHeader(c)
		if token == "" {
			if required {
				h.abort(c, http.StatusUnauthorized, errors.New("нужна авторизация: передайте заголовок Authorization: Bearer <токен>"))
				return
			}
			c.Next() // гость
			return
		}

		claims, err := auth.ParseToken(token, h.JWTSecret)
		if err == nil {
			var blacklisted bool
			blacklisted, err = h.Redis.IsJWTBlacklisted(c.Request.Context(), token)
			if err != nil {
				h.abort(c, http.StatusInternalServerError, err)
				return
			}
			if blacklisted {
				err = errors.New("токен отозван: пользователь вышел из системы")
			}
		}
		if err != nil {
			if required {
				h.abort(c, http.StatusUnauthorized, err)
				return
			}
			c.Next() // для чтения недействительный токен = гость
			return
		}

		auth.SetCurrent(c, claims, token)
		c.Next()
	}
}

// tokenFromHeader достаёт JWT из заголовка Authorization (префикс Bearer
// можно не указывать — так удобнее вставлять токен в Swagger).
func tokenFromHeader(c *gin.Context) string {
	header := strings.TrimSpace(c.GetHeader("Authorization"))
	if after, ok := strings.CutPrefix(header, bearerPrefix); ok {
		return strings.TrimSpace(after)
	}
	return header
}

// abort завершает запрос с ошибкой, не вызывая обработчик метода.
func (h *Handler) abort(c *gin.Context, code int, err error) {
	h.fail(c, code, err)
	c.Abort()
}
