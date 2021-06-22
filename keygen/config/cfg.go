package config

import "github.com/demeero/pocket-link/keygen/key"

type Config struct {
	GRPC            GRPC
	RedisUsedKeys   RedisUsedKeys
	RedisUnusedKeys RedisUnusedKeys
	Keys            key.KeysConfig
	Generator       key.GeneratorConfig
}

type GRPC struct {
	Port int
}

type RedisUsedKeys struct {
	Addr string
	DB   uint8
}

type RedisUnusedKeys struct {
	Addr string
	DB   uint8
}
