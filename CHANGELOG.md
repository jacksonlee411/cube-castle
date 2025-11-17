# Cube Castle 项目变更日志

## v1.6.3 - CI：Plan‑254 恢复并稳定（3×绿）+ JWT mint 调用修复 (2025-11-17)

### ✅ 门禁与流水线
- Plan‑254 Gate（前端端点与代理整合）恢复为正式工作流并稳定通过（连续 3 次成功）  
  - 端点统一：前端与 E2E 统一经单体进程（:9090）访问 `/api/v1` 与 `/graphql`；Vite 代理与端点 SSoT 对齐  
  - E2E：DevServer 自启（PW_SKIP_SERVER=0，PW_BASE_URL=http://localhost:3000），证据与报告作为工件上传（plan254-logs）  
  - 运行链接：run 53/59/60（详见 215 执行日志）
- CI 修复：`make jwt-dev-mint` 在 runner 上改为 `bash scripts/dev/mint-dev-jwt.sh` 调用，规避可执行位差异导致的 Permission denied

### 📚 文档与登记
- 215 执行日志新增“Plan‑254 门禁恢复并稳定”条目，登记 run 链接与工件索引

## v1.6.2 - 文档：PR 策略与回切规则澄清 (2025-11-17)

### 📚 文档
- AGENTS.md：明确“主干（本地 master）+ 远程仅通过分支 PR（squash-merge）”策略，并新增“PR 合并后安全回切 master（ff-only + 清理分支）”强制要求。
- docs/reference/01-DEVELOPER-QUICK-REFERENCE.md：新增“分支与 PR 工作流（主干 + 远程 PR）”与回切示例命令。
- .github/pull_request_template.md：增加远程策略提醒（master 禁直推、PR 必有 Issue 与 Required checks）。
- Makefile/scripts：新增 `scripts/ops/configure-branch-protection.sh` 和 `make protect-branch` 入口（一键配置仓库保护），不影响运行时行为。

> 说明：仅文档/工具补充，未改变现有 CI 或业务行为；以 `AGENTS.md` 为唯一事实来源。

## v1.6.1 - REST 业务查询端点弃用公告（Plan 259‑T4 启动）(2025-11-16)

### ⚠️ 弃用（Deprecation）
- 弃用 REST 查询端点：`GET /api/v1/positions/{code}/assignments`（迁移至 GraphQL）
  - 目标：清零“业务查询类 REST GET”，避免 OpenAPI/GraphQL 双事实来源；与 PostgreSQL 原生 CQRS 对齐（命令=REST、查询=GraphQL）
  - 迁移路径（GraphQL）：
    - `positionAssignments(positionCode, filter, pagination, sorting)`
    - `assignments(organizationCode, positionCode, filter, pagination, sorting)`
  - Sunset 时间：2025‑12‑20 00:00:00Z（届时计划移除 REST 端点）
  - 合同标识：OpenAPI 已设置 `deprecated: true`，并在 200 响应示例中加入 `Sunset` 与 `Link` 响应头
    - `Sunset: Sat, 20 Dec 2025 00:00:00 GMT`
    - `Link: <https://api.yourcompany.com/docs/migrations/positions-assignments-to-graphql>; rel="deprecation"`
  - 权限不变：GraphQL 与 REST 均使用 `position:assignments:read`（Plan 259‑T3 已对齐）

### 📚 参考与登记
- 方案与决议：`docs/development-plans/259-protocol-strategy-review.md`、`docs/development-plans/259A-protocol-duplication-and-whitelist-hardening.md`
- 执行登记：`docs/development-plans/215-phase2-execution-log.md`（记录 T3 对齐与 T4 启动）

## v1.6.0 - 契约漂移门禁完成（Plan 258 关闭）(2025-11-16)

### ✅ 门禁与契约
- 启用“契约漂移门禁”双阻断（Plan 258）：
  - Phase A：OpenAPI ↔ GraphQL 枚举差异（阻断）
  - Phase B：主实体字段矩阵（字段/类型/可空/列表）差异（阻断）
- GraphQL 契约对齐：
  - `Organization.profile: JSON`（与 REST object 对齐）
  - `Organization.sortOrder: Int!`（与 REST 非空对齐）
- 白名单（短期）：仅保留“存在性差异”（审计/派生/写侧元信息），在 `scripts/contract/drift-allowlist.json` 登记并在 215 标注回收期。

### 🧪 CI 与证据
- 受保护分支 Required：`plan-258-gates`（Contract Drift Gate (Plan 258)）
- 最近一次成功运行 Run ID：19408157081（artifact：`plan258-drift-report`）
- 报告路径：`reports/contracts/drift-report.json`（字段矩阵差异仅剩存在性差异）

### 📚 文档与登记
- 258 方案文档：`docs/development-plans/258-contract-drift-validation-gate.md`（状态：已完成）
- 215 执行日志：登记 Run ID 及工件索引、回收计划

> 说明：保持“单一事实来源”，本条仅索引路径与 run id；不复制报告正文或链接。

## v1.5.9 - Plan 240 全面完成与归档（2025-11-15）

### ✨ 代码与框架
- 引入最小骨架 `TemporalEntityLayout.Shell` 并包裹组织/职位路由，注入性能标记（不改变 DOM/testid，不改对外契约）
  - 前端：`frontend/src/features/temporal/layout/TemporalEntityLayout.tsx`
  - 组织路由包裹：`frontend/src/features/temporal/pages/organizationRoute.tsx`
  - 职位路由包裹：`frontend/src/features/temporal/pages/positionRoute.tsx`
- 选择器统一（职位域）
  - 增补集中选择器 `position.form(mode)`：`frontend/src/shared/testids/temporalEntity.ts`
  - 组件替换为 SSoT：
    - `frontend/src/features/positions/components/PositionForm/index.tsx`
    - `frontend/src/features/positions/components/dashboard/PositionHeadcountDashboard.tsx`
    - `frontend/src/features/positions/components/dashboard/PositionVacancyBoard.tsx`
    - `frontend/src/features/positions/components/transfer/PositionTransferDialog.tsx`

### 📚 文档与治理
- Plan 240：标记“已完成（验收通过）”，新增“0.1 影响评估：240 先于 241 完成的回补计划”
  - `docs/development-plans/240-position-management-page-refactor.md`
- Plan 240B：新增硬依赖“240BT 路由解耦完成”标注
  - `docs/development-plans/240B-position-loading-governance.md`
- Plan 240BT：验收完成并归档；开发目录下改为“已归档占位符”
  - 归档：`docs/archive/development-plans/240bt-org-detail-blank-page-mitigation.md`
  - 占位：`docs/development-plans/240bt-org-detail-blank-page-mitigation.md`
- Plan 240E：登记本地 Smoke 与守卫证据，新增“关闭确认”段落；215 执行日志同步
  - `docs/development-plans/240E-position-regression-and-runbook.md`
  - `docs/development-plans/215-phase2-execution-log.md`
- 文档索引：更新 240 为“已完成”，列出 241 子计划（A/B/C）
  - `docs/development-plans/HRMS-DOCUMENTATION-INDEX.md`
- 临时标签规范：统一为 `// TODO-TEMPORARY(YYYY-MM-DD): ...`
  - `AGENTS.md`、相关计划文档与参考手册同步修订

### 🧪 验收与证据
- 守卫（通过）：
  - 选择器守卫：`logs/plan240/E/selector-guard.log`
  - 架构守卫：`logs/plan240/E/architecture-validator.log`
  - 临时标签检查：`logs/plan240/E/temporary-tags.log`
- Smoke（Chromium）：6 passed / 1 skipped（通过）
  - 证据：`logs/plan240/E/playwright-smoke-20251115142851.log`
- CI 与工具链：
  - 新增工作流：`.github/workflows/plan-240e-regression.yml`
  - 统一脚本：`scripts/plan240/run-240e.sh`、`scripts/plan240/trigger-240e-ci.sh`、`scripts/plan240/record-240e-acceptance.sh`

### 🔄 后续（241 对接）
- 241 完成后按 240“0.1 影响评估”回补：骨架切换至共享 Layout、Hook/Loader 统一、可观测性归一、Feature Flag 收敛与 E2E 复跑（不引入第二事实来源；契约先行）。

## v1.5.7 - 职位详情可观测性落地（Plan 240D）(2025-11-15)

### ✨ 新增
- 前端观测发射器（极薄封装）：`frontend/src/shared/observability/obs.ts`
- 观测事件注入（职位详情）：
  - 首屏 Hydration：`position.hydrate.start/.done`（含 `durationMs`）
  - 页签切换：`position.tab.change`（`tabFrom/tabTo`）
  - 版本选择：`position.version.select`
  - 版本导出：`position.version.export.start/.done/.error`（含 `durationMs/sizeBytes`）
  - GraphQL 错误：`position.graphql.error`（`queryName/status`）

### 🔧 变更
- 移除运行时别名事件与重复定义；事件/Schema 仅引用 `docs/reference/temporal-entity-experience-guide.md`（单一事实来源）
- 统一落盘路径：`logs/plan240/D/obs-*.log`（E2E 采集写入，运行时代码不落盘）
- 门控与通道：`VITE_OBS_ENABLED` + `VITE_ENABLE_MUTATION_LOGS`；生产不输出信息级 `[OBS]`

### ✅ 验收与证据
- 用例：`frontend/tests/e2e/position-observability.spec.ts`
- 报告：`frontend/playwright-report/index.html`
- 证据：`logs/plan240/D/obs-position-observability-chromium.log`

> 240D 已完成并登记，详见 `docs/development-plans/240D-position-observability.md` 的“完成登记”章节。

## v1.5.8 - Timeline/Status 抽象完成（Plan 244）(2025-11-15)

### ✨ 抽象合并
- 统一 Temporal Timeline Adapter 与 Status 元数据：`frontend/src/features/temporal/entity/timelineAdapter.ts`、`statusMeta.ts`
- 组织/职位页面全面引用新命名空间，移除旧路径引用

### ✅ 验收与证据
- E2E（Chromium/Firefox 各 1 轮）：
  - `frontend/tests/e2e/smoke-org-detail.spec.ts`（通过）
  - `frontend/tests/e2e/temporal-header-status-smoke.spec.ts`（通过）
  - `frontend/tests/e2e/temporal-management-integration.spec.ts`（8 passed / 4 skipped）
- 证据：
  - `logs/plan242/t2/244-e2e-acceptance.log`
  - `frontend/playwright-report/index.html`
  - 命名空间扫描：`logs/plan242/t2/244-namespace-scan.log`

## 未发布 - 数据库基线重建（2025-11-06）

### 🛠️ 基础设施
- **数据库迁移体系重建**：生成 `database/schema.sql` 为唯一事实来源，输出 Goose 基线迁移 `20251106000000_base_schema.sql`，并清理历史脚本。
- **迁移工具链切换**：新增 `goose.yaml`、`atlas.hcl` 与 `scripts/generate-migration.sh`，Makefile 和 CI 改为使用 Goose `up/down`。
- **验证资产**：新增 `tests/integration/migration_roundtrip_test.go`，确保 Goose 迁移支持 up→down→up 循环；在 `schema/` 输出备份与 diff 清单（现已归档至 `docs/archive/schema-snapshots/`）。

### 📚 文档
- `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`docs/reference/03-API-AND-TOOLS-GUIDE.md` 同步更新 Goose/Atlas 使用方式。
- `scripts/README.md` 与 Plan 210 文档补充新流程、回滚策略与检查清单。

## v1.5.6 - 文档与治理对齐（Plan 247）(2025-11-14)

### 📖 文档治理
- 新增并确立唯一事实来源：《Temporal Entity Experience Guide》：`docs/reference/temporal-entity-experience-guide.md`
- 旧文档处理：Positions 早期指南路径改为“Deprecated 占位符”（无业务正文，仅提示迁移路径），避免第二事实来源；计划下个迭代移除
- 快速参考更新：在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 增补“Temporal Entity 命名与文档入口”，仅链接权威文档
- 计划日志登记：`docs/development-plans/06-integrated-teams-progress-log.md` 与 `docs/development-plans/215-phase2-execution-log.md` 登记 Plan 247 完成与证据路径

### 🧾 验收与证据
- 零引用检查（排除 `docs/archive/**`）：`logs/plan242/t5/rg-zero-ref-check.txt`
- 文档同步与架构守护运行日志：`logs/plan242/t5/document-sync.log`、`logs/plan242/t5/architecture-validator.log`
- 实现清单快照哈希：`logs/plan242/t5/inventory-sha.txt`（对应 `reports/implementation-inventory.*`）

> 本次仅涉及文档与治理，未引入代码行为变更。

## v1.5.5 - 前端日志统一与ESLint零告警方案 (2025-10-02)

### ✨ 新增
- **统一日志工具**：新增 `frontend/src/shared/utils/logger.ts`，按环境分级输出并提供 `mutation` 专用日志接口
- **日志单元测试**：覆盖调试模式、生产模式及分组日志行为，确保桥接层行为稳定

### 🔧 改进
- **移除 `console.*`**：`frontend/src` 目录全部迁移至 `logger`，强化 CQRS 调试日志与错误报表一致性
- **ESLint 门禁**：`no-console` 升级为 `error`，新增 CI 步骤校验 `eslint-disable-next-line camelcase` 例外说明
- **架构验证器扩展**：`scripts/quality/architecture-validator.js` 新增 `eslintExceptionComment` 规则和 `--rule` 过滤能力
- **文档更新**：开发者速查手册补充日志规范，Plan 20 验收标准同步调整

### ✅ 验证
- `npm run lint:frontend-api`（已知配置缺陷触发循环引用错误，详见提交备注）
- `npm run test`（覆盖 logger 单测）

---

## v1.5.4 - 移除未经审批的组织架构复制链接功能 (2025-09-21)

### 🔧 合规性修正
- **移除违规功能**：清理未经产品/治理流程审批的复制链接功能
  - 移除 `OrganizationDashboard.tsx` 中的"复制列表链接"按钮
  - 移除 `OrganizationTree.tsx` 中的复制链接、复制名称路径、复制编码路径按钮
  - 移除 `TemporalMasterDetailView.tsx` 中的相应复制功能
- **清理工具代码**：删除 `frontend/src/shared/utils/clipboard.ts` 及相关测试
- **测试文件更新**：移除相关测试用例并清理依赖

### ✅ 质量验证
- **ESLint**：✅ 通过，无 lint 错误
- **TypeScript**：✅ 通过类型检查
- **代码清理**：✅ 移除所有未使用的导入和依赖

### 📚 项目原则强化
- 重申"先契约后实现"核心原则的重要性
- 加强合规性检查流程，防止类似违规情况再次发生
- 完善文档生命周期管理和归档机制

---

## v1.5.3 - 质量门禁工具链升级与Go 1.23支持 (2025-09-17)

### 🔧 开发工具升级
- **golangci-lint 升级**：v1.55.2 → v1.61.0
  - 解决 Go 1.23 兼容性问题（之前版本构建于 go1.21.3，无法识别新语法特性）
  - 支持 Go 1.23 新特性：`for range` 常量语法、`slices` 扩展等
  - 安装路径：`~/.local/bin/golangci-lint`
- **gosec 配置优化**：v2.22.8
  - 创建符号链接至 `~/.local/bin/gosec` 便于访问
  - PATH 环境变量配置完善

### ✅ 质量门禁验证
- **make lint**：✅ 执行成功，发现并记录代码质量问题
  - errcheck: JSON encoder 错误未检查
  - unused: 未使用函数和字段清理
  - gosimple、staticcheck: 代码简化建议
- **make security**：✅ 执行成功，gosec 安全扫描正常运行
- **工具可执行性**：✅ `golangci-lint` 与 `gosec` 均可直接执行

### 📄 文档更新
- 更新 `docs/development-plans/06-integrated-teams-progress-log.md`：
  - 记录工具升级过程和版本信息
  - 更新质量门禁检查清单状态
  - 添加执行进度记录和验证结果

### 🎯 下一步
- 代码质量问题修复（基于 lint 结果）
- RS256 认证依赖配置验证

## v1.5.2 - 精简CLAUDE与跨文档导航完善 (2025-09-13)

### 📄 文档结构
- 精简 `CLAUDE.md` 为“核心原则 + 单一事实来源索引”，移除易变细节（变更通告、流程清单、脚本说明）。
- 易变内容迁移路径：
  - 变更通告/进展 → `docs/development-plans/` 与本 `CHANGELOG.md`
  - 开发前必检/禁止事项/操作清单 → `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
  - API 一致性细则与工具 → `docs/reference/03-API-AND-TOOLS-GUIDE.md`
  - 文档治理与目录边界 → `docs/DOCUMENT-MANAGEMENT-GUIDELINES.md`、`docs/README.md`

### 🔗 跨文档导航
- 在以下文档中补充/统一指向权威链接：
  - `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` → 增补指向 `CLAUDE.md`、`AGENTS.md`、`docs/api/*`、文档治理。
  - `docs/reference/03-API-AND-TOOLS-GUIDE.md` → 增补同类“进一步阅读与治理”链接。
  - `docs/development-plans/06-integrated-teams-progress-log.md` → 增补相关规范与参考链接。

### ✅ 影响范围
- 不涉及代码行为变更，仅文档组织与导航优化；CI 目录边界与文档同步检查不受影响。

## v1.5.1 - 文档与规则加固 (2025-09-13)

### 🔧 规则与门禁
- 架构验证器降误报：
  - 端口检查仅在 URL/port 键值对场景触发，跳过注释/样式数字。
  - CQRS 检查移除通用 `.get(` 匹配，保留 `fetch()/axios.get()` 精确检测。
  - 契约检查跳过注释、增加白名单（`client_credentials`、`cube_castle_oauth_token`）。
- 结果：`node scripts/quality/architecture-validator.js` 全通过（CQRS=0 / 端口=0 / 契约=0）。

### 🧩 契约命名归零（M-1）
- 修正 temporal 相关组件与类型的 snake_case → camelCase：
  - TimelineComponent、TemporalMasterDetailView、PlannedOrganizationForm、TemporalSettings、shared/types/temporal.ts。

### 🪝 废弃 Hook 替换（M-2）
- 业务侧使用 useEnterpriseOrganizations；测试 mock 同步调整。
- ESLint 限制：禁止导入 `shared/hooks/useOrganizations`（防回归）。

### 📝 文档/模板
- PR 模板：新增“契约命名自查”项（前端字段 camelCase 自检）。
- 06 进展日志：记录修复清单与里程碑完成状态。

## v1.5.0 - 文档治理与目录边界 (2025-09-13)

### 🗂️ 文档结构与治理强化
- 新增 `docs/reference/` 目录：承载长期稳定的参考文档（开发者快速参考、实现清单、API 使用指南、质量手册）。
- 开发计划归档迁移：`docs/development-plans/archived/` → `docs/archive/development-plans/`，统一归档入口新增 `docs/archive/README.md`。
- 文档导航更新：`docs/README.md` 提供 Reference vs Plans 分区导航与边界说明。
- 目录边界规则加入规范：更新 `docs/DOCUMENT-MANAGEMENT-GUIDELINES.md`，新增“目录边界（强制）”与“月度审计”检查项。

### 🔧 审查与CI门禁
- PR 模板（.github/pull_request_template.md）：新增“文档治理与目录边界（Reference vs Plans）”检查清单。
- CI（.github/workflows/document-sync.yml）：新增“目录边界检查”与“文档同步检查”，违规将自动评论并阻断；质量门禁输出纳入总判定。

### 🗺️ 文档链接修正
- 全面修正指向旧的 `docs/development-plans/archived/` 的链接为 `docs/archive/development-plans/`。
- 更新 `CLAUDE.md`、`AGENTS.md`、根 `README.md`，同步目录结构与最新规范。

---

## v1.4.0 - 企业级生产就绪版本 (2025-08-25)

### 🏆 重大架构革命
- **PostgreSQL原生CQRS架构**: 性能提升70-90%，查询响应时间1.5-8ms
- **架构简化**: 移除Neo4j依赖，简化架构60%，单一PostgreSQL数据源
- **数据一致性**: 消除CDC同步复杂性，实现强数据一致性

### 🎯 契约测试自动化体系
- **测试覆盖**: 32个契约测试100%通过
- **质量门禁**: CI/CD自动化验证，GitHub Actions + Pre-commit Hook
- **API一致性**: 字段命名规范100%合规，Schema验证完全通过
- **分支保护**: 企业级合并阻塞机制配置完成

### 🔧 关键修复成果
- **OAuth认证修复**: 解决client_id/client_secret字段名特例问题
- **GraphQL Schema映射**: 修复前端查询字段与后端Schema不匹配问题
- **企业级响应结构**: 统一API响应信封格式
- **JWT认证体系**: 开发/生产模式完善支持

### 📊 监控与质量保证
- **Prometheus集成**: 企业级监控指标收集
- **契约测试监控**: React监控仪表板集成到主应用
- **实时质量状态**: 契约遵循度实时监控
- **自动化验证**: Pre-commit Hook提供秒级反馈

### 🚀 生产就绪特性
- **Canvas Kit v13**: 完整兼容集成，TypeScript零错误构建
- **API契约遵循**: 85%符合度，核心功能完全达标
- **构建系统**: 2.47s生产构建时间，完全稳定
- **开发体验**: IDE配置优化，开发工具链完善

---

## v1.3.0 及更早版本

# 开发进展记录 - 2025年8月10日

## 🎯 重要里程碑: E2E测试体系完成

### 📊 测试成果
- **E2E覆盖率**: 92% (超过90%目标)
- **测试用例数**: 64个测试用例，6个测试文件
- **跨浏览器**: Chrome + Firefox 全面支持
- **测试框架**: Playwright + TypeScript

### 🔧 关键修复
1. **数据一致性测试修复**
   - 问题: API返回"ACTIVE"，前端显示"启用"
   - 解决: 添加状态字段本地化映射
   - 文件: `frontend/tests/e2e/business-flow-e2e.spec.ts:322-330`

2. **API兼容性测试修复**  
   - 问题: REST API数据结构误判
   - 解决: 修正数据结构断言 (REST API直接返回数据，不包装在'data'字段)
   - 文件: `frontend/tests/e2e/regression-e2e.spec.ts:71-72`

### 📈 性能验证结果
- **页面加载时间**: 0.5-0.9秒 (< 1秒目标 ✅)
- **API响应时间**: 0.01-0.6秒 (< 1秒目标 ✅)  
- **CDC同步延迟**: < 300ms (企业级标准 ✅)
- **内存使用**: ~23MB (优化后)

### 📋 生成文档
- `e2e-coverage-report.md`: 详细的E2E测试覆盖率报告
- 更新 `CLAUDE.md`: 项目最新状态和E2E测试成果
- 更新 `README.md`: 版本信息和测试验收结果

### 🎉 质量保证达成
| 质量指标 | 目标 | 实际 | 状态 |
|---------|------|------|------|
| E2E覆盖率 | ≥90% | 92% | ✅ 达标 |
| 页面响应时间 | <1秒 | 0.5-0.9秒 | ✅ 优秀 |  
| API响应时间 | <1秒 | 0.01-0.6秒 | ✅ 优秀 |
| 跨浏览器兼容 | Chrome+Firefox | ✅ 支持 | ✅ 达标 |

## 🚀 下一步计划
- [ ] 提交所有变更到Git仓库
- [ ] 标记v1.1-E2E版本Tag  
- [ ] 准备生产环境部署计划
- [ ] 增强压力测试(可选)

---
**执行时间**: 2025-08-10 12:00  
**执行环境**: WSL2 + Docker + 完整CQRS服务栈  
**验证状态**: ✅ E2E测试体系完整，生产环境部署就绪
## 2025-11-15 – Plan 252 权限一致性与契约对齐（完成）
- 新增：权限契约校验器（scripts/quality/auth-permission-contract-validator.js），生成并校验
  - OpenAPI security scopes：路径引用→注册表一致性（未注册引用=0 作为门禁）
  - GraphQL “Permissions Required: …” 注释→Query→scope 映射（SSoT 衍生物）
  - Resolver 授权覆盖（绕过=0 作为门禁；不在 schema 的历史查询仅信息提示）
  - 报告落盘至 reports/permissions/*；证据快照 logs/plan252/*
- PBAC 对齐：查询服务 PBAC 消费生成映射（go:embed）
  - 修正与补齐映射：hierarchyStatistics、assignments、assignmentHistory、assignmentStats、positionAssignmentAudit、jobCatalog*
  - 保留 org:write 临时兼容（TODO‑TEMPORARY，2025‑12‑15 回收）
- OpenAPI scopes 注册补齐：position:assignments:read/write/audit
- CI 门禁：新增 Plan 252 守卫与 DEV_MODE 默认禁用检查
- DEV：查询服务 DEV_MODE 默认 false；开发容器显式启用 DEV_MODE=true
- 测试：新增 PBAC 映射单测（cmd/hrms-server/query/internal/auth/pbac_mapping_test.go）
## [Unreleased]
- feat(gate-255): add ESLint architecture guard to plan-255 workflow (CQRS/ports/contracts), complementing static scanner
- chore(gate-255): pin golangci-lint to v1.59.1 for reproducible CI
- docs(plan-255): clarify JWKS not a permanent frontend exception; provide DEV_MODE temporary strategy; unify status fields to status/isCurrent/isFuture/isTemporal; require plan-250/253 as protected-branch checks alongside 255
- docs(215): add protected-branch evidence template to include plan-250/253/255 and sample failure links
- docs(agents): add “决策建议原则” — for decision items/open questions, provide best-practice advice and a default path (with rollback window), not Q&A only
- refactor(health-alerting): migrate JSON tags to camelCase (resolvedAt, maxRetries, enabledBy, statusEquals, responseTimeGt, consecutiveFails); remove temporary nolints
  - Note: webhook consumers expecting snake_case should align within one iteration; no inbound JSON parser is impacted by this change
