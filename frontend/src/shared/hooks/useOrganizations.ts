/**
 * DEPRECATED: 该文件已被 useEnterpriseOrganizations 替代
 * 
 * 请使用：
 * import { useEnterpriseOrganizations, useOrganizationDetails } from '@/shared/hooks';
 * 
 * 迁移映射：
 * - useOrganizations -> useEnterpriseOrganizations
 * - useOrganization -> useOrganizationDetails (来自 useEnterpriseOrganizations 文件)
 */

// 临时兼容封装，在所有引用替换完成后删除
import { useEnterpriseOrganizations } from './useEnterpriseOrganizations';
import type { OrganizationQueryParams } from '../types/organization';

// DEPRECATED: 使用 useEnterpriseOrganizations 替代
export const useOrganizations = (params?: OrganizationQueryParams) => {
  const result = useEnterpriseOrganizations(params);
  return {
    data: result.organizations,
    isLoading: result.loading,
    error: result.error,
    refetch: result.fetchOrganizations // 使用fetchOrganizations作为refetch
  };
};

// DEPRECATED: 使用 useEnterpriseOrganizations 的 fetchOrganizationByCode 替代
export const useOrganization = (code: string) => {
  const { fetchOrganizationByCode, loading, error } = useEnterpriseOrganizations();
  
  return {
    data: null, // DEPRECATED: 使用 fetchOrganizationByCode 方法
    isLoading: loading,
    error,
    refetch: () => fetchOrganizationByCode(code)
  };
};

