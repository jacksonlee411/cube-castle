# 30号文档：审计日志 operation_reason 字段一致性修复计划

## 背景与单一事实来源
- 命令服务运行日志 `/tmp/run-dev.log` 在执行激活操作 (`POST /api/v1/organization-units/1000004/activate`) 时持续出现 `pq: column "operation_reason" of relation "audit_logs" does not exist` 的数据库错误，说明审计写入逻辑试图持久化该字段。
- 现有审计表结构可通过 `SELECT column_name FROM information_schema.columns WHERE table_name='audit_logs';` 验证，仅包含 `modified_fields` 等 22 个字段，缺少 `operation_reason` 列。
- 审计写入逻辑位于 `cmd/organization-command-service/internal/audit/logger.go`，调用 `LogOrganizationUpdate` / `LogOrganizationSuspend` 等方法时会传入 `operationReason`，需要数据库结构配合。

## 目标与范围
- 为 `audit_logs` 表补充缺失的 `operation_reason` 字段，并保证迁移按序执行，不破坏既有数据。
- 调整审计写入逻辑/DTO 以使用唯一事实来源，确保前端展示与 API 响应保持一致。
- 覆盖命令服务在激活、停用、更新等路径触发的审计事件，确认不再抛出错误。

## 风险评估
- 变更审计表结构需评估历史数据兼容性，避免影响现网 BI / 报表查询。
- 若迁移未同步到所有环境，会导致旧服务继续写旧结构，引发列不存在或数据丢失风险。
- 与 GraphQL 查询审计历史 (`docs/api/schema.graphql` 中 `auditHistory` 定义) 相关字段需确认是否也同步扩充。
- ALTER 表时需要评估 `audit_logs` 数据量，避免长时间锁表影响线上写入。

## 实施步骤
1. 设计数据库迁移：在 `database/migrations/` 添加顺序迁移脚本，创建 `operation_reason TEXT` 列，并针对历史数据提供合理默认值（可为空）。
2. 更新审计模型与写入逻辑：在 `cmd/organization-command-service/internal/audit/logger.go` 等位置，确保 `OperationReason` 字段持久化并在结构体/DTO 中暴露。
3. 校验 API/GraphQL 契约：查看 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 是否已有字段定义；若缺失需同步补充并更新前端消费逻辑。
4. 执行回归测试：通过 `npm run lint`、相关 Go 单元测试与手动调用激活/停用接口，确认日志不再输出列缺失错误。
5. 产出迁移执行记录：在 `docs/development-plans/` 更新计划状态，并在 `docs/archive/development-plans/` 归档完成报告。

## 验收标准
- [x] 新迁移脚本执行后，`audit_logs` 表包含 `operation_reason` 列，历史数据保持可查询。
- [x] 再次执行组织激活/停用接口，命令服务日志不再出现列缺失错误，返回 200 且審计记录写入成功。
- [x] GraphQL `auditHistory` 与 REST 审计接口（若有）能透出 `operationReason` 字段，前端展示正常。
- [x] 相关单元与集成测试通过，`npm run lint`、`make test` 或等效命令无错误。

## 一致性校验说明
- 数据库结构以最新迁移脚本为唯一事实来源，所有服务需通过迁移保持一致。
- 审计字段命名遵循 camelCase，与 `docs/api/openapi.yaml` / `docs/api/schema.graphql` 对齐；若存在差异需先更新契约。

## 现状记录
- 2025-10-09：本地开发环境在命令服务激活请求过程中确认 `audit_logs.operation_reason` 缺失，计划立项。
- 2025-10-09：新增 039/040 数据库迁移，补齐 `operation_reason` 列并清理 JSON 默认值；更新命令/查询服务以返回 `operationReason`。
