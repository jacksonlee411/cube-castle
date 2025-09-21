import { useMutation, useQueryClient } from '@tanstack/react-query';
import { unifiedRESTClient } from '../api';
import type { OrganizationUnit, OrganizationRequest, APIResponse } from '../types';

const ensureSuccess = (response: APIResponse<OrganizationUnit>, fallbackMessage: string): OrganizationUnit => {
  if (!response.success || !response.data) {
    const error = new Error(response.error?.message ?? fallbackMessage) as Error & {
      code?: string
      details?: unknown
    }
    if (response.error?.code) {
      error.code = response.error.code
    }
    if (response.error?.details) {
      error.details = response.error.details
    }
    throw error;
  }
  return response.data;
};


// 新增组织单元
export const useCreateOrganization = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: OrganizationRequest): Promise<OrganizationUnit> => {
      console.log('[Mutation] Creating organization:', data);
      const response = await unifiedRESTClient.request<APIResponse<OrganizationUnit>>('/organization-units', {
        method: 'POST',
        body: JSON.stringify(data),
        headers: { 'Content-Type': 'application/json' }
      });
      console.log('[Mutation] Create successful:', response);
      return ensureSuccess(response, '创建组织失败');
    },
    onSettled: () => {
      console.log('[Mutation] Create settled, invalidating queries');
      
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
  });
};

// 更新组织单元
export const useUpdateOrganization = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: OrganizationRequest): Promise<OrganizationUnit> => {
      console.log('[Mutation] Updating organization:', data);
      const response = await unifiedRESTClient.request<APIResponse<OrganizationUnit>>(`/organization-units/${data.code}`, {
        method: 'PUT',
        body: JSON.stringify(data),
        headers: { 'Content-Type': 'application/json' }
      });
      console.log('[Mutation] Update successful:', response);
      return ensureSuccess(response, '更新组织失败');
    },
    onSettled: (data, error, variables) => {
      console.log('[Mutation] Update settled:', variables.code);
      
      // 立即失效所有相关查询缓存
      queryClient.invalidateQueries({ 
        queryKey: ['organizations'],
        exact: false
      });
      
      queryClient.invalidateQueries({ 
        queryKey: ['organization', variables.code!],
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
      if (data) {
        queryClient.setQueryData(['organization', variables.code!], data);
      }
      
      // 新增：移除过时的缓存数据
      queryClient.removeQueries({ 
        queryKey: ['organizations'],
        exact: false,
        type: 'inactive'
      });
      
      console.log('[Mutation] Update cache invalidation and refetch completed');
    },
  });
};

// === 新增：操作驱动状态管理Hooks ===

// 停用组织
export const useSuspendOrganization = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ code, reason }: { code: string; reason: string }): Promise<OrganizationUnit> => {
      console.log('[Mutation] Suspending organization:', code, reason);
      const response = await unifiedRESTClient.request<APIResponse<OrganizationUnit>>(`/organization-units/${code}/suspend`, {
        method: 'POST',
        body: JSON.stringify({ reason }),
        headers: { 'Content-Type': 'application/json' }
      });
      console.log('[Mutation] Suspend successful:', response);
      return ensureSuccess(response, '暂停组织失败');
    },
    onSettled: (data, error, variables) => {
      console.log('[Mutation] Suspend settled:', variables.code);
      
      // 立即失效所有相关查询缓存
      queryClient.invalidateQueries({ 
        queryKey: ['organizations'],
        exact: false
      });
      
      queryClient.invalidateQueries({ 
        queryKey: ['organization', variables.code!],
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
      
      // 直接设置缓存数据以提供即时反馈
      if (data) {
        queryClient.setQueryData(['organization', variables.code!], data);
      }
      
      console.log('[Mutation] Suspend cache invalidation and refetch completed');
    },
  });
};

// 重新启用组织
export const useActivateOrganization = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ code, reason }: { code: string; reason: string }): Promise<OrganizationUnit> => {
      console.log('[Mutation] Activating organization:', code, reason);
      const response = await unifiedRESTClient.request<APIResponse<OrganizationUnit>>(`/organization-units/${code}/activate`, {
        method: 'POST',
        body: JSON.stringify({ reason }),
        headers: { 'Content-Type': 'application/json' }
      });
      console.log('[Mutation] Activate successful:', response);
      return ensureSuccess(response, '重新启用组织失败');
    },
    onSettled: (data, error, variables) => {
      console.log('[Mutation] Activate settled:', variables.code);
      
      // 立即失效所有相关查询缓存
      queryClient.invalidateQueries({ 
        queryKey: ['organizations'],
        exact: false
      });
      
      queryClient.invalidateQueries({ 
        queryKey: ['organization', variables.code!],
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
      
      // 直接设置缓存数据以提供即时反馈
      if (data) {
        queryClient.setQueryData(['organization', variables.code!], data);
      }
      
      console.log('[Mutation] Activate cache invalidation and refetch completed');
    },
  });
};

