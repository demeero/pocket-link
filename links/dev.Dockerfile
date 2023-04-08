FROM golang:1.20-alpine

RUN go install github.com/cosmtrek/air@v1.42.0

WORKDIR /workdir

CMD ["air", "-c", ".air.toml"]
