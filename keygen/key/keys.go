package key

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go/v3"
	"go.uber.org/zap"

	"github.com/demeero/pocket-link/bricks/zaplogger"
)

type Key struct {
	Val       string
	ExpiresAt time.Time
}

var ErrKeyAlreadyUsed = errors.New("key already used")
var ErrKeyNotFound = errors.New("key not found")

//go:generate mockgen -destination=unused_keys_mock.go -package=key github.com/demeero/pocket-link/keygen/key UnusedKeysRepository
type UnusedKeysRepository interface {
	LoadAndDelete(context.Context) (string, error)
	Store(context.Context, ...string) (int64, error)
	Size(ctx context.Context) (int64, error)
}

//go:generate mockgen -destination=used_keys_mock.go -package=key github.com/demeero/pocket-link/keygen/key UsedKeysRepository
type UsedKeysRepository interface {
	Store(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Exists(context.Context, string) (bool, error)
}

type KeysConfig struct {
	TTL time.Duration
}

type Keys struct {
	cfg    KeysConfig
	used   UsedKeysRepository
	unused UnusedKeysRepository
}

func New(cfg KeysConfig, used UsedKeysRepository, unused UnusedKeysRepository) *Keys {
	return &Keys{
		cfg:    cfg,
		used:   used,
		unused: unused,
	}
}

func (k *Keys) Use(ctx context.Context) (Key, error) {
	var result Key

	job := func() error {
		loadedKey, err := k.unused.LoadAndDelete(ctx)
		if err != nil {
			return err
		}

		expiresAt := time.Now().Add(k.cfg.TTL)
		stored, err := k.used.Store(ctx, loadedKey, k.cfg.TTL)
		if err != nil {
			return err
		}
		if !stored {
			return ErrKeyAlreadyUsed
		}

		result.ExpiresAt = expiresAt
		result.Val = loadedKey
		return nil
	}

	retryCond := func(err error) bool {
		if errors.Is(err, ErrKeyAlreadyUsed) {
			zaplogger.From(ctx).Warn("expected free key is already in use - retry", zap.Error(err))
			return true
		}
		if errors.Is(err, ErrKeyNotFound) {
			zaplogger.From(ctx).Warn("no free keys - retry", zap.Error(err))
			return true
		}
		return false
	}

	err := retry.Do(job,
		retry.RetryIf(retryCond),
		retry.Context(ctx),
		retry.LastErrorOnly(true),
	)
	if err != nil {
		return Key{}, err
	}
	return result, nil
}
