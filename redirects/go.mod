module github.com/demeero/pocket-link/redirects

go 1.16

require (
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/demeero/pocket-link/bricks v0.0.0-20210623211941-2ffed6921ef5
	github.com/demeero/pocket-link/proto/gen/go v0.0.0-20210623125708-5227b6f1e5a1
	github.com/go-redis/redis/v8 v8.10.0
	github.com/go-redis/redismock/v8 v8.0.6
	github.com/golang/mock v1.6.0
	github.com/gomodule/redigo v1.8.5 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo/v4 v4.3.0
	github.com/stretchr/testify v1.7.0
	github.com/yuin/gopher-lua v0.0.0-20210529063254-f4c35e4016d9 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho v0.20.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	go.uber.org/zap v1.17.0
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)
