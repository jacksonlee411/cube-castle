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
    
    return apiClient.get<OrganizationListResponse>(endpoint);
  },

  // 获取单个组织单元
  getByCode: async (code: string): Promise<OrganizationUnit> => {
    return apiClient.get<OrganizationUnit>(`/organization-units/${code}`);
  },

  // 获取组织统计信息
  getStats: async (): Promise<OrganizationStats> => {
    return apiClient.get<OrganizationStats>('/organization-units/stats');
  },
};