import { apiClient } from './client';
import type { 
  OrganizationUnit, 
  OrganizationListResponse, 
  OrganizationStats, 
  APIResponse,
  GraphQLOrganizationResponse,
  CreateOrganizationResponse,
  UpdateOrganizationResponse,
  GraphQLResponse,
  GraphQLVariables,
  OrganizationUnitType,
  OrganizationStatus
} from '../types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../hooks/useOrganizationMutations';
import { 
  orgValidation,
  safeTransform,
  SimpleValidationError,
  isSimpleValidationError,
  isAPIError,
  isNetworkError
} from '../validation/simple-validation';

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
        const validationResult = orgValidation.validateQueryParams(params);
        if (!validationResult.isValid) {
          throw new SimpleValidationError('Invalid query parameters', validationResult.errors);
        }
      }

      // 构建GraphQL查询
      const graphqlQuery = `
        query GetOrganizations($first: Int, $offset: Int) {
          organizations(first: $first, offset: $offset) {
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
        offset: ((params?.page || 1) - 1) * (params?.pageSize || 50)
      };

      const response = await apiClient.post<GraphQLResponse<{
        organizations: any[];
        organizationStats: { totalCount: number };
      }>>('/graphql', {
        query: graphqlQuery,
        variables
      });

      if (response.data.errors) {
        throw new Error(`GraphQL Error: ${response.data.errors[0].message}`);
      }

      const data = response.data.data;
      if (!data) {
        throw new Error('No data returned from GraphQL');
      }

      // 简化的数据转换 - 无需复杂的Zod验证
      const organizations = data.organizations.map((org: any) => {
        try {
          return safeTransform.graphqlToOrganization(org);
        } catch (error) {
          console.warn('Failed to transform organization:', org, error);
          return null;
        }
      }).filter(Boolean);

      return {
        data: organizations,
        total: data.organizationStats.totalCount,
        page: params?.page || 1,
        pageSize: organizations.length,
        totalPages: Math.ceil(data.organizationStats.totalCount / (params?.pageSize || 50))
      };

    } catch (error) {
      console.error('Error fetching organizations:', error);
      
      if (isSimpleValidationError(error)) {
        throw error;
      }
      if (isNetworkError(error)) {
        throw new Error('Network connection failed. Please check your internet connection.');
      }
      if (isAPIError(error)) {
        throw new Error(`Server error: ${error.response?.data?.message || 'Unknown error'}`);
      }
      
      throw new Error('Failed to fetch organizations. Please try again.');
    }
  },

  // 根据代码获取单个组织 - 使用GraphQL而非REST
  getByCode: async (code: string): Promise<OrganizationUnit> => {
    try {
      if (!code || typeof code !== 'string') {
        throw new SimpleValidationError('Invalid organization code', { code: 'Code is required' });
      }

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

      const response = await apiClient.post<GraphQLResponse<{
        organization: any;
      }>>('/graphql', {
        query: graphqlQuery,
        variables: { code }
      });

      if (response.data.errors) {
        throw new Error(`GraphQL Error: ${response.data.errors[0].message}`);
      }

      const data = response.data.data;
      if (!data?.organization) {
        throw new Error(`Organization not found: ${code}`);
      }

      return safeTransform.graphqlToOrganization(data.organization);

    } catch (error) {
      console.error('Error fetching organization by code:', code, error);
      
      if (isSimpleValidationError(error)) {
        throw error;
      }
      
      throw new Error(`Failed to fetch organization ${code}. Please try again.`);
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

      const response = await apiClient.post<GraphQLResponse<{
        organizationStats: any;
      }>>('/graphql', {
        query: graphqlQuery
      });

      if (response.data.errors) {
        throw new Error(`GraphQL Error: ${response.data.errors[0].message}`);
      }

      const stats = response.data.data?.organizationStats;
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
  create: async (input: CreateOrganizationInput): Promise<CreateOrganizationResponse> => {
    try {
      // 基础前端验证 (用户体验)
      const validationResult = orgValidation.validateCreateInput(input);
      if (!validationResult.isValid) {
        throw new SimpleValidationError('Validation failed', validationResult.errors);
      }

      // 转换为API格式
      const apiData = safeTransform.createInputToAPI(input);

      const response = await apiClient.post<CreateOrganizationResponse>('/api/v1/organization-units', apiData);
      
      // 简单的响应验证
      if (!response.data?.code) {
        throw new Error('Invalid response from server');
      }

      return response.data;

    } catch (error) {
      console.error('Error creating organization:', error);
      
      if (isSimpleValidationError(error)) {
        throw error;
      }
      if (isAPIError(error)) {
        // 服务器端验证错误
        const serverMessage = error.response?.data?.message || error.response?.data?.error;
        throw new Error(serverMessage || 'Failed to create organization');
      }
      
      throw new Error('Failed to create organization. Please try again.');
    }
  },

  // 更新组织 - 依赖后端统一验证
  update: async (code: string, input: UpdateOrganizationInput): Promise<UpdateOrganizationResponse> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', { code: 'Code is required' });
      }

      // 基础前端验证 (用户体验)
      const validationResult = orgValidation.validateUpdateInput(input);
      if (!validationResult.isValid) {
        throw new SimpleValidationError('Validation failed', validationResult.errors);
      }

      // 转换为API格式
      const apiData = safeTransform.updateInputToAPI(input);

      const response = await apiClient.put<UpdateOrganizationResponse>(`/api/v1/organization-units/${code}`, apiData);
      
      if (!response.data?.code) {
        throw new Error('Invalid response from server');
      }

      return response.data;

    } catch (error) {
      console.error('Error updating organization:', code, error);
      
      if (isSimpleValidationError(error)) {
        throw error;
      }
      if (isAPIError(error)) {
        const serverMessage = error.response?.data?.message || error.response?.data?.error;
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

      await apiClient.delete(`/api/v1/organization-units/${code}`);

    } catch (error) {
      console.error('Error deleting organization:', code, error);
      
      if (isAPIError(error)) {
        const serverMessage = error.response?.data?.message || error.response?.data?.error;
        throw new Error(serverMessage || 'Failed to delete organization');
      }
      
      throw new Error('Failed to delete organization. Please try again.');
    }
  }
};

// 导出简化的API
export default organizationAPI;