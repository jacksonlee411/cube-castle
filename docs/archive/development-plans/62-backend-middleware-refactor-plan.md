# 62号文档：后端观测与运维巩固计划（Phase 2 精简版）

**版本**: v2.1
**创建日期**: 2025-10-10
**最后更新**: 2025-10-10
**维护团队**: 全栈工程师（单人执行）
**状态**: 第一批交付完成 —— Prometheus 指标与监控文档已完成；运维开关配置待后续补充
**遵循原则**: CLAUDE.md 资源唯一性与跨层一致性原则
**关联计划**: 60号总体计划、61号执行计划第一阶段验收、06号推进记录

---

## 1. 背景与目标

### 1.1 背景
- Phase 1 已完成契约与类型统一，后端服务当前运行稳定。
- 评审确认响应封装、事务/审计逻辑、中间件均已具备成熟实现（详见 `internal/utils/response.go`、`internal/audit/logger.go` 等）。
- 现阶段真正缺口集中在 **观测指标补充**、**运维开关梳理** 与 **操作手册更新**。需以最小成本巩固后端服务的可观测性与日常运维体验。

### 1.2 目标
1. **指标完善**：对命令服务补充与 Phase 1 契约成果相关的 Prometheus 指标，并验证 Query/Command 监控一致性。
2. **策略文档更新**：沉淀运维操作、监控告警及回滚步骤到手册，形成长期可依赖的运维指引。

---

## 2. 范围

### 2.1 涉及模块
- `cmd/organization-command-service/internal/middleware/*.go`
- `cmd/organization-command-service/internal/handlers/devtools.go`
- `cmd/organization-command-service/internal/handlers/operational.go`
- `docs/reference/` 中与监控/运维相关的章节（若缺少则新增小节）

### 2.2 明确不在范围
- 事务/审计封装、统一响应结构、双写机制等已存在能力，此阶段不再重复实现。
- 数据库结构调整与大规模后端重构。
- 前端或 Query 服务的大型改造（另有阶段覆盖）。

---

## 3. 时间线（预估 1 周）

| 周次 | 里程碑 | 交付物 |
|------|--------|--------|
| Week 3 | 指标补充与运维开关梳理 | Prometheus 指标、Dev/Operational 开关、运维手册更新 |

---

## 4. 详细任务清单

### 4.1 指标补充
- [x] 盘点命令服务现有指标（`/metrics`）与 Query 服务差异。✅ 2025-10-10
- [x] 增补与契约执行相关的指标，实现以下 Prometheus Counter：✅ 2025-10-10
  - `temporal_operations_total{operation, status}` —— 时态操作计数（CreateVersion/UpdateVersionEffectiveDate/DeleteVersion/SuspendOrganization/ActivateOrganization）
  - `audit_writes_total{status}` —— 审计日志写入计数（internal/audit/logger.go 和 repository.AuditWriter）
  - `http_requests_total{method, route, status}` —— HTTP 请求计数（由性能中间件统一记录）
  - 实现位置：`cmd/organization-command-service/internal/utils/metrics.go`
  - 暴露端点：`cmd/organization-command-service/main.go` line 202-207
- [x] 创建 `scripts/quality/validate-metrics.sh` 自动化验证脚本。✅ 2025-10-10
  - 支持关键指标和业务触发指标分类验证
  - 返回适当退出码供 CI 集成
  - 提供详细使用说明和示例操作

### 4.2 文档与验收
- [ ] 更新 60-execution-tracker 第二阶段进展。
- [x] 更新或新增 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 中监控/运维章节。✅ 2025-10-10
  - 新增"📊 运行监控（Prometheus）"完整章节
  - 记录三类指标的名称、说明、标签、触发条件
  - 提供自动化验证脚本使用方法
  - 提供手动验证步骤与命令示例
  - 说明 Prometheus Counter 行为与代码位置
- [x] 编写 Phase 2 验收报告草稿（62-phase-2-acceptance-draft.md），记录指标验证与运行时测试结果。✅ 2025-10-10
  - 已完成 v0.2 版本，包含完整验收清单与验证结果
  - 记录了指标实现位置与运行时验证步骤
  - 识别并说明了业务触发指标的可见性行为
  - 提出后续改进建议（自动化测试、完整业务流程验证）

---

## 5. 验收标准与测试

1. **指标验证**：
   - Prometheus 中能看到新增指标，执行 `curl /metrics` 出现对应字段。
   - 指标说明文档中记录含义与参考阈值。
2. **开关验证**：
   - 通过环境变量/配置关闭 Dev/Operational Handler 后，接口返回 403 或 404。
   - 白名单生效示例（白名单用户可访问，其他用户被拒绝）。
3. **文档更新**：
   - 60-execution-tracker 第二阶段进度勾选。
   - 63号验收报告草稿完成。
   - 若对外行为变更，更新 `docs/reference` 中相关章节。

---

## 6. 风险与缓解

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| 指标过多影响性能 | 新增指标可能带来额外开销 | 控制指标粒度，仅保留关键指标，并在文档注明采样策略 |
| 运维开关默认异常 | 默认配置可能阻止合法访问 | 提供默认配置模板，保留安全兜底（例如仅对生产环境启用严格限制） |
| 文档滞后 | 新能力未同步至手册 | 在任务中明确文档更新步骤，并纳入验收标准 |

---

## 7. 交付物
- [x] 更新后的 Prometheus 指标及收集脚本说明 ✅ 2025-10-10
  - Prometheus 指标实现：`cmd/organization-command-service/internal/utils/metrics.go`
  - `/metrics` 端点暴露：`cmd/organization-command-service/main.go:202-207`
  - 自动化验证脚本：`scripts/quality/validate-metrics.sh`
- [ ] Dev/Operational Handler 开关与白名单示例配置（已有实现，待文档化）
- [x] 运维/监控手册更新 ✅ 2025-10-10
  - `docs/reference/03-API-AND-TOOLS-GUIDE.md` 新增"📊 运行监控（Prometheus）"章节
- [x] Phase 2 验收报告草稿 ✅ 2025-10-10
  - `docs/development-plans/62-phase-2-acceptance-draft.md` v0.2

---

## 8. 追踪与后续
- 计划文档：62号本文件（v2.0）。
- 执行跟踪：`docs/development-plans/60-execution-tracker.md`（第二阶段条目）。
- 验收报告：63号文档（待创建）。
- 若后续发现新的后端缺口，再另立专题计划，避免与现有能力重复。

