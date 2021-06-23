package config

type Config struct {
	HTTP   HTTP
	Mongo  Mongo
	Keygen KeygenClient
}

type HTTP struct {
	Port int
}

type Mongo struct {
	URI string // e.g. mongodb://localhost:27017
}

type KeygenClient struct {
	Addr string
}
