# 时态同步服务 Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app  
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o temporal-sync main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget
WORKDIR /root/

COPY --from=builder /app/temporal-sync .

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8092}/health || exit 1

# 暴露端口
EXPOSE 8092

# 启动命令
CMD ["./temporal-sync"]