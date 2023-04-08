FROM golang:1.20-alpine AS builder

WORKDIR /src/links
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/links /src/links/main.go

FROM alpine:3 AS runner
COPY --from=builder /bin/links /bin/links
ENTRYPOINT ["/bin/links"]
