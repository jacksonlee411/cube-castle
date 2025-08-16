# 🚀 Cube Castle 部署指南

> **版本**: v2.1.0 | **更新日期**: 2025年8月6日  
> **适用环境**: 开发、测试、生产环境  
> **架构状态**: Vite + Canvas Kit 现代化前端架构

## 📋 概述

本文档提供 Cube Castle 企业级 HR SaaS 平台的完整部署指南，涵盖从开发环境到生产环境的各种部署场景，特别针对最新的 Vite + Canvas Kit 前端架构进行了优化。

### 🎯 部署架构概览

```yaml
架构组件:
  前端: Vite + React 18 + Canvas Kit (端口: 3000)
  后端: Go 1.23+ API服务 (端口: 8080)
  AI服务: Python gRPC (端口: 50051)
  数据库: PostgreSQL 16+ + Neo4j 5+
  缓存: Redis 7.x
  工作流: Temporal 1.25+
  消息队列: Kafka + Zookeeper (企业级)
  监控: Prometheus + Grafana
```

## 🛠️ 环境要求

### 基础环境依赖

#### 开发环境
```yaml
必需组件:
  - Node.js 18+ (前端Vite构建)
  - Go 1.23+ (后端服务)
  - Python 3.12+ (AI服务)
  - Docker 24+ & Docker Compose v2
  - Git 2.40+

数据库:
  - PostgreSQL 16+ (主数据存储)
  - Neo4j 5+ (图数据库)
  - Redis 7.x (缓存和会话)

系统资源:
  - 内存: 最少8GB，推荐16GB
  - CPU: 最少4核，推荐8核
  - 存储: 最少50GB可用空间
  - 网络: 千兆网络连接
```

#### 生产环境
```yaml
硬件配置:
  - 内存: 32GB+ (推荐64GB)
  - CPU: 16核+ (推荐32核)
  - 存储: SSD 500GB+ (推荐1TB)
  - 网络: 万兆网络连接

云服务推荐:
  - AWS: c6i.4xlarge或更高
  - Azure: Standard_D16s_v3或更高
  - GCP: c2-standard-16或更高
  - 阿里云: ecs.c7.4xlarge或更高
```

## 🏗️ 开发环境部署

### 1. 项目克隆和初始化

```bash
# 克隆项目
git clone <repository-url>
cd cube-castle

# 验证分支状态
git status
git branch -a
```

### 2. 前端开发环境 🆕

```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 验证Canvas Kit依赖
npm list @workday/canvas-kit-react

# 启动开发服务器
npm run dev

# 验证服务状态
curl http://localhost:3000/
```

#### 前端环境变量配置
```bash
# frontend/.env.development
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
VITE_APP_TITLE=Cube Castle - 开发环境
VITE_APP_VERSION=2.1.0

# 开发环境特定配置
VITE_DEV_TOOLS=true
VITE_HOT_RELOAD=true
VITE_SOURCE_MAPS=true
```

### 3. 后端服务配置

```bash
# 进入Go后端目录
cd go-app

# 配置环境变量
cp .env.example .env

# 编辑环境配置
vim .env
```

#### 关键环境变量配置
```bash
# .env 开发环境配置
# 服务配置
APP_ENV=development
APP_PORT=8080
LOG_LEVEL=debug

# 数据库配置
DATABASE_URL=postgresql://postgres:password@localhost:5432/cubecastle?sslmode=disable
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password

# Redis配置
REDIS_URL=redis://localhost:6379

# AI服务配置
INTELLIGENCE_SERVICE_GRPC_TARGET=localhost:50051

# 安全配置
JWT_SECRET=your-super-secret-jwt-key-for-development
CORS_ORIGINS=http://localhost:3000

# 前端集成配置 🆕
FRONTEND_URL=http://localhost:3000
STATIC_FILE_PATH=../frontend/dist

# Temporal工作流配置
TEMPORAL_HOST_PORT=localhost:7233
TEMPORAL_NAMESPACE=cube-castle-dev
```

### 4. 基础设施启动

#### 使用Docker Compose (推荐)
```bash
# 启动所有基础服务
docker-compose up -d postgres neo4j redis

# 验证服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f postgres neo4j redis
```

#### 手动服务启动 (备选)
```bash
# PostgreSQL
docker run --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=cubecastle \
  -p 5432:5432 -d postgres:16

# Neo4j
docker run --name neo4j \
  -e NEO4J_AUTH=neo4j/password \
  -p 7474:7474 -p 7687:7687 \
  -d neo4j:5-community

# Redis
docker run --name redis \
  -p 6379:6379 -d redis:7-alpine
```

### 5. 数据库初始化

```bash
# 等待服务就绪
./scripts/wait-for-services.sh

# 初始化数据库
cd go-app
go run cmd/server/main.go --init-db

# 验证数据库连接
psql -h localhost -U postgres -d cubecastle -c "SELECT version();"

# 验证Neo4j连接
curl -u neo4j:password http://localhost:7474/db/data/
```

### 6. 启动应用服务

```bash
# 方案一：并行启动 (推荐)
cd frontend && npm run dev &
cd go-app && go run cmd/server/main.go &
cd python-ai && python main.py &

# 方案二：使用tmux会话管理
tmux new-session -d -s cube-castle
tmux send-keys -t cube-castle:0 'cd frontend && npm run dev' C-m
tmux split-window -t cube-castle:0
tmux send-keys -t cube-castle:1 'cd go-app && go run cmd/server/main.go' C-m
tmux split-window -t cube-castle:1
tmux send-keys -t cube-castle:2 'cd python-ai && python main.py' C-m
```

### 7. 验证开发环境

```bash
# 前端验证
curl http://localhost:3000
# 期望: Vite React应用加载页面

# 后端API验证
curl http://localhost:8080/health
# 期望: {"status": "ok", "timestamp": "..."}

# 组织架构API验证 (重构后的核心功能)
curl http://localhost:8080/api/v1/organizations
# 期望: 组织数据列表

# AI服务验证
curl -X POST http://localhost:8080/api/v1/intelligence/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello"}'
# 期望: AI响应消息
```

## 🔧 测试环境部署

### 1. 测试环境配置

```bash
# 创建测试环境配置
cp .env .env.test

# 修改测试环境参数
sed -i 's/development/test/g' .env.test
sed -i 's/cubecastle/cubecastle_test/g' .env.test
sed -i 's/3000/3001/g' frontend/.env.test
sed -i 's/8080/8081/g' .env.test
```

### 2. Docker Compose测试环境

```yaml
# docker-compose.test.yml
version: '3.8'
services:
  frontend-test:
    build:
      context: ./frontend
      dockerfile: Dockerfile.test
    ports:
      - "3001:3000"
    environment:
      - VITE_API_BASE_URL=http://backend-test:8081/api/v1
    depends_on:
      - backend-test

  backend-test:
    build:
      context: ./go-app
      dockerfile: Dockerfile
    ports:
      - "8081:8080"
    environment:
      - APP_ENV=test
      - DATABASE_URL=postgresql://postgres:password@postgres-test:5432/cubecastle_test
    depends_on:
      - postgres-test
      - redis-test

  postgres-test:
    image: postgres:16
    environment:
      POSTGRES_DB: cubecastle_test
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5433:5432"

  redis-test:
    image: redis:7-alpine
    ports:
      - "6380:6379"
```

### 3. 自动化测试执行

```bash
# 启动测试环境
docker-compose -f docker-compose.test.yml up -d

# 前端测试
cd frontend
npm run test              # 单元测试
npm run test:e2e         # Playwright端到端测试
npm run test:e2e:ui      # 可视化E2E测试

# 后端测试
cd go-app
go test ./... -v         # 单元测试
go test ./tests -tags=integration -v  # 集成测试

# 性能测试
cd frontend
npm run build            # 构建优化验证
npm run preview          # 预览构建结果
```

## 🏭 生产环境部署

### 1. 生产环境准备

#### 系统配置优化
```bash
# 系统参数优化
echo 'net.core.somaxconn = 4096' >> /etc/sysctl.conf
echo 'fs.file-max = 65536' >> /etc/sysctl.conf
sysctl -p

# 用户限制配置
echo '* soft nofile 65536' >> /etc/security/limits.conf
echo '* hard nofile 65536' >> /etc/security/limits.conf

# 防火墙配置
ufw allow 80/tcp          # HTTP
ufw allow 443/tcp         # HTTPS
ufw allow 22/tcp          # SSH
ufw --force enable
```

#### SSL/TLS证书配置
```bash
# 使用Let's Encrypt (推荐)
apt-get install certbot nginx
certbot --nginx -d your-domain.com

# 或使用自签名证书 (开发/测试)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/cube-castle.key \
  -out /etc/ssl/certs/cube-castle.crt
```

### 2. 生产环境Docker配置

#### 前端生产构建 🆕
```dockerfile
# frontend/Dockerfile.production
FROM node:18-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

# 生产运行时
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.prod.conf /etc/nginx/nginx.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### Nginx配置优化
```nginx
# frontend/nginx.prod.conf
events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    
    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript 
               application/javascript application/json application/xml+rss;
    
    server {
        listen 80;
        server_name your-domain.com;
        root /usr/share/nginx/html;
        index index.html;
        
        # 前端路由支持
        location / {
            try_files $uri $uri/ /index.html;
        }
        
        # 静态资源缓存
        location /assets/ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
        
        # API代理
        location /api/ {
            proxy_pass http://backend:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
```

### 3. 生产环境Docker Compose

```yaml
# docker-compose.production.yml
version: '3.8'

services:
  # 前端服务 🆕
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.production
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./ssl:/etc/ssl/certs:ro
    depends_on:
      - backend
    restart: unless-stopped

  # 后端服务
  backend:
    build:
      context: ./go-app
      dockerfile: Dockerfile.production
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DATABASE_URL=postgresql://postgres:${POSTGRES_PASSWORD}@postgres:5432/cubecastle
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - redis
      - neo4j
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '1.0'

  # AI服务
  ai-service:
    build:
      context: ./python-ai
      dockerfile: Dockerfile.production
    ports:
      - "50051:50051"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    restart: unless-stopped

  # 数据库服务
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: cubecastle
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 4G

  neo4j:
    image: neo4j:5-enterprise
    environment:
      NEO4J_AUTH: neo4j/${NEO4J_PASSWORD}
      NEO4J_ACCEPT_LICENSE_AGREEMENT: "yes"
    volumes:
      - neo4j_data:/data
      - neo4j_logs:/logs
    ports:
      - "7474:7474"
      - "7687:7687"
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  neo4j_data:
  neo4j_logs:
  redis_data:
```

### 4. 生产环境启动

```bash
# 设置环境变量
export POSTGRES_PASSWORD=$(openssl rand -base64 32)
export NEO4J_PASSWORD=$(openssl rand -base64 32)
export JWT_SECRET=$(openssl rand -base64 64)

# 启动生产环境
docker-compose -f docker-compose.production.yml up -d

# 验证服务状态
docker-compose -f docker-compose.production.yml ps

# 查看日志
docker-compose -f docker-compose.production.yml logs -f
```

### 5. 健康检查和监控

```bash
# 应用健康检查
curl https://your-domain.com/health

# 数据库连接检查
curl https://your-domain.com/api/v1/health/database

# 系统监控检查
curl https://your-domain.com/metrics
```

## ☁️ 云平台部署

### Kubernetes部署

#### 前端Deployment 🆕
```yaml
# k8s/frontend-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cube-castle-frontend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cube-castle-frontend
  template:
    metadata:
      labels:
        app: cube-castle-frontend
    spec:
      containers:
      - name: frontend
        image: cube-castle/frontend:2.1.0
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: cube-castle-frontend-service
spec:
  selector:
    app: cube-castle-frontend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: LoadBalancer
```

#### Ingress配置
```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cube-castle-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
spec:
  tls:
  - hosts:
    - your-domain.com
    secretName: cube-castle-tls
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: cube-castle-frontend-service
            port:
              number: 80
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: cube-castle-backend-service
            port:
              number: 8080
```

### AWS ECS部署

```json
{
  "family": "cube-castle-frontend",
  "taskRoleArn": "arn:aws:iam::account:role/ecsTaskRole",
  "executionRoleArn": "arn:aws:iam::account:role/ecsTaskExecutionRole",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "frontend",
      "image": "your-account.dkr.ecr.region.amazonaws.com/cube-castle-frontend:latest",
      "portMappings": [
        {
          "containerPort": 80,
          "protocol": "tcp"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/cube-castle-frontend",
          "awslogs-region": "us-west-2",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

## 📊 监控和日志

### Prometheus配置
```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'cube-castle-backend'
    static_configs:
      - targets: ['backend:8080']
    metrics_path: '/metrics'

  - job_name: 'cube-castle-frontend'
    static_configs:
      - targets: ['frontend:80']
    metrics_path: '/metrics'
```

### Grafana仪表板
```json
{
  "dashboard": {
    "title": "Cube Castle - Frontend Performance",
    "panels": [
      {
        "title": "前端页面加载时间",
        "type": "stat",
        "targets": [
          {
            "expr": "avg(frontend_page_load_duration_seconds)"
          }
        ]
      },
      {
        "title": "API响应时间",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, http_request_duration_seconds_bucket{job=\"cube-castle-backend\"})"
          }
        ]
      }
    ]
  }
}
```

### 日志聚合 (ELK Stack)
```yaml
# logging/filebeat.yml
filebeat.inputs:
- type: container
  paths:
    - /var/lib/docker/containers/*/*.log
  processors:
    - add_docker_metadata:
        host: "unix:///var/run/docker.sock"

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "cube-castle-%{+yyyy.MM.dd}"

setup.kibana:
  host: "kibana:5601"
```

## 🔒 安全配置

### 安全检查清单
```yaml
网络安全:
  - [x] HTTPS强制启用
  - [x] 防火墙配置正确
  - [x] VPC网络隔离 (云环境)
  - [x] 安全组配置 (云环境)

应用安全:
  - [x] JWT密钥足够强度
  - [x] 数据库密码复杂度
  - [x] API访问控制
  - [x] CORS配置正确

数据安全:
  - [x] 数据库加密存储
  - [x] 传输层TLS加密
  - [x] 备份数据加密
  - [x] 敏感数据脱敏

合规要求:
  - [x] GDPR数据保护
  - [x] 审计日志完整
  - [x] 访问控制记录
  - [x] 数据保留策略
```

## 🔄 CI/CD集成

### GitHub Actions工作流
```yaml
# .github/workflows/deploy.yml
name: Deploy Cube Castle

on:
  push:
    branches: [master]

jobs:
  # 前端构建和部署
  frontend-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      
      - name: Install dependencies
        working-directory: ./frontend
        run: npm ci
      
      - name: Run tests
        working-directory: ./frontend
        run: |
          npm run test
          npm run test:e2e
      
      - name: Build production
        working-directory: ./frontend
        run: npm run build
      
      - name: Deploy to production
        run: |
          docker build -t cube-castle-frontend:latest ./frontend
          docker push your-registry/cube-castle-frontend:latest

  # 后端构建和部署
  backend-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      
      - name: Run tests
        working-directory: ./go-app
        run: go test ./... -v
      
      - name: Build and deploy
        run: |
          docker build -t cube-castle-backend:latest ./go-app
          docker push your-registry/cube-castle-backend:latest
```

## 🚨 故障排查

### 常见问题和解决方案

#### 前端相关问题 🆕
```yaml
问题: Vite开发服务器启动失败
解决: 
  - 检查Node.js版本 >= 18
  - 清理node_modules: rm -rf node_modules && npm install
  - 检查端口3000是否被占用: lsof -i :3000

问题: Canvas Kit组件样式异常
解决:
  - 验证Canvas Kit版本兼容性
  - 检查CSS导入顺序
  - 清理浏览器缓存

问题: API请求失败 (CORS错误)
解决:
  - 检查后端CORS配置
  - 验证API基础URL配置
  - 确认网络连通性
```

#### 后端相关问题
```yaml
问题: 数据库连接失败
解决:
  - 检查数据库服务状态: docker ps
  - 验证连接字符串格式
  - 检查防火墙设置

问题: 内存使用过高
解决:
  - 调整Go垃圾回收参数: GOGC=100
  - 检查goroutine泄漏: go tool pprof
  - 优化数据库查询

问题: API响应时间慢
解决:
  - 启用数据库查询日志
  - 检查索引使用情况
  - 优化Neo4j查询语句
```

### 诊断工具
```bash
# 系统资源监控
htop
iostat -x 1
free -h

# 网络连接检查
netstat -tulpn | grep :3000
netstat -tulpn | grep :8080

# Docker容器诊断
docker stats
docker logs <container-id>
docker exec -it <container-id> /bin/sh

# 数据库连接测试
psql -h localhost -U postgres -d cubecastle -c "SELECT 1;"
curl -u neo4j:password http://localhost:7474/db/data/

# 应用健康检查
curl -f http://localhost:3000/ || echo "前端服务异常"
curl -f http://localhost:8080/health || echo "后端服务异常"
```

## 📈 性能优化

### 前端性能优化 🆕
```yaml
Vite构建优化:
  - 代码分割: 动态导入大型组件
  - 树摇优化: 移除未使用的Canvas Kit组件
  - 压缩优化: Gzip/Brotli压缩静态资源
  - 缓存策略: 长期缓存不变资源

运行时优化:
  - React.memo: 防止不必要的重渲染
  - useMemo/useCallback: 优化计算和函数创建
  - 虚拟滚动: 处理大数据列表
  - 图片懒加载: 减少初始加载时间
```

### 后端性能优化
```yaml
数据库优化:
  - 连接池配置: max_connections=200
  - 查询优化: 使用EXPLAIN分析
  - 索引优化: 基于查询模式创建索引
  - 缓存策略: Redis缓存热点数据

应用优化:
  - Goroutine池: 控制并发数量
  - 内存管理: 及时释放大对象
  - CPU优化: 避免CPU密集型操作阻塞
  - 网络优化: 启用HTTP/2和压缩
```

## 📚 相关资源

### 官方文档
- [Vite 官方文档](https://vitejs.dev/)
- [Canvas Kit 组件库](https://workday.github.io/canvas-kit/)
- [Docker 部署指南](https://docs.docker.com/)
- [Kubernetes 部署文档](https://kubernetes.io/docs/)

### 监控工具
- [Prometheus](https://prometheus.io/)
- [Grafana](https://grafana.com/)
- [ELK Stack](https://www.elastic.co/elk-stack/)

### 云服务文档
- [AWS ECS](https://aws.amazon.com/ecs/)
- [Azure Container Instances](https://azure.microsoft.com/en-us/services/container-instances/)
- [Google Cloud Run](https://cloud.google.com/run)

---

> **更新日期**: 2025年8月6日  
> **文档维护**: Cube Castle DevOps团队  
> **部署状态**: 生产就绪 ✅  
> **技术支持**: support@cubecastle.com