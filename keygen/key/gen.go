package key

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/rs/zerolog/log"
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
				log.Error().Err(err).Msg("failed get unused keys size")
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
		rndKey, err := randStringRunes(keyLen)
		if err != nil {
			log.Error().Err(err).Msg("failed get random string")
			break
		}
		existed, err := used.Exists(ctx, rndKey)
		if err != nil {
			log.Error().Err(err).Msg("failed check used key existence")
			break
		}
		if existed {
			log.Debug().Msgf("key %s already exists in used keys repository - try another one", rndKey)
			continue
		}
		stored, err := unused.Store(ctx, rndKey)
		if err != nil {
			log.Error().Err(err).Msg("failed store new key")
			break
		}
		if stored == 0 {
			log.Debug().Msgf("key %s already exists in unused keys repository - try another one", rndKey)
			continue
		}
		log.Debug().Msgf("stored new key %s", rndKey)
		i++
	}
}

func randStringRunes(n int) (string, error) {
	b := make([]rune, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))
		if err != nil {
			return "", fmt.Errorf("failed pick random letter: %w", err)
		}
		b[i] = letterRunes[idx.Int64()]
	}
	return string(b), nil
}
