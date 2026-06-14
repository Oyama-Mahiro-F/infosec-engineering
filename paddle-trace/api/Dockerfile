# =============================================================================
# 乒乓球拍防伪溯源系统 - API服务Dockerfile
# 多阶段构建：编译阶段 + 运行阶段
# =============================================================================

# ---- 阶段1：编译 ----
FROM golang:1.21-alpine AS builder

# 安装编译依赖
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译（静态链接，减小镜像体积）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w -X main.version=1.0.0" \
    -o /app/paddle-trace-api ./cmd/main.go

# ---- 阶段2：运行 ----
FROM alpine:3.19

# 安装CA证书（HTTPS请求需要）
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从编译阶段复制二进制文件
COPY --from=builder /app/paddle-trace-api .

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# 启动服务
ENTRYPOINT ["./paddle-trace-api"]
