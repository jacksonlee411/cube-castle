# 81号附录：GraphQL 契约草拟片段（Phase 2）

**版本**: v0.1 草案  
**创建日期**: 2025-10-14  
**维护团队**: 查询服务团队（与架构组联审）  
**用途**: 配合 81 号计划 Phase 2，给出职位管理相关 GraphQL Schema 草拟片段，评审通过后合入 `docs/api/schema.graphql`。

---

## 1. 契约范围说明
- 命名空间保持在现有 Schema 顶层，新增类型 `Position`、`JobFamilyGroup` 等扩展字段。  
- 统一使用 camelCase，租户上下文仍由服务器从请求头 `X-Tenant-ID` 注入。  
- Enum / Scalar 与 OpenAPI 在字段和 pattern 上保持一致，遵循 81 号主文档第 2.1.2 节。  
- 查询默认需要权限 Scope：`position:read` / `job-catalog:read`；带历史或未来数据需额外 Scope。

---

## 2. Scalar & Enum 片段

```graphql
scalar PositionCode @constraint(pattern: "^P[0-9]{7}$")
scalar OrganizationCode @constraint(pattern: "^[0-9]{7}$")
scalar JobFamilyGroupCode @constraint(pattern: "^[A-Z]{4,6}$")
scalar JobFamilyCode @constraint(pattern: "^[A-Z]{4,6}-[A-Z0-9]{3,6}$")
scalar JobRoleCode @constraint(pattern: "^[A-Z]{4,6}-[A-Z0-9]{3,6}-[A-Z0-9]{3,6}$")
scalar JobLevelCode @constraint(pattern: "^[A-Z][0-9]{1,2}$")

enum PositionStatus {
  PLANNED
  ACTIVE
  FILLED
  VACANT
  INACTIVE
  DELETED
}

enum PositionType {
  REGULAR
  TEMPORARY
  CONTRACTOR
}

enum EmploymentType {
  FULL_TIME
  PART_TIME
  INTERN
}
```

> 说明：若 gqlgen 版本暂不支持 `@constraint`，将以自定义 Scalar + 验证函数实现，评审时明确落地方式。

---

## 3. Type 片段

```graphql
type Position {
  code: PositionCode!
  recordId: ID!
  tenantId: ID!
  title: String!

  jobProfileCode: String
  jobProfileName: String

  jobFamilyGroupCode: JobFamilyGroupCode!
  jobFamilyGroup: JobFamilyGroup!
  jobFamilyCode: JobFamilyCode!
  jobFamily: JobFamily!
  jobRoleCode: JobRoleCode!
  jobRole: JobRole!
  jobLevelCode: JobLevelCode!
  jobLevel: JobLevel!

  organizationCode: OrganizationCode!
  organizationName: String
  organization: OrganizationUnit

  positionType: PositionType!
  employmentType: EmploymentType!
  gradeLevel: String

  headcountCapacity: Float!
  headcountInUse: Float!
  availableHeadcount: Float!

  reportsToPositionCode: PositionCode
  reportsToPosition: Position

  status: PositionStatus!
  effectiveDate: Date!
  endDate: Date
  isCurrent: Boolean!
  isFuture: Boolean!
  createdAt: DateTime!
  updatedAt: DateTime!

  timeline: [PositionTimelineEntry!]!
}

type PositionTimelineEntry {
  recordId: ID!
  status: PositionStatus!
  title: String!
  effectiveDate: Date!
  endDate: Date
  isCurrent: Boolean!
  changeReason: String
}

type PositionConnection {
  edges: [PositionEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type PositionEdge {
  cursor: String!
  node: Position!
}

type HeadcountStats {
  organizationCode: OrganizationCode!
  organizationName: String!
  totalCapacity: Float!
  totalFilled: Float!
  totalAvailable: Float!
  fillRate: Float!
  byLevel: [LevelHeadcount!]!
  byType: [TypeHeadcount!]!
}

type LevelHeadcount {
  jobLevelCode: JobLevelCode!
  capacity: Float!
  utilized: Float!
  available: Float!
}

type TypeHeadcount {
  positionType: PositionType!
  capacity: Float!
  filled: Float!
  available: Float!
}
```

---

## 4. Query 片段

```graphql
type Query {
  positions(
    filter: PositionFilterInput
    pagination: PaginationInput
    sorting: [PositionSortInput!]
  ): PositionConnection! @requiresPermissions(permissions: ["position:read"])

  position(code: PositionCode!, asOfDate: Date): Position
    @requiresPermissions(permissions: ["position:read"])

  positionTimeline(
    code: PositionCode!
    startDate: Date
    endDate: Date
  ): [PositionTimelineEntry!]!
    @requiresPermissions(permissions: ["position:read:history"])

  vacantPositions(
    organizationCode: OrganizationCode
    positionType: PositionType
    includeSubordinates: Boolean = true
  ): [Position!]!
    @requiresPermissions(permissions: ["position:read"])

  positionHeadcountStats(
    organizationCode: OrganizationCode!
    includeSubordinates: Boolean = true
  ): HeadcountStats!
    @requiresPermissions(permissions: ["position:read:stats"])

  jobFamilyGroups(includeInactive: Boolean = false, asOfDate: Date): [JobFamilyGroup!]!
    @requiresPermissions(permissions: ["job-catalog:read"])

  jobFamilies(
    groupCode: JobFamilyGroupCode!
    includeInactive: Boolean = false
    asOfDate: Date
  ): [JobFamily!]!
    @requiresPermissions(permissions: ["job-catalog:read"])

  jobRoles(
    familyCode: JobFamilyCode!
    includeInactive: Boolean = false
    asOfDate: Date
  ): [JobRole!]!
    @requiresPermissions(permissions: ["job-catalog:read"])

  jobLevels(
    roleCode: JobRoleCode!
    includeInactive: Boolean = false
    asOfDate: Date
  ): [JobLevel!]!
    @requiresPermissions(permissions: ["job-catalog:read"])
}
```

---

## 5. Input & Filtering 片段

```graphql
input PositionFilterInput {
  organizationCode: OrganizationCode
  positionCodes: [PositionCode!]
  status: PositionStatus
  jobFamilyGroupCodes: [JobFamilyGroupCode!]
  jobFamilyCodes: [JobFamilyCode!]
  jobRoleCodes: [JobRoleCode!]
  jobLevelCodes: [JobLevelCode!]
  positionTypes: [PositionType!]
  employmentTypes: [EmploymentType!]
  effectiveRange: DateRangeInput
}

input PositionSortInput {
  field: PositionSortField!
  direction: SortDirection! = ASC
}

enum PositionSortField {
  CODE
  TITLE
  EFFECTIVE_DATE
  STATUS
}
```

> 与现有分页、排序结构保持一致；`DateRangeInput`、`PaginationInput` 已在现有 Schema 定义，仅复用。

---

## 6. 待确认事项
- `@requiresPermissions` 采用现有自定义 directive，如命令服务同步变更请提前通知。  
- GraphQL 输出中 `timeline` 是否需要限制长度或引入分页，待业务评审确认。  
- `vacantPositions` 是否需补充 `jobFamilyCode` 等过滤条件，如需扩展请在评审会上说明。  
- 与命令服务对齐临时端点回收计划，确保当 Assignment Phase 4 完成后字段/指示同步更新。

---

> 本草案用于 Phase 2 评审使用，后续调整请更新本附录并在主文档记录变更。
