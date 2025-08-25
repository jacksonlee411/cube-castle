/**
 * 企业级组织单元API客户端
 * 集成GraphQL企业级响应信封适配器
 * 支持后端P0级响应格式统一修复
 */

import type { 
  OrganizationUnit, 
  OrganizationListResponse, 
  OrganizationStats,
  OrganizationQueryParams
} from '../types';
import type { 
  TemporalQueryParams
} from '../types/temporal';
import type { APIResponse } from '../types/api';
import { safeTransform } from '../validation/simple-validation';
import { graphqlEnterpriseAdapter } from './graphql-enterprise-adapter';

// 扩展查询参数以支持时态查询
interface ExtendedOrganizationQueryParams extends OrganizationQueryParams {
  searchText?: string;
  pageSize?: number;
  temporalParams?: TemporalQueryParams;
}

/**
 * 企业级组织单元API客户端
 * 自动适配GraphQL企业级响应信封格式
 */
export const enterpriseOrganizationAPI = {
  /**
   * 获取组织单元列表 - 使用企业级GraphQL适配器
   */
  getAll: async (params?: ExtendedOrganizationQueryParams): Promise<APIResponse<OrganizationListResponse>> => {
    try {
      // 轻量级参数验证
      if (params) {
        if (params.page && params.page < 1) {
          return {
            success: false,
            error: {
              code: 'INVALID_PARAMETER',
              message: '页码必须大于0',
              details: { field: 'page', value: params.page }
            },
            timestamp: new Date().toISOString()
          };
        }
        if (params.pageSize && (params.pageSize < 1 || params.pageSize > 100)) {
          return {
            success: false,
            error: {
              code: 'INVALID_PARAMETER', 
              message: '页面大小必须在1-100之间',
              details: { field: 'pageSize', value: params.pageSize }
            },
            timestamp: new Date().toISOString()
          };
        }
      }

      // 构建GraphQL查询
      const useTemporalQuery = params?.temporalParams && Object.keys(params.temporalParams).length > 0;
      
      let graphqlQuery, variables;
      
      if (useTemporalQuery) {
        // 时态查询版本
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
                name
                unitType
                status
                isDeleted
                level
                hierarchyDepth
                codePath
                namePath
                sortOrder
                description
                profile
                effectiveDate
                endDate
                isCurrent
                isFuture
                createdAt
                updatedAt
                operationType
                operatedBy {
                  id
                  name
                }
                operationReason
                recordId
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
                historicalCount
                futureCount
              }
            }
          }
        `;
        
        variables = {
          filter: {
            unitType: params?.unitType,
            status: params?.status,
            level: params?.level,
            searchText: params?.searchText,
            temporal: params?.temporalParams
          },
          pagination: {
            page: params?.page || 1,
            pageSize: params?.pageSize || 50
          }
        };
      } else {
        // 基础查询版本
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
                name
                unitType
                status
                isDeleted
                level
                hierarchyDepth
                codePath
                namePath
                sortOrder
                description
                profile
                effectiveDate
                endDate
                isCurrent
                isFuture
                createdAt
                updatedAt
                operationType
                operatedBy {
                  id
                  name
                }
                operationReason
                recordId
              }
              pagination {
                total
                page
                pageSize
                hasNext
                hasPrevious
              }
            }
          }
        `;
        
        variables = {
          filter: {
            unitType: params?.unitType,
            status: params?.status,
            level: params?.level,
            searchText: params?.searchText
          },
          pagination: {
            page: params?.page || 1,
            pageSize: params?.pageSize || 50
          }
        };
      }

      // 使用企业级GraphQL适配器发送请求
      const response = await graphqlEnterpriseAdapter.request<{
        organizations: {
          data: Array<Partial<OrganizationUnit>>;
          pagination: {
            total: number;
            page: number;
            pageSize: number;
            hasNext: boolean;
            hasPrevious: boolean;
          };
          temporal?: {
            asOfDate: string;
            currentCount: number;
            historicalCount: number;
            futureCount: number;
          };
        };
      }>(graphqlQuery, variables);

      // 检查响应成功性
      if (!response.success || !response.data) {
        return {
          success: false,
          error: response.error || {
            code: 'NO_DATA',
            message: '未获取到组织数据'
          },
          timestamp: response.timestamp || new Date().toISOString(),
          requestId: response.requestId
        };
      }

      // 转换数据格式
      const organizations = (response.data.organizations?.data || []).map((org: Partial<OrganizationUnit>) => {
        try {
          return safeTransform.graphqlToOrganization ? 
            safeTransform.graphqlToOrganization(org) : 
            org;
        } catch (error) {
          console.warn('Failed to transform organization:', org, error);
          return null;
        }
      }).filter(Boolean);

      const totalCount = response.data.organizations?.pagination?.total || 0;
      
      const result: OrganizationListResponse = {
        organizations: organizations.filter((org): org is OrganizationUnit => org !== null),
        totalCount: totalCount,
        page: params?.page || 1,
        pageSize: organizations.length,
        totalPages: Math.ceil(totalCount / (params?.pageSize || 50))
      };

      return {
        success: true,
        data: result,
        message: `成功获取 ${result.organizations.length} 个组织单元`,
        timestamp: response.timestamp,
        requestId: response.requestId
      };

    } catch (error) {
      console.error('Error fetching organizations:', error);
      
      return {
        success: false,
        error: {
          code: 'FETCH_ERROR',
          message: error instanceof Error ? error.message : '获取组织列表失败',
          details: error
        },
        timestamp: new Date().toISOString()
      };
    }
  },

  /**
   * 根据代码获取单个组织 - 企业级响应格式
   */
  getByCode: async (code: string, temporalParams?: TemporalQueryParams): Promise<APIResponse<OrganizationUnit>> => {
    try {
      if (!code || typeof code !== 'string') {
        return {
          success: false,
          error: {
            code: 'INVALID_CODE',
            message: '组织代码不能为空',
            details: { code }
          },
          timestamp: new Date().toISOString()
        };
      }

      const useTemporalQuery = temporalParams && Object.keys(temporalParams).length > 0;
      
      let graphqlQuery, variables;
      
      if (useTemporalQuery) {
        graphqlQuery = `
          query GetOrganizationByCode($code: String!, $temporal: TemporalInput) {
            organization(code: $code, temporal: $temporal) {
              code
              parentCode
              tenantId
              name
              unitType
              status
              isDeleted
              level
              hierarchyDepth
              codePath
              namePath
              sortOrder
              description
              profile
              effectiveDate
              endDate
              isCurrent
              isFuture
              createdAt
              updatedAt
              operationType
              operatedBy {
                id
                name
              }
              operationReason
              recordId
            }
          }
        `;
        variables = { code, temporal: temporalParams };
      } else {
        graphqlQuery = `
          query GetOrganizationByCode($code: String!) {
            organization(code: $code) {
              code
              parentCode
              tenantId
              name
              unitType
              status
              isDeleted
              level
              hierarchyDepth
              codePath
              namePath
              sortOrder
              description
              profile
              effectiveDate
              endDate
              isCurrent
              isFuture
              createdAt
              updatedAt
              operationType
              operatedBy {
                id
                name
              }
              operationReason
              recordId
            }
          }
        `;
        variables = { code };
      }

      const response = await graphqlEnterpriseAdapter.request<{
        organization: Partial<OrganizationUnit>;
      }>(graphqlQuery, variables);

      if (!response.success || !response.data) {
        return {
          success: false,
          error: response.error || {
            code: 'NOT_FOUND',
            message: `未找到组织单元: ${code}`
          },
          timestamp: response.timestamp || new Date().toISOString(),
          requestId: response.requestId
        };
      }

      // 数据转换和验证
      try {
        const organization = safeTransform.graphqlToOrganization ? 
          safeTransform.graphqlToOrganization(response.data.organization) as unknown as OrganizationUnit : 
          response.data.organization as unknown as OrganizationUnit;

        return {
          success: true,
          data: organization,
          message: `成功获取组织单元: ${organization.name}`,
          timestamp: response.timestamp,
          requestId: response.requestId
        };
      } catch (transformError) {
        return {
          success: false,
          error: {
            code: 'TRANSFORM_ERROR',
            message: '组织数据格式转换失败',
            details: transformError
          },
          timestamp: response.timestamp || new Date().toISOString(),
          requestId: response.requestId
        };
      }

    } catch (error) {
      console.error('Error fetching organization by code:', error);
      
      return {
        success: false,
        error: {
          code: 'FETCH_ERROR',
          message: error instanceof Error ? error.message : '获取组织失败',
          details: error
        },
        timestamp: new Date().toISOString()
      };
    }
  },

  /**
   * 获取组织统计信息 - 企业级响应格式
   */
  getStats: async (): Promise<APIResponse<OrganizationStats>> => {
    try {
      const graphqlQuery = `
        query GetOrganizationStats {
          organizationStats {
            totalCount
            temporalStats {
              totalVersions
              averageVersionsPerOrg
              oldestEffectiveDate
              newestEffectiveDate
            }
            byType {
              unitType
              count
            }
            byStatus {
              status
              count
            }
          }
        }
      `;

      const response = await graphqlEnterpriseAdapter.request<{
        organizationStats: OrganizationStats;
      }>(graphqlQuery);

      if (!response.success || !response.data) {
        return {
          success: false,
          error: response.error || {
            code: 'NO_STATS',
            message: '未能获取统计信息'
          },
          timestamp: response.timestamp || new Date().toISOString(),
          requestId: response.requestId
        };
      }

      return {
        success: true,
        data: response.data.organizationStats,
        message: '成功获取组织统计信息',
        timestamp: response.timestamp,
        requestId: response.requestId
      };

    } catch (error) {
      console.error('Error fetching organization stats:', error);
      
      return {
        success: false,
        error: {
          code: 'STATS_ERROR',
          message: error instanceof Error ? error.message : '获取统计信息失败',
          details: error
        },
        timestamp: new Date().toISOString()
      };
    }
  }
};

// 导出企业级API客户端作为默认导出
export default enterpriseOrganizationAPI;