# E2E 测试指南

18 号《E2E 测试完善计划》定义了端到端测试的范围、流程与质量门禁。本指南汇总最新约定，帮助团队在本地与 CI 环境稳定运行 Playwright 测试。

## 快速开始

1. **准备基础服务**
   ```bash
   make docker-up
   make run-auth-rs256-sim
   ```
   - PostgreSQL、Redis、命令服务（9090）和查询服务（8090）启动后再继续。
   - `run-auth-rs256-sim` 会确保 RS256 密钥对存在并提供 JWKS。

2. **生成开发 JWT**
   ```bash
   make jwt-dev-mint
   cat .cache/dev.jwt
   ```
   - 默认租户：`3b99930c-4dc6-4cc9-8e4d-7d960a931cb9`
   - 命令失败时先确认 `http://localhost:9090/health` 返回 200。

3. **运行全部 E2E 测试**
   ```bash
   cd frontend
   PW_JWT=$(cat ../.cache/dev.jwt) \
   PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 \
   npm run test:e2e
   ```

4. **查看报告**
   ```bash
   npx playwright show-report
   ```
   - 失败案例会保留 trace、video、screenshot。

## 测试套件结构

| 套件 | 文件 | 说明 |
|------|------|------|
| 业务流程 | `tests/e2e/business-flow-e2e.spec.ts` | 覆盖创建、编辑、删除等完整 CRUD 剧本；长流程默认超时 180s。|
| 基础功能 | `tests/e2e/basic-functionality-test.spec.ts` | 核心入口与组织列表可用性；调试用 `/test` 剧本已跳过。|
| 优化验证 | `tests/e2e/optimization-verification-e2e.spec.ts` | 性能与前端优化验证，用于采集页面加载指标。|
| 回归测试 | `tests/e2e/regression-e2e.spec.ts` | 关键回归场景回放，保障历史问题不复现。|

单独执行示例：

```bash
PW_JWT=$(cat ../.cache/dev.jwt) PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 \
npm run test:e2e -- tests/e2e/regression-e2e.spec.ts
```

常用 `data-testid`：
- `create-organization-button`、`back-to-organization-list`
- `table-row-{code}`、`status-pill-{code}`、`temporal-manage-button-{code}`
- `form-field-effective-date`、`form-field-name`、`form-submit-button`
- `temporal-delete-record-button`、`deactivate-confirm-button`

## 编写规范

- **选择器策略**：所有交互元素必须暴露 `data-testid`，避免依赖可视文本。先使用 `page.waitForSelector('[data-testid="..."]')` 再操作。
- **断言一致性**：界面展示与 API 值不一致时，通过映射表统一断言，例如状态字段使用 `ACTIVE → "✓ 启用"`。
- **超时管理**：长流程使用 `test.setTimeout(180000)` 或拆分多个场景，避免静态 `waitForTimeout`。必要的等待必须基于条件。
- **诊断资产**：`playwright.config.ts` 已启用 `trace/video/screenshot` 的 `retain-on-failure`，无需在测试内重复截图逻辑。
- **安全与租户**：统一从 `.cache/dev.jwt` 读取令牌，并确保 `PW_TENANT_ID` 始终与测试数据一致。

## 调试技巧

- **Inspector**：`npm run test:e2e -- --debug`
- **单步定位**：在用例内插入 `await page.pause()`，配合 Inspector 观察 DOM 变化。
- **Trace Viewer**：失败后执行 `npx playwright show-report`，点击对应用例的 `Trace`。
- **HAR 捕获**：在 `business-flow` 套件中启用 `context.tracing.start({ screenshots: true, snapshots: true })` 即可导出关键请求。

## CI 质量门禁

- `.github/workflows/e2e-tests.yml` 会在 PR 上运行 Playwright 并阻止未通过的合并。
- 关键环境变量通过 `make jwt-dev-mint` 自动注入，测试失败时会上传 `frontend/playwright-report` 与 `frontend/test-results`。
- PR 作者应在描述中附上本地或 CI 运行截图，确认证据与报告一致。

## 常见问题

| 问题 | 处理办法 |
|------|----------|
| `make jwt-dev-mint` 执行失败 | 确保 9090/8090 健康，或重新执行 `make run-auth-rs256-sim`。|
| Playwright 报错 `401 Unauthorized` | 检查 `PW_JWT` 是否为空或过期，重新 mint 后重跑。|
| 前端选择器失效 | 给组件添加稳定的 `data-testid`，并在测试中使用同样的标识符。|
| CI 超时 | 减少长流程中的固定等待，使用条件等待并拆分测试用例。|

如需补充脚本或新增套件，请更新此指南并同步到 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`。
