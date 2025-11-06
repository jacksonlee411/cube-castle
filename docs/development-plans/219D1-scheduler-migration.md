# Plan 219D1 – Scheduler / Temporal 代码迁移

**文档编号**: 219D1  
**关联路线图**: Plan 219 → 219D  
**依赖子计划**: 219A 目录、219B 查询、219C 审计/验证  
**目标周期**: Week 4 Day 21（219D 第一阶段）  
**负责人**: 后端团队（Scheduler 代码 Owner）

---

## 1. 目标

1. 盘点并迁移 `organization_temporal_service.go`、`operational_scheduler.go` 及相关 workflow/activity/cron 逻辑到 `internal/organization/scheduler/`。
2. 建立统一的 Scheduler Facade/Service，在 `cmd/hrms-server/command/main.go` 等入口中完成依赖注入，确保构建通过。
3. 保留回退路径：记录原目录结构与入口，必要时可一键回滚。

---

## 2. 范围

| 模块 | 内容 |
|------|------|
| 代码迁移 | 识别调度/Temporal 相关文件，搬迁至新目录并更新包名 |
| 依赖注入 | main/di/初始化脚本引用更新，确保 REST 与后台任务均指向新的 Facade |
| 行为校验 | 冒烟运行 position version 激活、timeline 修复等关键 workflow，确认日志/队列不变 |

不包含：配置集中化（由 219D2）、监控指标（219D3）、深度测试与文档（219D4/219D5）。

---

## 3. 详细任务

1. **现状盘点**
   - 使用 `rg "Temporal"`、`rg "scheduler"` 于 `cmd/hrms-server/command/internal/` 定位全部相关文件及依赖。
   - 列出 workflow 名称/队列/调用方，形成迁移清单并附路径。

2. **目录迁移与重构**
   - 将 workflow/activity/cron/job struct 移入 `internal/organization/scheduler/`，必要时拆分 `workflow/activities/cron` 子目录。
   - 调整包命名、可见性（私有/导出）以符合 internal 约束，新增 Facade（如 `scheduler.Service`）。

3. **依赖注入更新**
   - 在 `cmd/hrms-server/command/main.go`、`internal/app/bootstrap.go` 等初始化流程中注入新的 Service。
   - 替换旧引用，确保命令处理器、后台任务均使用统一入口。

4. **冒烟验证与回退策略**
   - 本地 `make build`、`make run-dev`，确认工作流注册及队列监听成功。
   - 记录回退步骤：保留 git tag/branch，说明如何恢复旧目录。

---

## 4. 验收标准

- [ ] 所有 Scheduler/Temporal 文件已移动到 `internal/organization/scheduler/`，旧目录无残留。
- [ ] 构建、启动脚本通过，关键 workflow 能够注册与触发（冒烟结果附日志片段）。
- [ ] 回退说明文档化（可附在子计划附录或提交说明）。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 隐藏依赖遗漏 | 中 | 迁移前通过 `rg` 与 `go list` 列表核对；迁移后运行 `go test ./...` 捕获编译失败 |
| 包路径调整导致循环依赖 | 中 | 先引入 Facade 接口，再逐步替换；必要时引入适配器过渡 |
| 冒烟验证不足 | 中 | 与 219D4 协调测试用例，至少验证两个关键 workflow |

---

## 6. 交付物

- 更新后的 `internal/organization/scheduler/` 目录及 Facade。
- 迁移清单与回退说明（附在 PR 描述或 `docs/development-plans/219D1-scheduler-migration.md` 附录）。
- 构建/冒烟验证记录。
