# 多阶段构建
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o chatroom main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o client client/client.go

# 运行阶段
FROM alpine:latest

# 安装ca-certificates
RUN apk --no-cache add ca-certificates

# 创建非root用户
RUN addgroup -g 1001 -S chatroom && \
    adduser -u 1001 -S chatroom -G chatroom

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/chatroom .
COPY --from=builder /app/client .

# 创建日志目录
RUN mkdir -p /app/logs && chown -R chatroom:chatroom /app

# 切换到非root用户
USER chatroom

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 设置环境变量
ENV CHATROOM_HOST=0.0.0.0
ENV CHATROOM_PORT=8080
ENV CHATROOM_MAX_USERS=100
ENV CHATROOM_TIMEOUT=40

# 启动命令
CMD ["./chatroom"]
