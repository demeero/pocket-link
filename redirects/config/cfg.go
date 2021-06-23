package config

type Config struct {
	HTTP     HTTP
	Links    LinksClient
	RedisLRU RedisLRU
}

type HTTP struct {
	Port int
}

type LinksClient struct {
	Addr string
}

type RedisLRU struct {
	Addr string
}
