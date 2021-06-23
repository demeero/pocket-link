package config

import "github.com/demeero/pocket-link/keygen/key"

type Config struct {
	GRPC                   GRPC
	UsedKeysRepositoryType UsedKeysRepositoryType
	RedisUsedKeys          RedisUsedKeys
	MongoUsedKeys          MongoUsedKeys
	RedisUnusedKeys        RedisUnusedKeys
	Keys                   key.KeysConfig
	Generator              key.GeneratorConfig
	Telemetry              Telemetry
}

type GRPC struct {
	Port int
}

type UsedKeysRepositoryType string

const (
	UsedKeysRepositoryTypeRedis UsedKeysRepositoryType = "redis"
	UsedKeysRepositoryTypeMongo UsedKeysRepositoryType = "mongo"
)

type MongoUsedKeys struct {
	URI string // e.g. mongodb://localhost:27017
}

type RedisUsedKeys struct {
	Addr string
	DB   uint8
}

type RedisUnusedKeys struct {
	Addr string
	DB   uint8
}

type Telemetry struct {
	Collector struct {
		Addr string
	}
}
