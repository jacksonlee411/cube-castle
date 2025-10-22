# 109 号方案：职位审计历史缺失整改计划

**版本**：v1.0（执行中）  
**创建日期**：2025-10-22  
**责任团队**：后端查询服务组（主责） · 职位领域前端组（协同）  
**关联需求**：职位详情页展示完整时态版本与审计轨迹  

---

## 1. 背景与触发
- 实际现象：访问 `职位详情：P1000000` 时，`AuditHistorySection` 组件仅显示“暂无审计记录”。
- 前端调用链：`frontend/src/features/positions/PositionTemporalPage.tsx` 将版本 `recordId` 传递给 `AuditHistorySection`，后者通过 `auditHistory` GraphQL 查询获取数据。
- 权威契约：`docs/api/schema.graphql`（v4.6.0）定义 `auditHistory(recordId)` 为记录级审计查询，目前描述集中在组织时态。

## 2. 复现与事实基线
1. **GraphQL 查询返回空数组**  
   ```bash
   curl -s -X POST http://localhost:8090/graphql \
     -H "Authorization: Bearer $(cat .cache/dev.jwt)" \
     -H "X-Tenant-ID: 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9" \
     -d '{"query":"query($id:String!){ auditHistory(recordId:$id){ auditId } }","variables":{"id":"22ebed15-f902-41b6-bdd6-769bfe856832"}}'
   ```
   → 查询日志记录 `返回 0 条记录`。
2. **数据库实际存在审计日志**  
   ```sql
   SELECT resource_type, changes, response_data
     FROM audit_logs
    WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
      AND resource_id::uuid = '22ebed15-f902-41b6-bdd6-769bfe856832';
   ```
   - 返回两行，`resource_type = 'POSITION'`，`changes = []`，`response_data` 包含版本快照。
3. **命令服务写审计的事实来源**  
   - `cmd/organization-command-service/internal/services/position_service.go` 在 `logPositionEvent` 中固定 `ResourceType = "POSITION"`。
4. **查询服务原始实现**  
   - `cmd/organization-query-service/internal/repository/postgres_audit.go` 先前约束 `resource_type = 'ORGANIZATION'`，并在 `processAuditRowsStrict` 中丢弃 `changes`、`modified_fields` 同为空的记录，导致职位审计被过滤。

## 3. 根因分析
| 层级 | 事实来源 | 结论 |
|------|----------|------|
| 审计写入 | 命令服务 `logPositionEvent` | 职位版本创建/更新均以 `POSITION` 类型写入 `audit_logs`。 |
| 数据存储 | `audit_logs` 表 | 存在合法记录，但 `changes` / `modified_fields` 为空数组。 |
| 查询实现 | GraphQL 仓库 `postgres_audit.go` | 仅允许 `resource_type = 'ORGANIZATION'` 并跳过“无字段差异”记录。 |
| 前端调用 | `AuditHistorySection` | 依约定通过 `recordId` 查询，但收到空数组触发“暂无审计记录”。 |

根因总结：查询服务的资源类型过滤与严格校验逻辑未覆盖职位审计；尽管日志已写入数据库，但在 GraphQL 层被直接忽略，造成前端空态。

## 4. 整改方案与执行
| 序号 | 动作 | 实施状态 |
|------|------|----------|
| A | 扩展 `auditHistory` 查询允许 `POSITION` / `JOB_CATALOG` 资源类型（`postgres_audit.go`）。 | ✅ 已提交 |
| B | 为空 `changes` 但存在快照的事件放行（`processAuditRowsStrict` 新增快照检测）。 | ✅ 已提交 |
| C | 更新 GraphQL 契约描述为“支持组织 / 职位 / 职位分类记录”（`docs/api/schema.graphql`）。 | ✅ 已提交 |
| D | 调整前端审计组件文案，确保与契约一致（`AuditHistorySection.tsx`）。 | ✅ 已提交 |
| E | 验证：`go test ./cmd/organization-query-service/...`、`curl` 查询确认至少返回两条记录。 | ✅ 已完成 |

> 注：所有代码改动均遵循唯一事实来源原则，并与 `docs/api/schema.graphql` 同步保持一致。

## 5. 验证与验收标准
- ✅ `go test ./cmd/organization-query-service/...` 通过，覆盖新增逻辑的现有测试套件。
- ✅ GraphQL 查询返回包含 `CREATE` 与 `UPDATE` 事件的数组，`afterData` 字段包含职位快照。
- ✅ 前端刷新职位详情页后，`AuditHistorySection` 显示两条审计记录（需确保 `VITE_POSITIONS_MOCK_MODE=false`）。
- ✅ 手工 SQL 校验 `audit_logs` 表未出现额外重复写入，`resource_type` 字段保持原有值。

## 6. 未决事项与后续跟踪
- ⌛️ 若未来扩展更多资源（如 `POSITION_ASSIGNMENT`），需评估 `auditHistory` 查询的授权范围与数据整形策略。
- ⌛️ 建议在 `cmd/organization-command-service` 审计写入逻辑中补充 `changes` 字段，以便前端展示字段差异。
- ⌛️ 观察期：上线后一周监控 `graphql-service` 日志中 `[PERF] record_id审计查询` 的返回条数及平均耗时，确认无性能退化。

---

**记录人**：后端查询服务组 · 王小松  
**最后更新**：2025-10-22 03:50 UTC  

## 7. 前端验证结果
- ✅ 职位详情页（P1000000）已在 `VITE_POSITIONS_MOCK_MODE=false` 下实时调用 REST+GraphQL 链路，`AuditHistorySection` 展示出两条审计记录（CREATE / UPDATE）。
- 验证环境：本地 dev 环境，命令服务 & 查询服务均通过 `make run-dev` 启动。
- 触发步骤：
  1. 刷新职位详情页并选择任一时态版本；
  2. 审计历史卡片自动拉取 `auditHistory`；
  3. 确认 UI 已包含创建与更新事件，内容与数据库快照一致。

