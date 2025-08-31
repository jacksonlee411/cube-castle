/**
 * 企业级组织单元管理Hook
 * 集成企业级响应信封格式和错误处理
 * 支持后端P0级响应格式统一修复
 */

import { useState, useEffect, useCallback } from 'react';
import type { 
  OrganizationUnit, 
  OrganizationListResponse, 
  OrganizationStats,
  OrganizationQueryParams
} from '../types';
import type { APIResponse } from '../types/api';
import type { TemporalQueryParams } from '../types/temporal';
import { enterpriseOrganizationAPI } from '../api/organizations-enterprise';

// 扩展查询参数
interface ExtendedOrganizationQueryParams extends OrganizationQueryParams {
  searchText?: string;
  pageSize?: number;
  temporalParams?: TemporalQueryParams;
}

// Hook状态接口
interface OrganizationState {
  organizations: OrganizationUnit[];
  totalCount: number;
  page: number;
  pageSize: number;
  totalPages: number;
  loading: boolean;
  error: string | null;
  stats: OrganizationStats | null;
  lastRequestId?: string;
  lastUpdate?: string;
}

// Hook操作接口
interface OrganizationOperations {
  fetchOrganizations: (params?: ExtendedOrganizationQueryParams) => Promise<APIResponse<OrganizationListResponse>>;
  fetchOrganizationByCode: (code: string, temporalParams?: TemporalQueryParams) => Promise<APIResponse<OrganizationUnit>>;
  fetchStats: () => Promise<APIResponse<OrganizationStats>>;
  refreshData: () => void;
  clearError: () => void;
}

/**
 * 企业级组织单元管理Hook
 * 自动处理企业级响应信封格式
 */
export const useEnterpriseOrganizations = (
  initialParams?: ExtendedOrganizationQueryParams
): OrganizationState & OrganizationOperations => {
  // 状态管理
  const [state, setState] = useState<OrganizationState>({
    organizations: [],
    totalCount: 0,
    page: 1,
    pageSize: 50,
    totalPages: 0,
    loading: false,
    error: null,
    stats: null
  });

  // 获取组织列表
  const fetchOrganizations = useCallback(async (
    params?: ExtendedOrganizationQueryParams
  ): Promise<APIResponse<OrganizationListResponse>> => {
    setState(prev => ({ ...prev, loading: true, error: null }));
    
    try {
      const response = await enterpriseOrganizationAPI.getAll(params);
      
      if (response.success && response.data) {
        setState(prev => ({
          ...prev,
          organizations: response.data!.organizations,
          totalCount: response.data!.totalCount,
          page: response.data!.page,
          pageSize: response.data!.pageSize,
          totalPages: response.data!.totalPages,
          loading: false,
          error: null,
          lastRequestId: response.requestId,
          lastUpdate: response.timestamp
        }));
      } else {
        const errorMessage = response.error?.message || '获取组织列表失败';
        setState(prev => ({ 
          ...prev, 
          loading: false, 
          error: errorMessage,
          lastRequestId: response.requestId
        }));
      }
      
      return response;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '网络请求失败';
      setState(prev => ({ ...prev, loading: false, error: errorMessage }));
      
      return {
        success: false,
        error: {
          code: 'HOOK_ERROR',
          message: errorMessage,
          details: error
        },
        timestamp: new Date().toISOString()
      };
    }
  }, []);

  // 根据代码获取单个组织
  const fetchOrganizationByCode = useCallback(async (
    code: string, 
    temporalParams?: TemporalQueryParams
  ): Promise<APIResponse<OrganizationUnit>> => {
    setState(prev => ({ ...prev, loading: true, error: null }));
    
    try {
      const response = await enterpriseOrganizationAPI.getByCode(code, temporalParams);
      
      setState(prev => ({ 
        ...prev, 
        loading: false, 
        error: response.success ? null : (response.error?.message || '获取组织失败'),
        lastRequestId: response.requestId,
        lastUpdate: response.timestamp
      }));
      
      return response;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '网络请求失败';
      setState(prev => ({ ...prev, loading: false, error: errorMessage }));
      
      return {
        success: false,
        error: {
          code: 'HOOK_ERROR',
          message: errorMessage,
          details: error
        },
        timestamp: new Date().toISOString()
      };
    }
  }, []);

  // 获取统计信息
  const fetchStats = useCallback(async (): Promise<APIResponse<OrganizationStats>> => {
    try {
      const response = await enterpriseOrganizationAPI.getStats();
      
      if (response.success && response.data) {
        setState(prev => ({
          ...prev,
          stats: response.data!,
          lastRequestId: response.requestId,
          lastUpdate: response.timestamp
        }));
      }
      
      return response;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '获取统计信息失败';
      
      return {
        success: false,
        error: {
          code: 'HOOK_ERROR',
          message: errorMessage,
          details: error
        },
        timestamp: new Date().toISOString()
      };
    }
  }, []);

  // 刷新数据
  const refreshData = useCallback(() => {
    if (initialParams) {
      fetchOrganizations(initialParams);
    }
    fetchStats();
  }, [initialParams, fetchOrganizations, fetchStats]);

  // 清除错误
  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  // 初始化加载
  useEffect(() => {
    if (initialParams) {
      fetchOrganizations(initialParams);
    }
    fetchStats();
  }, [initialParams, fetchOrganizations, fetchStats]);

  return {
    // 状态
    ...state,
    
    // 操作
    fetchOrganizations,
    fetchOrganizationByCode,
    fetchStats,
    refreshData,
    clearError
  };
};

// 简化版Hook，只获取组织列表
export const useOrganizationList = (
  params?: ExtendedOrganizationQueryParams
) => {
  const {
    organizations,
    totalCount,
    page,
    pageSize,
    totalPages,
    loading,
    error,
    lastRequestId,
    lastUpdate,
    fetchOrganizations,
    clearError
  } = useEnterpriseOrganizations(params);

  return {
    organizations,
    totalCount,
    page,
    pageSize,
    totalPages,
    loading,
    error,
    lastRequestId,
    lastUpdate,
    fetchOrganizations,
    clearError,
    // 便捷方法
    refresh: () => fetchOrganizations(params)
  };
};


export default useEnterpriseOrganizations;