package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"rip/lab1/internal/app/ds"
)

// issuer — кто выпустил токен (поле iss).
const issuer = "amphora"

// Claims — полезные данные JWT: стандартные поля (срок действия, время
// выпуска, издатель) и наши — пользователь и его роль.
type Claims struct {
	jwt.RegisteredClaims
	UserID      uint   `json:"user_id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

// NewToken выпускает JWT для пользователя, подписанный секретом (HS256).
// Возвращает строку токена и момент, когда он перестанет действовать.
func NewToken(user ds.User, secret []byte, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID:      user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("подпись токена: %w", err)
	}
	return token, expiresAt, nil
}

// ParseToken проверяет подпись и срок действия токена и возвращает его данные.
func ParseToken(raw string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(issuer))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("срок действия токена истёк")
		}
		return nil, errors.New("токен недействителен")
	}
	if claims.UserID == 0 {
		return nil, errors.New("в токене нет пользователя")
	}
	return claims, nil
}

// ExpiresIn — сколько ещё действует токен (0, если уже истёк).
func (c *Claims) ExpiresIn() time.Duration {
	if c.ExpiresAt == nil {
		return 0
	}
	if left := time.Until(c.ExpiresAt.Time); left > 0 {
		return left
	}
	return 0
}
