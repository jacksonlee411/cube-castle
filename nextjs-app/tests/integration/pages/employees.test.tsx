// tests/integration/pages/employees.test.tsx
import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import EmployeesPage from '@/pages/employees/index';

// 模拟Next.js组件
jest.mock('next/head', () => {
  return function Head({ children }: { children: React.ReactNode }) {
    return <div data-testid="head">{children}</div>;
  };
});

// 模拟组件
jest.mock('@/components/business/employee-table', () => {
  return function EmployeeTable({ employees, loading }: any) {
    if (loading) {
      return <div data-testid="employee-table-loading">Loading employees...</div>;
    }
    return (
      <div data-testid="employee-table">
        <div data-testid="employee-count">{employees?.length || 0} employees</div>
        {employees?.map((emp: any) => (
          <div key={emp.id} data-testid={`employee-${emp.id}`}>
            {emp.legalName}
          </div>
        ))}
      </div>
    );
  };
});

jest.mock('@/components/business/employee-create-dialog', () => {
  return function EmployeeCreateDialog({ open, onOpenChange }: any) {
    if (!open) return null;
    return (
      <div data-testid="employee-create-dialog">
        <button onClick={() => onOpenChange(false)}>关闭</button>
      </div>
    );
  };
});

jest.mock('@/components/business/employee-filters', () => {
  return function EmployeeFilters({ onFiltersChange }: any) {
    return (
      <div data-testid="employee-filters">
        <input
          data-testid="search-input"
          placeholder="搜索员工"
          onChange={(e) => onFiltersChange({ search: e.target.value })}
        />
      </div>
    );
  };
});

// 模拟hook
jest.mock('@/hooks/useEmployees', () => ({
  useEmployees: jest.fn(() => ({
    employees: [
      {
        id: 'emp-1',
        employeeId: 'EMP001',
        legalName: '张三',
        email: 'zhangsan@example.com',
        status: 'ACTIVE',
        hireDate: '2023-01-01',
        currentPosition: {
          positionTitle: '软件工程师',
          department: '技术部',
          employmentType: 'FULL_TIME'
        }
      }
    ],
    loading: false,
    error: null,
    totalCount: 1,
    hasNextPage: false,
    loadMore: jest.fn(),
    refresh: jest.fn(),
  })),
  useEmployeeFilters: jest.fn(() => ({
    filters: {},
    updateFilter: jest.fn(),
    clearFilters: jest.fn(),
    hasActiveFilters: false,
  })),
  useEmployeeStats: jest.fn(() => ({
    stats: {
      total: 1,
      active: 1,
      inactive: 0,
      pending: 0,
      byDepartment: { '技术部': 1 },
      byEmploymentType: { 'FULL_TIME': 1 },
    },
    loading: false,
  })),
}));

describe('员工管理页面集成测试', () => {
  it('正确渲染员工管理页面', async () => {
    render(<EmployeesPage />);

    // 检查页面标题
    expect(screen.getByTestId('head')).toBeInTheDocument();

    // 检查主要组件是否存在
    await waitFor(() => {
      expect(screen.getByTestId('employee-filters')).toBeInTheDocument();
      expect(screen.getByTestId('employee-table')).toBeInTheDocument();
    });

    // 检查员工数据是否显示
    expect(screen.getByTestId('employee-count')).toHaveTextContent('1 employees');
    expect(screen.getByTestId('employee-emp-1')).toHaveTextContent('张三');
  });

  it('显示员工统计信息', async () => {
    render(<EmployeesPage />);

    await waitFor(() => {
      // 统计卡片应该显示正确的数据
      expect(screen.getByText('1')).toBeInTheDocument(); // 总数
    });
  });

  it('显示加载状态', async () => {
    // 修改mock返回加载状态
    const { useEmployees } = require('@/hooks/useEmployees');
    useEmployees.mockReturnValue({
      employees: [],
      loading: true,
      error: null,
      totalCount: 0,
      hasNextPage: false,
      loadMore: jest.fn(),
      refresh: jest.fn(),
    });

    render(<EmployeesPage />);

    expect(screen.getByTestId('employee-table-loading')).toBeInTheDocument();
    expect(screen.getByText('Loading employees...')).toBeInTheDocument();
  });

  it('显示空数据状态', async () => {
    // 修改mock返回空数据
    const { useEmployees } = require('@/hooks/useEmployees');
    useEmployees.mockReturnValue({
      employees: [],
      loading: false,
      error: null,
      totalCount: 0,
      hasNextPage: false,
      loadMore: jest.fn(),
      refresh: jest.fn(),
    });

    render(<EmployeesPage />);

    await waitFor(() => {
      expect(screen.getByTestId('employee-count')).toHaveTextContent('0 employees');
    });
  });

  it('处理错误状态', async () => {
    // 修改mock返回错误状态
    const { useEmployees } = require('@/hooks/useEmployees');
    useEmployees.mockReturnValue({
      employees: [],
      loading: false,
      error: new Error('Failed to fetch employees'),
      totalCount: 0,
      hasNextPage: false,
      loadMore: jest.fn(),
      refresh: jest.fn(),
    });

    render(<EmployeesPage />);

    // 页面应该能正常渲染，即使有错误
    await waitFor(() => {
      expect(screen.getByTestId('employee-table')).toBeInTheDocument();
    });
  });
});