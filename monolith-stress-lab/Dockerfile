# Start from golang alpine image to build the Go binary
FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Prepare the final Alpine image
FROM alpine:3.19

# Change to aliyun mirror for faster apk downloads
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && apk add --no-cache \
    postgresql \
    postgresql-contrib \
    sysstat \
    procps \
    curl \
    vim \
    bash \
    su-exec

WORKDIR /app

# Ensure directories exist and have correct permissions
RUN mkdir -p /var/run/postgresql && chown -R postgres:postgres /var/run/postgresql && \
    mkdir -p /var/lib/postgresql/data && chown -R postgres:postgres /var/lib/postgresql

# Copy the Pre-built binary file and scripts
COPY --from=builder /app/main .
COPY db/init.sql ./init.sql
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

EXPOSE 8080

CMD ["./entrypoint.sh"]
