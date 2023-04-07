package key

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

var letterRunes = []rune("-_1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// GeneratorConfig is a configuration for key generator.
type GeneratorConfig struct {
	// PredefinedKeysCount is a number of keys that should be generated in advance.
	PredefinedKeysCount uint
	// Delay is a delay between key generation.
	Delay time.Duration
	// KeyLen is a length of generated keys.
	KeyLen uint8
}

// Generate generates keys by specified GeneratorConfig.
func Generate(ctx context.Context, cfg GeneratorConfig, used UsedKeysRepository, unused UnusedKeysRepository) {
	t := time.NewTicker(cfg.Delay)
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
			size, err := unused.Size(ctx)
			if err != nil {
				zap.L().Error("failed get unused keys size", zap.Error(err))
				continue
			}
			if size > int64(cfg.PredefinedKeysCount) {
				continue
			}
			n := cfg.PredefinedKeysCount/10 + 1
			gen(ctx, int(n), int(cfg.KeyLen), used, unused)
		}
	}
}

func gen(ctx context.Context, n, keyLen int, used UsedKeysRepository, unused UnusedKeysRepository) {
	for i := 0; i < n; {
		rndKey := randStringRunes(keyLen, rand.New(rand.NewSource(time.Now().UnixNano())))
		existed, err := used.Exists(ctx, rndKey)
		if err != nil {
			zap.L().Error("failed check used key existence", zap.Error(err))
			break
		}
		if existed {
			zap.L().Debug("key already exists in used keys repository - try another one", zap.String("key", rndKey))
			continue
		}
		stored, err := unused.Store(ctx, rndKey)
		if err != nil {
			zap.L().Error("failed store new key", zap.Error(err))
			break
		}
		if stored == 0 {
			zap.L().Debug("key already exists in unused keys repository - try another one", zap.String("key", rndKey))
			continue
		}
		zap.L().Debug("stored new key", zap.String("key", rndKey))
		i++
	}
}

func randStringRunes(n int, rnd *rand.Rand) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rnd.Intn(len(letterRunes))]
	}
	return string(b)
}
