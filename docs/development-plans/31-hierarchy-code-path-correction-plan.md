# 31号文档：组织层级路径编码一致性修复方案

## 背景与单一事实来源
- 命令服务层级刷新逻辑位于 `cmd/organization-command-service/internal/repository/hierarchy.go`，权威字段为 `code_path`/`name_path`；遗留字段 `path` 仍保留在部分查询中，存在历史数据残留风险。
- 前端层级路径展示逻辑 `frontend/src/features/temporal/components/inlineNewVersionForm/utils.ts` 中的 `deriveCodePath` 会优先读取版本数据的 `path`，仅在缺失时才回退至 GraphQL 返回的 `codePath`，导致 UI 受旧字段影响。
- 现网反馈：将组织 `1000003` 重挂至 `1000002` 后，其子组织 `1000004` 的“组织路径（名称）”已按 `name_path` 更新为 `/高谷集团/市场部/国际业务部`，而“组织路径（编码）”仍显示旧值 `/1000000/1000003/1000004`，反映出跨层对 `path` 的依赖造成一致性偏差。

## 目标与范围
- 统一依赖 `code_path`/`name_path` 作为唯一事实来源，全面替换 UI 与服务响应中对 `path` 字段的读取。
- 更新 GraphQL 契约、服务实现与 DTO，逐步移除遗留 `path` 字段，避免重复事实来源。
- 对现有数据执行层级一致性扫描（`sql/hierarchy-consistency-check.sql`），清理残留数据并在迁移完成后考虑数据库列级淘汰。

## 风险评估
- 修改层级刷新 SQL 需确认不会破坏历史时态记录或触发器逻辑，应在测试环境验证大规模级联更新场景。
- 前端优先使用 `codePath` 可能暴露尚未迁移的数据，需要在发布前完成数据库修复与数据校验。
- 批量修正 `path` 字段需在事务中执行，评估对大规模组织树的锁表时间。

## 实施步骤
1. **契约与模型更新**  
   - 在 `docs/api/schema.graphql` 中将 `Organization` 类型的 `path` 字段标记为废弃并准备移除，同时补充 `codePath`/`namePath` 至主模型，确保契约与 OpenAPI 对齐。  
   - 更新 `cmd/organization-command-service/internal/types` 与 Response DTO，仅透出 `code_path`/`name_path`，逐步删除 `path` 字段引用。
2. **后端实现调整**  
   - 清理命令服务、查询服务中对 `path` 字段的读写逻辑，统一走 `code_path` 与 `name_path`；必要时保留兼容层映射，但默认不再返回 `path`。  
   - 对 GraphQL 解析与仓储查询（如 `organizationVersions`）移除 `path` 字段映射，更新测试用例。
3. **前端替换**  
   - 更新 `deriveCodePath` 及相关组件，直接使用 GraphQL 返回的 `codePath`，移除对 `path` 的引用。  
   - 清理类型定义（`InlineVersionRecord` 等）中的 `path` 字段，运行 `npm run lint`/单测确保通过。
4. **数据一致性治理**  
   - 运行 `sql/hierarchy-consistency-check.sql`，确认历史 `path` 与 `code_path` 的差异范围。  
   - 编写一次性数据脚本或迁移，确保在彻底删除 `path` 列前，所有数据均与 `code_path` 对齐。  
   - 评估并在后续迁移中安全移除 `path` 列（需另立迁移脚本和变更单）。
5. **回归验证与发布**  
   - 在测试环境复现重挂场景，验证 API 响应与 UI 展示均使用 `codePath`。  
   - 通过 `make test`、`npm run lint` 等质量门禁，发布后监控日志与前端异常。

## 验收标准
- [ ] GraphQL 契约更新并发布，`Organization` 类型默认提供 `codePath`/`namePath`，`path` 标记废弃或移除，客户端同步升级。
- [ ] 命令服务与查询服务响应不再暴露 `path` 字段，相关处理逻辑统一依赖 `code_path`/`name_path`。
- [ ] 前端“组织路径（编码）”展示 `/1000002/1000003/1000004` 且无 `path` 字段依赖，回归测试通过。
- [ ] 层级一致性扫描通过；如需保留 `path` 列用于历史迁移，确认数据已与 `code_path` 完全一致并有移除计划。
- [ ] Go/前端质量门禁（`make test`、`npm run lint` 等）全部通过，发布与监控记录齐备。

## 一致性校验说明
- 契约字段以 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 为唯一事实来源，迁移后以 `codePath`/`namePath` 为唯一层级路径字段。
- 数据修复脚本与执行日志需引用唯一事实来源（数据库迁移或运维记录），避免产生新的平行事实。

## 现状记录
- 2025-10-09：收到 1000003 → 1000002 重挂后路径不一致问题反馈；定位命令服务与前端逻辑，方案立项。
