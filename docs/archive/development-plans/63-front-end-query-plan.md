# 63号文档：前端 API / Hooks / 配置整治计划（Phase 3）

> **归档说明**：本计划已于 2025-10-12 完成验收，移至 `docs/archive/development-plans/` 供历史参考。所有执行细节与结项记录请参考 06、60、64 号文档。

**版本**: v1.0
**创建日期**: 2025-10-10
**维护团队**: 全栈工程师（单人执行）
**状态**: 执行完成（待结项评审）
**遵循原则**: CLAUDE.md 资源唯一性与跨层一致性原则
**关联计划**: 60号总体计划、61号执行计划第一阶段验收、62号运维巩固计划

---

## 1. 背景与目标

### 1.1 背景
- Phase 1 已完成契约与类型统一，Phase 2 聚焦后端观测与运维即将收尾。
- 前端层仍存在 **状态管理不一致**、**Hooks 重复实现**、**环境配置分散** 等问题，影响开发效率与可维护性。
- 61号文档第三阶段要求统一前端 API 客户端、清理 Hooks、重构配置。需在已完成的契约成果基础上，确保前端消费契约的一致性与可观测性。

### 1.2 目标
1. **统一 React Query 客户端**：提供共享的 queryClient 与错误处理包装，替换零散实现。
2. **重构 Hooks**：按“查询先、写操作后”顺序迁移现有 Hooks，保持契约类型与新客户端一致。
3. **配置整治**：重写端口/环境助手，沉淀环境变量与运行时配置，提升 QA 与运维的可用性。
4. **质量验证**：通过复用的 Vitest/Playwright 测试与 bundle 分析，确保性能与稳定性达标。

---

## 2. 范围

### 2.1 主要目录
- `frontend/src/shared/api/`
- `frontend/src/shared/hooks/`
- `frontend/src/shared/config/`
- `frontend/src/features/organizations/`（调用方验证）
- `frontend/tests/`、`frontend/src/**/*/__tests__`

### 2.2 非范围内容
- 契约脚本与生成器（Phase 1 已完成）。
- 后端服务改动（Phase 2 专注后端）。
- Playwright 全量脚本重写（本阶段聚焦冒烟场景）。

---

## 3. 时间线（预估 3 周）

| 周次 | 里程碑 | 交付物 |
|------|--------|--------|
| Week 6 | 统一 React Query 客户端与错误包装 | `shared/api/queryClient.ts`, 错误包装工具, 单元测试 |
| Week 7 | Hooks 迁移与桥接层 | 重构后的 Hooks (`useOrganizationsQuery` 等)、`legacyOrganizationApi` 桥接层 |
| Week 8 | 配置与 QA 验证 | 新的端口/环境助手、QA 冒烟报告、Vitest/Playwright 结果、bundle 分析 |

---

## 4. 详细任务清单

### 4.1 Week 6：统一 API 客户端与错误包装
- [x] 梳理现有 API 调用入口，确认保留/废弃路径。
- [x] 实现共享的 queryClient（含缓存策略、重试策略）。
- [x] 编写标准化错误包装（含 requestId、错误码、用户友好提示）。
- [x] 编写单元测试覆盖 queryClient 与错误包装。
- [x] 更新调用示例与文档（README / 计划记录）。*2025-10-12：frontend/README.md 新增统一 Hook 与客户端使用示例*

### 4.2 Week 7：Hooks 迁移与桥接
- [x] 逐一迁移组织相关 Hooks（优先查询类：`useOrganizationsQuery`、`useOrganizationQuery`）。
- [x] 重构写操作 Hooks（创建、更新、停用/启用）以复用新客户端与错误包装。
- [x] 提供 `legacyOrganizationApi` 桥接层（如需），确保旧调用方逐步迁移。（评估结果：当前调用方已整体切换，暂不新增桥接层）
- [x] 更新单元测试/集成测试覆盖迁移的 Hooks。

### 4.3 Week 8：配置整治与验证
- [x] 重写端口/环境助手（`frontend/src/shared/config/`），明确环境变量读取、默认值与校验。
- [x] 整理 QA 冒烟场景（Playwright 在 `frontend/tests/`），确保关键路径通过。（2025-10-11：`npm run test:e2e:smoke` 全部通过，报告已归档）
- [x] 运行 Vitest 覆盖率，目标 ≥ 75%（记录报告）。*2025-10-11：语句 84.1%，范围限定在 Phase 3 模块*
- [x] 分析 bundle 大小，确保与基线相比下降 ≥ 5% 或给出优化说明。*2025-10-12：dist 主包 gzip≈82.97 kB，保持较 10 月初基线的 ≥5% 优化幅度*
- [x] 更新文档（配置指南、运行手册）。*2025-10-12：同步 06 号常见故障/TODO 与 03-API-AND-TOOLS-GUIDE 代理说明*

---

## 5. 验收标准

1. **API 客户端统一**：所有前端 API 调用通过 `shared/api/queryClient.ts` 或统一错误包装。
2. **Hooks 重构完成**：核心组织 Hooks 迁移到新客户端；保留桥接层时需注明使用期限。
3. **测试与质量指标**：
   - Vitest 覆盖率 ≥ 75%
   - Playwright 冒烟场景通过
   - Bundle 大小下降 ≥ 5%（若无法达到，提供原因与替代指标）
4. **配置文档更新**：环境变量、端口、QA 运行指南都记录在案。

---

## 6. 风险与缓解

当前阶段无新增开放风险；既有风险（回归、Bundle 优化、配置一致性）已在 2025-10-12 前全部完成并通过 06/64 号文档的运行验证。

---

## 7. 交付物
- 统一 queryClient / 错误包装代码与测试
- 重构后的 Hooks 及桥接层（如有）
- 新的环境配置工具与文档
- QA 冒烟报告、Vitest 覆盖率报告、bundle 分析
- Phase 3 验收报告草稿（64号文档 v0.2：`docs/development-plans/64-phase-3-acceptance-draft.md`）

---

## 8. 追踪
- 计划文档：63号（本文件）。
- 执行跟踪：`docs/development-plans/60-execution-tracker.md` 第三阶段进度。
- 验收报告：预留 64 号文档。

---

## 9. 附录
- 参考文档：Phase 1、Phase 2 验收记录；`frontend/src/shared/api/`、`frontend/src/shared/hooks/` 现状清单。
- 相关脚本：`frontend/scripts/validate-field-naming*.js`、`frontend/tests/`。
- 配置说明：待第三阶段执行时补充。

---

## 10. 结项说明
- 2025-10-12：全部交付物已完成并在 06/60/64 号文档登记验证结果。
- Phase 3 验收草案（64 号文档 v0.2）已具备评审条件，待确认后即可归档本计划。
