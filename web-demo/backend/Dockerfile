# ==========================================
# 阶段1: 编译构建
# ==========================================
ARG REGISTRY_PREFIX=""
FROM ${REGISTRY_PREFIX}golang:1.25-alpine AS builder

WORKDIR /build

# 设置 Go 代理，加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct

# 先只拷贝依赖描述文件，利用 Docker 缓存机制，把依赖包先缓存起来
COPY go.mod go.sum ./
# 仅下载依赖
RUN go mod download

# 拷贝业务源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o server cmd/main.go

# ==========================================
# 阶段2: 运行镜像
# ==========================================
# 彻底锁死：静态底座 + Debian 12 + 非特权用户
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# 从 builder 拷贝可执行文件
COPY --chown=nonroot:nonroot --from=builder /build/server .

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./server"]