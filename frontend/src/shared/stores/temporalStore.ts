/**
 * 组织详情全局状态存储 (统一字符串类型版本)
 * 使用Zustand提供轻量级状态管理
 */
import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';
import type { 
  TemporalMode,
  TemporalQueryParams,
  DateRange,
  TemporalOrganizationUnit,
  TimelineEvent,
  TemporalTimelineViewConfig,
  TemporalContext
} from '../types/temporal';
import { TemporalConverter } from '../utils/temporal-converter';

// 时态存储状态接口
export interface TemporalState {
  // 当前时态上下文
  context: TemporalContext;
  
  // 时态查询参数
  queryParams: TemporalQueryParams;
  
  // 时间线视图配置
  viewConfig: TemporalTimelineViewConfig;
  
  // 缓存的时态数据
  cache: {
    organizations: Map<string, TemporalOrganizationUnit[]>;
    timelines: Map<string, TimelineEvent[]>;
    lastUpdated: Map<string, number>;
  };
  
  // 加载状态
  loading: {
    organizations: boolean;
    timeline: boolean;
    history: boolean;
  };
  
  // 错误状态
  error: string | null;
}

// 时态存储操作接口 (统一字符串类型)
export interface TemporalActions {
  // 设置时态模式
  setMode: (mode: TemporalMode) => void;
  
  // 设置查询时间点 (统一为字符串)
  setAsOfDate: (date: string) => void;
  
  // 设置时间范围 (统一为字符串)
  setDateRange: (range: DateRange) => void;
  
  // 设置查询参数
  setQueryParams: (params: Partial<TemporalQueryParams>) => void;
  
  // 设置时间线视图配置
  setViewConfig: (config: Partial<TemporalTimelineViewConfig>) => void;
  
  // 缓存操作
  cacheOrganizations: (key: string, data: TemporalOrganizationUnit[]) => void;
  cacheTimeline: (key: string, data: TimelineEvent[]) => void;
  getCachedOrganizations: (key: string) => TemporalOrganizationUnit[] | undefined;
  getCachedTimeline: (key: string) => TimelineEvent[] | undefined;
  clearCache: (key?: string) => void;
  
  // 加载状态管理
  setLoading: (type: keyof TemporalState['loading'], loading: boolean) => void;
  
  // 错误状态管理
  setError: (error: string | null) => void;
  
  // 重置状态
  reset: () => void;
}

// 默认状态 (统一字符串类型)
const defaultState: TemporalState = {
  context: {
    mode: 'current',
    currentDate: TemporalConverter.getCurrentISOString(),
    viewConfig: {
      showEvents: true,
      showRecords: true,  // 修正：替换showVersions为showRecords
      dateFormat: 'YYYY-MM-DD',
      timeRange: {
        start: TemporalConverter.dateToIso(new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)),
        end: TemporalConverter.getCurrentISOString()
      },
      eventTypes: []
    },
    permissions: {
      canViewHistory: true,
      canViewFuture: true,
      canCreatePlannedChanges: false,
      canModifyHistory: false,
      canCancelPlannedChanges: false
    },
    cacheConfig: {
      currentDataTTL: 300,
      historicalDataTTL: 3600,
      maxRecordsCache: 100,  // 修正：替换maxVersionsCache为maxRecordsCache
      enablePrefetch: false
    }
  },
  
  queryParams: {
    asOfDate: TemporalConverter.getCurrentISOString(),
    dateRange: {
      start: TemporalConverter.dateToIso(new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)),
      end: TemporalConverter.getCurrentISOString()
    },
    includeHistory: false,
    includeFuture: false
  },
  
  viewConfig: {
    showEvents: true,
    showRecords: true,  // 修正：替换showVersions为showRecords
    dateFormat: 'YYYY-MM-DD',
    timeRange: {
      start: TemporalConverter.dateToIso(new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)),
      end: TemporalConverter.getCurrentISOString()
    },
    eventTypes: []
  },
  
  cache: {
    organizations: new Map(),
    timelines: new Map(),
    lastUpdated: new Map()
  },
  
  loading: {
    organizations: false,
    timeline: false,
    history: false
  },
  
  error: null
};

// 创建时态状态存储
export const useTemporalStore = create<TemporalState & TemporalActions>()(
  subscribeWithSelector((set, get) => ({
    ...defaultState,

    // 设置时态模式
    setMode: (mode: TemporalMode) => {
      set((state) => ({
        context: { ...state.context, mode },
        queryParams: state.queryParams,
        error: null
      }));
    },

    // 设置查询时间点 (统一字符串类型)
    setAsOfDate: (date: string) => {
      // 验证并标准化日期字符串
      const normalizedDate = TemporalConverter.dateToIso(date);
      set((state) => ({
        context: state.context,
        queryParams: { ...state.queryParams, asOfDate: normalizedDate },
        error: null
      }));
    },

    // 设置时间范围 (统一字符串类型)
    setDateRange: (range: DateRange) => {
      // 标准化时间范围
      const normalizedRange: DateRange = {
        start: TemporalConverter.dateToIso(range.start),
        end: TemporalConverter.dateToIso(range.end)
      };
      set((state) => ({
        queryParams: { ...state.queryParams, dateRange: normalizedRange },
        error: null
      }));
    },

    // 设置查询参数 (统一字符串类型)
    setQueryParams: (params: Partial<TemporalQueryParams>) => {
      // 标准化查询参数中的日期字段
      const normalizedParams = TemporalConverter.normalizeTemporalQueryParams(params);
      set((state) => ({
        queryParams: { ...state.queryParams, ...normalizedParams },
        error: null
      }));
    },

    // 设置时间线视图配置
    setViewConfig: (config: Partial<TemporalTimelineViewConfig>) => {
      set((state) => ({
        viewConfig: { ...state.viewConfig, ...config },
        error: null
      }));
    },

    // 缓存组织数据
    cacheOrganizations: (key: string, data: TemporalOrganizationUnit[]) => {
      set((state) => {
        const newCache = { ...state.cache };
        newCache.organizations.set(key, data);
        newCache.lastUpdated.set(key, Date.now());
        return { cache: newCache };
      });
    },

    // 缓存时间线数据
    cacheTimeline: (key: string, data: TimelineEvent[]) => {
      set((state) => {
        const newCache = { ...state.cache };
        newCache.timelines.set(key, data);
        newCache.lastUpdated.set(key, Date.now());
        return { cache: newCache };
      });
    },

    // 获取缓存的组织数据
    getCachedOrganizations: (key: string) => {
      const { cache } = get();
      const cached = cache.organizations.get(key);
      const lastUpdated = cache.lastUpdated.get(key);
      
      // 缓存有效期5分钟
      if (cached && lastUpdated && (Date.now() - lastUpdated < 5 * 60 * 1000)) {
        return cached;
      }
      
      return undefined;
    },

    // 获取缓存的时间线数据
    getCachedTimeline: (key: string) => {
      const { cache } = get();
      const cached = cache.timelines.get(key);
      const lastUpdated = cache.lastUpdated.get(key);
      
      // 缓存有效期5分钟
      if (cached && lastUpdated && (Date.now() - lastUpdated < 5 * 60 * 1000)) {
        return cached;
      }
      
      return undefined;
    },

    // 清除缓存
    clearCache: (key?: string) => {
      set((state) => {
        if (key) {
          const newCache = { ...state.cache };
          newCache.organizations.delete(key);
          newCache.timelines.delete(key);
          newCache.lastUpdated.delete(key);
          return { cache: newCache };
        } else {
          return {
            cache: {
              organizations: new Map(),
              timelines: new Map(),
              lastUpdated: new Map()
            }
          };
        }
      });
    },

    // 设置加载状态
    setLoading: (type: keyof TemporalState['loading'], loading: boolean) => {
      set((state) => ({
        loading: { ...state.loading, [type]: loading }
      }));
    },

    // 设置错误状态
    setError: (error: string | null) => {
      set({ error });
    },

    // 重置状态
    reset: () => {
      set(defaultState);
    }
  }))
);

// 选择器函数 - 优化性能
export const temporalSelectors = {
  // 获取当前时态上下文
  useContext: () => useTemporalStore((state) => state.context),
  
  // 获取查询参数
  useQueryParams: () => useTemporalStore((state) => state.queryParams),
  
  // 获取视图配置
  useViewConfig: () => useTemporalStore((state) => state.viewConfig),
  
  // 获取加载状态
  useLoading: () => useTemporalStore((state) => state.loading),
  
  // 获取错误状态
  useError: () => useTemporalStore((state) => state.error),
  
  // 获取是否为历史模式
  useIsHistoricalMode: () => useTemporalStore((state) => 
    state.context.mode === 'historical'
  ),
  
  // 获取是否为规划模式
  useIsPlanningMode: () => useTemporalStore((state) => 
    state.context.mode === 'planning'
  ),
  
  // 获取缓存状态
  useCacheStats: () => useTemporalStore((state) => ({
    organizationsCount: state.cache.organizations.size,
    timelinesCount: state.cache.timelines.size,
    totalCacheSize: state.cache.organizations.size + state.cache.timelines.size
  }))
};

// 创建稳定的actions选择器，避免无限循环
const actionsSelector = (state: TemporalState & TemporalActions) => ({
  setMode: state.setMode,
  setAsOfDate: state.setAsOfDate,
  setDateRange: state.setDateRange,
  setQueryParams: state.setQueryParams,
  setViewConfig: state.setViewConfig,
  cacheOrganizations: state.cacheOrganizations,
  cacheTimeline: state.cacheTimeline,
  getCachedOrganizations: state.getCachedOrganizations,
  getCachedTimeline: state.getCachedTimeline,
  clearCache: state.clearCache,
  setLoading: state.setLoading,
  setError: state.setError,
  reset: state.reset
});

// 时态操作钩子 - 修复无限循环问题，使用浅比较优化
export const useTemporalActions = () => {
  return useTemporalStore(
    actionsSelector
  );
};

// 导出类型
export type { TemporalState as TemporalStoreState, TemporalActions as TemporalStoreActions };