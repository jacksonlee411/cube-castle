# Phase1 回归验证记录

**计划编号**：Plan 211 Phase1
**执行负责人**：Codex（全栈）
**完整规范**：`reports/DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md`
**相关计划**：`docs/development-plans/06-integrated-teams-progress-log.md`

---

## Day8 数据一致性巡检

- **脚本**：`scripts/tests/test-data-consistency.sh`
- **数据真源**：`scripts/data-consistency-check.sql`
- **官方规范**：`reports/DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md`（包含完整的环境准备、命令、产出、判定、异常处理）
- **运行方式**（快速参考）：
  1. 确保 Docker PostgreSQL 已通过 `make docker-up` 启动，并导入最新迁移。
  2. 先验证脚本可用：`scripts/tests/test-data-consistency.sh --dry-run`
  3. 正式执行脚本：`scripts/tests/test-data-consistency.sh`
  4. 预期产出：
     - 原始结果：`reports/consistency/data-consistency-<timestamp>.csv`
     - 概要报告：`reports/consistency/data-consistency-summary-<timestamp>.md`
  5. 在下方表格登记结果

- **当前状态**：✅ 已执行（2025-11-03T02:02:29Z UTC；详见运行记录）
- **扩展检查**：`scripts/phase1-acceptance-check.sh` 已于 2025-11-03T03:39:38Z UTC 运行，产出 `reports/acceptance/phase1-acceptance-summary-20251103T033918Z.md`。
- **异常处理**：
  - 若判定为 `❌ FAIL`，详见规范五（异常处理流程）
  - 常见问题与排查方法见规范五、七
  - 修复后需重新执行并更新表格记录

> **重要**：本文件与官方规范保持一致。执行时需严格按规范进行，确保事实唯一来源。任何疑问请参考 `reports/DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md`。

---

## 运行记录

| 执行时间 (UTC) | 环境 | 脚本版本 | 判定 | 异常数 | 审计日志(7d) | 关键结论 | 附件 |
|----------------|------|---------|------|--------|------------|----------|------|
| 2025-11-03T02:02:29Z | Dev | v1.0 | ✅ PASS（审计豁免） | 0/0/0/0 | 0 | 无异常；近7日无业务操作，QA+架构确认审计豁免 | consistency/data-consistency-summary-20251103T020229Z.md |

**表格字段说明**：
- **执行时间**：脚本输出的 UTC 时间戳（格式 `YYYY-MM-DDTHH:MM:SSZ`）
- **环境**：Dev/Test/Staging/Prod
- **脚本版本**：脚本的当前版本（参见脚本顶部注释）
- **判定**：✅ PASS 或 ❌ FAIL
- **异常数**：格式 `M/C/IP/DC`，分别表示 MULTIPLE_CURRENT / TEMPORAL_OVERLAP / INVALID_PARENT / DELETED_BUT_CURRENT 的计数
- **审计日志(7d)**：近 7 天审计日志记录数
- **关键结论**：问题摘要或"无异常"
- **附件**：指向摘要报告相对路径（示例：`consistency/data-consistency-summary-20251103T143022Z.md`）

---

## 待办清单

### Day8 执行
- [x] 确认前置环境就绪（Docker、迁移、服务健康）
- [x] 执行干运行：`scripts/tests/test-data-consistency.sh --dry-run`
- [x] 执行正式验证：`scripts/tests/test-data-consistency.sh`
- [x] 检查摘要报告判定（✅ PASS 或 ❌ FAIL）
- [x] 在上表新增运行记录行
- [ ] 若 FAIL，按规范五异常处理流程推进
- [x] 审计日志计数为 0，经 QA+架构确认属于近7日无操作场景，按规范 4.1 豁免处理并记录

### Day9-10 延伸测试
- [ ] REST/GraphQL 接口对照测试（与 Day6 基线对比）
- [ ] E2E 核心流程验证
- [ ] 补充延伸测试结果至运行记录
- [ ] 准备 Day10 复盘与最终交付

---

## Go 1.24 基线评审（Plan 213）

### 评估概述
- 执行 `go env GOVERSION` → `go1.24.9`，与仓库 `go.mod` 中 `toolchain go1.24.9` 一致。
- 使用 `go list -m -json all` 盘点依赖 `GoVersion` 字段，仅根模块 `cube-castle` 标记为 `1.24.0`；第三方依赖均未要求高于 1.22 的版本。
- 未发现需要额外补丁的第三方库，CI/Docker 构建脚本在 1.24 下保持原状。

### 验证记录
- `go test ./... -count=1` ✅（2025-11-04 16:20 CST 执行，覆盖所有命令/查询服务、共享库、测试套件）。
- 现有 `scripts/phase1-acceptance-check.sh`（2025-11-03）已在 Go 1.24 基线上通过，作为回归基线继续有效。

### 风险复核
- **团队环境**：需通知开发者将本地 Go 版本升级至 ≥1.24（记录于 06 号进展文档）。
- **回退预案**：当前无触发条件；若未来依赖要求回落至 1.22，将按 PLAN200 流程启动专门变更。
- **结论**：Steering 同意维持 Go 1.24 基线，无需进一步回退操作。

---

## 历史执行日志（如有）

> 此区域用于保存历次执行的详细信息。若脚本执行出现问题或需要追溯历史，请参考本区域或检查 `reports/consistency/` 目录中的完整日志文件。

## Phase1 验收脚本运行记录

| 执行时间 (UTC) | 环境 | 脚本 | 判定 | 备注 | 附件 |
|----------------|------|------|------|------|------|
| 2025-11-03T03:39:38Z | Dev | phase1-acceptance-check v1.0 | ✅ PASS | Go build/test + npm lint/test + data check 全绿 | acceptance/phase1-acceptance-summary-20251103T033918Z.md |
