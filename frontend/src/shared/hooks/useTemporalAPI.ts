/**
 * 组织详情API客户端钩子 (纯日期生效模型)
 * 连接到端口9091的组织详情服务
 */
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { TemporalConverter } from '../utils/temporal-converter';

// 组织详情API基础URL
const TEMPORAL_API_BASE = 'http://localhost:9091/api/v1';

// 查询参数接口 (纯日期生效模型)
export interface TemporalQueryParams {
  asOfDate?: string;      // 时间点查询 YYYY-MM-DD
  effectiveFrom?: string;  // 时间范围开始 YYYY-MM-DD  
  effectiveTo?: string;    // 时间范围结束 YYYY-MM-DD
}

// 时态组织记录 (纯日期生效模型)
export interface TemporalOrganizationRecord {
  tenantId: string;
  code: string;
  name: string;
  unitType: string;
  status: string;
  level: number;
  path: string;
  sortOrder: number;
  description?: string;
  createdAt: string;
  updatedAt: string;
  effectiveDate: string;   // 生效日期
  endDate?: string;        // 结束日期 (可选)
  isCurrent: boolean;      // 是否当前有效
  changeReason?: string;   // 变更原因
  approvedBy?: string;     // 批准人
  approvedAt?: string;     // 批准时间
}

// 时态查询响应
export interface TemporalQueryResponse {
  organizations: TemporalOrganizationRecord[];
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
      const response = await fetch(`${TEMPORAL_API_BASE.replace('/api/v1', '')}/health`);
      if (!response.ok) {
        throw new Error('组织详情服务不可用');
      }
      return response.json();
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
      const url = `${TEMPORAL_API_BASE}/organization-units/${organizationCode}/temporal?as_of_date=${asOfDate}`;
      const response = await fetch(url);
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || '时态查询失败');
      }
      
      return response.json();
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
      const url = `${TEMPORAL_API_BASE}/organization-units/${organizationCode}/temporal?effective_from=${effectiveFrom}&effective_to=${effectiveTo}`;
      const response = await fetch(url);
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || '时态范围查询失败');
      }
      
      return response.json();
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
          const url = `${TEMPORAL_API_BASE}/organization-units/${organizationCode}/temporal?as_of_date=${params.asOfDate}`;
          const response = await fetch(url);
          return response.json();
        },
        staleTime: 5 * 60 * 1000,
      });
    }
    
    if (params.effectiveFrom && params.effectiveTo) {
      await queryClient.prefetchQuery({
        queryKey: ['temporal', 'dateRange', organizationCode, params.effectiveFrom, params.effectiveTo],
        queryFn: async () => {
          const url = `${TEMPORAL_API_BASE}/organization-units/${organizationCode}/temporal?effective_from=${params.effectiveFrom}&effective_to=${params.effectiveTo}`;
          const response = await fetch(url);
          return response.json();
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
  const isRecordValidAt = useCallback((record: TemporalOrganizationRecord, date: Date): boolean => {
    const effectiveDate = new Date(record.effectiveDate);
    const endDate = record.endDate ? new Date(record.endDate) : null;
    
    return effectiveDate <= date && (!endDate || date < endDate);
  }, []);

  // 获取记录的有效期描述
  const getRecordValidityDescription = useCallback((record: TemporalOrganizationRecord): string => {
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