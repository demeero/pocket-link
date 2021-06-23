package key

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type GeneratorConfig struct {
	PredefinedKeysCount uint
	Delay               time.Duration
	KeyLen              uint8
}

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
				zap.L().Error("error get unused keys size", zap.Error(err))
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
		rndKey := randStringRunes(keyLen)
		existed, err := used.Exists(ctx, rndKey)
		if err != nil {
			zap.L().Error("error check used key existence", zap.Error(err))
			break
		}
		if existed {
			zap.L().Debug("key already exists in used keys repository - try another one", zap.String("key", rndKey))
			continue
		}
		stored, err := unused.Store(ctx, rndKey)
		if err != nil {
			zap.L().Error("error store new key", zap.Error(err))
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("-_1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
