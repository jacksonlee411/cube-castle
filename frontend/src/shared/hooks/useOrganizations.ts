import { useQuery } from '@tanstack/react-query';
import { organizationAPI } from '../api';

// 组织单元列表查询
export const useOrganizations = (params?: {
  unit_type?: string;
  status?: string;
  limit?: number;
  offset?: number;
}) => {
  return useQuery({
    queryKey: ['organizations', params],
    queryFn: () => organizationAPI.getAll(),
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