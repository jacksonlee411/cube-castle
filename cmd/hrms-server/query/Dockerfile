# GraphQL 查询服务 Dockerfile（PostgreSQL 原生架构）
FROM golang:1.24-alpine AS builder

# 配置 Go 模块代理（国内镜像）
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates tzdata curl

# 复制全部源码（包括依赖的 pkg 目录）
WORKDIR /workspace
COPY . .

# 下载依赖并构建
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /query-service ./cmd/hrms-server/query

FROM golang:1.24-alpine AS dev
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GOTOOLCHAIN=auto
RUN apk add --no-cache git ca-certificates tzdata curl bash
WORKDIR /workspace
COPY . .
WORKDIR /workspace/cmd/hrms-server/query
RUN go mod download
RUN go install github.com/cosmtrek/air@v1.52.3
ENTRYPOINT ["air","-c","cmd/hrms-server/query/.air.toml"]

FROM alpine:latest AS release

RUN apk --no-cache add ca-certificates tzdata curl jq

WORKDIR /app

# 复制可执行文件与 GraphQL Schema（运行时按 docs/api/schema.graphql 读取）
COPY --from=builder /query-service /query-service
RUN mkdir -p docs/api
COPY --from=builder /workspace/docs/api/schema.graphql docs/api/schema.graphql

EXPOSE 8090

# 健康检查：命令服务会依赖查询服务健康状态
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -fs http://localhost:8090/health | jq -r .status | grep -E "(healthy|degraded)"

ENTRYPOINT ["/query-service"]
