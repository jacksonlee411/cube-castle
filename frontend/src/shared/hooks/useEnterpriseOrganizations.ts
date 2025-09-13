/**
 * 企业级组织单元管理Hook
 * 集成企业级响应信封格式和错误处理
 * 支持后端P0级响应格式统一修复
 */

import { useState, useEffect, useCallback } from 'react';
import type { 
  OrganizationUnit, 
  OrganizationListResponse, 
  OrganizationQueryParams
} from '../types';

// 临时类型定义，待完善
interface OrganizationStats {
  total: number;
  active: number;
  inactive: number;
}
import type { APIResponse } from '../types/api';
import type { TemporalQueryParams } from '../types/temporal';
// TODO: Replace with proper API implementation

// 使用统一的组织查询参数接口，无需重复定义
// OrganizationQueryParams 已包含所有必要字段

/**
 * 企业级组织单元管理Hook
 * 自动处理企业级响应信封格式
 */
export const useEnterpriseOrganizations = (
  initialParams?: OrganizationQueryParams
) => {
  // 状态管理
  const [state, setState] = useState<{
    organizations: OrganizationUnit[];
    totalCount: number;
    page: number;
    pageSize: number;
    totalPages: number;
    loading: boolean;
    error: string | null;
    stats: OrganizationStats | null;
  }>({
    organizations: [],
    totalCount: 0,
    page: 1,
    pageSize: 50,
    totalPages: 0,
    loading: false,
    error: null,
    stats: null
  });

  // 获取组织列表
  const fetchOrganizations = useCallback(async (
    params?: OrganizationQueryParams
  ): Promise<APIResponse<OrganizationListResponse>> => {
    setState(prev => ({ ...prev, loading: true, error: null }));
    
    try {
      // 构建GraphQL查询
      const query = `
        query GetOrganizations($filter: OrganizationFilter, $pagination: PaginationInput) {
          organizations(filter: $filter, pagination: $pagination) {
            data {
              code
              parentCode
              tenantId
              name
              unitType
              status
              level
              sortOrder
              description
              profile
              effectiveDate
              endDate
              createdAt
              updatedAt
              recordId
              isFuture
              hierarchyDepth
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

      // 构建查询变量
      const variables: any = {};
      
      if (params) {
        variables.filter = {};
        variables.pagination = {};
        
        // 映射查询参数
        if (params.unitType) variables.filter.unitType = params.unitType;
        if (params.status) variables.filter.status = params.status;
        if (params.parentCode) variables.filter.parentCode = params.parentCode;
        if (params.level) variables.filter.level = params.level;
        if (params.searchText) variables.filter.searchText = params.searchText;
        if (params.asOfDate) variables.filter.asOfDate = params.asOfDate;
        
        // 分页参数
        if (params.page) variables.pagination.page = params.page;
        if (params.pageSize) variables.pagination.pageSize = params.pageSize;
        if (params.sortBy) variables.pagination.sortBy = params.sortBy;
        if (params.sortOrder) variables.pagination.sortOrder = params.sortOrder;
      }

      // 使用统一的GraphQL客户端
      const { unifiedGraphQLClient } = await import('../api/unified-client');
      const graphqlData = await unifiedGraphQLClient.request<{
        organizations: {
          data: OrganizationUnit[];
          pagination: {
            total: number;
            page: number;
            pageSize: number;
            hasNext: boolean;
            hasPrevious: boolean;
          };
          temporal: {
            asOfDate: string;
            currentCount: number;
            futureCount: number;
            historicalCount: number;
          };
        };
      }>(query, variables);
      
      if (graphqlData?.organizations) {
        const orgData = graphqlData.organizations;
        
        const response: APIResponse<OrganizationListResponse> = {
          success: true,
          data: {
            organizations: orgData.data || [],
            totalCount: orgData.pagination?.total || 0,
            page: orgData.pagination?.page || 1,
            pageSize: orgData.pagination?.pageSize || 50,
            totalPages: Math.ceil((orgData.pagination?.total || 0) / (orgData.pagination?.pageSize || 50))
          },
          timestamp: new Date().toISOString()
        };

        setState(prev => ({
          ...prev,
          organizations: response.data!.organizations,
          totalCount: response.data!.totalCount,
          page: response.data!.page,
          pageSize: response.data!.pageSize,
          totalPages: response.data!.totalPages,
          loading: false,
          error: null,
          lastUpdate: response.timestamp
        }));
        
        return response;
      } else {
        const errorMessage = '获取组织列表失败：无数据返回';
        setState(prev => ({ 
          ...prev, 
          loading: false, 
          error: errorMessage
        }));
        
        return {
          success: false,
          error: {
            code: 'GRAPHQL_ERROR',
            message: errorMessage
          },
          timestamp: new Date().toISOString()
        };
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '网络请求失败';
      setState(prev => ({ ...prev, loading: false, error: errorMessage }));
      
      return {
        success: false,
        error: {
          code: 'HOOK_ERROR',
          message: errorMessage,
          details: error
        },
        timestamp: new Date().toISOString()
      };
    }
  }, []);

  // 根据代码获取单个组织
  const fetchOrganizationByCode = useCallback(async (
    code: string, 
    temporalParams?: TemporalQueryParams
  ): Promise<APIResponse<OrganizationUnit>> => {
    setState(prev => ({ ...prev, loading: true, error: null }));
    
    try {
      // 构建GraphQL查询
      const query = `
        query GetOrganization($code: String!, $asOfDate: String) {
          organization(code: $code, asOfDate: $asOfDate) {
            code
            parentCode
            tenantId
            name
            unitType
            status
            level
            sortOrder
            description
            profile
            effectiveDate
            endDate
            createdAt
            updatedAt
            recordId
            isFuture
            hierarchyDepth
          }
        }
      `;

      // 构建查询变量
      const variables: any = { code };
      if (temporalParams?.asOfDate) {
        variables.asOfDate = temporalParams.asOfDate;
      }

      // 使用统一的GraphQL客户端
      const { unifiedGraphQLClient } = await import('../api/unified-client');
      const graphqlData = await unifiedGraphQLClient.request<{
        organization: OrganizationUnit | null;
      }>(query, variables);
      
      if (graphqlData?.organization) {
        const response: APIResponse<OrganizationUnit> = {
          success: true,
          data: graphqlData.organization,
          timestamp: new Date().toISOString()
        };

        setState(prev => ({ 
          ...prev, 
          loading: false, 
          error: null,
          lastUpdate: response.timestamp
        }));
        
        return response;
      } else {
        const errorMessage = '获取组织失败：组织不存在或无权限访问';
        setState(prev => ({ 
          ...prev, 
          loading: false, 
          error: errorMessage
        }));
        
        return {
          success: false,
          error: {
            code: 'GRAPHQL_ERROR',
            message: errorMessage
          },
          timestamp: new Date().toISOString()
        };
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '网络请求失败';
      setState(prev => ({ ...prev, loading: false, error: errorMessage }));
      
      return {
        success: false,
        error: {
          code: 'HOOK_ERROR',
          message: errorMessage,
          details: error
        },
        timestamp: new Date().toISOString()
      };
    }
  }, []);

  // 获取统计信息
  const fetchStats = useCallback(async (asOfDate?: string): Promise<APIResponse<OrganizationStats>> => {
    try {
      // 构建GraphQL查询
      const query = `
        query GetOrganizationStats($asOfDate: String, $includeHistorical: Boolean) {
          organizationStats(asOfDate: $asOfDate, includeHistorical: $includeHistorical) {
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

      // 构建查询变量
      const variables: any = { 
        includeHistorical: false 
      };
      if (asOfDate) {
        variables.asOfDate = asOfDate;
      }

      // 使用统一的GraphQL客户端
      const { unifiedGraphQLClient } = await import('../api/unified-client');
      const graphqlData = await unifiedGraphQLClient.request<{
        organizationStats: {
          totalCount: number;
          activeCount: number;
          inactiveCount: number;
          plannedCount: number;
          deletedCount: number;
          byType: Array<{ unitType: string; count: number; }>;
          byStatus: Array<{ status: string; count: number; }>;
          byLevel: Array<{ level: number; count: number; }>;
          temporalStats: {
            totalVersions: number;
            averageVersionsPerOrg: number;
            oldestEffectiveDate: string;
            newestEffectiveDate: string;
          };
        };
      }>(query, variables);
      
      if (graphqlData?.organizationStats) {
        const statsData = graphqlData.organizationStats;
        
        // 映射到本地类型
        const mappedStats: OrganizationStats = {
          total: statsData.totalCount || 0,
          active: statsData.activeCount || 0,
          inactive: statsData.inactiveCount || 0
        };

        const response: APIResponse<OrganizationStats> = {
          success: true,
          data: mappedStats,
          timestamp: new Date().toISOString()
        };

        setState(prev => ({
          ...prev,
          stats: response.data!,
          lastUpdate: response.timestamp
        }));
        
        return response;
      } else {
        const errorMessage = '获取统计信息失败：无数据返回';
        
        return {
          success: false,
          error: {
            code: 'GRAPHQL_ERROR',
            message: errorMessage
          },
          timestamp: new Date().toISOString()
        };
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '获取统计信息失败';
      
      return {
        success: false,
        error: {
          code: 'HOOK_ERROR',
          message: errorMessage,
          details: error
        },
        timestamp: new Date().toISOString()
      };
    }
  }, []);

  // 刷新数据
  const refreshData = useCallback(() => {
    if (initialParams) {
      fetchOrganizations(initialParams);
    }
    fetchStats();
  }, [initialParams, fetchOrganizations, fetchStats]);

  // 清除错误
  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  // 初始化加载
  useEffect(() => {
    // 总是调用fetchOrganizations获取组织列表，如果没有参数则使用默认参数
    fetchOrganizations(initialParams);
    fetchStats();
  }, [initialParams, fetchOrganizations, fetchStats]);

  return {
    // 状态
    ...state,
    
    // 操作
    fetchOrganizations,
    fetchOrganizationByCode,
    fetchStats,
    refreshData,
    clearError
  };
};

// DEPRECATED: useOrganizationList 是不必要的重复代码
// 直接使用 useEnterpriseOrganizations，它已经包含所有相同功能
// 如需 refresh 方法，直接使用 result.fetchOrganizations(params)

// TODO-TEMPORARY: 该Hook将在 2025-09-16 后删除，所有使用应替换为 useEnterpriseOrganizations


export default useEnterpriseOrganizations;