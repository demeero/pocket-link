module github.com/demeero/pocket-link/keygen

go 1.16

require (
	github.com/avast/retry-go/v3 v3.1.1
	github.com/demeero/pocket-link/bricks v0.0.0-20210623194847-f53aadbd5048
	github.com/demeero/pocket-link/proto/gen/go v0.0.0-20210622173658-fb5b34dca4e2
	github.com/go-redis/redis/v8 v8.10.0
	github.com/golang/snappy v0.0.3 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.7.0 // indirect
	go.mongodb.org/mongo-driver v1.5.3
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	go.uber.org/zap v1.17.0
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)
