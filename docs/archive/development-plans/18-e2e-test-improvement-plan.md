# 18 — E2E 测试完善计划

**创建日期**: 2025-10-02
**最后更新**: 2025-10-05 (Phase 1.3 验收完成)
**责任团队**: 前端团队 + QA 团队
**状态**: ✅ **Phase 1.3 完成（已归档）**
**关联文档**: [06-integrated-teams-progress-log.md](../../development-plans/06-integrated-teams-progress-log.md)

---

## 📊 当前状态 (2025-10-05)

### 测试通过率：22/22 (100%)

| 测试类别 | 状态 | 通过率 | 说明 |
|---------|------|--------|------|
| PBAC Scope 验证 | ✅ | 100% | 访问令牌 RS256，Tenant 校验通过 |
| 架构契约 E2E | ✅ | 100% | `tests/e2e/architecture-e2e.spec.ts` (Chromium) 全绿，GraphQL 200 |
| 优化验证 E2E | ✅ | 100% | 指标与性能剧本全部通过 |
| 回归测试 E2E | ✅ | 100% | 8/8 绿灯，网络恢复脚本稳定 |
| 基础功能 E2E | ✅ | 100% | 4/4 绿灯 |
| 业务流程 E2E | ✅ | 100% | `tests/e2e/business-flow-e2e.spec.ts` (Chromium + Firefox) 各 5/5 |

### 🛠️ 最终修复摘要（2025-10-05）

- ✅ **数据库兼容层闭环**：新增 `database/migrations/032_phase_b_remove_legacy_columns.sql`，在重建 `organization_temporal_current` 视图后移除 `is_deleted`、`operation_reason` 列；`make db-migrate-all` 全量执行通过。
- ✅ **触发器/时态补强**：`030_fix_is_current_with_utc_alignment.sql`、`031_cleanup_temporal_triggers.sql`、`032_phase_b_remove_legacy_columns.sql` 组合完成 Phase B，彻底移除旧列依赖，审计/版本触发器幂等。
- ✅ **前端认证工具链**：`frontend/tests/e2e/utils/authToken.ts` + `auth-setup.ts` 支持 RS256 令牌自动续签、过期检测与统一头注入；`playwright.config.ts` 自动读取 `.cache/dev.jwt`。
- ✅ **E2E 稳定性**：`business-flow-e2e.spec.ts` 修复错误恢复步骤，Firefox 场景新增对“重新加载”按钮的兜底；`OrganizationTree` 提供 `data-testid="organization-tree-retry-button"` 以支撑跨浏览器定位。
- ✅ **脚本验收**：Chromium/Firefox 分别执行 `npm --prefix frontend run test:e2e -- --project=<browser> tests/e2e/business-flow-e2e.spec.ts`，报告均 5/5 绿灯；`architecture-e2e.spec.ts` 再跑确认 GraphQL 200。

### 📈 Phase 1.3 验收结果 (2025-10-05 19:40 UTC)

**测试人员**: QA Automation Team + 前端工具组
**浏览器矩阵**: Chromium 118、Firefox 118
**命令**:
```bash
npm --prefix frontend run test:e2e -- --project=chromium tests/e2e/business-flow-e2e.spec.ts
npm --prefix frontend run test:e2e -- --project=firefox tests/e2e/business-flow-e2e.spec.ts
npm --prefix frontend run test:e2e -- --project=chromium tests/e2e/architecture-e2e.spec.ts
```

#### ✅ 主要结论
- 业务流程端到端剧本在 Chromium、Firefox 均 5/5 通过，错误恢复流程验证通过。
- 架构契约剧本返回 GraphQL 200，确认 Authorization / X-Tenant-ID 头配置生效。
- `make db-migrate-all` 在新增 032 迁移后无阻塞，可重复执行。
- Playwright 自动续签 `PW_JWT`，测试日志显示 `✅ 认证设置已注入 localStorage`。

#### 📦 佐证材料
- reports/iig-guardian/plan18-phase1.3-validation-20251005.md (已更新)
- reports/iig-guardian/plan18-business-flow-20251005T1930.log
- frontend/test-results/business-flow-e2e-* (Chromium/Firefox trace、video、截图)
- reports/iig-guardian/plan18-migration-20251005T1930.log

#### 🟢 阻塞项清单
- 无。Firefox 错误恢复剧本已通过。

### 📜 历史记录（保留原始记录）

### 🔍 Phase 1.3 手动测试结果 (2025-10-03 21:28-21:33)

**测试人员**: Claude Code (自动化)
**测试覆盖率**: 40% (部分场景完成)

#### ✅ 成功项
- ✅ 环境健康检查: 所有服务正常
- ✅ 创建组织: 成功创建 1000023
- ✅ 更新组织: 成功修改名称
- ✅ 删除组织: 成功删除记录
- ✅ 分页功能: 正常切换20/50条

#### ❌ 发现的问题

**🔴 P0 严重问题**:
1. **重新启用功能500错误**
   - 位置: `/organizations/{code}/temporal` 重新启用按钮
   - 错误: 服务器内部错误
   - 影响: 无法恢复组织为启用状态，阻塞完整状态流转

2. **搜索筛选完全失效**
   - 位置: `/organizations` 列表页名称搜索
   - 现象: UI显示"已激活筛选条件"但列表未筛选
   - 影响: 用户无法按名称查找组织

**🟡 P1 中等问题**:
3. **数据库is_current初始化错误**
   - 问题: 1000000的is_current初始为false
   - 影响: 创建子组织时报"父组织不存在或不可用"
   - 临时修复: 手动UPDATE is_current=true

4. **修改操作副作用**
   - 现象: 修改组织名称后状态从"启用"变为"停用"
   - 影响: 意外的状态变更，用户体验差

**🔵 P2 低优先级**:
5. Canvas Kit图标类型警告 (控制台大量错误)
6. make jwt-dev-mint命令Python依赖失败

详细报告: `test-results/manual/plan18-phase1.3/e2e-test-report.md`

### 🔁 Phase 1.3 自动化复测 (2025-10-05 11:32)

- ✅ 执行脚本：`scripts/plan18/run-business-flow-e2e.sh`（自动完成 Docker → 全量迁移 → `/auth/dev-token` → Playwright）。
- ✅ 产物：
  - `reports/iig-guardian/plan18-migration-20251005T113248.log` — 迁移 008–031 全量通过。
  - `reports/iig-guardian/plan18-business-flow-20251005T113248.log` — E2E 输出（Chromium/Firefox 各 5 场景）。
  - `reports/iig-guardian/plan18-phase1.3-validation-20251005.md` — 验证报告（9/10 通过）。
- ⚠️ 未通过剧本：`业务流程端到端测试 › 错误处理和恢复测试（Firefox）`
  - 定位：`getByRole('button', { name: '重试' })` 15s 超时未出现；API 调用均成功，页面未进入错误态。
  - 判定：脚本模拟错误流程与实际 UI 状态不一致（按钮未渲染或定位符需更新），与命令服务本次修复无直接关联。
- ✅ 验证要点：Chromium/Firefox “完整 CRUD 流程” 均通过；命令服务再无 `CREATE_ERROR`，请求 `requestId=6dcc6e79-3e51-471a-ac6c-b3d501e22a6b` 返回 200。

### ❌ 剩余阻塞 (P0)

**问题**: `business-flow-e2e › 错误处理和恢复测试（Firefox）` 未出现“重试”按钮
**文件**: `tests/e2e/business-flow-e2e.spec.ts`

**最新现象 (2025-10-05)**:
- 请求 `/api/v1/organization-units` 已成功返回 201；页面未进入错误态，`getByRole('button', { name: '重试' })` 超时。
- Trace 未捕获额外错误，推测 UI 逻辑已改为自动恢复或按钮选择器失效。

**当前要求**:
- 与前端确认错误恢复流程是否仍展示“重试”按钮；若逻辑改为自动恢复，应同步调整测试脚本和文档说明。
- 若按钮仍应存在，请为该元素添加稳定 `data-testid`，并更新测试定位逻辑后复测。
- 保留最新日志与 trace（`plan18-business-flow-20251005T113248.log`、`trace.zip`）作为修复参考。

---

## ✅ 已完成修复 (Phase 1.1-1.2)

### Phase 1.1 (2025-10-02)
- ✅ 页面加载时机优化 (三阶段等待逻辑)
- ✅ GraphQL 认证修复 (代理配置)
- ✅ ESLint 配置调整 (测试文件 no-console → warn)

### Phase 1.2 (2025-10-16)
- ✅ Vite import 别名修复 (`@/shared/*` 统一)
- ✅ 认证懒加载补偿 (`auth.ts` localStorage 回读)
- ✅ 优化验证断言对齐 (资源体积 < 4MB,Prometheus `/metrics`)
- ✅ 回归测试稳定化 (网络中断场景 `page.reload()` 异常捕获)

---

## 🎯 Phase 1.3: 修复创建表单渲染 (待启动)

### 目标
修复 `useTemporalMasterDetail` 创建模式初始状态,确保表单即时可见

### 执行步骤

0. **后端归一化逻辑生效** (平台团队, 即时)
   - 重启 `organization-command-service`，确保最新 `parentCode=0000000` 归一化已加载：
     ```bash
     make run-dev # 或单独重启命令服务进程
     ```
   - 冒烟验证：
     ```bash
     curl -sS http://localhost:9090/health
     curl -sS -X POST http://localhost:9090/api/v1/organization-units \
       -H "Authorization: Bearer $PW_JWT" \
       -H "X-Tenant-ID: $PW_TENANT_ID" \
       -H "Content-Type: application/json" \
     -d '{"name":"Smoke 根组织","unitType":"DEPARTMENT","parentCode":"0000000","effectiveDate":"2025-10-03"}'
     ```
     期望返回 201/200。
   - 如执行 `make db-migrate-all` 时出现 `there is no unique constraint matching given keys for referenced table "organization_units"`，需按以下步骤排查：
     1. 使用 `psql` 确认 `organization_units` 是否已存在预期的主键/唯一索引 (`record_id` PK、`tenant_id,code` 部分唯一)。
     2. 若缺失，先按照 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 的数据库初始化流程重建基础结构，再重跑迁移。
     3. 迁移成功后记录 `psql` 输出并更新本计划文档的“修复进展”小节，便于归档。

1. **前端创建模式初始化修复** (前端团队, ✅ 2025-10-03 完成)
   - `useTemporalMasterDetail` 已改为 `const [isLoading] = useState(Boolean(organizationCode));`
   - 创建模式下追加 `useEffect` 重置 `formMode`/`formInitialData`。

2. **完整回归** (QA 团队,0.5 天)
   ```bash
   chmod +x scripts/plan18/run-business-flow-e2e.sh
   scripts/plan18/run-business-flow-e2e.sh
   ```
   - 脚本将自动执行 `make docker-up`、`make db-migrate-all`、调用 `/auth/dev-token` 生成 RS256 令牌，并运行 `tests/e2e/business-flow-e2e.spec.ts`（日志默认输出至 `reports/iig-guardian/`）。
   - 如需仅运行 Playwright，可跳过脚本并按照 `PW_JWT=$(cat .cache/dev.jwt)`、`npm --prefix frontend run test:e2e -- tests/e2e/business-flow-e2e.spec.ts` 手动执行。
   - 当前阻塞：Firefox “错误处理与恢复” 场景无法复现“重试”按钮，脚本 15s 超时；需更新剧本或补充 UI 错误提示逻辑。
   - 测试脚本在“创建”步骤必须明确选择已有组织 `1000000` 作为上级，禁止留空；Playwright 默认流程已补充对 `ParentOrganizationSelector` 的操作。

3. **归档与记录** (QA 团队,0.25 天)
   - 创建 `reports/iig-guardian/plan18-phase1.3-validation-<date>.md`
   - 更新 `06-integrated-teams-progress-log.md` 当前状态

### 验收标准
- [x] 业务流程 E2E 通过率 ≥ 95% (5/5)
- [x] 创建请求返回 201/200，且请求体 `parentCode` 为现有上级 `1000000`
- [x] 创建/编辑/删除完整流程截图与视频
- [x] 测试报告归档至 `reports/iig-guardian/`

---

## 📋 Phase 2-3: 长期优化（已移交）

> 注：以下任务已纳入 QA 自动化路线图，将在新的计划文档中跟进；在本计划归档时保持原样记录。

### Phase 2: 质量门禁
- [ ] 建立 `.github/workflows/e2e-tests.yml`
- [ ] PR 合并前自动运行 E2E 测试
- [ ] 失败时自动上传 trace/screenshot/video

### Phase 3: 稳定性提升
- [ ] 优化 Playwright 配置 (超时/并发/重试)
- [ ] 测试总耗时优化至 < 5 分钟
- [ ] 补充 E2E 测试文档 (`docs/development-tools/e2e-testing-guide.md`)

---

## 🚀 快速执行指南

### 本地验证环境
```bash
# 1. 启动服务栈
make docker-up
export JWT_PRIVATE_KEY_PATH=/home/shangmeilin/cube-castle/secrets/dev-jwt-private.pem
export JWT_PUBLIC_KEY_PATH=/home/shangmeilin/cube-castle/secrets/dev-jwt-public.pem
go run ./cmd/organization-command-service/main.go &
go run ./cmd/organization-query-service/main.go &

# 2. 生成 JWT
make jwt-dev-mint
export PW_JWT=$(cat .cache/dev.jwt)
export PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9

# 3. 执行 E2E 测试
cd frontend
npm run test:e2e -- --project=chromium

# 4. 查看报告
npx playwright show-report
```

### 单个测试执行
```bash
# 仅执行业务流程测试
npm run test:e2e -- tests/e2e/business-flow-e2e.spec.ts

# 调试模式
npm run test:e2e -- --debug tests/e2e/business-flow-e2e.spec.ts
```

---

## 📊 归档条件评估

### 必须完成 (阻塞归档)
- [ ] **Phase 1.3**: 修复创建表单渲染,业务流程 E2E ≥ 95%
- [ ] 验证报告归档至 `reports/iig-guardian/`

### 建议完成 (长期价值)
- [ ] **Phase 2**: CI E2E 门禁建立
- [ ] **Phase 3**: Playwright 配置优化
- [ ] E2E 测试文档完善

### 预计归档日期
- **最早**: 2025-10-05 (仅 Phase 1.3 完成)
- **推荐**: 2025-10-12 (含 Phase 2-3)

---

## 📚 参考资料

### 内部文档
- [06-integrated-teams-progress-log.md](../../development-plans/06-integrated-teams-progress-log.md)
- [16-code-smell-analysis-and-improvement-plan.md](./16-code-smell-analysis-and-improvement-plan.md)
- [Playwright RS256 验证报告](../../reports/iig-guardian/playwright-rs256-verification-20251002.md)
- [E2E 测试指南](../../docs/development-tools/e2e-testing-guide.md)

### 技术参考
- [Playwright 官方文档](https://playwright.dev/)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)

---

**本文档状态**: ✅ 已精简,聚焦核心待办事项
**下一步行动**: 重启命令服务确认根编码归一化 → 复跑业务流程 E2E (Phase 1.3)
