/**
 * 时态存储单元测试
 */
import { renderHook, act } from '@testing-library/react';
import { useTemporalStore, useTemporalActions, temporalSelectors } from '../temporalStore';
import type { TemporalMode, TemporalOrganizationUnit, TimelineEvent } from '../../types/temporal';

// 模拟数据
const mockOrganization: TemporalOrganizationUnit = {
  code: '1000001',
  name: 'Test Department',
  unit_type: 'DEPARTMENT',
  status: 'ACTIVE',
  level: 1,
  path: '/1000001',
  sort_order: 1,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-06-01T00:00:00Z',
  effective_from: '2024-01-01T00:00:00Z',
  effective_to: '2024-12-31T23:59:59Z',
  is_temporal: true,
  version: 1
};

const mockTimelineEvent: TimelineEvent = {
  id: 'evt-001',
  organizationCode: '1000001',
  eventType: 'create',
  eventDate: '2024-01-01T00:00:00Z',
  status: 'completed',
  title: 'Organization Created',
  createdAt: '2024-01-01T00:00:00Z'
};

describe('temporalStore', () => {
  beforeEach(() => {
    // 重置存储状态
    const { reset } = useTemporalStore.getState();
    reset();
  });

  describe('useTemporalStore', () => {
    it('should have default state', () => {
      const { result } = renderHook(() => useTemporalStore());
      const state = result.current;

      expect(state.context.mode).toBe('current');
      expect(state.queryParams.mode).toBe('current');
      expect(state.loading.organizations).toBe(false);
      expect(state.loading.timeline).toBe(false);
      expect(state.error).toBe(null);
    });

    it('should update mode', () => {
      const { result } = renderHook(() => useTemporalStore());

      act(() => {
        result.current.setMode('historical');
      });

      expect(result.current.context.mode).toBe('historical');
      expect(result.current.queryParams.mode).toBe('historical');
      expect(result.current.error).toBe(null);
    });

    it('should update asOfDate', () => {
      const { result } = renderHook(() => useTemporalStore());
      const testDate = '2024-06-01T00:00:00Z';

      act(() => {
        result.current.setAsOfDate(testDate);
      });

      expect(result.current.context.asOfDate).toBe(testDate);
      expect(result.current.queryParams.asOfDate).toBe(testDate);
    });

    it('should update date range', () => {
      const { result } = renderHook(() => useTemporalStore());
      const dateRange = {
        start: '2024-01-01T00:00:00Z',
        end: '2024-12-31T23:59:59Z'
      };

      act(() => {
        result.current.setDateRange(dateRange);
      });

      expect(result.current.queryParams.dateRange).toEqual(dateRange);
    });

    it('should cache organizations', () => {
      const { result } = renderHook(() => useTemporalStore());
      const cacheKey = 'test-org-key';
      const organizations = [mockOrganization];

      act(() => {
        result.current.cacheOrganizations(cacheKey, organizations);
      });

      const cached = result.current.getCachedOrganizations(cacheKey);
      expect(cached).toEqual(organizations);
    });

    it('should cache timeline', () => {
      const { result } = renderHook(() => useTemporalStore());
      const cacheKey = 'test-timeline-key';
      const timeline = [mockTimelineEvent];

      act(() => {
        result.current.cacheTimeline(cacheKey, timeline);
      });

      const cached = result.current.getCachedTimeline(cacheKey);
      expect(cached).toEqual(timeline);
    });

    it('should return undefined for expired cache', async () => {
      const { result } = renderHook(() => useTemporalStore());
      const cacheKey = 'test-expired-key';
      const organizations = [mockOrganization];

      // 缓存数据
      act(() => {
        result.current.cacheOrganizations(cacheKey, organizations);
      });

      // 手动设置过期时间（实际测试中可能需要模拟时间）
      const state = useTemporalStore.getState();
      state.cache.lastUpdated.set(cacheKey, Date.now() - 6 * 60 * 1000); // 6分钟前

      const cached = result.current.getCachedOrganizations(cacheKey);
      expect(cached).toBeUndefined();
    });

    it('should clear cache', () => {
      const { result } = renderHook(() => useTemporalStore());
      const cacheKey = 'test-clear-key';
      const organizations = [mockOrganization];

      // 先缓存数据
      act(() => {
        result.current.cacheOrganizations(cacheKey, organizations);
      });

      // 验证缓存存在
      expect(result.current.getCachedOrganizations(cacheKey)).toEqual(organizations);

      // 清除特定缓存
      act(() => {
        result.current.clearCache(cacheKey);
      });

      expect(result.current.getCachedOrganizations(cacheKey)).toBeUndefined();
    });

    it('should clear all cache', () => {
      const { result } = renderHook(() => useTemporalStore());
      const orgKey = 'test-org-key';
      const timelineKey = 'test-timeline-key';

      // 缓存多个数据
      act(() => {
        result.current.cacheOrganizations(orgKey, [mockOrganization]);
        result.current.cacheTimeline(timelineKey, [mockTimelineEvent]);
      });

      // 清除所有缓存
      act(() => {
        result.current.clearCache();
      });

      expect(result.current.getCachedOrganizations(orgKey)).toBeUndefined();
      expect(result.current.getCachedTimeline(timelineKey)).toBeUndefined();
    });

    it('should manage loading states', () => {
      const { result } = renderHook(() => useTemporalStore());

      act(() => {
        result.current.setLoading('organizations', true);
      });

      expect(result.current.loading.organizations).toBe(true);
      expect(result.current.loading.timeline).toBe(false);

      act(() => {
        result.current.setLoading('organizations', false);
        result.current.setLoading('timeline', true);
      });

      expect(result.current.loading.organizations).toBe(false);
      expect(result.current.loading.timeline).toBe(true);
    });

    it('should manage error state', () => {
      const { result } = renderHook(() => useTemporalStore());
      const errorMessage = 'Test error message';

      act(() => {
        result.current.setError(errorMessage);
      });

      expect(result.current.error).toBe(errorMessage);

      act(() => {
        result.current.setError(null);
      });

      expect(result.current.error).toBe(null);
    });

    it('should reset to default state', () => {
      const { result } = renderHook(() => useTemporalStore());

      // 修改一些状态
      act(() => {
        result.current.setMode('historical');
        result.current.setError('Test error');
        result.current.setLoading('organizations', true);
      });

      // 验证状态已改变
      expect(result.current.context.mode).toBe('historical');
      expect(result.current.error).toBe('Test error');
      expect(result.current.loading.organizations).toBe(true);

      // 重置状态
      act(() => {
        result.current.reset();
      });

      // 验证状态已重置
      expect(result.current.context.mode).toBe('current');
      expect(result.current.error).toBe(null);
      expect(result.current.loading.organizations).toBe(false);
    });
  });

  describe('useTemporalActions', () => {
    it('should return all actions', () => {
      const { result } = renderHook(() => useTemporalActions());
      const actions = result.current;

      expect(actions.setMode).toBeDefined();
      expect(actions.setAsOfDate).toBeDefined();
      expect(actions.setDateRange).toBeDefined();
      expect(actions.cacheOrganizations).toBeDefined();
      expect(actions.clearCache).toBeDefined();
      expect(actions.reset).toBeDefined();
    });
  });

  describe('temporalSelectors', () => {
    it('should select context', () => {
      const { result } = renderHook(() => temporalSelectors.useContext());
      expect(result.current.mode).toBe('current');
    });

    it('should select query params', () => {
      const { result } = renderHook(() => temporalSelectors.useQueryParams());
      expect(result.current.mode).toBe('current');
    });

    it('should select loading state', () => {
      const { result } = renderHook(() => temporalSelectors.useLoading());
      expect(result.current.organizations).toBe(false);
      expect(result.current.timeline).toBe(false);
    });

    it('should select error state', () => {
      const { result } = renderHook(() => temporalSelectors.useError());
      expect(result.current).toBe(null);
    });

    it('should detect historical mode', () => {
      const { result: storeResult } = renderHook(() => useTemporalStore());
      const { result: selectorResult } = renderHook(() => temporalSelectors.useIsHistoricalMode());

      expect(selectorResult.current).toBe(false);

      act(() => {
        storeResult.current.setMode('historical');
      });

      expect(selectorResult.current).toBe(true);
    });

    it('should detect planning mode', () => {
      const { result: storeResult } = renderHook(() => useTemporalStore());
      const { result: selectorResult } = renderHook(() => temporalSelectors.useIsPlanningMode());

      expect(selectorResult.current).toBe(false);

      act(() => {
        storeResult.current.setMode('planning');
      });

      expect(selectorResult.current).toBe(true);
    });

    it('should get cache stats', () => {
      const { result: storeResult } = renderHook(() => useTemporalStore());
      const { result: statsResult } = renderHook(() => temporalSelectors.useCacheStats());

      expect(statsResult.current.organizationsCount).toBe(0);
      expect(statsResult.current.timelinesCount).toBe(0);
      expect(statsResult.current.totalCacheSize).toBe(0);

      act(() => {
        storeResult.current.cacheOrganizations('test-key', [mockOrganization]);
        storeResult.current.cacheTimeline('test-key', [mockTimelineEvent]);
      });

      expect(statsResult.current.organizationsCount).toBe(1);
      expect(statsResult.current.timelinesCount).toBe(1);
      expect(statsResult.current.totalCacheSize).toBe(2);
    });
  });
});