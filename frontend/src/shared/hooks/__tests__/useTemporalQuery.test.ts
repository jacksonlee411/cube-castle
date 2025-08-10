/**
 * 时态查询钩子单元测试
 */
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode } from 'react';
import {
  useTemporalOrganizations,
  useTemporalOrganization,
  useOrganizationHistory,
  useOrganizationTimeline,
  useTemporalMode
} from '../useTemporalQuery';
import organizationAPI from '../../api/organizations-simplified';
import { useTemporalStore } from '../../stores/temporalStore';
import type { OrganizationUnit } from '../../types/organization';
import type { TemporalOrganizationUnit, TimelineEvent } from '../../types/temporal';

// 模拟API
jest.mock('../../api/organizations-simplified');
const mockOrganizationAPI = organizationAPI as jest.Mocked<typeof organizationAPI>;

// 模拟数据
const mockOrganization: OrganizationUnit = {
  code: '1000001',
  name: 'Test Department',
  unit_type: 'DEPARTMENT',
  status: 'ACTIVE',
  level: 1,
  path: '/1000001',
  sort_order: 1,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-06-01T00:00:00Z'
};

const mockTemporalOrganization: TemporalOrganizationUnit = {
  ...mockOrganization,
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

// 创建测试包装器
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useTemporalQuery hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // 重置存储状态
    const { reset } = useTemporalStore.getState();
    reset();
  });

  describe('useTemporalOrganizations', () => {
    it('should fetch organizations successfully', async () => {
      const mockResponse = {
        organizations: [mockOrganization],
        total_count: 1,
        page: 1,
        page_size: 1,
        total_pages: 1
      };

      mockOrganizationAPI.getAll.mockResolvedValue(mockResponse);

      const { result } = renderHook(
        () => useTemporalOrganizations({}),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual([mockOrganization]);
      expect(result.current.temporalContext.mode).toBe('current');
      expect(result.current.isHistorical).toBe(false);
      expect(mockOrganizationAPI.getAll).toHaveBeenCalledWith(
        expect.objectContaining({
          temporalParams: expect.any(Object)
        })
      );
    });

    it('should handle API errors', async () => {
      const errorMessage = 'Failed to fetch organizations';
      mockOrganizationAPI.getAll.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(
        () => useTemporalOrganizations({}),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toBeTruthy();
    });

    it('should use cache when available', async () => {
      // 首次调用API
      const mockResponse = {
        organizations: [mockOrganization],
        total_count: 1,
        page: 1,
        page_size: 1,
        total_pages: 1
      };

      mockOrganizationAPI.getAll.mockResolvedValue(mockResponse);

      const { result, rerender } = renderHook(
        () => useTemporalOrganizations({}),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // 确保API被调用
      expect(mockOrganizationAPI.getAll).toHaveBeenCalledTimes(1);

      // 重新渲染，应该使用缓存
      rerender();

      // API不应该被再次调用（使用了缓存或React Query缓存）
      expect(result.current.data).toEqual([mockOrganization]);
    });
  });

  describe('useTemporalOrganization', () => {
    it('should fetch single organization successfully', async () => {
      mockOrganizationAPI.getByCode.mockResolvedValue(mockOrganization);

      const { result } = renderHook(
        () => useTemporalOrganization('1000001'),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockOrganization);
      expect(mockOrganizationAPI.getByCode).toHaveBeenCalledWith(
        '1000001',
        expect.any(Object)
      );
    });

    it('should not fetch when disabled', () => {
      const { result } = renderHook(
        () => useTemporalOrganization('1000001', false),
        { wrapper: createWrapper() }
      );

      expect(result.current.isFetching).toBe(false);
      expect(mockOrganizationAPI.getByCode).not.toHaveBeenCalled();
    });

    it('should not fetch when code is empty', () => {
      const { result } = renderHook(
        () => useTemporalOrganization(''),
        { wrapper: createWrapper() }
      );

      expect(result.current.isFetching).toBe(false);
      expect(mockOrganizationAPI.getByCode).not.toHaveBeenCalled();
    });
  });

  describe('useOrganizationHistory', () => {
    it('should fetch organization history successfully', async () => {
      const mockHistory = [mockTemporalOrganization, { ...mockTemporalOrganization, version: 2 }];
      mockOrganizationAPI.getHistory.mockResolvedValue(mockHistory);

      const { result } = renderHook(
        () => useOrganizationHistory('1000001'),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockHistory);
      expect(result.current.hasHistory).toBe(true);
      expect(result.current.latestVersion).toEqual(mockHistory[0]);
      expect(mockOrganizationAPI.getHistory).toHaveBeenCalledWith(
        '1000001',
        expect.any(Object)
      );
    });

    it('should indicate no history when only one version exists', async () => {
      const mockHistory = [mockTemporalOrganization];
      mockOrganizationAPI.getHistory.mockResolvedValue(mockHistory);

      const { result } = renderHook(
        () => useOrganizationHistory('1000001'),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.hasHistory).toBe(false);
    });
  });

  describe('useOrganizationTimeline', () => {
    it('should fetch organization timeline successfully', async () => {
      const mockTimeline = [mockTimelineEvent, { ...mockTimelineEvent, id: 'evt-002' }];
      mockOrganizationAPI.getTimeline.mockResolvedValue(mockTimeline);

      const { result } = renderHook(
        () => useOrganizationTimeline('1000001'),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockTimeline);
      expect(result.current.hasEvents).toBe(true);
      expect(result.current.eventCount).toBe(2);
      expect(result.current.latestEvent).toEqual(mockTimeline[0]);
      expect(mockOrganizationAPI.getTimeline).toHaveBeenCalledWith(
        '1000001',
        expect.any(Object)
      );
    });

    it('should indicate no events when timeline is empty', async () => {
      mockOrganizationAPI.getTimeline.mockResolvedValue([]);

      const { result } = renderHook(
        () => useOrganizationTimeline('1000001'),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.hasEvents).toBe(false);
      expect(result.current.eventCount).toBe(0);
      expect(result.current.latestEvent).toBeUndefined();
    });
  });

  describe('useTemporalMode', () => {
    it('should provide current mode and switch functions', () => {
      const { result } = renderHook(() => useTemporalMode());

      expect(result.current.mode).toBe('current');
      expect(result.current.isCurrent).toBe(true);
      expect(result.current.isHistorical).toBe(false);
      expect(result.current.isPlanning).toBe(false);

      expect(result.current.switchToCurrent).toBeDefined();
      expect(result.current.switchToHistorical).toBeDefined();
      expect(result.current.switchToPlanning).toBeDefined();
    });
  });
});