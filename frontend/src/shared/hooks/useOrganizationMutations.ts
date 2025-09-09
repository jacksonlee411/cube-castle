import { useMutation, useQueryClient } from '@tanstack/react-query';
import { organizationAPI } from '../api';
import type { OrganizationUnit, OrganizationRequest } from '../types';


// 新增组织单元
export const useCreateOrganization = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: OrganizationRequest): Promise<OrganizationUnit> => {
      console.log('[Mutation] Creating organization:', data);
      const response = await organizationAPI.create(data);
      console.log('[Mutation] Create successful:', response);
      return response;
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
      const response = await organizationAPI.update(data.code!, data);
      console.log('[Mutation] Update successful:', response);
      return response;
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
      const response = await organizationAPI.suspend(code, reason);
      console.log('[Mutation] Suspend successful:', response);
      return response;
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
      const response = await organizationAPI.activate(code, reason);
      console.log('[Mutation] Activate successful:', response);
      return response;
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

