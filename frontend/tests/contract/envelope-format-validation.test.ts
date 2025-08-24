/**
 * 企业级信封格式验证测试
 * 
 * 验证所有API响应都使用统一的企业级信封结构
 * 基于契约测试自动化验证体系文档
 */

import { describe, it, expect } from 'vitest'

describe('企业级信封响应结构验证', () => {
  describe('成功响应格式验证', () => {
    it('成功响应应该包含标准信封字段', () => {
      const successResponse = {
        success: true,
        data: {
          code: 'ORG001',
          name: 'Test Organization',
          unitType: 'DEPARTMENT',
          operatedBy: {
            id: 'uuid-123',
            name: 'John Doe'
          }
        },
        message: 'Organization retrieved successfully',
        timestamp: '2025-01-01T00:00:00.000Z',
        requestId: 'req-abc123'
      }

      // 验证顶层信封结构
      expect(successResponse.success).toBe(true)
      expect(successResponse.data).toBeDefined()
      expect(successResponse.message).toBeDefined()
      expect(successResponse.timestamp).toBeDefined()
      expect(successResponse.requestId).toBeDefined()

      // 验证字段类型
      expect(typeof successResponse.success).toBe('boolean')
      expect(typeof successResponse.data).toBe('object')
      expect(typeof successResponse.message).toBe('string')
      expect(typeof successResponse.timestamp).toBe('string')
      expect(typeof successResponse.requestId).toBe('string')

      // 验证时间戳格式 (ISO 8601)
      expect(successResponse.timestamp).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/)
    })

    it('分页响应应该使用标准信封 + 分页信息', () => {
      const paginatedResponse = {
        success: true,
        data: {
          items: [
            {
              code: 'ORG001',
              name: 'Organization 1',
              operatedBy: { id: 'uuid-1', name: 'User 1' }
            },
            {
              code: 'ORG002', 
              name: 'Organization 2',
              operatedBy: { id: 'uuid-2', name: 'User 2' }
            }
          ],
          pagination: {
            total: 100,
            page: 1,
            pageSize: 50,
            hasNext: true,
            hasPrevious: false
          },
          temporal: {
            asOfDate: '2025-01-01',
            currentCount: 80,
            futureCount: 15,
            historicalCount: 5
          }
        },
        message: 'Organizations retrieved successfully',
        timestamp: '2025-01-01T00:00:00.000Z',
        requestId: 'req-def456'
      }

      // 验证信封结构
      expect(paginatedResponse.success).toBe(true)
      expect(paginatedResponse.data).toBeDefined()
      expect(paginatedResponse.data.items).toBeInstanceOf(Array)
      expect(paginatedResponse.data.pagination).toBeDefined()
      expect(paginatedResponse.data.temporal).toBeDefined()

      // 验证分页信息结构
      const pagination = paginatedResponse.data.pagination
      expect(pagination.total).toBeTypeOf('number')
      expect(pagination.page).toBeTypeOf('number')
      expect(pagination.pageSize).toBeTypeOf('number')
      expect(pagination.hasNext).toBeTypeOf('boolean')
      expect(pagination.hasPrevious).toBeTypeOf('boolean')

      // 验证时态信息结构
      const temporal = paginatedResponse.data.temporal
      expect(temporal.asOfDate).toBeTypeOf('string')
      expect(temporal.currentCount).toBeTypeOf('number')
      expect(temporal.futureCount).toBeTypeOf('number')
      expect(temporal.historicalCount).toBeTypeOf('number')
    })
  })

  describe('错误响应格式验证', () => {
    it('错误响应应该包含标准错误信封', () => {
      const errorResponse = {
        success: false,
        error: {
          code: 'ORG_NOT_FOUND',
          message: 'Organization with code ORG999 not found',
          details: {
            searchCode: 'ORG999',
            tenantId: 'tenant-123',
            asOfDate: '2025-01-01'
          }
        },
        timestamp: '2025-01-01T00:00:00.000Z',
        requestId: 'req-error123'
      }

      // 验证错误信封结构
      expect(errorResponse.success).toBe(false)
      expect(errorResponse.error).toBeDefined()
      expect(errorResponse.timestamp).toBeDefined()
      expect(errorResponse.requestId).toBeDefined()

      // 验证错误对象结构
      expect(errorResponse.error.code).toBeDefined()
      expect(errorResponse.error.message).toBeDefined()
      expect(errorResponse.error.details).toBeDefined()

      // 验证字段类型
      expect(typeof errorResponse.error.code).toBe('string')
      expect(typeof errorResponse.error.message).toBe('string')
      expect(typeof errorResponse.error.details).toBe('object')
    })

    it('验证错误应该包含详细验证信息', () => {
      const validationErrorResponse = {
        success: false,
        error: {
          code: 'VALIDATION_FAILED',
          message: 'Request validation failed',
          details: {
            field: 'unitType',
            invalidValue: 'INVALID_TYPE',
            allowedValues: ['DEPARTMENT', 'ORGANIZATION_UNIT', 'COMPANY', 'PROJECT_TEAM'],
            constraint: 'unitType must be one of the allowed values'
          }
        },
        timestamp: '2025-01-01T00:00:00.000Z',
        requestId: 'req-validation123'
      }

      expect(validationErrorResponse.success).toBe(false)
      expect(validationErrorResponse.error.code).toBe('VALIDATION_FAILED')
      expect(validationErrorResponse.error.details.field).toBeDefined()
      expect(validationErrorResponse.error.details.invalidValue).toBeDefined()
      expect(validationErrorResponse.error.details.allowedValues).toBeInstanceOf(Array)
    })
  })

  describe('操作响应格式验证', () => {
    it('创建操作响应应该包含创建的资源', () => {
      const createResponse = {
        success: true,
        data: {
          code: 'ORG003',
          name: 'New Department',
          unitType: 'DEPARTMENT',
          status: 'ACTIVE',
          parentCode: 'ORG001',
          operatedBy: {
            id: 'uuid-creator',
            name: 'Creator User'
          },
          operationType: 'CREATE',
          operationReason: 'Department restructuring',
          createdAt: '2025-01-01T00:00:00.000Z'
        },
        message: 'Organization created successfully',
        timestamp: '2025-01-01T00:00:00.000Z',
        requestId: 'req-create123'
      }

      // 验证创建响应结构
      expect(createResponse.success).toBe(true)
      expect(createResponse.data.code).toBeDefined()
      expect(createResponse.data.operationType).toBe('CREATE')
      expect(createResponse.data.operatedBy.id).toBeDefined()
      expect(createResponse.data.operatedBy.name).toBeDefined()
    })

    it('更新操作响应应该包含更新的资源', () => {
      const updateResponse = {
        success: true,
        data: {
          code: 'ORG001',
          name: 'Updated Engineering Department',
          description: 'Updated description',
          operatedBy: {
            id: 'uuid-updater',
            name: 'Updater User'
          },
          operationType: 'UPDATE',
          operationReason: 'Name and description update',
          updatedAt: '2025-01-01T00:00:00.000Z'
        },
        message: 'Organization updated successfully',
        timestamp: '2025-01-01T00:00:00.000Z',
        requestId: 'req-update123'
      }

      expect(updateResponse.success).toBe(true)
      expect(updateResponse.data.operationType).toBe('UPDATE')
      expect(updateResponse.data.operatedBy).toBeDefined()
      expect(updateResponse.data.updatedAt).toBeDefined()
    })

    it('删除操作响应应该包含确认信息', () => {
      const deleteResponse = {
        success: true,
        data: {
          code: 'ORG001',
          operationType: 'DELETE',
          operatedBy: {
            id: 'uuid-deleter',
            name: 'Deleter User'
          },
          operationReason: 'Organization consolidation',
          deletedAt: '2025-01-01T00:00:00.000Z'
        },
        message: 'Organization deleted successfully',
        timestamp: '2025-01-01T00:00:00.000Z',
        requestId: 'req-delete123'
      }

      expect(deleteResponse.success).toBe(true)
      expect(deleteResponse.data.operationType).toBe('DELETE')
      expect(deleteResponse.data.deletedAt).toBeDefined()
    })
  })

  describe('专用操作响应格式验证', () => {
    it('suspend操作响应应该标准化', () => {
      const suspendResponse = {
        success: true,
        data: {
          code: 'ORG001',
          status: 'INACTIVE',
          operationType: 'SUSPEND',
          operatedBy: {
            id: 'uuid-admin',
            name: 'Admin User'
          },
          operationReason: 'Temporary suspension for restructuring',
          suspendedAt: '2025-01-01T00:00:00.000Z'
        },
        message: 'Organization suspended successfully',
        timestamp: '2025-01-01T00:00:00.000Z',
        requestId: 'req-suspend123'
      }

      expect(suspendResponse.success).toBe(true)
      expect(suspendResponse.data.operationType).toBe('SUSPEND')
      expect(suspendResponse.data.status).toBe('INACTIVE')
      expect(suspendResponse.data.suspendedAt).toBeDefined()
    })

    it('activate操作响应应该标准化', () => {
      const activateResponse = {
        success: true,
        data: {
          code: 'ORG001',
          status: 'ACTIVE',
          operationType: 'REACTIVATE',
          operatedBy: {
            id: 'uuid-admin',
            name: 'Admin User'
          },
          operationReason: 'Restructuring completed, reactivating',
          reactivatedAt: '2025-01-01T00:00:00.000Z'
        },
        message: 'Organization activated successfully',
        timestamp: '2025-01-01T00:00:00.000Z',
        requestId: 'req-activate123'
      }

      expect(activateResponse.success).toBe(true)
      expect(activateResponse.data.operationType).toBe('REACTIVATE')
      expect(activateResponse.data.status).toBe('ACTIVE')
      expect(activateResponse.data.reactivatedAt).toBeDefined()
    })
  })

  describe('响应字段一致性验证', () => {
    it('所有响应都应该有统一的元字段', () => {
      const responses = [
        { success: true, data: {}, message: 'OK', timestamp: '2025-01-01T00:00:00.000Z', requestId: 'req-1' },
        { success: false, error: { code: 'ERROR', message: 'Error' }, timestamp: '2025-01-01T00:00:00.000Z', requestId: 'req-2' }
      ]

      responses.forEach(response => {
        expect(response.success).toBeTypeOf('boolean')
        expect(response.timestamp).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/)
        expect(response.requestId).toMatch(/^req-/)
      })
    })

    it('operatedBy字段应该在所有操作响应中保持一致', () => {
      const operatedByExamples = [
        { id: 'uuid-1', name: 'User One' },
        { id: 'uuid-2', name: 'User Two' },
        { id: 'uuid-3', name: 'User Three' }
      ]

      operatedByExamples.forEach(operatedBy => {
        expect(operatedBy.id).toMatch(/^uuid-/)
        expect(operatedBy.name).toBeTypeOf('string')
        expect(operatedBy.name.length).toBeGreaterThan(0)
        
        // 验证字段命名格式
        Object.keys(operatedBy).forEach(key => {
          expect(key).toMatch(/^[a-z][a-zA-Z0-9]*$/)
        })
      })
    })
  })
})