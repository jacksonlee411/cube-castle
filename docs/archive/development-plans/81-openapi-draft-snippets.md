# 81号附录：OpenAPI 契约草拟片段（Phase 1）

**版本**: v0.1 草案  
**创建日期**: 2025-10-14  
**维护团队**: 命令服务团队（与架构组共审）  
**用途**: 配合 81 号计划 Phase 1，提供职位管理相关 REST 契约的草拟片段，待评审通过后合入 `docs/api/openapi.yaml`。

---

## 1. 契约范围说明
- 路径前缀：`/api/v1/positions` 与 Job Catalog 相关子资源。  
- 安全模型：继承现有 `oauth` securityScheme，需声明职位权限 Scope。  
- 命名规范：所有字段保持 camelCase，路径参数统一 `{code}`，依赖 81 号主文档第 2.1.2 节的 pattern。  
- 临时端点：`/fill`、`/vacate`、`/transfer` 受 Assignment Phase 4 影响，使用 `x-temporary` 标注，deadline 统一 2025-12-31。

---

## 2. Components 片段

```yaml
components:
  schemas:
    PositionResource:
      type: object
      required:
        - code
        - title
        - organizationCode
        - status
        - effectiveDate
        - recordId
        - headcountCapacity
        - headcountInUse
        - availableHeadcount
      properties:
        code:
          type: string
          pattern: ^P[0-9]{7}$
          description: 职位编码（唯一）
        title:
          type: string
          maxLength: 120
        organizationCode:
          type: string
          pattern: ^[0-9]{7}$
        status:
          $ref: '#/components/schemas/PositionStatus'
        jobFamilyGroupCode:
          type: string
          pattern: ^[A-Z]{4,6}$
        jobFamilyCode:
          type: string
          pattern: ^[A-Z]{4,6}-[A-Z0-9]{3,6}$
        jobRoleCode:
          type: string
          pattern: ^[A-Z]{4,6}-[A-Z0-9]{3,6}-[A-Z0-9]{3,6}$
        jobLevelCode:
          type: string
          pattern: ^[A-Z][0-9]{1,2}$
        headcountCapacity:
          type: number
          format: float
        headcountInUse:
          type: number
          format: float
        availableHeadcount:
          type: number
          format: float
        effectiveDate:
          type: string
          format: date
        endDate:
          type: string
          format: date
          nullable: true
        isCurrent:
          type: boolean
        isFuture:
          type: boolean
        recordId:
          type: string
          format: uuid
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time

    CreatePositionRequest:
      type: object
      required:
        - title
        - jobFamilyGroupCode
        - jobFamilyCode
        - jobRoleCode
        - jobLevelCode
        - organizationCode
        - positionType
        - employmentType
        - headcountCapacity
        - effectiveDate
        - operationReason
      properties:
        title:
          type: string
          maxLength: 120
        jobProfileCode:
          type: string
          maxLength: 64
        jobProfileName:
          type: string
          maxLength: 120
        jobFamilyGroupCode:
          type: string
          pattern: ^[A-Z]{4,6}$
        jobFamilyGroupRecordId:
          type: string
          format: uuid
          nullable: true
        jobFamilyCode:
          type: string
          pattern: ^[A-Z]{4,6}-[A-Z0-9]{3,6}$
        jobFamilyRecordId:
          type: string
          format: uuid
          nullable: true
        jobRoleCode:
          type: string
          pattern: ^[A-Z]{4,6}-[A-Z0-9]{3,6}-[A-Z0-9]{3,6}$
        jobRoleRecordId:
          type: string
          format: uuid
          nullable: true
        jobLevelCode:
          type: string
          pattern: ^[A-Z][0-9]{1,2}$
        jobLevelRecordId:
          type: string
          format: uuid
          nullable: true
        organizationCode:
          type: string
          pattern: ^[0-9]{7}$
        positionType:
          $ref: '#/components/schemas/PositionType'
        employmentType:
          $ref: '#/components/schemas/EmploymentType'
        gradeLevel:
          type: string
          maxLength: 16
        headcountCapacity:
          type: number
          format: float
          minimum: 0
        reportsToPositionCode:
          type: string
          pattern: ^P[0-9]{7}$
          nullable: true
        effectiveDate:
          type: string
          format: date
        operationReason:
          type: string
          maxLength: 200

    PositionStatus:
      type: string
      enum:
        - PLANNED
        - ACTIVE
        - FILLED
        - VACANT
        - INACTIVE
        - DELETED

    PositionType:
      type: string
      enum:
        - REGULAR
        - TEMPORARY
        - CONTRACTOR

    EmploymentType:
      type: string
      enum:
        - FULL_TIME
        - PART_TIME
        - INTERN

    CreatePositionVersionRequest:
      allOf:
        - $ref: '#/components/schemas/CreatePositionRequest'
      required:
        - effectiveDate
      properties:
        recordId:
          type: string
          format: uuid
          nullable: true

    FillPositionRequest:
      type: object
      required:
        - effectiveDate
        - positionHolderId
        - operationReason
      properties:
        positionHolderId:
          type: string
          maxLength: 64
        assignmentType:
          type: string
          enum:
            - PRIMARY
            - SECONDARY
        fte:
          type: number
          format: float
          minimum: 0
          maximum: 1
          default: 1
        effectiveDate:
          type: string
          format: date
        operationReason:
          type: string
          maxLength: 200

    VacatePositionRequest:
      type: object
      required:
        - effectiveDate
        - operationReason
      properties:
        effectiveDate:
          type: string
          format: date
        fte:
          type: number
          format: float
          minimum: 0
          maximum: 1
          default: 1
        operationReason:
          type: string
          maxLength: 200
```

---

## 3. Paths 片段

```yaml
/api/v1/positions:
  post:
    summary: Create position
    operationId: createPosition
    tags: [Positions]
    security:
      - oauth:
          - position:create
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/CreatePositionRequest'
    responses:
      '201':
        description: Position created
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PositionResource'
      '409':
        $ref: '#/components/responses/Conflict'

/api/v1/positions/{code}:
  parameters:
    - name: code
      in: path
      required: true
      schema:
        type: string
        pattern: ^P[0-9]{7}$
  put:
    summary: Replace position
    operationId: replacePosition
    tags: [Positions]
    security:
      - oauth:
          - position:update
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/CreatePositionRequest'
    responses:
      '200':
        description: Updated position
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PositionResource'

/api/v1/positions/{code}/versions:
  post:
    summary: Insert position version
    operationId: createPositionVersion
    tags: [Positions]
    security:
      - oauth:
          - position:modify:history
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/CreatePositionVersionRequest'
    responses:
      '201':
        description: Version created
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PositionResource'

/api/v1/positions/{code}/fill:
  post:
    summary: Fill position (TEMPORARY – depends on Assignment Phase 4)
    description: |
      **⚠️ TEMPORARY IMPLEMENTATION**
      Deadline: 2025-12-31
      Migration: docs/development-plans/80-position-management-with-temporal-tracking.md#7.6
    x-temporary:
      reason: Assignment table not yet implemented
      deadline: '2025-12-31'
      migrationPlan: docs/development-plans/80-position-management-with-temporal-tracking.md#7.6
      owner: backend-architect-developer
    operationId: fillPosition
    tags: [Positions]
    security:
      - oauth:
          - position:fill
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/FillPositionRequest'
    responses:
      '200':
        description: Position filled
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PositionResource'

/api/v1/positions/{code}/vacate:
  post:
    summary: Vacate position (TEMPORARY – depends on Assignment Phase 4)
    description: |
      **⚠️ TEMPORARY IMPLEMENTATION**
      Deadline: 2025-12-31
      Migration: docs/development-plans/80-position-management-with-temporal-tracking.md#7.6
    x-temporary:
      reason: Assignment table not yet implemented
      deadline: '2025-12-31'
      migrationPlan: docs/development-plans/80-position-management-with-temporal-tracking.md#7.6
      owner: backend-architect-developer
    operationId: vacatePosition
    tags: [Positions]
    security:
      - oauth:
          - position:vacate
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/VacatePositionRequest'
    responses:
      '200':
        description: Position vacated
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PositionResource'

/api/v1/positions/{code}/transfer:
  post:
    summary: Transfer position (TEMPORARY – depends on Assignment Phase 4)
    description: |
      **⚠️ TEMPORARY IMPLEMENTATION**
      Deadline: 2025-12-31
      Migration: docs/development-plans/80-position-management-with-temporal-tracking.md#7.6
    x-temporary:
      reason: Assignment table not yet implemented
      deadline: '2025-12-31'
      migrationPlan: docs/development-plans/80-position-management-with-temporal-tracking.md#7.6
      owner: backend-architect-developer
    operationId: transferPosition
    tags: [Positions]
    security:
      - oauth:
          - position:transfer
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - targetOrganizationCode
              - effectiveDate
              - operationReason
            properties:
              targetOrganizationCode:
                type: string
                pattern: ^[0-9]{7}$
              effectiveDate:
                type: string
                format: date
              operationReason:
                type: string
                maxLength: 200
    responses:
      '200':
        description: Position transferred
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PositionResource'
```

---

## 4. 待确认事项
- `positionType`、`employmentType` 枚举是否与现有实现一致，若需扩展请在 80 号方案或实现清单中同步。  
- `jobProfileCode/name` 是否需要引入 pattern（暂保持字符串，依赖主数据校验）。  
- 临时端点在 Phase 4 前仍需与 17 号治理计划保持同步，如 Assignment 表落地需回收 `x-temporary`。  
- 与查询团队对齐响应体中 `job*` 相关嵌套对象的最小字段集，保持 GraphQL 与 REST 一致。

---

> 本草案供 Phase 1 评审使用，如需调整请在 81 号主文档或本附录中追加修订记录。
