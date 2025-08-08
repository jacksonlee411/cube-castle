import { useQuery } from '@tanstack/react-query';
import { organizationAPI, type OrganizationQueryParams } from '../api';

// 组织单元列表查询
export const useOrganizations = (params?: OrganizationQueryParams) => {
  return useQuery({
    queryKey: ['organizations', params],
    queryFn: () => organizationAPI.getAll(params),
    // 启用缓存，但当参数变化时重新获取
    staleTime: 5 * 60 * 1000, // 5分钟内认为数据是新鲜的
  });
};

// 单个组织单元查询
export const useOrganization = (code: string) => {
  return useQuery({
    queryKey: ['organization', code],
    queryFn: () => organizationAPI.getByCode(code),
    enabled: !!code,
  });
};

// 组织统计信息查询
export const useOrganizationStats = () => {
  return useQuery({
    queryKey: ['organization-stats'],
    queryFn: () => organizationAPI.getStats(),
  });
};