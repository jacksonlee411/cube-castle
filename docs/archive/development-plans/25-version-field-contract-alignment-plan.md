# 25 号计划：version 字段契约纳管方案

## 1. 背景与问题陈述
- 组织数据表与触发器：`database/migrations/031_cleanup_temporal_triggers.sql` 定义 `organization_version_trigger()`，在每次插入/更新 `organization_units` 时维护 `version` 自增，表明后端已具备乐观锁版本号。  
- 前端现状：`frontend/src/shared/types/organization.ts` 与 `frontend/src/features/temporal/components/TemporalInfoDisplay.tsx` 预留 `version` 字段，用于未来的版本可视化与冲突提示。  
- 契约缺口：`docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 均未暴露 `version`，导致 API 权威来源缺失字段描述，违背“先契约后实现”原则，也让客户端无法依赖正式字段。

## 2. 目标与验收标准
| 目标 | 验收标准 |
| --- | --- |
| 契约补齐 | `docs/api/openapi.yaml` 中 `OrganizationUnit`、相关响应示例新增 `version`，并说明语义与数值来源；`docs/api/schema.graphql` 的 `Organization` 类型新增对应字段。 |
| 实现对齐 | 命令服务与查询服务返回结果实际包含 `version`；GraphQL `organization` / `organizationVersions` 查询可读取整数字段。 |
| 客户端可用 | 前端组织详情页通过契约字段展示版本号，移除临时占位逻辑。 |
| 测试覆盖 | OpenAPI 合约校验、GraphQL schema 校验、E2E 时间轴用例（含版本号断言）全部通过。 |
| 文档同步 | 更新 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 和 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 中关于组织字段的描述，保持单一事实来源。 |

## 3. 工作范围
- **在范围**  
  1. 契约文件更新：OpenAPI、GraphQL、示例响应、错误码说明。  
  2. 后端实现：命令服务 `OrganizationResponse`、查询服务 resolver/SQL 映射补充 `version`，并确保触发器逻辑与返回值一致。  
  3. 客户端：统一改为消费契约字段，整理 `TemporalInfoDisplay` 中版本展示逻辑。  
  4. 回归：契约验证脚本、集成/E2E 测试、实现清单刷新。
- **不在范围**  
  - 与 `ETag` / `If-Match` 乐观锁处理流程相关的新业务规则变更。  
  - 重写数据库触发器（若触发器存在缺陷另行立项）。

## 4. 事实来源与一致性校验
- **唯一事实来源列表**  
  - 数据触发器：`database/migrations/031_cleanup_temporal_triggers.sql`。  
  - REST 契约：`docs/api/openapi.yaml`。  
  - GraphQL 契约：`docs/api/schema.graphql`。  
  - 前端现状：`frontend/src/shared/types/organization.ts`、`frontend/src/features/temporal/components/TemporalInfoDisplay.tsx`。  
  - 实现清单：`docs/reference/02-IMPLEMENTATION-INVENTORY.md`。
- **一致性检查步骤**  
  1. 变更契约 → 运行 `node scripts/generate-implementation-inventory.js` 校验字段登记。  
  2. 变更后端 → 运行 `make test`、`make test-integration`。  
  3. 变更前端 → `cd frontend && npm run test`、`npm run lint`。  
  4. 端到端 → `make e2e-full` 确认版本号在 UI 呈现。  
  5. 提交前 → `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 与契约字段保持同步。

## 5. 阶段划分与任务清单
| 阶段 | 任务 | 负责人 | 输出物 |
| --- | --- | --- | --- |
| Phase 1 契约设计 | (1) 评估 `OrganizationUnit` 响应及 `SuccessResponse` 示例；(2) 补全 GraphQL `Organization` 与 `organizationVersions` 返回字段；(3) 更新错误码文案说明 `version` 用途 | 架构组 | 契约 MR |
| Phase 2 实现对齐 | (1) 更新命令服务 DTO/Mapper；(2) 查询服务 SQL/Resolver 输出版本号；(3) 增补相关单元/集成测试 | 后端组 | Go 代码 + 测试报告 |
| Phase 3 客户端接入 | (1) 清理前端临时字段占位；(2) 补充版本展示与冲突提示；(3) 更新 Vitest/E2E 断言 | 前端组 | Frontend MR + E2E 结果 |
| Phase 4 文档与发布 | (1) 刷新实现清单、API 工具指南；(2) 若需迁移步骤，更新 `CHANGELOG.md` 与团队播报；(3) 计划归档 | 架构组 | 文档更新、计划归档 |

## 6. 风险与缓解措施
| 风险 | 影响 | 缓解措施 |
| --- | --- | --- |
| 契约更新未同步所有调用者 | 造成部分客户端解析失败 | 提前通知所有调用方，并准备 feature flag 如需灰度 |
| 版本号返回类型不一致（int vs string） | 类型错误导致序列化失败 | 契约明确 `integer`，后端统一转换；测试覆盖 |
| 数据中存在旧记录 version null | 返回响应缺失 | 在 Phase 2 加入迁移/后置脚本，将历史 null 填成 1 |
| GraphQL 缓存破坏 | 前端查询未刷新 | 更新后运行 `graphql-codegen`、确认缓存键未变 |

## 7. 时间安排与里程碑（建议）
- T0：立项通过，完成契约草稿（+1 天）。  
- T0 +3 天：后端与契约 MR 合并，通过回归测试。  
- T0 +5 天：前端接入完成，E2E 通过。  
- T0 +6 天：文档与实现清单更新，计划归档。

## 8. 验收清单
1. `docs/api/openapi.yaml` / `docs/api/schema.graphql` 含 `version` 字段及示例。  
2. 命令服务/查询服务响应实际返回 `version`，单元与集成测试覆盖。  
3. 前端组织详情页显示版本号，`TemporalInfoDisplay` 不再依赖占位逻辑。  
4. `node scripts/generate-implementation-inventory.js` 输出字段统计准确，`docs/reference/02-IMPLEMENTATION-INVENTORY.md` 同步更新。  
5. QA 提供回归报告（含 E2E 版本场景截图/trace）。  
6. 计划文档移入 `docs/archive/development-plans/` 并在提交信息中引用。

---
*编写日期：2025-10-09 · 维护人：架构组*
