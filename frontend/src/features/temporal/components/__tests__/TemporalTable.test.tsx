/**
 * 时态表格组件单元测试
 */
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TemporalTable } from '../TemporalTable';
import type { OrganizationUnit } from '../../../shared/types/organization';

// 导入要mock的模块
import * as useTemporalQuery from '../../../shared/hooks/useTemporalQuery';
import * as temporalStore from '../../../shared/stores/temporalStore';

// 模拟钩子
jest.mock('../../../shared/hooks/useTemporalQuery');
jest.mock('../../../shared/stores/temporalStore');

// 类型断言mock函数
const mockUseTemporalQuery = useTemporalQuery as jest.Mocked<typeof useTemporalQuery>;
const mockTemporalStore = temporalStore as jest.Mocked<typeof temporalStore>;

// 模拟数据
const mockOrganizations: OrganizationUnit[] = [
  {
    code: '1000001',
    name: '测试部门1',
    unit_type: 'DEPARTMENT',
    status: 'ACTIVE',
    level: 1,
    path: '/1000001',
    sort_order: 1,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-06-01T00:00:00Z'
  },
  {
    code: '1000002',
    name: '测试部门2',
    unit_type: 'DEPARTMENT',
    status: 'INACTIVE',
    level: 2,
    path: '/1000001/1000002',
    sort_order: 2,
    parent_code: '1000001',
    created_at: '2024-02-01T00:00:00Z',
    updated_at: '2024-07-01T00:00:00Z'
  }
];

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

describe('TemporalTable', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // 模拟钩子返回值
    mockUseTemporalQuery.useTemporalOrganizations.mockReturnValue({
      data: mockOrganizations,
      isLoading: false,
      isError: false,
      error: null,
      temporalContext: {
        mode: 'current',
        asOfDate: '2024-08-10T00:00:00.000Z'
      }
    });

    mockTemporalStore.temporalSelectors = {
      useContext: jest.fn().mockReturnValue({
        mode: 'current',
        asOfDate: '2024-08-10T00:00:00.000Z'
      })
    };
  });

  it('should render table with organization data', () => {
    render(<TemporalTable />, { wrapper: createWrapper() });

    expect(screen.getByText('组织架构 (2)')).toBeInTheDocument();
    expect(screen.getByText('测试部门1')).toBeInTheDocument();
    expect(screen.getByText('测试部门2')).toBeInTheDocument();
    expect(screen.getByText('1000001')).toBeInTheDocument();
    expect(screen.getByText('1000002')).toBeInTheDocument();
  });

  it('should display table headers correctly', () => {
    render(<TemporalTable />, { wrapper: createWrapper() });

    expect(screen.getByText('组织代码')).toBeInTheDocument();
    expect(screen.getByText('组织名称')).toBeInTheDocument();
    expect(screen.getByText('类型')).toBeInTheDocument();
    expect(screen.getByText('状态')).toBeInTheDocument();
    expect(screen.getByText('层级')).toBeInTheDocument();
  });

  it('should show temporal indicators when enabled', () => {
    render(<TemporalTable showTemporalIndicators={true} />, { wrapper: createWrapper() });

    // 应该有时态状态指示器列
    const indicators = screen.getAllByRole('columnheader');
    expect(indicators.some(indicator => indicator.textContent?.includes('时态状态'))).toBeTruthy();
  });

  it('should show selection column when enabled', () => {
    render(<TemporalTable showSelection={true} />, { wrapper: createWrapper() });

    // 应该有选择框列
    const checkboxes = screen.getAllByRole('checkbox');
    expect(checkboxes.length).toBeGreaterThan(0);
  });

  it('should show action buttons when enabled', () => {
    render(<TemporalTable showActions={true} />, { wrapper: createWrapper() });

    // 应该显示操作列
    expect(screen.getByText('操作')).toBeInTheDocument();
  });

  it('should call onRowClick when row is clicked', async () => {
    const mockOnRowClick = jest.fn();
    render(<TemporalTable onRowClick={mockOnRowClick} />, { wrapper: createWrapper() });

    const firstRow = screen.getByText('测试部门1').closest('tr');
    if (firstRow) {
      fireEvent.click(firstRow);
    }

    await waitFor(() => {
      expect(mockOnRowClick).toHaveBeenCalledWith(mockOrganizations[0]);
    });
  });

  it('should call onEdit when edit button is clicked', async () => {
    const mockOnEdit = jest.fn();
    render(
      <TemporalTable showActions={true} onEdit={mockOnEdit} />, 
      { wrapper: createWrapper() }
    );

    const editButtons = screen.getAllByRole('button', { name: /编辑组织/ });
    fireEvent.click(editButtons[0]);

    await waitFor(() => {
      expect(mockOnEdit).toHaveBeenCalledWith(mockOrganizations[0]);
    });
  });

  it('should call onViewHistory when history button is clicked', async () => {
    const mockOnViewHistory = jest.fn();
    render(
      <TemporalTable showActions={true} onViewHistory={mockOnViewHistory} />, 
      { wrapper: createWrapper() }
    );

    const historyButtons = screen.getAllByRole('button', { name: /查看历史版本/ });
    fireEvent.click(historyButtons[0]);

    await waitFor(() => {
      expect(mockOnViewHistory).toHaveBeenCalledWith(mockOrganizations[0]);
    });
  });

  it('should handle selection changes correctly', async () => {
    const mockOnSelectionChange = jest.fn();
    render(
      <TemporalTable showSelection={true} onSelectionChange={mockOnSelectionChange} />, 
      { wrapper: createWrapper() }
    );

    const checkboxes = screen.getAllByRole('checkbox');
    // 点击第一个数据行的复选框（跳过表头的全选复选框）
    fireEvent.click(checkboxes[1]);

    await waitFor(() => {
      expect(mockOnSelectionChange).toHaveBeenCalledWith([mockOrganizations[0]]);
    });
  });

  it('should handle select all functionality', async () => {
    const mockOnSelectionChange = jest.fn();
    render(
      <TemporalTable showSelection={true} onSelectionChange={mockOnSelectionChange} />, 
      { wrapper: createWrapper() }
    );

    const checkboxes = screen.getAllByRole('checkbox');
    // 点击表头的全选复选框
    fireEvent.click(checkboxes[0]);

    await waitFor(() => {
      expect(mockOnSelectionChange).toHaveBeenCalledWith(mockOrganizations);
    });
  });

  it('should show loading state', () => {
    mockUseTemporalQuery.useTemporalOrganizations.mockReturnValue({
      data: [],
      isLoading: true,
      isError: false,
      error: null,
      temporalContext: { mode: 'current' }
    });

    render(<TemporalTable />, { wrapper: createWrapper() });

    expect(screen.getByText('🔄 加载组织数据...')).toBeInTheDocument();
  });

  it('should show error state', () => {
    mockUseTemporalQuery.useTemporalOrganizations.mockReturnValue({
      data: [],
      isLoading: false,
      isError: true,
      error: { message: 'Test error' },
      temporalContext: { mode: 'current' }
    });

    render(<TemporalTable />, { wrapper: createWrapper() });

    expect(screen.getByText(/❌ 加载数据失败: Test error/)).toBeInTheDocument();
  });

  it('should show empty state when no data', () => {
    mockUseTemporalQuery.useTemporalOrganizations.mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
      error: null,
      temporalContext: { mode: 'current' }
    });

    render(<TemporalTable />, { wrapper: createWrapper() });

    expect(screen.getByText('📭 没有找到符合条件的组织数据')).toBeInTheDocument();
  });

  it('should display status badges correctly', () => {
    render(<TemporalTable />, { wrapper: createWrapper() });

    expect(screen.getByText('启用')).toBeInTheDocument();
    expect(screen.getByText('停用')).toBeInTheDocument();
  });

  it('should display organization types correctly', () => {
    render(<TemporalTable />, { wrapper: createWrapper() });

    const departmentElements = screen.getAllByText('部门');
    expect(departmentElements).toHaveLength(2);
  });

  it('should disable edit and delete buttons in historical mode', () => {
    mockTemporalStore.temporalSelectors = {
      useContext: jest.fn().mockReturnValue({
        mode: 'historical',
        asOfDate: '2024-06-01T00:00:00.000Z'
      })
    };

    mockUseTemporalQuery.useTemporalOrganizations.mockReturnValue({
      data: mockOrganizations,
      isLoading: false,
      isError: false,
      error: null,
      temporalContext: {
        mode: 'historical',
        asOfDate: '2024-06-01T00:00:00.000Z'
      }
    });

    render(<TemporalTable showActions={true} />, { wrapper: createWrapper() });

    const editButtons = screen.getAllByRole('button', { name: /历史模式下不可编辑/ });
    expect(editButtons[0]).toBeDisabled();
  });

  it('should show temporal fields in historical mode', () => {
    mockTemporalStore.temporalSelectors = {
      useContext: jest.fn().mockReturnValue({
        mode: 'historical',
        asOfDate: '2024-06-01T00:00:00.000Z'
      })
    };

    mockUseTemporalQuery.useTemporalOrganizations.mockReturnValue({
      data: mockOrganizations,
      isLoading: false,
      isError: false,
      error: null,
      temporalContext: {
        mode: 'historical',
        asOfDate: '2024-06-01T00:00:00.000Z'
      }
    });

    render(<TemporalTable />, { wrapper: createWrapper() });

    expect(screen.getByText('生效时间')).toBeInTheDocument();
    expect(screen.getByText('失效时间')).toBeInTheDocument();
  });

  it('should render in compact mode', () => {
    render(<TemporalTable compact={true} />, { wrapper: createWrapper() });

    // 在紧凑模式下，更新时间列应该被隐藏
    expect(screen.queryByText('更新时间')).not.toBeInTheDocument();
  });
});