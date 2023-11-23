package main

import (
	"github.com/demeero/bricks/configbrick"
)

type config struct {
	configbrick.AppMeta
	Links    linksClient       `json:"links"`
	OTEL     configbrick.OTEL  `json:"otel"`
	RedisLRU configbrick.Redis `json:"redis_lru" split_words:"true"`
	Log      configbrick.Log   `json:"log"`
	HTTP     configbrick.HTTP  `json:"http"`
}

type linksClient struct {
	// Addr is a target address for Keygen GRPC server (e.g. localhost:8081).
	Addr string `required:"true" json:"addr"`
}
