import { useMutation, useQueryClient } from '@tanstack/react-query';
import { organizationAPI } from '../api/organizations-simplified';
import type { OrganizationUnit, OrganizationStatus } from '../types';

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
  unit_type?: 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM';
  status?: 'ACTIVE' | 'INACTIVE' | 'PLANNED';
  description?: string;
  sort_order?: number;
  level?: number;
  parent_code?: string;
}

// 状态切换输入类型
export interface ToggleStatusInput {
  code: string;
  status: OrganizationStatus;
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
    onSuccess: (createdOrganization) => {
      console.log('[Mutation] Create success, invalidating queries');
      
      // 立即失效所有相关查询缓存
      queryClient.invalidateQueries({ 
        queryKey: ['organizations'],
        exact: false
      });
      
      queryClient.invalidateQueries({ 
        queryKey: ['organization-stats'],
        exact: false
      });
      
      // 强制重新获取数据以确保立即显示新创建的组织
      queryClient.refetchQueries({ 
        queryKey: ['organizations'],
        type: 'active'
      });
      
      queryClient.refetchQueries({ 
        queryKey: ['organization-stats'],
        type: 'active'
      });
      
      console.log('[Mutation] Create cache invalidation and refetch completed');
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
    onSuccess: (updatedOrganization, variables) => {
      console.log('[Mutation] Update success, invalidating queries for:', variables.code);
      
      // 立即失效所有相关查询缓存
      queryClient.invalidateQueries({ 
        queryKey: ['organizations'],
        exact: false
      });
      
      queryClient.invalidateQueries({ 
        queryKey: ['organization', variables.code],
        exact: false
      });
      
      queryClient.invalidateQueries({ 
        queryKey: ['organization-stats'],
        exact: false
      });
      
      // 强制重新获取数据以确保立即显示更新的组织
      queryClient.refetchQueries({ 
        queryKey: ['organizations'],
        type: 'active'
      });
      
      queryClient.refetchQueries({ 
        queryKey: ['organization-stats'],
        type: 'active'
      });
      
      // 新增：直接设置缓存数据以提供即时反馈
      queryClient.setQueryData(['organization', variables.code], updatedOrganization);
      
      // 新增：移除过时的缓存数据
      queryClient.removeQueries({ 
        queryKey: ['organizations'],
        exact: false,
        type: 'inactive'
      });
      
      console.log('[Mutation] Update cache invalidation and refetch completed');
    },
    onError: (error) => {
      console.error('[Mutation] Update failed:', error);
    },
  });
};

// 状态切换操作 (替代删除操作)
export const useToggleOrganizationStatus = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: ToggleStatusInput): Promise<OrganizationUnit> => {
      console.log('[Mutation] Toggling organization status:', data);
      const response = await organizationAPI.update(data.code, { status: data.status });
      console.log('[Mutation] Toggle status successful:', response);
      return response;
    },
    onSuccess: (updatedOrganization, variables) => {
      console.log('[Mutation] Status toggle success, invalidating queries for:', variables.code);
      
      // 立即失效所有相关查询缓存 - 遵循CQRS原则
      queryClient.invalidateQueries({ 
        queryKey: ['organizations'],
        exact: false
      });
      
      queryClient.invalidateQueries({ 
        queryKey: ['organization', variables.code],
        exact: false
      });
      
      queryClient.invalidateQueries({ 
        queryKey: ['organization-stats'],
        exact: false
      });
      
      // 强制重新获取数据以确保立即显示状态变更
      queryClient.refetchQueries({ 
        queryKey: ['organizations'],
        type: 'active'
      });
      
      queryClient.refetchQueries({ 
        queryKey: ['organization-stats'],
        type: 'active'
      });
      
      // 新增：直接设置缓存数据以提供即时反馈
      queryClient.setQueryData(['organization', variables.code], updatedOrganization);
      
      // 新增：移除过时的缓存数据
      queryClient.removeQueries({ 
        queryKey: ['organizations'],
        exact: false,
        type: 'inactive'
      });
      
      // 新增：强制清除所有可能的缓存变体
      queryClient.clear();
      
      console.log('[Mutation] Status toggle cache invalidation and refetch completed');
    },
    onError: (error) => {
      console.error('[Mutation] Toggle status failed:', error);
    },
  });
};