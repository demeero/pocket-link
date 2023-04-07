package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

// Config represents the configuration of application.
type Config struct {
	Log                    Log                    `json:"log"`
	UsedKeysRepositoryType UsedKeysRepositoryType `required:"true" split_words:"true" json:"used_keys_repository_type"`
	MongoUsedKeys          MongoUsedKeys          `split_words:"true" json:"mongo_used_keys"`
	Telemetry              Telemetry              `json:"telemetry"`
	RedisUsedKeys          RedisUsedKeys          `split_words:"true" json:"redis_used_keys"`
	RedisUnusedKeys        RedisUnusedKeys        `split_words:"true" json:"redis_unused_keys"`
	Generator              Generator              `json:"generator"`
	GRPC                   GRPC                   `json:"grpc"`
	Keys                   Keys                   `json:"keys"`
	ShutdownTimeout        time.Duration          `default:"10s" split_words:"true" json:"shutdown_timeout"`
}

// New creates a new Config.
func New() Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal().Err(err).Msg("failed process config")
	}
	return cfg
}

// Log represents the log configuration.
type Log struct {
	// Level is the log level. "disabled" value disables logging.
	Level string `default:"debug" json:"log_level"`
	// Pretty enables human-friendly, colorized output instead of JSON.
	Pretty bool `json:"pretty"`
	// Caller adds file and line number to log.
	Caller bool `default:"true" json:"caller"`
	// UnixTimestamp enables unix timestamp in log instead of human-readable timestamps.
	UnixTimestamp bool `default:"true" json:"unix_timestamp"`
}

type Generator struct {
	// PredefinedKeysCount is a number of keys that should be generated in advance.
	PredefinedKeysCount uint `default:"100" split_words:"true" json:"predefined_keys_count"`
	// Delay is a delay between key generation.
	Delay time.Duration `default:"1m" split_words:"true" json:"delay"`
	// KeyLen is a length of generated keys.
	KeyLen uint8 `default:"10" split_words:"true" json:"key_len"`
}

// Keys is a configuration for Keys.
type Keys struct {
	// TTL is a time to live for used keys
	TTL time.Duration `default:"24h" json:"ttl"`
}

type GRPC struct {
	Port             int  `required:"true" json:"port"`
	EnableReflection bool `default:"true" split_words:"true" json:"enable_reflection"`
}

type UsedKeysRepositoryType string

const (
	UsedKeysRepositoryTypeRedis UsedKeysRepositoryType = "redis"
	UsedKeysRepositoryTypeMongo UsedKeysRepositoryType = "mongo"
)

type MongoUsedKeys struct {
	URI string `json:"uri"`
}

type RedisUsedKeys struct {
	Addr string `json:"addr"`
	DB   uint8  `json:"db"`
}

type RedisUnusedKeys struct {
	Addr string `required:"true" json:"addr"`
	DB   uint8  `json:"db"`
}

type Telemetry struct {
	Collector struct {
		Addr string `json:"addr"`
	} `json:"collector"`
}
