// Package redis — клиент Redis. В Redis хранится чёрный список (blacklist)
// JWT: при выходе пользователя его токен записывается сюда и перестаёт
// приниматься, пока не истечёт его срок действия.
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Префиксы ключей: «сервис.назначение.значение», чтобы ключи разных
// приложений в одном Redis не пересекались.
const (
	servicePrefix = "amphora."
	jwtPrefix     = "jwt."
)

// Client — соединение с Redis.
type Client struct {
	client *goredis.Client
}

// New подключается к Redis и проверяет соединение командой PING.
func New(ctx context.Context, addr, password string) (*Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:        addr,
		Password:    password,
		DB:          0,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis %s: %w", addr, err)
	}
	return &Client{client: client}, nil
}

// Close закрывает соединение.
func (c *Client) Close() error { return c.client.Close() }

func jwtKey(token string) string { return servicePrefix + jwtPrefix + token }

// WriteJWTToBlacklist кладёт токен в чёрный список до конца срока его
// действия (ttl). Значение — логин пользователя: так в redis-cli видно,
// чей это токен.
func (c *Client) WriteJWTToBlacklist(ctx context.Context, token, login string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil // токен уже истёк — хранить незачем
	}
	return c.client.Set(ctx, jwtKey(token), login, ttl).Err()
}

// IsJWTBlacklisted — есть ли токен в чёрном списке.
func (c *Client) IsJWTBlacklisted(ctx context.Context, token string) (bool, error) {
	err := c.client.Get(ctx, jwtKey(token)).Err()
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, goredis.Nil):
		return false, nil
	default:
		return false, err
	}
}
