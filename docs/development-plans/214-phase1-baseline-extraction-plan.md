# 214-Phase1 Baseline Extraction Plan

**编号**: 214  
**标题**: 数据库基线重建 · Phase1 基线萃取执行方案  
**创建日期**: 2025-11-04  
**最近更新**: 2025-11-03  
**状态**: 🟢 已完成（执行时间 2025-11-03，提前于原 Week1 计划）  
**关联文档**:
- `210-database-baseline-reset-plan.md`（母计划）
- `212-shared-architecture-alignment-plan.md`（架构目录与复用决议）  
- `213-go-toolchain-baseline-plan.md`（Go 1.24 工具链基线确认）  
- `reports/phase1-regression.md`（Go 工具链验证记录）

---

## 1. 范围与目标

**范围**：落实 210 号计划中的 Phase1 基线萃取工作，聚焦以下事项：
1. 生成当前数据库 Schema 快照并比对差异。
2. 整理声明式 Schema（`database/schema.sql`）并形成唯一事实来源。
3. 基于 Schema 产出带 `-- +goose Up/Down` 标记的基线迁移。
4. 完成 DBA、架构组联合审阅与签字。

**目标**：在 Week1 结束前交付可复用的基线迁移文件，为 Phase2 目录重构与 Goose/Atlas 落地提供前提条件。  
**现状**：Plan 210 已于 2025-11-04 完成全部阶段（参见 `reports/PLAN-210-IMPLEMENTATION-VERIFICATION-REPORT.md`），Plan 214 现已获授权进入执行阶段，需在 2025-11-06 日间窗口内启动本计划并按 Day1-4 完成各项交付。

---

## 2. 关键交付物

| 编号 | 交付物 | 内容描述 | 归档位置 | 验收方 |
|------|--------|----------|----------|--------|
| D1 | `schema/current_schema.sql` | 通过 `pg_dump --schema-only` 获取的实时 Schema 快照 | `database/schema/current_schema.sql` | DBA |
| D2 | 声明式 Schema | 统一的 `database/schema.sql`（或 `schema.hcl` 附件） | `database/schema.sql` | 架构组 |
| D3 | 基线迁移文件 | `database/migrations/20251105000000_base_schema.sql`（含 Up/Down） | migrations 目录 | 架构组 + DBA |
| D4 | 审阅记录 | 审核意见与签字纪要 | `docs/archive/development-plans/214-signoff-*.md` | 架构组 + DBA |
| D5 | 执行日志 | 抽取、对比、生成脚本执行记录 | `logs/214-phase1-baseline/` | PM/QA |

---

## 3. 工作分解与时间线

| 日期 | 任务 | 负责人 | 输出 |
|------|------|--------|------|
| Day1（2025-11-03 上午） | 冻结/确认当前数据库状态（依赖 210 Phase0 已完成） | 基础设施组 | 状态记录 |
| Day1（2025-11-03 下午） | 生成 `current_schema.sql`、产出 `schema-inspect.hcl`、记录 diff | DBA | D1 |
| Day2（2025-11-03 下午） | 整理 `database/schema.sql`，补齐 Goose 元数据；准备迁移 | 架构组 | D2 |
| Day3（2025-11-03 傍晚） | 执行 Goose up/down/up + `go test ./...` | DBA + QA | 验证日志 |
| Day4（2025-11-03 晚间） | 归档日志、生成签字纪要、更新索引文档 | PM + Codex | D3 最终版 + D4, D5 |

> 所有命令需在 Docker 容器内执行，禁止使用宿主机数据库服务。

---

## 4. 执行步骤与命令指引

### 4.1 Schema 快照
```bash
export PG_BASELINE_DSN="postgres://user:password@postgres:5432/cubecastle?sslmode=disable"
docker compose exec -T postgres \
  pg_dump --schema-only --no-owner --no-privileges "$PG_BASELINE_DSN" \
  > database/schema/current_schema.sql
```
- 补充执行 `atlas schema inspect --url "$PG_BASELINE_DSN"`，生成 `database/schema/schema-inspect.hcl` 供 diff 使用。
- 结果存放于 `database/schema/`，比对文件由 DBA 记录至 `logs/214-phase1-baseline/schema-diff.txt`。

### 4.2 声明式 Schema 整理
- 以 `current_schema.sql` 为基础，合并业务命名规范、注释、索引定义。
- 确保与 203/205 计划中的数据模型保持一致（命名、外键、索引）。
- 若需要 `schema.hcl` 以便 Atlas diff，放置于同目录并在文档中引用。

### 4.3 基线迁移生成
- 首选 Atlas:
  ```bash
  bin/atlas migrate diff \
    --dir "file://database/migrations" \
    --dev "$PG_BASELINE_DSN" \
    --to "file://database/schema.sql" \
    --format goose
  ```
- 若 Atlas 不适用，则手工拆分 Schema，逐对象编写 Up/Down。
- 迁移文件命名：`database/migrations/20251106000000_base_schema.sql`，需包含：
  ```sql
  -- +goose Up
  ...
  -- +goose Down
  ...
  ```

### 4.4 本地验证
```bash
make docker-up
make db-migrate-all      # goose up
make db-rollback-last    # goose down
go test ./... -count=1   # 确认新基线不破坏现有测试
```
- 将执行日志保存到 `logs/214-phase1-baseline/`。

### 4.5 审阅与签字
- 通过会议/文档收集意见，按照“问题-处理-确认人”记录。
- 生成 `docs/archive/development-plans/214-signoff-20251103.md`，包含最终版本摘要、链接、签字人。

---

## 5. 依赖与前置条件

| 依赖项 | 状态 | 说明 |
|--------|------|------|
| Plan 212 Day6-7 架构审查 | ✅ 完成 | 目录与共享复用决议明确，可按统一结构整理 Schema |
| Plan 213 Go 工具链基线 | ✅ 完成 | Go 1.24.9 工具链已确认，可使用最新 `goose`/`atlas` 版本 |
| Plan 210 Phase0 冻结与备份 | ✅ 完成 | 备份与保护分支已归档，可安全进行萃取工作 |
| Docker Compose 环境可用 | ✅ | 需使用容器内的 PostgreSQL 实例 |

---

## 6. 风险与对策

| 风险 | 等级 | 说明 | 对策 |
|------|------|------|------|
| Atlas 生成迁移不完整 | 中 | 自定义函数/触发器可能未完全导出 | 预备手工方案；审阅环节逐项对照 Schema |
| Down 脚本遗漏对象 | 高 | 影响回滚能力 | 手工审核 + `goose down` 实测验证 |
| Schema 与业务命名不一致 | 中 | 影响后续模块开发 | 参考 203/205 文档逐字段核对 |
| 执行日志缺失 | 低 | 影响审计与复盘 | 所有命令保留脚本与输出；提交 PR 时附链接 |

---

## 7. 验收标准

- `database/schema/current_schema.sql`、`database/schema.sql`、基线迁移文件均入库，且通过代码审查。✅
- `make db-migrate-all`、`make db-rollback-last`、`go test ./...` 均在 Go 1.24 环境下通过。✅（日志参见 `logs/214-phase1-baseline/`）
- `docs/archive/development-plans/214-signoff-20251103.md` 保存联合签字纪要。✅
- `reports/PHASE-TRANSITION-REPORT-20251104.md`、`docs/development-plans/210-database-baseline-reset-plan.md` 等跨计划文档已同步完成状态。✅

---

## 8. 完成情况概览

- 实际执行时间：2025-11-03（较原计划提早 3 天），全量交付物当日完成。
- Atlas CLI 已在 `bin/atlas` 落地，可离线复用；`database/schema/schema-inspect.hcl` 与 SQL 快照保持一致。
- Goose 基线迁移回放验证成功，日志位于 `logs/214-phase1-baseline/day3-*`。
- 后续工作：转入 Plan 203 Phase 2（预计 2025-11-13 启动），沿用本次萃取成果。

---

**版本历史**
- v1.2 (2025-11-03)：标记计划完成，补充执行结果与验收状态。  
- v1.1 (2025-11-04)：更新状态为待执行，明确 Week1 日程与责任。  
- v1.0 (2025-11-04)：初稿，等待评审。
