import { useQuery } from '@tanstack/react-query';
import { organizationAPI, type OrganizationQueryParams } from '../api';

// 组织单元列表查询
export const useOrganizations = (params?: OrganizationQueryParams) => {
  return useQuery({
    queryKey: ['organizations', params],
    queryFn: () => organizationAPI.getAll(params),
    // 缩短缓存时间以确保数据更新更及时
    staleTime: 30 * 1000, // 30秒内认为数据是新鲜的
    gcTime: 5 * 60 * 1000, // 5分钟后清理缓存
    refetchOnWindowFocus: true, // 窗口获得焦点时重新获取数据
    refetchOnMount: true, // 组件挂载时重新获取数据
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
    // 缩短缓存时间以确保统计数据更新及时
    staleTime: 30 * 1000, // 30秒内认为数据是新鲜的
    gcTime: 5 * 60 * 1000, // 5分钟后清理缓存
    refetchOnWindowFocus: true, // 窗口获得焦点时重新获取数据
    refetchOnMount: true, // 组件挂载时重新获取数据
  });
};