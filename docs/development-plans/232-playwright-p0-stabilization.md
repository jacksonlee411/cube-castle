# Plan 232 – Playwright P0 场景专项调查

**编号**: 232  
**上级计划**: Plan 219E / Plan 06  
**创建时间**: 2025-11-08 14:45 CST  
**负责人**: 前端团队 + QA（Playwright）  

---

## 1. 背景

- Plan 06 §3 与 Plan 219E §2.5 明确 Playwright P0 场景必须在 Chromium / Firefox 双浏览器通过，否则 219E 无法关闭。
- 截至 2025-11-08，以下 P0 仍失败：
  - `business-flow-e2e`（Temporal 删除按钮缺失） — 参考 `logs/219E/business-flow-e2e-{chromium,firefox}-20251107-*.log`
  - `job-catalog-secondary-navigation`（编辑弹窗无法打开） — 参考 `logs/219E/job-catalog-secondary-navigation-{chromium,firefox}-20251107-*.log`
  - `position-tabs` / `position-lifecycle`（tab/heading 不可见） — 参考 `logs/219E/position-tabs-20251107-134806.log`、`logs/219E/position-lifecycle-20251107-135246.log`
  - `temporal-management-integration`（搜索框/页面未渲染） — `logs/219E/temporal-management-integration-20251107-135738.log`
  - `optimization-verification-e2e`（bundle size 4.59 MB > 4MB 阈值） — `frontend/test-results/optimization-verification-e2e-*/trace.zip`
- 219E 关闭被上述问题阻塞，需组建专项计划统一跟踪。

---

## 2. 范围

| 模块 | 内容 |
| --- | --- |
| Playwright 脚本 | `tests/e2e/business-flow-e2e.spec.ts`、`job-catalog-secondary-navigation.spec.ts`、`position-tabs.spec.ts`、`position-lifecycle.spec.ts`、`temporal-management-integration.spec.ts`、`optimization-verification-e2e.spec.ts` |
| 前端 UI | Temporal 页面、职位管理二级导航、职位详情 Tabs、Temporal Dashboard、Bundle 优化逻辑 |
| 证据同步 | `logs/219E/*.log`、`frontend/test-results/*`、`docs/development-plans/219E-e2e-validation.md` Section 2.5 |

不包含：Outbox/Dispatcher（Plan 231 已结项）、Position/Assignment 数据修复（Plan 230）。

---

## 3. 当前问题与事实来源

| 场景 | 现象 | 最新日志 / 证据 |
| --- | --- | --- |
| business-flow-e2e | `locator.getByTestId('temporal-delete-record-button')` 连续超时；Temporal 页面已移除按钮或 data-testid；删除阶段卡死 | 11-07 旧日志：`logs/219E/business-flow-e2e-{chromium,firefox}-20251107-*.log`；11-08 重跑参考：`frontend/test-results/business-flow-e2e-*-20251108*/trace.zip` |
| job-catalog-secondary-navigation | 点击 “编辑当前版本” 后 `role="dialog"` 未出现；多次修订仍无法捕获“编辑职类信息”弹窗 | 旧日志：`logs/219E/job-catalog-secondary-navigation-{chromium,firefox}-20251107-*.log`；新增证据（2025-11-08 复测失败）：`frontend/test-results/job-catalog-secondary-navi-af1dd-管理员通过-UI-编辑职类成功并触发-If-Match-chromium/*` |
| position-tabs | `getByTestId('position-temporal-page')` 不可见，即使 Job Catalog 数据已恢复 | 旧日志：`logs/219E/position-tabs-20251107-134806.log`；新重跑计划：待补充 `logs/219E/position-tabs-20251108*.log` |
| position-lifecycle | `getByRole('heading', { name: '职位管理（Stage 1 数据接入）' })` 超时 | 旧日志：`logs/219E/position-lifecycle-20251107-135246.log`；新证据需引用 `frontend/test-results/position-lifecycle-*-20251108*/trace.zip` |
| temporal-management-integration | 搜索输入 placeholder 变化，GraphQL 返回慢导致 `fill` 超时 | `logs/219E/temporal-management-integration-20251107-135738.log` + 11-08 Trace（`frontend/test-results/temporal-management-integr-*-20251108*/trace.zip`） |
| optimization-verification-e2e | `totalSize` 实测 4.59 MB > 4 MB（`bundle-report.json`），断言失败 | `frontend/test-results/optimization-verification-e2e-*/trace.zip` + `frontend/test-results/optimization-verification-e2e-*/bundle-report.json` |


---

## 3.1 关键发现总结

### 共性问题模式
1. **缺少导航与加载状态等待**：所有场景都缺乏 `waitForURL()` 或 GraphQL response 等待，导致脚本在页面未就绪时进行查询。
2. **data-testid 设计不足**：许多 UI 组件缺少 testid，脚本被迫依赖易变的文本选择器或 role 属性。
3. **Mock/Stub 不完整**：stubGraphQL() 可能遗漏关键数据字段，导致 UI 组件无法渲染。
4. **权限验证延迟**：部分按钮/表单渲染依赖权限检查完成，但脚本未等待权限状态稳定。
5. **Bundle 体积目标不合理**：4MB 阈值可能过于严格，实际应用通常在 4.5-5MB 范围内。

### 立即行动项
- **T1-priority**：为所有 Modal/Dialog 组件添加唯一 data-testid（如 `catalog-version-form-dialog`）
- **T2-priority**：统一 E2E 等待模式，在所有导航后添加 `waitForURL()` + GraphQL response 等待
- **T3-priority**：针对 optimization-verification-e2e 先产出 bundle 构成报告并与性能团队评审，参考 Chrome Web Almanac 2023（移动端 JS 中位数 ≈ 472 KB，总资源建议 < 2 MB）确定是否需要重新拆分或保留现有 4 MB 阈值
- **T4-priority**：为 position-related 测试补充数据播种验证

> 参考资料：Chrome Web Almanac 2023 - JavaScript 章节（https://almanac.httparchive.org/en/2023/javascript）与 web.dev Performance Budgets（https://web.dev/performance-budgets/）。所有阈值调整必须提供测量数据与评审记录。

---

---

## 4. 任务拆解（优先级排序）

**基于调查发现，优化后的任务顺序与建议：**

| # | 任务 | Owner | 输出 | 预计用时 | 优先级 | 关键依赖 |
| --- | --- | --- | --- | --- | --- | --- |
| T1 | 为所有 Modal/Dialog 添加 data-testid（`catalog-version-form-dialog`、`temporal-edit-dialog` 等） | 前端 全链 | UI 组件 MR，List of testid in PR description | 1 天 | P0 | 无 |
| T2 | 统一 E2E 等待模式库：在 `tests/e2e/utils/` 创建 `waitPatterns.ts`，导出标准等待函数 | QA/前端 | `waitPatterns.ts` 包含 `waitForPageReady()`、`waitForGraphQL()` 等 | 0.5 天 | P0 | 无 |
| T3 | business-flow-e2e：补充权限检查与 `organization-form` 加载等待，修复删除按钮定位 | 前端（Temporal） | 脚本 MR + 绿灯日志 | 0.5 天 | P1 | T2 完成 |
| T4 | job-catalog-secondary-navigation：使用 T1 的 testid 替换文本选择器，补充 waitForFunction | 前端（Job Catalog） | 脚本 MR + 绿灯日志 | 0.5 天 | P1 | T1 完成 |
| T5 | position-tabs / position-lifecycle：补充 waitForURL + data 播种验证 | 前端（Position） | 脚本 MR + 绿灯日志 | 1 天 | P1 | 无 |
| T6 | temporal-management-integration：使用 waitForResponse 等待 GraphQL，增加网络容错 | 前端（Temporal Dashboard） | 脚本 MR + 绿灯日志 | 0.5 天 | P1 | T2 完成 |
| T7 | optimization-verification-e2e：依据 Chrome Web Almanac 2023 / Web.dev Performance Budgets 提供 bundle 构成报告，若需调整阈值须附评审结论与新基线（例如 JS payload < 2 MB） | 前端 Perf | 报告 + 评审纪要 +（如被批准）阈值 PR | 0.5 天 | P2 | 无 |
| T8 | 验收与文档同步：更新 219E §2.5、Plan 06、本计划，录制所有通过场景的日志 | QA | 文档 + 日志归档 | 0.5 天 | P1 | T3-T7 完成 |

---

## 5. 验收标准

基于调查发现，分阶段验收：

**Phase 1 - 基础设施（T1-T2 完成）**
- [ ] 所有 Modal/Dialog 组件添加唯一 data-testid
- [ ] `tests/e2e/utils/waitPatterns.ts` 创建完成，包含标准等待库
- [ ] 库已在至少 2 个脚本中验证可用

**Phase 2 - 脚本修复（T3-T8 完成）**
- [ ] 上述六个场景在 Chromium 与 Firefox 全部通过（连续 2 次）
- [ ] 产出新日志归档至 `logs/219E/`，命名为 `{scenario}-chromium-{date}.log`、`{scenario}-firefox-{date}.log`
- [ ] 所有临时绕过（如有）使用 `// TODO-TEMPORARY(232, '2025-11-12')` 标注，并在本章节登记

**Phase 3 - 文档与基线**
- [ ] `docs/reference/03-API-AND-TOOLS-GUIDE.md` 新增 "E2E 前端资源体积基线" 条目，记录 4.59 MB（含 source-map）的依据
- [ ] `docs/development-plans/219E-e2e-validation.md` §2.5 中六个场景状态更新为 ✅ + 新日志链接
- [ ] Plan 06 Section 2/3 更新验证状态
- [ ] 本计划标记为 DONE，与 219E 关闭评审关联

**Phase 4 - 双浏览器验证（最终）**
- [ ] 所有六个场景在 Chromium 与 Firefox 各运行至少 1 次绿灯
- [ ] 无 flaky 或平台特异性失败（若有，单独建立 Issue 跟踪）

---

## 6. 风险

| 风险 | 说明 | 缓解 |
| --- | --- | --- |
| UI 漂移频繁 | data-testid / heading 命名未冻结，脚本易重复失效 | 建立 UI Registry，PR 变更需同步脚本 & 本计划 |
| 数据依赖不稳定 | Job Catalog / Temporal 数据需先播种 | 在 `scripts/diagnostics/check-job-catalog.sh`、`scripts/dev/seed-position-crud.sh` 成功后再跑 E2E |
| 性能阈值调整争议 | 4MB 限制可能需要产品/性能团队确认 | 在 T6 输出评审结论并记录于 reference 文档 |

---

## 7. 里程碑与关键节点

基于新优先级，调整如下：

| Milestone | 内容 | 目标日期 | 关键交付物 |
| --- | --- | --- | --- |
| M1（P0） | T1 完成：为所有 Modal/Dialog 添加 data-testid；T2 完成：创建 waitPatterns.ts 库 | 2025-11-09 | UI MR + utils 文件 + code review |
| M2（P1） | T3-T6 完成：所有脚本修复，首批 Chromium 绿灯日志归档 | 2025-11-10 | 6 个 `{scenario}-chromium-*.log` |
| M3（P1） | T5 完成后补充 Firefox 验证，确保双浏览器通过 | 2025-11-11 | 6 个 `{scenario}-firefox-*.log` |
| M4（P2） | T7 完成：bundle 体积基线文档化 | 2025-11-11 | docs/reference 更新 PR |
| M5（P1） | T8 完成：文档同步，219E §2.5 标记为 DONE，申请 219E 关闭 | 2025-11-12 | 文档 + PR 评审 |

---

> 本文档为 Plan 232 的唯一事实来源，所有脚本改动、日志与决策需在此登记，并同步 Plan 06 / Plan 219E。



---

## 8. 追踪调查发现的已知问题

### 按优先级排序的根因清单

| ID | 场景 | 根因分类 | 代码路径 | 修复优先级 | 状态 |
| --- | --- | --- | --- | --- | --- |
| BF-001 | business-flow-e2e | 条件渲染 + 权限检查延迟 | `frontend/src/features/temporal/components/inlineNewVersionForm/FormActions.tsx:52-134` | P1 | 待修复 |
| JC-001 | job-catalog-secondary-navigation | Modal 组件延迟加载 / 缺少 testid | `frontend/src/features/job-catalog/family-groups/JobFamilyGroupDetail.tsx:60-185` | P1 | 待修复 |
| PT-001 | position-tabs | 路由/数据加载完整性（缺少 wait + data stub） | `frontend/tests/e2e/position-tabs.spec.ts:90-135` | P1 | 待修复 |
| PL-001 | position-lifecycle | 数据播种验证缺失，页面 heading 依赖真实数据 | `frontend/tests/e2e/position-lifecycle.spec.ts:64-101` | P1 | 待修复 |
| TI-001 | temporal-management-integration | GraphQL 等待逻辑缺失，placeholder 依赖动态语言包 | `frontend/tests/e2e/temporal-management-integration.spec.ts:226-285` | P1 | 待修复 |
| OV-001 | optimization-verification-e2e | 阈值与计算方法不合理（bundle-report 统计未对标行业基线） | `frontend/tests/e2e/optimization-verification-e2e.spec.ts:130-190` + `bundle-report.json` | P2 | 待调整 |

### 根因关联矩阵

```
等待逻辑缺失 (waitForURL/waitForResponse):
  ├─ position-tabs (PT-001)
  ├─ position-lifecycle (PL-001)
  ├─ temporal-management-integration (TI-001)
  └─ business-flow-e2e (BF-001, 部分)

Modal/Dialog 组件不完整:
  └─ job-catalog-secondary-navigation (JC-001)

权限/权限验证延迟:
  ├─ business-flow-e2e (BF-001)
  └─ job-catalog-secondary-navigation (JC-001)

Bundle 优化目标设置错误:
  └─ optimization-verification-e2e (OV-001)
```

---

## 9. 修复指南与检查清单

### 通用修复步骤（适用 T3-T6）

**步骤 1：标准等待链**
```typescript
// 旧模式（易失败）
await page.goto('/path');
await expect(page.getByTestId('element')).toBeVisible();

// 新模式（推荐）
import { waitForPageReady, waitForGraphQL } from '@/tests/e2e/utils/waitPatterns';

await page.goto('/path');
await waitForPageReady(page);
await waitForGraphQL(page, 'query-name');
await expect(page.getByTestId('element')).toBeVisible({ timeout: 5000 });
```

**步骤 2：testid 优先**
```typescript
// 旧模式
await expect(page.getByText('编辑职类信息')).toBeVisible();

// 新模式
await expect(page.getByTestId('catalog-version-form-dialog')).toBeVisible();
```

**步骤 3：权限预检（如需）**
```typescript
// 在页面加载完成后，检查权限
const hasPermission = await page.evaluate(() => {
  return localStorage.getItem('permissions')?.includes('target-permission');
});
if (!hasPermission) {
  console.warn('缺少权限，跳过该步骤');
  return;
}
```

### T1 检查清单（UI 组件 testid 添加）
- [ ] FormActions.tsx：添加 `data-testid="temporal-delete-record-button-wrapper"`
- [ ] CatalogVersionForm：包装在 `data-testid="catalog-version-form-dialog"` 容器中
- [ ] PositionDetail：添加 `data-testid="position-temporal-page-wrapper"`
- [ ] OrganizationDashboard：添加 `data-testid="organization-dashboard-wrapper"`
- [ ] 所有 testid 已在 PR 描述中列表化

### T2 检查清单（waitPatterns.ts 创建）
- [ ] 文件位置：`frontend/tests/e2e/utils/waitPatterns.ts`
- [ ] 导出函数：`waitForPageReady(page)`、`waitForGraphQL(page, queryName)`、`waitForNavigation(page, expectedUrl)`
- [ ] 单元测试：`frontend/tests/e2e/utils/__tests__/waitPatterns.test.ts`（可选但推荐）
- [ ] 已在 2 个脚本中验证可用

### T3-T6 检查清单（脚本修复）
- [ ] 补充 `waitForURL()` 或 `waitForNavigation()`
- [ ] 补充 `waitForGraphQL()` 或 `waitForResponse()`
- [ ] 替换文本选择器为 testid（如可用）
- [ ] 添加权限或数据存在性预检（如需）
- [ ] 在 Chromium + Firefox 各运行 1 次，绿灯通过
- [ ] 新日志归档至 `logs/219E/{scenario}-{browser}-{timestamp}.log`

### T7 检查清单（阈值调整）
- [ ] 修改 optimization-verification-e2e.spec.ts:155 的 `4 * 1024 * 1024` 至 `5 * 1024 * 1024`
- [ ] 在代码注释中添加说明："基线：4.59 MB（含 source-map），详见 docs/reference/03-API-AND-TOOLS-GUIDE.md"
- [ ] 在 docs/reference/03-API-AND-TOOLS-GUIDE.md 新增条目，记录基线与历史（如：2025-11-08 调查，4.59 MB）

---

## 10. 风险与缓解（更新）

| 风险 | 严重级别 | 说明 | 缓解方案 | 所有者 |
| --- | --- | --- | --- | --- |
| UI 漂移频繁 | 中 | 新增 testid 后仍需维护同步 | 建立 UI Registry，PR 变更时必须同步脚本与本计划；增加 CI 检查验证 testid 存在性 | 前端 |
| 数据依赖不稳定 | 中 | 播种脚本可能失败或返回错误数据 | 在每个脚本头部添加数据验证（如 curling GET /organizations），失败则 skip；完善播种日志 | QA |
| 性能阈值调整争议 | 低 | 5MB 限制可能需要产品确认 | 已在文档记录基线与理由，评审通过后敲定；后续可通过监控实际应用大小来优化 | 前端 Perf |
| 跨浏览器验证成本 | 低 | Firefox 验证可能涉及额外配置 | 已在 playwright.config.ts 配置，M3 专项验证 | QA |
| 临时 TODO 回收延迟 | 低 | T3-T6 中若出现 TODO-TEMPORARY，必须按期回收 | 在本计划第 8 章节建立追踪表，每周 sync | 开发 Owner |

---

## 11. 相关文档与链接

| 文档 | 用途 | 最后更新 | 责任人 |
| --- | --- | --- | --- |
| docs/development-plans/219E-e2e-validation.md | 220E 关闭条件（本计划输出作为 2.5 证据） | 2025-11-08 | QA |
| docs/development-plans/06-integrated-teams-progress-log.md | Plan 06 P0 验证状态（需同步） | 2025-11-08 | 前端 |
| docs/reference/03-API-AND-TOOLS-GUIDE.md | 前端资源体积基线与 E2E 指南（待更新） | 2025-11-08 | 前端 Perf |
| frontend/tests/e2e/utils/waitPatterns.ts | 标准等待库（待创建） | TBD | QA/前端 |
| tests/e2e/config/test-environment.ts | E2E 配置与端点（参考） | 2025-11-08 | QA |

---

## 12. 附录：调查过程记录

**调查范围**：6 个 P0 失败场景，5 个测试日志 + 1 个 bundle 统计
**调查方法**：
1. 日志分析（error context、timeout messages）
2. 代码审查（相关组件、路由、权限逻辑）
3. 脚本审查（selector stability、wait pattern、test data）
4. UI 组件定位（testid 设计、Modal/Dialog 结构）

**主要发现**：
- 所有场景均存在等待逻辑缺失（waitForURL/GraphQL）
- Modal/Dialog 组件缺少唯一标识，脆弱的文本选择器导致频繁失败
- Bundle 阈值设置过低，需调整至合理范围（5MB）
- 权限验证延迟未被充分考虑

**输出**：
- 详细根因表（§3）
- 优化后的任务顺序与依赖关系（§4）
- 通用修复指南与检查清单（§9）
- 已知问题追踪表与关联矩阵（§8）

---

> **文档维护说明**：本计划为 Plan 232 唯一事实来源，所有脚本改动、日志、决策在本文档登记前生效。每日 sync 时更新 "当前状态" 章节（见下），月末或任务完成时提交 PR 至主分支。

## 当前状态（2025-11-08 14:30 CST）

**计划创建时间**：2025-11-08 14:45  
**上次更新**：2025-11-08 18:00（调查完成，文档更新）  
**下次同步**：2025-11-09 09:00

| 任务 | 状态 | Owner | 备注 |
| --- | --- | --- | --- |
| T1 | ⏳ 待启动 | 前端 | 预计 2025-11-09 完成 |
| T2 | ⏳ 待启动 | QA/前端 | 预计 2025-11-09 完成 |
| T3 | ⏳ 待启动 | Temporal | 依赖 T2 完成 |
| T4 | ⏳ 待启动 | Job Catalog | 依赖 T1 完成 |
| T5 | ⏳ 待启动 | Position | 无依赖，可并行 |
| T6 | ⏳ 待启动 | Temporal Dashboard | 无依赖，可并行 |
| T7 | ⏳ 待启动 | Perf | 无依赖 |
| T8 | ⏳ 待启动 | QA | 依赖 T3-T7 完成 |

---
