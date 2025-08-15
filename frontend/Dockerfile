# 前端服务生产级Dockerfile
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY package*.json ./
RUN npm ci --only=production --silent

# 复制源代码并构建
COPY . .
RUN npm run build

# 生产镜像
FROM nginx:1.25-alpine

# 复制构建结果
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制Nginx配置
COPY nginx.conf /etc/nginx/nginx.conf

# 暴露端口
EXPOSE 3000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=30s \
  CMD curl -f http://localhost:3000/ || exit 1

# 启动命令
CMD ["nginx", "-g", "daemon off;"]