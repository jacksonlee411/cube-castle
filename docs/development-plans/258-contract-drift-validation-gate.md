# Plan 258 - 契约漂移校验门禁

文档编号: 258  
标题: 契约漂移校验门禁（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.2  
关联计划: 202、256（SSOT 生成）、CI 门禁

---

## 1. 目标
- 在 CI 设置“契约漂移”门禁：实现/文档/脚本任意一端变更需保持一致，发现漂移即阻断；
- 输出可读报告，指向单一事实来源文件位置。

## 2. 范围与交付物
- 范围（分阶段实施）：
  - Phase A（已启用，阻断）：OpenAPI ↔ GraphQL 枚举差异（UnitType/Status/OperationType）
  - Phase B（计划中，先报告后阻断）：主实体字段矩阵（字段/类型/可空/描述）
- 门禁工作流与脚本（仅索引，不复制实现）：
  - 阶段 A：`.github/workflows/plan-258-gates.yml`、`scripts/contract/drift-check.js`、`scripts/contract/drift-allowlist.json`
  - 阶段 B：扩展 `scripts/contract/openapi-to-json.js` 与 `scripts/contract/graphql-to-json.js` 输出字段矩阵；`drift-check.js` 增强矩阵比对（按阶段推进）
- 漂移报告工件：`reports/contracts/drift-report.json`（Actions Artifact: plan258-drift-report）
- 证据索引：`logs/plan258/*`（由工作流/脚本落盘）

## 3. 前置依赖
- 256（SSoT 生成链路）可运行；契约作为唯一真源（不可从实现反向生成）

## 4. 验收标准
- 阶段 A：枚举差异检测稳定、门禁阻断生效、白名单受控并有回收期；
- 阶段 B：字段矩阵比对先报告验证 1-2 个工作日后切换阻断；
- 新增/修改契约均可在 PR 中一体化审查（报告可下载），未登记差异不得放行。

## 5. 步骤
1) 运行生成链路（Plan 256）：`scripts/contract/sync.sh`
2) 阶段 A 接入（阻断）：
   - 工作流：`plan-258-gates.yml` → `node scripts/contract/drift-check.js --fail-on-diff`
   - 报告：`reports/contracts/drift-report.json`（artifact: plan258-drift-report）
   - 受保护分支：将 `plan-258-gates` 设为 Required check
3) 阶段 B 推进（先报告后阻断）：
   - 扩展生成脚本输出字段矩阵 → `drift-check.js` 增强矩阵比对（初期仅报告）
   - 收敛误报后改为 `--fail-on-diff`
4) 在 215 登记证据与 Required checks 变更，保持单一事实来源。

## 6. 白名单与回滚（AGENTS 对齐）
- 临时白名单（仅少量、短期差异）：
  - 文件：`scripts/contract/drift-allowlist.json`
  - 代码处必须添加 `// TODO-TEMPORARY(YYYY-MM-DD):` 注释并在 215 登记回收期（≤1 迭代）
  - CI 同时运行 `scripts/check-temporary-tags.sh` 校验 TODO 标注规范
- 回滚与处置：
  - 发现误报 → 修复脚本或白名单条目并复跑
  - 必须放行 → TODO 标注 + allowlist 登记 + 回收日期，逾期 CI 阻断

---

维护者: 合同治理小组（与平台/QA 协作）

---

## 7. 实施现状
- 枚举门禁（Phase A）：已接入 `plan-258-gates.yml`，通过 `drift-check.js --fail-on-diff` 阻断
- 字段矩阵（Phase B）：按“先报告后阻断”推进，待生成脚本输出矩阵后在 `drift-check.js` 增强
- 与 256 协同：256 负责“契约→生成→快照一致”，258 负责“差异检测与门禁”

---

## 8. 已知临时差异（登记）
- 占位：按 PR 引入差异时在此登记，并在 `scripts/contract/drift-allowlist.json` 与 215 同步索引

---

## 9. 本地与 CI 命令（规范索引）
- 生成：`make generate-contracts`（调用 `scripts/contract/sync.sh`）
- 漂移检测：`node scripts/contract/drift-check.js [--fail-on-diff]`
- CI 工作流：`.github/workflows/plan-258-gates.yml`（阻断，产出 `plan258-drift-report` 工件）
- 建议 Required checks：`api-compliance`（Spectral）+ `contract-testing`（256）+ `plan-258-gates`（本计划）

---

## 10. 非空/可空语义裁决（规则表）
- 术语：
  - REST 非空：OpenAPI `required=true` 且 `nullable=false`
  - REST 可空：OpenAPI `nullable=true`（无论 required 是否为 true）
  - GraphQL 非空：`Type!`（顶层 NonNull）
- 裁决：
  - REST 非空 ↔ GraphQL 非空：一致（通过）
  - REST 非空 ↔ GraphQL 可空：不一致（阻断）
  - REST 可空 ↔ GraphQL 非空：不一致（阻断）
  - REST 可空 ↔ GraphQL 可空：一致（通过）
  - 备注：
    - OpenAPI 出现 `required=true && nullable=true` 为边缘用法；按“可空”处理（允许为空，非空不匹配）
    - 列表类型：仅比较“是否列表”与元素基础类型；若 OpenAPI 未提供 items 基础类型，则只判定“是否列表”

---

变更记录
- v1.2（2025-11-16）
  - 明确分阶段范围与现状；对齐已实现的 `drift-check.js` 与 `plan-258-gates.yml`
  - 引入 `scripts/contract/drift-allowlist.json` 白名单与 TODO 联动规范
- v1.1（2025-11-16）
  - 初稿扩展，补充矩阵比对裁决规则与实施清单
