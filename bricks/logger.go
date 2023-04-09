package bricks

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogConfig represents the log configuration.
type LogConfig struct {
	// Level is the log level. "disabled" value disables logging.
	Level string `default:"debug" json:"log_level"`
	// Pretty enables human-friendly, colorized output instead of JSON.
	Pretty bool `json:"pretty"`
	// Caller adds file and line number to log.
	Caller bool `default:"true" json:"caller"`
	// UnixTimestamp enables unix timestamp in log instead of human-readable timestamps.
	UnixTimestamp bool `default:"true" json:"unix_timestamp"`
}

func ConfigureLogger(cfg LogConfig) {
	if cfg.UnixTimestamp {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
	if cfg.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}
	if cfg.Caller {
		log.Logger = log.Logger.With().Caller().Logger()
	}
	zerolog.DefaultContextLogger = &log.Logger
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		log.Fatal().Err(err).Msg("failed parse log level")
	}
	zerolog.SetGlobalLevel(level)
}
