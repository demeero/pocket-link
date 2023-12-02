package main

import (
	"github.com/demeero/bricks/configbrick"
)

// config represents the configuration of application.
type config struct {
	configbrick.AppMeta
	Keygen   KeygenClient                  `json:"keygen"`
	OTEL     configbrick.OTEL              `json:"otel"`
	Log      configbrick.Log               `json:"log"`
	Mongo    configbrick.Mongo             `json:"mongo"`
	GRPC     configbrick.GRPC              `json:"grpc"`
	HTTP     configbrick.HTTP              `json:"http"`
	Profiler configbrick.PyroscopeProfiler `json:"profiler"`
}

// KeygenClient is a configuration for Keygen GRPC client.
type KeygenClient struct {
	// Addr is a target address for Keygen GRPC server (e.g. localhost:8081).
	Addr string `required:"true" json:"addr"`
}
