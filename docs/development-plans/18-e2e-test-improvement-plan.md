# 18 — E2E 测试完善计划

**创建日期**: 2025-10-02
**最后更新**: 2025-10-03
**责任团队**: 前端团队 + QA 团队
**状态**: 🚧 **Phase 1.3 待启动（剩余 1 项核心阻塞）**
**关联文档**: [06-integrated-teams-progress-log.md](./06-integrated-teams-progress-log.md)

---

## 📊 当前状态 (2025-10-03)

### 测试通过率: 21/22 (95.5%)

| 测试类别 | 状态 | 通过率 | 说明 |
|---------|------|--------|------|
| PBAC Scope 验证 | ✅ | 100% | - |
| 架构契约 E2E | ✅ | 100% | 6/6 通过 |
| 优化验证 E2E | ✅ | 100% | 6/6 通过,Prometheus 指标已集成 |
| 回归测试 E2E | ✅ | 100% | 8/8 通过,网络中断剧本稳定 |
| 基础功能 E2E | ✅ | 100% | 4/4 通过 |
| **业务流程 E2E** | ⚠️ | **80%** | **4/5 通过,创建流程阻塞** |

### ❌ 剩余阻塞 (P0)

**问题**: `business-flow-e2e › 完整CRUD业务流程测试` 失败
**文件**: `tests/e2e/business-flow-e2e.spec.ts:38`

**现象**:
```
Timed out 10000ms waiting for expect(locator).toBeVisible()
Locator: getByTestId('organization-form')
```

**根因**:
- `/organizations/new` 路由进入创建模式,但 `useTemporalMasterDetail` hook 在 `isCreateMode=true` 时仍以 `isLoading=true` 初始化
- `InlineNewVersionForm` 被"加载中"状态阻塞,表单在首帧未渲染
- Playwright 无法找到 `data-testid="organization-form"`

**影响**:
- 阻塞完整 CRUD 流程验证
- 无法自动化回归创建新组织功能

**证据**:
- 截图: `test-results/business-flow-e2e-业务流程端到端测试-完整CRUD业务流程测试-chromium/test-failed-1.png`
- Trace: `test-results/.../trace.zip`

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

1. **调整 Hook 初始化逻辑** (前端团队,0.5 天)
   ```typescript
   // useTemporalMasterDetail.ts
   // 建议修复:创建模式下直接设置 isLoading=false
   const [isLoading, setIsLoading] = useState(
     organizationCode !== null // 仅编辑模式需要加载
   );
   ```

2. **确保表单模式初始化** (前端团队,0.5 天)
   ```typescript
   // 创建模式下立即初始化 formMode 和 formInitialData
   useEffect(() => {
     if (isCreateMode) {
       setFormMode('create');
       setFormInitialData(defaultFormData);
       setIsLoading(false);
     }
   }, [isCreateMode]);
   ```

3. **回归验证** (QA 团队,0.5 天)
   ```bash
   # 执行完整业务流程测试
   PW_JWT=$(cat .cache/dev.jwt) PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 \
   npm --prefix frontend run test:e2e -- --project=chromium \
   tests/e2e/business-flow-e2e.spec.ts
   ```

4. **归档验证报告** (QA 团队,0.25 天)
   - 创建 `reports/iig-guardian/plan18-phase1.3-validation-<date>.md`
   - 更新 06 号日志"当前状态"

### 验收标准
- [ ] 业务流程 E2E 通过率 ≥ 95% (5/5)
- [ ] 创建/编辑/删除完整流程截图与视频
- [ ] 测试报告归档至 `reports/iig-guardian/`

---

## 📋 Phase 2-3: 长期优化 (待排期)

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
- [06-integrated-teams-progress-log.md](./06-integrated-teams-progress-log.md)
- [16-code-smell-analysis-and-improvement-plan.md](./16-code-smell-analysis-and-improvement-plan.md)
- [Playwright RS256 验证报告](../../reports/iig-guardian/playwright-rs256-verification-20251002.md)
- [E2E 测试指南](../../docs/development-tools/e2e-testing-guide.md)

### 技术参考
- [Playwright 官方文档](https://playwright.dev/)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)

---

**本文档状态**: ✅ 已精简,聚焦核心待办事项
**下一步行动**: 启动 Phase 1.3 修复创建表单渲染问题
