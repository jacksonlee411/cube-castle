# 文档更新建议 – Plan 06 验收完成后

## 1. `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` – 补充项

### 新增部分：Scheduler 配置快速参考

```markdown
## Scheduler 配置（219D2 集中化）

### 环境变量清单

| 参数 | 默认值 | 说明 | 是否必填 |
|------|--------|------|---------|
| `SCHEDULER_ENABLED` | `true` | 启用/禁用 Scheduler | 否 |
| `SCHEDULER_TEMPORAL_ENDPOINT` | `temporal:7233` | Temporal 服务端点 | 是 |
| `SCHEDULER_NAMESPACE` | `cube-castle` | Temporal 命名空间 | 是 |
| `SCHEDULER_TASK_QUEUE` | `organization-maintenance` | 任务队列名称 | 是 |
| `SCHEDULER_WORKER_CONCURRENCY` | `4` | Worker 并发数 | 否 |
| `SCHEDULER_WORKER_POLLER_COUNT` | `2` | Activity poller 数量 | 否 |
| `SCHEDULER_RETRY_MAX_ATTEMPTS` | `3` | 重试最大次数 | 否 |
| `SCHEDULER_RETRY_INITIAL_INTERVAL` | `1s` | 初始重试间隔 | 否 |
| `SCHEDULER_RETRY_BACKOFF_COEFFICIENT` | `2.0` | 重试指数退避系数 | 否 |
| `SCHEDULER_RETRY_MAX_INTERVAL` | `1m` | 最大重试间隔 | 否 |

### 关闭 Scheduler（调试/禁用）

```bash
make run-dev SCHEDULER_ENABLED=false
# 或通过 .env 文件设置
echo "SCHEDULER_ENABLED=false" >> .env
make run-dev
```

### 验证 Scheduler 状态

```bash
# 检查启动日志中的 Scheduler 配置输出
docker logs cubecastle-rest-service | grep -i scheduler

# 检查健康状态端点
curl http://localhost:9090/health
```

### 常见问题

**Q: 如何临时禁用 Scheduler？**
A: 设置 `SCHEDULER_ENABLED=false` 环境变量后重启服务。

**Q: Temporal 连接失败怎么办？**
A: 确保 Docker Compose 已启动 Temporal 服务：`make docker-up`

**Q: 如何修改任务队列名称？**
A: 修改 `.env` 中的 `SCHEDULER_TASK_QUEUE` 并重启服务。
```

---

## 2. `internal/organization/README.md#scheduler` – 更新内容

### 新增或更新的部分

```markdown
## Scheduler 配置结构

### 配置加载层级

1. **默认值** – `internal/config/scheduler.go` 中定义的常量
2. **环境变量** – `.env` 或 `docker-compose*.yml` 中的 `SCHEDULER_*` 前缀配置
3. **运行时覆盖** – `make run-dev SCHEDULER_ENABLED=false` 命令行参数

### 配置验证

所有 Scheduler 配置在启动时通过 `config.ValidateSchedulerConfig()` 验证：
- 检查 Cron 表达式有效性
- 检查必填项（任务队列、Temporal 端点）
- 检查重试参数合理性

**验证失败将阻断启动**，错误日志见 `logs/219D2/config-validation.log`。

### 测试与验收

参考 Plan 06 验收记录：`docs/development-plans/06-integrated-teams-progress-log.md`

验收日志存档：`logs/219D2/ACCEPTANCE-RECORD-2025-11-06.md`

---

### Scheduler 服务集成点

**启动时初始化：**
```go
// cmd/hrms-server/command/main.go
scheduler := NewScheduler(cfg.SchedulerConfig)
scheduler.Start(ctx)
```

**关闭时清理：**
```go
scheduler.Stop(ctx)
```

**与 Temporal 的连接：**
- 端点：`SCHEDULER_TEMPORAL_ENDPOINT`
- 命名空间：`SCHEDULER_NAMESPACE`
- 任务队列：`SCHEDULER_TASK_QUEUE`
```

---

## 3. `CHANGELOG.md` – 版本记录更新

### 推荐条目

```markdown
### [Unreleased]

#### 配置管理 (Configuration)
- **Plan 06**: 调度器配置集中化（219D2）
  - 将所有 Scheduler 配置迁移至 `SCHEDULER_*` 命名空间前缀
  - 统一通过 `.env` 和 `docker-compose.dev.yml` 管理
  - 新增配置验证机制（`config.ValidateSchedulerConfig`）
  - 支持运行时通过环境变量覆盖关闭 Scheduler
  - 详见：`docs/development-plans/06-integrated-teams-progress-log.md`
  - 验收报告：`logs/219D2/ACCEPTANCE-RECORD-2025-11-06.md`

#### 测试与质量
- 新增 Scheduler 配置单元测试 ✅
- 新增启动自检与环境变量覆盖测试 ✅
- 新增失败恢复演练 ✅
```

---

## 4. 同步清单

- [ ] 更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
  - 添加 Scheduler 配置快速参考表
  - 添加常见问题 FAQ

- [ ] 更新 `internal/organization/README.md`
  - 详述配置加载层级
  - 说明验证机制
  - 指向验收报告

- [ ] 更新 `CHANGELOG.md`
  - 记录 Plan 06 的变更
  - 链接到验收报告和配置清单

- [ ] 存档 Plan 06 文档
  - 若计划已完成，移至 `docs/archive/`
  - 保留 `logs/219D2/` 作为永久参考

---

## 5. 后续验证

如需重复验收或环境升级，参考以下模板：

**验收步骤：**
```bash
# 1. 单元测试
go test ./internal/config/... -v

# 2. 启动验证
make run-dev

# 3. 关闭验证
make run-dev SCHEDULER_ENABLED=false

# 4. 失败演练
export SCHEDULER_TASK_QUEUE=""
make run-dev  # 应被阻断

# 5. 恢复
export SCHEDULER_TASK_QUEUE="organization-maintenance"
make run-dev
```

**日志位置：** `logs/219D2/`

