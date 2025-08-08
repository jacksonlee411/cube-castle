import { apiClient } from './client';
import type { 
  OrganizationUnit, 
  OrganizationListResponse, 
  OrganizationStats, 
  APIResponse,
  OrganizationListAPIResponse,
  OrganizationStatsAPIResponse,
  GraphQLOrganizationResponse,
  GraphQLStatsTypeItem,
  GraphQLStatsStatusItem
} from '../types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../hooks/useOrganizationMutations';

export const organizationAPI = {
  // 获取组织单元列表
  getAll: async (params?: {
    unit_type?: string;
    status?: string;
    limit?: number;
    offset?: number;
  }): Promise<OrganizationListResponse> => {
    // 直接调用GraphQL
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
    const organizations = graphqlResponse.data.organizations;
    
    // 适配后端返回的数据格式 - 修复字段名映射问题
    const adaptedOrganizations: OrganizationUnit[] = organizations.map((org: GraphQLOrganizationResponse) => ({
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
    }));
    
    return {
      organizations: adaptedOrganizations,
      total_count: adaptedOrganizations.length,
      page: 1,
      page_size: adaptedOrganizations.length,
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
    const response = await apiClient.post<APIResponse<GraphQLOrganizationResponse>>('/organization-units', {
      code: data.code,
      parent_code: data.parent_code,
      name: data.name,
      unit_type: data.unit_type,
      status: data.status,
      level: data.level,
      sort_order: data.sort_order,
      description: data.description,
    });
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

  // 更新组织单元
  update: async (code: string, data: Omit<UpdateOrganizationInput, 'code'>): Promise<OrganizationUnit> => {
    const response = await apiClient.put<APIResponse<GraphQLOrganizationResponse>>(`/organization-units/${code}`, {
      name: data.name,
      status: data.status,
      description: data.description,
      sort_order: data.sort_order,
    });
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

  // 删除组织单元
  delete: async (code: string): Promise<void> => {
    await apiClient.delete(`/organization-units/${code}`);
  },
};