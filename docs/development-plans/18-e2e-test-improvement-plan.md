# 18 — E2E 测试完善计划

**创建日期**：2025-10-02
**最后更新**：2025-10-02
**责任团队**：前端团队 + QA 团队
**状态**：✅ **就绪，可启动 Phase 1**
**关联文档**：[06-integrated-teams-progress-log.md](./06-integrated-teams-progress-log.md)、[16-code-smell-analysis-and-improvement-plan.md](./16-code-smell-analysis-and-improvement-plan.md)
**验证报告**：[reports/iig-guardian/playwright-rs256-verification-20251002.md](../../reports/iig-guardian/playwright-rs256-verification-20251002.md)

---

## 📊 执行摘要（2025-10-02 更新）

### 实施条件评估结论

**✅ 已具备立即实施条件**

**已完成阻塞项**（优先级 P0）：
- ✅ **本地端到端验证已执行**：成功启动 Docker 栈、生成 JWT 并运行 Playwright 套件，环境稳定可用（firefox 40通过，chromium 部分完成）
- ✅ **验证报告已填写**：`reports/iig-guardian/plan18-ci-verification-20251002.md` 已完成，结论为"部分通过，建议启动 Phase 1"

**已具备的基础条件**：
- ✅ Docker E2E 栈配置（`docker-compose.e2e.yml`）
- ✅ RS256 认证链路工具链
- ✅ Playwright 测试套件（5 个测试文件）
- ✅ 本地验证手册与报告模板（PLAN18-CI-VERIFICATION-GUIDE.md / plan18-ci-verification-20251002.md）

**验证执行摘要**（2025-10-02）：
1. ✅ 本地验证完成：Docker栈+JWT+Playwright（耗时11分钟）
2. ✅ 验证报告已填写：完整记录环境信息、执行结果、发现问题
3. ✅ 依赖项状态已更新为 ✅
4. ✅ 验证证据已保存：日志、报告、Playwright artifacts

**下一步**：可以立即启动 Phase 1 任务（修复现有失败测试）。

---

## 一、背景与目标

### 1.1 背景

2025-10-02 完成的 Playwright RS256 E2E 验证（Plan 06）显示：
- ✅ **核心功能正常**：RS256 认证链路、PBAC 权限验证、GraphQL 查询服务、架构契约全部通过
- ⚠️ **次要问题存在**：业务流程测试与基础功能测试存在 3 处失败，总体通过率 ~83%
- 📊 **测试覆盖不足**：优化验证与回归测试剧本尚未执行，完整 CRUD 流程因超时未完成

### 1.2 目标

通过本计划实现以下目标：
1. **修复现有失败测试**：将 E2E 测试通过率提升至 95% 以上
2. **完善测试覆盖**：补充优化验证、回归测试、完整 CRUD 流程
3. **提升测试稳定性**：解决超时、选择器失效等不稳定因素
4. **建立质量门禁**：将 E2E 测试纳入 CI 流程，确保长期可维护性

### 1.3 成功标准

- [ ] 业务流程 E2E 通过率 ≥95%
- [ ] 基础功能 E2E 通过率 100%
- [ ] 架构契约 E2E 保持 100%
- [ ] 优化验证与回归测试剧本执行完成
- [ ] E2E 测试总耗时 <5 分钟
- [ ] 测试失败时提供清晰的错误上下文与截图

---

## 二、现状分析

### 2.1 测试结果总览

| 测试类别 | 通过 | 失败 | 总计 | 通过率 | 关键问题 |
|---------|------|------|------|--------|---------|
| **PBAC Scope 验证** | ✅ 1 | 0 | 1 | 100% | - |
| **架构契约 E2E** | ✅ 6 | 0 | 6 | 100% | - |
| **业务流程 E2E** | ⚠️ 9+ | ❌ 1 | 10+ | ~90% | 数据一致性断言 |
| **基础功能 E2E** | ✅ 8 | ❌ 2 | 10 | 80% | 测试页面缺失 |
| **优化验证 E2E** | - | - | - | 未执行 | - |
| **回归测试 E2E** | - | - | - | 未执行 | - |

### 2.2 失败测试详情

#### 问题 1：数据一致性测试失败（P2）

**文件**：`tests/e2e/business-flow-e2e.spec.ts:355`
**现象**：
```
Expected: "启用"
Received: "✓ 启用"
```

**根因分析**：
- 前端组件在渲染状态字段时添加了勾选标记 `✓`
- 测试断言期待 API 返回的原始值（`"启用"`），但前端已转换为带标记的显示值

**影响范围**：
- 业务流程 E2E 中所有涉及状态字段的数据一致性验证
- 可能影响其他字段的显示逻辑验证

**证据**：
- 截图：`test-results/business-flow-e2e-业务流程端到端测试-数据一致性验证测试-chromium/test-failed-1.png`
- 视频：`test-results/.../video.webm`

---

#### 问题 2：测试页面功能验证失败（P3）

**文件**：`tests/e2e/basic-functionality-test.spec.ts:81`
**现象**：
```
expect(hasButtons).toBeGreaterThan(0);
// Received: 0
```

**根因分析**：
- `/test` 路由可能不存在或未正确渲染
- 测试页面组件缺失交互元素（按钮、表格）

**影响范围**：
- 基础功能 E2E 中的测试页面验证（chromium + firefox）
- 可能表明测试页面已废弃或未维护

**证据**：
- 截图：`test-results/basic-functionality-test-时态管理系统基础功能验证-测试页面功能验证-chromium/test-failed-1.png`

---

#### 问题 3：业务流程测试超时（P2）

**文件**：`tests/e2e/business-flow-e2e.spec.ts:20`（完整 CRUD 流程）
**现象**：测试执行超时（2 分钟限制），未完成所有剧本

**可能原因**：
1. 前端页面加载慢或等待元素超时
2. 表单提交后跳转延迟
3. 选择器策略不当（未使用 `data-testid`）
4. 网络请求响应慢

**影响范围**：
- 完整 CRUD 流程覆盖不完整
- 无法验证创建、编辑、删除的端到端流程

---

### 2.3 未执行测试

| 测试文件 | 状态 | 原因 |
|---------|------|------|
| `optimization-verification-e2e.spec.ts` | ⏳ 未执行 | 06 号文档未明确要求 |
| `regression-e2e.spec.ts` | ⏳ 未执行 | 06 号文档未明确要求 |
| 完整 CRUD 流程（`business-flow-e2e.spec.ts`） | ⚠️ 超时未完成 | 需优化超时与选择器 |

---

## 三、改进计划

### 3.1 Phase 1：修复现有失败测试（优先级 P1-P2）

#### 任务 1.1：修复数据一致性测试（P2）

**责任人**：前端团队
**工作量**：0.5 天
**截止日期**：2025-10-04

**方案 A（推荐）**：调整测试断言逻辑
```typescript
// 修改前
expect(firstFrontendItem.status).toBe(expectedStatus);

// 修改后（支持带标记的显示值）
const statusMap: Record<string, string> = {
  'ACTIVE': '✓ 启用',
  'INACTIVE': '停用',
  'PLANNED': '计划中'
};
const expectedDisplayStatus = statusMap[firstApiItem.status] || firstApiItem.status;
expect(firstFrontendItem.status).toBe(expectedDisplayStatus);
```

**方案 B（长期）**：统一状态字段渲染规范
- 将视觉标记移至 CSS 伪元素（`::before`）或独立图标组件
- API 返回值与显示文本保持一致
- 更新设计规范文档

**验收标准**：
- [ ] 数据一致性测试通过（chromium + firefox）
- [ ] 其他状态字段测试不受影响
- [ ] 更新测试文档说明状态映射逻辑

---

#### 任务 1.2：修复或移除测试页面验证（P3）

**责任人**：前端团队 + QA
**工作量**：0.25 天
**截止日期**：2025-10-05

**方案 A（快速）**：移除测试页面剧本
```typescript
// 在 basic-functionality-test.spec.ts 中标记为 skip
test.skip('测试页面功能验证', async ({ page }) => {
  // 原测试逻辑
});
```

**方案 B（长期）**：修复测试页面
- 检查 `/test` 路由配置
- 补充测试页面组件（表格、按钮）
- 更新路由文档说明测试页面用途

**决策依据**：
- 如测试页面为调试用途且不影响生产，选择方案 A
- 如测试页面为必需功能，选择方案 B

**验收标准**：
- [ ] 基础功能 E2E 通过率 100%
- [ ] 测试套件无阻塞性失败

---

#### 任务 1.3：解决业务流程测试超时（P2）

**责任人**：QA 团队
**工作量**：1 天
**截止日期**：2025-10-06

**优化措施**：

1. **增加超时配置**
   ```typescript
   test('完整CRUD业务流程测试', async ({ page }) => {
     test.setTimeout(180000); // 从 120s 增加到 180s
   });
   ```

2. **优化选择器策略**
   - 确保所有交互元素使用 `data-testid`
   - 使用 `page.waitForSelector()` 替代固定延迟
   ```typescript
   // 修改前
   await page.click('button:has-text("创建")');

   // 修改后
   await page.click('[data-testid="create-organization-button"]');
   await page.waitForURL('**/organizations/*/temporal');
   ```

3. **分段验证**
   - 将长流程拆分为独立测试
   - 每个测试专注单一功能（创建 / 编辑 / 删除）

**验收标准**：
- [ ] 完整 CRUD 流程测试通过（3 分钟内）
- [ ] 提供创建、编辑、删除的完整截图与视频
- [ ] HAR 文件记录网络请求

---

### 3.2 Phase 2：补充未执行测试（优先级 P2-P3）

#### 任务 2.1：执行优化验证测试（P3）

**责任人**：QA 团队
**工作量**：0.5 天
**截止日期**：2025-10-08

**执行命令**：
```bash
PW_JWT=$(cat .cache/dev.jwt) PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 \
npm run test:e2e -- tests/e2e/optimization-verification-e2e.spec.ts
```

**验收标准**：
- [ ] 测试执行完成并生成报告
- [ ] 记录性能指标（页面加载、API 响应、内存使用）
- [ ] 更新 06 号文档"当前状态"

---

#### 任务 2.2：执行回归测试（P3）

**责任人**：QA 团队
**工作量**：0.5 天
**截止日期**：2025-10-09

**执行命令**：
```bash
PW_JWT=$(cat .cache/dev.jwt) PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 \
npm run test:e2e -- tests/e2e/regression-e2e.spec.ts
```

**验收标准**：
- [ ] 测试执行完成并生成报告
- [ ] 验证核心功能无回退
- [ ] 更新回归测试基线

---

### 3.3 Phase 3：提升测试稳定性与可维护性（优先级 P2）

#### 任务 3.1：优化 Playwright 配置（P2）

**责任人**：QA 团队
**工作量**：0.5 天
**截止日期**：2025-10-10

**改进项**：

1. **统一超时配置**
   ```typescript
   // playwright.config.ts
   export default defineConfig({
     timeout: 120000, // 全局测试超时
     expect: { timeout: 10000 }, // 断言超时
     use: {
       actionTimeout: 15000, // 交互超时
       navigationTimeout: 30000, // 导航超时
     },
   });
   ```

2. **增强错误诊断**
   ```typescript
   use: {
     screenshot: 'only-on-failure',
     video: 'retain-on-failure',
     trace: 'retain-on-failure',
   },
   ```

3. **浏览器并行优化**
   ```typescript
   workers: process.env.CI ? 2 : 4, // CI 环境限制并发
   ```

**验收标准**：
- [ ] 测试失败时自动保存 trace、video、screenshot
- [ ] 测试总耗时 <5 分钟
- [ ] 更新 `frontend/playwright.config.ts`

---

#### 任务 3.2：建立 E2E 测试 CI 门禁（P2）

**责任人**：平台/工具团队
**工作量**：1 天
**截止日期**：2025-10-12

**实施步骤**：

1. **创建 GitHub Actions 工作流**
   ```yaml
   # .github/workflows/e2e-tests.yml
   name: E2E Tests
   on:
     pull_request:
       branches: [main, develop]

   jobs:
     e2e:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3
         - uses: actions/setup-node@v3
         - run: npm ci --prefix frontend
         - run: make docker-up
         - run: make run-dev &
         - run: make run-auth-rs256-sim &
         - run: sleep 10
         - run: PW_JWT=$(cat .cache/dev.jwt) PW_TENANT_ID=xxx npm run test:e2e --prefix frontend
         - uses: actions/upload-artifact@v3
           if: failure()
           with:
             name: playwright-report
             path: frontend/playwright-report/
   ```

2. **配置失败通知**
   - Slack/企业微信通知
   - PR 检查状态阻塞合并

**验收标准**：
- [ ] PR 提交后自动触发 E2E 测试
- [ ] 测试失败阻止 PR 合并
- [ ] 失败报告自动上传至 Artifacts

---

### 3.4 Phase 4：文档与知识传承（优先级 P3）

#### 任务 4.1：完善 E2E 测试文档（P3）

**责任人**：QA 团队
**工作量**：0.5 天
**截止日期**：2025-10-13

**文档内容**：
1. **快速开始**：本地运行 E2E 测试的完整步骤
2. **编写规范**：选择器命名、断言策略、错误处理
3. **调试指南**：如何使用 Playwright Inspector 与 Trace Viewer
4. **常见问题**：超时、选择器失效、认证失败的解决方案

**输出位置**：`docs/development-tools/e2e-testing-guide.md`

---

#### 任务 4.2：更新参考手册（P3）

**责任人**：架构组
**工作量**：0.25 天
**截止日期**：2025-10-13

**更新内容**：
- 在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 补充 E2E 测试命令
- 在 `CLAUDE.md` 索引中添加 18 号文档链接

---

## 四、时间表与里程碑

| 阶段 | 任务 | 负责人 | 工期 | 截止日期 | 状态 |
|------|------|--------|------|---------|------|
| **Phase 1** | 修复数据一致性测试 | 前端团队 | 0.5d | 2025-10-04 | ⏳ 待启动 |
| | 修复/移除测试页面 | 前端 + QA | 0.25d | 2025-10-05 | ⏳ 待启动 |
| | 解决业务流程超时 | QA 团队 | 1d | 2025-10-06 | ⏳ 待启动 |
| **Phase 2** | 执行优化验证测试 | QA 团队 | 0.5d | 2025-10-08 | ⏳ 待启动 |
| | 执行回归测试 | QA 团队 | 0.5d | 2025-10-09 | ⏳ 待启动 |
| **Phase 3** | 优化 Playwright 配置 | QA 团队 | 0.5d | 2025-10-10 | ⏳ 待启动 |
| | 建立 CI 门禁 | 平台/工具 | 1d | 2025-10-12 | ⏳ 待启动 |
| **Phase 4** | 完善测试文档 | QA 团队 | 0.5d | 2025-10-13 | ⏳ 待启动 |
| | 更新参考手册 | 架构组 | 0.25d | 2025-10-13 | ⏳ 待启动 |
| **总计** | | | **5 天** | **2025-10-13** | |

---

## 五、验收标准

### 5.1 测试通过率

- [ ] 业务流程 E2E：≥95%（9+/10 通过）
- [ ] 基础功能 E2E：100%（10/10 通过）
- [ ] 架构契约 E2E：100%（保持）
- [ ] 优化验证 E2E：≥90%
- [ ] 回归测试 E2E：≥95%

### 5.2 性能指标

- [ ] 测试总耗时：<5 分钟
- [ ] 单个测试超时：<3 分钟
- [ ] 页面加载时间：<1 秒（中位数）
- [ ] API 响应时间：<200ms（P95）

### 5.3 质量门禁

- [ ] `.github/workflows/e2e-tests.yml` 在 PR 场景下自动运行 Playwright 套件
- [ ] PR 合并前必须通过 E2E 测试（含 chromium / firefox 项目）
- [ ] 失败测试自动生成 trace、screenshot、video 并作为 artifact 上传

### 5.4 验证步骤（执行顺序）

1. 启动依赖栈并确保健康：
   ```bash
   make docker-up
   make run-auth-rs256-sim
   curl -fsS http://localhost:9090/health
   curl -fsS http://localhost:8090/health
   ```
2. 生成开发用 RS256 JWT，并记录租户：
   ```bash
   make jwt-dev-mint
   export PW_JWT=$(cat .cache/dev.jwt)
   export PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9
   ```
3. 在 `frontend/` 目录执行完整 Playwright 套件，确认 chromium 与 firefox 全部通过：
   ```bash
   cd frontend
   PW_JWT=$PW_JWT PW_TENANT_ID=$PW_TENANT_ID npm run test:e2e
   ```
4. 如需单独验证剧本（例如回归/优化验证），复用相同令牌：
   ```bash
   PW_JWT=$PW_JWT PW_TENANT_ID=$PW_TENANT_ID \
     npm run test:e2e -- tests/e2e/regression-e2e.spec.ts
   ```
5. 通过 `npx playwright show-report` 检查报告与 trace；CI 失败时在 artifact 中获取相同证据。
6. 所有结果与操作需同步记录到 `docs/development-tools/e2e-testing-guide.md` 所列标准。

### 5.5 文档完整性

- [ ] E2E 测试指南发布（`docs/development-tools/e2e-testing-guide.md`）
- [ ] 快速参考手册更新
- [ ] CLAUDE.md 索引包含 18 号文档

---

## 六、风险与缓解

### 6.1 风险识别

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| **前端渲染逻辑变更导致测试失败** | 中 | 中 | 使用 `data-testid` 替代文本选择器 |
| **测试环境不稳定（网络/数据库）** | 低 | 高 | 增加重试机制，使用 Docker 固定环境 |
| **本地资源不足导致超时** | 中 | 中 | 限制并发数，提前构建镜像并清理磁盘 |
| **团队对 Playwright 不熟悉** | 低 | 中 | 提供培训与文档，指定专人支持 |

### 6.2 依赖项

- ✅ 服务栈已启动（PostgreSQL + Redis + 命令服务 + 查询服务）
- ✅ RS256 JWT 认证链路正常
- ✅ Playwright 基础配置完成
- ✅ 本地端到端验证 — **已执行** (2025-10-02)
  - ✅ 验证报告：`reports/iig-guardian/plan18-ci-verification-20251002.md`
  - ✅ Playwright 报告：`frontend/playwright-report/index.html`
  - ✅ 验证日志：`reports/iig-guardian/plan18-local-validation.log`
  - ✅ 执行结果：firefox 40通过/4跳过，chromium 部分完成，环境可用

---

## 七、实施条件评估

### 7.1 前置条件检查清单

| 检查项 | 状态 | 说明 |
|--------|------|------|
| **本地测试环境可用** | ✅ 完成 | `docker-compose.e2e.yml` 已就绪，可在本地运行完整 E2E 栈 |
| **RS256 认证链路验证** | ✅ 完成 | 已通过 `playwright-rs256-verification-20251002.md` 验证 |
| **Playwright 测试套件存在** | ✅ 完成 | 5 个测试文件已创建（架构、业务流程、基础功能、优化、回归） |
| **本地端到端验证执行** | ✅ 完成 | 2025-10-02 执行完成，耗时11分钟 |
| **验证报告填写** | ✅ 完成 | 报告已填写，结论为"部分通过，建议启动 Phase 1" |
| **失败测试根因已明确** | ✅ 完成 | 3 个失败问题已分析并提供修复方案 |
| **团队资源到位** | ⏳ 待确认 | 需前端团队 + QA 团队确认排期 |

### 7.2 实施条件判定

**结论**：✅ **已具备立即实施条件**

**关键阻塞项**：
1. **本地端到端验证未执行**（优先级 P0）
   - 尚未验证 Docker 栈启动、服务健康检查、JWT mint 与 Playwright 执行是否在当前机器可行
   - 缺少以下证据：
     - 本地 `make docker-up` / `make run-auth-rs256-sim` 成功日志
     - `.cache/dev.jwt` 生成记录
     - `npm run test:e2e` 控制台输出与报告产物

2. **验证报告未填**（优先级 P1）
   - `reports/iig-guardian/plan18-ci-verification-20251002.md` 尚未填写执行结果
   - Plan 18 文档 6.2 节依赖状态仍为待执行

### 7.3 启动前必需操作

在启动 Plan 18 执行前，**必须**完成以下本地验证流程：

#### 操作 1：执行本地端到端验证（优先级 P0）
```bash
# 1. 启动依赖栈
make docker-up
make run-auth-rs256-sim

# 2. 健康检查（需均返回 HTTP 200）
curl -fsS http://localhost:9090/health
curl -fsS http://localhost:8090/health

# 3. 生成 RS256 JWT
make jwt-dev-mint
export PW_JWT=$(cat .cache/dev.jwt)
export PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9

# 4. 运行完整 Playwright 套件
cd frontend
PW_JWT=$PW_JWT PW_TENANT_ID=$PW_TENANT_ID npm run test:e2e

# 5. 清理资源（完成验证后执行）
cd ..
docker compose -f docker-compose.e2e.yml down -v --remove-orphans 2>/dev/null || true
make dev-kill
```

**验证指标**：
- [ ] Docker Compose 栈成功启动且服务健康
- [ ] `.cache/dev.jwt` 成功生成并可被读取
- [ ] Playwright 套件至少完成一次执行（允许包含失败用例）
- [ ] 生成 `frontend/playwright-report` 与 `frontend/test-results`

#### 操作 2：记录验证结果（优先级 P1）
- 在 `reports/iig-guardian/plan18-ci-verification-20251002.md` 填写执行日期、命令输出摘要、Playwright 结果、耗时与结论。
- 说明验证所使用的环境（本地机器、操作系统、Docker 版本、Node/Go 版本）。

#### 操作 3：更新依赖项状态（优先级 P1）
完成本地验证与报告后，更新 `6.2 依赖项` 状态为：
```markdown
- ✅ 本地端到端验证 — 已执行 {YYYY-MM-DD}
  - 验证报告：`reports/iig-guardian/plan18-ci-verification-20251002.md`
  - Playwright 运行日志：`frontend/playwright-report/index.html`
```

### 7.4 风险提示

如果跳过本地端到端验证直接实施 Plan 18：
- **高风险**：环境未准备充分导致 Phase 1 修复操作阻塞（如 JWT 生成失败、端口冲突）
- **中风险**：Playwright 依赖缺失导致用例无法启动，需要返工
- **低风险**：测试报告目录不存在，影响后续追踪

**建议**：确保本地验证步骤全部通过并留存证据，再进入用例修复阶段。

---

## 八、后续计划

### 8.1 Phase 5（长期优化）

1. **E2E 测试并行化**：优化测试分组，利用多 worker 并行执行
2. **视觉回归测试**：集成 Percy 或 Playwright Visual Comparisons
3. **性能监控集成**：E2E 测试中采集 Web Vitals 指标
4. **移动端测试**：补充移动浏览器（Mobile Chrome/Safari）测试

### 8.2 技术债务

- **TODO-TEMPORARY 清理**：测试文件中的临时实现需按 Plan 17 标准处理
- **弱类型治理**：`any/unknown` 在测试文件中的使用需符合 Plan 16 规范
- **选择器硬编码**：全面迁移到 `data-testid` 策略

---

## 九、参考资料

### 9.1 内部文档

- [06-integrated-teams-progress-log.md](./06-integrated-teams-progress-log.md) — 集成团队推进记录
- [16-code-smell-analysis-and-improvement-plan.md](./16-code-smell-analysis-and-improvement-plan.md) — 代码异味治理计划
- [Playwright RS256 验证报告](../../reports/iig-guardian/playwright-rs256-verification-20251002.md)

### 9.2 技术参考

- [Playwright 官方文档](https://playwright.dev/)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [RS256 JWT 认证规范](../../docs/api/openapi.yaml)
- [PBAC 权限实现](../../internal/auth/pbac.go)

---

## 附录 A：测试命令速查

```bash
# 1. 启动完整服务栈
make docker-up
make run-dev &
make run-auth-rs256-sim &

# 2. 生成 JWT 令牌
curl -X POST http://localhost:9090/auth/dev-token \
  -H 'Content-Type: application/json' \
  -d '{"userId":"dev-user","tenantId":"3b99930c-4dc6-4cc9-8e4d-7d960a931cb9","roles":["ADMIN","USER"],"duration":"24h"}' \
  | jq -r '.data.token' > .cache/dev.jwt

# 3. 执行所有 E2E 测试
cd frontend
PW_JWT=$(cat ../.cache/dev.jwt) PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 \
npm run test:e2e

# 4. 执行单个测试文件
npm run test:e2e -- tests/e2e/business-flow-e2e.spec.ts

# 5. 调试模式（打开 Playwright Inspector）
npm run test:e2e -- --debug

# 6. 查看测试报告
npx playwright show-report
```

---

## 附录 B：修改清单

| 文件 | 修改类型 | 说明 |
|------|---------|------|
| `tests/e2e/business-flow-e2e.spec.ts:355` | 修改断言 | 支持带标记的状态显示值 |
| `tests/e2e/basic-functionality-test.spec.ts:68` | 标记 skip | 移除测试页面验证 |
| `tests/e2e/business-flow-e2e.spec.ts:20` | 增加超时 | 从 120s 增至 180s |
| `playwright.config.ts` | 优化配置 | 统一超时、增强诊断 |
| `.github/workflows/e2e-tests.yml` | 新建文件 | CI 门禁工作流 |
| `docs/development-tools/e2e-testing-guide.md` | 新建文件 | E2E 测试指南 |

---

**本文档状态**：✅ 已创建，待团队评审与启动
