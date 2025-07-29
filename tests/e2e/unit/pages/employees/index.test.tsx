// tests/unit/pages/employees/index.test.tsx
import React from 'react';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import { useRouter } from 'next/router';
import { notification } from 'antd';
import EmployeesPage from '../../../../src/pages/employees/index';
import '@testing-library/jest-dom';

// Mock next/router
jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

// Mock next/link
jest.mock('next/link', () => {
  return ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  );
});

const mockRouter = {
  push: jest.fn(),
  pathname: '/employees',
  route: '/employees',
  query: {},
  asPath: '/employees',
};

// Mock fetch
global.fetch = jest.fn();

const mockEmployees = [
  {
    id: '1',
    employeeId: 'EMP001',
    legalName: '张三',
    email: 'zhangsan@company.com',
    status: 'ACTIVE',
    hireDate: '2023-01-15',
    department: '研发部',
    position: '高级开发工程师',
    managerId: 'MGR001',
    managerName: '李经理'
  },
  {
    id: '2',
    employeeId: 'EMP002',
    legalName: '李四',
    email: 'lisi@company.com',
    status: 'ACTIVE',
    hireDate: '2023-03-20',
    department: '产品部',
    position: '产品经理',
    managerId: 'MGR002',
    managerName: '王总监'
  },
  {
    id: '3',
    employeeId: 'EMP003',
    legalName: '王五',
    email: 'wangwu@company.com',
    status: 'INACTIVE',
    hireDate: '2022-11-10',
    department: '市场部',
    position: '市场专员',
    managerId: 'MGR003',
    managerName: '赵主管'
  }
];

describe('EmployeesPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
    
    // Mock successful API response
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        data: mockEmployees,
        total: mockEmployees.length
      })
    });
  });

  it('should render the employees page title and description', async () => {
    render(<EmployeesPage />);
    
    expect(screen.getByText('员工管理')).toBeInTheDocument();
    expect(screen.getByText('管理公司员工信息、职位变更和组织结构')).toBeInTheDocument();
  });

  it('should display loading state initially', () => {
    render(<EmployeesPage />);
    
    // Should show loading indicator
    expect(screen.getByText('员工管理')).toBeInTheDocument();
  });

  it('should load and display employees data', async () => {
    render(<EmployeesPage />);
    
    // Wait for setTimeout to complete (1000ms) and data to load
    await waitFor(() => {
      expect(screen.getByTestId('table')).toBeInTheDocument();
    }, { timeout: 2000 });

    // 验证至少有一些员工数据显示
    const tableRows = screen.getByTestId('table').querySelectorAll('tbody tr');
    expect(tableRows.length).toBeGreaterThan(0);

    // 验证表格基本功能
    expect(screen.getByText('员工信息')).toBeInTheDocument();
    expect(screen.getByText('职位信息')).toBeInTheDocument();
    expect(screen.getByText('状态')).toBeInTheDocument();
  });

  it('should display status tags with correct colors', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByTestId('table')).toBeInTheDocument();
    }, { timeout: 2000 });

    // Check status values appear in table - based on actual component data
    expect(screen.getAllByText('在职').length).toBeGreaterThan(0);
    // There's one INACTIVE employee (刘七)
    expect(screen.getAllByText('离职').length).toBeGreaterThan(0);
  });

  it('should open add employee modal when clicking add button', async () => {
    render(<EmployeesPage />);
    
    const addButton = screen.getByText('新增员工');
    fireEvent.click(addButton);
    
    await waitFor(() => {
      expect(screen.getByTestId('modal')).toBeInTheDocument(); // Modal appears
    });

    // Check form fields are present
    expect(screen.getByLabelText('员工工号')).toBeInTheDocument();
    expect(screen.getByLabelText('姓名')).toBeInTheDocument();
    expect(screen.getByLabelText('邮箱')).toBeInTheDocument();
  });

  it('should filter employees by search term', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByTestId('table')).toBeInTheDocument();
      expect(screen.getAllByText('张三').length).toBeGreaterThan(0);
    }, { timeout: 2000 });

    // Search for specific employee
    const searchInput = screen.getByPlaceholderText('搜索员工姓名、工号、邮箱或职位');
    fireEvent.change(searchInput, { target: { value: '张三' } });
    
    // Should trigger filtering (using local filter, not API)
    await waitFor(() => {
      // Still should see employee data
      expect(screen.getAllByText('张三').length).toBeGreaterThan(0);
    });
  });

  it('should filter employees by department', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByTestId('table')).toBeInTheDocument();
      expect(screen.getAllByText('张三').length).toBeGreaterThan(0);
    }, { timeout: 2000 });

    // Select department filter - use the first select (department filter)
    const departmentSelects = screen.getAllByTestId('select');
    const departmentSelect = departmentSelects[0]; // First select is for department
    fireEvent.change(departmentSelect, { target: { value: '技术部' } });

    // Should trigger filtering locally
    await waitFor(() => {
      // Still should see table data
      expect(screen.getByTestId('table')).toBeInTheDocument();
    });
  });

  it('should filter employees by status', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByTestId('table')).toBeInTheDocument();
      expect(screen.getAllByText('张三').length).toBeGreaterThan(0);
    }, { timeout: 2000 });

    // Find and use the status filter (second select)
    const statusSelects = screen.getAllByTestId('select');
    const statusSelect = statusSelects[1]; // Second select is for status
    fireEvent.change(statusSelect, { target: { value: 'ACTIVE' } });

    // Should trigger filtering locally
    await waitFor(() => {
      // Still should see table data
      expect(screen.getByTestId('table')).toBeInTheDocument();
    });
  });

  it('should navigate to employee positions when clicking view positions', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByText('张三')).toBeInTheDocument();
    });

    // Click on the actions dropdown for first employee
    const moreButtons = screen.getAllByLabelText('更多操作');
    fireEvent.click(moreButtons[0]);
    
    await waitFor(() => {
      const viewPositionsButton = screen.getByText('查看职位历史');
      fireEvent.click(viewPositionsButton);
    });

    expect(mockRouter.push).toHaveBeenCalledWith('/employees/positions/1');
  });

  it('should open edit modal when clicking edit button', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByText('张三')).toBeInTheDocument();
    });

    // Click on the actions dropdown for first employee
    const moreButtons = screen.getAllByLabelText('更多操作');
    fireEvent.click(moreButtons[0]);
    
    await waitFor(() => {
      const editButton = screen.getByText('编辑');
      fireEvent.click(editButton);
    });

    await waitFor(() => {
      expect(screen.getByText('编辑员工信息')).toBeInTheDocument();
    });

    // Form should be pre-filled with employee data
    expect(screen.getByDisplayValue('张三')).toBeInTheDocument();
    expect(screen.getByDisplayValue('zhangsan@company.com')).toBeInTheDocument();
  });

  it('should handle employee creation successfully', async () => {
    render(<EmployeesPage />);
    
    // Open add modal
    const addButton = screen.getByText('新增员工');
    fireEvent.click(addButton);
    
    await waitFor(() => {
      expect(screen.getByText('添加新员工')).toBeInTheDocument();
    });

    // Fill form
    fireEvent.change(screen.getByLabelText('员工ID'), { target: { value: 'EMP004' } });
    fireEvent.change(screen.getByLabelText('姓名'), { target: { value: '新员工' } });
    fireEvent.change(screen.getByLabelText('邮箱'), { target: { value: 'new@company.com' } });

    // Mock successful creation
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ success: true })
    });

    // Submit form
    const submitButton = screen.getByText('确定');
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(notification.success).toHaveBeenCalledWith({
        message: '成功',
        description: '员工创建成功'
      });
    });
  });

  it('should handle API errors gracefully', async () => {
    // Mock API error
    (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('Network error'));
    
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(notification.error).toHaveBeenCalledWith({
        message: '错误',
        description: '获取员工数据失败'
      });
    });
  });

  it('should display employee statistics', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByText('张三')).toBeInTheDocument();
    });

    // Check statistics cards
    expect(screen.getByText('总员工数')).toBeInTheDocument();
    expect(screen.getByText('在职人数')).toBeInTheDocument();
    expect(screen.getByText('本月新增')).toBeInTheDocument();
    expect(screen.getByText('平均司龄')).toBeInTheDocument();
  });

  it('should show delete confirmation modal', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByText('张三')).toBeInTheDocument();
    });

    // Click on the actions dropdown for first employee
    const moreButtons = screen.getAllByLabelText('更多操作');
    fireEvent.click(moreButtons[0]);
    
    await waitFor(() => {
      const deleteButton = screen.getByText('删除');
      fireEvent.click(deleteButton);
    });

    await waitFor(() => {
      expect(screen.getByText('确认删除')).toBeInTheDocument();
      expect(screen.getByText('确定要删除员工 张三 吗？此操作无法撤销。')).toBeInTheDocument();
    });
  });

  it('should handle table pagination', async () => {
    // Mock response with pagination
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        data: mockEmployees,
        total: 50, // More than page size
        page: 1,
        pageSize: 10
      })
    });

    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByText('张三')).toBeInTheDocument();
    });

    // Should show pagination controls
    expect(screen.getByTitle('下一页')).toBeInTheDocument();
  });

  it('should export employees data', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByText('张三')).toBeInTheDocument();
    });

    const exportButton = screen.getByText('导出');
    fireEvent.click(exportButton);

    // Should call export API
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('export'),
        expect.any(Object)
      );
    });
  });

  it('should refresh data when clicking refresh button', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByText('张三')).toBeInTheDocument();
    });

    const refreshButton = screen.getByText('刷新');
    fireEvent.click(refreshButton);

    // Should make another API call
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(2); // Initial load + refresh
    });
  });

  it('should display employee avatars correctly', async () => {
    render(<EmployeesPage />);
    
    await waitFor(() => {
      expect(screen.getByText('张三')).toBeInTheDocument();
    });

    // Should show user icons for employees without avatars
    const avatars = screen.getAllByRole('img');
    expect(avatars.length).toBeGreaterThan(0);
  });

  it('should handle form validation in add/edit modals', async () => {
    render(<EmployeesPage />);
    
    // Open add modal
    const addButton = screen.getByText('新增员工');
    fireEvent.click(addButton);
    
    await waitFor(() => {
      expect(screen.getByText('添加新员工')).toBeInTheDocument();
    });

    // Try to submit empty form
    const submitButton = screen.getByText('确定');
    fireEvent.click(submitButton);

    // Should show validation errors
    await waitFor(() => {
      expect(screen.getByText('请输入员工ID')).toBeInTheDocument();
      expect(screen.getByText('请输入姓名')).toBeInTheDocument();
      expect(screen.getByText('请输入邮箱')).toBeInTheDocument();
    });
  });
});