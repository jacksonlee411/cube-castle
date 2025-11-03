# Phase1 模块统一化执行日志

## Day1 · 2025-11-03

### Kick-off 核对
- 执行负责人：Codex（全栈） — 依据 `docs/development-plans/211-phase1-module-unification-plan.md` 与 06 号评审报告确认执行权与范围。
- 当前分支：`feature/204-phase1-unify`（HEAD `c714c16b`）。
- 同步状态：已完成计划复评，无新增契约变更；执行窗口锁定 Week1-2。

### 资产盘点与模块现状
- go.mod / go.work 清单（通过 `find . -name go.mod` 验证）：
  - `go.mod`
  - `cmd/hrms-server/command/go.mod`
  - `cmd/hrms-server/query/go.mod`
  - `pkg/health/go.mod`
  - `shared/go.mod`
  - `go.work`
- `node scripts/generate-implementation-inventory.js` 于 2025-11-02T23:21:10Z 执行，刷新 `reports/implementation-inventory.json` 供后续差异比对。
- 未检测到新的宿主服务依赖或端口冲突，Docker 使用前置条件保持有效。

### 今日待办与后续跟踪
- Day1 动作完成：Kick-off 确认、资产盘点脚本执行。
- 持续跟踪项：Day2 输出模块命名锁定结论与分支策略说明；Day3 准备 go.mod 合并策略草案。

## Day2 准备草案（初稿）

### 模块命名现状
| 目录 | 当前 module 声明 |
|------|------------------|
| `go.mod` | `module cube-castle-deployment-test` |
| `cmd/hrms-server/command/go.mod` | `module organization-command-service` |
| `cmd/hrms-server/query/go.mod` | `module cube-castle-deployment-test/cmd/hrms-server/query` |
| `pkg/health/go.mod` | `module cube-castle-deployment-test/pkg/health` |
| `shared/go.mod` | `module shared` |

- 决议：统一将根模块命名为 `module cube-castle`，所有子模块依赖归并至单一 go.mod；各服务在目录迁移完成前临时使用 `cube-castle/cmd/<service>` 包路径。（详见 `docs/development-plans/211-Day2-Module-Naming-Record.md`）
- 结果记录：本决议为 Day2 命名基线，后续执行情况同步 `docs/development-plans/211-phase1-module-unification-plan.md` 与本日志。

### 命名与合并策略要点
- 目标命名：根模块统一为 `module cube-castle`，子模块依赖通过单一 go.mod 管理。
- Day3 将输出 `go.mod` 合并策略草案，按“根模块保留 + 子模块替换为相对路径引用”思路编排迁移顺序。
- 子模块内引用将统一切换为根模块包路径，避免形成平行事实来源。

### 分支策略（初稿）
- 主执行分支保持 `feature/204-phase1-unify`，所有阶段性提交保持线性历史，必要时使用临时工作分支（命名 `feature/204-phase1-unify/<task>`）承载大规模 refactor。
- 合并至 `feature/204-phase1-unify` 前需通过 `make fmt`、`go test ./...`、关键 CI 作业，保持分支可随时回滚。

## Day3 go.mod 合并策略草案

### 任务拆解顺序
1. **重命名根模块**：`go.mod` 中 `module cube-castle-deployment-test` → `module cube-castle`，同步更新 `go` 版本为计划指定的 `go 1.22.x`（若偏差需先与 Plan 204 对齐）。
2. **整合 shared 模块**：将现有 `shared/go.mod` 中依赖合并至根模块；移除 `replace shared => ./shared`，统一采用 `cube-castle/shared/...` 导入路径。
3. **逐步回收子模块**：
   - 删除 `cmd/hrms-server/command/go.mod`，在根模块 `require` 中补齐所需依赖。
   - 删除 `cmd/hrms-server/query/go.mod`，同步处理其 `replace cube-castle-deployment-test/pkg/health => ../../pkg/health`。
   - 删除 `pkg/health/go.mod`，统一改为包路径 `cube-castle/pkg/health`。
4. **批量更新导入路径**：
   - `organization-command-service/...` → `cube-castle/cmd/hrms-server/command/...`
   - `cube-castle-deployment-test/cmd/hrms-server/query/...` → `cube-castle/cmd/hrms-server/query/...`
   - `cube-castle-deployment-test/internal/...` → `cube-castle/internal/...`
   - `cube-castle-deployment-test/pkg/health` → `cube-castle/pkg/health`
   - `shared` → `cube-castle/shared`
5. **移除 go.work**：在上述依赖调整完成且 `go list ./...` 通过后删除 `go.work`。
6. **整理 go.sum**：运行 `go mod tidy`，确保 `go.sum` 与依赖列表一致；记录差异用于后续审计。

### 依赖映射与兼容性检查
| 原 module path | 目标 module path | 校验要点 |
|----------------|------------------|----------|
| `organization-command-service/...` | `cube-castle/cmd/hrms-server/command/...` | 确认命令服务内部 import 全量替换；`gofmt` 后确保构建通过 |
| `cube-castle-deployment-test/cmd/hrms-server/query/...` | `cube-castle/cmd/hrms-server/query/...` | GraphQL 层引用与测试脚本需同步调整 |
| `cube-castle-deployment-test/internal/...` | `cube-castle/internal/...` | 关注共享中间件、配置、鉴权代码的包路径一致性 |
| `cube-castle-deployment-test/pkg/health` | `cube-castle/pkg/health` | 确保 Playwright/运维脚本引用更新 |
| `shared` | `cube-castle/shared` | `goimports` 后确认无循环依赖 |

### 校验与辅助脚本计划
- 快速替换：使用 `rg` + `sd`/`perl -pi -e`（或编写临时脚本 `scripts/migrations/rename-module-paths.sh`）批量修改导入路径；执行前后保留 diff。
- 验证命令：
  - `go list ./...`（确保所有包可枚举）
  - `go test ./cmd/hrms-server/command/...`
  - `go test ./cmd/hrms-server/query/...`
  - `go env GOPRIVATE`（确认无需额外配置，如需私有模块在 Day3 记录）
- 记录：迁移完成后在本日志追加执行摘要，并更新 06 号评审日志的 P0#1/Day4 进展。

## Day4-5 并行迁移与 CI 清理 Checklist（草案）

### 轨道A：命令服务迁移
- 将 `cmd/hrms-server/command` 入口与内部包移动到 `cmd/hrms-server/command`（保留 CQRS 边界）。
- 更新 `main.go`、`internal/*` 下 import，确认编译通过。
- 编写一次性脚本（若必要）搬运文件，执行后标记为 `// TODO-TEMPORARY:` 并记录回收计划。

### 轨道B：查询服务迁移
- 同步迁移 GraphQL 入口至 `cmd/hrms-server/query`，同步更新生成脚本与 resolver 引用。
- 运行 `make generate-graphql`（如存在）并比较 `schema.graphql` 输出，确保契约未改动。

### 轨道C：DevOps/CI
- 更新 `Makefile`、`docker-compose*.yml`、`Dockerfile` 中的模块路径与二进制位置。
- 清理 `.github/workflows/*` 中 Neo4j 相关步骤；统一 Go 版本至 `1.22.x`。
- 执行 `make test`、`make build`、`docker compose -f docker-compose.dev.yml build hrms-server`，记录日志至 `reports/phase1-module-unification.md`。

### 统一验收
- `go test ./...`、`npm run lint`、`npm run test`（前端受影响时）全部通过。
- `curl http://localhost:9090/health` 与 `curl http://localhost:8090/health` 返回 200。
- CI 绿灯截图及日志归档至 `reports/phase1-module-unification.md` 与 `reports/phase1-regression.md`。

## Day3 执行记录（go.mod 合并 & 校验）

- **路径调整**：批量将旧模块引用替换为 `cube-castle/...`，涵盖命令/查询服务、共享与内部工具（脚本：`python3 - <<'PY' ...`）。
- **模块统一**：
  - 删除 `cmd/hrms-server/command/go.mod`、`cmd/hrms-server/query/go.mod`、`pkg/health/go.mod`、`shared/go.mod` 及关联 `go.sum` / `go.work`。
  - 根 `go.mod` 重命名为 `module cube-castle`，聚合原子依赖并固定 `github.com/graph-gophers/graphql-go v1.5.0` 等关键版本。
- **Go 版本基线**：因 `github.com/jackc/pgx/v5 v5.7.5`、`github.com/gin-gonic/gin v1.9.1` 等依赖要求 Go ≥1.23，当前统一至 `go 1.24.0`（toolchain `go1.24.9`）。已在 06 号评审日志登记为风险说明，后续与 Plan 204/Steering 协调是否调整官方基线。
- **校验命令**：
- `go mod tidy`（完成依赖归并及 go.sum 更新）。
- `go list ./...`（成功枚举 27 个包，含命令/查询服务、shared/config、tests 等）。
- `go test ./...` — **结果**：内部包 `cube-castle/internal/auth` 由于预存的 512-bit 测试密钥触发安全检查而失败（Day5 已替换为 2048-bit 密钥并恢复通过）；其余包测试通过。
- **执行证据**：详见 `go.mod`、`go.sum` 最新 diff，命令输出保留在本地终端记录；Day3 完成情况同步 `docs/development-plans/06-integrated-teams-progress-log.md`。

## Day5 执行记录（服务目录迁移 & CI 清理）

- **目录迁移**：
  - 将命令/查询服务分别迁移至 `cmd/hrms-server/command/` 与 `cmd/hrms-server/query/`，保留原有 `internal/` 结构与开发脚本。
  - 全局更新路径引用（Makefile、测试脚本、Docker Compose、质量脚本与关键文档），确保新目录为唯一事实来源。
  - 命令与查询服务 Dockerfile 切换至 `golang:1.24-alpine`，直接在根模块下载依赖并构建。
- **CI / 工具链**：
  - `.github/workflows/ci.yml`、`test.yml` 移除 Neo4j 服务，新增 Redis 依赖，统一 Go 版本至 1.24，并补充前端 `npm run lint` / `npm run test` 检查。
  - `.golangci.yml`、`scripts/quality/lint-validation.js` 等策略工具同步新路径约束。
- **验证记录**：
  - `go test ./cmd/hrms-server/command/... ./cmd/hrms-server/query/...` ✅
  - `go test ./...` ✅（升级 `internal/auth/auth_test.go` 内部 RSA 测试密钥至 2048-bit，消除安全校验失败）
  - `npm run lint` ⚠️（历史遗留：`frontend/src/features/positions/timelineAdapter.ts` CamelCase 命名与 Storybook 配置缺失；保留为后续修复项）
- **未决事项**：
  - 完成前端命名/Storybook 配置整改，恢复 Lint 绿灯。
  - 按 Day6-7 计划继续推进共享代码抽取与架构审查。
