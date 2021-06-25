FROM golang:1.16.2 AS base

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPRIVATE=github.com/demeero/pocket-link

WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

FROM base AS build
RUN go build -o /out/app main.go

FROM alpine:3.13.3 AS bin
ENV GRPC_PORT=8081
COPY --from=build /out/app /
EXPOSE ${GRPC_PORT}
ENTRYPOINT ["/app"]