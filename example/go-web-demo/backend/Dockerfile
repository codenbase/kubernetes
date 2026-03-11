# 阶段1: 编译构建
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置 Go 代理，加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct

# 拷贝 go.mod
COPY go.mod ./
# 注意：暂时没有 go.sum，因为我们还没有执行 go mod tidy，如果之后添加了，也可以把 go.sum 复制过来
# 运行 download 如果失败会自动跳过缺失的依赖，但在下一步 go build 也会下载
RUN go mod download || true

# 拷贝源代码
COPY . .

# 触发 go mod tidy 生成完整的 go.sum（确保本地编译一致性）
RUN go mod tidy

# 静态编译，禁用 CGO 以确保alpine/scratch运行正常
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o server cmd/main.go

# 阶段2: 运行镜像
FROM alpine:latest

# 安装根证书和时区数据库
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从 builder 拷贝可执行文件
COPY --from=builder /app/server .

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./server"]
