package key

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avast/retry-go/v3"
	"github.com/demeero/bricks/errbrick"
	"github.com/rs/zerolog/log"
)

// Key is a key for short link.
type Key struct {
	ExpiresAt time.Time
	Val       string
}

// UnusedKeysRepository is a repository for unused keys.
//
//go:generate mockgen -destination=unused_keys_mock.go -package=key github.com/demeero/pocket-link/keygen/key UnusedKeysRepository
type UnusedKeysRepository interface {
	LoadAndDelete(context.Context) (string, error)
	Store(context.Context, ...string) (int64, error)
	Size(ctx context.Context) (int64, error)
}

// UsedKeysRepository is a repository for used keys.
//
//go:generate mockgen -destination=used_keys_mock.go -package=key github.com/demeero/pocket-link/keygen/key UsedKeysRepository
type UsedKeysRepository interface {
	Store(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Exists(context.Context, string) (bool, error)
}

// Keys is a service for generating keys.
type Keys struct {
	used   UsedKeysRepository
	unused UnusedKeysRepository
	ttl    time.Duration
}

// New creates a new Keys.
func New(ttl time.Duration, used UsedKeysRepository, unused UnusedKeysRepository) *Keys {
	return &Keys{
		ttl:    ttl,
		used:   used,
		unused: unused,
	}
}

// Use returns a key for short link.
func (k *Keys) Use(ctx context.Context) (Key, error) {
	var result Key

	job := func() error {
		loadedKey, err := k.unused.LoadAndDelete(ctx)
		if err != nil {
			return fmt.Errorf("failed load key: %w", err)
		}

		expiresAt := time.Now().Add(k.ttl)
		stored, err := k.used.Store(ctx, loadedKey, k.ttl)
		if err != nil {
			return fmt.Errorf("failed store key: %w", err)
		}
		if !stored {
			return fmt.Errorf("%w: key already exist", errbrick.ErrConflict)
		}

		result.ExpiresAt = expiresAt
		result.Val = loadedKey
		return nil
	}

	retryCond := func(err error) bool {
		if errors.Is(err, errbrick.ErrConflict) {
			log.Ctx(ctx).Info().Msg("expected free key is already in use - retry")
			return true
		}
		if errors.Is(err, errbrick.ErrNotFound) {
			log.Ctx(ctx).Info().Msg("no free keys - retry")
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
