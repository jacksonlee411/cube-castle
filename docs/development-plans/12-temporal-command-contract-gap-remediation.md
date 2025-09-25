# 12. 时态命令契约回正计划

**文档类型**: 契约修复 / 去重专项  
**创建日期**: 2025-09-24  
**优先级**: P0（唯一性失守 + 生产阻断风险）  
**负责团队**: 命令服务团队（Owner） / 前端组织域团队（Co-owner）  
**关联文档**: `CLAUDE.md`、`AGENTS.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`docs/api/openapi.yaml`

---

## 1. 背景与触发

- 组织时态功能的合同事实来源始终是 `/api/v1/organization-units/{code}/versions` 等既有 REST 端点，且当前命令服务依旧通过 `cmd/organization-command-service/internal/handlers/organization.go:167` 提供该能力。
- 历史上曾计划引入新的 `/organization-units/temporal` 路由，并准备了 `temporal_handlers.go.disabled` 中的处理器草稿，但重构在契约更新前即被搁置，未进入主干、未获得 OpenAPI 支持。
- 前端在未更新契约的情况下，自行调用 `/organization-units/temporal` / `/organization-units/{code}/temporal`，导致运行时 404 与重复实现，直接违反“先契约后实现”“资源唯一性与跨层一致性为最高优先级”的原则。
- 该重复路径风险已触发 IIG 与质量门禁告警，必须回收半途而废的实现草稿，恢复单一事实来源。

## 2. 问题画像（证据）

| 证据位置 | 问题说明 |
| --- | --- |
| `frontend/src/features/organizations/components/OrganizationForm/index.tsx` & `frontend/src/shared/hooks/useOrganizationMutations.ts:24-72` | 表单与 Hook 新增了 `/temporal` 调用，绕过契约端点。 |
| `docs/api/openapi.yaml` | `rg "organization-units/temporal" docs/api/openapi.yaml` 无匹配结果，说明契约未更新。 |
| `cmd/organization-command-service/internal/handlers/organization.go:167` | 仍以 `/versions`、`/events` 为唯一可执行端点。 |
| `cmd/organization-command-service/internal/handlers/temporal_handlers.go.disabled` | 未编译的草稿处理器，证明重构停滞。 |
| `reports/implementation-inventory.json` | IIG 未登记 `/temporal` 系列，显示事实来源仍指向 `/versions`。 |

## 3. 根因与约束

1. **半途而废的重构计划**：后端草稿未合并却残留在仓库，给调用方造成“新端点可用”的错觉。
2. **契约流程缺失**：前端跳过 `docs/api/openapi.yaml` 变更与评审，违反 `CLAUDE.md` 的契约先行原则。
3. **跨团队同步失败**：IIG 报表和实现清单未发现重复调用，导致前端误判可以试用 `/temporal`。
4. **唯一性原则被破坏**：同一能力出现两个事实来源（契约内 `/versions` vs. 未发布 `/temporal`），直接违背最高优先级指令。

## 4. 目标与验收标准

| 目标 | 验收标准 |
| --- | --- |
| 恢复唯一事实来源 | OpenAPI、命令服务路由、前端调用全部回归 `/api/v1/organization-units/{code}/versions` 等既有端点；IIG 报表仅登记该能力。 |
| 端到端一致 | 前端 payload 字段与契约字段完全一致（`operationReason`、`effectiveDate` 等），通过 Jest/Vitest + Go 集成测试。 |
| 去除遗留草稿 | `temporal_handlers.go.disabled` 删除或存档至 `archive/`，确保仓库不再出现未发布处理器。 |
| 质量门禁恢复 | `node scripts/generate-implementation-inventory.js`、`node scripts/quality/architecture-validator.js` 与 Playwright 时态场景全部绿灯。 |

## 5. 指导原则（强制）

- ⚠️ **最高优先级**：资源唯一性与跨层一致性优先于一切交付，如果任何环节仍依赖 `/temporal`，必须立即阻断合并。
- 契约 → 实现 → 调用的顺序不得倒置；不得在契约之外创建新端点或代理层。
- 所有修复提交需在 PR 描述中说明“如何验证唯一事实来源与契约一致性”。

## 6. 方案评估与决策

| 方案 | 概述 | 结论 |
| --- | --- | --- |
| A. 完成 `/temporal` 重构 | 重新启用草稿处理器，补充契约并迁移调用 | ❌ 与目标相悖，会再次制造双事实来源，增加维护面。 |
| B. 回退至契约端点并清理遗留 | 前端撤回 `/temporal` 调用，后端删除草稿，维持 `/versions` 为唯一入口 | ✅ 直接恢复唯一性与一致性，符合 CLAUDE.md 与 AGENTS.md 要求。 |
| C. 新增 BFF 代理层 | 由 BFF 翻译 `/temporal` 请求至 `/versions` | ❌ 引入额外事实来源与复杂度，风险更高。 |

**决策**：执行方案 B。

## 7. 行动计划

| 阶段 | 任务 | Owner | 截止 |
| --- | --- | --- | --- |
| Phase 1 | 在 IIG 与架构周会上通报 `/temporal` 重复问题，冻结相关 PR；更新本计划至 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 变更日志 | IIG Guardian | 2025-09-25 |
| Phase 1 | 删除前端 `/temporal` 调用：统一使用契约 Hook，修正请求 payload（`operationReason`、`effectiveDate` 为 `YYYY-MM-DD`），补充单测 | Frontend Org Team | 2025-09-27 |
| Phase 1 | 清理后端遗留：移除 `temporal_handlers.go.disabled` 或归档至 `docs/archive/`，确认路由表仅包含契约端点 | Command Service Team | 2025-09-27 |
| Phase 2 | 运行 `make test`、`make test-integration`、`npm --prefix frontend run test`、Playwright 时态场景，更新实现清单与 QA 记录 | 联合 QA | 2025-09-28 |
| Phase 2 | 在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`06-integrated-teams-progress-log.md` 登记结果，关闭本计划 | IIG Guardian | 2025-09-29 |

## 8. 验证步骤

1. `node scripts/generate-implementation-inventory.js`，确认输出无 `/temporal`。  
2. `node scripts/quality/architecture-validator.js`，验证 CQRS 路由与文档一致。  
3. `curl -X POST http://localhost:9090/api/v1/organization-units/{code}/versions` 搭配新 payload，验证 2xx 与审计记录。  
4. Playwright `organization-create.spec.ts` / `temporal-management-integration.spec.ts` 通过。  
5. 记录在 `reports/iig-guardian/` 中，作为唯一事实来源恢复的证据。

## 9. 风险与应对

| 风险 | 影响 | 缓解措施 |
| --- | --- | --- |
| 前端移除 `/temporal` 后仍有隐藏调用 | 用户仍会触发 404 | 通过 `rg "temporal" frontend/` 全量扫描；在 CI 添加 deny list。 |
| 删除草稿影响后续架构讨论 | 历史资料丢失 | 将原草稿存档到 `docs/archive/development-plans/` 或 `archive/deprecated-api-design/`，附带说明。 |
| Payload 格式未一次调整完成 | 契约测试失败 | 使用共享转换工具 + Vitest 覆盖，强制类型校验。 |

## 10. 交付物

- 前端修复 PR（撤销 `/temporal`、统一契约字段、单元测试）。
- 后端清理 PR（删除草稿处理器、确认路由表、更新审计文档）。
- 更新后的 IIG 报表与实现清单条目，明确“时态命令契约回正”状态。  
- QA 验证记录 & Playwright 报告归档至 `reports/`。

---

**后续追踪**：计划关闭后，两周内在 `06-integrated-teams-progress-log.md` 回顾验证状态，确保没有团队重新引入 `/temporal`；若出现类似策略，必须在 PR 模板新增“唯一事实来源确认”段落。
