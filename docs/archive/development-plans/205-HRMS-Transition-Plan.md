# 205-HRMS 系统过渡方案详解

**文档编号**: 205
**标题**: 从当前架构到模块化单体的过渡方案详细指南
**创建日期**: 2025-11-03
**最后更新**: 2025-11-04
**状态**: ✅ **已完成（分解为 Plan 210/211/212）** — 本文档作为设计决议与参考资料保留
**相关文档**:
- `203-hrms-module-division-plan.md`（主文档）
- `204-HRMS-Implementation-Roadmap.md`（路线图）
- `206-Alignment-With-200-201.md`（对齐分析）
- `210-database-baseline-reset-plan.md`（前置条件：数据库基线重建）
- `211-phase1-module-unification-plan.md`（执行方案：Phase1 Day1-10）
- `212-shared-architecture-alignment-plan.md`（架构审查：Day6-7）
- `205-Plan-Alignment-Assessment.md`（本文档与后续计划的对标评估）

---

## ⚠️ 重要说明

本计划的**核心目标**已分解为以下下游计划执行：

| 执行环节 | 对应计划 | 说明 |
|---------|--------|------|
| **前置条件** | Plan 210 | 数据库基线重建（保证迁移幂等性） |
| **核心执行** | Plan 211 | Phase1 模块统一化（Day1-10） |
| **架构审查** | Plan 212 | 共享代码复用决策与目录标准化（Day6-7） |

**本文档保留的参考价值**：
- **第一部分**：问题诊断与现状分析（为什么需要 go.mod 统一化）
- **第四部分**：详细的风险应对指南（症状→原因→解决方案）
- **附加资源**：常用命令速查表（开发者手册）

若需了解**最新执行进度**，请参见 [Plan 211 - Phase1 模块统一化实施方案](./211-phase1-module-unification-plan.md)。

---

## 概述

本文档提供了从当前项目架构（多个独立 go.mod）向目标架构（统一的模块化单体）过渡的**具体操作指南**。包括：

- 当前项目的详细分析
- go.mod 统一化的步骤
- 代码迁移的具体方法
- 风险识别与应对
- 验证与回滚计划

---

## 第一部分：当前项目的详细分析

### 1.1 当前 go.mod 的混乱状态

#### 主 go.mod
```go
module cube-castle-deployment-test
```
**问题**：模块名与项目名不符，容易产生困惑

#### organization-command-service go.mod
```
/cmd/organization-command-service/go.mod:
module organization-command-service
```
**问题**：独立的模块，与主模块不一致

#### organization-query-service go.mod
```
/cmd/organization-query-service/go.mod:
module cube-castle-deployment-test/cmd/organization-query-service
```
**问题**：与主模块不一致，而且路径冗长

#### 影响

- 两个服务难以共享代码
- 无法在 `internal/` 中创建共享包，被所有服务复用
- 新模块的集成变得复杂
- 项目整体显得不专业

### 1.2 当前的代码重复

```
/cmd/organization-command-service/internal/
├── auth/              # 认证逻辑副本
├── cache/             # 缓存管理副本
├── config/            # 配置管理副本
└── ...

/cmd/organization-query-service/internal/
├── auth/              # 认证逻辑副本
├── ...

/internal/            # 全局 internal（被忽视）
├── auth/              # 实际使用的认证逻辑
└── cache/             # 实际使用的缓存逻辑
```

**影响**：代码维护困难，容易产生不一致

### 1.3 当前的项目结构问题

```
/cube-castle/
├── cmd/
│   ├── organization-command-service/  # 独立二进制
│   │   ├── main.go
│   │   ├── go.mod                    # ❌ 独立模块
│   │   └── internal/                 # ❌ 独立的内部包
│   ├── organization-query-service/   # 独立二进制
│   │   ├── main.go
│   │   ├── go.mod                    # ❌ 独立模块（或路径冗长）
│   │   └── internal/                 # ❌ 独立的内部包
│   └── oauth-service/
├── internal/                         # ⚠️ 被忽视，没有被充分利用
│   ├── auth/
│   ├── cache/
│   └── ...
├── pkg/
│   └── health/
└── go.mod                            # 主模块（混乱的模块名）
```

---

## 第二部分：go.mod 统一化的详细步骤

### 步骤 1：分支准备

```bash
# 在新分支上进行试验
git checkout -b feat/unify-go-modules

# 创建备份分支（以防需要回滚）
git checkout -b backup/modular-before-unify
git checkout feat/unify-go-modules
```

### 步骤 2：修改主 go.mod

#### 当前状态
```go
// go.mod
module cube-castle-deployment-test

go 1.21
```

#### 目标状态
```go
// go.mod
module cube-castle

go 1.21

require (
    github.com/... // 所有第三方依赖
)
```

#### 操作
```bash
# 修改文件
vi go.mod

# 将 module 名改为：cube-castle
# 保留所有 require 语句
# 删除 replace 语句（如果有）
```

### 步骤 3：删除冗余 go.mod

#### organization-command-service

```bash
# 删除这个文件
rm /cmd/organization-command-service/go.mod
rm /cmd/organization-command-service/go.sum

# 验证：该服务现在被视为主模块的子包
# 其导入路径变为：cube-castle/cmd/organization-command-service
```

#### organization-query-service

```bash
# 删除这个文件
rm /cmd/organization-query-service/go.mod
rm /cmd/organization-query-service/go.sum

# 验证：该服务现在被视为主模块的子包
# 其导入路径变为：cube-castle/cmd/organization-query-service
```

### 步骤 4：合并依赖

```bash
# 整合所有依赖到主 go.mod
go mod tidy

# 这个命令会：
# 1. 删除未使用的依赖
# 2. 添加缺失的依赖
# 3. 整理 go.mod 和 go.sum
```

### 步骤 5：验证导入路径

#### 在 organization-command-service 中，检查是否有需要调整的导入

```bash
# 查看所有导入
grep -r "import" /cmd/organization-command-service/main.go

# 应该看到的导入：
# "cube-castle/cmd/organization-command-service/..."
# "cube-castle/internal/auth"
# 不应该看到：
# "organization-command-service/..."
```

#### 如果存在旧的导入路径，需要更新

```bash
# 在整个项目中搜索和替换
grep -r "import.*organization-command-service" ./

# 手动更新为：
# "cube-castle/cmd/organization-command-service/..."
```

### 步骤 6：调整代码结构（可选但推荐）

#### 当前结构
```
/cmd/organization-command-service/
├── main.go
├── internal/
│   ├── auth/
│   ├── cache/
│   ├── handlers/
│   ├── repository/
│   ├── services/
│   └── ...
```

#### 推荐结构（立即调整）
```
/cmd/organization-command-service/
├── main.go
├── handlers/
├── models/
└── ...

/cmd/organization-query-service/
├── main.go
├── resolvers/
├── models/
└── ...

/internal/              # 共享代码
├── auth/
├── cache/
├── config/
├── organization/       # 🆕 organization 模块即将在这里
│   ├── api.go
│   └── internal/
└── ...
```

**或者（分阶段调整，风险更低）**：
```
# 暂时保持原结构，只是去掉独立的 go.mod
# 在第二阶段（Week 3-4）再进行结构重构
```

### 步骤 7：编译验证

```bash
# 主程序编译
go build -v ./cmd/organization-command-service

# 应该看到：
# cube-castle/cmd/organization-command-service

# 查询服务编译
go build -v ./cmd/organization-query-service

# 应该看到：
# cube-castle/cmd/organization-query-service

# 完整编译（包括所有包）
go build ./...

# 查看所有模块
go list ./...
# 应该看到：
# cube-castle
# cube-castle/cmd/organization-command-service
# cube-castle/cmd/organization-query-service
# cube-castle/internal/auth
# cube-castle/internal/cache
# ...
```

### 步骤 8：运行测试

```bash
# 运行所有测试
go test ./... -v

# 统计测试覆盖率
go test ./... -cover

# 运行 race detector（检测并发竞态）
go test ./... -race
```

### 步骤 9：本地验证

```bash
# 启动命令服务
./bin/organization-command-service

# 在另一个终端启动查询服务
./bin/organization-query-service

# 测试 API
curl http://localhost:9090/health
curl http://localhost:8090/health

# 测试功能
curl -X GET http://localhost:9090/org/organizations
```

### 步骤 10：同步基础设施标准

1. **数据库连接池显式配置**
   ```go
   // cmd/organization-command-service/main.go
   db.SetMaxOpenConns(25)
   db.SetMaxIdleConns(5)
   db.SetConnMaxIdleTime(5 * time.Minute)
   db.SetConnMaxLifetime(30 * time.Minute)
   ```
   `organization-query-service` 应复用相同配置，确保跨服务一致。

2. **引入 sqlc/atlas/goose 流程**
   ```bash
   make sqlc-generate      # 生成仓储代码
   make db-migrate-verify  # goose up/down + atlas diff 预演
   ```
   任何迁移 MR 必须携带 `-- +goose Down` 与更新后的 `atlas.hcl`。

3. **事务性发件箱基座**
   - 在 `pkg/database/outbox`（或等效目录）创建共享 `InsertEvent`/`FetchUnpublished` 封装，统一 `event_id` 生成逻辑。
   - 命令服务启动时注册 outbox relay，并将发布失败的事件增加 `retry_count`。

4. **Docker 集成测试基线**
   - 添加 `docker-compose.test.yml` 启动 PostgreSQL。
   - 将 `make test-db` (或 `go test -tags integration`) 纳入本地和 CI 的默认检查。

---

## 第三部分：代码迁移与清理

### 3.1 共享代码的提取

#### 认证代码（auth）

**当前状态**：在两个服务的 internal/ 中各有一份副本

```bash
# 查看 cmd/organization-command-service/internal/auth/
ls -la cmd/organization-command-service/internal/auth/
# 可能看到：jwt.go, middleware.go, pbac.go 等

# 查看 cmd/organization-query-service/internal/auth/
ls -la cmd/organization-query-service/internal/auth/
# 可能看到：jwt.go, middleware.go 等

# 查看全局 internal/auth/
ls -la internal/auth/
# 可能看到：jwt.go, middleware.go, pbac.go, graphql_middleware.go
```

**迁移步骤**：

1. 确认全局 `/internal/auth/` 是最完整的版本
2. 更新两个服务，使它们导入全局 auth：
   ```go
   // 在 cmd/organization-command-service/main.go 中
   import "cube-castle/internal/auth"

   // 不再导入
   // import "./internal/auth"  ❌
   ```
3. 删除服务内的 auth 副本：
   ```bash
   rm -rf cmd/organization-command-service/internal/auth
   rm -rf cmd/organization-query-service/internal/auth
   ```
4. 验证编译
   ```bash
   go build ./cmd/organization-command-service
   ```

#### 缓存代码（cache）

同样的过程：

```bash
# 确认全局 /internal/cache/ 或 /pkg/cache/ 是最完整的
ls -la internal/cache/ pkg/cache/

# 更新导入路径
# 删除服务内的副本
rm -rf cmd/organization-command-service/internal/cache
rm -rf cmd/organization-query-service/internal/cache

# 验证编译
go build ./...
```

#### 其他共享代码

按相同方式处理 config/, types/, middleware/ 等。

#### 数据访问脚手架（sqlc）

1. 在 `configs/sqlc/sqlc.yaml` 创建或更新 sqlc 配置，指向 `internal/` 与 `database/queries/`。
2. 提供至少一个示例查询（如 `internal/organization/repository/queries.sql`），运行：
   ```bash
   make sqlc-generate
   ```
3. 将生成的包（例：`internal/organization/repository/sqlc`）替换手写 `Scan` 逻辑，确保编译通过。
4. 在 CI 中新增步骤：
   ```yaml
   - name: Generate sqlc code
     run: make sqlc-generate
   - name: Verify no diff
     run: git diff --exit-code
   ```

#### 事务性发件箱基础设施

1. 在 `pkg/database/outbox` 创建复用组件（生成 `event_id`、插入、轮询、重试）。
2. 命令服务与未来模块使用统一 Outbox API：
   ```go
   outbox.Insert(ctx, tx, outbox.Event{
       EventID: uuid.New(),
       AggregateID: empID,
       AggregateType: "employee",
       EventType: "employee.terminated",
       Payload: payload,
   })
   ```
3. Relay 使用公共封装读取批次，并在发布成功后标记 `published_at`，失败时自增 `retry_count`。
4. 在 `pkg/metrics` 内注册 `outbox_unpublished_events_total`、`outbox_retry_total` 指标。

### 3.2 清理 internal/ 目录

**目标**：使 `/internal/` 真正成为全局共享包

```
/internal/
├── auth/                    # 认证/授权（共享）
├── cache/                   # 缓存管理（共享）
├── config/                  # 配置管理（共享）
├── middleware/              # 中间件（共享）
├── types/                   # 共享类型定义
├── graphql/                 # GraphQL 工具（共享）
├── organization/            # 🆕 organization 模块
│   ├── api.go               # 公开接口
│   └── internal/            # 模块内部实现
│       ├── service/
│       ├── repository/
│       ├── handler/
│       ├── resolver/
│       └── domain/
└── workforce/               # 🆕 workflow 模块（未来）
    ├── api.go
    └── internal/
        └── ...
```

### 3.3 Docker 集成测试基座

1. 在仓库根目录添加 `docker-compose.test.yml`：
   ```yaml
   version: "3.8"
   services:
     postgres:
       image: postgres:15
       environment:
         POSTGRES_PASSWORD: password
         POSTGRES_USER: cube
         POSTGRES_DB: cube
       ports:
         - "6543:5432"
   ```
2. 创建 `scripts/testdb/wait-for-postgres.sh` 或使用 `docker compose run` 等待数据库就绪。
3. 新增 Make 目标：
   ```make
   test-db:
       docker compose -f docker-compose.test.yml up -d postgres
       goose -dir database/migrations postgres "$$TEST_DB_DSN" up
       go test ./tests/integration/... -tags integration
       goose -dir database/migrations postgres "$$TEST_DB_DSN" down
       docker compose -f docker-compose.test.yml down
   ```
4. 在 QA 与 CI 流程中执行 `make test-db`，确保真实数据库路径完全覆盖关键仓储。

### 3.4 更新导入路径

#### 在所有文件中查找旧导入

```bash
# 查找所有可能的旧导入
grep -r "organization-command-service" ./cmd/
grep -r "organization-query-service" ./cmd/

# 或搜索相对导入（可能导致问题）
grep -r "^\\./" ./cmd/
```

#### 替换导入路径

```bash
# 使用 sed 或编辑器进行替换
# 例如，在 organization-command-service 中：
sed -i 's|"organization-command-service/|"cube-castle/cmd/organization-command-service/|g' cmd/organization-command-service/**/*.go

# 验证替换结果
grep -r "import" cmd/organization-command-service/main.go | head -20
```

---

## 第四部分：风险识别与应对

### 4.1 高风险项

#### 风险 1：编译失败

**症状**：
```
go build ./...
error: package "organization-command-service" not found
```

**原因**：
- 导入路径未正确更新
- 依赖缺失
- go.mod 文件损坏

**应对**：
```bash
# 1. 检查 go.mod 语法
go mod validate

# 2. 检查所有导入
grep -r "import" ./cmd/ | grep -v "cube-castle"

# 3. 清理缓存并重新下载
go clean -modcache
go mod tidy

# 4. 如果还是不行，回滚到备份分支
git checkout backup/modular-before-unify
```

#### 风险 2：功能缺失

**症状**：
- API 返回 404
- GraphQL Query 不工作
- 认证失败

**原因**：
- 代码未完全迁移
- 配置文件指向错误的路径
- 数据库初始化失败

**应对**：
```bash
# 1. 检查服务是否正常启动
./bin/organization-command-service &
# 如果报错，检查日志

# 2. 查看服务日志
tail -f /var/log/hrms/*.log

# 3. 验证数据库连接
psql -h localhost -U user -d cubecastle -c "\dt"

# 4. 如果问题无法解决，回滚
git reset --hard HEAD
```

#### 风险 3：性能下降

**症状**：
- 响应时间明显增加
- CPU 使用率上升
- 内存使用量增加

**原因**：
- 导入路径变长，编译优化不同
- 缓存失效
- 依赖版本变化

**应对**：
```bash
# 1. 进行性能对比测试
./scripts/performance-test.sh

# 2. 分析热点
go tool pprof -http=:8080 cpu.prof

# 3. 优化导入（延迟加载）
# 4. 如果无法快速解决，考虑回滚
```

#### 风险 4：迁移回滚或集成测试缺失

**症状**：
- goose `down` 执行失败或缺失
- Docker 集成测试长时间未运行，CI 未覆盖
- 发布后数据库无法回滚

**原因**：
- down.sql 未补齐或与 up.sql 不匹配
- `make db-migrate-verify` 未执行
- `docker-compose.test.yml` 环境缺失或配置错误

**应对**：
```bash
make db-migrate-verify   # 重新运行 up/down/atlas diff
make test-db             # 在 Docker PostgreSQL 中跑集成测试
git status               # 确认 sqlc 生成代码已提交
```
若以上命令失败，禁止合并并回退到备份分支排查。

### 4.2 中风险项

#### 风险 5：部分服务不可用

**症状**：
- 命令服务可用，查询服务不可用（或反之）
- 某些 API 端点返回 500

**应对**：
```bash
# 1. 检查该服务的启动日志
docker logs organization-query-service

# 2. 检查依赖是否完整
go mod verify

# 3. 逐个验证该服务的模块
go build -v ./cmd/organization-query-service

# 4. 如果是特定功能，可能是导入路径问题
grep -n "import" cmd/organization-query-service/internal/**/*.go
```

#### 风险 6：第三方库冲突

**症状**：
```
go mod tidy 出错：
conflict: github.com/some-lib requires version 1.0, but 2.0 is already required
```

**应对**：
```bash
# 1. 查看依赖树
go mod graph | grep some-lib

# 2. 检查哪些包需要这个库
grep -r "some-lib" ./

# 3. 更新相关包版本
go get -u github.com/some-lib@latest

# 4. 清理和验证
go mod tidy
go mod verify
```

### 4.3 低风险项

#### 风险 7：文档过时

**症状**：
- README.md 中的导入路径示例不正确
- 开发指南提到的目录不存在

**应对**：
```bash
# 1. 更新所有文档
find ./docs -name "*.md" -exec grep -l "organization-command-service" {} \;

# 2. 批量替换
sed -i 's|organization-command-service|cube-castle/cmd/organization-command-service|g' docs/**/*.md

# 3. 手动审查关键文档
```

---

## 第五部分：验证清单

### 5.1 编译验证

```bash
□ go build ./cmd/organization-command-service 成功
□ go build ./cmd/organization-query-service 成功
□ go build ./... 成功
□ go mod verify 无错误
□ go mod graph 中无循环依赖
```

### 5.2 测试验证

```bash
□ go test ./... 全部通过
□ 测试覆盖率 > 70%
□ go test -race ./... 无竞态条件
□ make sqlc-generate && git diff --exit-code
□ make db-migrate-verify 成功（goose up/down + atlas diff）
□ make test-db 成功（Docker PostgreSQL 集成测试）
□ 所有集成测试通过
```

### 5.3 功能验证

```bash
□ REST API /org/organizations 返回正确数据
□ GraphQL query organizations 返回正确数据
□ 认证系统正常工作
□ 缓存系统正常工作
□ 错误处理正常
□ 命令/查询服务均显式设置连接池参数
□ Outbox relay 正常运行（事件已入库并发布）
□ Prometheus 暴露 outbox/数据库指标
```

### 5.4 性能验证

```bash
□ 平均响应时间与之前相同（±10%）
□ CPU 使用率与之前相同（±10%）
□ 内存使用量与之前相同（±10%）
□ 并发请求处理正常
```

### 5.5 部署验证

```bash
□ Docker 镜像编译成功
□ Docker 容器启动正常
□ 容器内服务正常运行
□ 所有健康检查通过
□ docker-compose -f docker-compose.test.yml up/down 正常
```

---

## 第六部分：回滚计划

### 6.1 快速回滚（如果问题在 1-2 小时内发现）

```bash
# 情况 1：代码尚未提交
git reset --hard HEAD
# 或恢复到特定提交
git checkout abc123def

# 情况 2：代码已提交但未推送
git reset --soft HEAD~1
git checkout .

# 情况 3：代码已推送（生产环境有问题）
git revert commit-hash
git push
```

### 6.2 分阶段回滚（如果问题在生产发现）

```bash
# 步骤 1：标记为不稳定
git tag -a v-unstable-module-merge

# 步骤 2：切回上一个稳定版本
git checkout tag/v4.7.0

# 步骤 3：打包并部署旧版本
make docker-build VERSION=4.7.0-rollback
docker push cube-castle:4.7.0-rollback

# 步骤 4：更新运行环境
docker-compose down
docker pull cube-castle:4.7.0-rollback
docker-compose up -d
```

### 6.3 问题分析与修复

```bash
# 如果只是小问题，尝试快速修复：
# 1. 识别问题
# 2. 在新分支修复
git checkout -b fix/module-merge-issue

# 3. 修复并测试
# 4. 提交和部署
git commit -m "fix: 解决 go.mod 合并后的xxx问题"
git push origin fix/module-merge-issue
# 提交 PR 进行审查
```

---

## 第七部分：成功指标

| 指标 | 目标值 | 检查方法 |
|------|--------|--------|
| 编译通过率 | 100% | `go build ./...` |
| 测试通过率 | 100% | `go test ./...` |
| 测试覆盖率 | > 70% | `go test ./... -cover` |
| 迁移可回滚率 | 100% | `make db-migrate-verify` |
| 集成测试通过率 | 100% | `make test-db` |
| 功能完整性 | 100% | 逐个验证 API |
| 性能对比 | ±10% | 基准测试对比 |
| 首次成功部署 | 一次 | 部署日志 |

---

## 第八部分：附加资源

### 常用命令速查表

```bash
# go.mod 管理
go mod init cube-castle              # 初始化新模块
go mod tidy                          # 整理依赖
go mod verify                        # 验证 go.mod
go mod graph                         # 查看依赖树
go mod download                      # 下载依赖
go clean -modcache                   # 清理缓存

# 编译与测试
go build -v ./cmd/...               # 详细编译
go test ./... -v                    # 详细测试
go test ./... -cover                # 测试覆盖率
go test -race ./...                 # 竞态检测
go test -bench ./...                # 基准测试

# 代码分析
go vet ./...                        # 静态分析
go fmt ./...                        # 代码格式
golint ./...                        # 代码规范

# Docker 操作
docker build -t cube-castle:latest .
docker run -p 9090:9090 -p 8090:8090 cube-castle:latest
docker compose up -d
docker compose logs -f

# 数据库与代码生成
make db-migrate-verify               # goose up/down + atlas diff
make sqlc-generate                   # 生成类型安全仓储代码
make test-db                         # 在 Docker PostgreSQL 中运行集成测试
```

---

**文档版本历史**:
- v1.0 (2025-11-03): 初始版本，详细的 go.mod 统一化过渡方案
