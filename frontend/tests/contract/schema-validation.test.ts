/**
 * GraphQL Schema Validation Tests
 * 
 * 验证前端GraphQL查询与Schema v4.2.1的一致性
 * 基于契约测试自动化验证体系文档
 */

import { describe, it, expect } from 'vitest'
import { readFileSync } from 'fs'
import { parse, validate, buildSchema } from 'graphql'
import { join } from 'path'

describe('GraphQL Schema 契约验证', () => {
  let schema: ReturnType<typeof buildSchema>

  beforeAll(() => {
    // 读取Schema文件
    const schemaPath = join(process.cwd(), '../docs/api/schema.graphql')
    const schemaString = readFileSync(schemaPath, 'utf-8')
    schema = buildSchema(schemaString)
  })

  describe('L1 - 语法层验证', () => {
    it('Schema语法应该正确无误', () => {
      expect(schema).toBeDefined()
      expect(schema.getQueryType()).toBeDefined()
    })

    it('所有核心查询字段应该存在', () => {
      const queryType = schema.getQueryType()
      const fields = queryType.getFields()
      
      // 验证9个核心查询端点 (v4.6.0)
      const expectedQueries = [
        'organizations',
        'organization', 
        'organizationStats',
        'organizationHierarchy',
        'organizationSubtree',
        'hierarchyStatistics',
        'auditHistory',
        'auditLog',
        'hierarchyConsistencyCheck'
      ]
      
      expectedQueries.forEach(queryName => {
        expect(fields[queryName]).toBeDefined()
      })
    })

    it('Organization核心字段应该完整', () => {
      const organizationType = schema.getType('Organization')
      const fields = organizationType.getFields()
      
      // 验证核心业务字段存在
      const requiredFields = [
        'code', 'parentCode', 'tenantId', 'name', 'unitType', 'status',
        'level', 'effectiveDate', 'endDate', 'isCurrent', 'isFuture',
        'createdAt', 'updatedAt', 'operationType', 'operatedBy'
      ]
      
      requiredFields.forEach(fieldName => {
        expect(fields[fieldName]).toBeDefined()
      })
    })
  })

  describe('L2 - 语义层验证', () => {
    it('字段命名应该遵循camelCase规范', () => {
      const organizationType = schema.getType('Organization')
      const fields = Object.keys(organizationType.getFields())
      
      fields.forEach(fieldName => {
        // 验证camelCase命名格式
        expect(fieldName).toMatch(/^[a-z][a-zA-Z0-9]*$/)
        // 确保没有snake_case
        expect(fieldName).not.toMatch(/_/)
      })
    })

    it('operatedBy字段应该是标准对象结构', () => {
      const organizationType = schema.getType('Organization')
      const operatedByField = organizationType.getFields().operatedBy
      const operatedByType = schema.getType('OperatedBy')
      
      expect(operatedByField).toBeDefined()
      expect(operatedByType).toBeDefined()
      
      const operatedByFields = operatedByType.getFields()
      expect(operatedByFields.id).toBeDefined()
      expect(operatedByFields.name).toBeDefined()
    })

    it('时态字段命名应该标准化', () => {
      const organizationType = schema.getType('Organization')
      const fields = organizationType.getFields()
      
      // 验证时态字段命名标准
      expect(fields.effectiveDate).toBeDefined()
      expect(fields.endDate).toBeDefined()
      expect(fields.isCurrent).toBeDefined()
      expect(fields.isFuture).toBeDefined()
      expect(fields.createdAt).toBeDefined()
      expect(fields.updatedAt).toBeDefined()
      
      // 确保没有旧式命名
      expect(fields.effective_date).toBeUndefined()
      expect(fields.end_date).toBeUndefined()
      expect(fields.is_current).toBeUndefined()
    })
  })

  describe('L3 - 集成层验证', () => {
    it('分页查询结构应该标准化', () => {
      const connectionType = schema.getType('OrganizationConnection')
      const fields = connectionType.getFields()
      
      expect(fields.data).toBeDefined()
      expect(fields.pagination).toBeDefined()
      expect(fields.temporal).toBeDefined()
      
      // 验证分页信息结构
      const paginationType = schema.getType('PaginationInfo')
      const paginationFields = paginationType.getFields()
      expect(paginationFields.total).toBeDefined()
      expect(paginationFields.page).toBeDefined()
      expect(paginationFields.pageSize).toBeDefined()
      expect(paginationFields.hasNext).toBeDefined()
      expect(paginationFields.hasPrevious).toBeDefined()
    })

    it('查询参数应该支持时态过滤', () => {
      const filterType = schema.getType('OrganizationFilter')
      const fields = filterType.getFields()
      
      expect(fields.asOfDate).toBeDefined()
      expect(fields.includeFuture).toBeDefined()
      expect(fields.onlyFuture).toBeDefined()
    })

    it('audit相关字段应该完整', () => {
      const auditLogType = schema.getType('AuditLogDetail')
      const auditFields = auditLogType.getFields()
      
      expect(auditFields.auditId).toBeDefined()
      expect(auditFields.recordId).toBeDefined()
      expect(auditFields.operation).toBeDefined()
      expect(auditFields.timestamp).toBeDefined()
      expect(auditFields.userInfo).toBeDefined()
    })
  })
})

describe('实际查询验证', () => {
  let schema: ReturnType<typeof buildSchema>

  beforeAll(() => {
    const schemaPath = join(process.cwd(), '../docs/api/schema.graphql')
    const schemaString = readFileSync(schemaPath, 'utf-8')
    schema = buildSchema(schemaString)
  })

  it('基础组织查询应该有效', () => {
    const query = `
      query GetOrganizations($filter: OrganizationFilter, $pagination: PaginationInput) {
        organizations(filter: $filter, pagination: $pagination) {
          data {
            code
            name
            unitType
            status
            parentCode
            level
            effectiveDate
            isCurrent
            operatedBy {
              id
              name
            }
          }
          pagination {
            total
            page
            pageSize
            hasNext
          }
          temporal {
            asOfDate
            currentCount
            futureCount
          }
        }
      }
    `

    const document = parse(query)
    const errors = validate(schema, document)
    expect(errors).toHaveLength(0)
  })

  it('层级查询应该有效', () => {
    const query = `
      query GetOrganizationHierarchy($code: String!, $tenantId: UUID!) {
        organizationHierarchy(code: $code, tenantId: $tenantId) {
          code
          name
          level
          hierarchyDepth
          codePath
          namePath
          parentChain
          childrenCount
          isRoot
          isLeaf
          children {
            code
            name
            level
          }
        }
      }
    `

    const document = parse(query)
    const errors = validate(schema, document)
    expect(errors).toHaveLength(0)
  })

  it('审计查询应该有效', () => {
    const query = `
      query GetAuditHistory($recordId: UUID!) {
        auditHistory(recordId: $recordId) {
          auditId
          recordId
          operation
          timestamp
          userInfo {
            userId
            userName
          }
          operationReason
          dataChanges {
            beforeData
            afterData
            modifiedFields
          }
        }
      }
    `

    const document = parse(query)
    const errors = validate(schema, document)
    expect(errors).toHaveLength(0)
  })
})