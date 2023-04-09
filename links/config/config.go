package config

import (
	"time"

	"github.com/demeero/pocket-link/bricks"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

// Config represents the configuration of application.
type Config struct {
	Telemetry       Telemetry        `json:"telemetry"`
	Mongo           Mongo            `json:"mongo"`
	Keygen          KeygenClient     `json:"keygen"`
	Log             bricks.LogConfig `json:"log"`
	HTTP            HTTP             `json:"http"`
	GRPC            GRPC             `json:"grpc"`
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

// GRPC is a configuration for GRPC server.
type GRPC struct {
	// Port is the port on which the server will listen.
	Port int `required:"true" json:"port"`
	// EnableReflection enables gRPC reflection.
	EnableReflection bool `default:"true" split_words:"true" json:"enable_reflection"`
}

// Mongo is a configuration for MongoDB.
type Mongo struct {
	// URI is a connection string for MongoDB (e.g. mongodb://localhost:27017).
	URI string `required:"true" json:"uri"`
}

// KeygenClient is a configuration for Keygen GRPC client.
type KeygenClient struct {
	// Addr is a target address for Keygen GRPC server (e.g. localhost:8081).
	Addr string `required:"true" json:"addr"`
}

// Telemetry is a configuration for telemetry.
type Telemetry struct {
	Collector struct {
		Addr string `json:"addr"`
	} `json:"collector"`
}
