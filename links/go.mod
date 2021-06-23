module github.com/demeero/pocket-link/links

go 1.16

require (
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/demeero/pocket-link/bricks v0.0.0-20210623205401-0a2609554d80
	github.com/demeero/pocket-link/proto/gen/go v0.0.0-20210623113856-58d68e5c85f4
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo/v4 v4.3.0
	go.mongodb.org/mongo-driver v1.5.3
	go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho v0.20.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	go.uber.org/zap v1.17.0
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)
