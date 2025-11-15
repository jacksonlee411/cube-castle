# HRMS 文档体系索引

**更新时间**: 2025-11-15
**版本**: v1.2

## 概述

本索引聚焦两条主线：
- 203/204/205/206 主线：总体规划、路线与过渡、与 200/201 的对齐。
- 24x 主线（Temporal Entity 命名/框架/页面/E2E）：以 242 为轴的命名与框架治理，贯穿页面重构与测试资产统一。

严格遵循“资源唯一性与跨层一致性”：本索引仅提供导读与执行顺序，不重复落地文档的实现细节。

---

## 203-206 主线（概览）

```
203 ← 主计划（模块划分、架构设计）
 ├── 204 ← 实施路线图（时间表、里程碑）
 ├── 205 ← 过渡方案（具体操作步骤）
 └── 206 ← 对齐分析（与200、201一致性校验）
```

### 203：HRMS 系统模块化演进与领域划分（主计划）

文件: `203-hrms-module-division-plan.md`

- 章节 1-3: 模块蓝图与支撑能力；章节 4: 实施策略与项目结构（含数据访问层演进 4.5）；章节 5: 契约管理（含权限策略 5.4）；章节 6-10 与附录 A-E：优先级、测试、部署、连接池、迁移治理与对齐矩阵
- 对齐度：从 60% 提升至 95%+
- 引用：发件箱（200:341-399）、数据访问（200:207-241）、权限（200:403-417）、连接池（200:261-270）、迁移（200:243-257）

### 204：实施路线图

文件: `204-HRMS-Implementation-Roadmap.md`

- 四阶段节奏与验收要素（行动项、依赖、成果、风险）
- 与 203 第 7 章“过渡方案”对应

### 205：过渡方案

文件: `../archive/development-plans/205-HRMS-Transition-Plan.md`（已归档）

- 多服务/多 go.mod → 模块统一化、组织化结构、workforce/contract 渐进实施
- 与 203 第 7.2-7.4 对齐（操作步骤级别）

### 206：与 200/201 对齐分析

文件: `206-Alignment-With-200-201.md`

 - 关键差异审查、补充实施要求、修订清单与对齐矩阵（建议均已并入 203 v2.1）

---

## 24x 主线（Temporal Entity 命名/框架/页面/E2E）

说明：24x 文档围绕“Temporal Entity”中性抽象推进，从命名统一（242）→ 页面入口抽象（243）→ 时间线/状态抽象（244）→ 类型与契约统一（245/245A/245T）→ 选择器/fixtures 统一（246）→ 文档与治理对齐（247）。

### 242：时态实体命名抽象与统一治理
文件: `242-temporal-naming-abstraction-plan.md`（状态：草案·T0/T1/T2/T3/T4/T5 任务集合）

### 243：Temporal Entity Page 抽象实施
文件: `243-temporal-entity-page-plan.md`（状态：已完成，2025-11-10）

### 244：Timeline & Status 抽象
文件: `244-temporal-timeline-status-plan.md`（状态：进行中；核心改造已落地，E2E/CI 验收仍需收尾）

### 245：类型 & 契约统一（含统一 Hook）
文件: `245-temporal-entity-type-contract-plan.md`（状态：已完成，2025-11-14）

### 245A：统一 Hook 渐进采纳（组织侧）
文件: `245A-unified-hook-adoption.md`（状态：已完成，2025-11-14）

### 245T：OpenAPI no-$ref-siblings 修复
文件: `245T-openapi-no-ref-siblings-fix.md`（状态：已完成，2025-11-14）

### 246：选择器与 Fixtures 统一
文件: `246-temporal-entity-selectors-fixtures-plan.md`（状态：已完成，2025-11-14，Phase 1；组件 testid 收敛 Phase 2 持续推进）

### 247：文档与治理对齐（本次完成项）
文件: `../archive/development-plans/247-temporal-entity-docs-alignment-plan.md`（状态：已完成，已归档；按本索引更新）

### 240：职位管理页面重构与稳定化（已完成）
文件: `240-position-management-page-refactor.md`（状态：已完成 · 验收通过；A–D/BT 完成，E 回归与守卫已登记）

### 241：前端页面与框架一体化重构（已完成）
文件: `241-frontend-framework-refactor.md`（状态：已完成；241C 连跑证据已登记）
- 子计划：  
  - 241A – TemporalEntityLayout 合流与最小接入：`241A-temporal-entity-layout-integration.md`  
  - 241B – 统一 Hook 薄封装与选择器门禁：`241B-unified-hook-and-selector-guard.md`  
  - 241C – E2E 验收与可观测性证据登记：`241C-e2e-acceptance-and-observability-evidence.md`

---

## 执行与监控（活跃）

- 215：Phase2 执行日志与进度
  - 文件: `215-phase2-execution-log.md`（状态：活跃；记录里程碑、验证与证据路径）
- 215：Phase2 概览与里程碑
  - 文件: `215-phase2-summary-overview.md`（状态：活跃；阶段目标与达成情况）
- 06：集成团队协作进展
  - 文件: `06-integrated-teams-progress-log.md`（状态：活跃；跨团队对齐与阻塞）

备注：所有执行与验收证据应登记至 215（必要时在 06 中同步公告），避免出现第二事实来源。

---

## 专项与质量（活跃/记录）

- 220：模块开发模板与规范文档
  - 文件: `220-module-template-documentation.md`（状态：已完成）
- 300：平台化 UI 蓝图（插槽/清单/元数据生成/脚手架/远程功能包）
  - 文件: `300-platformized-ui-blueprint.md`（状态：提案；包含 P0/P1/P2 路线与验收标准）
- 301：字段可扩展标准（标准化装配能力）
  - 文件: `301-field-extensibility-standard.md`（状态：提案；字段复制/生成链/门禁/回归标准）
- 221：Docker 集成测试基座
  - 文件: `221-docker-integration-testing.md`（状态：已完成（本地验收）；CI 常态化推进）
- 221T：Docker 集成测试基座验证记录
  - 文件: `221t-docker-integration-validation.md`（状态：验证记录/滚动维护）
- 222：organization 模块 Phase2 验收
  - 文件: `222-organization-verification.md`（状态：进行中；最终验收）
- 231：Outbox Dispatcher 接入差距分析
  - 文件: `231-outbox-dispatcher-gap.md`（状态：记录/整改跟踪）
- 232：Playwright P0 稳定专项
  - 文件: `232-playwright-p0-stabilization.md`（状态：进行中）
- 232T：Playwright P0 测试清单
  - 文件: `232t-test-checklist.md`（状态：清单/滚动维护）
- 233：审计历史页签失败调查
  - 文件: `233-audit-history-tab-failure.md`（状态：进行中）
- 234：触发器清理评估与验证
  - 文件: `234-trigger-cleanup-assessment.md`（状态：进行中）
- 234T：触发器验证补充记录
  - 文件: `234t.md`（状态：记录/补充）

说明：本分区列出与质量保障、集成验证、缺陷分析相关的活跃专项，作为“执行顺序”之外的并行工作面。详细步骤与证据以各计划文档为唯一事实来源。

---

## 执行顺序（247 完成后）

基于各计划的依赖标注与当前完成度，后续执行顺序建议如下（仅列顺序与准入条件，具体实施细节以各计划文档为唯一事实来源）：

1) 244 收尾验收（必要门槛）
   - 准入条件：`TemporalEntityTimelineAdapter/StatusMeta` 已合并；补齐 E2E（Chromium/Firefox 各 3 轮）与 lint/守卫门禁；在 `215-phase2-execution-log.md` 登记完成状态。
2) 241 恢复并交付（先于 240）
   - 产出统一 `TemporalEntityLayout`、共享数据 Hook 接口与可观测性基线；为 240 提供落地框架。
3) 240 在新框架上实施
   - 职位页面重构与 DOM/testid 稳定性整改，直接复用 241/242/246 的抽象与选择器；关闭 232/232T 中与职位相关的 P0 不稳定条目。
4) 246 Phase 2（并行小步）
   - 组件内 `data-testid` 值逐步收敛到 `temporal-*`，仅更新映射常量；`guard:selectors-246` 计数持续下降。
5) 245A 后续子任务（并行小步）
   - 组织详情子组件继续增量采纳统一 Hook（Header/Alerts/EditForm/ParentSelector），每步提交附守卫/类型检查/Vitest 证据。

备注：
- 若 244 的 E2E 验收在 CI 侧受限，可不阻塞 241/240 的代码开发，但在合并前必须提供本地/CI 复核证据并在 Plan 215/Plan 06 留痕。
- 任何契约或命名调整必须先更新 `docs/api/schema.graphql`、`docs/api/openapi.yaml` 并运行 `node scripts/generate-implementation-inventory.js`，以防“第二事实来源”。

---

## 近期关键推进（基于 202 与 215 评审）

仅给出顺序与索引，实施细节与验收标准以对应计划文档为唯一事实来源。

- P0 · 222 收口验收与文档更新  
  - 索引：`215-phase2-execution-log.md`、`215-phase2-summary-overview.md`、`222-organization-verification.md`  
  - 说明：补齐单元/集成/REST/GraphQL/E2E/性能与覆盖率证据，更新 README/开发者指南，形成 Phase2 验收报告

- P0 · 202 阶段1：模块化单体合流（不改协议，仅合并进程与共享中间件/连接池）  
  - 索引：`202-CQRS混合架构深度分析与演进建议.md`（“阶段 1: 架构回归”）  
  - 说明：统一端口与健康/metrics/JWT/数据库连接池，保留“命令=REST、查询=GraphQL”，降低运维与一致性成本

- P1 · 221 基座 CI 化与常态运行  
  - 索引：`221-docker-integration-testing.md`  
  - 说明：将本地已通过的 Docker 集成测试基座接入 CI，预拉取镜像、冷启动 < 10s、Goose up/down 循环复跑、覆盖率产物上报

- P1 · 202 阶段2：契约 SSoT 与前端 API Facade  
  - 索引：`202-CQRS混合架构深度分析与演进建议.md`（“阶段 2: 工程优化”）  
  - 说明：以 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 为唯一事实来源，固化生成/校验流水线；前端统一领域 Facade，削减双协议耦合

- P1 · 219E 回归补强（面向 222 覆盖率目标）  
  - 索引：`215-phase2-execution-log.md`（Plan 219 完成登记）、`222-organization-verification.md`、`internal/organization/README.md`  
  - 说明：在组织模块上补齐关键路径与回归脚本，作为模板为后续新模块提供可复制的验收脚本

一致性与登记
- 执行证据一律登记到 `215-phase2-execution-log.md`；跨团队同步到 `06-integrated-teams-progress-log.md`
- 若 217B/220 等计划在子文档与索引状态存在漂移，以 `215-phase2-execution-log.md` 为真源，先更新其状态标记与证据路径

---

## 25x 主线（202 计划分解实施）

- 250：模块化单体合流实施计划（阶段1·核心）
  - 文件: `250-modular-monolith-merge.md`（状态：未启动；合流后单进程/单端口/共享中间件）
- 251：运行时统一与健康/指标整合
  - 文件: `251-runtime-unification-health-metrics.md`（状态：已完成；单体主路径/统一健康与指标/配置来源单一）
- 252：权限一致性与契约对齐（REST x-scopes ↔ GraphQL）
  - 文件: `../archive/development-plans/252-permission-consistency-and-contract-alignment.md`（状态：已完成 · 已归档；签字：`../archive/development-plans/252-signoff-20251115.md`）
- 253：部署与流水线简化（单体优先）
  - 文件: `253-deployment-pipeline-simplification.md`（状态：进行中）
- 254：前端端点与代理收敛（单基址）
  - 文件: `254-frontend-endpoint-and-proxy-consolidation.md`（状态：未启动）
- 256：契约 SSoT 生成流水线（阶段2）
  - 文件: `256-contract-ssot-generation-pipeline.md`（状态：未启动；生成与门禁接线）
- 257：前端领域 API 门面采纳（阶段2）
  - 文件: `257-frontend-domain-api-facade-adoption.md`（状态：未启动）
- 258：契约漂移校验与门禁（阶段2）
  - 文件: `258-contract-drift-validation-gate.md`（状态：未启动）
- 259：协议策略复盘与评估（可选）
  - 文件: `259-protocol-strategy-review.md`（状态：未启动；阶段1+2 完成后评估）

说明：25x 仅分解 202 的实施目标与门禁，不重述技术结论；唯一事实来源为 `202-CQRS混合架构深度分析与演进建议.md`，执行与证据登记以 215 为准。

---

## 文档关系图（扩展）

```
┌─────────────────────────────────────────────────┐
│ 200/201（参考标准与现状）                        │
└───────────────┬─────────────────────────────────┘
                │
                ▼
        ┌───────────────┐
        │ 206 对齐分析   │
        └───────┬───────┘
                │
                ▼
        ┌───────────────┐
        │ 203 主计划     │
        └───────┬───────┘
        ┌───────▼───────┐
        │ 204 路线图     │
        │ 205 过渡方案   │
        └───────────────┘

  24x（Temporal Entity 线）：

        ┌───────────────┐
        │ 242 命名抽象   │
        └──┬────┬────┬──┘
           │    │    │
           ▼    ▼    ▼
        243   244   246（并行/交错，按计划准入）
                 │
                 ▼
                245/245A/245T
                 │
                 ▼
                247 文档/治理对齐
                 │
                 ▼
                241 → 240（在新框架上落地）
```

---

## 快速导航

| 角色 | 优先级 | 建议读物 |
|------|--------|---------|
| 架构师/决策者 | 1 | 203（核心）→ 206（对齐分析）→ 247（治理收口） |
| 项目经理 | 1 | 204（路线图）→ 215（阶段执行日志）→ 247（对齐完成核验） |
| 开发工程师 | 1 | 242/243/244/245（命名/页面/类型）→ 241（框架）→ 240（页面重构） |
| QA/测试 | 1 | 246（selector/fixtures 统一）→ 232/232T（E2E 稳定）→ 244（验收） |
| DevOps | 2 | 203（部署运维章）→ 205（部署清单）→ 247（README/Quick Reference 更新） |

### 关键章节查询

| 内容 | 位置 |
|------|------|
| 模块划分 | 203-第2章 |
| 架构设计 | 203-第4章 |
| API 契约 | 203-第5章 |
| 实施优先级 | 203-第6章 |
| 操作步骤 | 205-详细步骤 |
| 时间表 | 204-四个阶段 |
| 数据访问层演进 | 203-第4.5节 |
| 事务性发件箱 | 203-第4.3.3节 |
| 权限策略 | 203-第5.4节 |
| 数据库迁移 | 203-附录D |
| 对齐评估 | 206-全文 |
| Temporal 命名抽象 | 242-全文 |
| Temporal 页面抽象 | 243-全文 |
| Timeline & 状态抽象 | 244-全文 |
| 类型 & 契约统一 | 245/245A/245T-全文 |
| Selector/Fixture 统一 | 246-全文 |
| 文档与治理对齐 | 247-全文 |

---

## 文档版本历史

### 203-hrms-module-division-plan.md
- v2.1 (2025-11-04): 更新 Phase2 就绪与状态描述，持续与 200/201 对齐
- v2.0 (2025-11-03): 完全对齐 200/201，新增 7 个补充需求，对齐度 60% → 95%+
- v1.0 (2025-11-03): 初始版本

### 204-HRMS-Implementation-Roadmap.md
- v1.0 (2025-11-03): 初始版本（编号已更新为 204）

### 205-HRMS-Transition-Plan.md
- v1.0 (2025-11-03): 初始版本（编号已更新为 205；2025-11-04 已归档）

### 206-Alignment-With-200-201.md
- v1.0 (2025-11-03): 初始版本（编号已更新为 206）

### 24x 文档（新增索引）
- 243-temporal-entity-page-plan.md: 已完成（2025-11-10）
- 244-temporal-timeline-status-plan.md: 进行中（收尾验收中）
- 245-temporal-entity-type-contract-plan.md: 已完成（2025-11-14）
- 245A-unified-hook-adoption.md: 已完成（2025-11-14）
- 245T-openapi-no-ref-siblings-fix.md: 已完成（2025-11-14）
- 246-temporal-entity-selectors-fixtures-plan.md: 已完成（2025-11-14，Phase 1）
- 247-temporal-entity-docs-alignment-plan.md: 已完成（与本索引同步；已归档）
- 240-position-management-page-refactor.md: 暂缓（待 241 完成后执行）
- 241-frontend-framework-refactor.md: 暂缓（247/242 完成后恢复优先）

---

## 重要约定

### 文档编号系统

```
203 系列文档（HRMS 计划）：
├── 203：主计划（模块划分、架构、实施策略）
├── 204：路线图（时间表、里程碑、交付物）
├── 205：过渡方案（具体操作指南）
└── 206：对齐分析（与 200/201 的一致性验证）
```

### 跨文档引用规则

- **203 引用 204/205/206**：通过文件名和章节号
- **204/205 引用 203**：通过"第 X 章"或"附录 X"
- **206 引用 203/204/205**：通过文件名和相关性说明

### 更新维护

- 所有文档在同一次提交中更新，保持一致性
- 编号变更后，相关文档的内部引用自动更新
- 版本号在顶部清晰标注

---

## 实施开始

**推荐的文档阅读顺序**:

1. **快速概览** (30 分钟)
   - 203-第 1-2 章
   - 206-执行摘要

2. **深入理解** (2 小时)
   - 203-第 3-6 章
   - 204-总体时间表

3. **实施准备** (3 小时)
   - 205-第一、二阶段详细步骤
   - 203-第 7 章和第 10 章

4. **长期规划** (1 小时)
   - 203-附录 C、D、E
   - 206-第二、三部分

—

统一约束与登记
- 执行与验收证据：请统一登记至 `docs/development-plans/215-phase2-execution-log.md`；跨团队同步请使用 `docs/development-plans/06-integrated-teams-progress-log.md`。
- 契约先行：任何契约或命名调整必须先更新 `docs/api/schema.graphql`、`docs/api/openapi.yaml` 并运行 `node scripts/generate-implementation-inventory.js`，防止出现“第二事实来源”。

---

**本索引最后更新**: 2025-11-15
**维护者**: 架构评审组
