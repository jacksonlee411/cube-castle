# 命令服务生产级Dockerfile (带健康检查)
FROM golang:1.24-alpine AS builder

# 安装依赖和curl用于健康检查
RUN apk add --no-cache git ca-certificates tzdata curl

# 允许 Go 自动下载所需 toolchain
ENV GOTOOLCHAIN=auto
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct

# 设置工作目录并复制源码
WORKDIR /workspace
COPY . .

# 下载依赖并构建可执行文件
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /main ./cmd/hrms-server/command

FROM golang:1.24-alpine AS dev
ENV GOTOOLCHAIN=auto
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
RUN apk add --no-cache git ca-certificates tzdata curl bash
WORKDIR /workspace
COPY . .
WORKDIR /workspace/cmd/hrms-server/command
RUN go mod download
RUN go install github.com/cosmtrek/air@v1.52.3
ENTRYPOINT ["air","-c","cmd/hrms-server/command/.air.toml"]

# 最终运行镜像
FROM alpine:latest AS release

# 安装必要工具用于健康检查
RUN apk --no-cache add ca-certificates tzdata curl jq

# 复制可执行文件
COPY --from=builder /main /main

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
