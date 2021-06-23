FROM golang:1.16.2 AS base

ARG ssh_key

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPRIVATE=github.com/demeero/pocket-link

RUN apt-get update && apt-get install -y ca-certificates git-core ssh

# Authorize SSH Host
RUN mkdir -p /root/.ssh && \
    chmod 0700 /root/.ssh && \
    ssh-keyscan github.com > /root/.ssh/known_hosts

# Add the keys and set permissions (not secure because ssh key remains in image if image is not deleted)
RUN echo "$ssh_key" > /root/.ssh/id_rsa && chmod 600 /root/.ssh/id_rsa && \
    echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config && \
    git config --global url.ssh://git@github.com/.insteadOf https://github.com/

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