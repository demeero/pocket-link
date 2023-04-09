package config

import (
	"time"

	"github.com/demeero/pocket-link/bricks"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Links           LinksClient      `json:"links"`
	Telemetry       Telemetry        `json:"telemetry"`
	Log             bricks.LogConfig `json:"log"`
	RedisLRU        RedisLRU         `json:"redis_lru" split_words:"true"`
	HTTP            HTTP             `json:"http"`
	ShutdownTimeout time.Duration    `default:"10s" split_words:"true" json:"shutdown_timeout"`
}

// New creates a new Config.
func New() Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal().Err(err).Msg("failed process config")
	}
	return cfg
}

// HTTP is a configuration for HTTP server.
type HTTP struct {
	// Port is the port on which the server will listen.
	Port int `required:"true" json:"port"`
	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	ReadTimeout time.Duration `default:"30s" split_words:"true" json:"read_timeout"`
	// ReadHeaderTimeout is the amount of time allowed to read request headers.
	ReadHeaderTimeout time.Duration `default:"10s" split_words:"true" json:"read_header_timeout"`
	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration `default:"30s" split_words:"true" json:"write_timeout"`
}

type LinksClient struct {
	// Addr is a target address for Keygen GRPC server (e.g. localhost:8081).
	Addr string `required:"true" json:"addr"`
}

// RedisLRU is a configuration for Redis LRU to cache links.
type RedisLRU struct {
	Addr string `json:"addr"`
	DB   uint8  `json:"db"`
}

// Telemetry is a configuration for telemetry.
type Telemetry struct {
	Collector struct {
		Addr string `json:"addr"`
	} `json:"collector"`
}
