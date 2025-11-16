# Plan 257 - 前端领域 API Facade 采纳

文档编号: 257  
标题: 前端领域 API Facade 采纳（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.0  
关联计划: 202、241（前端合流）、242/244/245（命名与类型统一）

---

## 1. 目标
- 在前端层收敛领域 API Facade，隔离契约调整、提升可测性与可替换性；
- 与 242/244/245 的命名与类型统一协同，避免双事实来源。

## 2. 交付物
- Facade 最小模式与接入示例（只引用前端 shared/api/）；
- 统一错误与重试策略（索引约定）；
- 证据：logs/plan257/*（采纳前/后对比）。

## 3. 验收标准
- 关键页面迁移至 Facade 调用；
- ESLint/AST 规则启用并在 CI 中统计“Facade 覆盖率”；
- E2E 对关键路径稳定（不退化）。

---

维护者: 前端（与契约治理协作）

---

## 4. 覆盖率度量与门禁（CI）
- 覆盖率口径：
  - 分子：通过领域 Facade 的业务调用点数量（按 import + 方法签名识别）
  - 分母：所有对统一客户端（unifiedGraphQLClient/unifiedRESTClient）或 fetch/axios 的直接业务调用点数量
  - 覆盖率=分子/分母（目标门槛：组织/职位模块 ≥80%）
- 实现方式：
  - ESLint 规则禁止直接 fetch/axios；对 unified 客户端直连仅告警但计数
  - AST 扫描脚本输出覆盖率报表（JSON），作为 CI 工件上传；支持识别从 `@/shared/api/facade/*` 与从 `@/shared/api` 重新导出的 Facade 函数（如 `getOrganizationByCode`）两种导入方式
  - 低于门槛阻断合并；临时豁免必须以 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注并在 215/06 登记

---

## 5. 范围与不做的事（Scope / Non-Goals）
- 范围（Phase 1 → Phase 2）：
  - Phase 1：组织、职位模块中的业务页面与共享 hooks 调用统一收敛到 Facade（最小可用）
  - Phase 2：其余模块逐步接入，结合 256/258 契约治理确保一致性
- Non-Goals：
  - 不在本计划内修改后端契约；契约演进由 256/258 驱动，Facade 仅做兼容适配

## 6. 目录与约定（SSoT 对齐）
- Facade 入口：`frontend/src/shared/api/facade/*`（唯一事实来源）
- 统一客户端：`frontend/src/shared/api/unified-client.ts`（仅在 Facade/适配层使用）
- 业务层使用方式：`import { getOrganizationByCode } from '@/shared/api/facade/organization'`
- 类型来源：`docs/api/*` + Plan 256 生成物（禁止手写重复类型）

## 7. 门禁与报告
- 本地：
  - 运行 `make guard-plan257` 生成 `reports/facade/coverage.json`，同时在 `logs/plan257/` 产出证据
  - 环境变量 `THRESHOLD` 控制阈值（默认 0.0 报告模式；建议迁移达标后提升至 0.8）：`THRESHOLD=0.8 make guard-plan257`
- CI：
  - 工作流：`.github/workflows/plan-257-gates.yml`（artifact: `plan257-facade-coverage`）
  - Required checks：已切换为“阻断（阈值 0.8）”，请在受保护分支勾选 “Plan 257 - Facade Coverage Gate” 为必需检查

## 8. 阈值与例外治理
- 阈值：组织、职位模块覆盖率 ≥0.80（其余模块按 Phase 2 推进时设定）
- 临时例外：必须以 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注原因与回收日期（≤1 迭代），在 215/06 登记

## 9. 回滚与处置
- 发现误报：修正 `scripts/facade/coverage-scan.js` 口径或补充 Facade，并复跑
- 必须放行：按“临时例外”规范登记，设置回收期；逾期 CI 阻断

---

## 10. 实施现状
- 目录与示例已创建：`frontend/src/shared/api/facade/organization.ts`
- ESLint 规则：禁止 features 层直连 `shared/api/unified-client`（提示改用 Facade）
- 覆盖率门禁：`scripts/facade/coverage-scan.js` + `make guard-plan257` + CI 工作流
- 状态：进行中（阻断门禁已启用；阈值 0.8）

变更记录
- v1.2（2025-11-16）
  - 切换 CI 门禁为阻断，阈值 0.8；建议将 “Plan 257 - Facade Coverage Gate” 设为 Required check
- v1.1（2025-11-16）
  - 明确 Scope/Non-Goals、目录与 SSoT 绑定、门禁与阈值、回滚与实施现状
  - 修正 Facade 命令端点与 OpenAPI 一致（/api/v1/organization-units/*）；更新支持 If-Match/ETag
  - 覆盖率扫描增强：识别从 `@/shared/api` 入口导入的 Facade 函数；输出 `offenders` 清单辅助迁移
