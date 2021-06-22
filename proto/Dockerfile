FROM alpine:latest as base

RUN apk add --no-cache curl go protoc git

# Install the protocol compiler plugins for Go
RUN GO111MODULE=on go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.26 \
        google.golang.org/grpc/cmd/protoc-gen-go-grpc

# Install buf.build
RUN VERSION=0.41.0 && \
    curl -sSL https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m) \
        -o /usr/local/bin/buf && \
    chmod +x /usr/local/bin/buf

WORKDIR proto
COPY . .

FROM base as lint
RUN export PATH=$PATH:$(go env GOPATH)/bin && buf lint

FROM base as gen
RUN export PATH=$PATH:$(go env GOPATH)/bin && buf generate
RUN cd gen/go && go mod init github.com/demeero/pocket-link/proto/gen/go && go mod tidy