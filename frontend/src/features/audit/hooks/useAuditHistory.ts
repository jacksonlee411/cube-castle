import { useState, useEffect, useCallback } from 'react';
import { AuditAPI } from '../../../shared/api/audit';
import type { AuditQueryParams, OrganizationAuditHistory, AuditTimelineEntry } from '../../../shared/api/audit';

// Hook状态接口
interface UseAuditHistoryState {
  auditHistory: OrganizationAuditHistory | null;
  auditTimeline: AuditTimelineEntry[];
  loading: boolean;
  error: string | null;
  hasMore: boolean;
  totalRecords: number;
}

// Hook返回值接口
interface UseAuditHistoryReturn extends UseAuditHistoryState {
  refetch: () => Promise<void>;
  loadMore: () => Promise<void>;
  clearError: () => void;
  updateFilters: (params: AuditQueryParams) => void;
}

// Hook参数接口
interface UseAuditHistoryParams {
  code: string;
  initialParams?: AuditQueryParams;
  autoFetch?: boolean;
}

/**
 * 审计信息数据获取Hook
 * 提供审计数据获取、过滤、分页加载等功能
 */
export const useAuditHistory = ({
  code,
  initialParams = {},
  autoFetch = true
}: UseAuditHistoryParams): UseAuditHistoryReturn => {
  
  // 状态管理
  const [state, setState] = useState<UseAuditHistoryState>({
    auditHistory: null,
    auditTimeline: [],
    loading: false,
    error: null,
    hasMore: true,
    totalRecords: 0
  });

  // 当前查询参数
  const [currentParams, setCurrentParams] = useState<AuditQueryParams>({
    limit: 50,
    ...initialParams
  });

  // 已加载记录数追踪
  const [loadedCount, setLoadedCount] = useState(0);
  
  // 请求缓存机制 - 避免重复请求
  const [lastRequestKey, setLastRequestKey] = useState<string>('');

  /**
   * 获取审计信息数据
   */
  const fetchAuditHistory = useCallback(async (
    params: AuditQueryParams,
    append: boolean = false
  ) => {
    if (!code) return;
    
    // 生成请求唯一键避免重复请求
    const requestKey = `${code}-${JSON.stringify(params)}-${append}`;
    if (requestKey === lastRequestKey && !append) {
      console.log('[AuditHistory] 跳过重复请求:', requestKey);
      return; // 跳过重复请求
    }

    try {
      setState(prev => ({ ...prev, loading: true, error: null }));
      setLastRequestKey(requestKey);

      const auditHistory = await AuditAPI.getOrganizationAuditHistory(code, params);
      
      setState(prev => ({
        ...prev,
        auditHistory,
        auditTimeline: append ? 
          [...prev.auditTimeline, ...auditHistory.auditTimeline] :
          auditHistory.auditTimeline,
        loading: false,
        totalRecords: auditHistory.meta.totalAuditRecords,
        hasMore: auditHistory.auditTimeline.length === (params.limit || 50)
      }));

      if (append) {
        setLoadedCount(prev => prev + auditHistory.auditTimeline.length);
      } else {
        setLoadedCount(auditHistory.auditTimeline.length);
      }

    } catch (error) {
      console.error('Error fetching audit history:', error);
      setState(prev => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : '获取审计信息失败'
      }));
    }
  }, [code, lastRequestKey]);

  /**
   * 重新获取数据
   */
  const refetch = useCallback(async () => {
    setLoadedCount(0);
    await fetchAuditHistory(currentParams, false);
  }, [fetchAuditHistory, currentParams]);

  /**
   * 加载更多数据 (无限滚动)
   */
  const loadMore = useCallback(async () => {
    if (state.loading || !state.hasMore) return;

    // 计算偏移量进行分页
    const paramsWithOffset = {
      ...currentParams,
      // 这里可以根据后端API支持情况添加offset参数
      // offset: loadedCount
    };

    await fetchAuditHistory(paramsWithOffset, true);
  }, [state.loading, state.hasMore, loadedCount, currentParams, fetchAuditHistory]);

  /**
   * 清除错误状态
   */
  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  /**
   * 更新过滤参数 - 添加参数比较避免不必要的更新
   */
  const updateFilters = useCallback((params: AuditQueryParams) => {
    setCurrentParams(prev => {
      const newParams = {
        ...prev,
        ...params
      };
      
      // 比较参数是否真的发生了变化
      const prevStr = JSON.stringify(prev);
      const newStr = JSON.stringify(newParams);
      
      if (prevStr === newStr) {
        return prev; // 参数没有变化，返回原来的对象
      }
      
      return newParams;
    });
    setLoadedCount(0);
  }, []);

  // 自动获取数据 - 添加防抖机制避免频繁请求
  useEffect(() => {
    if (autoFetch && code) {
      // 防抖延迟200ms
      const debounceTimer = setTimeout(() => {
        fetchAuditHistory(currentParams, false);
      }, 200);
      
      return () => clearTimeout(debounceTimer);
    }
  }, [code, currentParams, autoFetch, fetchAuditHistory]);

  // 组织代码变化时重置状态
  useEffect(() => {
    setState({
      auditHistory: null,
      auditTimeline: [],
      loading: false,
      error: null,
      hasMore: true,
      totalRecords: 0
    });
    setLoadedCount(0);
    setLastRequestKey(''); // 清除请求缓存
  }, [code]);

  return {
    ...state,
    refetch,
    loadMore,
    clearError,
    updateFilters
  };
};