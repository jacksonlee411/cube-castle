import { apiClient } from './client';
import type { OrganizationUnit, OrganizationListResponse, OrganizationStats } from '../types';

export const organizationAPI = {
  // 获取组织单元列表
  getAll: async (params?: {
    unit_type?: string;
    status?: string;
    limit?: number;
    offset?: number;
  }): Promise<OrganizationListResponse> => {
    const searchParams = new URLSearchParams();
    
    if (params?.unit_type) searchParams.set('unit_type', params.unit_type);
    if (params?.status) searchParams.set('status', params.status);
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    
    const queryString = searchParams.toString();
    const endpoint = `/organization-units${queryString ? `?${queryString}` : ''}`;
    
    const response = await apiClient.get<{data: {organizations: any[]}}>(endpoint);
    
    // 适配后端返回的数据格式
    const adaptedOrganizations = response.data.organizations.map(org => ({
      code: org.code,
      parent_code: org.parentCode,
      name: org.name,
      unit_type: org.unitType,
      status: org.status,
      level: org.level,
      path: org.path,
      sort_order: org.sortOrder,
      description: org.description,
      created_at: org.createdAt,
      updated_at: org.updatedAt,
    }));
    
    return {
      organizations: adaptedOrganizations,
      total_count: response.data.organizations.length,
      page: 1,
      page_size: response.data.organizations.length,
    };
  },

  // 获取单个组织单元
  getByCode: async (code: string): Promise<OrganizationUnit> => {
    const response = await apiClient.get<{data: any}>(`/organization-units/${code}`);
    const org = response.data;
    
    return {
      code: org.code,
      parent_code: org.parentCode,
      name: org.name,
      unit_type: org.unitType,
      status: org.status,
      level: org.level,
      path: org.path,
      sort_order: org.sortOrder,
      description: org.description,
      created_at: org.createdAt,
      updated_at: org.updatedAt,
    };
  },

  // 获取组织统计信息
  getStats: async (): Promise<OrganizationStats> => {
    const response = await apiClient.get<{data: {organizationStats: any}}>('/organization-units/stats');
    const stats = response.data.organizationStats;
    
    return {
      total_count: stats.totalCount || 0,
      by_type: stats.byType || {},
      by_status: stats.byStatus || {},
    };
  },
};