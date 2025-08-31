import type { 
  OrganizationUnit, 
  OrganizationListResponse, 
  OrganizationStats,
  OrganizationQueryParams
} from '../types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../hooks/useOrganizationMutations';
import type { 
  TemporalQueryParams,
  TemporalOrganizationUnit
} from '../types/temporal';
import { 
  validateOrganizationBasic,
  validateOrganizationUpdate,
  validateStatusUpdate,
  safeTransform,
  SimpleValidationError,
  formatValidationErrors
} from '../validation/simple-validation';
import { unifiedGraphQLClient, unifiedRESTClient } from './unified-client';

// 扩展查询参数以支持时态查询
interface ExtendedOrganizationQueryParams extends OrganizationQueryParams {
  searchText?: string;
  pageSize?: number;
  temporalParams?: TemporalQueryParams;
}

export const organizationAPI = {
  // 获取组织单元列表 - 使用GraphQL (修复getByCode问题)
  getAll: async (params?: ExtendedOrganizationQueryParams): Promise<OrganizationListResponse> => {
    try {
      // 轻量级参数验证
      if (params) {
        // 简化的参数验证，依赖后端详细验证
        if (params.page && params.page < 1) {
          throw new SimpleValidationError('页码必须大于0', [
            { field: 'page', message: '页码必须大于0' }
          ]);
        }
        if (params.pageSize && (params.pageSize < 1 || params.pageSize > 100)) {
          throw new SimpleValidationError('页面大小必须在1-100之间', [
            { field: 'pageSize', message: '页面大小必须在1-100之间' }
          ]);
        }
      }

      // 构建GraphQL查询和变量 (基础版本，不含时态参数)
      const useTemporalQuery = params?.temporalParams && Object.keys(params.temporalParams).length > 0;
      
      let graphqlQuery, variables;
      
      if (useTemporalQuery) {
        // 时态查询版本 - 使用真实的organizations查询
        graphqlQuery = `
          query GetOrganizations(
            $filter: OrganizationFilter,
            $pagination: PaginationInput
          ) {
            organizations(filter: $filter, pagination: $pagination) {
              data {
                code
                parentCode
                tenantId
                recordId
                name
                unitType
                status
                level
                path
                sortOrder
                description
                profile
                effectiveDate
                endDate
                isCurrent
                isTemporal
                createdAt
                updatedAt
              }
              pagination {
                total
                page
                pageSize
                hasNext
                hasPrevious
              }
              temporal {
                asOfDate
                currentCount
                futureCount
                historicalCount
              }
            }
          }
        `;
        variables = {
          filter: {
            asOfDate: params?.temporalParams?.asOfDate || null,
            searchText: params?.searchText || null
          },
          pagination: {
            page: params?.page || 1,
            pageSize: params?.pageSize || 50
          }
        };
      } else {
        // 基础查询版本（不含时态参数）- 使用正确的OrganizationConnection结构
        graphqlQuery = `
          query GetOrganizations {
            organizations {
              data {
                code
                name
                unitType
                status
                level
                path
                sortOrder
                description
                parentCode
                createdAt
                updatedAt
              }
              pagination {
                total
                page
                pageSize
              }
            }
          }
        `;
        variables = {};
      }

      const data = await unifiedGraphQLClient.request<{
        organizations: {
          data: Partial<OrganizationUnit>[];
          pagination: {
            total: number;
            page: number;
            pageSize: number;
          };
        };
      }>(graphqlQuery, variables);

      // 🔧 修复P0级数据契约问题: 使用OrganizationConnection结构
      // 后端返回: organizations: {data: [...], pagination: {total, page, pageSize}}
      // 前端期望: organizations: [...], totalCount: number
      const organizations = (data.organizations?.data || []).map((org: Partial<OrganizationUnit>) => {
        try {
          return safeTransform.graphqlToOrganization ? 
            safeTransform.graphqlToOrganization(org) : 
            org; // 直接返回原始数据，依赖后端格式
        } catch (error) {
          console.warn('Failed to transform organization:', org, error);
          return null;
        }
      }).filter(Boolean);

      // 🔧 修复: 从organizations.pagination.total获取总数，符合OrganizationConnection结构
      const totalCount = data.organizations?.pagination?.total || 0;
      
      return {
        organizations: organizations.filter((org): org is OrganizationUnit => org !== null),
        totalCount: totalCount,
        page: params?.page || 1,
        pageSize: organizations.length,
        totalPages: Math.ceil(totalCount / (params?.pageSize || 50))
      };

    } catch (error) {
      console.error('Error fetching organizations:', error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to fetch organizations. Please try again.');
    }
  },

  // 根据代码获取单个组织 - ✅ 修复协议违反，统一使用GraphQL (支持时态查询)
  getByCode: async (code: string, temporalParams?: TemporalQueryParams): Promise<OrganizationUnit> => {
    try {
      if (!code || typeof code !== 'string') {
        throw new SimpleValidationError('Invalid organization code', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      // ✅ 使用GraphQL查询，遵循"查询统一用GraphQL"原则 (基础版本)
      const useTemporalQuery = temporalParams && Object.keys(temporalParams).length > 0;
      
      let graphqlQuery, variables;
      
      if (useTemporalQuery) {
        // 时态查询版本 - 使用真实的organization查询
        graphqlQuery = `
          query GetOrganization(
            $code: String!,
            $asOfDate: Date
          ) {
            organization(code: $code, asOfDate: $asOfDate) {
              code
              parentCode
              tenantId
              recordId
              name
              unitType
              status
              level
              path
              sortOrder
              description
              profile
              effectiveDate
              endDate
              isCurrent
              isTemporal
              createdAt
              updatedAt
            }
          }
        `;
        variables = {
          code,
          asOfDate: temporalParams?.asOfDate || null
        };
      } else {
        // 基础查询版本（不含时态参数）
        graphqlQuery = `
          query GetOrganization($code: String!) {
            organization(code: $code) {
              code
              recordId
              name
              unitType
              status
              level
              path
              sortOrder
              description
              parentCode
              createdAt
              updatedAt
            }
          }
        `;
        variables = { code };
      }

      const data = await unifiedGraphQLClient.request<{
        organization: Partial<OrganizationUnit>;
      }>(graphqlQuery, variables);

      const organization = data.organization;
      if (!organization) {
        throw new Error(`组织 ${code} 不存在`);
      }

      // 简单数据转换，依赖后端格式
      if (safeTransform.graphqlToOrganization) {
        const transformed = safeTransform.graphqlToOrganization(organization) as unknown as OrganizationUnit;
        // 确保转换后的对象包含所有必需字段
        if (transformed && typeof transformed === 'object' && 'code' in transformed && 'name' in transformed) {
          return transformed as OrganizationUnit;
        }
      }
      
      return organization as OrganizationUnit;

    } catch (error: unknown) {
      console.error('Error fetching organization by code:', code, error);
      
      if (error instanceof Error && 'response' in error && 
          (error as Error & { response?: { status?: number } }).response?.status === 404) {
        throw new Error(`组织 ${code} 不存在`);
      }
      
      throw new Error(`获取组织 ${code} 失败，请重试`);
    }
  },

  // 获取组织统计信息 - 使用organizations查询获取统计数据
  getStats: async (): Promise<OrganizationStats> => {
    try {
      const graphqlQuery = `
        query GetOrganizationStats {
          organizationStats {
            totalCount
            activeCount
            inactiveCount
            plannedCount
            deletedCount
            byType {
              unitType
              count
            }
            byStatus {
              status
              count
            }
            byLevel {
              level
              count
            }
            temporalStats {
              totalVersions
              averageVersionsPerOrg
              oldestEffectiveDate
              newestEffectiveDate
            }
          }
        }
      `;

      const data = await unifiedGraphQLClient.request<{
        organizationStats: {
          totalCount: number;
          activeCount: number;
          inactiveCount: number;
          plannedCount: number;
          deletedCount: number;
          byType: Array<{ unitType: string; count: number }>;
          byStatus: Array<{ status: string; count: number }>;
          byLevel: Array<{ level: number; count: number }>;
          temporalStats: {
            totalVersions: number;
            averageVersionsPerOrg: number;
            oldestEffectiveDate: string;
            newestEffectiveDate: string;
          };
        };
      }>(graphqlQuery);

      const stats = data.organizationStats;
      
      if (!stats) {
        throw new Error('No statistics data returned');
      }

      // 转换为前端期望的格式
      const byType: Record<string, number> = {};
      const byStatus: Record<string, number> = {};
      
      stats.byType.forEach(item => {
        byType[item.unitType] = item.count;
      });
      
      stats.byStatus.forEach(item => {
        byStatus[item.status] = item.count;
      });

      return {
        totalCount: stats.totalCount,
        byType,
        byStatus,
        temporal: {
          current: stats.activeCount,
          future: stats.plannedCount,
          historical: stats.temporalStats.totalVersions
        },
        lastUpdated: new Date().toISOString()
      };

    } catch (error) {
      console.error('Error fetching organization stats:', error);
      throw new Error('Failed to fetch organization statistics. Please try again.');
    }
  },

  // 创建组织 - 依赖后端统一验证
  create: async (input: CreateOrganizationInput): Promise<OrganizationUnit> => {
    try {
      // 基础前端验证 (用户体验)
      const validationResult = validateOrganizationBasic(input);
      if (!validationResult.isValid) {
        throw new SimpleValidationError(
          '输入验证失败：' + formatValidationErrors(validationResult.errors), 
          validationResult.errors
        );
      }

      // 转换为API格式
      const apiData = safeTransform.cleanCreateInput(input);

      const response = await unifiedRESTClient.request<OrganizationUnit>('/organization-units', {
        method: 'POST',
        body: JSON.stringify(apiData),
      });
      
      // 简单的响应验证
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: unknown) {
      console.error('Error creating organization:', error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      if (error?.message?.includes('REST Error:')) {
        // 服务器端验证错误
        const serverMessage = error.message;
        throw new Error(serverMessage || 'Failed to create organization');
      }
      
      throw new Error('Failed to create organization. Please try again.');
    }
  },

  // 更新组织 - 智能验证，根据更新内容选择合适的验证策略
  update: async (code: string, input: UpdateOrganizationInput): Promise<OrganizationUnit> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      // 智能验证策略：根据更新的字段选择验证方法
      let validationResult;
      
      const inputKeys = Object.keys(input);
      const isStatusOnlyUpdate = inputKeys.length === 1 && inputKeys[0] === 'status';
      
      if (isStatusOnlyUpdate) {
        // 仅状态更新，使用状态专用验证
        console.log('[API] Status-only update detected, using validateStatusUpdate');
        validationResult = validateStatusUpdate(input);
      } else {
        // 完整更新，使用更新专用验证（不验证unitType）
        console.log('[API] Full update detected, using validateOrganizationUpdate');
        validationResult = validateOrganizationUpdate(input);
      }
      
      if (!validationResult.isValid) {
        throw new SimpleValidationError(
          '输入验证失败：' + formatValidationErrors(validationResult.errors),
          validationResult.errors
        );
      }

      // 转换为API格式
      const apiData = safeTransform.cleanUpdateInput(input);

      const response = await unifiedRESTClient.request<OrganizationUnit>(`/organization-units/${code}`, {
        method: 'PUT',
        body: JSON.stringify(apiData),
      });
      
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: unknown) {
      console.error('Error updating organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      if (error?.message?.includes('REST Error:')) {
        const serverMessage = error.message;
        throw new Error(serverMessage || 'Failed to update organization');
      }
      
      throw new Error('Failed to update organization. Please try again.');
    }
  },

  // 删除组织
  delete: async (code: string): Promise<void> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      await unifiedRESTClient.request<void>(`/organization-units/${code}`, {
        method: 'DELETE'
      });

    } catch (error) {
      console.error('Error deleting organization:', code, error);
      
      if (error && typeof error === 'object' && 'message' in error && typeof error.message === 'string' && error.message.includes('REST Error:')) {
        const serverMessage = error.message;
        throw new Error(serverMessage || 'Failed to delete organization');
      }
      
      throw new Error('Failed to delete organization. Please try again.');
    }
  },

  // ====== 组织详情API方法 ======

  // 获取组织的审计历史记录 - 集成新的审计API模块
  getAuditHistory: async (code: string, params?: TemporalQueryParams): Promise<Record<string, unknown>[]> => {
    try {
      // 导入审计API (动态导入以避免循环依赖)
      const { AuditAPI } = await import('./audit');
      
      // 将TemporalQueryParams转换为AuditQueryParams
      const auditParams = {
        startDate: params?.dateRange?.start,
        endDate: params?.dateRange?.end,
        limit: params?.limit || 50
      };

      const auditHistory = await AuditAPI.getOrganizationAuditHistory(code, auditParams);
      return auditHistory.auditTimeline || [];

    } catch (error) {
      console.error('Error fetching organization audit history:', code, error);
      throw new Error(`获取组织 ${code} 审计历史失败，请重试`);
    }
  },

  // 获取组织的时间线事件 - 已移除，使用temporal-graphql-client.ts中的实现

  // 创建时态组织记录
  createTemporal: async (input: CreateOrganizationInput & { 
    effectiveFrom: string; 
    effectiveTo?: string; 
    changeReason?: string;
  }): Promise<TemporalOrganizationUnit> => {
    try {
      // 基础前端验证
      const validationResult = validateOrganizationBasic(input);
      if (!validationResult.isValid) {
        throw new SimpleValidationError(
          '输入验证失败：' + formatValidationErrors(validationResult.errors), 
          validationResult.errors
        );
      }

        // 转换为API格式 - 修正字段命名为camelCase
        const apiData = {
          ...safeTransform.cleanCreateInput(input),
          effectiveDate: input.effectiveFrom, // 修正：使用camelCase
          endDate: input.effectiveTo,      // 修正：使用camelCase
          operationReason: input.changeReason, // 修正：使用camelCase
          isTemporal: true
        };

      const response = await unifiedRESTClient.request<TemporalOrganizationUnit>('/organization-units/temporal', {
        method: 'POST',
        body: JSON.stringify(apiData),
      });
      
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: unknown) {
      console.error('Error creating temporal organization:', error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to create temporal organization. Please try again.');
    }
  },

  // 更新时态组织记录 - 使用事件驱动API
  updateTemporal: async (code: string, input: UpdateOrganizationInput & {
    effectiveDate?: string;
    endDate?: string;
    changeReason?: string;
  }): Promise<TemporalOrganizationUnit> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      // 智能验证策略
      const validationResult = validateOrganizationUpdate(input);
      if (!validationResult.isValid) {
        throw new SimpleValidationError(
          '输入验证失败：' + formatValidationErrors(validationResult.errors),
          validationResult.errors
        );
      }

      // 转换为事件驱动API格式 - 修正为camelCase命名
      const eventData = {
        eventType: "UPDATE", // 修正：使用camelCase
        effectiveDate: input.effectiveDate ? new Date(input.effectiveDate + 'T00:00:00Z').toISOString() : new Date().toISOString(),
        endDate: input.endDate ? new Date(input.endDate + 'T00:00:00Z').toISOString() : null,
        changeData: safeTransform.cleanUpdateInput(input), // 修正：使用camelCase
        operationReason: input.changeReason || "组织信息更新" // 修正：使用camelCase
      };

      // 使用事件驱动端点
      const response = await unifiedRESTClient.request<TemporalOrganizationUnit>(`/organization-units/${code}/events`, {
        method: 'POST',
        body: JSON.stringify(eventData),
      });
      
      // 验证响应是否有效 - 修正：检查核心字段而非event_id
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: unknown) {
      console.error('Error updating temporal organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to update temporal organization. Please try again.');
    }
  },

  // === 新增：操作驱动状态管理API ===

  // 停用组织
  suspend: async (code: string, reason: string): Promise<OrganizationUnit> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      if (!reason || !reason.trim()) {
        throw new SimpleValidationError('Suspend reason is required', [
          { field: 'reason', message: 'Reason is required' }
        ]);
      }

      const response = await unifiedRESTClient.request<OrganizationUnit>(`/organization-units/${code}/suspend`, {
        method: 'POST',
        body: JSON.stringify({ reason: reason.trim() }),
      });
      
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: unknown) {
      console.error('Error suspending organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to suspend organization. Please try again.');
    }
  },

  // 重新启用组织
  reactivate: async (code: string, reason: string): Promise<OrganizationUnit> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      if (!reason || !reason.trim()) {
        throw new SimpleValidationError('Reactivate reason is required', [
          { field: 'reason', message: 'Reason is required' }
        ]);
      }

      const response = await unifiedRESTClient.request<OrganizationUnit>(`/organization-units/${code}/reactivate`, {
        method: 'POST',
        body: JSON.stringify({ reason: reason.trim() }),
      });
      
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: unknown) {
      console.error('Error reactivating organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to reactivate organization. Please try again.');
    }
  }
};

// 导出简化的API
export default organizationAPI;