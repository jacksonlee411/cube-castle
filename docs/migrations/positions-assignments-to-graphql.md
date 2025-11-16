# 迁移指南：Positions Assignments（REST → GraphQL）

文档编号：MIG‑PA‑GQL  
创建日期：2025‑11‑16  
适用范围：将 REST 业务查询端点 `GET /api/v1/positions/{code}/assignments` 迁移至 GraphQL 查询  
关联计划：259（主）、259A、257、258、215、AGENTS.md  

---

## 1. 背景与目标
- 背景：为消除“REST 业务查询”与 GraphQL 查询的双事实来源，遵循 PostgreSQL 原生 CQRS（命令=REST、查询=GraphQL），259‑T4 启动“弃用 REST assignments 查询”的治理。
- 目标：前端与集成调用统一迁移至 GraphQL：
  - `positionAssignments(positionCode, filter, pagination, sorting)`
  - `assignments(organizationCode, positionCode, filter, pagination, sorting)`
- 权限：`position:assignments:read`（已与 REST 对齐，见 259‑T3）

---

## 2. GraphQL 查询示例

查询 1：按职位 code 分页获取任职记录
```graphql
query PositionAssignments($positionCode: PositionCode!, $page: Int = 1, $pageSize: Int = 25) {
  positionAssignments(
    positionCode: $positionCode
    pagination: { page: $page, pageSize: $pageSize }
  ) {
    data {
      assignmentId
      employeeId
      employeeName
      type
      status
      effectiveDate
      endDate
      fte
    }
    pagination { total page pageSize }
  }
}
```

查询 2：按组织/职位过滤（等价 REST 查询参数）
```graphql
query Assignments($organizationCode: String, $positionCode: PositionCode, $filter: PositionAssignmentFilterInput, $page: Int = 1, $pageSize: Int = 25) {
  assignments(
    organizationCode: $organizationCode
    positionCode: $positionCode
    filter: $filter
    pagination: { page: $page, pageSize: $pageSize }
  ) {
    data { assignmentId employeeId status effectiveDate endDate }
    pagination { total page pageSize }
  }
}
```

常用过滤器映射（REST → GraphQL）：
- `assignmentTypes` → `filter: { types: [...] }`
- `status` → `filter: { status: ... }`
- `asOfDate` → `filter: { asOfDate: "YYYY‑MM‑DD" }`
- `includeHistorical` → `filter: { includeHistorical: true }`
- `includeActingOnly` → `filter: { actingOnly: true }`

---

## 3. 客户端调用建议（前端）
- 统一通过领域 API 门面与 `UnifiedGraphQLClient` 发起查询（Plan 257）；
- 禁止在业务组件中直接使用 `fetch/axios` 访问 REST 查询端点。

示例（TypeScript）：
```ts
import { unifiedGraphQLClient } from "@/shared/api/unified-client";

export async function listPositionAssignments(positionCode: string, page = 1, pageSize = 25) {
  const query = `
    query($positionCode: PositionCode!, $page: Int!, $pageSize: Int!) {
      positionAssignments(positionCode: $positionCode, pagination: { page: $page, pageSize: $pageSize }) {
        data { assignmentId employeeId employeeName status effectiveDate endDate }
        pagination { total page pageSize }
      }
    }`;
  const variables = { positionCode, page, pageSize };
  return unifiedGraphQLClient.request(query, variables);
}
```

---

## 4. 兼容与时间表
- OpenAPI 已标注弃用：`GET /api/v1/positions/{code}/assignments`（deprecated=true）
- Sunset：2025‑12‑20 00:00:00Z（届时计划移除 REST 端点）
- CI 门禁：Plan 259A 软门禁（业务 GET 阈值=1），迁移完成后改为 0，转为硬门禁；随后移除 REST 端点

---

## 5. 验收与证据
- 验收：
  - 前端调用点替换为 GraphQL，E2E/集成测试通过（Playwright/后端集成均绿）
  - CI：Plan 259A 阈值=0 时通过
- 证据：
  - `reports/plan259/protocol-duplication-matrix.json`（restBusinessGetCount=0）
  - 215 执行日志登记迁移完成与门禁切换时间点

---

## 6. 回滚策略
- 客户端回退：保持原 REST 端点未移除前，可切回旧调用；建议通过 Feature Flag 控制
- 服务器端：仅为契约弃用标注与文档公告，无数据/迁移变更，不涉及 DB 回滚

---

## 7. 关联
- 方案与决议：`docs/development-plans/259-protocol-strategy-review.md`、`docs/development-plans/259A-protocol-duplication-and-whitelist-hardening.md`
- 门禁与产物：`make guard-plan259`、`.github/workflows/plan-258-gates.yml`
- 原则：`AGENTS.md`

