# 06 — 恢复“显示所有版本”的实现方案（GraphQL 推荐）

最后更新：2025-09-13  
维护团队：架构组 + 前端组 + 查询服务组  
文档状态：方案评审通过后执行（不新增自动化脚本）

---

## 1. 目标（Outcome）
- 为前端提供“按业务编码 code 返回组织全部版本（按生效日排序）”的查询能力，恢复时间轴多版本展示。
- 保持 CQRS：查询走 GraphQL；命令走 REST。命名一致（camelCase），多租户隔离（X-Tenant-ID）。

---

## 2. GraphQL 契约（新增查询，推荐）

### 2.1 Query 定义（最小可用）
- 名称：`organizationVersions`
- 签名：
```
organizationVersions(
  code: String!,
  includeDeleted: Boolean = false
): [Organization!]!
```
- 语义：返回指定 `code` 的全部版本（不分页），按 `effectiveDate` 升序；默认过滤 `status='DELETED'` 与 `deleted_at IS NULL`。
- 权限：`org:read:history`

### 2.2 字段与命名
- 复用既有 `Organization` 类型（camelCase：`effectiveDate/endDate/isCurrent/...`）。
- 仅在 Query 层组合，避免重复类型定义。

---

## 3. 查询服务实现（后端）

### 3.1 位置
- 查询侧（GraphQL Read）：`cmd/organization-query-service`

### 3.2 数据访问（示意）
- 过滤：`tenant_id = $tenant AND code = $code`
  - `includeDeleted=false`：`status != 'DELETED' AND deleted_at IS NULL`
  - `includeDeleted=true`：放开上述条件
- 排序：`ORDER BY effective_date ASC`
- 映射：DB snake_case → API camelCase（例如 `effective_date → effectiveDate`）。

### 3.3 权限与隔离
- 从 JWT/Header 解析 tenantId；
- 校验 PBAC：`org:read:history`。

---

## 4. 前端改造（时间轴数据源切换）

### 4.1 修改点
- 文件：`frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`
- `loadVersions` 由“单体快照 `organization(code)`”改为“列表查询 `organizationVersions(code)`”。
- 将返回数组直接 map 为 `TimelineVersion[]`，按 `effectiveDate` ASC/或服务端已排序；选中当前版本 = “生效日 ≤ 今日”的最大者。

### 4.2 回退策略
- 新查询不可用/返回空时，回退到现有“单体快照”逻辑，避免页面空白；但在 UI 上提示“历史列表不可用，展示当前快照”。

### 4.3 交互与错误提示
- 插入中间时点后调用 `loadVersions()`，应显示多条（旧版 endDate=新版本前一日；新版本；后续版本）。
- 409 `TEMPORAL_POINT_CONFLICT`：提示“存在同一生效日记录，请选择非重复日期”。
- 400 `INVALID_DATE_FORMAT`：提示“日期格式为 YYYY-MM-DD（如 2025-08-01）”。

---

## 5. 文档与治理（reference 对齐）

### 5.1 实现清单（reference/02）
- GraphQL 查询章节新增：`organizationVersions(code, includeDeleted)`（权限：`org:read:history`）。
- 维护类 REST 端点（`refresh-hierarchy/batch-refresh-hierarchy/corehr/organizations`）标注“契约存在/未实现”，避免误导。

### 5.2 API 使用指南（reference/03）
- 新增“时态最佳实践与常见错误”：
  - 中间时点插入：201 成功并重算；
  - 重复时点：409 `TEMPORAL_POINT_CONFLICT`；
  - 日期格式：强制 `YYYY-MM-DD`；
  - 父组织：7 位且需存在（若提供）。

---

## 6. 验证与验收（手动，不新增脚本）

### 6.1 GraphQL 验收
- 在 GraphQL Playground 执行：
```
query {
  organizationVersions(code: "1000002") {
    recordId code name status effectiveDate endDate isCurrent
  }
}
```
- 期待：按生效日升序的多条版本；仅一条 `isCurrent=true`。

### 6.2 插入后验证
- 执行 REST：POST `/api/v1/organization-units/{code}/versions` 插入 `2025-08-01`；
- 再执行 `organizationVersions` 查询：出现 5/1、8/1、9/1 … 等多条，边界正确（旧版 endDate=新版本前一日）。

### 6.3 错误场景
- 重复时点：期望 409 `TEMPORAL_POINT_CONFLICT`；
- 错误日期格式（如 2025/8/1）：期望 400 `INVALID_DATE_FORMAT`。

---

## 7. 迭代计划（渐进增强）

### 7.1 第 1 步（最小）
- GraphQL 新增 `organizationVersions`，查询服务实现 resolver；
- 前端 `loadVersions` 切换至新查询（保留回退）。

### 7.2 第 2 步（可选优化）
- 新增 `organizationHistory(code, startDate, endDate, includeDeleted, pagination)`；
- 前端支持按时间范围筛选/分页（当版本量较大时）。

---

## 8. 风险与回退
- Schema 变更带来的前端类型不一致：避免新增复杂类型，直接复用 `Organization`；
- 若查询临时不可用，维持现有“单体快照 + 提示”的回退，不影响核心操作。

---

## 9. 附：字段与错误速查
- 字段约束：
  - name：2–255；unitType：`DEPARTMENT|ORGANIZATION_UNIT|PROJECT_TEAM`；
  - effectiveDate：`YYYY-MM-DD`；operationReason：5–500；
  - parentCode（可选）：7 位且存在于当前租户。
- 典型错误：
```json
// 400
{ "code": "INVALID_DATE_FORMAT", "message": "生效日期格式无效" }
// 409
{ "code": "TEMPORAL_POINT_CONFLICT", "message": "生效日期与现有版本冲突" }
```

---

## 10. 附录（蓝图）：Schema 片段 + Resolver 伪代码 + 前端 loadVersions 伪代码

### 10.1 GraphQL Schema 片段（新增 Query）
```graphql
extend type Query {
  """
  Return all temporal versions for an organization code (ascending by effectiveDate).
  Requires scope: org:read:history
  """
  organizationVersions(
    code: String!
    includeDeleted: Boolean = false
  ): [Organization!]!
}
```

### 10.2 Resolver 伪代码（查询服务 / Go）
```go
// Resolver: organizationVersions
func (r *queryResolver) OrganizationVersions(ctx context.Context, code string, includeDeleted *bool) ([]*model.Organization, error) {
  tenantID := auth.FromContext(ctx).TenantID // 从上下文解析多租户
  incDel := false
  if includeDeleted != nil { incDel = *includeDeleted }

  // 构建 SQL
  qb := sqlBuilder.Select(
    "record_id", "code", "name", "unit_type", "status", "level", "path",
    "sort_order", "description", "effective_date", "end_date",
    "created_at", "updated_at", "parent_code", "tenant_id", "is_current",
  ).From("organization_units").
    Where("tenant_id = ? AND code = ?", tenantID, code)

  if !incDel {
    qb = qb.Where("status != 'DELETED' AND deleted_at IS NULL")
  }

  qb = qb.OrderBy("effective_date ASC")

  rows, err := db.QueryContext(ctx, qb.SQL(), qb.Args()...)
  if err != nil { return nil, err }
  defer rows.Close()

  var out []*model.Organization
  for rows.Next() {
    var rec dbOrg // 承载 snake_case 列
    if err := rows.Scan(&rec.RecordID, &rec.Code, &rec.Name, &rec.UnitType, &rec.Status,
      &rec.Level, &rec.Path, &rec.SortOrder, &rec.Description,
      &rec.EffectiveDate, &rec.EndDate, &rec.CreatedAt, &rec.UpdatedAt,
      &rec.ParentCode, &rec.TenantID, &rec.IsCurrent); err != nil {
      return nil, err
    }
    out = append(out, mapToAPI(rec)) // 映射为 camelCase 的 GraphQL 类型
  }
  return out, nil
}
```

### 10.3 前端 loadVersions 伪代码（TypeScript/React）
```ts
// GraphQL 查询
const QUERY = gql`
  query OrganizationVersions($code: String!) {
    organizationVersions(code: $code) {
      recordId code name unitType status level
      effectiveDate endDate isCurrent createdAt updatedAt parentCode description
    }
  }
`;

async function loadVersions(isRetry = false) {
  try {
    setIsLoading(true);
    setLoadingError(null);
    if (!isRetry) setRetryCount(0);

    const data = await unifiedGraphQLClient.request<{ organizationVersions: Org[] }>(QUERY, { code: organizationCode });
    const list = data?.organizationVersions ?? [];
    const mapped: TimelineVersion[] = list.map(o => ({
      recordId: o.recordId,
      code: o.code,
      name: o.name,
      unitType: o.unitType,
      status: o.status,
      level: o.level,
      effectiveDate: o.effectiveDate,
      endDate: o.endDate ?? null,
      isCurrent: o.isCurrent,
      createdAt: o.createdAt,
      updatedAt: o.updatedAt,
      parentCode: o.parentCode ?? undefined,
      description: o.description ?? undefined,
      lifecycleStatus: o.isCurrent ? 'CURRENT' : 'HISTORICAL',
      businessStatus: o.status === 'ACTIVE' ? 'ACTIVE' : 'INACTIVE',
      dataStatus: 'NORMAL',
      path: '',
      sortOrder: 1,
      changeReason: '',
    }));
    const sorted = mapped.sort((a,b) => new Date(a.effectiveDate).getTime() - new Date(b.effectiveDate).getTime());
    setVersions(sorted);
    setSelectedVersion(sorted.find(v => v.isCurrent) ?? sorted.at(-1) ?? null);
  } catch (e) {
    setLoadingError(e instanceof Error ? e.message : String(e));
    // 回退：旧的单体快照逻辑（可选）
    await loadSingleSnapshotFallback();
  } finally {
    setIsLoading(false);
  }
}
```

