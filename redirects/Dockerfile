FROM golang:1.20-alpine AS builder

WORKDIR /src/redirects
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/redirects /src/redirects/main.go

FROM alpine:3 AS runner
COPY --from=builder /bin/redirects /bin/redirects
ENTRYPOINT ["/bin/redirects"]
