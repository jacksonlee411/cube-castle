# 402D · 切换与回收

**关联计划**：Plan 400、Plan 401、Plan 402、Plan 222/255、Plan 300  
**状态**：待启动（依赖 402A~402C 完成）  
**范围**：一次性迁移、Feature Flag 切换、前端/查询适配、门禁与回滚演练  
**日志要求**：`logs/plan402/migration/*.log`、`logs/plan402/validator/*.json`、`logs/plan402/ui/*.log`、`logs/plan402/verification/*.log`、`logs/plan402/rollback/*.log`、`logs/plan402/capability/*.log`

> 目标：将所有读写切换至 Standard Object 三表，停用旧 `organization_units` 写入，完成前端/查询适配与门禁验收，并验证回滚路径。

---

## 1. 目标

1. 执行最终迁移与校验，关闭旧仓储写路径，确保只读或归档。
2. 更新前端与查询层，使其完全依赖 SOM，并完成 Playwright/OBS 证据。
3. 跑通所有门禁（Go/前端/质量脚本），并完成 Goose Down + 数据回滚演练。
4. 输出切换 Runbook、执行报告与能力契约核对结果。

---

## 2. 工作项

### D1 · 切换执行
- 运行 `cmd/tools/standardobject-migrator` 对生产级数据进行最终导入，日志写入 `logs/plan402/migration/*.log`。
- 执行 `standardobject-validator` 确认 0 差异，附 `logs/plan402/validator/*.json`。
- 关闭旧仓储写路径（Feature Flag 默认开启 SOM），旧表仅保留只读视图或直接冻结。
- 输出切换 Runbook（步骤、负责人、回滚条件）与 DEC/OCL 体检报告，确保 `schema-registry.json` 与能力契约无缺口。

### D2 · 前端与查询适配
- 前端（组织/职位页面）完全接入 `standardObjectAdapter`；清理组织特有冗余字段。
- 重新运行 `npm run quality:preflight`、`npm run test`、`npm run test:e2e`，并把日志写入 `logs/plan402/ui/*.log`。
- GraphQL/REST 查询切换至 SOM 数据源，更新缓存/selector/OBS 事件；Manifest/Slot 记录 DEC 列表与视点。

### D3 · 门禁与回滚
- 跑通 `make test`、`make test-db`、`scripts/quality/*`（含 capabilityContracts 规则），输出 `logs/plan402/verification/*.log`。
- 编写 Goose Down + 数据回滚脚本，并在 staging 演练；日志存入 `logs/plan402/rollback/*.log`。
- 在 `scripts/quality/architecture-validator.js` 中运行 capability contract 完整性检查，确认 4.3 表格条目覆盖全部 Federate。

---

## 3. 交付物

- 切换 Runbook 与执行报告、迁移/校验日志、DEC/OCL 体检报告。
- 前端/查询适配 diff、Manifest/Slot 更新、Playwright/OBS 证据 (`logs/plan402/ui/*.log`)。
- 门禁运行日志 (`logs/plan402/verification/*.log`)、回滚脚本与演练记录 (`logs/plan402/rollback/*.log`)。
- 能力契约核对结果 (`logs/plan402/capability/*.log`)。

---

## 4. 验收标准

1. 切换后所有写操作仅落在 `standard_objects*`，旧表只读或归档，监控显示 0 双写差异。
2. 前端 UI 与 GraphQL/REST 功能在 SOM 模式下通过全量测试（含 Playwright 核心场景），日志齐全。
3. DEC/OCL 体检报告 0 漏项、0 违规，能力契约矩阵覆盖所有视点；相关证据已存档。
4. 回滚演练可在 30 分钟内恢复旧实现并保证数据一致。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 切换期间出现异常 | 数据不一致或服务中断 | 严格按 Runbook 执行，设置回滚点并实时监控 |
| 前端回归遗漏 | UI/交互异常 | 强制执行 Playwright/OBS 证据，Plan 222 负责人复核 |
| 监控缺失 | 无法及时发现问题 | 切换前配置 SOM 专用指标与告警，纳入 Plan 272 |

---

402D 完成后方可启动 402E；若出现重大问题，必须通过回滚脚本恢复旧数据路径并记录 `logs/plan402/rollback/*.log`。
