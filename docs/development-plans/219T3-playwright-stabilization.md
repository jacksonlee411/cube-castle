# 219T3 – Playwright 用例稳定性子方案

## 1. 背景
- `npm run test:e2e` 报告 9 组双浏览器用例失败，原因包括 UI selector 变更、API 返回码更新、读模型缺失等。
- 219T 报告要求对 UI/契约/Mock 模式做区分，确保测试与实现保持一致。

## 2. 工作范围
1. **UI 选择器同步**：收集最新 data-testid/placeholder 规范，更新 `tests/e2e/*.spec.ts` 中的 selector。
2. **契约更新**：对 REST 返回 200/422 的接口调整断言，与 `docs/api/openapi.yaml` 保持一致。
3. **Mock/真实切换**：为依赖 GraphQL 实时数据的场景提供可配置 stub 或 fallback，以便在读模型未恢复前运行。
4. **结果归档**：修复后重新执行 Playwright，生成新的 `frontend/test-results/*` 目录，并在 219T 报告中登记。

## 3. 任务清单
| 编号 | 场景 | 处理策略 |
| --- | --- | --- |
| T3-1 | business-flow-e2e | 确认 `temporal-delete-record-button` 是否重命名或流程被移除，更新按钮定位或改用 API 校验 |
| T3-2 | job-catalog-secondary-navigation | 调查编辑弹窗未出现原因（UI 权限、异步条件），调整等待策略或修复页面 |
| T3-3 | name-validation-parentheses | 更新断言为 200，同时在 API 文档中记录括号合法性 |
| T3-4 | position-* 系列 | 补足必填字段、422 错误处理逻辑，并分离 CRUD 与错误用例 |
| T3-5 | position-tabs / position-lifecycle | 更新 GraphQL stub 或 data-testid，确保页面加载后元素存在 |
| T3-6 | temporal-management-integration | 确认 `/organizations` 页面加载时机，必要时增加 `waitForResponse` 或更换 selector |

## 4. 输出
- 更新后的 spec 文件及辅助工具（如 `tests/e2e/utils`）。
- 新的 Playwright 报告（`frontend/playwright-report/`）与 `frontend/test-results/`。
- 在 `docs/development-plans/219T-e2e-validation-report.md` “Playwright 用例整改” 条目中附上结论。

## 5. 验收
1. `npm run test:e2e` 在 Chromium/Firefox 均达到既定通过率（允许 Mock 场景跳过）。
2. 所有更新附带 changelog（如 README 或脚本说明）。
3. 对外契约（OpenAPI、GraphQL schema）同步完成。

---

> 唯一事实来源：`frontend/tests/e2e/`、`frontend/test-results/`、`docs/api/openapi.yaml`。  
> 更新时间：2025-11-07。
