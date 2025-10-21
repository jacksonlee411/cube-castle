# 81号附录：职位管理 API 权限映射对照表

**版本**: v0.1 草案  
**创建日期**: 2025-10-14  
**维护团队**: 架构组 + 安全团队  
**来源**: `docs/archive/development-plans/80-position-management-with-temporal-tracking.md` 第 6 节、OpenAPI/GraphQL 草拟片段  

---

## 1. REST 命令端点 → 权限 Scope

| 接口 | 说明 | 所需 Scope | 参考文档 |
|------|------|------------|----------|
| `POST /api/v1/positions` | 创建职位 | `position:create` | 80号文档 §4.1.1 |
| `PUT /api/v1/positions/{code}` | 替换职位（完全更新） | `position:update` | 80号文档 §4.1.2 |
| `POST /api/v1/positions/{code}/versions` | 新增职位时态版本 | `position:modify:history` | 80号文档 §4.1.3 |
| `POST /api/v1/positions/{code}/fill` | 填充职位（临时端点） | `position:fill` | 80号文档 §4.1.5 |
| `POST /api/v1/positions/{code}/vacate` | 空缺职位（临时端点） | `position:vacate` | 80号文档 §4.1.5 |
| `POST /api/v1/positions/{code}/transfer` | 转移职位（临时端点） | `position:transfer` | 80号文档 §4.1.6 |
| `POST /api/v1/positions/{code}/suspend` | 暂停职位 | `position:suspend` | 80号文档 §4.1.7 |
| `POST /api/v1/positions/{code}/activate` | 激活职位 | `position:activate` | 80号文档 §4.1.7 |
| `POST /api/v1/positions/{code}/events` | 删除/历史调整 | `position:modify:history` | 80号文档 §4.1.8 |
| `POST /api/v1/job-family-groups` 等 Job Catalog CRUD | 分类主数据维护 | `job-catalog:write` | 80号文档 §4.4 |
| `POST /api/v1/job-catalog/sync` | 分类同步任务 | `job-catalog:write` | 80号文档 §4.4 |

> 临时端点均在 OpenAPI 草案中以 `x-temporary` 标注，deadline 2025-12-31，迁移计划见 80号文档 §7.6。

---

## 2. GraphQL 查询 → 权限 Scope

| 查询 | 功能 | 最小 Scope | 参考文档 |
|------|------|------------|----------|
| `positions` / `position` | 职位列表/详情 | `position:read` | GraphQL 草案 §4 |
| `positionTimeline` | 职位时间线 | `position:read:history` | GraphQL 草案 §4 |
| `vacantPositions` | 空缺职位列表 | `position:read` | GraphQL 草案 §4 |
| `positionHeadcountStats` | 编制统计 | `position:read:stats` | GraphQL 草案 §4 |
| `jobFamilyGroups` / `jobFamilies` / `jobRoles` / `jobLevels` | 职位体系主数据查询 | `job-catalog:read` | GraphQL 草案 §4 |

> 若查询需要未来版本（`asOfDate > today`），服务端仍适用 `position:read:future` Scope 校验，后续契约落地时一并在 resolver 中验证。

---

## 3. 权限治理注意事项

1. **统一声明**：OpenAPI `components.securitySchemes.oauth.scopes` 与 GraphQL 自定义 directive 必须与本表保持一致，禁止新增未声明 Scope。  
2. **最小权限原则**：前端菜单与按钮应依据上述 Scope 做细粒度控制，避免过度授权。  
3. **审计要求**：命令服务每次调用需在审计日志记录 `scope`，方便后续追踪权限滥用。  
4. **临时端点回收**：当 Assignment Phase 4 上线并移除 `x-temporary` 实现后，再评审是否需要新增 `position:assignment:*` 相关 Scope。  
5. **变更流程**：若新增 Scope，需同步更新 80 号文档、第 5 节质量门禁清单以及安全团队的权限矩阵。

---

> 本附录将在联合评审时随同 81 号主计划提交，评审通过后若 Scope 有新增或调整，请同步更新此表。
