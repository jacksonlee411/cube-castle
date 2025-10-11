# 06 号文档：测试任务与执行要求（最新版）

> 本文为集成团队在所有阶段执行测试的唯一事实来源。涉及环境准备、测试范围、执行命令、结果记录与缺陷反馈，请严格遵守。若与其他文档冲突，以本文件为准。

---

## 1. 环境准备（必须满足后方测试才可执行）

| 组件 | 目标状态 | 启动/校验命令 | 说明 |
|------|-----------|---------------|------|
| **PostgreSQL & Redis** | 已启动 | `make docker-up` | 仅需最小依赖，无需额外容器 |
| **命令服务 (REST, :9090)** | 运行中、/health 返回 200 | `make run-dev` 或手动执行 `go run ./cmd/organization-command-service/main.go` | 必须使用 RS256，`secrets/dev-jwt-*.pem` 已存在 |
| **查询服务 (GraphQL, :8090)** | 运行中、/health 返回 200 | 同上（`go run ./cmd/organization-query-service/main.go`） | `JWT_JWKS_URL` 指向命令服务 JWKS |
| **前端 (Vite Dev, :3000)** | 可访问，首页返回 HTML | `npm run dev`（默认 webpack dev server，已在 Playwright 中复用） | Playwright 取 `baseURL` 时需判定该服务是否已经就绪 |
| **RS256 开发令牌** | `.cache/dev.jwt` 存在且未过期 | `make jwt-dev-mint`（如需刷新） | `PW_JWT`、`PW_TENANT_ID` 从此文件导出 |
| **本地环境变量** | 已设置 | ```bash 
export PW_JWT=$(cat .cache/dev.jwt)  
export PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9
``` | Playwright 会自动回退读取 `.cache/dev.jwt`，但建议显式导出 |

**验证脚本**（全部通过后才允许进入测试环节）：
```bash
curl -fsS http://localhost:9090/health
curl -fsS http://localhost:8090/health
curl -fsS http://localhost:3000/ | head -n 1
node scripts/validate-field-naming-simple.js
```

---

## 2. 测试范围与执行顺序

1. **单元与集合测试（可并行）**
   - `npm run test -- queryClient`
   - `npm run test -- useEnterpriseOrganizations`
   - `npm run test -- OrganizationDashboard`

2. **前端静态校验**
   - `npm run validate:field-naming`
   - `node scripts/generate-implementation-inventory.js`

3. **E2E 冒烟测试（Chromium）**
   - 命令：`npm run test:e2e:smoke`
   - 覆盖规格：
     - `tests/e2e/simple-connection-test.spec.ts`
     - `tests/e2e/basic-functionality-test.spec.ts`
     - `tests/e2e/organization-create.spec.ts`
   - Playwright 会自动判定前端是否启用；若检测到“环境不可用”即视为失败，需先恢复服务再重试。

4. **（可选）全量 Playwright / Vitest 覆盖**
   - `npm run test:e2e`（全浏览器矩阵）
   - `npm run test`（Vitest 全量）
   - `npm run coverage`（如需覆盖率报告）

> **执行顺序不可调换**：只有在 1~2 步全部通过后才能进入 E2E；若冒烟失败需先修复环境，禁止忽略后续步骤直接归档。

---

## 3. 结果记录模板

| 日期 | 测试类型 | 命令 | 结果 | 主要产物/日志 | 备注 |
|------|----------|------|------|---------------|------|
| 2025-10-11 | Unit (queryClient) | `npm run test -- queryClient` | ✅ | `frontend/test-results/` | 4/4 通过，耗时 1.34s |
| 2025-10-11 | Unit (useEnterpriseOrganizations) | `npm run test -- useEnterpriseOrganizations` | ✅ | `frontend/test-results/` | 4/4 通过，耗时 950ms |
| 2025-10-11 | Unit (OrganizationDashboard) | `npm run test -- OrganizationDashboard` | ✅ | `frontend/test-results/` | 2/2 通过，耗时 1.01s |
| 2025-10-11 | Field Naming | `npm run validate:field-naming` | ✅ | `reports/implementation-inventory.json` | 144个文件，0违规项 |
| 2025-10-11 | Implementation Inventory | `node scripts/generate-implementation-inventory.js` | ✅ | `reports/implementation-inventory.json` | 26 REST + 9 GraphQL + 45 Go + 172 TS |
| 2025-10-11 | Vitest 覆盖率 | `npx vitest run --coverage --run` | ✅ | `frontend/coverage/` | 语句 84.1% / 分支 71.3% / 函数 75.9%。范围限定在 Phase3 相关模块 |
| 2025-10-11 | Bundle 分析 | `npm run build:analyze` | ✅ | `frontend/dist/` | Vite 构建通过，核心 bundle (vendor-state) gzip≈12.45 kB |
| 2025-10-11 | E2E 冒烟 | `npm run test:e2e:smoke` | ✅ | `frontend/playwright-report/` `frontend/test-results/` | 6 通过 / 1 跳过；开发代理出现 `.well-known/jwks.json` EPROTO 告警，已确认不影响用例 |
| YYYY-MM-DD | 全量 Playwright (可选) | `npm run test:e2e` | ✅/⚠️ | `frontend/playwright-report/` | |
| YYYY-MM-DD | 覆盖率 (可选) | `npm run coverage` | ✅/⚠️ | `coverage/` | |

- 所有失败项必须附带：命令输出、关键日志、截图/trace（Playwright 自动保存在 `frontend/playwright-report/` 与 `frontend/test-results/`）。
- 在 63 号计划或其他执行文档中引用测试结果时，应链接至本表格对应行，并明示产物路径。

---

## 4. 常见故障与处理

| 现象 | 原因 | 处理方式 |
|------|------|----------|
| Playwright 报 “测试环境不可用” | 前端/命令/查询服务尚未启动，或端口被占用 | 按第 1 节重启服务；确认 `validateTestEnvironment()` 输出路径 |
| `setupAuth` 抛出 “无法获取 RS256 开发令牌” | 命令服务未开启 `/auth/dev-token` | 重新执行 `make run-dev` 并确认日志 |
| Vitest 报错 `import.meta.env` 未定义 | 环境脚本在 Node 环境执行 | 确保测试中引用 `env` 模块前已 mock 或使用默认值 |
| `npm run test:e2e` 无响应 | 本地未安装 Playwright 浏览器 | 执行 `npx playwright install` |

若问题不在上述范围：
1. 收集日志、截图、trace（如 `playwright-report`）。
2. 在 `docs/development-plans/63-front-end-query-plan.md` 的风险章节登记。
3. 通过 Slack `#cube-castle-testing` 通知运行保障组，说明影响范围与复现步骤。

---

## 5. 归档要求

- 每轮测试结束后，必须将第 3 节表格更新到最新状态，并附上失败分析或验收结论。
- 阶段性验收需打包以下文件：
  1. `frontend/playwright-report/index.html` 及 `data/`、`trace/` 全量目录。
  2. `frontend/test-results/` 下的截图与 trace（如有失败）。
  3. `reports/implementation-inventory.json` 最新版本。
  4. 任何附加脚本或调试记录（存放在 `reports/` 或 `docs/archive/`，严禁散落）。
- 归档提交前请再次执行 `node scripts/generate-implementation-inventory.js`，确认前端导出清单与最新改动保持一致。

---

**维护人**：集成团队测试负责人（全栈工程师）  
**最后更新**：2025-10-11 08:50 CST  
**变更摘要**：清空旧的运维记录，重新定义标准化测试流程与验收要求，覆盖环境准备、执行顺序、结果登记与故障处理。若需调整流程，请先提 PR 修改本文件，并通知所有相关团队。

---

## 6. 待办事项（与 63 号计划同步）

以下项目需在 Phase 3 完成前持续跟踪（所有条目完成后再更新 63/06 文档）：

1. **提升 Vitest 覆盖率 ≥ 75%**  
   - 优先补齐 `shared/hooks/useOrganizationMutations`, `shared/utils/organization-helpers`, `features/organizations/*` 等高价值模块。
   - 每次补测后更新第 3 节结果表与 63 号计划验收状态。

2. **修复 `vite build` 阻塞的 TypeScript 校验**  
   - 重点文件：`AuditEntryCard/AuditHistorySection`, `OrganizationForm`, `Temporal*` 组件、`logger` 类型定义等。  
   - 完成后重新执行 `npm run build:analyze`，记录 bundle 体积及是否达到 ≥5% 优化目标。

3. **更新配置/QA 文档及验收草稿**  
   - 将环境变量、端口说明等同步至 `docs/reference/`；  
   - 归档 QA 冒烟流程、截图、报告；  
   - 起草 64 号验收文档，完成 Phase 3 关闭条件。
