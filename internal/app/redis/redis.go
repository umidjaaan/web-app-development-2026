// Package redis — клиент Redis. В Redis хранится ЧЁРНЫЙ СПИСОК (blacklist)
// отозванных JWT — и больше ничего.
//
// При входе в Redis ничего не пишется: JWT самодостаточен, сервер проверяет
// его подпись и срок действия без хранилища. Но выданный JWT нельзя
// «отменить» до истечения срока, поэтому при выходе (logout) идентификатор
// токена (jti) записывается в чёрный список. Пока запись есть, токен
// не принимается (401). Запись живёт ровно столько, сколько осталось жить
// токену (TTL), — потом Redis удаляет её сам.
//
// Запись — хеш (HSET) с ключом amphora:blacklist:<jti>:
//
//	user_id    — ид пользователя, чей токен отозван
//	login      — его логин
//	revoked_at — когда отозван (момент выхода)
//	expires_at — когда истекает сам токен
package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// blacklistPrefix — префикс ключей чёрного списка: «сервис:назначение:jti»,
// чтобы ключи разных приложений в одном Redis не пересекались.
const blacklistPrefix = "amphora:blacklist:"

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

func blacklistKey(jti string) string { return blacklistPrefix + jti }

// RevokedToken — запись чёрного списка: чей токен и когда отозван.
type RevokedToken struct {
	JTI       string
	UserID    uint
	Login     string
	ExpiresAt time.Time
}

// AddToBlacklist — выход: кладёт jti токена в чёрный список до конца срока
// действия токена. HSET + EXPIRE выполняются одной транзакцией (MULTI/EXEC).
func (c *Client) AddToBlacklist(ctx context.Context, t RevokedToken) error {
	ttl := time.Until(t.ExpiresAt)
	if ttl <= 0 {
		return nil // токен уже истёк — в чёрный список не нужен
	}
	key := blacklistKey(t.JTI)
	_, err := c.client.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.HSet(ctx, key, map[string]any{
			"user_id":    strconv.FormatUint(uint64(t.UserID), 10),
			"login":      t.Login,
			"revoked_at": time.Now().Format(time.RFC3339),
			"expires_at": t.ExpiresAt.Format(time.RFC3339),
		})
		pipe.Expire(ctx, key, ttl)
		return nil
	})
	return err
}

// IsBlacklisted — отозван ли токен (есть ли его jti в чёрном списке).
func (c *Client) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := c.client.Exists(ctx, blacklistKey(jti)).Result()
	return n > 0, err
}
