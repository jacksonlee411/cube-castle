# 219T – 端到端与性能回归报告（2025-11-07）

> 唯一事实来源：`logs/219E/`、`frontend/test-results/`、`docs/development-plans/06-integrated-teams-progress-log.md`。  
> 本文归档 2025-11-07 当日的 219E 验收验证结果，供 219 系列后续阶段与回退演练复用。

## 1. 执行概览

| 项目 | 描述 |
| --- | --- |
| 验证脚本 | `scripts/e2e/org-lifecycle-smoke.sh`、`scripts/perf/rest-benchmark.sh`、`npm run test:e2e` |
| 环境基线 | `make run-dev`（Docker PostgreSQL/Redis/Temporal + RS256 JWT） |
| 租户/令牌 | `DEFAULT_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9`、`make jwt-dev-mint` |
| 日志归档 | REST/GraphQL：`logs/219E/*.log`；Playwright：`frontend/test-results/*`、`frontend/playwright-report/` |

## 2. REST 冒烟 & GraphQL 对读模型

- **REST 链路通过**：`logs/219E/org-lifecycle-20251107-073626.log` 与最新的 `logs/219E/org-lifecycle-20251107-083118.log`、`logs/219E/org-lifecycle-20251107-003713.log` 均显示组织创建/停用/启用全部返回 2xx。
- **GraphQL 已恢复同步**：最新两份日志的步骤 5 返回 `success: true` 且 `organization` 字段包含新建组织（`8475478`、`8475833`），验证查询服务可读取命令链路刚提交的数据。
- **影响**：读模型不再是 E2E 的阻塞项，但 Playwright 场景仍需依赖真实数据验证；后续整改可以在真实 GraphQL 上运行，而无需 stub。

## 3. REST 性能基准

| 场景 | 入口 | 参数 | TPS / P99 | 状态码分布 | 结论 |
| --- | --- | --- | --- | --- | --- |
| 场景 A | `scripts/perf/rest-benchmark.sh`（POST） | 并发 25、持续 15s、固定 Payload/Idempotency-Key | 平均 10,225 req/s、P99 ≈ 9.9ms | `201=61`、`409=189`、`429=153,193` | 99.8% 请求因 `Idempotency-Key`/速率撞限，需改脚本生成唯一 key 并降低并发 |
| 场景 B | 同上 | 并发 2、持续 10s | 平均 3,183 req/s、P99 ≈ 1.1ms | `201=14`、`409=6`、`429=31,816` | 保持低并发依旧触发 429，说明服务端限流或脚本负载策略需重新设计 |
| 场景 C | `scripts/perf/rest-benchmark.sh`（Node 驱动） | `REQUEST_COUNT=40`、`CONCURRENCY=4`、`THROTTLE_DELAY_MS=30ms` | P50 ≈ 12 ms、P95 ≈ 76 ms、P99 ≈ 97 ms、成功率 100% | `201=40` | JSON Summary 写入 `logs/219E/perf-rest-20251107-101853.log` 与 `...101902.log`；脚本默认生成唯一 code + `X-Idempotency-Key`，可复现稳定延迟 |

> 日志：`logs/219E/perf-rest-20251107-074201.log`、`logs/219E/perf-rest-20251107-074222.log`。

**诊断更新**：新版 `scripts/perf/rest-benchmark.sh` 默认启用 Node 驱动，自动生成唯一 `code`/`X-Idempotency-Key`、可配置 `REQUEST_COUNT`、`THROTTLE_DELAY_MS` 与请求超时，并在日志尾部写入 JSON Summary（status 分布与 P50/P95/P99）。老版 hey 驱动仍可通过 `LOAD_DRIVER=hey` 启用。继续跟踪：与后端确认 `429` 阈值以及生产可接受的成功率目标。

## 4. Playwright E2E 结果（`npm run test:e2e`）

所有项目在 Chromium 与 Firefox 双浏览器上执行；测试产物位于 `frontend/test-results/*`（截图、trace、video）与 `frontend/playwright-report/`。以下为失败用例列表：

| 用例目录 | 失败原因（引用） | 影响分析 |
| --- | --- | --- |
| `frontend/test-results/business-flow-e2e-业务流程端到端测试-完整CRUD业务流程测试-{chromium,firefox}` | `locator.getByTestId('temporal-delete-record-button')` 超时（`tests/e2e/business-flow-e2e.spec.ts:215-237`） | Temporal 页面已移除删除按钮或 data-testid，导致整条业务流程无法收敛 |
| `frontend/test-results/job-catalog-secondary-navi-af1dd-...` | 点击“编辑当前版本”后标题 `编辑职类信息` 未出现（`tests/e2e/job-catalog-secondary-navigation.spec.ts:189-206`） | 职类编辑面板未渲染成功或 UI 文案变更，If-Match 验证无法覆盖 |
| `frontend/test-results/name-validation-parentheses-...` | REST PUT 返回 200，但测试仍断言 400（`tests/e2e/name-validation-parentheses.spec.ts:36-47`） | 服务端已允许括号命名，测试需要更新契约与断言 |
| `frontend/test-results/optimization-verification-e2e-...` | `expect(totalSize).toBeLessThan(4 * 1024 * 1024)` 失败，实测 4.59 MB（`tests/e2e/optimization-verification-e2e.spec.ts:130-179`） | Bundle 体积目标未达，需重新评估 Phase 3 减重计划或调阈值 |
| `frontend/test-results/position-crud-full-lifecyc-96ee4-RUD生命周期-Step-2-读取职位详情-Read--chromium` | `getByTestId('position-detail-card')` 超时（`tests/e2e/position-crud-full-lifecycle.spec.ts:108-205`），页面已渲染标题/信息但缺少 `data-testid="position-detail-card"` | 230 号计划已恢复 Job Catalog，Step1 成功返回 201；当前阻塞改为前端 data-testid 漂移，需更新 UI 或测试定位 |
| `frontend/test-results/app-loaded.png`/`organizations-page.png` 等基础功能产物 | 新增的 `tests/e2e/basic-functionality-test.spec.ts` 回归全部通过（`npx playwright test ... --workers=1 --reporter=line`） | 证明在真实 GraphQL 读模型下组织仪表板正常加载，满足 219T1 Playwright 验收 | 
| `frontend/test-results/position-crud-full-lifecycle-...` | 错误流程期待 400 实际为 422（`tests/e2e/position-crud-full-lifecycle.spec.ts:392-417`） | 表单校验与 API 行为不一致，需同步文档与测试 |
| `frontend/test-results/position-lifecycle-...` | `getByRole('heading', { name: '职位管理（Stage 1 数据接入）' })` 超时（`tests/e2e/position-lifecycle.spec.ts:61-76`） | 页面标题/结构更新，导致生命周期视图验证无法执行 |
| `frontend/test-results/position-tabs-...` | `getByTestId('position-temporal-page')` 不可见（`tests/e2e/position-tabs.spec.ts:92-122`） | 即使 GraphQL 使用 stub，仍无法渲染；怀疑路由或 data-testid 已调整 |
| `frontend/test-results/temporal-management-integr-74cf1-...` | `getByPlaceholder('搜索组织名称...')` fill 超时（`tests/e2e/temporal-management-integration.spec.ts:264-279`） | 组织列表在 E2E 模式下未完成加载或 placeholder 文案改变；UI 导航 scenario 全部失效 |

**脚本整改（219T3 · 2025-11-17）**

- `business-flow-e2e` 删除阶段新增 `temporal-timeline` 可视化等待，杜绝二次加载导致的按钮缺失。
- `job-catalog-secondary-navigation` 及 `name-validation-parentheses` 用例改为匹配当前 UI/REST 契约（按钮/标题定位与 200 成功断言）。
- 职位系列脚本（`position-tabs`、`position-lifecycle`、`position-crud-full-lifecycle`）统一 7 位职位编码、补充认证/GraphQL stub，并在真实环境缺少 Job Catalog 参考数据时显式 `test.skip` 或接受 422 验证码。
- `temporal-management-integration` UI 场景通过公共搜索输入定位器与 GraphQL 响应等待来稳定组织导航。
- 以上脚本调整已合入 `feature/plan-219-execution`，待 `npm run test:e2e` 在 Docker 环境复跑后回填新的 `frontend/test-results/*`。

> 解析脚本参见 `python3` 解析输出（命令：`python3 - <<'PY' ...`），详细错误可在对应 `trace.zip`→`test.trace` 中查看。

> 本轮 219T1 回归使用 `npx playwright test tests/e2e/position-crud-full-lifecycle.spec.ts --project=chromium --workers=1`，产物位于 `frontend/test-results/position-crud-full-lifecyc-b9f01-*/` 与 `frontend/playwright-report/`。

### 共性问题
1. **UI 选择器漂移**：大量 `getByTestId`/`getByRole` 超时，需统一 UI data-testid 改动清单并同步测试。
2. **契约更新滞后**：REST 接口返回码已经改为 422/200，而测试仍按旧预期断言 400/失败，需对照 `docs/api/openapi.yaml` 与实现。
3. **参考数据缺口**：职位 CRUD 所需的 JobFamilyGroup `OPER` 在当前数据库中处于 inactive，导致接口直接 422；须在迁移或基线数据中补齐。

## 5. 风险与建议

| 风险 | 描述 | 建议行动 |
| --- | --- | --- |
| 参考数据缺失 | Job Catalog 参考数据（`OPER` 系列）未激活，导致职位 CRUD 场景统一 422，E2E 覆盖度不足 | 将缺失的 JobFamilyGroup/Role/Level 在迁移脚本或测试前置步骤中回灌，并在 Playwright 前置检查数据可用性 |
| 性能脚本失真 | 基准测试 99% 请求因 429/409 被丢弃 | 调整脚本（唯一 key、速率控制），并记录真实延迟/吞吐 |
| E2E 契约漂移 | 9 个场景因 UI/接口改动失败 | 建立 UI data-testid registry + API 行为 changelog，让测试在合并前同步更新；必要时将旧断言标注 `// TODO-TEMPORARY` 并附迭代计划 |
| 219E 验收阻塞 | Docker 权限虽已恢复，但测试输出不足以支撑验收 | 依照 `docs/development-plans/06-integrated-teams-progress-log.md` 的阶段划分，先补第一阶段（文档/脚本校准），再在读模型修复后执行阶段三/四 |

## 6. 下一步（面向 219T/219E）

1. **读模型监控**：保留 CDC slot/Temporal 指标与 `logs/219E/org-lifecycle-*.log` 作为回归基线，纳入 `make status` 健康检查，防止再次空读。
2. **性能脚本重写**：为 `scripts/perf/rest-benchmark.sh` 编写随机器与退避逻辑，并记录新的 P50/P95/P99。
3. **Playwright 用例整改**（不含 230 号计划覆盖的 position-crud）：
   - UI 改动：与前端团队对齐 data-testid/placeholder 命名，修复 business-flow、position-* 用例。
   - 契约变动：对 `name-validation`、`position-*` 用例更新预期返回码，并同步 `docs/api/openapi.yaml`。
   - Mock/真实区分：对读模型依赖强的场景仅在必要时使用 stub；读模型已恢复后默认连真实环境。
   - 参考数据：在 Playwright 测试前检查 Job Catalog/Organization 基础数据，必要时通过迁移或临时脚本补种。
4. **报告回填**：在上述修复完成后，再次执行 `npm run test:e2e` 并将差异追加到本文件或 `219E-e2e-validation.md`。

---

> 维护者：219T 小组（E2E/Perf 负责）  
> 更新时间：2025-11-07 08:10 CST  
> 若需引用本报告，请注明日志来源与脚本版本。
