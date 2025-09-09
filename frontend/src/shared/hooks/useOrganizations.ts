// TODO-TEMPORARY: 此Hook将逐步迁移到useEnterpriseOrganizations，当前作为简化版本包装器
// 迁移期限: 2025-09-23 (2周后)
import { useQuery } from '@tanstack/react-query';
import { organizationAPI } from '../api';
import { useEnterpriseOrganizations } from './useEnterpriseOrganizations';
import type { OrganizationQueryParams } from '../types/organization';

// 简化版组织查询Hook - 包装useEnterpriseOrganizations
export const useOrganizations = (params?: OrganizationQueryParams) => {
  // 使用企业级Hook但只返回基础字段，保持接口兼容性
  const {
    organizations,
    loading,
    error,
    refetch
  } = useEnterpriseOrganizations(params);

  return {
    data: organizations,
    isLoading: loading,
    error,
    refetch
  };
};

// 单个组织单元查询
export const useOrganization = (code: string) => {
  return useQuery({
    queryKey: ['organization', code],
    queryFn: () => organizationAPI.getByCode(code),
    enabled: !!code,
  });
};

