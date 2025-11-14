# Plan 246 – Temporal Entity Selector & Fixture 统一（完善版）

关联主计划: Plan 242（T4）  
目标窗口: Day 12-15  
范围: 统一 Playwright selector/testid、测试 fixtures 与 E2E 工具  
状态: 已完成（2025-11-14）— Phase 1 交付，Phase 2 组件 testid 收敛按迭代小步推进

—

## 1. 背景
- 现有 E2E 用例使用 `organization-*`、`position-*` testid 与专有 fixtures（如 frontend/tests/e2e/position-tabs.spec.ts），维护成本高，选择器命名易漂移。
- Plan 242 要求以“Temporal Entity”作为中性抽象统一命名；本计划负责在“选择器与测试资产层”落地唯一事实来源，避免新增旧命名。

权威参考
- 选择器与 E2E 指南: docs/development-tools/e2e-testing-guide.md  
- 契约（只读）: docs/api/openapi.yaml, docs/api/schema.graphql  
- 架构门禁: scripts/quality/architecture-validator.js（contracts/ports/forbidden 等）

—

## 2. 前置条件（Docker 强制 + 鉴权）
- 环境前置
  - make docker-up → make run-dev（9090/8090 健康）  
  - make jwt-dev-mint（.cache/dev.jwt 存在）  
  - 快速验证：`curl http://localhost:9090/health`、`curl http://localhost:8090/health` → 200
- E2E 前置
  - 在 frontend 下执行：`PW_JWT=$(cat ../.cache/dev.jwt) PW_TENANT_ID=<tenant-uuid> npm run test:e2e`
  - 允许 Mock 回退：当后端不可用时由用例内切换到 Mock 模式（参考 temporal-management-integration.spec.ts）

—

## 3. 命名与选择器规范（唯一事实来源）
- 选择器常量位置调整
  - 新文件: `frontend/src/shared/testids/temporalEntity.ts`（运行时代码与测试共用）  
  - 导出对象: `temporalEntitySelectors`（统一常量与构造器）
- 命名风格
  - testid 字面值使用 kebab-case，前缀优先 `temporal-*`；可分层 `temporal-page-*` / `temporal-list-*` / `temporal-action-*`  
  - 动态选择器统一由函数构造，避免手写拼接  
    - 例如: `row(code) => \`temporal-row-\${code}\``，`manageButton(code) => \`temporal-manage-button-\${code}\``
- 渐进收敛策略
  - 第一阶段（不破坏）：`temporalEntitySelectors` 的值映射到现有 testid（如 `position-dashboard`、`organization-dashboard-wrapper` 等），先统一“引用点”，保持零行为变更。
  - 第二阶段（收敛命名）：组件内逐步将 `data-testid` 替换为新 `temporal-*` 值，仅需更新 `temporalEntitySelectors` 映射，测试无感知。

—

## 4. 目录与资产规划
- 选择器常量
  - 新增: `frontend/src/shared/testids/temporalEntity.ts`（导出常量 + 构造器 + 旧命名别名）
- E2E fixtures
  - 新增: `frontend/tests/e2e/utils/temporalEntityFixtures.ts`  
    - `createFixtures(entityType, options): { graphql, rest }`  
    - GraphQL 操作名与字段集遵循 docs/api/schema.graphql，不引入新的“第二事实来源”
  - 兼容期: `frontend/tests/e2e/utils/positionFixtures.ts` 仅 re-export 并打印弃用告警；一个迭代后删除
- 工具与 util
  - `waitPatterns`, `auth-setup` 维持路径不变，但命名去实体化（避免 `position*` 前缀）

—

## 5. 守卫与 CI 接入（防新增旧命名）
- 新增守卫脚本
  - `scripts/quality/selector-guard-246.js`  
    - 首次运行生成 `reports/plan246/baseline.json`（记录仓库中 `data-testid="(organization|position)-"` 及 `getByTestId('(organization|position)-')` 的计数）  
    - 后续运行：计数上升则失败（error）；允许 `reports/plan246/allowlist.txt`
    - 匹配范围（覆盖真实用法，减少漏检）：
      - 直接属性：`data-testid="(organization|position)-..."`、`data-testid='(organization|position)-...'`
      - Attribute Locator：`\[data-testid\s*=\s*['"\`](organization|position)-`
      - 前缀匹配：`\[data-testid\^\s*=\s*['"\`](organization|position)-`
      - 测试 API：`getByTestId\(\s*['"\`](organization|position)-`
      - 模板字面量：<code>getByTestId\\(\\s*\`(organization|position)-\\$\\{</code>（检测以字面量起始的模板字符串）
    - 输出内容：新旧计数对比、触发文件清单、提示替代的 `temporalEntitySelectors` 名称
- NPM 脚本与 CI
  - package.json 增加 `"guard:selectors-246": "node scripts/quality/selector-guard-246.js"`
  - 在 `.github/workflows/frontend-quality-gate.yml` 与 `.github/workflows/agents-compliance.yml` 中，在安装依赖与构建前（fail-fast）执行：  
    - `npm run guard:plan245`（既有）  
    - `npm run guard:selectors-246`（新增）

—

## 6. Codemod 方案（测试侧先行）
- 目标：替换测试代码中的硬编码 `getByTestId('position-*'|'organization-*')` 为 `temporalEntitySelectors.*`
- 步骤
  1) 建立映射表（以“不破坏”为原则）：  
     - `position-dashboard` → `selectors.position.dashboard`  
     - `organization-dashboard-wrapper` → `selectors.organization.dashboardWrapper`  
     - `table-row-{code}` → `selectors.list.row(code)`  
     - `temporal-manage-button-{code}` → `selectors.action.manageButton(code)`  
  2) jscodeshift/简单正则脚本分批替换；对无法自动匹配的场景出具 TODO 注释
  3) 验证（见第 8 节），提交 MR

—

## 7. 验收标准（可度量）
- 引用统一
  - 测试与组件侧均通过 `frontend/src/shared/testids/temporalEntity.ts` 引用选择器常量  
  - 仓库中对 `data-testid="(organization|position)-"` 的新增使用计数为 0（selector-guard-246 通过）
  - ESLint 禁止硬编码：新增规则禁止除 `frontend/src/shared/testids/temporalEntity.ts` 外任何位置直接出现 `data-testid="..."` 字面量
- fixtures 统一
  - `positionFixtures.ts` 仅保留 re-export（含弃用告警），新增用例全部改用 `temporalEntityFixtures.ts`
- 稳定性门禁
  - Chromium/Firefox 各运行 3 次：`position-tabs.spec.ts`、`organization-create.spec.ts`、`temporal-management-integration.spec.ts`  
  - 失败率 < 5%（按 6 次总运行计）；报告归档至 `logs/plan242/t4/`，并上传 `frontend/playwright-report`
- 文档同步
  - 更新 docs/development-tools/e2e-testing-guide.md：新增“temporal 选择器清单与编写规范”  
  - 在 `215-phase2-execution-log.md` 记录执行证据（命令与关键截图/trace）

—

## 8. 验证与留痕
- 本地指令
  - `node scripts/quality/selector-guard-246.js`（首次生成基线后再次运行验门禁）
  - `cd frontend && npm run test:e2e -- tests/e2e/position-tabs.spec.ts`（重复运行 3 次 × 2 浏览器）
  - `npx playwright show-report`（本地查看）
- 日志与报告
  - `logs/plan242/t4/`：收集 guard、e2e、typecheck、vitest 输出  
  - `frontend/playwright-report`：作为 CI artifact

—

## 9. 风险与回滚
- 风险
  - 选择器替换引发短期不稳定；组件 testid 与测试不一致  
  - fixtures 与 GraphQL 操作名不对齐导致 mock 失效
- 缓解
  - 分阶段：先统一“引用点”，再逐步收敛组件内 testid 值  
  - 在 `temporalEntitySelectors` 中保留旧命名别名（仅导出，不在新代码中引用）
  - 对关键组件 PR 增加可视化回归（截图对照/关键元素存在性断言）
- 回滚
  - 守卫降级为 warn：`SELECTOR_GUARD_STRICT=0 node scripts/quality/selector-guard-246.js`  
  - 保留 alias 时间上限 1 个迭代；如需延长，必须在计划文档记录原因并设新截止日期；所有 alias/re-export/allowlist 项必须附 `// TODO-TEMPORARY: 原因/计划/截止日期` 注记

—

## 10. 里程碑
- Day 14：提交 selectors + fixtures MR，Playwright 绿灯（3 次 × 2 浏览器）  
- Day 15：alias 清理与文档更新（指南 + 执行日志），CI 接入守卫生效

—

## 11. 变更影响与不做的事
- 不改动 API 契约与 GraphQL schema；仅统一选择器与测试资产命名  
- 不引入新的“选择器第二事实来源”；唯一来源为 `frontend/src/shared/testids/temporalEntity.ts`

—

## 12. 结项产物
- 代码
  - `frontend/src/shared/testids/temporalEntity.ts`（新）  
  - `frontend/tests/e2e/utils/temporalEntityFixtures.ts`（新）  
  - `scripts/quality/selector-guard-246.js`（新）
- 文档
  - 本计划（完善版）与 e2e 指南更新  
  - `logs/plan242/t4/` 留痕（guard、e2e 报告、截图/trace 索引）

—

## 16. 完成说明与证据
- 守卫基线：`reports/plan246/baseline.json`（total=125）
- 当前扫描：`npm run guard:selectors-246` → total=83（legacy -42），通过
- CI 集成：已在 `agents-compliance.yml` 与 `frontend-quality-gate.yml` 中启用 Plan 246 Guard（安装依赖/构建前）
- 用例采纳（示例）
  - 组织：`business-flow-e2e.spec.ts`、`basic-functionality-test.spec.ts`、`organization-create.spec.ts`、`canvas-e2e.spec.ts`、`temporal-editform-defaults-smoke.spec.ts`、`optimization-verification-e2e.spec.ts`
  - 职位：`position-tabs.spec.ts`、`position-crud-live.spec.ts`、`position-crud-full-lifecycle.spec.ts`、`position-lifecycle.spec.ts`、`job-catalog-layout-baseline.spec.ts`
- SSoT 扩展
  - `organization.{dashboardWrapper,dashboard,form,table}`
  - `position.{dashboard,temporalPageWrapper,temporalPage,overviewCard,versionToolbar,versionList,detailCard,tabVersions}`

—

## 13. 附录：选择器与类型规范（代码片段）
```ts
// frontend/src/shared/testids/temporalEntity.ts
export type TemporalSelectors = {
  page: {
    wrapper: string;         // e.g. 'temporal-master-detail-view'
    timeline: string;        // e.g. 'temporal-timeline'
  };
  list: {
    table: string;           // e.g. 'temporal-entity-table'
    rowPrefix: string;       // e.g. 'temporal-row-'
    row: (code: string) => string; // => `${rowPrefix}${code}`
  };
  action: {
    manageButton: (code: string) => string; // => `temporal-manage-button-${code}`
    deleteRecord: string;    // e.g. 'temporal-delete-record-button'
  };
  organization: {
    dashboardWrapper: string; // e.g. 'organization-dashboard-wrapper' (兼容期内映射旧值)
  };
  position: {
    dashboard: string;        // e.g. 'position-dashboard' (兼容期内映射旧值)
  };
};

export const temporalEntitySelectors: TemporalSelectors = {
  page: {
    wrapper: 'temporal-master-detail-view',
    timeline: 'temporal-timeline',
  },
  list: {
    table: 'temporal-entity-table',
    rowPrefix: 'temporal-row-',
    row: (code) => `temporal-row-${code}`,
  },
  action: {
    manageButton: (code) => `temporal-manage-button-${code}`,
    deleteRecord: 'temporal-delete-record-button',
  },
  // 兼容期：值映射到现有 testid，后续组件内收敛为 temporal-* 再切换为新值
  organization: { dashboardWrapper: 'organization-dashboard-wrapper' },
  position: { dashboard: 'position-dashboard' },
} as const;
```

—

## 14. 附录：ESLint 禁止硬编码 testid（策略）
- 目标：在组件与测试中禁止直接硬编码 `data-testid="..."`；仅允许在 `frontend/src/shared/testids/temporalEntity.ts` 定义
- 建议规则（添加到 `frontend/.eslintrc.api-compliance.cjs`）：
  - `no-restricted-syntax` 或自定义规则，匹配包含 `data-testid="..."` 的 JSXAttribute/字符串字面量
  - 允许文件白名单：`frontend/src/shared/testids/temporalEntity.ts`
  - 例外需 `// TODO-TEMPORARY:` 注记，并纳入 allowlist 文件（与 selector-guard-246 同步）

—

## 15. 附录：fixtures 强类型校验（契约从属）
- 选项 A（推荐）：使用 `graphql` 包 AST 解析 fixtures 对象的字段，并与 schema.graphql 选定的 SelectionSet 比对（无网络依赖）
- 选项 B：在 `npm run test:contract` 中执行 `@graphql-inspector/cli` 或轻量校验逻辑，确保 fixtures 字段集与 schema 同步
- 要求：不新增“字段/枚举常量”的第二事实来源，所有字段名均来自 schema.graphql 或 GraphQL 文本查询
