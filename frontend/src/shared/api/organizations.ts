import { apiClient } from './client';
import type { 
  OrganizationUnit, 
  OrganizationListResponse, 
  OrganizationStats, 
  APIResponse,
  GraphQLOrganizationResponse,
  CreateOrganizationResponse,
  UpdateOrganizationResponse
} from '../types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../hooks/useOrganizationMutations';

export interface OrganizationQueryParams {
  searchText?: string;
  unit_type?: string;
  status?: string;
  level?: number;
  page?: number;
  pageSize?: number;
}

export const organizationAPI = {
  // 获取组织单元列表
  getAll: async (params?: OrganizationQueryParams): Promise<OrganizationListResponse> => {
    // 构建GraphQL查询变量
    const variables: any = {};
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
    
    const graphqlResponse = await response.json();
    
    // 如果后端不支持新的查询参数，回退到旧的查询方式
    if (graphqlResponse?.errors) {
      console.warn('GraphQL errors:', graphqlResponse.errors);
      return organizationAPI.getAllFallback(params);
    }
    
    const organizationsData = graphqlResponse?.data?.organizations;
    
    // 如果后端返回的是新格式（包含items、totalCount等）
    if (organizationsData && typeof organizationsData === 'object' && 'items' in organizationsData) {
      const adaptedOrganizations: OrganizationUnit[] = organizationsData.items.map((org: GraphQLOrganizationResponse) => ({
        code: org.code,
        parent_code: org.parentCode || '',
        name: org.name,
        unit_type: org.unitType as 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM',
        status: org.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
        level: org.level,
        path: org.path,
        sort_order: org.sortOrder || 0,
        description: org.description || '',
        created_at: org.createdAt || '',
        updated_at: org.updatedAt || '',
      }));
      
      return {
        organizations: adaptedOrganizations,
        total_count: organizationsData.totalCount || adaptedOrganizations.length,
        page: organizationsData.page || 1,
        page_size: organizationsData.pageSize || adaptedOrganizations.length,
      };
    }
    
    // 如果后端返回的是旧格式，回退到客户端筛选
    return organizationAPI.getAllFallback(params);
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
    
    const graphqlResponse = await response.json();
    const organizations = graphqlResponse?.data?.organizations;
    
    // 适配数据格式
    let adaptedOrganizations: OrganizationUnit[] = (organizations || []).map((org: GraphQLOrganizationResponse) => ({
      code: org.code,
      parent_code: org.parentCode || '',
      name: org.name,
      unit_type: org.unitType as 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM',
      status: org.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
      level: org.level,
      path: org.path,
      sort_order: org.sortOrder || 0,
      description: org.description || '',
      created_at: org.createdAt || '',
      updated_at: org.updatedAt || '',
    }));

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
    const response = await apiClient.get<APIResponse<GraphQLOrganizationResponse>>(`/organization-units/${code}`);
    const org = response.data;
    
    return {
      code: org.code,
      parent_code: org.parentCode || '', // 处理null值
      name: org.name,
      unit_type: org.unitType as 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM',
      status: org.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
      level: org.level,
      path: org.path,
      sort_order: org.sortOrder || 0, // 处理null值
      description: org.description || '', // 处理null值
      created_at: org.createdAt || '',
      updated_at: org.updatedAt || '',
    };
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
    
    const graphqlResponse = await response.json();
    const stats = graphqlResponse.data.organizationStats;
    
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
      
      // 构建请求体，过滤undefined字段
      const requestBody: Record<string, any> = {
        name: data.name,
        unit_type: data.unit_type,
        status: data.status,
        level: data.level,
        sort_order: data.sort_order,
        description: data.description,
      };
      
      // 只有当code不为undefined时才添加
      if (data.code !== undefined) {
        requestBody.code = data.code;
      }
      
      // 只有当parent_code不为undefined时才添加
      if (data.parent_code !== undefined) {
        requestBody.parent_code = data.parent_code;
      }
      
      const response = await apiClient.post<CreateOrganizationResponse>('/organization-units', requestBody);
      
      console.log('[Organization API] Create response:', response);
      
      // 构造返回的组织单元对象
      const newOrg: OrganizationUnit = {
        code: response.code,
        parent_code: data.parent_code || '', 
        name: response.name || data.name,
        unit_type: (response.unit_type || data.unit_type) as 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM',
        status: (response.status || data.status) as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
        level: data.level || 1,
        path: response.path || '',
        sort_order: data.sort_order || 0,
        description: data.description || '',
        created_at: response.created_at || new Date().toISOString(),
        updated_at: response.updated_at || new Date().toISOString(),
      };
      
      console.log('[Organization API] Created org:', newOrg);
      return newOrg;
    } catch (error) {
      console.error('[Organization API] Create failed:', error);
      throw error;
    }
  },

  // 更新组织单元
  update: async (code: string, data: Omit<UpdateOrganizationInput, 'code'>): Promise<OrganizationUnit> => {
    try {
      console.log(`[Organization API] Updating ${code}:`, data);
      
      const response = await apiClient.put<UpdateOrganizationResponse>(`/organization-units/${code}`, {
        name: data.name,
        status: data.status,
        description: data.description,
        sort_order: data.sort_order,
      });
      
      console.log(`[Organization API] Update response:`, response);
      
      // 重新获取完整数据以确保数据一致性
      try {
        const updatedOrg = await organizationAPI.getByCode(code);
        console.log(`[Organization API] Refreshed data:`, updatedOrg);
        return updatedOrg;
      } catch (fetchError) {
        console.warn(`[Organization API] Failed to fetch updated data, using response:`, fetchError);
        
        // 如果获取失败，构造返回数据
        return {
          code: code,
          parent_code: '',
          name: data.name || '',
          unit_type: 'DEPARTMENT',
          status: data.status || 'ACTIVE',
          level: 1,
          path: '',
          sort_order: data.sort_order || 0,
          description: data.description || '',
          created_at: '',
          updated_at: response.updated_at || new Date().toISOString(),
        };
      }
    } catch (error) {
      console.error(`[Organization API] Update failed:`, error);
      throw error;
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