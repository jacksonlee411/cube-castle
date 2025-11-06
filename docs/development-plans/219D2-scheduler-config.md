# Plan 219D2 – 调度配置集中化与启动流程更新

**文档编号**: 219D2  
**关联路线图**: Plan 219 → 219D  
**依赖子计划**: 219D1 完成代码迁移；219A 目录结构约束  
**目标周期**: Week 4 Day 21（219D 并行阶段）  
**负责人**: 后端团队（配置 Owner）

---

## 1. 目标

1. 汇总调度相关配置（cron 表达式、队列名称、重试策略、worker 并发度），统一落在配置包与 `.env`/YAML，并在 `internal/organization/README.md` 的 Scheduler 章节维护唯一事实来源链接；盘点对象需至少覆盖：`internal/organization/scheduler/operational_scheduler.go`、`internal/organization/scheduler/temporal_monitor.go`、`internal/organization/api.go`、`cmd/hrms-server/command/main.go`、`docker-compose.dev.yml`、`.env.example`、`Makefile`。
2. 更新启动链路（Makefile、`make run-dev`、Docker Compose env）以确保 Scheduler 默认启用且可调试，执行前确认 219D1 已完成目录/依赖注入迁移。
3. 建立配置变更的校验流程（默认值说明 + 变更 checklist），并将校验过程与回滚记录纳入 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 对应章节。

---

## 2. 范围

| 模块 | 内容 |
|------|------|
| 配置项归集 | 抓取所有 scheduler/Temporal 相关配置，统一声明并去重，并在 `internal/organization/README.md#scheduler` 更新清单（含原始位置、目标配置键、覆盖层级） |
| 启动流程 | Makefile、`cmd/hrms-server/command/main.go`、`config/` 中的初始化参数及 `docker-compose*.yml` 的环境变量映射 |
| 检查机制 | 添加配置校验/日志，记录默认值及覆盖来源，并在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 记录审计/回滚步骤及常见报错处理 |

不包含：代码迁移（219D1）、指标埋点（219D3）、测试与文档编写（219D4/219D5）。

---

## 3. 详细任务

1. **配置盘点**
   - 遍历 `internal/organization/scheduler/operational_scheduler.go`、`internal/organization/scheduler/temporal_monitor.go`、`internal/organization/api.go`、`cmd/hrms-server/command/main.go`、`config/`、`.env*`、`docker-compose.*`、`Makefile`，搜索 `cron`、`temporal`、`scheduler` 等关键字。
   - 形成表格：参数名称、当前位置、默认值、用途、依赖调用方、当前覆盖方式（环境变量/常量/脚本），并将最终表格纳入 PR 附录，同时同步摘要至 `internal/organization/README.md#scheduler`。
   - 对比 219D1 迁移清单，确认无遗留配置项仍在旧目录；如发现临时脚本（例如 `cmd/hrms-server/command/scripts/setup-cron.sh`）引用环境变量，需要标注后续移除或兼容策略。

2. **集中化实现**
   - 在 `internal/config/` 中新增 `scheduler.go` 与 `scheduler_validator.go`（与项目现有 `jwt.go` 风格一致），定义：
     ```go
     type SchedulerConfig struct {
       Enabled          bool
       TemporalEndpoint string
       Namespace        string
       TaskQueue        string
       Worker           struct {
         Concurrency int
         PollerCount int
       }
       CronJobs map[string]CronDefinition
       Retry    RetryPolicy
     }
     ```
     `CronDefinition` 至少包含 `Schedule string`、`InitialDelay time.Duration`、`Enabled bool`；`RetryPolicy` 包含 `MaxAttempts`, `InitialInterval`, `BackoffCoefficient`, `MaxInterval`，默认值遵循 Temporal 官方推荐（3 次、初始 1s、指数 2.0、最大 1m）。
   - 定义环境变量映射（前缀 `SCHEDULER_`）：例如 `SCHEDULER_ENABLED`、`SCHEDULER_TEMPORAL_ENDPOINT`、`SCHEDULER_NAMESPACE`、`SCHEDULER_TASK_QUEUE`、`SCHEDULER_WORKER_CONCURRENCY` 等；支持 `.env`、Docker Compose、Kubernetes（后续）三层读取，默认值写入 `.env.example` 并在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 标注覆盖方式。
   - 更新 `cmd/hrms-server/command/main.go` 与 219D1 新 Facade（`internal/organization/scheduler`）读取 `config.GetSchedulerConfig()`，替换所有散落的 `os.Getenv`、硬编码 cron 表达式及默认路径。
   - 若需保留 YAML 或 JSON 清单，统一放置于 `config/scheduler.yaml` 并在 config 包加载（遵循现有 config 体系的查找顺序：env > YAML > 默认值）。

3. **启动流程对齐**
   - 在确认 219D1 已合入并稳定（目录、依赖注入完成）后，再调整 Makefile 目标（如 `make run-dev`、`make run-scheduler`），确保 scheduler 默认启用：新增 `SCHEDULER_ENABLED ?= true`，本地调试通过 `make run-dev SCHEDULER_ENABLED=false` 覆盖，并与 219D1 Owner 联合评审 Facade 接入点，确认配置结构与新目录/模块职责一致。
   - 更新 `docker-compose.dev.yml`、`docker-compose.e2e.yml` 加入上述环境变量；删除或标记未来淘汰的宿主机 cron 脚本依赖（参考 `cmd/hrms-server/command/scripts/README.md` 的说明）。
   - 补充说明如何在本地禁用或覆盖参数（示例命令、环境变量列表），在 `internal/organization/README.md#scheduler` 中同步调试指引。

4. **配置校验与回滚**
   - 为关键参数在 `internal/config/scheduler_validator.go` 中添加启动时校验：检查 Temporal endpoint URL、TaskQueue 非空、Cron 表达式合法（使用 `github.com/robfig/cron/v3` 的 parser），并在日志中输出来源优先级和建议修复。
   - 新增单元测试 `internal/config/scheduler_test.go`、`internal/config/scheduler_validator_test.go`，覆盖默认值、环境变量覆盖、非法配置报错；纳入 `make test` 路径（最少执行 `go test ./internal/config/...`）。
   - 将校验失败的日志示例与回滚操作记录于 `logs/219D2/`，同时在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 的 Scheduler 配置小节登记 checklist，含“恢复默认值”“回退到 219D1 配置”的指引。

---

## 4. 验收标准

- [ ] 所有调度相关参数集中在单一配置入口，并在 `.env.example`、`internal/organization/README.md#scheduler` 中列明默认值与来源说明。
- [ ] `make run-dev`、Docker Compose 启动后 Scheduler 正常读取新配置，日志无报错，且 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 完成调试指引更新。
- [ ] 配置校验机制可在参数缺失或格式错误时阻止启动并给出指引，失败示例记录于 `logs/219D2/` 并在速查文档留存链接，相关单元测试 (`go test ./internal/config/...`) 全部通过并在验收记录中列出命令输出。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 遗漏散落配置项 | 中 | 使用脚本抓取关键字并与 219D1 迁移清单交叉验证，评审配置盘点表由 219D1 Owner 联合确认 |
| 配置改动影响已有环境 | 中 | 升级前通知平台团队，提供回滚默认值（`SCHEDULER_*` 环境变量）与 `.env` 变更说明；在 sandbox 先行验证并保留旧 `.env` 快照 |
| 启动脚本修改引入回归 | 中 | 运行 `make run-dev`、`make docker-up`、`make test` 全量验证，并增加 `make run-dev SCHEDULER_ENABLED=false` 验证可控关闭场景 |

---

## 6. 交付物

- 更新后的配置文件与 `.env.example`。
- 配置盘点表与检查清单（附于 PR，并同步摘要至 `internal/organization/README.md` 与 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`，包含旧值对照与变更审批记录）。
- 经验证的启动日志与失败示例（记录于 `logs/219D2/` 并在速查文档添加引用；附上 `go test ./internal/config/...` 输出与 `make run-dev` 日志关键片段）。
