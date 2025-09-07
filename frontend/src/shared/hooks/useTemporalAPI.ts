/**
 * 组织详情API客户端钩子 (纯日期生效模型)
 * 连接到端口9091的组织详情服务
 */
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { TemporalConverter } from '../utils/temporal-converter';
import { unifiedRESTClient } from '../api/unified-client';

// 组织详情API基础URL (通过unifiedRESTClient统一处理)

// 查询参数接口 (纯日期生效模型)
export interface TemporalQueryParams {
  asOfDate?: string;      // 时间点查询 YYYY-MM-DD
  effectiveFrom?: string;  // 时间范围开始 YYYY-MM-DD  
  effectiveTo?: string;    // 时间范围结束 YYYY-MM-DD
}

// 使用统一的时态组织单元接口
import type { TemporalOrganizationUnit } from '../types/temporal';

// 时态查询响应
export interface TemporalQueryResponse {
  organizations: TemporalOrganizationUnit[];
  queriedAt: string;
  queryOptions: TemporalQueryParams;
  resultCount: number;
}

// 健康检查响应
export interface TemporalHealthResponse {
  service: string;
  status: string;
  features: string[];
  timestamp: string;
}

/**
 * 组织详情服务健康检查
 */
export function useTemporalHealth() {
  return useQuery({
    queryKey: ['temporal', 'health'],
    queryFn: async (): Promise<TemporalHealthResponse> => {
      return await unifiedRESTClient.request('/health');
    },
    staleTime: 30 * 1000, // 30秒内认为数据新鲜
    gcTime: 60 * 1000, // 缓存1分钟
    retry: 3,
  });
}

/**
 * 时间点查询钩子
 */
export function useTemporalAsOfDateQuery(organizationCode: string, asOfDate: string, enabled = true) {
  return useQuery({
    queryKey: ['temporal', 'asOfDate', organizationCode, asOfDate],
    queryFn: async (): Promise<TemporalQueryResponse> => {
      return await unifiedRESTClient.request(`/organization-units/${organizationCode}/temporal?asOfDate=${asOfDate}`);
    },
    enabled: enabled && !!organizationCode && !!asOfDate,
    staleTime: 5 * 60 * 1000, // 5分钟内认为数据新鲜
    gcTime: 10 * 60 * 1000, // 缓存10分钟
  });
}

/**
 * 时间范围查询钩子
 */
export function useTemporalDateRangeQuery(
  organizationCode: string, 
  effectiveFrom: string, 
  effectiveTo: string,
  enabled = true
) {
  return useQuery({
    queryKey: ['temporal', 'dateRange', organizationCode, effectiveFrom, effectiveTo],
    queryFn: async (): Promise<TemporalQueryResponse> => {
      return await unifiedRESTClient.request(`/organization-units/${organizationCode}/temporal?effectiveFrom=${effectiveFrom}&effectiveTo=${effectiveTo}`);
    },
    enabled: enabled && !!organizationCode && !!effectiveFrom && !!effectiveTo,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
}

/**
 * 时态查询工具钩子
 */
export function useTemporalQueryUtils() {
  const queryClient = useQueryClient();

  // 清除时态查询缓存
  const clearTemporalCache = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: ['temporal'] });
  }, [queryClient]);

  // 预取组织的时态数据
  const prefetchTemporalData = useCallback(async (
    organizationCode: string,
    params: TemporalQueryParams
  ) => {
    if (params.asOfDate) {
      await queryClient.prefetchQuery({
        queryKey: ['temporal', 'asOfDate', organizationCode, params.asOfDate],
        queryFn: async () => {
          return await unifiedRESTClient.request(`/organization-units/${organizationCode}/temporal?asOfDate=${params.asOfDate}`);
        },
        staleTime: 5 * 60 * 1000,
      });
    }
    
    if (params.effectiveFrom && params.effectiveTo) {
      await queryClient.prefetchQuery({
        queryKey: ['temporal', 'dateRange', organizationCode, params.effectiveFrom, params.effectiveTo],
        queryFn: async () => {
          return await unifiedRESTClient.request(`/organization-units/${organizationCode}/temporal?effectiveFrom=${params.effectiveFrom}&effectiveTo=${params.effectiveTo}`);
        },
        staleTime: 5 * 60 * 1000,
      });
    }
  }, [queryClient]);

  // 格式化日期为API需要的格式
  const formatDateForAPI = useCallback((date: Date): string => {
    return date.toISOString().slice(0, 10); // YYYY-MM-DD
  }, []);

  // 解析API返回的日期
  const parseAPIDate = useCallback((dateString: string): Date => {
    return new Date(dateString);
  }, []);

  // 检查记录是否在指定时间点有效
  const isRecordValidAt = useCallback((record: TemporalOrganizationUnit, date: Date): boolean => {
    const effectiveDate = new Date(record.effectiveDate);
    const endDate = record.endDate ? new Date(record.endDate) : null;
    
    return effectiveDate <= date && (!endDate || date < endDate);
  }, []);

  // 获取记录的有效期描述
  const getRecordValidityDescription = useCallback((record: TemporalOrganizationUnit): string => {
    const effectiveDate = new Date(record.effectiveDate);
    const endDate = record.endDate ? new Date(record.endDate) : null;
    
    const effectiveDateStr = effectiveDate.toLocaleDateString('zh-CN');
    
    if (endDate) {
      const endDateStr = endDate.toLocaleDateString('zh-CN');
      return `${effectiveDateStr} - ${endDateStr}`;
    } else {
      return `${effectiveDateStr} 起生效`;
    }
  }, []);

  return {
    clearTemporalCache,
    prefetchTemporalData,
    formatDateForAPI,
    parseAPIDate,
    isRecordValidAt,
    getRecordValidityDescription,
  };
}

/**
 * 时态查询统计钩子
 */
export function useTemporalQueryStats() {
  const queryClient = useQueryClient();
  
  // 获取缓存统计
  const getCacheStats = useCallback(() => {
    const queryCache = queryClient.getQueryCache();
    const temporalQueries = queryCache.findAll({ queryKey: ['temporal'] });
    
    return {
      totalQueries: temporalQueries.length,
      cachedQueries: temporalQueries.filter(q => q.state.data).length,
      failedQueries: temporalQueries.filter(q => q.state.error).length,
      loadingQueries: temporalQueries.filter(q => q.state.status === 'pending').length,
    };
  }, [queryClient]);
  
  return {
    getCacheStats,
  };
}

// 错误类型
export class TemporalAPIError extends Error {
  constructor(
    message: string,
    public errorCode?: string,
    public details?: string
  ) {
    super(message);
    this.name = 'TemporalAPIError';
  }
}

// 导出常用日期格式化函数 (统一使用TemporalConverter)
export const TemporalDateUtils = {
  today: () => TemporalConverter.getCurrentDateString(),
  yesterday: () => {
    const date = new Date();
    date.setDate(date.getDate() - 1);
    return TemporalConverter.dateToDateString(date);
  },
  nextWeek: () => {
    const date = new Date();
    date.setDate(date.getDate() + 7);
    return TemporalConverter.dateToDateString(date);
  },
  startOfMonth: () => {
    const date = new Date();
    date.setDate(1);
    return TemporalConverter.dateToDateString(date);
  },
  endOfMonth: () => {
    const date = new Date();
    date.setMonth(date.getMonth() + 1, 0);
    return TemporalConverter.dateToDateString(date);
  },
};