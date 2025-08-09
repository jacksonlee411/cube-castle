import type { 
  OrganizationUnit, 
  OrganizationListResponse, 
  OrganizationStats,
  GraphQLResponse,
  OrganizationUnitType,
  OrganizationStatus
} from '../types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../hooks/useOrganizationMutations';
import { 
  validateOrganizationBasic,
  safeTransform,
  SimpleValidationError,
  formatValidationErrors
} from '../validation/simple-validation';

// GraphQL客户端 - 使用正确的端口8090
const GRAPHQL_ENDPOINT = 'http://localhost:8090/graphql';

const graphqlClient = {
  async request<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
    const response = await fetch(GRAPHQL_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        variables
      }),
    });

    if (!response.ok) {
      throw new Error(`GraphQL Error: ${response.status} ${response.statusText}`);
    }

    const result = await response.json() as GraphQLResponse<T>;
    
    if (result.errors) {
      throw new Error(`GraphQL Error: ${result.errors[0].message}`);
    }

    if (!result.data) {
      throw new Error('No data returned from GraphQL');
    }

    return result.data;
  }
};

// REST API客户端 - 使用命令服务端口9090
const REST_ENDPOINT = 'http://localhost:9090/api/v1';

const restClient = {
  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${REST_ENDPOINT}${endpoint}`;
    
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        'X-Tenant-ID': '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
        ...options.headers,
      },
      ...options,
    });

    if (!response.ok) {
      throw new Error(`REST Error: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }
};

export interface OrganizationQueryParams {
  searchText?: string | undefined;
  unit_type?: OrganizationUnitType | undefined;
  status?: OrganizationStatus | undefined;
  level?: number | undefined;
  page?: number;
  pageSize?: number;
}

export const organizationAPI = {
  // 获取组织单元列表 - 使用GraphQL (修复getByCode问题)
  getAll: async (params?: OrganizationQueryParams): Promise<OrganizationListResponse> => {
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

      // 构建GraphQL查询和变量
      const graphqlQuery = `
        query GetOrganizations($first: Int, $offset: Int, $searchText: String) {
          organizations(first: $first, offset: $offset, searchText: $searchText) {
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
          organizationStats {
            totalCount
          }
        }
      `;

      const variables = {
        first: params?.pageSize || 50,
        offset: ((params?.page || 1) - 1) * (params?.pageSize || 50),
        searchText: params?.searchText || null
      };

      const data = await graphqlClient.request<{
        organizations: any[];
        organizationStats: { totalCount: number };
      }>(graphqlQuery, variables);

      // 简化的数据转换 - 无需复杂的Zod验证
      const organizations = data.organizations.map((org: any) => {
        try {
          return safeTransform.graphqlToOrganization ? 
            safeTransform.graphqlToOrganization(org) : 
            org; // 直接返回原始数据，依赖后端格式
        } catch (error) {
          console.warn('Failed to transform organization:', org, error);
          return null;
        }
      }).filter(Boolean);

      return {
        organizations: organizations,
        total_count: data.organizationStats.totalCount,
        page: params?.page || 1,
        page_size: organizations.length,
        total_pages: Math.ceil(data.organizationStats.totalCount / (params?.pageSize || 50))
      };

    } catch (error) {
      console.error('Error fetching organizations:', error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to fetch organizations. Please try again.');
    }
  },

  // 根据代码获取单个组织 - ✅ 修复协议违反，统一使用GraphQL
  getByCode: async (code: string): Promise<OrganizationUnit> => {
    try {
      if (!code || typeof code !== 'string') {
        throw new SimpleValidationError('Invalid organization code', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      // ✅ 使用GraphQL查询，遵循"查询统一用GraphQL"原则
      const graphqlQuery = `
        query GetOrganization($code: String!) {
          organization(code: $code) {
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
        }
      `;

      const data = await graphqlClient.request<{
        organization: any;
      }>(graphqlQuery, { code });

      const organization = data.organization;
      if (!organization) {
        throw new Error(`组织 ${code} 不存在`);
      }

      // 简单数据转换，依赖后端格式
      return safeTransform.graphqlToOrganization ? 
        safeTransform.graphqlToOrganization(organization) : 
        organization;

    } catch (error) {
      console.error('Error fetching organization by code:', code, error);
      
      if (error.response?.status === 404) {
        throw new Error(`组织 ${code} 不存在`);
      }
      
      throw new Error(`获取组织 ${code} 失败，请重试`);
    }
  },

  // 获取组织统计信息 - 使用GraphQL
  getStats: async (): Promise<OrganizationStats> => {
    try {
      const graphqlQuery = `
        query GetOrganizationStats {
          organizationStats {
            totalCount
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
          }
        }
      `;

      const data = await graphqlClient.request<{
        organizationStats: any;
      }>(graphqlQuery);

      const stats = data.organizationStats;
      if (!stats) {
        throw new Error('No statistics data returned');
      }

      // 简化的数据转换
      return {
        total: stats.totalCount || 0,
        by_type: stats.byType?.reduce((acc: any, item: any) => {
          acc[item.unitType] = item.count;
          return acc;
        }, {}) || {},
        by_status: stats.byStatus?.reduce((acc: any, item: any) => {
          acc[item.status] = item.count;
          return acc;
        }, {}) || {},
        by_level: stats.byLevel?.reduce((acc: any, item: any) => {
          acc[item.level] = item.count;
          return acc;
        }, {}) || {}
      };

    } catch (error) {
      console.error('Error fetching organization stats:', error);
      throw new Error('Failed to fetch organization statistics. Please try again.');
    }
  },

  // 创建组织 - 依赖后端统一验证
  create: async (input: CreateOrganizationInput): Promise<any> => {
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

      const response = await restClient.request<any>('/organization-units', {
        method: 'POST',
        body: JSON.stringify(apiData),
      });
      
      // 简单的响应验证
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: any) {
      console.error('Error creating organization:', error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      if (error.message?.includes('REST Error:')) {
        // 服务器端验证错误
        const serverMessage = error.message;
        throw new Error(serverMessage || 'Failed to create organization');
      }
      
      throw new Error('Failed to create organization. Please try again.');
    }
  },

  // 更新组织 - 依赖后端统一验证
  update: async (code: string, input: UpdateOrganizationInput): Promise<any> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      // 基础前端验证 (用户体验)
      const validationResult = validateOrganizationBasic(input);
      if (!validationResult.isValid) {
        throw new SimpleValidationError(
          '输入验证失败：' + formatValidationErrors(validationResult.errors),
          validationResult.errors
        );
      }

      // 转换为API格式
      const apiData = safeTransform.cleanUpdateInput(input);

      const response = await restClient.request<any>(`/organization-units/${code}`, {
        method: 'PUT',
        body: JSON.stringify(apiData),
      });
      
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: any) {
      console.error('Error updating organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      if (error.message?.includes('REST Error:')) {
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
        throw new SimpleValidationError('Organization code is required', { code: 'Code is required' });
      }

      await restClient.request<void>(`/organization-units/${code}`, {
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
  }
};

// 导出简化的API
export default organizationAPI;