## 变更说明

- 概述本次变更的目的与范围：
- 是否涉及后向不兼容变更：
- 验证方式与可复现实验步骤：

## 对齐 CLAUDE.md 与 AGENTS.md 的合规检查

- [ ] 诚实/谨慎：描述基于可验证事实，无夸大用语
- [ ] 临时方案：如有 `// TODO-TEMPORARY:`，已包含原因、计划、截止日期(YYYY-MM-DD)
- [ ] 新功能审批：新增 API/组件/服务/表 已获用户批准（附链接/证据）
- [ ] CQRS 分工：查询走 GraphQL，命令走 REST；未引入额外数据源
- [ ] API 命名：字段均为 camelCase；组织路由路径参数使用 `{code}`（未使用 `/{id}`）
- [ ] 契约命名自查：前端新增/修改字段均遵循 camelCase；无 `*_status`/`*_date` 等 snake_case 遗留
- [ ] 权限契约：相关端点权限已在 `docs/api/openapi.yaml` 定义并与实现一致
- [ ] 测试与验证：包含必要单测/集成/契约测试或最小验证
- [ ] 文档更新：READMEs/规范/变更日志已同步

## 文档治理与目录边界（Reference vs Plans）

- [ ] 已在 `docs/development-plans/` 建立/更新本次变更的计划/进展（完成后将归档至 `archived/`）
- [ ] 未将计划/进展类文档置于 `docs/reference/`（reference 仅保留长期稳定参考）
- [ ] 如新增参考类文档，已确认其长期稳定性并与 plans 分离
- [ ] `docs/README.md` 与各目录 `00-README.md` 的导航与边界说明保持一致

## 风险与回滚

- 风险点：
- 监控/告警：
- 回滚方案：
