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
			for i := 0; i < int(n); {
				rndKey := randStringRunes(cfg.KeyLen)
				existed, err := used.Exists(ctx, rndKey)
				if err != nil {
					zap.L().Error("error check used key existence", zap.Error(err))
					break
				}
				if existed {
					continue
				}
				stored, err := unused.Store(ctx, rndKey)
				if err != nil {
					zap.L().Error("error store new key", zap.Error(err))
					break
				}
				if stored == 0 {
					continue
				}
				i++
			}
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("-_1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n uint8) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
