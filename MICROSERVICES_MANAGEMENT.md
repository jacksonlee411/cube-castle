# Cube Castle 微服务管理指南

## 概述

Cube Castle项目采用CQRS微服务架构，包含多个独立的服务组件。本指南介绍如何使用统一的管理脚本来管理这些微服务。

## 当前架构

### 核心组织管理服务
- **organization-api-gateway** (端口8000): 统一API网关，路由请求到不同的后端服务
- **organization-api-server** (端口8080): CQRS查询端，提供REST API查询接口
- **organization-graphql-service** (端口8090): GraphQL查询服务，支持复杂查询
- **organization-command-server** (端口9090): CQRS命令端，处理写操作（增删改）

### 其他业务服务
- **employee-server** (端口8081): 员工管理服务
- **position-server** (端口8082): 岗位管理服务

### 基础设施（Docker容器）
- PostgreSQL (端口5432): 主数据库
- Neo4j (端口7474/7687): 图数据库，用于复杂关系查询
- Redis (端口6379): 缓存层
- Kafka生态系统: 事件驱动和数据同步
- Temporal: 工作流引擎

## 微服务管理脚本

使用 `scripts/microservices-manager.sh` 脚本来管理所有微服务：

### 基本命令

```bash
# 显示所有服务状态
./scripts/microservices-manager.sh status

# 启动所有服务
./scripts/microservices-manager.sh start

# 停止所有服务
./scripts/microservices-manager.sh stop

# 重启所有服务
./scripts/microservices-manager.sh restart

# 编译所有服务
./scripts/microservices-manager.sh build

# 清理过期PID文件
./scripts/microservices-manager.sh cleanup
```

### 单个服务管理

```bash
# 启动特定服务
./scripts/microservices-manager.sh start organization-api-gateway

# 停止特定服务
./scripts/microservices-manager.sh stop organization-api-server

# 重启特定服务
./scripts/microservices-manager.sh restart organization-graphql-service
```

## API调用路径

### 前端应用调用路径
前端应用应该通过API网关统一访问：
- **基础URL**: `http://localhost:8000/api/v1`
- **查询操作**: 网关自动路由到8080端口（REST API）或8090端口（GraphQL）
- **命令操作**: 网关自动路由到9090端口（命令服务器）

### 直接服务调用（开发调试用）
- 查询操作: `http://localhost:8080/api/v1/organization-units`
- GraphQL查询: `http://localhost:8090/graphql`
- 命令操作: `http://localhost:9090/api/v1/organization-units`

## 服务启动顺序

建议的启动顺序（脚本会自动按此顺序启动）：

1. **organization-command-server** - 命令端服务
2. **organization-api-server** - 查询端服务 
3. **organization-graphql-service** - GraphQL服务
4. **organization-api-gateway** - API网关
5. **employee-server** - 员工服务
6. **position-server** - 岗位服务

## 健康检查

每个服务都提供健康检查端点：
- `http://localhost:{PORT}/health`

管理脚本会自动进行健康检查并显示状态。

## 日志文件

所有服务的日志文件位于：
- `cmd/{service-name}/logs/{service-name}.log`

例如：
- `cmd/organization-api-gateway/logs/organization-api-gateway.log`
- `cmd/organization-graphql-service/logs/organization-graphql-service.log`

## 故障排查

### 1. 服务启动失败
```bash
# 检查服务状态
./scripts/microservices-manager.sh status

# 查看服务日志
tail -f cmd/{service-name}/logs/{service-name}.log
```

### 2. 端口占用问题
```bash
# 检查端口占用
lsof -i :{PORT}

# 终止占用进程
kill -9 {PID}
```

### 3. 数据库连接问题
确保Docker容器正常运行：
```bash
docker-compose ps
```

### 4. 前端删除操作失败
确认API网关(8000)和命令服务器(9090)都在运行：
```bash
curl http://localhost:8000/health
curl http://localhost:9090/health
```

## 开发最佳实践

1. **启动开发环境**:
   ```bash
   # 启动基础设施
   docker-compose up -d
   
   # 启动微服务
   ./scripts/microservices-manager.sh start
   ```

2. **代码更改后重新部署**:
   ```bash
   # 重新编译并重启
   ./scripts/microservices-manager.sh build
   ./scripts/microservices-manager.sh restart
   ```

3. **停止开发环境**:
   ```bash
   # 停止微服务
   ./scripts/microservices-manager.sh stop
   
   # 停止基础设施（可选）
   docker-compose down
   ```

## 监控和运维

- 使用管理脚本定期检查服务状态
- 监控日志文件中的错误信息
- 定期清理过期的PID文件
- 确保基础设施服务正常运行

## 架构优势

1. **服务分离**: 查询和命令操作分离，支持独立扩展
2. **多协议支持**: 同时支持REST API和GraphQL
3. **统一网关**: 前端只需要知道一个API入口
4. **事件驱动**: 通过Kafka实现服务间异步通信
5. **缓存优化**: Redis缓存提升查询性能
6. **可观测性**: 统一的日志和健康检查机制

## 待优化项

1. 实施分布式追踪（Jaeger/Zipkin）
2. 集成Prometheus监控指标
3. 实施自动重启和故障恢复
4. 添加服务发现机制
5. 实施配置中心统一管理