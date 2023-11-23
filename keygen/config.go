package main

import (
	"time"

	"github.com/demeero/bricks/configbrick"
)

// Config represents the configuration of application.
type Config struct {
	configbrick.AppMeta
	UsedKeysRepositoryType UsedKeysRepositoryType `required:"true" split_words:"true" json:"used_keys_repository_type"`
	OTEL                   configbrick.OTEL       `json:"otel"`
	RedisUsedKeys          configbrick.Redis      `split_words:"true" json:"redis_used_keys"`
	RedisUnusedKeys        configbrick.Redis      `split_words:"true" json:"redis_unused_keys"`
	Log                    configbrick.Log        `json:"log"`
	MongoUsedKeys          configbrick.Mongo      `split_words:"true" json:"mongo_used_keys"`
	GRPC                   configbrick.GRPC       `json:"grpc"`
	Generator              Generator              `json:"generator"`
	Keys                   Keys                   `json:"keys"`
	ShutdownTimeout        time.Duration          `default:"10s" split_words:"true" json:"shutdown_timeout"`
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

type UsedKeysRepositoryType string

const (
	UsedKeysRepositoryTypeRedis UsedKeysRepositoryType = "redis"
	UsedKeysRepositoryTypeMongo UsedKeysRepositoryType = "mongo"
)
