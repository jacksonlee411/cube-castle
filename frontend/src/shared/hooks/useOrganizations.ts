import { useQuery } from '@tanstack/react-query';
import { organizationAPI } from '../api';
import type { OrganizationQueryParams } from '../types/organization';

// 组织单元列表查询 - 修复无限循环问题
export const useOrganizations = (params?: OrganizationQueryParams) => {
  return useQuery({
    queryKey: ['organizations', JSON.stringify(params || {})], // 序列化参数避免引用变化
    queryFn: () => organizationAPI.getAll(params),
    // 合理的缓存时间设置
    staleTime: 30 * 1000, // 30秒内认为数据是新鲜的
    gcTime: 5 * 60 * 1000, // 5分钟后清理缓存
    refetchOnWindowFocus: false, // 避免过度刷新
    refetchOnMount: true, // 组件挂载时获取数据
    refetchOnReconnect: true, // 网络重连时重新获取
    // 移除自动轮询，避免无限循环
    refetchInterval: false,
    refetchIntervalInBackground: false,
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

// 组织统计信息查询 - 修复无限循环问题
export const useOrganizationStats = () => {
  return useQuery({
    queryKey: ['organization-stats'],
    queryFn: () => organizationAPI.getStats(),
    // 合理的缓存时间设置
    staleTime: 60 * 1000, // 60秒内认为数据是新鲜的
    gcTime: 5 * 60 * 1000, // 5分钟后清理缓存
    refetchOnWindowFocus: false, // 避免过度刷新
    refetchOnMount: true, // 组件挂载时获取数据
    refetchOnReconnect: true, // 网络重连时重新获取
    // 移除自动轮询，避免无限循环
    refetchInterval: false,
    refetchIntervalInBackground: false,
  });
};