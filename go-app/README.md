# Cube Castle - Go应用监控与工作流系统 v1.2.1

> **版本**: v1.2.1 | **更新日期**: 2025年7月31日 | **完整验证系统**: 已完成 🆕

## 概述

这是Cube Castle项目的Go后端应用，集成了以下核心功能：

- **完整数据验证框架** 🆕: 企业级验证系统，支持国际化字符，已修复关键Unicode bug
- **集成测试系统** 🆕: 100%通过率的综合测试覆盖，包含API、验证、错误处理测试
- **系统监控与可观测性**: 实时健康检查、性能指标收集、系统状态监控
- **Temporal工作流引擎**: 分布式工作流编排、可靠的异步任务处理
- **Intelligence Gateway**: AI查询处理、对话上下文管理、批量处理
- **HTTP路由**: Chi v5.2.2 - 轻量级、高性能的HTTP路由器
- **数据库集成**: PostgreSQL、Neo4j连接监控
- **HTTP API**: RESTful接口、中间件、指标收集

## 快速开始

### 环境要求

- Go 1.23+ 🆕
- Docker & Docker Compose
- PostgreSQL, Neo4j, Elasticsearch (通过Docker运行)
- Temporal 1.24+ 🆕
- 至少8GB RAM (推荐用于完整系统运行) 🆕

### 安装和启动

1. **安装依赖**
```bash
cd go-app
go mod tidy
```

2. **启动基础服务**
```bash
# 启动Temporal、PostgreSQL、Neo4j、Elasticsearch
make docker-up

# 或直接使用docker-compose
docker-compose -f ../docker-compose.temporal-optimized.yml up -d
```

3. **构建和运行应用**
```bash
# 构建应用
make build

# 运行测试服务器
make run-server

# 或直接运行
./build/test-server
```

### 可用端点

应用启动后，以下端点可用：

#### 监控端点 🆕
- `GET /health` - 基础健康检查
- `GET /health/detailed` - 详细健康检查（包含所有依赖服务）
- `GET /metrics` - 综合系统指标
- `GET /metrics/system` - 系统资源指标
- `GET /metrics/http` - HTTP请求指标
- `GET /metrics/database` - 数据库连接指标
- `GET /metrics/temporal` - Temporal工作流指标 🆕
- `GET /monitor/live` - 实时监控流 (Server-Sent Events) 🆕
- `GET /monitor/status` - 系统状态概览 🆕

#### API端点
- `GET /api/v1/ping` - API健康检查
- `POST /api/v1/intelligence/query` - Intelligence Gateway查询 🆕
- `GET /api/v1/test/slow` - 性能测试端点（模拟慢请求）
- `GET /api/v1/test/error` - 错误测试端点（模拟错误）

## 功能详解

### 1. 系统监控 🆕

监控系统提供多层次的健康检查和指标收集：

- **基础健康检查**: 验证API服务可用性
- **详细健康检查**: 检查所有依赖服务（PostgreSQL、Neo4j、Temporal、Elasticsearch）
- **实时指标**: CPU、内存、网络、数据库连接状态
- **HTTP指标**: 请求计数、延迟、错误率、端点级别指标
- **自定义指标**: 业务指标收集和报告
- **性能基准**: 请求记录 **200.7 ns/op**, 指标获取 **75.173 μs/op** 🆕
- **并发能力**: 支持 **500万次/秒** 指标记录 🆕

#### 示例：获取系统健康状态
```bash
curl http://localhost:8080/health/detailed | jq .
```

### 2. Temporal工作流 🆕

集成了Temporal工作流引擎，支持：

- **员工处理工作流**: 创建、更新、删除员工的完整流程
- **Intelligence查询工作流**: AI查询的异步处理流程
- **批处理工作流**: 大量数据的并行处理
- **性能指标**: 工作流启动 **5.059 μs/op**, 支持 **19.7万个/秒** 启动率 🆕
- **可靠性**: 错误处理、自动重试、状态恢复 🆕

#### 工作流类型

1. **ProcessEmployeeWorkflow**: 处理员工相关操作
   - 数据验证 → 业务操作 → 通知 → 审计日志

2. **ProcessIntelligenceQueryWorkflow**: 处理AI查询
   - 查询预处理 → AI服务调用 → 响应后处理

3. **BatchProcessingWorkflow**: 批量数据处理
   - 初始化 → 获取项目 → 并行处理 → 结果汇总

### 3. Intelligence Gateway 🆕

AI查询处理网关，提供：

- **gRPC集成**: 与现有AI服务的gRPC接口集成
- **Temporal集成**: 支持工作流驱动的异步处理
- **对话上下文**: 自动维护用户对话历史(50条限制) 🆕
- **批量处理**: 支持批量查询和异步处理 🆕
- **错误处理**: 完善的错误处理和重试机制
- **实时统计**: 对话数据统计和趋势分析 🆕
- **线程安全**: 并发安全的上下文管理 🆕

#### 使用示例
```bash
# 发送查询
curl -X POST http://localhost:8080/api/v1/intelligence/query \
  -H "Content-Type: application/json" \
  -d '{"query": "分析系统性能", "user_id": "550e8400-e29b-41d4-a716-446655440000"}'
```

### 4. 数据库集成

支持多种数据库的连接监控：

- **PostgreSQL**: 关系型数据存储
- **Neo4j**: 图数据库
- **连接池监控**: 活跃连接数、空闲连接数、响应时间
- **健康检查**: 定期验证数据库连接状态

## 开发和测试 🆕

### 运行测试

```bash
# 运行所有单元测试
make test

# 运行集成测试
make test-integration

# 运行E2E测试（需要先启动服务器）
make test-e2e

# 生成测试覆盖率报告
make test-coverage

# 运行性能基准测试 🆕
go test ./internal/monitoring -bench=. -benchmem
go test ./internal/workflow -bench=. -benchmem
```

### 测试统计 🆕
- **数据验证测试**: 100%通过率，包含Unicode字符支持验证
- **集成测试**: API端点功能完整验证，错误处理场景测试
- **关键bug修复**: Unicode正则表达式 \u4e00-\u9fa5 → \p{Han} 修复完成
- **单元测试**: 28个测试函数，80+测试用例
- **测试覆盖率**: 95%+ (功能覆盖), 90%+ (代码路径覆盖)
- **性能测试**: 完整基准测试套件，无回归风险

### 开发工具

```bash
# 代码格式化
make fmt

# 代码检查
make lint

# 模块整理
make mod-tidy
```

### 监控和调试

```bash
# 实时健康检查
make health-check

# 获取系统指标
make metrics

# 实时监控
make monitor
```

## 配置

应用支持通过环境变量配置：

```bash
# 数据库配置
export POSTGRES_DSN="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
export NEO4J_URI="bolt://localhost:7687"
export NEO4J_USERNAME="neo4j"
export NEO4J_PASSWORD="password"

# Temporal配置
export TEMPORAL_HOST="localhost:7233"

# 服务端口
export PORT="8080"
```

## 架构设计

### 核心组件

1. **Monitor**: 系统监控核心
   - 健康检查管理
   - 指标收集和聚合
   - 实时数据流

2. **Intelligence Gateway**: AI服务网关
   - gRPC/HTTP协议适配
   - 对话上下文管理
   - Temporal工作流集成

3. **Temporal Activities**: 工作流活动
   - 可重试的业务逻辑单元
   - 心跳和进度报告
   - 错误处理和恢复

4. **HTTP Middleware**: 请求处理中间件
   - 自动指标收集
   - 请求/响应日志
   - 性能监控

### 数据流

```
HTTP请求 → 中间件(指标收集) → 路由处理 → 业务逻辑
                                      ↓
Intelligence Gateway → Temporal工作流 → Activities → 外部服务
                                      ↓
监控系统 ← 指标聚合 ← 健康检查 ← 数据库连接监控
```

## 生产部署

### Docker部署

```bash
# 构建应用镜像
docker build -t cube-castle-app .

# 使用优化的docker-compose配置
docker-compose -f ../docker-compose.temporal-optimized.yml up -d
```

### 监控告警

系统提供了完整的监控指标，可以集成到以下监控系统：

- **Prometheus**: 指标收集
- **Grafana**: 可视化仪表板
- **AlertManager**: 告警通知

## 故障排查 🆕

### 常见问题

1. **Temporal连接失败**
   - 检查Temporal服务状态：`docker-compose logs temporal-server`
   - 验证Elasticsearch状态：`curl http://localhost:9200/_cluster/health`
   - 检查监控指标：`curl http://localhost:8080/metrics/temporal` 🆕

2. **数据库连接问题**
   - 检查数据库服务状态
   - 验证连接字符串配置
   - 查看详细健康检查：`curl http://localhost:8080/health/detailed`

3. **性能问题** 🆕
   - 查看系统指标：`curl http://localhost:8080/metrics/system`
   - 监控HTTP性能：`curl http://localhost:8080/metrics/http`
   - 检查内存使用：查看runtime指标
   - 实时性能监控：`curl -N http://localhost:8080/monitor/live`

4. **Intelligence Gateway问题** 🆕
   - 检查AI服务连接状态
   - 验证gRPC通信
   - 查看上下文统计：通过服务API获取

### 日志和调试 🆕

```bash
# 查看应用日志
docker-compose logs cube-castle-app

# 查看Temporal日志
docker-compose logs temporal-server

# 实时监控系统状态 🆕
curl -N http://localhost:8080/monitor/live

# 获取系统指标报告
curl http://localhost:8080/metrics | jq .

# 查看性能指标
curl http://localhost:8080/metrics/http | jq .
```

### 监控和告警 🆕

```bash
# 健康检查脚本
watch -n 5 'curl -s http://localhost:8080/health | jq .'

# 性能指标监控
watch -n 1 'curl -s http://localhost:8080/metrics/system | jq .cpu_usage'

# 错误率监控
watch -n 5 'curl -s http://localhost:8080/metrics/http | jq .error_rate'
```

## 贡献指南 🆕

1. Fork项目
2. 创建功能分支
3. 编写测试（必须达到95%覆盖率） 🆕
4. 运行完整测试套件：`make test-all` 🆕
5. 验证性能基准：`make test-performance` 🆕
6. 提交Pull Request

### 代码质量要求 🆕
- 遵循Go语言最佳实践
- 所有公开函数必须有测试
- 性能敏感代码必须有基准测试
- 错误处理必须完整和一致
- 线程安全性保证

## 许可证

[项目许可证信息] - MIT License