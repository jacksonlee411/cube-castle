/**
 * 时态管理React钩子 - 基于GraphQL时态查询
 * 提供organizationAsOfDate和organizationHistory的React集成
 */
import { useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import temporalAPI from '../api/temporal-graphql-client';
import type { 
  TemporalQueryParams,
  TemporalOrganizationUnit,
  TimelineEvent,
  TemporalMode
} from '../types/temporal';

// 时态查询键生成器
const TEMPORAL_QUERY_KEYS = {
  temporal: ['temporal'] as const,
  asOfDate: (code: string, asOfDate: string) => 
    [...TEMPORAL_QUERY_KEYS.temporal, 'asOfDate', code, asOfDate] as const,
  history: (code: string, fromDate?: string, toDate?: string) => 
    [...TEMPORAL_QUERY_KEYS.temporal, 'history', code, fromDate, toDate] as const,
  timeline: (code: string, params?: TemporalQueryParams) => 
    [...TEMPORAL_QUERY_KEYS.temporal, 'timeline', code, params] as const,
  stats: (code: string, fromDate?: string, toDate?: string) => 
    [...TEMPORAL_QUERY_KEYS.temporal, 'stats', code, fromDate, toDate] as const,
  batch: (codes: string[], asOfDate: string) => 
    [...TEMPORAL_QUERY_KEYS.temporal, 'batch', codes.sort().join(','), asOfDate] as const,
};

// 缓存配置
const TEMPORAL_CACHE_CONFIG = {
  staleTime: 10 * 60 * 1000, // 10分钟 - 历史数据更稳定
  gcTime: 30 * 60 * 1000,    // 30分钟
  refetchOnWindowFocus: false,
  refetchOnMount: false,      // 历史数据不需要重新加载
  retry: 2,
};

/**
 * 时间点查询钩子 - 查询特定时间点的组织状态
 */
export function useOrganizationAsOfDate(
  code: string,
  asOfDate: string,
  options?: {
    enabled?: boolean;
    onSuccess?: (data: TemporalOrganizationUnit | null) => void;
    onError?: (error: Error) => void;
  }
): UseQueryResult<TemporalOrganizationUnit | null> & {
  hasData: boolean;
  isEmpty: boolean;
  isHistoricalRecord: boolean;
} {
  const query = useQuery({
    queryKey: TEMPORAL_QUERY_KEYS.asOfDate(code, asOfDate),
    queryFn: () => temporalAPI.getOrganizationAsOfDate(code, asOfDate),
    ...TEMPORAL_CACHE_CONFIG,
    enabled: !!(code && asOfDate && options?.enabled !== false),
    onSuccess: options?.onSuccess,
    onError: options?.onError,
  });

  const hasData = !!query.data;
  const isEmpty = !query.isLoading && !query.data;
  const isHistoricalRecord = hasData ? !query.data!.is_current : false;

  return {
    ...query,
    hasData,
    isEmpty,
    isHistoricalRecord
  };
}

/**
 * 组织历史查询钩子 - 查询完整历史记录
 */
export function useOrganizationHistory(
  code: string,
  params?: {
    fromDate?: string;
    toDate?: string;
    enabled?: boolean;
  }
): UseQueryResult<TemporalOrganizationUnit[]> & {
  hasHistory: boolean;
  historyCount: number;
  latestRecord: TemporalOrganizationUnit | undefined;
  currentRecord: TemporalOrganizationUnit | undefined;
  historicalRecords: TemporalOrganizationUnit[];
} {
  const { fromDate, toDate, enabled = true } = params || {};

  const query = useQuery({
    queryKey: TEMPORAL_QUERY_KEYS.history(code, fromDate, toDate),
    queryFn: () => temporalAPI.getOrganizationHistory(code, { fromDate, toDate }),
    ...TEMPORAL_CACHE_CONFIG,
    enabled: !!(code && enabled),
  });

  const historyCount = query.data?.length || 0;
  const hasHistory = historyCount > 1;
  const latestRecord = query.data?.[0]; // 按时间倒序，第一个是最新的
  const currentRecord = query.data?.find(record => record.is_current);
  const historicalRecords = query.data?.filter(record => !record.is_current) || [];

  return {
    ...query,
    hasHistory,
    historyCount,
    latestRecord,
    currentRecord,
    historicalRecords
  };
}

/**
 * 组织时间线查询钩子 - 基于历史记录生成时间线
 */
export function useOrganizationTimeline(
  code: string,
  params?: TemporalQueryParams & { enabled?: boolean }
): UseQueryResult<TimelineEvent[]> & {
  hasEvents: boolean;
  eventCount: number;
  latestEvent: TimelineEvent | undefined;
  eventsGroupedByYear: Record<string, TimelineEvent[]>;
} {
  const { enabled = true, ...timelineParams } = params || {};

  const query = useQuery({
    queryKey: TEMPORAL_QUERY_KEYS.timeline(code, timelineParams),
    queryFn: () => temporalAPI.getOrganizationTimeline(code, timelineParams),
    ...TEMPORAL_CACHE_CONFIG,
    enabled: !!(code && enabled),
  });

  const eventCount = query.data?.length || 0;
  const hasEvents = eventCount > 0;
  const latestEvent = query.data?.[0]; // 最新事件

  // 按年份分组事件
  const eventsGroupedByYear = query.data?.reduce((groups, event) => {
    const year = new Date(event.eventDate).getFullYear().toString();
    if (!groups[year]) {
      groups[year] = [];
    }
    groups[year].push(event);
    return groups;
  }, {} as Record<string, TimelineEvent[]>) || {};

  return {
    ...query,
    hasEvents,
    eventCount,
    latestEvent,
    eventsGroupedByYear
  };
}

/**
 * 批量时间点查询钩子 - 多个组织的同一时间点查询
 */
export function useBatchOrganizationsAsOfDate(
  codes: string[],
  asOfDate: string,
  options?: {
    enabled?: boolean;
    onSuccess?: (data: Record<string, TemporalOrganizationUnit | null>) => void;
  }
): UseQueryResult<Record<string, TemporalOrganizationUnit | null>> & {
  successCount: number;
  failureCount: number;
  hasAnyData: boolean;
} {
  const query = useQuery({
    queryKey: TEMPORAL_QUERY_KEYS.batch(codes, asOfDate),
    queryFn: () => temporalAPI.getBatchOrganizationsAsOfDate(codes, asOfDate),
    ...TEMPORAL_CACHE_CONFIG,
    enabled: !!(codes.length > 0 && asOfDate && options?.enabled !== false),
    onSuccess: options?.onSuccess,
  });

  const successCount = Object.values(query.data || {}).filter(Boolean).length;
  const failureCount = Object.values(query.data || {}).filter(data => data === null).length;
  const hasAnyData = successCount > 0;

  return {
    ...query,
    successCount,
    failureCount,
    hasAnyData
  };
}

/**
 * 组织变更统计钩子
 */
export function useOrganizationChangeStats(
  code: string,
  params?: {
    fromDate?: string;
    toDate?: string;
    enabled?: boolean;
  }
) {
  const { fromDate, toDate, enabled = true } = params || {};

  const query = useQuery({
    queryKey: TEMPORAL_QUERY_KEYS.stats(code, fromDate, toDate),
    queryFn: () => temporalAPI.getOrganizationChangeStats(code, { fromDate, toDate }),
    ...TEMPORAL_CACHE_CONFIG,
    enabled: !!(code && enabled),
  });

  return query;
}

/**
 * 时态查询缓存管理钩子
 */
export function useTemporalCacheManager() {
  const queryClient = useQueryClient();

  const invalidateAsOfDateCache = useCallback(
    (code?: string) => {
      const queryKey = code 
        ? TEMPORAL_QUERY_KEYS.asOfDate(code, '*' as any)
        : TEMPORAL_QUERY_KEYS.temporal;
      return queryClient.invalidateQueries({ queryKey });
    },
    [queryClient]
  );

  const invalidateHistoryCache = useCallback(
    (code?: string) => {
      const queryKey = code
        ? TEMPORAL_QUERY_KEYS.history(code)
        : TEMPORAL_QUERY_KEYS.temporal;
      return queryClient.invalidateQueries({ queryKey });
    },
    [queryClient]
  );

  const clearAllTemporalCache = useCallback(
    () => {
      return queryClient.invalidateQueries({ queryKey: TEMPORAL_QUERY_KEYS.temporal });
    },
    [queryClient]
  );

  const prefetchAsOfDate = useCallback(
    (code: string, asOfDate: string) => {
      return queryClient.prefetchQuery({
        queryKey: TEMPORAL_QUERY_KEYS.asOfDate(code, asOfDate),
        queryFn: () => temporalAPI.getOrganizationAsOfDate(code, asOfDate),
        ...TEMPORAL_CACHE_CONFIG
      });
    },
    [queryClient]
  );

  const prefetchHistory = useCallback(
    (code: string, params?: { fromDate?: string; toDate?: string }) => {
      return queryClient.prefetchQuery({
        queryKey: TEMPORAL_QUERY_KEYS.history(code, params?.fromDate, params?.toDate),
        queryFn: () => temporalAPI.getOrganizationHistory(code, params),
        ...TEMPORAL_CACHE_CONFIG
      });
    },
    [queryClient]
  );

  return {
    invalidateAsOfDateCache,
    invalidateHistoryCache,
    clearAllTemporalCache,
    prefetchAsOfDate,
    prefetchHistory
  };
}

/**
 * 时态查询工具钩子 - 提供常用的时态查询辅助功能
 */
export function useTemporalQueryUtils() {
  // 生成常用的时间点
  const getCommonDatePoints = useCallback(() => {
    const now = new Date();
    const currentYear = now.getFullYear();
    
    return {
      today: now.toISOString().split('T')[0],
      yesterday: new Date(now.getTime() - 24 * 60 * 60 * 1000).toISOString().split('T')[0],
      lastWeek: new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
      lastMonth: new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
      yearStart: `${currentYear}-01-01`,
      lastYearEnd: `${currentYear - 1}-12-31`,
    };
  }, []);

  // 格式化时态查询结果用于展示
  const formatTemporalRecord = useCallback(
    (record: TemporalOrganizationUnit) => {
      return {
        displayName: record.name,
        effectivePeriod: record.end_date 
          ? `${record.effective_date} 至 ${record.end_date}`
          : `${record.effective_date} 起生效`,
        status: record.is_current ? '当前有效' : '历史记录',
        changeReason: record.change_reason || '无变更说明',
        organizationType: record.unit_type,
        organizationStatus: record.status
      };
    },
    []
  );

  // 比较两个时态记录的差异
  const compareRecords = useCallback(
    (record1: TemporalOrganizationUnit, record2: TemporalOrganizationUnit) => {
      const changes: Array<{
        field: string;
        oldValue: any;
        newValue: any;
        displayName: string;
      }> = [];

      const fieldMappings = {
        name: '组织名称',
        unit_type: '组织类型',
        status: '状态',
        parent_code: '上级组织',
        description: '描述',
        level: '级别'
      };

      Object.entries(fieldMappings).forEach(([field, displayName]) => {
        const oldValue = (record1 as any)[field];
        const newValue = (record2 as any)[field];
        
        if (oldValue !== newValue) {
          changes.push({
            field,
            oldValue,
            newValue,
            displayName
          });
        }
      });

      return changes;
    },
    []
  );

  return {
    getCommonDatePoints,
    formatTemporalRecord,
    compareRecords
  };
}

// 导出查询键和缓存配置供外部使用
export {
  TEMPORAL_QUERY_KEYS,
  TEMPORAL_CACHE_CONFIG
};