/**
 * 时态查询钩子函数
 * 提供时态数据查询、缓存和状态管理功能
 */
import { useState, useEffect, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import organizationAPI from '../api/organizations-simplified';
import { useTemporalStore, useTemporalActions, temporalSelectors } from '../stores/temporalStore';
import type { 
  TemporalQueryParams,
  TemporalOrganizationUnit,
  TimelineEvent,
  TemporalMode
} from '../types/temporal';
import type { OrganizationUnit, OrganizationQueryParams } from '../types/organization';

// 查询键生成器
const QUERY_KEYS = {
  temporal: ['temporal'] as const,
  organizations: (params: TemporalQueryParams) => 
    [...QUERY_KEYS.temporal, 'organizations', params] as const,
  organization: (code: string, params?: TemporalQueryParams) => 
    [...QUERY_KEYS.temporal, 'organization', code, params] as const,
  history: (code: string, params?: TemporalQueryParams) => 
    [...QUERY_KEYS.temporal, 'history', code, params] as const,
  timeline: (code: string, params?: TemporalQueryParams) => 
    [...QUERY_KEYS.temporal, 'timeline', code, params] as const,
};

// 缓存配置
const CACHE_CONFIG = {
  staleTime: 5 * 60 * 1000, // 5分钟
  cacheTime: 10 * 60 * 1000, // 10分钟
  refetchOnWindowFocus: false,
  refetchOnMount: true,
};

// 时态组织查询钩子
export function useTemporalOrganizations(
  queryParams?: Partial<OrganizationQueryParams>
): UseQueryResult<OrganizationUnit[]> & {
  temporalContext: ReturnType<typeof temporalSelectors.useContext>;
  isHistorical: boolean;
} {
  const temporalContext = temporalSelectors.useContext();
  const temporalQueryParams = temporalSelectors.useQueryParams();
  
  // 使用稳定的actions引用
  const { getCachedOrganizations, cacheOrganizations } = useTemporalActions();

  const queryKey = QUERY_KEYS.organizations(temporalQueryParams);
  const cacheKey = JSON.stringify(queryKey);

  const query = useQuery({
    queryKey,
    queryFn: async () => {
      // 尝试从缓存获取
      const cached = getCachedOrganizations(cacheKey);
      if (cached) {
        return cached;
      }

      // 合并查询参数
      const params: OrganizationQueryParams = {
        ...queryParams,
        temporalParams: temporalQueryParams
      };

      // 调用API
      const response = await organizationAPI.getAll(params);
      const organizations = response.organizations;

      // 缓存结果
      cacheOrganizations(cacheKey, organizations);

      return organizations;
    },
    ...CACHE_CONFIG,
    enabled: !!temporalContext,
  });

  return {
    ...query,
    temporalContext,
    isHistorical: temporalContext.mode === 'historical'
  };
}

// 时态单个组织查询钩子
export function useTemporalOrganization(
  code: string,
  enabled: boolean = true
): UseQueryResult<OrganizationUnit> & {
  temporalContext: ReturnType<typeof temporalSelectors.useContext>;
  isHistorical: boolean;
} {
  const temporalContext = temporalSelectors.useContext();
  const temporalQueryParams = temporalSelectors.useQueryParams();

  const query = useQuery({
    queryKey: QUERY_KEYS.organization(code, temporalQueryParams),
    queryFn: async () => {
      return await organizationAPI.getByCode(code, temporalQueryParams);
    },
    ...CACHE_CONFIG,
    enabled: enabled && !!code && !!temporalContext,
  });

  return {
    ...query,
    temporalContext,
    isHistorical: temporalContext.mode === 'historical'
  };
}

// 组织历史查询钩子
export function useOrganizationHistory(
  code: string,
  params?: Partial<TemporalQueryParams>,
  enabled: boolean = true
): UseQueryResult<TemporalOrganizationUnit[]> & {
  hasHistory: boolean;
  latestVersion: TemporalOrganizationUnit | undefined;
} {
  const temporalQueryParams = temporalSelectors.useQueryParams();
  
  // 使用稳定的actions引用
  const { getCachedOrganizations, cacheOrganizations } = useTemporalActions();

  const finalParams = { ...temporalQueryParams, ...params };
  const queryKey = QUERY_KEYS.history(code, finalParams);
  const cacheKey = JSON.stringify(queryKey);

  const query = useQuery({
    queryKey,
    queryFn: async () => {
      // 尝试从缓存获取
      const cached = getCachedOrganizations(cacheKey);
      if (cached) {
        return cached;
      }

      // 调用API
      const history = await organizationAPI.getHistory(code, finalParams);

      // 缓存结果
      cacheOrganizations(cacheKey, history);

      return history;
    },
    ...CACHE_CONFIG,
    enabled: enabled && !!code,
  });

  const hasHistory = (query.data?.length ?? 0) > 1;
  const latestVersion = query.data?.[0]; // 假设按时间倒序排列

  return {
    ...query,
    hasHistory,
    latestVersion
  };
}

// 组织时间线查询钩子
export function useOrganizationTimeline(
  code: string,
  params?: Partial<TemporalQueryParams>,
  enabled: boolean = true
): UseQueryResult<TimelineEvent[]> & {
  hasEvents: boolean;
  eventCount: number;
  latestEvent: TimelineEvent | undefined;
} {
  const temporalQueryParams = temporalSelectors.useQueryParams();
  
  // 使用稳定的actions引用
  const { getCachedTimeline, cacheTimeline } = useTemporalActions();

  const finalParams = { ...temporalQueryParams, ...params };
  const queryKey = QUERY_KEYS.timeline(code, finalParams);
  const cacheKey = JSON.stringify(queryKey);

  const query = useQuery({
    queryKey,
    queryFn: async () => {
      // 尝试从缓存获取
      const cached = getCachedTimeline(cacheKey);
      if (cached) {
        return cached;
      }

      // 调用API
      const timeline = await organizationAPI.getTimeline(code, finalParams);

      // 缓存结果
      cacheTimeline(cacheKey, timeline);

      return timeline;
    },
    ...CACHE_CONFIG,
    enabled: enabled && !!code,
  });

  const eventCount = query.data?.length ?? 0;
  const hasEvents = eventCount > 0;
  const latestEvent = query.data?.[0]; // 假设按时间倒序排列

  return {
    ...query,
    hasEvents,
    eventCount,
    latestEvent
  };
}

// 时态模式切换钩子
export function useTemporalMode() {
  const mode = useTemporalStore((state) => state.context.mode);
  
  // 使用稳定的actions引用
  const { setMode, setAsOfDate, reset } = useTemporalActions();
  const queryClient = useQueryClient();

  const switchToMode = useCallback(
    async (newMode: TemporalMode, options?: { asOfDate?: string; clearCache?: boolean }) => {
      // 清除相关查询缓存
      if (options?.clearCache) {
        await queryClient.invalidateQueries({ queryKey: QUERY_KEYS.temporal });
      }

      // 设置新模式
      setMode(newMode);

      // 设置时间点（如果提供）
      if (options?.asOfDate) {
        setAsOfDate(options.asOfDate);
      } else if (newMode === 'current') {
        setAsOfDate(new Date().toISOString());
      }
    },
    [setMode, setAsOfDate, queryClient]
  );

  const switchToCurrent = useCallback(() => 
    switchToMode('current', { clearCache: true }), [switchToMode]);

  const switchToHistorical = useCallback((asOfDate: string) => 
    switchToMode('historical', { asOfDate, clearCache: true }), [switchToMode]);

  const switchToPlanning = useCallback((asOfDate?: string) => 
    switchToMode('planning', { asOfDate, clearCache: true }), [switchToMode]);

  return {
    mode,
    switchToMode,
    switchToCurrent,
    switchToHistorical,
    switchToPlanning,
    isHistorical: mode === 'historical',
    isPlanning: mode === 'planning',
    isCurrent: mode === 'current'
  };
}

// 时态查询状态钩子
export function useTemporalQueryState() {
  const loading = temporalSelectors.useLoading();
  const error = temporalSelectors.useError();
  const context = temporalSelectors.useContext();
  const queryParams = temporalSelectors.useQueryParams();
  const cacheStats = temporalSelectors.useCacheStats();

  // 使用稳定的actions引用
  const { setError, clearCache } = useTemporalActions();

  const clearError = useCallback(() => setError(null), [setError]);

  const refreshCache = useCallback(async () => {
    const queryClient = useQueryClient();
    clearCache();
    await queryClient.invalidateQueries({ queryKey: QUERY_KEYS.temporal });
  }, [clearCache]);

  return {
    loading,
    error,
    context,
    queryParams,
    cacheStats,
    clearError,
    refreshCache,
    isLoading: Object.values(loading).some(Boolean),
    hasError: !!error
  };
}

// 时态数据预加载钩子
export function useTemporalPreloader() {
  const queryClient = useQueryClient();
  const temporalQueryParams = temporalSelectors.useQueryParams();

  const preloadOrganizations = useCallback(
    async (queryParams?: Partial<OrganizationQueryParams>) => {
      const params = { ...queryParams, temporalParams: temporalQueryParams };
      await queryClient.prefetchQuery({
        queryKey: QUERY_KEYS.organizations(temporalQueryParams),
        queryFn: () => organizationAPI.getAll(params),
        ...CACHE_CONFIG,
      });
    },
    [queryClient, temporalQueryParams]
  );

  const preloadOrganization = useCallback(
    async (code: string) => {
      await queryClient.prefetchQuery({
        queryKey: QUERY_KEYS.organization(code, temporalQueryParams),
        queryFn: () => organizationAPI.getByCode(code, temporalQueryParams),
        ...CACHE_CONFIG,
      });
    },
    [queryClient, temporalQueryParams]
  );

  const preloadHistory = useCallback(
    async (code: string, params?: Partial<TemporalQueryParams>) => {
      const finalParams = { ...temporalQueryParams, ...params };
      await queryClient.prefetchQuery({
        queryKey: QUERY_KEYS.history(code, finalParams),
        queryFn: () => organizationAPI.getHistory(code, finalParams),
        ...CACHE_CONFIG,
      });
    },
    [queryClient, temporalQueryParams]
  );

  const preloadTimeline = useCallback(
    async (code: string, params?: Partial<TemporalQueryParams>) => {
      const finalParams = { ...temporalQueryParams, ...params };
      await queryClient.prefetchQuery({
        queryKey: QUERY_KEYS.timeline(code, finalParams),
        queryFn: () => organizationAPI.getTimeline(code, finalParams),
        ...CACHE_CONFIG,
      });
    },
    [queryClient, temporalQueryParams]
  );

  return {
    preloadOrganizations,
    preloadOrganization,
    preloadHistory,
    preloadTimeline
  };
}

// 时态查询工具钩子
export function useTemporalUtils() {
  // 使用稳定的actions引用
  const { setQueryParams, setViewConfig } = useTemporalActions();

  const setDateRange = useCallback(
    (start: string, end: string) => {
      setQueryParams({ dateRange: { start, end } });
    },
    [setQueryParams]
  );

  const setAsOfDate = useCallback(
    (date: string) => {
      setQueryParams({ asOfDate: date });
    },
    [setQueryParams]
  );

  const toggleIncludeInactive = useCallback(() => {
    setQueryParams((prev) => ({ includeInactive: !prev.includeInactive }));
  }, [setQueryParams]);

  const setEventTypes = useCallback(
    (eventTypes: string[]) => {
      setQueryParams({ eventTypes });
    },
    [setQueryParams]
  );

  return {
    setDateRange,
    setAsOfDate,
    toggleIncludeInactive,
    setEventTypes
  };
}

// 导出所有钩子
export {
  QUERY_KEYS as temporalQueryKeys,
  CACHE_CONFIG as temporalCacheConfig
};