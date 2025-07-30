# 🚀 元合约可视化编辑器 - 部署指南

## 📋 部署概述

本指南提供Cube Castle元合约可视化编辑器的完整部署方案。基于城堡蓝图的雄伟单体架构，支持本地开发、测试和生产环境的一键部署。

## 🏗️ 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                    Cube Castle Editor                       │
│                   (Single Docker Container)                 │
├─────────────────────────────────────────────────────────────┤
│  🎨 React Frontend     │  🏰 Go Backend (Monolith)         │
│  - Visual Editor       │  - MetaContract Compiler          │
│  - Monaco Editor       │  - WebSocket Server               │
│  - Template System     │  - Local AI Service              │
│  - Multi-Panel Preview │  - REST API                      │
├─────────────────────────────────────────────────────────────┤
│  🗄️ PostgreSQL 15     │  ⚡ Redis 7        │ 🔍 Neo4j    │
│  - Meta Contract Data  │  - Session Cache   │ - Relations  │
│  - User Sessions      │  - Real-time Data  │ - Graph Data │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 环境要求

### **基础要求**
- **操作系统**: Linux/macOS/Windows + WSL2
- **Docker**: 20.10+ & Docker Compose v2+
- **内存**: 最低4GB，推荐8GB
- **磁盘**: 最低10GB可用空间

### **开发环境**
- **Node.js**: 18.0+ (前端开发)
- **Go**: 1.21+ (后端开发)
- **Git**: 版本控制

### **生产环境**
- **CPU**: 2核心以上
- **内存**: 8GB以上
- **网络**: 稳定的网络连接
- **SSL证书**: HTTPS支持(推荐)

## 🛠️ 安装步骤

### **Step 1: 获取代码**
```bash
# 克隆项目仓库
git clone <cube-castle-repository-url>
cd cube-castle

# 检查项目结构
ls -la
# 应该看到: go-app/, nextjs-app/, docs/, docker-compose.*.yml
```

### **Step 2: 环境配置**
```bash
# 复制环境配置模板
cp .env.example .env

# 编辑配置文件 (重要!)
nano .env
```

**关键配置项**:
```bash
# 数据库配置
DATABASE_URL=postgres://cube_user:cube_pass@postgres:5432/cube_castle
REDIS_URL=redis://redis:6379
NEO4J_URI=bolt://neo4j:7687

# 应用配置
APP_ENV=development  # development/production
APP_PORT=8080
FRONTEND_PORT=3000

# AI服务配置 (可选)
AI_ENABLED=true
AI_MODEL_PATH=/app/models

# 安全配置
JWT_SECRET=your-super-secret-jwt-key
ENCRYPTION_KEY=your-32-char-encryption-key
```

### **Step 3: 一键启动**

#### **开发环境启动**
```bash
# 启动开发环境 (包含热重载)
docker-compose -f docker-compose.editor-dev.yml up -d

# 查看服务状态
docker-compose -f docker-compose.editor-dev.yml ps

# 查看日志
docker-compose -f docker-compose.editor-dev.yml logs -f
```

#### **生产环境启动**
```bash
# 构建生产镜像
docker-compose -f docker-compose.editor-prod.yml build

# 启动生产环境
docker-compose -f docker-compose.editor-prod.yml up -d

# 健康检查
curl http://localhost/health
```

### **Step 4: 验证部署**

#### **服务检查**
```bash
# 检查所有容器状态
docker ps

# 检查数据库连接
docker exec -it cube-castle-postgres psql -U cube_user -d cube_castle -c "SELECT version();"

# 检查Redis连接
docker exec -it cube-castle-redis redis-cli ping

# 检查应用健康状态
curl http://localhost:8080/health
```

#### **功能验证**
1. **前端访问**: http://localhost:3000
2. **编辑器访问**: http://localhost:3000/metacontract-editor/advanced
3. **API文档**: http://localhost:8080/swagger
4. **健康检查**: http://localhost:8080/health

## 🔒 安全配置

### **SSL/TLS配置**
```bash
# 创建SSL证书目录
mkdir -p ./ssl

# 使用Let's Encrypt (生产环境)
sudo apt install certbot
sudo certbot certonly --standalone -d your-domain.com

# 复制证书到项目目录
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem ./ssl/
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem ./ssl/

# 更新Nginx配置
nano ./nginx/nginx.conf
```

### **防火墙配置**
```bash
# Ubuntu/Debian
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 22/tcp
sudo ufw enable

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --reload
```

### **数据库安全**
```bash
# 设置强密码
export POSTGRES_PASSWORD=$(openssl rand -base64 32)
export REDIS_PASSWORD=$(openssl rand -base64 32)

# 更新配置文件
echo "POSTGRES_PASSWORD=$POSTGRES_PASSWORD" >> .env
echo "REDIS_PASSWORD=$REDIS_PASSWORD" >> .env
```

## 🔄 数据备份和恢复

### **自动备份脚本**
```bash
#!/bin/bash
# backup.sh - 自动备份脚本

BACKUP_DIR="/backups/cube-castle"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份PostgreSQL
docker exec cube-castle-postgres pg_dump -U cube_user cube_castle > $BACKUP_DIR/postgres_$DATE.sql

# 备份Redis
docker exec cube-castle-redis redis-cli SAVE
docker cp cube-castle-redis:/data/dump.rdb $BACKUP_DIR/redis_$DATE.rdb

# 备份应用配置
tar -czf $BACKUP_DIR/config_$DATE.tar.gz .env docker-compose*.yml nginx/

# 清理旧备份 (保留30天)
find $BACKUP_DIR -type f -mtime +30 -delete

echo "Backup completed: $DATE"
```

### **恢复脚本**
```bash
#!/bin/bash
# restore.sh - 数据恢复脚本

BACKUP_FILE=$1
if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: ./restore.sh <backup_file>"
    exit 1
fi

# 停止服务
docker-compose -f docker-compose.editor-prod.yml down

# 恢复PostgreSQL
docker run --rm -v $(pwd):/backup -v cube-castle_postgres_data:/var/lib/postgresql/data postgres:15 \
    sh -c "pg_restore -U cube_user -d cube_castle /backup/$BACKUP_FILE"

# 重启服务
docker-compose -f docker-compose.editor-prod.yml up -d

echo "Restore completed"
```

## 📊 监控和日志

### **日志管理**
```bash
# 查看应用日志
docker-compose logs -f cube-castle

# 查看特定服务日志
docker-compose logs -f postgres
docker-compose logs -f redis

# 日志轮转配置
cat > /etc/logrotate.d/cube-castle << EOF
/var/log/cube-castle/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 root root
}
EOF
```

### **性能监控**
```bash
# 容器资源使用情况
docker stats

# 系统资源监控
htop
iostat -x 1
free -h

# 数据库性能
docker exec -it cube-castle-postgres psql -U cube_user -d cube_castle -c "
SELECT schemaname,tablename,attname,n_distinct,correlation 
FROM pg_stats WHERE schemaname = 'public';
"
```

### **健康检查脚本**
```bash
#!/bin/bash
# health-check.sh - 系统健康检查

echo "=== Cube Castle Health Check ==="

# 检查容器状态
echo "1. Container Status:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 检查服务响应
echo -e "\n2. Service Health:"
curl -s http://localhost:8080/health | jq .
curl -s http://localhost:3000 > /dev/null && echo "Frontend: OK" || echo "Frontend: ERROR"

# 检查数据库
echo -e "\n3. Database Status:"
docker exec cube-castle-postgres pg_isready -U cube_user && echo "PostgreSQL: OK" || echo "PostgreSQL: ERROR"
docker exec cube-castle-redis redis-cli ping | grep -q "PONG" && echo "Redis: OK" || echo "Redis: ERROR"

# 检查磁盘空间
echo -e "\n4. Disk Usage:"
df -h | grep -E "/$|docker"

# 检查内存使用
echo -e "\n5. Memory Usage:"
free -h
```

## 🔧 故障排除

### **常见问题**

#### **问题1: 容器启动失败**
```bash
# 检查日志
docker-compose logs <service-name>

# 检查端口占用
netstat -tulpn | grep :8080
netstat -tulpn | grep :3000

# 解决方案
sudo lsof -i :8080  # 找到占用进程
sudo kill -9 <PID>  # 终止进程
```

#### **问题2: 数据库连接失败**
```bash
# 检查数据库容器
docker exec -it cube-castle-postgres psql -U cube_user -d cube_castle

# 检查网络连接
docker network ls
docker network inspect cube-castle_default

# 重建数据库
docker-compose down -v
docker-compose up -d
```

#### **问题3: 前端编译错误**
```bash
# 清理Node.js缓存
docker exec -it cube-castle-frontend npm cache clean --force

# 重新安装依赖
docker exec -it cube-castle-frontend rm -rf node_modules package-lock.json
docker exec -it cube-castle-frontend npm install

# 重启前端服务
docker-compose restart frontend
```

### **性能优化**

#### **数据库优化**
```sql
-- PostgreSQL性能调优
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;

-- 重启数据库使配置生效
SELECT pg_reload_conf();
```

#### **Redis优化**
```bash
# Redis内存优化
echo "maxmemory 512mb" >> /etc/redis/redis.conf
echo "maxmemory-policy allkeys-lru" >> /etc/redis/redis.conf

# 持久化配置
echo "save 900 1" >> /etc/redis/redis.conf
echo "save 300 10" >> /etc/redis/redis.conf
echo "save 60 10000" >> /etc/redis/redis.conf
```

## 📈 扩容和升级

### **垂直扩容 (增加资源)**
```bash
# 更新Docker Compose资源限制
nano docker-compose.editor-prod.yml

# 示例配置
services:
  cube-castle:
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '2.0'
        reservations:
          memory: 1G
          cpus: '1.0'
```

### **应用升级**
```bash
#!/bin/bash
# upgrade.sh - 应用升级脚本

echo "Starting Cube Castle upgrade..."

# 1. 备份当前数据
./backup.sh

# 2. 拉取最新代码
git fetch origin
git checkout v2.0.0  # 替换为目标版本

# 3. 构建新镜像
docker-compose -f docker-compose.editor-prod.yml build --no-cache

# 4. 滚动更新
docker-compose -f docker-compose.editor-prod.yml up -d

# 5. 验证升级
sleep 10
curl -f http://localhost:8080/health || {
    echo "Upgrade failed, rolling back..."
    git checkout v1.0.0
    docker-compose -f docker-compose.editor-prod.yml up -d
    exit 1
}

echo "Upgrade completed successfully!"
```

## 🚀 快速命令参考

### **日常运维命令**
```bash
# 启动所有服务
docker-compose -f docker-compose.editor-prod.yml up -d

# 停止所有服务
docker-compose -f docker-compose.editor-prod.yml down

# 重启特定服务
docker-compose restart cube-castle

# 查看服务状态
docker-compose ps

# 查看实时日志
docker-compose logs -f --tail=100

# 进入容器调试
docker exec -it cube-castle-app bash
docker exec -it cube-castle-postgres psql -U cube_user cube_castle

# 数据备份
./scripts/backup.sh

# 性能监控
docker stats --no-stream
```

### **开发调试命令**
```bash
# 启动开发环境
docker-compose -f docker-compose.editor-dev.yml up -d

# 热重载开发
cd nextjs-app && npm run dev
cd go-app && air

# 代码格式化
cd nextjs-app && npm run lint:fix
cd go-app && gofmt -w .

# 运行测试
cd nextjs-app && npm test
cd go-app && go test ./...
```

---

## 📞 支持和帮助

如果在部署过程中遇到问题，请：

1. **查看日志**: `docker-compose logs -f`
2. **检查配置**: 确认`.env`文件配置正确
3. **运行健康检查**: `./scripts/health-check.sh`
4. **查阅文档**: `/docs/metacontract-editor/`

**部署完成后，您可以通过以下地址访问系统：**
- 🎨 **可视化编辑器**: http://localhost:3000/metacontract-editor/advanced
- 📊 **管理面板**: http://localhost:3000/admin
- 🔧 **API文档**: http://localhost:8080/swagger
- ❤️ **健康检查**: http://localhost:8080/health