/**
 * 时态导航栏组件单元测试
 */
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TemporalNavbar } from '../TemporalNavbar';

// 导入要mock的模块
import * as useTemporalQuery from '../../../shared/hooks/useTemporalQuery';
import * as temporalStore from '../../../shared/stores/temporalStore';

// 模拟钩子
jest.mock('../../../shared/hooks/useTemporalQuery');
jest.mock('../../../shared/stores/temporalStore');

// 类型断言mock函数
const mockUseTemporalQuery = useTemporalQuery as jest.Mocked<typeof useTemporalQuery>;
const mockTemporalStore = temporalStore as jest.Mocked<typeof temporalStore>;

// 创建测试包装器
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

// 模拟默认状态
const mockDefaultState = {
  mode: 'current' as const,
  switchToCurrent: jest.fn(),
  switchToHistorical: jest.fn(),
  switchToPlanning: jest.fn(),
  isCurrent: true,
  isHistorical: false,
  isPlanning: false,
  loading: {
    organizations: false,
    timeline: false,
    history: false
  },
  error: null,
  context: {
    mode: 'current' as const,
    asOfDate: '2024-08-10T00:00:00.000Z',
    effectiveDate: '2024-08-10T00:00:00.000Z',
    timezone: 'UTC',
    version: 1
  },
  cacheStats: {
    organizationsCount: 0,
    timelinesCount: 0,
    totalCacheSize: 0
  },
  refreshCache: jest.fn(),
  setError: jest.fn()
};

describe('TemporalNavbar', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // 模拟钩子返回值
    mockUseTemporalQuery.useTemporalMode.mockReturnValue(mockDefaultState);
    mockUseTemporalQuery.useTemporalQueryState.mockReturnValue({
      loading: mockDefaultState.loading,
      error: mockDefaultState.error,
      context: mockDefaultState.context,
      cacheStats: mockDefaultState.cacheStats,
      refreshCache: mockDefaultState.refreshCache
    });
    mockTemporalStore.useTemporalActions.mockReturnValue({
      setError: mockDefaultState.setError
    });
    mockTemporalStore.temporalSelectors.useQueryParams.mockReturnValue({
      mode: 'current',
      asOfDate: '2024-08-10T00:00:00.000Z'
    });
  });

  it('should render temporal navbar with mode buttons', () => {
    render(<TemporalNavbar />, { wrapper: createWrapper() });

    expect(screen.getByText('当前')).toBeInTheDocument();
    expect(screen.getByText('历史')).toBeInTheDocument();
    expect(screen.getByText('规划')).toBeInTheDocument();
  });

  it('should show current mode as active', () => {
    render(<TemporalNavbar />, { wrapper: createWrapper() });

    const currentButton = screen.getByRole('button', { name: /当前/ });
    expect(currentButton).toHaveAttribute('aria-pressed', 'true');
  });

  it('should call switchToCurrent when current button is clicked', async () => {
    const mockSwitchToCurrent = jest.fn();
    mockUseTemporalQuery.useTemporalMode.mockReturnValue({
      ...mockDefaultState,
      switchToCurrent: mockSwitchToCurrent
    });

    render(<TemporalNavbar />, { wrapper: createWrapper() });

    fireEvent.click(screen.getByText('当前'));

    await waitFor(() => {
      expect(mockSwitchToCurrent).toHaveBeenCalled();
    });
  });

  it('should show date picker when historical button is clicked', async () => {
    render(<TemporalNavbar />, { wrapper: createWrapper() });

    fireEvent.click(screen.getByText('历史'));

    // 应该显示日期选择器弹窗
    await waitFor(() => {
      expect(screen.getByText('选择历史查看时点')).toBeInTheDocument();
    });
  });

  it('should call switchToPlanning when planning button is clicked', async () => {
    const mockSwitchToPlanning = jest.fn();
    mockUseTemporalQuery.useTemporalMode.mockReturnValue({
      ...mockDefaultState,
      switchToPlanning: mockSwitchToPlanning
    });

    render(<TemporalNavbar />, { wrapper: createWrapper() });

    fireEvent.click(screen.getByText('规划'));

    await waitFor(() => {
      expect(mockSwitchToPlanning).toHaveBeenCalled();
    });
  });

  it('should display current mode badge and description', () => {
    render(<TemporalNavbar />, { wrapper: createWrapper() });

    expect(screen.getByText('当前视图')).toBeInTheDocument();
    expect(screen.getByText('显示当前有效的组织架构')).toBeInTheDocument();
  });

  it('should display historical mode when in historical mode', () => {
    mockUseTemporalQuery.useTemporalMode.mockReturnValue({
      ...mockDefaultState,
      mode: 'historical',
      isCurrent: false,
      isHistorical: true,
      context: {
        ...mockDefaultState.context,
        mode: 'historical',
        asOfDate: '2024-06-01T00:00:00.000Z'
      }
    });
    mockUseTemporalQuery.useTemporalQueryState.mockReturnValue({
      ...mockDefaultState,
      context: {
        ...mockDefaultState.context,
        mode: 'historical',
        asOfDate: '2024-06-01T00:00:00.000Z'
      }
    });

    render(<TemporalNavbar />, { wrapper: createWrapper() });

    expect(screen.getByText('历史视图')).toBeInTheDocument();
  });

  it('should show loading indicator when loading', () => {
    mockUseTemporalQuery.useTemporalQueryState.mockReturnValue({
      ...mockDefaultState,
      loading: {
        organizations: true,
        timeline: false,
        history: false
      }
    });

    render(<TemporalNavbar />, { wrapper: createWrapper() });

    expect(screen.getByText('🔄 加载组织数据...')).toBeInTheDocument();
  });

  it('should show error message when there is an error', () => {
    mockUseTemporalQuery.useTemporalQueryState.mockReturnValue({
      ...mockDefaultState,
      error: 'Test error message'
    });

    render(<TemporalNavbar />, { wrapper: createWrapper() });

    expect(screen.getByText('⚠️ Test error message')).toBeInTheDocument();
  });

  it('should show cache stats when cache has data', () => {
    mockUseTemporalQuery.useTemporalQueryState.mockReturnValue({
      ...mockDefaultState,
      cacheStats: {
        organizationsCount: 5,
        timelinesCount: 3,
        totalCacheSize: 8
      }
    });

    render(<TemporalNavbar />, { wrapper: createWrapper() });

    expect(screen.getByText('8')).toBeInTheDocument();
  });

  it('should call refreshCache when refresh button is clicked', async () => {
    const mockRefreshCache = jest.fn();
    mockUseTemporalQuery.useTemporalQueryState.mockReturnValue({
      ...mockDefaultState,
      refreshCache: mockRefreshCache
    });

    render(<TemporalNavbar />, { wrapper: createWrapper() });

    const refreshButton = screen.getByRole('button', { name: /刷新数据缓存/ });
    fireEvent.click(refreshButton);

    await waitFor(() => {
      expect(mockRefreshCache).toHaveBeenCalled();
    });
  });

  it('should disable buttons when loading', () => {
    mockUseTemporalQuery.useTemporalQueryState.mockReturnValue({
      ...mockDefaultState,
      loading: {
        organizations: true,
        timeline: false,
        history: false
      }
    });

    render(<TemporalNavbar />, { wrapper: createWrapper() });

    const currentButton = screen.getByText('当前');
    const historicalButton = screen.getByText('历史');
    const planningButton = screen.getByText('规划');

    expect(currentButton).toBeDisabled();
    expect(historicalButton).toBeDisabled();
    expect(planningButton).toBeDisabled();
  });

  it('should render in compact mode', () => {
    render(<TemporalNavbar compact={true} />, { wrapper: createWrapper() });

    // 在紧凑模式下，不应该显示详细描述
    expect(screen.queryByText('显示当前有效的组织架构')).not.toBeInTheDocument();
  });

  it('should hide advanced settings when showAdvancedSettings is false', () => {
    render(<TemporalNavbar showAdvancedSettings={false} />, { wrapper: createWrapper() });

    expect(screen.queryByRole('button', { name: /时态查询设置/ })).not.toBeInTheDocument();
  });

  it('should call onModeChange when mode changes', async () => {
    const mockOnModeChange = jest.fn();
    
    render(<TemporalNavbar onModeChange={mockOnModeChange} />, { wrapper: createWrapper() });

    fireEvent.click(screen.getByText('规划'));

    await waitFor(() => {
      expect(mockOnModeChange).toHaveBeenCalledWith('planning');
    });
  });
});