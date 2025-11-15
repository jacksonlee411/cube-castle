# Plan 240BT – 组织详情“白屏”治理与路由模块解耦

编号: 240bt  
上游: Plan 240B（数据装载与等待治理）  
状态: 已完成（验收通过 · 2025-11-15）

—

## 问题陈述
- 现象：进入组织架构详情页（如 `/organizations/1000001/temporal`）后出现空白页面（无错误提示）。
- 影响：阻断详情功能使用，误导为数据问题；与 240B“避免白屏”的验收目标冲突。

## 根因与证据（唯一事实来源）
- 路由懒加载的模块静态耦合，导致模块评估失败时整体白屏（高概率主因）
  - 组织与职位两个详情路由共用同一懒加载模块 `entityRoutes.tsx`；即便仅访问组织详情，也会在模块评估阶段加载职位详情依赖。一旦职位侧依赖链出现运行期错误，整个懒加载模块评估失败，因缺少边界处理，呈现“白屏”。
  - 证据：
    - 懒加载位置：`frontend/src/App.tsx:13`、`frontend/src/App.tsx:18`
    - 共用模块导入组织与职位详情：`frontend/src/features/temporal/pages/entityRoutes.tsx:6`、`frontend/src/features/temporal/pages/entityRoutes.tsx:7`
    - Dev HMR 关闭错误浮层，错误不直观：`frontend/vite.config.ts:36`

- 认证/JWKS 校验失败导致请求级异常（中高概率并发问题，容易与上项叠加）
  - 统一客户端在请求前强制校验 RS256 并访问 `/.well-known/jwks.json`；若命令服务未暴露 RS256 JWKS，则 `getAccessToken()` 会抛错，所有请求失败。该错误在没有上层 ErrorBoundary 时，易被“懒加载失败”掩盖为白屏。
  - 证据：JWKS 校验逻辑 `ensureRS256()`：`frontend/src/shared/api/auth.ts:272`

- 数据层兜底逻辑健全，非白屏直接原因（旁证）
  - 组织版本 GraphQL 失败会回退 snapshot 并提示：`frontend/src/features/temporal/components/hooks/temporalMasterDetailApi.ts:157`
  - 审计历史失败呈现错误卡片而非白屏：`frontend/src/features/audit/components/AuditHistorySection.tsx:71`
  - 说明：即使数据不可用，页面应显示错误提示，不应“白屏”。因此“白屏”更符合“路由模块评估失败”的特征。

## 与 240B 的关系
- 240B 目标强调“避免白屏、统一装载、可取消/可重试”。本问题的主因是“路由懒加载模块静态耦合”，不直接属于装载策略，但会先于装载阶段导致页面无法呈现。解决该耦合后，再继续 241/245 统一 Loader/Hook 的推广，可共同满足“避免白屏”的验收目标。

## 准入与约束（遵循 AGENTS.md）
- Docker 强制：通过 `make docker-up` 启动所有基础设施，禁止在宿主机直接安装 PostgreSQL/Redis/Temporal。
- 健康检查：`curl http://localhost:9090/health`、`curl http://localhost:8090/health` 均返回 200。
- 鉴权链路：`make jwt-dev-mint` 生成 `.cache/dev.jwt`；`curl http://localhost:9090/.well-known/jwks.json` 需 200。
- 契约先行：不修改 GraphQL/OpenAPI 契约；如需扩展，先更新 `docs/api/*` 并跑脚本，后实现。

## 范围与不做
- 范围：
  - 路由模块解耦（组织与职位详情分离懒加载单元）
  - 增设路由级 ErrorBoundary（捕获 React.lazy 加载/评估异常）
  - 环境守卫：JWKS/RS256 校验自检与提示（不改客户端契约）
- 不做：
  - 不改 API 契约与端口映射（严格遵守 Docker 端口与代理配置）
  - 不引入第二事实来源的数据读取逻辑（继续复用现有 Hook/Adapter）

## 方案设计
1) 路由模块拆分（推荐，必做）
   - 新增：
     - `frontend/src/features/temporal/pages/organizationRoute.tsx` 仅导出 `OrganizationTemporalEntityRoute`（不导入职位）
     - `frontend/src/features/temporal/pages/positionRoute.tsx` 仅导出 `PositionTemporalEntityRoute`
   - 修改：`frontend/src/App.tsx` 将原对 `entityRoutes` 的两个懒加载替换为上述两个独立模块，杜绝跨实体静态耦合导致的“牵连白屏”。

2) 错误边界兜底（可选，建议）
   - 在 `renderWithAuth()` 外围增加 ErrorBoundary；对 React.lazy 加载/评估错误与运行期错误统一呈现错误卡片，避免白屏。

3) 环境与鉴权守卫（必做，配置项）
   - 启动链路固定化：`make docker-up` → `make run-dev` → `make frontend-dev`；如端口占用先卸载宿主冲突服务（严禁改映射）。
   - 若 JWKS 4xx/5xx：使用 `make run-auth-rs256-sim` 或 `make jwt-dev-mint` 修复；前端页面展示“认证配置未就绪”的提示文案，不白屏。

## 任务清单（可执行项）
1. 代码解耦
   - 新增两个独立路由模块（组织/职位），提取自 `entityRoutes.tsx`，仅保留对应实体依赖。
   - 更新 `frontend/src/App.tsx` 两处懒加载指向新模块（参考 `frontend/src/App.tsx:13`、`frontend/src/App.tsx:18`）。
2. 错误边界
   - 实现 `AppErrorBoundary`（轻量）：展示错误标题、错误信息、返回主页/重试按钮。
   - 将 `renderWithAuth()` 包裹进 ErrorBoundary；不改变 Suspense 结构与鉴权逻辑。
3. 环境守卫与文案
   - 在“登录/提示”界面，当捕获到 `ensureRS256()` 报错时显示明确提示（JWKS 未就绪），引导执行 `make jwt-dev-mint`。
4. 验证与记录
   - 本地手动验证：访问 `/organizations/:code/temporal` 渲染正常；控制台无“Chunk load/Module evaluation”报错。
   - 故障注入验证：暂时让职位侧页面抛出可控错误（开发态），确认组织详情不再白屏（仅作为本地验证，不提交此变更）。
   - E2E 冒烟：新增或复用“组织详情壳渲染”用例（smoke），记录日志与 HAR（与 240B 产物一致）。

## 验收标准
- 路由解耦：组织详情路由懒加载的模块不再依赖职位侧组件（静态依赖分析通过）。
- 白屏治理：
  - 正常环境：访问 `/organizations/:code/temporal` 显示主从视图骨架并渲染详情；
  - 非正常环境（JWKS 未就绪 / GraphQL 不可用）：页面显示错误提示或回退提示；“不出现整页白屏”。
- 控制台/网络：访问组织详情时无“Chunk load error”、“Module evaluation failed”报错；`/.well-known/jwks.json` 200。
- 测试证据：
  - 开发态冒烟日志：`logs/plan240/BT/smoke-org-detail.log`
  - HAR（可选）：`logs/plan240/BT/network-har-*.har`

## 风险与回滚
| 风险 | 影响 | 缓解/回滚 |
| --- | --- | --- |
| 路由拆分引入路径/导出名错误 | 详情页不可达 | 加 ErrorBoundary；快速回滚到 `entityRoutes`（同 MR 中保留切换开关） |
| 错误边界误捕获正常异常 | 阻断正常渲染 | 边界上报 `requestId`/路径，保留“继续渲染”选项 |
| 环境校验误判 | 误导用户操作 | 文案清晰标注“JWKS 未就绪”，并提示 `make` 命令 |

## 时间线（建议）
- Day 1：路由拆分与最小改动接入（App.tsx）；开发态验证通过；提交 MR。
- Day 2：ErrorBoundary 与文案补充；本地冒烟 + HAR；产出日志与截图。
- Day 3：E2E（smoke）落盘；与 240B 的 Loader/Hook 推进计划衔接。

## 一致性与质量守卫
- 不引入第二事实来源：继续使用现有 Hook/Adapter 与 GraphQL 文档。
- 代码风格与校验：`npm run guard:plan245`、`npm run guard:selectors-246`、`npm run lint`、`node scripts/quality/architecture-validator.js`。
- Docker 强制：遵循 `AGENTS.md`，禁止改端口映射；如占用，卸载宿主服务。

## 执行指令（冒烟 + HAR 落盘）
- 本地/CI（默认由 Playwright 启动 dev server；如已运行本地前端可设 `PW_SKIP_SERVER=1`）
```bash
cd frontend
npm run test:e2e:240bt
```
- 产物：`logs/plan240/BT/health-checks.log`、`logs/plan240/BT/network-har-*.har`、`frontend/playwright-report/`

## 附：关键引用（用于复核）
- 懒加载组织/职位的共用模块导入：`frontend/src/features/temporal/pages/entityRoutes.tsx:6`、`frontend/src/features/temporal/pages/entityRoutes.tsx:7`
- App 懒加载 `entityRoutes`：`frontend/src/App.tsx:13`、`frontend/src/App.tsx:18`
- Dev HMR 关闭 overlay：`frontend/vite.config.ts:36`
- JWKS/RS256 强制校验：`frontend/src/shared/api/auth.ts:272`
- 组织版本回退兜底：`frontend/src/features/temporal/components/hooks/temporalMasterDetailApi.ts:157`
- 审计历史错误呈现：`frontend/src/features/audit/components/AuditHistorySection.tsx:71`

—

备注：本计划不改动 API 契约；实施完成后，请在 240B 计划的“避免白屏/统一装载”验收记录中标注“路由解耦完成”，并将本计划归档到 `docs/archive/development-plans/`。 

---

## 完成登记（2025-11-15）

验收总结（以唯一事实来源为准；不复制实现细节）：

- 路由解耦已完成  
  - App 路由单元不再依赖共用 `entityRoutes.tsx`，改为分别懒加载：  
    - `frontend/src/App.tsx:15`、`frontend/src/App.tsx:20`  
    - `frontend/src/features/temporal/pages/organizationRoute.tsx:49`  
    - `frontend/src/features/temporal/pages/positionRoute.tsx:50`  
  - 组织详情模块文件未导入职位依赖；职位模块独立。旧 `entityRoutes.tsx` 已不被 App 入口消费（保留仅为迁移期兼容，不作为入口）。
- 错误边界已接入  
  - 懒加载入口包裹于 `AppErrorBoundary`：`frontend/src/App.tsx:27`、`frontend/src/App.tsx:35`。  
  - 目的：React.lazy 加载/评估异常不再导致整页白屏。
- 冒烟用例通过（Chromium/Firefox 各 1 轮）  
  - 证据：`logs/plan240/BT/smoke-org-detail.log`（最终 2 passed），并附带失败到通过的重试记录与 trace 路径。  
  - 健康检查均 200：`logs/plan240/BT/health-checks.log`（包含 `/.well-known/jwks.json` 200）。  
  - 用例入口：`frontend/tests/e2e/smoke-org-detail.spec.ts`
- 控制台错误  
  - 冒烟成功轮次未出现 “Chunk load error / Module evaluation failed”。如遇失败，ErrorBoundary 已确保不会出现“整页白屏”。  

对齐与指针：
- 本计划不新增选择器与事件定义；选择器来源：`frontend/src/shared/testids/temporalEntity.ts`。  
- 事件词汇表与采集策略：`docs/reference/temporal-entity-experience-guide.md`。  
- 影响同步：已满足 240B “避免白屏/统一装载”的前置路由解耦条件；建议在 240B 验收记录中标注“路由解耦完成”。  

归档说明：  
- 建议将本计划在后续 MR 中归档至 `docs/archive/development-plans/`，并在主计划 240 的“完成登记与现状评估”章节保留索引与证据路径。 
