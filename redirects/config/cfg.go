package config

type Config struct {
	HTTP  HTTP
	Links LinksClient
}

type HTTP struct {
	Port int
}

type LinksClient struct {
	Addr string
}
