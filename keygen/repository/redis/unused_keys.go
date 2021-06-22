package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"

	"github.com/demeero/pocket-link/keygen/key"
)

const unusedSetName = "set_unusedkeys"

type UnusedKeys struct {
	rds redis.Cmdable
}

func NewUnusedKeys(rds redis.Cmdable) *UnusedKeys {
	return &UnusedKeys{
		rds: rds,
	}
}

func (u *UnusedKeys) LoadAndDelete(ctx context.Context) (string, error) {
	result, err := u.rds.SPop(ctx, unusedSetName).Result()
	if errors.Is(err, redis.Nil) {
		return "", key.ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	return result, nil
}

func (u *UnusedKeys) Store(ctx context.Context, k ...string) (int64, error) {
	return u.rds.SAdd(ctx, unusedSetName, k).Result()
}

func (u *UnusedKeys) Size(ctx context.Context) (int64, error) {
	return u.rds.SCard(ctx, unusedSetName).Result()
}
