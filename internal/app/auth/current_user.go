// Package auth — аутентификация и авторизация пользователей (ЛР4).
//
// В ЛР3 создатель услуг был зафиксирован константой через функцию-singleton.
// Теперь пользователь берётся из JWT: middleware проверяет токен из заголовка
// Authorization и кладёт его данные (claims) в контекст запроса, а методы
// API читают их функцией CurrentUserID.
package auth

import "github.com/gin-gonic/gin"

// Ключи контекста gin, под которыми middleware сохраняет данные токена.
const (
	claimsKey = "auth.claims" // *Claims — пользователь из токена
	tokenKey  = "auth.token"  // string — сам токен (нужен для выхода)
)

// SetCurrent сохраняет проверенный токен и его данные в контексте запроса.
func SetCurrent(c *gin.Context, claims *Claims, token string) {
	c.Set(claimsKey, claims)
	c.Set(tokenKey, token)
}

// Current — данные пользователя из токена. ok = false у гостя.
func Current(c *gin.Context) (*Claims, bool) {
	value, exists := c.Get(claimsKey)
	if !exists {
		return nil, false
	}
	claims, ok := value.(*Claims)
	return claims, ok
}

// CurrentUserID — идентификатор текущего пользователя (0 у гостя).
// Гость может только читать: у него признаки is_creator и is_liked равны 0.
func CurrentUserID(c *gin.Context) uint {
	if claims, ok := Current(c); ok {
		return claims.UserID
	}
	return 0
}

// CurrentToken — токен текущего запроса (пустая строка у гостя).
func CurrentToken(c *gin.Context) string {
	return c.GetString(tokenKey)
}
