# 时态管理系统故障排除指南

## 🔧 常见问题诊断与解决

### 问题分类索引
- [🚫 服务启动问题](#服务启动问题)
- [⚡ 性能问题](#性能问题) 
- [📊 数据同步问题](#数据同步问题)
- [🔍 查询问题](#查询问题)
- [✏️ 操作问题](#操作问题)
- [🌐 网络连接问题](#网络连接问题)

---

## 🚫 服务启动问题

### 问题1: 时态管理服务无法启动
**现象**: 访问 http://localhost:9091/health 返回连接错误

**诊断步骤**:
```bash
# 1. 检查服务进程
ps aux | grep main_no_version

# 2. 检查端口占用
lsof -i :9091

# 3. 查看服务日志
# 时态管理功能已整合到现有服务中
go run main_no_version.go
```

**常见原因和解决方案**:

#### 原因1: 数据库连接失败
```bash
# 检查数据库服务
docker-compose ps postgres

# 重启数据库
docker-compose restart postgres

# 测试连接
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "SELECT 1;"
```

#### 原因2: 端口被占用
```bash
# 查找占用进程
sudo lsof -i :9091

# 终止占用进程
sudo kill -9 <PID>

# 重新启动服务
# 时态管理功能已整合到现有服务中 && go run main_no_version.go
```

#### 原因3: Redis连接失败
```bash
# 检查Redis服务
docker-compose ps redis

# 重启Redis
docker-compose restart redis

# 测试连接
redis-cli -h localhost -p 6379 ping
```

### 问题2: 前端页面加载失败
**现象**: 访问 http://localhost:3000 显示错误或空白

**解决步骤**:
```bash
# 1. 检查前端开发服务器
cd frontend
npm run dev

# 2. 清理依赖和缓存
npm run clean
npm install
npm run dev

# 3. 检查代理配置
cat vite.config.ts | grep proxy
```

---

## ⚡ 性能问题

### 问题3: 查询响应缓慢
**现象**: API响应时间超过5秒

**性能诊断**:
```bash
# 1. 检查系统资源
top
free -m
df -h

# 2. 检查数据库性能
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "
SELECT query, mean_exec_time, calls 
FROM pg_stat_statements 
ORDER BY mean_exec_time DESC 
LIMIT 5;"

# 3. 检查缓存命中率
redis-cli -h localhost -p 6379 info stats
```

**优化方案**:

#### 方案1: 数据库索引优化
```bash
# 运行索引优化脚本
cd scripts
psql -h localhost -U user -d cubecastle -f optimize-temporal-performance.sql
```

#### 方案2: 缓存优化
```bash
# 清理过期缓存
redis-cli -h localhost -p 6379 FLUSHDB

# 重启缓存服务
docker-compose restart redis
```

#### 方案3: 查询参数优化
- 使用更精确的时间范围
- 避免查询过多历史记录
- 使用分页参数限制结果数量

### 问题4: 内存使用过高
**现象**: 系统内存使用率超过80%

**诊断和解决**:
```bash
# 1. 检查各服务内存使用
docker stats

# 2. 清理系统缓存
sync && echo 3 > /proc/sys/vm/drop_caches

# 3. 调整服务配置
# 编辑 docker-compose.yml，调整内存限制
docker-compose restart
```

---

## 📊 数据同步问题

### 问题5: CDC数据同步延迟
**现象**: 在命令服务创建数据后，查询服务查不到

**诊断步骤**:
```bash
# 1. 检查Kafka服务
docker-compose ps kafka zookeeper

# 2. 检查Debezium连接器状态
curl http://localhost:8083/connectors/postgres-connector/status

# 3. 查看Kafka主题消息
docker exec -it cube_castle_kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic postgres.public.organization_units \
  --from-beginning
```

**解决方案**:

#### 方案1: 重新配置CDC管道
```bash
# 运行CDC配置脚本
./scripts/setup-cdc-pipeline.sh

# 验证配置结果
curl http://localhost:8083/connectors/postgres-connector/config
```

#### 方案2: 手动触发同步
```bash
# 重启同步服务
cd cmd/organization-sync-service
go run main.go

# 手动运行同步脚本
python3 scripts/sync-organization-to-neo4j.py
```

### 问题6: 数据不一致
**现象**: PostgreSQL和Neo4j中的数据不匹配

**数据一致性检查**:
```bash
# 1. 检查PostgreSQL记录数
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "
SELECT COUNT(*) FROM organization_units WHERE is_current = true;"

# 2. 检查Neo4j记录数  
curl -X POST http://localhost:7474/db/data/cypher \
  -H "Content-Type: application/json" \
  -d '{"query": "MATCH (n:Organization) RETURN count(n)"}'

# 3. 比较关键字段
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "
SELECT code, name, updated_at FROM organization_units 
WHERE is_current = true ORDER BY updated_at DESC LIMIT 5;"
```

**修复数据不一致**:
```bash
# 重建Neo4j数据
./scripts/rebuild-neo4j-data.sh

# 或强制全量同步
./scripts/force-full-sync.sh
```

---

## 🔍 查询问题

### 问题7: 时态查询返回空结果
**现象**: API返回空数组或404错误

**排查步骤**:
```bash
# 1. 检查组织是否存在
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "
SELECT code, name, effective_date, end_date, is_current 
FROM organization_units WHERE code = 'YOUR_ORG_CODE';"

# 2. 检查查询时间范围
curl "http://localhost:9091/api/v1/organization-units/1000056/temporal?include_history=true&include_future=true"

# 3. 验证日期格式
curl "http://localhost:9091/api/v1/organization-units/1000056/temporal?as_of_date=2025-08-11"
```

**常见问题解决**:

#### 问题: 日期格式错误
```bash
# 错误格式
curl "...?as_of_date=2025/08/11"

# 正确格式  
curl "...?as_of_date=2025-08-11"
```

#### 问题: 查询时间点无数据
```bash
# 查看完整时间线
curl "http://localhost:9091/api/v1/organization-units/1000056/temporal?include_history=true"

# 确认有效期间
SELECT effective_date, end_date FROM organization_units WHERE code = 'YOUR_CODE';
```

### 问题8: GraphQL查询失败
**现象**: 前端查询报GraphQL错误

**诊断和修复**:
```bash
# 1. 检查GraphQL服务
curl http://localhost:8090/health

# 2. 测试GraphQL查询
curl -X POST http://localhost:8090/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ organizations { code name } }"}'

# 3. 检查服务日志
cd cmd/organization-query-service-unified
go run main.go
```

---

## ✏️ 操作问题

### 问题9: 版本创建失败
**现象**: 提交新版本时返回验证错误

**常见错误和解决方案**:

#### 错误1: "没有有效的字段变更"
```json
// 错误的请求格式
{
  "event_type": "UPDATE",
  "changes": {...}  // ❌ 错误字段名
}

// 正确的请求格式  
{
  "event_type": "UPDATE", 
  "change_data": {...}  // ✅ 正确字段名
}
```

#### 错误2: 生效日期冲突
```bash
# 查看现有版本的生效日期
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "
SELECT effective_date, end_date 
FROM organization_units 
WHERE code = 'YOUR_CODE' 
ORDER BY effective_date;"
```

#### 错误3: 必填字段缺失
检查以下必填字段：
- `event_type`: 必须是有效的事件类型
- `effective_date`: 必须是有效的日期格式
- `change_reason`: 不能为空
- `change_data`: 必须包含至少一个有效字段

### 问题10: 版本删除失败
**现象**: 删除操作返回错误或无响应

**检查和修复**:
```bash
# 1. 确认版本存在
curl "http://localhost:9091/api/v1/organization-units/1000056/temporal?include_history=true"

# 2. 检查删除权限
# 当前版本通常不允许直接删除

# 3. 使用正确的API端点
curl -X DELETE "http://localhost:9091/api/v1/organization-units/1000056/temporal/2025-09-01"
```

---

## 🌐 网络连接问题

### 问题11: 服务间通信失败
**现象**: 前端无法连接到后端API

**网络诊断**:
```bash
# 1. 检查端口连通性
telnet localhost 9091
telnet localhost 8090
telnet localhost 9090

# 2. 检查防火墙设置
sudo ufw status

# 3. 检查Docker网络
docker network ls
docker network inspect cube_castle_default
```

**修复网络问题**:
```bash
# 1. 重启Docker网络
docker-compose down
docker-compose up -d

# 2. 清理Docker网络
docker system prune -f

# 3. 重建容器
docker-compose up --build
```

### 问题12: 跨域请求问题
**现象**: 浏览器控制台显示CORS错误

**解决方案**:
```bash
# 检查CORS配置
# 时态管理功能已整合到现有服务中
grep -n "cors" main_no_version.go

# 确认CORS配置正确
AllowedOrigins: []string{"*"}
AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
```

---

## 🛠️ 系统维护

### 日常维护检查清单

#### 每日检查
```bash
# 1. 服务健康检查
curl http://localhost:9091/health
curl http://localhost:8090/health  
curl http://localhost:9090/health

# 2. 数据库连接测试
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "SELECT 1;"

# 3. 缓存状态检查
redis-cli -h localhost -p 6379 ping
```

#### 每周检查
```bash
# 1. 清理过期缓存
redis-cli -h localhost -p 6379 FLUSHEXPIRED

# 2. 数据库维护
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "VACUUM ANALYZE;"

# 3. 日志轮转
docker-compose logs --tail=1000 > logs/weekly-$(date +%Y%m%d).log
```

#### 每月检查
```bash
# 1. 数据备份
pg_dump -h localhost -U user cubecastle > backup/monthly-$(date +%Y%m%d).sql

# 2. 性能分析
PGPASSWORD=password psql -h localhost -U user -d cubecastle -f scripts/performance-report.sql

# 3. 存储空间检查
df -h
docker system df
```

---

## 📞 获取帮助

### 自助诊断工具
```bash
# 运行完整健康检查
./scripts/health-check-cqrs.sh

# 生成系统报告
./scripts/generate-system-report.sh
```

### 日志文件位置
- **应用日志**: `docker-compose logs [service-name]`
- **数据库日志**: `docker-compose logs postgres`
- **前端日志**: 浏览器开发者工具Console面板

### 紧急恢复程序
```bash
# 1. 完全重启系统
docker-compose down
docker-compose up -d

# 2. 重建所有服务
./scripts/rebuild-all-services.sh

# 3. 从备份恢复
./scripts/restore-from-backup.sh
```

---

*故障排除指南 - 快速解决常见问题*  
*最后更新: 2025-08-11*