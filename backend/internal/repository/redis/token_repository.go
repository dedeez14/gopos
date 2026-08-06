// Package redis berisi implementasi kontrak TokenRepository di atas Redis.
// Refresh token disimpan sebagai kunci "refresh:<token>" → userID dengan TTL
// — kedaluwarsa otomatis oleh Redis, dan pencabutan = DEL satu kunci.
package redis

import (
	"context"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/tuleh-pos/server/internal/domain"
)

type TokenRepository struct {
	client *goredis.Client
}

func NewTokenRepository(client *goredis.Client) *TokenRepository {
	return &TokenRepository{client: client}
}

func kunci(token string) string { return "refresh:" + token }

func (r *TokenRepository) SimpanRefresh(ctx context.Context, token string, userID uint, ttl time.Duration) error {
	return r.client.Set(ctx, kunci(token), userID, ttl).Err()
}

func (r *TokenRepository) AmbilRefresh(ctx context.Context, token string) (uint, error) {
	val, err := r.client.Get(ctx, kunci(token)).Result()
	if err == goredis.Nil {
		return 0, domain.ErrTokenTidakSah
	}
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, domain.ErrTokenTidakSah
	}
	return uint(id), nil
}

func (r *TokenRepository) HapusRefresh(ctx context.Context, token string) error {
	return r.client.Del(ctx, kunci(token)).Err()
}
