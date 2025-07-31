// tests/unit/components/business/employee-table.test.tsx
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { EmployeeTable } from '@/components/business/employee-table';
import { Employee, Organization, EmployeeStatus } from '@/types';

// 模拟 date-fns
jest.mock('date-fns', () => ({
  format: jest.fn(() => '2023年01月01日')
}));

jest.mock('date-fns/locale', () => ({
  zhCN: {}
}));

// 模拟 lucide-react 图标
jest.mock('lucide-react', () => ({
  MoreHorizontal: () => <span data-testid="more-horizontal">more</span>,
  Edit3: () => <span data-testid="edit3">edit</span>,
  Trash2: () => <span data-testid="trash2">trash</span>,
  Eye: () => <span data-testid="eye">eye</span>,
  Mail: () => <span data-testid="mail">mail</span>,
  Phone: () => <span data-testid="phone">phone</span>,
  Building2: () => <span data-testid="building2">building</span>,
  ChevronLeft: () => <span data-testid="chevron-left">left</span>,
  ChevronRight: () => <span data-testid="chevron-right">right</span>,
  ChevronsLeft: () => <span data-testid="chevrons-left">first</span>,
  ChevronsRight: () => <span data-testid="chevrons-right">last</span>,
}));

// 模拟 UI 组件
jest.mock('@/components/ui/button', () => ({
  Button: ({ children, onClick, disabled, variant, size, className, ...props }: any) => (
    <button 
      onClick={onClick} 
      disabled={disabled}
      data-testid="button"
      data-variant={variant}
      data-size={size}
      className={className}
      {...props}
    >
      {children}
    </button>
  )
}));

jest.mock('@/components/ui/badge', () => ({
  Badge: ({ children, variant }: any) => (
    <span data-testid="badge" data-variant={variant}>{children}</span>
  )
}));

jest.mock('@/components/ui/checkbox', () => ({
  Checkbox: ({ checked, onCheckedChange }: any) => (
    <input 
      type="checkbox" 
      checked={checked}
      onChange={(e) => onCheckedChange?.(e.target.checked)}
      data-testid="checkbox"
    />
  )
}));

jest.mock('@/components/ui/table', () => ({
  Table: ({ children }: any) => <table data-testid="table">{children}</table>,
  TableBody: ({ children }: any) => <tbody data-testid="table-body">{children}</tbody>,
  TableCell: ({ children, onClick, className }: any) => (
    <td onClick={onClick} className={className} data-testid="table-cell">{children}</td>
  ),
  TableHead: ({ children, className }: any) => (
    <th className={className} data-testid="table-head">{children}</th>
  ),
  TableHeader: ({ children }: any) => <thead data-testid="table-header">{children}</thead>,
  TableRow: ({ children, onClick, className }: any) => (
    <tr onClick={onClick} className={className} data-testid="table-row">{children}</tr>
  ),
}));

jest.mock('@/components/ui/dropdown-menu', () => ({
  DropdownMenu: ({ children }: any) => <div data-testid="dropdown-menu">{children}</div>,
  DropdownMenuContent: ({ children }: any) => <div data-testid="dropdown-content">{children}</div>,
  DropdownMenuItem: ({ children, onClick }: any) => (
    <div onClick={onClick} data-testid="dropdown-item">{children}</div>
  ),
  DropdownMenuLabel: ({ children }: any) => <div data-testid="dropdown-label">{children}</div>,
  DropdownMenuSeparator: () => <hr data-testid="dropdown-separator" />,
  DropdownMenuTrigger: ({ children }: any) => <div data-testid="dropdown-trigger">{children}</div>,
}));

// 测试数据
const mockEmployees: Employee[] = [
  {
    id: 'emp-1',
    createdAt: '2023-01-01T00:00:00Z',
    updatedAt: '2023-01-01T00:00:00Z',
    email: 'zhangsan@example.com',
    status: EmployeeStatus.ACTIVE,
    firstName: '三',
    lastName: '张',
    fullName: '张三',
    employeeNumber: 'EMP001',
    phoneNumber: '13800138000',
    organizationId: 'org-1',
    hireDate: '2023-01-01',
    tenantId: 'tenant-1',
  },
  {
    id: 'emp-2',
    createdAt: '2023-02-01T00:00:00Z',
    updatedAt: '2023-02-01T00:00:00Z',
    email: 'lisi@example.com',
    status: EmployeeStatus.INACTIVE,
    firstName: '四',
    lastName: '李',
    fullName: '李四',
    employeeNumber: 'EMP002',
    organizationId: 'org-2',
    hireDate: '2023-02-01',
    tenantId: 'tenant-1',
  },
];

const mockOrganizations: Organization[] = [
  { id: 'org-1', name: '技术部' } as Organization,
  { id: 'org-2', name: '市场部' } as Organization,
];

const defaultProps = {
  employees: mockEmployees,
  loading: false,
  error: null,
  organizations: mockOrganizations,
  pagination: {
    current: 1,
    pageSize: 10,
    total: 2,
  },
  selectedEmployees: [],
  onSelectionChange: jest.fn(),
  onPageChange: jest.fn(),
  onEmployeeClick: jest.fn(),
  onEmployeeEdit: jest.fn(),
  onEmployeeDelete: jest.fn(),
};

describe('EmployeeTable', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('正确渲染员工表格', () => {
    render(<EmployeeTable {...defaultProps} />);

    // 检查表格结构
    expect(screen.getByTestId('table')).toBeInTheDocument();
    expect(screen.getByTestId('table-header')).toBeInTheDocument();
    expect(screen.getByTestId('table-body')).toBeInTheDocument();

    // 检查员工数据 - 使用更灵活的查找方式
    expect(screen.getByText((content, element) => {
      return element?.textContent === '张三';
    })).toBeInTheDocument();
    
    expect(screen.getByText('#EMP001')).toBeInTheDocument();
    expect(screen.getByText('zhangsan@example.com')).toBeInTheDocument();
    expect(screen.getByText('技术部')).toBeInTheDocument();
  });

  it('显示加载状态', () => {
    render(<EmployeeTable {...defaultProps} loading={true} />);

    // 检查加载骨架屏
    expect(screen.getAllByText('').length).toBeGreaterThan(0);
    expect(document.querySelector('.animate-pulse')).toBeInTheDocument();
  });

  it('显示错误状态', () => {
    render(<EmployeeTable {...defaultProps} error="获取员工数据失败" />);

    expect(screen.getByText('加载失败')).toBeInTheDocument();
    expect(screen.getByText('获取员工数据失败')).toBeInTheDocument();
  });

  it('显示空数据状态', () => {
    render(<EmployeeTable {...defaultProps} employees={[]} />);

    expect(screen.getByText('暂无员工数据')).toBeInTheDocument();
    expect(screen.getByText('点击"新增员工"按钮开始添加员工信息')).toBeInTheDocument();
  });

  it('点击员工行时调用选择回调', () => {
    const onEmployeeClick = jest.fn();
    render(<EmployeeTable {...defaultProps} onEmployeeClick={onEmployeeClick} />);

    // 点击第一行
    const rows = screen.getAllByTestId('table-row');
    fireEvent.click(rows[1]); // 跳过表头行

    expect(onEmployeeClick).toHaveBeenCalledWith(mockEmployees[0]);
  });

  it('显示正确的员工状态', () => {
    render(<EmployeeTable {...defaultProps} />);

    // 检查状态徽章
    const badges = screen.getAllByTestId('badge');
    expect(badges).toHaveLength(2);
    expect(screen.getByText('在职')).toBeInTheDocument();
    expect(screen.getByText('离职')).toBeInTheDocument();
  });

  it('显示员工联系信息', () => {
    render(<EmployeeTable {...defaultProps} />);

    // 检查邮箱和电话
    expect(screen.getByText('zhangsan@example.com')).toBeInTheDocument();
    expect(screen.getByText('13800138000')).toBeInTheDocument();
  });

  it('处理复选框选择', () => {
    const onSelectionChange = jest.fn();
    render(<EmployeeTable {...defaultProps} onSelectionChange={onSelectionChange} />);

    // 点击第一个员工的复选框
    const checkboxes = screen.getAllByTestId('checkbox');
    fireEvent.click(checkboxes[1]); // 跳过全选复选框

    expect(onSelectionChange).toHaveBeenCalledWith(['emp-1']);
  });

  it('处理全选复选框', () => {
    const onSelectionChange = jest.fn();
    render(<EmployeeTable {...defaultProps} onSelectionChange={onSelectionChange} />);

    // 点击全选复选框
    const checkboxes = screen.getAllByTestId('checkbox');
    fireEvent.click(checkboxes[0]);

    expect(onSelectionChange).toHaveBeenCalledWith(['emp-1', 'emp-2']);
  });

  it('显示分页信息', () => {
    const propsWithPagination = {
      ...defaultProps,
      pagination: {
        current: 1,
        pageSize: 1,
        total: 2,
      },
    };
    
    render(<EmployeeTable {...propsWithPagination} />);

    expect(screen.getByText('显示 1 - 1 条，共 2 条记录')).toBeInTheDocument();
  });
});