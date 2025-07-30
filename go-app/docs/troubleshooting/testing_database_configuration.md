# 测试数据库配置指南

## 概述

本项目支持多种数据库后端进行测试，以平衡测试速度和生产环境一致性。

## 数据库架构

### 生产环境
- **PostgreSQL**: 主要关系型数据存储（通过Ent ORM）
- **Neo4j**: 图形数据库，用于复杂组织关系查询

### 测试环境选项
1. **SQLite Memory** (默认) - 最快速度
2. **SQLite File** - 中等速度，支持持久化
3. **PostgreSQL Test** - 与生产环境一致

## 配置方法

### 环境变量配置

```bash
# 使用SQLite内存数据库（默认，无需设置）
# TEST_DB_TYPE=sqlite_memory

# 使用SQLite文件数据库
export TEST_DB_TYPE=sqlite

# 使用PostgreSQL测试数据库
export TEST_DB_TYPE=postgresql
export TEST_DATABASE_URL="postgresql://postgres:password@localhost:5432/cubecastle_test?sslmode=disable"
```

### Makefile集成

```makefile
# 快速单元测试（使用SQLite内存）
test-unit:
	@TEST_DB_TYPE=sqlite_memory go test ./internal/handler/...

# 集成测试（使用PostgreSQL）
test-integration:
	@TEST_DB_TYPE=postgresql TEST_DATABASE_URL="postgresql://postgres:password@localhost:5432/cubecastle_test?sslmode=disable" go test ./internal/handler/...

# 完整测试套件
test-all: test-unit test-integration
```

## 测试类型建议

### 1. 单元测试
- **推荐**: SQLite Memory
- **原因**: 最快速度，完全隔离
- **适用**: API handler测试、业务逻辑测试

### 2. 集成测试
- **推荐**: PostgreSQL Test
- **原因**: 与生产环境数据库类型一致
- **适用**: 复杂查询测试、数据完整性测试

### 3. 端到端测试
- **推荐**: PostgreSQL + Neo4j
- **原因**: 完整生产环境配置
- **适用**: 完整业务流程测试

## 测试数据库信息

当前测试会自动显示使用的数据库类型：

```go
// 查看当前测试数据库信息
info := testutil.GetTestDatabaseInfo()
fmt.Printf("测试数据库: %s (%s)\n", info.Type, info.PerformanceLevel)
```

## PostgreSQL测试数据库设置

### Docker方式（推荐）

```bash
# 启动PostgreSQL测试容器
docker run --name postgres-test -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=cubecastle_test -d postgres:13

# 测试连接
export TEST_DATABASE_URL="postgresql://postgres:password@localhost:5432/cubecastle_test?sslmode=disable"
```

### 本地安装方式

```bash
# 创建测试数据库
createdb cubecastle_test

# 设置环境变量
export TEST_DATABASE_URL="postgresql://localhost/cubecastle_test?sslmode=disable"
```

## 性能对比

| 数据库类型 | 测试速度 | 内存使用 | 生产一致性 | 适用场景 |
|-----------|---------|---------|-----------|----------|
| SQLite Memory | 最快 | 最少 | 低 | 单元测试 |
| SQLite File | 快 | 少 | 低 | 本地调试 |
| PostgreSQL Test | 中等 | 中等 | 高 | 集成测试 |

## 最佳实践

1. **开发阶段**: 使用SQLite Memory进行快速测试
2. **CI/CD**: 同时运行SQLite和PostgreSQL测试
3. **发布前**: 必须通过PostgreSQL集成测试
4. **性能基准**: 使用PostgreSQL进行性能测试

## 故障排除

### PostgreSQL连接失败
```bash
# 检查PostgreSQL服务状态
pg_isready -h localhost -p 5432

# 检查数据库是否存在
psql -h localhost -c "SELECT 1" cubecastle_test
```

### SQLite权限问题
```bash
# 确保有写入权限
chmod 755 .
rm -f testdb.sqlite
```

## 注意事项

1. **测试隔离**: 每个测试都会获得独立的数据库实例
2. **自动清理**: 测试结束后会自动清理数据
3. **迁移支持**: PostgreSQL测试会自动运行数据库迁移
4. **并发安全**: 支持并发测试执行