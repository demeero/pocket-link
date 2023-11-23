FROM golang:1.21-alpine AS builder

WORKDIR /usr/local/src/redirects
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /usr/local/bin/redirects ./*.go

FROM alpine:3 AS runner
COPY --from=builder /usr/local/bin/redirects /usr/local/bin/redirects
ENTRYPOINT ["/usr/local/bin/redirects"]


