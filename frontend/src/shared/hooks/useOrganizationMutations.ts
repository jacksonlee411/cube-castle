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
      
      console.log('[Mutation] Update cache invalidation and refetch completed');
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
    onSuccess: (_, organizationCode) => {
      console.log('[Mutation] Delete success, updating cache for:', organizationCode);
      
      // 方案3: 直接更新缓存数据，立即反映删除结果
      
      // 1. 直接从组织列表缓存中移除已删除的组织
      queryClient.setQueryData(['organizations'], (oldData: any) => {
        if (!oldData) return oldData;
        
        console.log('[Cache Update] Removing organization from cache:', organizationCode);
        
        // 处理分页数据结构
        if (oldData.pages) {
          return {
            ...oldData,
            pages: oldData.pages.map((page: any) => ({
              ...page,
              data: page.data ? page.data.filter((org: any) => org.code !== organizationCode) : []
            }))
          };
        } 
        // 处理简单数组数据结构
        else if (Array.isArray(oldData)) {
          return oldData.filter((org: any) => org.code !== organizationCode);
        }
        // 处理带data属性的对象结构
        else if (oldData.data && Array.isArray(oldData.data)) {
          return {
            ...oldData,
            data: oldData.data.filter((org: any) => org.code !== organizationCode)
          };
        }
        
        return oldData;
      });
      
      // 2. 更新统计数据缓存
      queryClient.setQueryData(['organization-stats'], (oldStats: any) => {
        if (!oldStats) return oldStats;
        
        console.log('[Cache Update] Updating stats after deletion');
        
        // 假设统计数据结构包含总数和分类统计
        const newStats = { ...oldStats };
        
        // 减少总数
        if (typeof newStats.total === 'number') {
          newStats.total = Math.max(0, newStats.total - 1);
        }
        
        // 更新按状态统计（需要知道被删除组织的状态）
        // 这里可以从删除前的数据中获取，或者让API返回更多信息
        
        return newStats;
      });
      
      // 3. 移除被删除组织的单个查询缓存
      queryClient.removeQueries({ 
        queryKey: ['organization', organizationCode] 
      });
      
      // 4. 后台异步刷新数据以确保数据一致性（可选）
      setTimeout(() => {
        queryClient.invalidateQueries({ 
          queryKey: ['organizations'], 
          exact: false 
        });
        queryClient.invalidateQueries({ 
          queryKey: ['organization-stats'], 
          exact: false 
        });
        console.log('[Cache Update] Background refresh completed');
      }, 2000); // 2秒后后台刷新，确保服务端数据同步
      
      console.log('[Mutation] Direct cache update completed');
    },
    onError: (error) => {
      console.error('[Mutation] Delete failed:', error);
    },
  });
};