FROM golang:1.20-alpine AS builder

WORKDIR /src/keygen
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/keygen /src/keygen/main.go

FROM alpine:3 AS runner
COPY --from=builder /bin/keygen /bin/keygen
ENTRYPOINT ["/bin/keygen"]
