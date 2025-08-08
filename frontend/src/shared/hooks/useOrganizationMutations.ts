import { useMutation, useQueryClient } from '@tanstack/react-query';
import { organizationAPI } from '../api/organizations';
import type { OrganizationUnit } from '../types';

// 新增组织单元的输入类型
export interface CreateOrganizationInput {
  code?: string; // 修改为可选，支持自动生成
  parent_code?: string;
  name: string;
  unit_type: 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM';
  status: 'ACTIVE' | 'INACTIVE' | 'PLANNED';
  level: number;
  sort_order: number;
  description?: string;
}

// 更新组织单元的输入类型
export interface UpdateOrganizationInput {
  code: string;
  name?: string;
  status?: 'ACTIVE' | 'INACTIVE' | 'PLANNED';
  description?: string;
  sort_order?: number;
}

// 新增组织单元
export const useCreateOrganization = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: CreateOrganizationInput): Promise<OrganizationUnit> => {
      console.log('[Mutation] Creating organization:', data);
      const response = await organizationAPI.create(data);
      console.log('[Mutation] Create successful:', response);
      return response;
    },
    onSuccess: () => {
      console.log('[Mutation] Create success, invalidating queries');
      // 刷新组织列表和统计
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      queryClient.invalidateQueries({ queryKey: ['organization-stats'] });
    },
    onError: (error) => {
      console.error('[Mutation] Create failed:', error);
    },
  });
};

// 更新组织单元
export const useUpdateOrganization = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: UpdateOrganizationInput): Promise<OrganizationUnit> => {
      console.log('[Mutation] Updating organization:', data);
      const response = await organizationAPI.update(data.code, data);
      console.log('[Mutation] Update successful:', response);
      return response;
    },
    onSuccess: (_, variables) => {
      console.log('[Mutation] Update success, invalidating queries');
      // 刷新相关查询
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      queryClient.invalidateQueries({ queryKey: ['organization', variables.code] });
      queryClient.invalidateQueries({ queryKey: ['organization-stats'] });
    },
    onError: (error) => {
      console.error('[Mutation] Update failed:', error);
    },
  });
};

// 删除组织单元
export const useDeleteOrganization = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (code: string): Promise<void> => {
      console.log('[Mutation] Deleting organization:', code);
      await organizationAPI.delete(code);
      console.log('[Mutation] Delete successful for:', code);
    },
    onSuccess: () => {
      console.log('[Mutation] Delete success, invalidating queries');
      // 刷新组织列表和统计
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      queryClient.invalidateQueries({ queryKey: ['organization-stats'] });
    },
    onError: (error) => {
      console.error('[Mutation] Delete failed:', error);
    },
  });
};