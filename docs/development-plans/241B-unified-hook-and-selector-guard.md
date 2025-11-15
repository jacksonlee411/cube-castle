# 241B – 统一 Hook 薄封装与选择器门禁

编号: 241B  
标题: `useOrganizationDetail` 薄封装 + `useTemporalEntityDetail` 单测 + 选择器门禁（ESLint + Guard）  
创建日期: 2025-11-15  
状态: 待实施  
上游关联: 241（框架重构 · 收尾）、245/245A（统一类型与 Hook · 已完成）、246（选择器与 Fixtures · 已完成）

---

## 1. 背景与目标

- 背景：统一 Hook `useTemporalEntityDetail` 已落地并被消费，但组织侧缺少对外薄封装，且缺少单元测试覆盖；同时，Plan 246 的守卫脚本可冻结旧前缀，但未禁止新增硬编码 `data-testid`。  
- 目标：补齐组织薄封装、完善统一 Hook 的行为测试，并通过 ESLint + Guard 双门禁禁止硬编码选择器与旧前缀增长，确保 Selector 唯一事实来源。

---

## 2. 范围与产物

- 新增 Hook 薄封装：
  - `frontend/src/shared/hooks/useOrganizationDetail.ts`：仅透传 `useTemporalEntityDetail('organization', ...)`；提供必要的参数规范与类型导出
- 统一 Hook 单测：
  - `frontend/src/shared/hooks/__tests__/useTemporalEntityDetail.test.ts`：覆盖错误状态、租户切换、`includeDeleted`、失效/刷新与无重复 fetch
- 选择器门禁：
  - ESLint 自定义规则：禁止在 `frontend/src/**/*.{ts,tsx}` 直接硬编码 `data-testid` 字面量（白名单仅允许从 `shared/testids/temporalEntity.ts` 导入）
  - 继续使用 `scripts/quality/selector-guard-246.js`，并创建/更新基线 `reports/plan246/baseline.json`

---

## 3. 验收标准

1) 组织薄封装可用：组织侧读取统一 Hook 数据时不再直接拼装参数或绕开统一入口  
2) 统一 Hook 单测通过：覆盖场景的断言稳定，React Query 失效/刷新链路行为正确，无重复 fetch  
3) ESLint 规则生效：在非白名单文件中新增硬编码 `data-testid` 会报错；CI 开启该规则  
4) Guard 基线不升：`npm run guard:selectors-246` 通过，当前工作集对旧前缀无新增使用

---

## 4. 依赖与边界

- 只涉及前端 Hook 与规则；不涉及 GraphQL/REST 契约变更
- 选择器唯一来源：`frontend/src/shared/testids/temporalEntity.ts`
- 文档唯一来源：`docs/reference/temporal-entity-experience-guide.md`

---

## 5. 风险与缓解

| 风险 | 影响 | 概率 | 缓解 |
|---|---|---|---|
| ESLint 规则误伤历史代码 | 中 | 中 | 采用“只拦新增/改动文件”的策略，或提供短期 allowlist；同时保留 Guard 基线 |
| 单测依赖网络/容器 | 低 | 低 | 使用 mock/请求代理，避免真实网络；遵循仓库“网络受限”策略 |

---

## 6. 执行步骤

1) 新增 `useOrganizationDetail.ts` 薄封装并导出公共类型  
2) 为 `useTemporalEntityDetail` 编写 Vitest 覆盖；必要时对 GraphQL/REST 客户端 mock  
3) 在 `frontend/eslint.config.js` 增设“禁止硬编码 data-testid”规则并启用到 CI  
4) 跑 `npm run guard:selectors-246`，若无基线则建立，否则保证不升

---

## 7. 产出与登记

- 代码：`frontend/src/shared/hooks/useOrganizationDetail.ts`、`frontend/src/shared/hooks/__tests__/useTemporalEntityDetail.test.ts`、`frontend/eslint.config.js` 规则更新  
- 日志：`logs/plan241/B/hook-tests.log`、`logs/plan241/B/selector-guard.log`

---

## 8. 退出准则

- 单测通过；ESLint 与 Guard 门禁在本地与 CI 均生效；无新增硬编码 testid；Guard 基线不升  
- 改动登记于本文件并在 241 主计划更新“阶段性结论”与“里程碑达成”

