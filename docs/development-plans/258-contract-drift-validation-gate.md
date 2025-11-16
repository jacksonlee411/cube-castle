# Plan 258 - 契约漂移校验门禁

文档编号: 258  
标题: 契约漂移校验门禁（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.1  
关联计划: 202、256（SSOT 生成）、CI 门禁

---

## 1. 目标
- 在 CI 设置“契约漂移”门禁：实现/文档/脚本任意一端变更需保持一致，发现漂移即阻断；
- 输出可读报告，指向单一事实来源文件位置。

## 2. 范围与交付物
- 范围：对比 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 转 JSON 后的字段矩阵（类型/字段/可空/描述）
- 门禁工作流与脚本（仅索引 .github/workflows/* 与 scripts/quality/*）；
- 漂移报告样例与修复建议（登记，不复制正文）；
- 证据：logs/plan258/*。

## 3. 前置依赖
- 256（SSoT 生成链路）可运行；契约作为唯一真源（不可从实现反向生成）

## 4. 验收标准
- 漂移检测脚本可稳定发现主要不一致；
- 新增/修改契约均可在 MR 中一体化审查；
- 白名单与豁免机制符合 AGENTS 临时方案管控（见下）。

## 5. 步骤
1) 差异比对脚本完善与集成（OpenAPI↔GraphQL）
2) CI 工作流接线并上传差异报告产物
3) 文档与 README/指南更新引用

## 6. 白名单与回滚（AGENTS 对齐）
- 临时白名单项必须使用 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注，最长期限一个迭代
- 在 215 与 `scripts/check-temporary-tags.sh` 清单登记；超期自动阻断
- 发现误报 → 快速修订与复跑；必要时回滚 MR

---

维护者: 合同治理小组（与平台/QA 协作）

---

## 7. 实施（v1）
- 门禁脚本：`scripts/contract/drift-matrix-gate.js`
  - 输入：`docs/api/openapi.yaml`、`docs/api/schema.graphql`
  - 对比维度：字段存在性、类型（基础/枚举/引用名）、列表(List) 语义、非空（required/!）、描述（信息级）
  - 允许列表：`scripts/contract/drift-allowlist.json`（必须包含 `reason` 与未来 `expires`）
  - 产物：`reports/contracts/drift-matrix-report.json`
  - 行为：
    - 未允许的“缺字段 / 类型不匹配 / 列表语义不匹配 / 非空不匹配” → 阻断（退出码 5）
    - 描述差异（description-mismatch）→ 信息级（不阻断，用于审阅）
- CI 工作流：`.github/workflows/plan-258-gates.yml`（阻断）
  - 步骤：安装依赖 → 运行 gate → 上传报告
  - 保护分支建议：加入 Required checks
- 与 256 协同：
  - 256 保留“枚举差异报告（非阻断）”；258 负责字段矩阵阻断与白名单收敛

---

## 8. 已知临时差异（登记，需在到期前收敛）
// TODO-TEMPORARY(2025-11-23): 统一 REST/GraphQL 关于 profile/派生审计字段的表达与兼容策略；当前 allowlist 仅用于过渡。
- OrganizationUnit : Organization
  - profile（REST=object, GraphQL=String）
  - path（GraphQL 遗留字段，已 @deprecated）
  - childrenCount/isTemporal（GraphQL 派生字段）
  - deletedBy/deletionReason/suspendedAt/suspendedBy/suspensionReason（GraphQL 扩展审计字段；REST 以 operatedBy/operationReason 表达）

---

## 9. 本地与 CI 命令（规范索引）
- 本地阻断门禁：`make guard-plan258`（日志与报告会落盘 `logs/plan258/`）
- CI 工作流：`.github/workflows/plan-258-gates.yml`（阻断，产出 `plan258-drift-matrix-report` 工件）
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
