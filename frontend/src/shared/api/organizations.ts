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
  validateGraphQLVariables,
  validateGraphQLOrganizationList,
  validateGraphQLOrganizationResponse,
  validateOrganizationUnit,
  validateCreateOrganizationResponse,
  validateUpdateOrganizationInput,
  safeTransformGraphQLToOrganizationUnit,
  safeTransformCreateInputToAPI,
  isGraphQLError,
  isGraphQLSuccessResponse,
  ValidationError,
  isValidationError,
  isAPIError,
  isNetworkError
} from './type-guards';

export interface OrganizationQueryParams {
  searchText?: string | undefined;
  unit_type?: OrganizationUnitType | undefined;
  status?: OrganizationStatus | undefined;
  level?: number | undefined;
  page?: number;
  pageSize?: number;
}

export const organizationAPI = {
  // 获取组织单元列表
  getAll: async (params?: OrganizationQueryParams): Promise<OrganizationListResponse> => {
    try {
      // 构建并验证GraphQL查询变量
      const variables: GraphQLVariables = {};
      
      if (params) {
        // 验证输入参数
        const validatedParams = validateGraphQLVariables(params);
        Object.assign(variables, validatedParams);
      }
    let queryArgs = '';
    
    if (params) {
      const queryParts: string[] = [];
      
      if (params.searchText) {
        queryParts.push('$searchText: String');
        variables.searchText = params.searchText;
      }
      
      if (params.unit_type) {
        queryParts.push('$unitType: String');
        variables.unitType = params.unit_type;
      }
      
      if (params.status) {
        queryParts.push('$status: String');
        variables.status = params.status;
      }
      
      if (params.level) {
        queryParts.push('$level: Int');
        variables.level = params.level;
      }
      
      if (params.page) {
        queryParts.push('$page: Int');
        variables.page = params.page;
      }
      
      if (params.pageSize) {
        queryParts.push('$pageSize: Int');
        variables.pageSize = params.pageSize;
      }
      
      if (queryParts.length > 0) {
        queryArgs = `(${queryParts.join(', ')})`;
      }
    }
    
    // 构建GraphQL查询
    const graphqlQuery = {
      query: `
        query GetOrganizations${queryArgs} {
          organizations${queryArgs ? `(
            ${params?.searchText ? 'searchText: $searchText' : ''}
            ${params?.unit_type ? 'unitType: $unitType' : ''}
            ${params?.status ? 'status: $status' : ''}
            ${params?.level ? 'level: $level' : ''}
            ${params?.page ? 'page: $page' : ''}
            ${params?.pageSize ? 'pageSize: $pageSize' : ''}
          )`.replace(/\s+/g, ' ').replace(/,\s*,/g, ',').replace(/^\(\s*,?\s*/, '(').replace(/,?\s*\)$/, ')') : ''} {
            items {
              code
              name
              unitType
              status
              level
              parentCode
              path
              sortOrder
              description
              createdAt
              updatedAt
            }
            totalCount
            page
            pageSize
          }
        }
      `,
      variables: variables
    };
    
      const response = await fetch('http://localhost:8090/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Tenant-ID': '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
        },
        body: JSON.stringify(graphqlQuery),
      });
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const graphqlResponse = await response.json();
      
      // 检查GraphQL错误
      if (isGraphQLError(graphqlResponse)) {
        console.warn('GraphQL errors:', graphqlResponse.errors);
        return organizationAPI.getAllFallback(params);
      }
      
      // 验证GraphQL成功响应
      if (!isGraphQLSuccessResponse(graphqlResponse)) {
        throw new ValidationError('Invalid GraphQL response structure', [{
          message: 'Response missing data field',
          code: 'invalid_structure',
          path: ['data']
        }]);
      }
      
      const organizationsData = graphqlResponse.data?.organizations;
      
      // 如果后端返回的是新格式（包含items、totalCount等）
      if (organizationsData && typeof organizationsData === 'object' && 'items' in organizationsData) {
        // 使用运行时验证替换类型断言
        const validatedItems = validateGraphQLOrganizationList(organizationsData.items);
        const adaptedOrganizations: OrganizationUnit[] = validatedItems.map(safeTransformGraphQLToOrganizationUnit);
        
        return {
          organizations: adaptedOrganizations,
          total_count: organizationsData.totalCount || adaptedOrganizations.length,
          page: organizationsData.page || 1,
          page_size: organizationsData.pageSize || adaptedOrganizations.length,
        };
      }
      
      // 如果后端返回的是旧格式，回退到客户端筛选
      return organizationAPI.getAllFallback(params);
    } catch (error) {
      if (isValidationError(error)) {
        console.error('[Organization API] Validation failed:', error.details);
        throw error;
      } else if (isNetworkError(error)) {
        console.error('[Organization API] Network error:', error.message);
        throw new Error('Network connection failed. Please check your internet connection.');
      } else {
        console.error('[Organization API] Unexpected error:', error);
        throw error;
      }
    }
  },

  // 回退方法：如果后端不支持筛选，在前端进行筛选
  getAllFallback: async (params?: OrganizationQueryParams): Promise<OrganizationListResponse> => {
    const graphqlQuery = {
      query: `
        query {
          organizations {
            code
            name
            unitType
            status
            level
            parentCode
            path
            sortOrder
            description
            createdAt
            updatedAt
          }
        }
      `
    };
    
    const response = await fetch('http://localhost:8090/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Tenant-ID': '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
      },
      body: JSON.stringify(graphqlQuery),
    });
    
    const graphqlResponse: GraphQLResponse<{ organizations: GraphQLOrganizationResponse[] }> = await response.json();
    const organizations = graphqlResponse?.data?.organizations;
    
    // 使用运行时验证适配数据格式
    const validatedOrganizations = validateGraphQLOrganizationList(organizations || []);
    let adaptedOrganizations: OrganizationUnit[] = validatedOrganizations.map(safeTransformGraphQLToOrganizationUnit);

    // 在前端应用筛选条件
    if (params) {
      if (params.searchText) {
        const searchLower = params.searchText.toLowerCase();
        adaptedOrganizations = adaptedOrganizations.filter(org => 
          org.name.toLowerCase().includes(searchLower) ||
          org.code.toLowerCase().includes(searchLower)
        );
      }
      
      if (params.unit_type) {
        adaptedOrganizations = adaptedOrganizations.filter(org => org.unit_type === params.unit_type);
      }
      
      if (params.status) {
        adaptedOrganizations = adaptedOrganizations.filter(org => org.status === params.status);
      }
      
      if (params.level) {
        adaptedOrganizations = adaptedOrganizations.filter(org => org.level === params.level);
      }
    }

    const totalCount = adaptedOrganizations.length;
    const page = params?.page || 1;
    const pageSize = params?.pageSize || 20;
    
    // 应用分页
    const startIndex = (page - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    const paginatedOrganizations = adaptedOrganizations.slice(startIndex, endIndex);
    
    return {
      organizations: paginatedOrganizations,
      total_count: totalCount,
      page: page,
      page_size: pageSize,
    };
  },

  // 获取单个组织单元
  getByCode: async (code: string): Promise<OrganizationUnit> => {
    try {
      const response = await apiClient.get<APIResponse<GraphQLOrganizationResponse>>(`/organization-units/${code}`);
      
      // 验证响应数据
      const validatedOrg = validateGraphQLOrganizationResponse(response.data);
      
      return safeTransformGraphQLToOrganizationUnit(validatedOrg);
    } catch (error) {
      if (isValidationError(error)) {
        console.error(`[Organization API] Validation failed for org ${code}:`, error.details);
        throw error;
      } else if (isAPIError(error)) {
        console.error(`[Organization API] API error for org ${code}:`, error.status, error.statusText);
        throw error;
      } else {
        console.error(`[Organization API] Unexpected error for org ${code}:`, error);
        throw error;
      }
    }
  },

  // 获取组织统计信息
  getStats: async (): Promise<OrganizationStats> => {
    // 直接调用GraphQL
    const graphqlQuery = {
      query: `
        query {
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
          }
        }
      `
    };
    
    const response = await fetch('http://localhost:8090/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Tenant-ID': '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
      },
      body: JSON.stringify(graphqlQuery),
    });
    
    const graphqlResponse: GraphQLResponse<{ organizationStats: { totalCount: number; byType: Array<{unitType: string; count: number}>; byStatus: Array<{status: string; count: number}> } }> = await response.json();
    const stats = graphqlResponse.data?.organizationStats;
    
    if (!stats) {
      throw new Error('No organization stats data received');
    }
    
    // 将GraphQL返回的数组格式转换为前端期望的对象格式
    const byTypeMap: Record<string, number> = {};
    const byStatusMap: Record<string, number> = {};
    
    // 安全检查 - 如果byType不存在或不是数组，初始化为空对象
    if (stats.byType && Array.isArray(stats.byType)) {
      stats.byType.forEach((item: {unitType: string, count: number}) => {
        byTypeMap[item.unitType] = item.count;
      });
    }
    
    // 安全检查 - 如果byStatus不存在或不是数组，初始化为空对象
    if (stats.byStatus && Array.isArray(stats.byStatus)) {
      stats.byStatus.forEach((item: {status: string, count: number}) => {
        byStatusMap[item.status] = item.count;
      });
    }
    
    return {
      total_count: stats.totalCount || 0,
      by_type: byTypeMap,
      by_status: byStatusMap,
    };
  },

  // 新增组织单元
  create: async (data: CreateOrganizationInput): Promise<OrganizationUnit> => {
    try {
      console.log('[Organization API] Creating:', data);
      
      // 使用安全的类型转换构建请求体
      const requestBody = safeTransformCreateInputToAPI(data);
      
      const response = await apiClient.post<CreateOrganizationResponse>('/organization-units', requestBody);
      
      console.log('[Organization API] Create response:', response);
      
      // 使用专门的验证函数验证创建响应
      const validatedResponse = validateCreateOrganizationResponse(response);
      
      console.log('[Organization API] Validated create response:', validatedResponse);
      
      // 重新获取完整的组织单元数据，因为创建响应只包含部分字段
      try {
        const fullOrgData = await organizationAPI.getByCode(validatedResponse.code);
        console.log('[Organization API] Created org with full data:', fullOrgData);
        return fullOrgData;
      } catch (fetchError) {
        console.warn('[Organization API] Failed to fetch full data after create, constructing minimal data:', fetchError);
        
        // 如果无法获取完整数据，构造包含所有必要字段的对象
        const minimalOrgData = {
          code: validatedResponse.code,
          parent_code: data.parent_code || '',
          name: validatedResponse.name,
          unit_type: validatedResponse.unit_type,
          status: validatedResponse.status,
          level: data.level || 1,
          path: '',
          sort_order: data.sort_order || 0,
          description: data.description || '',
          created_at: validatedResponse.created_at,
          updated_at: validatedResponse.created_at,
        };
        
        // 验证构造的完整对象
        const validatedFullOrg = validateOrganizationUnit(minimalOrgData);
        console.log('[Organization API] Created org with minimal data:', validatedFullOrg);
        return validatedFullOrg;
      }
    } catch (error) {
      console.error('[Organization API] Create failed:', error);
      if (isValidationError(error)) {
        console.error('[Organization API] Validation failed during create:', error.details);
        throw error;
      } else if (isAPIError(error)) {
        console.error('[Organization API] API error during create:', error.status, error.statusText);
        throw error;
      } else {
        throw error;
      }
    }
  },

  // 更新组织单元
  update: async (code: string, data: Omit<UpdateOrganizationInput, 'code'>): Promise<OrganizationUnit> => {
    try {
      console.log(`[Organization API] Updating ${code}:`, data);
      
      // 验证更新输入数据
      const updateData = { ...data, code }; // 添加code用于验证
      const validatedUpdateInput = validateUpdateOrganizationInput(updateData);
      
      const response = await apiClient.put<UpdateOrganizationResponse>(`/organization-units/${code}`, {
        name: validatedUpdateInput.name,
        status: validatedUpdateInput.status,
        description: validatedUpdateInput.description,
        sort_order: validatedUpdateInput.sort_order,
      });
      
      console.log(`[Organization API] Update response:`, response);
      
      // 重新获取完整数据以确保数据一致性
      try {
        const updatedOrg = await organizationAPI.getByCode(code);
        console.log(`[Organization API] Refreshed data:`, updatedOrg);
        return updatedOrg;
      } catch (fetchError) {
        console.warn(`[Organization API] Failed to fetch updated data, constructing from response:`, fetchError);
        
        // 如果获取失败，构造返回数据并验证
        const fallbackData = {
          code: code,
          parent_code: '',
          name: validatedUpdateInput.name || '',
          unit_type: 'DEPARTMENT',
          status: validatedUpdateInput.status || 'ACTIVE',
          level: 1,
          path: '',
          sort_order: validatedUpdateInput.sort_order || 0,
          description: validatedUpdateInput.description || '',
          created_at: '',
          updated_at: response.updated_at || new Date().toISOString(),
        };
        
        return validateOrganizationUnit(fallbackData);
      }
    } catch (error) {
      console.error(`[Organization API] Update failed:`, error);
      if (isValidationError(error)) {
        console.error('[Organization API] Validation failed during update:', error.details);
        throw error;
      } else if (isAPIError(error)) {
        console.error('[Organization API] API error during update:', error.status, error.statusText);
        throw error;
      } else {
        throw error;
      }
    }
  },

  // 删除组织单元
  delete: async (code: string): Promise<void> => {
    try {
      console.log(`[Organization API] Deleting ${code}`);
      
      await apiClient.delete(`/organization-units/${code}`);
      
      console.log(`[Organization API] Successfully deleted ${code}`);
    } catch (error) {
      console.error(`[Organization API] Delete failed:`, error);
      throw error;
    }
  },
};