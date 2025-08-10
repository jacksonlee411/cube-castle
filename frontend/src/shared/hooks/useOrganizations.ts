import { useQuery } from '@tanstack/react-query';
import { organizationAPI, type OrganizationQueryParams } from '../api';

// 组织单元列表查询 - 优化缓存失效策略
export const useOrganizations = (params?: OrganizationQueryParams) => {
  return useQuery({
    queryKey: ['organizations', params],
    queryFn: () => organizationAPI.getAll(params),
    // 大幅缩短缓存时间以确保数据更新更及时
    staleTime: 5 * 1000, // 5秒内认为数据是新鲜的（从30秒减少）
    gcTime: 2 * 60 * 1000, // 2分钟后清理缓存（从5分钟减少）
    refetchOnWindowFocus: true, // 窗口获得焦点时重新获取数据
    refetchOnMount: true, // 组件挂载时重新获取数据
    // 新增：网络重连时自动重新获取
    refetchOnReconnect: true,
    // 新增：确保后台更新时立即重新获取
    refetchInterval: 30 * 1000, // 每30秒后台轮询一次
    refetchIntervalInBackground: false, // 仅在前台时轮询
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

// 组织统计信息查询 - 优化缓存失效策略
export const useOrganizationStats = () => {
  return useQuery({
    queryKey: ['organization-stats'],
    queryFn: () => organizationAPI.getStats(),
    // 大幅缩短缓存时间以确保统计数据更新及时
    staleTime: 5 * 1000, // 5秒内认为数据是新鲜的（从30秒减少）
    gcTime: 2 * 60 * 1000, // 2分钟后清理缓存（从5分钟减少）
    refetchOnWindowFocus: true, // 窗口获得焦点时重新获取数据
    refetchOnMount: true, // 组件挂载时重新获取数据
    // 新增：网络重连时自动重新获取
    refetchOnReconnect: true,
    // 新增：确保统计数据后台更新时立即重新获取
    refetchInterval: 30 * 1000, // 每30秒后台轮询一次
    refetchIntervalInBackground: false, // 仅在前台时轮询
  });
};