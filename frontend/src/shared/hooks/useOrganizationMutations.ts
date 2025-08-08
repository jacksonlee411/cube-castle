import { useMutation, useQueryClient } from '@tanstack/react-query';
import { organizationAPI } from '../api/organizations';
import type { OrganizationUnit } from '../types';

// 新增组织单元的输入类型
export interface CreateOrganizationInput {
  code: string;
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
      const response = await organizationAPI.create(data);
      return response;
    },
    onSuccess: () => {
      // 刷新组织列表和统计
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      queryClient.invalidateQueries({ queryKey: ['organization-stats'] });
    },
  });
};

// 更新组织单元
export const useUpdateOrganization = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: UpdateOrganizationInput): Promise<OrganizationUnit> => {
      const response = await organizationAPI.update(data.code, data);
      return response;
    },
    onSuccess: (_, variables) => {
      // 刷新相关查询
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      queryClient.invalidateQueries({ queryKey: ['organization', variables.code] });
      queryClient.invalidateQueries({ queryKey: ['organization-stats'] });
    },
  });
};

// 删除组织单元
export const useDeleteOrganization = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (code: string): Promise<void> => {
      await organizationAPI.delete(code);
    },
    onSuccess: () => {
      // 刷新组织列表和统计
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      queryClient.invalidateQueries({ queryKey: ['organization-stats'] });
    },
  });
};