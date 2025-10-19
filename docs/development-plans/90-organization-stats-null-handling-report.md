# 90号文档：组织列表 GraphQL 错误分析与修复建议

**更新时间**：2025-10-19 15:20
**触发来源**：前端组织架构管理页（`OrganizationDashboard`）显示“数据加载失败 API Error: Query completed with errors (请求ID: req_⋯)”
**影响范围**：组织单元列表、统计卡片、分页器；所有依赖 `organizations` 查询的页面都会受到影响

---

## 1. 问题概述

- 用户打开“组织架构管理 → 组织单元列表”时，总是收到“数据加载失败 API Error: Query completed with errors”提示，页面无法加载任何表格数据。
- 错误来自 GraphQL 查询封装层：当后端 GraphQL 返回 `errors` 数组时，`internal/middleware/graphql_envelope.go` 会统一转换成企业级错误响应，并写回 `"Query completed with errors"`（详见该文件第 62-108 行）。
- 前端统一客户端 `frontend/src/shared/api/unified-client.ts` 在检测到企业级错误后抛出异常，`useEnterpriseOrganizations` 钩子捕获失败并将错误透传给 `OrganizationDashboard`，进而渲染“数据加载失败”提示。

---

## 2. 复现路径

1. 启动 Docker 基础设施（`make docker-up`）、查询服务与前端。
2. 浏览器访问 `http://localhost:3000/organizations`。
3. 页面加载“组织单元列表”时出现上述错误弹窗，同时开发者工具 Network 面板可观察到 GraphQL 请求返回 `success: false` 与 `message: "Query completed with errors"`。

---

## 3. 调查发现

- **查询形态**：前端 `useEnterpriseOrganizations` 发送的 GraphQL 文档（同文件第 66-123 行）一次性请求 `organizations` 与 `organizationStats` 两个字段；只要其中任意字段出错，GraphQL runtime 会在响应 `errors` 中记录失败并继续返回部分数据。
- **错误拦截行为**：GraphQL Envelope 中间件会检测到 `errors` 列表非空，构造 `types.WriteErrorResponse("GRAPHQL_EXECUTION_ERROR", "Query completed with errors", …)` 并写回企业级错误体，从而让整次调用视为失败。
- **后端统计实现缺陷**：`cmd/organization-query-service/internal/repository/postgres_organization_hierarchy.go` 中 `GetOrganizationStats` 通过一次 SQL 聚合查询获取统计；当租户下没有满足 `status <> 'DELETED'` 条件的记录时，`MIN(effective_date)` / `MAX(effective_date)` 会返回 `NULL`，但当前代码仍尝试将结果扫描进 `time.Time` 类型（第 70-90 行），触发 `sql: Scan error on column index …`。
- **错误传播链路**：上述扫描错误向上传递 → resolver 返回 `error` → GraphQL runtime 写入 `errors` → Envelope 将其转为“Query completed with errors” → 前端统一客户端抛出 `API Error`，最终导致页面渲染失败。

---

## 4. 根因分析

> `GetOrganizationStats` 未对空数据集进行空值兜底，直接把 `NULL` 写入非空 Go 类型，导致 SQL 扫描失败。

- 在新环境或清理后的租户中，如果所有组织均被软删除或尚未创建数据，就会出现无满足条件记录的情况。
- 由于 `organizationStats` 字段声明为非空（`docs/api/schema.graphql` 第 356-410 行），任一错误都必须被视为 P0：统计缺失无法回退到前端默认值，同时阻断列表数据。

---

## 5. 修复建议

1. **后端防御性改造（首要）**
   - 将 `oldest_date` / `newest_date` 扫描目标改为 `sql.NullTime` 或在 SQL 中使用 `COALESCE(MIN(...), NOW())`，再格式化为字符串，以满足 GraphQL 非空字段约束。
   - 在统计结果为空时，明确返回 `TemporalStats{totalVersions: 0, averageVersionsPerOrg: 0, oldestEffectiveDate: "1970-01-01", newestEffectiveDate: "1970-01-01"}` 或其它契约认可的默认值，避免抛错。
   - 新增针对“无数据租户”场景的单元测试，覆盖 `GetOrganizationStats` 与 Resolver，使回归时自动阻止类似问题。

2. **GraphQL 查询健壮性（并行推进）**
   - 在 `useEnterpriseOrganizations` 里，将统计查询拆分为独立请求或允许失败降级（例如忽略统计模块，仅提示“统计暂不可用”），避免界面被整体阻断。

3. **监控与告警**
   - 在查询服务日志中增加 `GetOrganizationStats` 错误计数，并将“连续失败”接入现有质量门禁，防止问题再次被忽略。

---

## 6. 验收标准

- 租户下无任何“未删除”组织时，GraphQL `organizationStats` 正常返回 0 值统计，`organizations` 列表可正常展示。
- 浏览器刷新组织列表不再出现 “Query completed with errors”，统计面板应显示默认数值或“暂无数据”。
- 新增的单元测试覆盖“空数据”场景并在 CI 中强制执行。

---

## 7. 后续行动

| 任务 | 责任人 | 截止时间 | 备注 |
| --- | --- | --- | --- |
| 调整 `GetOrganizationStats` Null 处理并补充测试 | 查询服务团队 | 2025-10-21 | 遵循 CLAUDE.md “根因修复” 原则 |
| 评估前端统计降级策略并更新 UX 提示 | 前端团队 | 2025-10-22 | 与 `useEnterpriseOrganizations` 对齐 |
| 在 `docs/development-plans/06-integrated-teams-progress-log.md` 更新状态 | 架构组 | 完成后次日 | 保持跨文档一致性 |

---

> **合规说明**：分析与建议均依据当前唯一事实来源文件（`docs/api/schema.graphql`、`frontend/src/shared/hooks/useEnterpriseOrganizations.ts`、`cmd/organization-query-service/internal/repository/postgres_organization_hierarchy.go`、`internal/middleware/graphql_envelope.go`），符合“资源唯一性与跨层一致性”约束。

