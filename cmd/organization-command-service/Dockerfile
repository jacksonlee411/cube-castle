# 命令服务生产级Dockerfile (带健康检查)
FROM golang:1.22-alpine AS builder

# 安装依赖和curl用于健康检查
RUN apk add --no-cache git ca-certificates tzdata curl

# 设置工作目录
WORKDIR /app

# 复制所有源代码（包括依赖）
COPY . .
COPY ../../pkg ./pkg 2>/dev/null || echo "No pkg directory"

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main . || go mod tidy && go build -o main .

# 最终运行镜像
FROM alpine:latest

# 安装必要工具用于健康检查
RUN apk --no-cache add ca-certificates tzdata curl jq

# 复制可执行文件
COPY --from=builder /app/main /main

# 创建健康检查脚本
RUN echo '#!/bin/sh' > /healthcheck.sh && \
    echo 'curl -f http://localhost:9090/health | jq -r .status | grep -E "(healthy|degraded)" > /dev/null' >> /healthcheck.sh && \
    chmod +x /healthcheck.sh

# 暴露端口
EXPOSE 9090

# 健康检查配置
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD /healthcheck.sh

# 启动命令
ENTRYPOINT ["/main"]