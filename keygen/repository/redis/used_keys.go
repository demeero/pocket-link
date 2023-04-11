package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type UsedKeys struct {
	rds redis.Cmdable
}

func NewUsedKeys(rds redis.Cmdable) *UsedKeys {
	return &UsedKeys{
		rds: rds,
	}
}

func (u *UsedKeys) Store(ctx context.Context, k string, ttl time.Duration) (bool, error) {
	return u.rds.SetNX(ctx, k, nil, ttl).Result()
}

func (u *UsedKeys) Exists(ctx context.Context, k string) (bool, error) {
	result, err := u.rds.Exists(ctx, k).Result()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}
