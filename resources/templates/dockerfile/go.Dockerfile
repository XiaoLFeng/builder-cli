# xbuilder Dockerfile 示例 - Go
# 多阶段构建，适用于 Go 应用

# ============================================
# 阶段 1: 构建
# ============================================
FROM golang:1.23-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum (利用 Docker 缓存)
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建参数
ARG VERSION=dev
ARG BUILD_DATE
ARG GIT_COMMIT

# 编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
    -X main.Version=${VERSION} \
    -X main.BuildDate=${BUILD_DATE} \
    -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/main .

# ============================================
# 阶段 2: 运行
# ============================================
FROM scratch

LABEL maintainer="your-email@example.com"

# 复制时区信息
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
# 复制 CA 证书
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 设置时区
ENV TZ=Asia/Shanghai

# 复制二进制文件
COPY --from=builder /app/main /app/main

# 暴露端口
EXPOSE 8080

# 启动命令
ENTRYPOINT ["/app/main"]
