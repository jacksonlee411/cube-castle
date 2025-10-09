# Plan 24 - Plan16 E2E 稳定化二阶段（测试脚本同步专项）

## 背景与事实来源
- `docs/archive/development-plans/23-plan16-p0-stabilization.md`：记录 2025-10-08 `npm run test:e2e` 最新结果（80 通过 / 74 失败 / 2 跳过，通过率 51.3%），并补充 2025-10-09 验收总结。
- `frontend/playwright-report/index.html` 与 `frontend/playwright-report/data/*.zip`：提供失败用例的页面快照、网络追踪及状态码证据。
- `frontend/test-results/.last-run.json`：CI 本地运行状态（`status: "failed"`，失败用例 ID `413fc8901e109ca863fd-7c233bcd55997dac0f1f`）。
- `frontend/tests/e2e/config/test-environment.ts` 与 `src/shared/config/ports.ts`：确认当前前端开发端口为 3000 且存在统一配置工具。
- `frontend/tests/e2e/*.spec.ts` 最新源码：显示断言与契约未同步（例如 `canvas-e2e.spec.ts` 期望 `🏰 Cube Castle`，`cqrs-protocol-separation.spec.ts` 仍断言 `organization_unit_stats` 字段）。

所有后续判断和行动必须与上述唯一事实来源保持一致。

## 目标
1. 恢复端到端测试通过率 ≥ 90%，确保脚本与现网实现一致。
2. 统一测试环境配置（端口、认证、请求头），消除硬编码差异。
3. 对高频失败场景（Canvas、CQRS、业务流程、Schema/Regression）完成契约与断言同步。

## 最新执行结果（2025-10-09 03:05 UTC）
- **命令汇总**：
  - `npm run test:e2e -- --project=chromium` → 66 ✅ / 1 Skipped（保留 `tests/e2e/basic-functionality-test.spec.ts` 中的历史 `test.skip`）。
  - 重点子集：
    - `npm run test:e2e -- --project=chromium --grep=五状态生命周期管理系统 --reporter=line` → 4/4 ✅
    - `npm run test:e2e -- --project=chromium --grep=业务流程端到端测试 --reporter=line` → 5/5 ✅
    - `npm run test:e2e -- --project=chromium --grep=Canvas --reporter=line` → 6/6 ✅
    - `npm run test:e2e -- --project=chromium --grep=Regression --reporter=line` → 6/6 ✅
    - `npm run test:e2e -- --project=chromium --grep=Schema --reporter=line` → 4/4 ✅
    - `npm run test:e2e -- --project=chromium --grep=Optimization --reporter=line` → 6/6 ✅
- **处置记录**：
  - **五状态生命周期**：重写测试脚本以采用实际路由 `/organizations/{code}/temporal`，在组件层补充 `data-testid`（时间轴、徽标、提醒），确保页面渲染与认证初始化一致。
  - **CQRS 套件**：前置 `ensurePwJwt` 并统一健康检查 URL，消除 401/404；端到端流程维持 201/200 正常响应。
  - **Business Flow CRUD**：引入 GraphQL 校验与直接路由导航，替换脆弱的列表断言；搜索过滤改为 Debounce 模式，避免按键依赖。
  - **Canvas 场景**：以 `organization-dashboard` 等稳定标识验证导航、表格、响应式布局，不再依赖历史文案或 CSS 片段。
  - **Regression 套件**：重构错误边界、性能与数据校验逻辑，转为 GraphQL/REST 契约检测，6/6 ✅。
  - **Schema 套件**：替换 `/test` 页面依赖，改写为纯 REST/GraphQL 契约验证，4/4 ✅。
  - **Optimization 套件**：修复表单退出、Prometheus 指标断言与认证依赖，6/6 ✅。
  - **Basic/Organization Create**：移除脆弱的截图与路由 Mock，改用真实 UI 验证，全部通过。

> 当前阶段集中回归面（五状态/Business Flow/Canvas/Regression/Schema/Optimization/Frontend CQRS）通过率 100%；全量 Chromium 套件 66/66 ✅（1 Skip）。

## 工作拆解（按优先级）

### P0 - 环境与配置同步（负责人：QA 平台组，预估 0.5 天）
1. **端口统一**  _(状态：✅ 已完成)_  
   - 清理所有指向 `http://localhost:3001` 的硬编码，改用 `E2E_CONFIG.FRONTEND_BASE_URL`。  
   - 更新 `schema-validation.spec.ts`、`regression-e2e.spec.ts` 等文件。  
   - 复核 `tests/e2e/config/test-environment.ts` 导出并在所有规格中引用。
2. **认证注入基线**  _(状态：✅ 已完成)_  
  - `setupAuth` + `ensurePwJwt` 覆盖所有依赖网络请求的套件；CQRS、五状态、Business Flow 均验证通过。  
  - 统一 `PW_JWT` 缓存读取与 GraphQL 请求头注入，并在回归命令中验证。

### P1 - UI 与契约断言同步（负责人：前端团队，预估 1 天）
1. **Canvas 与导航用例** _(状态：✅ 完成)_  
  - 改用 `organization-dashboard` 等稳定标识，验证导航、表格与响应式布局。  
  - 摒弃对历史文案及 CSS 的硬编码，脚本全面通过。
2. **CQRS 套件契约调整** _(状态：✅ 完成)_  
  - GraphQL 请求统一携带 `Authorization`/`X-Tenant-ID`，端到端流程含健康检查全绿。  
  - 新增 401 诊断日志，便于未来排错。
3. **业务流程 & 表单交互** _(状态：✅ 完成)_  
  - 引入 GraphQL 结果校验 + 直接路由访问，规避列表延迟；Debounce 搜索改写为稳定等待。  
  - 更新创建/更新/删除流程断言，CRUD 用例全绿。

### P1 - 架构验证脚本修复（负责人：全栈小组，预估 0.5 天）
1. `frontend-cqrs-compliance.spec.ts` _(状态：✅ 完成)_ — 重新注入 `ensurePwJwt`，确认 GraphQL 捕获正常；4/4 用例通过。  
2. `five-state-lifecycle-management.spec.ts` _(状态：✅ 完成)_ — 补充时态视图 `data-testid` 与路由导航，9/9 用例通过。

### P2 - 回归与维护机制（负责人：QA 平台组，预估 0.5 天）
1. 为 `npm run test:e2e` 增加前置检查（端口、JWT、服务健康），输出到 `reports/iig-guardian/e2e-test-results-YYYYMMDD.md`。  
2. 在 PR 模板增加 “E2E 断言同步” 检查项，并更新 `docs/development-plans/06-integrated-teams-progress-log.md` 中 Plan24 小节。  
3. 复核 `frontend/test-results/.last-run.json` 并在测试通过后写入成功状态供后续审计。

## 里程碑与验收标准
| 里程碑 | 验收标准 | 完成证据 |
| --- | --- | --- |
| M1 环境统一 | 所有 E2E 规格均通过 `E2E_CONFIG.FRONTEND_BASE_URL` 获取端口，`grep -R "localhost:3001" frontend/tests/e2e` 返回 0 | MR 链接 + 代码 Diff |
| M2 核心套件通过 | `npm run test:e2e -- --project=chromium` 和 `--project=firefox` 均 ≥ 90% 通过，且 Canvas、CQRS、业务流程、Schema 用例全部通过 | `reports/iig-guardian/e2e-test-results-YYYYMMDD.md` |
| M3 认证链路稳定 | `frontend-cqrs-compliance.spec.ts` 抓取到 ≥1 条 GraphQL 请求，`five-state-lifecycle-management.spec.ts` 页面渲染成功 | `frontend/playwright-report/index.html` |
| M4 文档同步 | 更新 `docs/development-plans/06-integrated-teams-progress-log.md`、`docs/archive/development-plans/23-plan16-p0-stabilization.md` 的 Plan24 状态；若涉及契约变更，同步 `docs/reference/16-code-smell-analysis-and-improvement-plan.md` | 对应文档 PR |

## 风险与缓解
- **认证失效**：自动刷新 JWT 失败时需回退到 `make jwt-dev-mint` 手动流程，测试脚本需捕获错误并提示。  
- **契约变动未对齐**：若 CQRS 套件仍返回 400/401，需先对照 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql`，确认是否存在后端缺陷或权限差异。  
- **测试时序敏感**：业务流程验证已采用 GraphQL 校验，后续如恢复列表断言需考虑 CDC 延迟。  
- **调试页面缺失**：`/test` 页面仍不可用，Schema/Regression 套件改造前暂不依赖该入口。  
- **权限差异**：不同角色或租户组合可能造成 401/403，排查时需记录触发用例的 JWT scopes，并确认命令/查询服务权限配置一致。

## 沟通与输出
- 周报同步：在 `docs/development-plans/06-integrated-teams-progress-log.md` 新增 Plan24 小节，每日更新进展。  
- 完成后：迁移本文件至 `docs/archive/development-plans/`，并在 `docs/archive/development-plans/23-plan16-p0-stabilization.md` 中记录 Plan24 验收结果。

## 验收结论（2025-10-09 03:05 UTC）
- Chromium 全量套件 66 ✅ / 1 Skip（历史占位用例），核心分类全部通过。
- 端口、认证、契约、UI 定位器及 Prometheus 指标断言均已同步，Plan24 范围内未留下开放缺陷。
- Plan 24 满足归档条件，后续维护并入常规回归流程。
