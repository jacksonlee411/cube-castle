// tests/unit/pages/employees/positions/[id].test.tsx
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useRouter } from 'next/router';
import EmployeePositionHistoryPage from '../../../../../src/pages/employees/positions/[id]';
import '@testing-library/jest-dom';

// Mock next/router
jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

// Mock custom hooks
jest.mock('../../../../../src/hooks/useEmployees', () => ({
  useEmployee: jest.fn(),
  useCreatePositionChange: jest.fn(),
  usePositionTimeline: jest.fn(),
}));

jest.mock('../../../../../src/hooks/useWorkflows', () => ({
  useWorkflowHistory: jest.fn(),
}));

// Mock antd - 移动到文件顶部并统一处理
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  notification: {
    success: jest.fn(),
    error: jest.fn(),
  },
}));

// Get reference to mocked notification
const { notification: mockNotification } = require('antd');

const mockRouter = {
  query: { id: 'emp-001' },
  push: jest.fn(),
  back: jest.fn(),
  pathname: '/employees/positions/[id]',
  route: '/employees/positions/[id]',
  asPath: '/employees/positions/emp-001',
};

const mockEmployee = {
  id: 'emp-001',
  employeeId: 'EMP001',
  legalName: '张三',
  email: 'zhangsan@company.com',
  status: 'ACTIVE',
  hireDate: '2023-01-15',
  currentPosition: {
    positionTitle: '高级开发工程师',
    department: '研发部',
    jobLevel: 'SENIOR',
    location: '北京',
    employmentType: 'FULL_TIME'
  }
};

const mockPositionHistory = [
  {
    id: 'pos-001',
    positionTitle: '高级开发工程师',
    department: '研发部',
    jobLevel: 'SENIOR',
    location: '北京',
    employmentType: 'FULL_TIME',
    effectiveDate: '2024-01-01',
    endDate: null,
    changeReason: '晋升',
    isRetroactive: false,
    minSalary: 18000,
    maxSalary: 22000,
    currency: 'CNY'
  },
  {
    id: 'pos-002',
    positionTitle: '开发工程师',
    department: '研发部',
    jobLevel: 'INTERMEDIATE',
    location: '北京',
    employmentType: 'FULL_TIME',
    effectiveDate: '2023-01-15',
    endDate: '2023-12-31',
    changeReason: '入职',
    isRetroactive: false,
    minSalary: 15000,
    maxSalary: 18000,
    currency: 'CNY'
  }
];

const mockWorkflows = [
  {
    id: 'wf-001',
    workflowType: 'POSITION_CHANGE',
    status: 'COMPLETED',
    createdAt: '2024-01-01T09:00:00Z',
    completedAt: '2024-01-01T18:00:00Z',
    positionChange: {
      fromTitle: '开发工程师',
      toTitle: '高级开发工程师',
      effectiveDate: '2024-01-01'
    }
  }
];

// Mock hooks implementation
const mockUseEmployee = require('../../../../../src/hooks/useEmployees').useEmployee as jest.Mock;
const mockUseCreatePositionChange = require('../../../../../src/hooks/useEmployees').useCreatePositionChange as jest.Mock;
const mockUsePositionTimeline = require('../../../../../src/hooks/useEmployees').usePositionTimeline as jest.Mock;
const mockUseWorkflowHistory = require('../../../../../src/hooks/useWorkflows').useWorkflowHistory as jest.Mock;

describe('EmployeePositionHistoryPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
    
    // Mock hooks return values
    mockUseEmployee.mockReturnValue({
      employee: mockEmployee,
      loading: false,
      refetch: jest.fn()
    });
    
    mockUseCreatePositionChange.mockReturnValue({
      create: jest.fn().mockResolvedValue({ success: true }),
      loading: false
    });
    
    mockUsePositionTimeline.mockReturnValue({
      positionTimeline: mockPositionHistory,
      loading: false,
      refetch: jest.fn()
    });
    
    mockUseWorkflowHistory.mockReturnValue({
      workflows: mockWorkflows,
      loading: false
    });
  });

  it('should render employee position history page with correct title', async () => {
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('职位历史')).toBeInTheDocument();
      expect(screen.getByText('张三')).toBeInTheDocument();
      expect(screen.getByText('EMP001')).toBeInTheDocument();
    });
  });

  it('should display loading state when data is loading', () => {
    mockUseEmployee.mockReturnValue({
      employee: null,
      loading: true,
      refetch: jest.fn()
    });
    
    render(<EmployeePositionHistoryPage />);
    
    // Check for Ant Design Spin component with aria-busy attribute
    expect(document.querySelector('[aria-busy="true"]')).toBeInTheDocument();
  });

  it('should display employee current position information', async () => {
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('当前职位')).toBeInTheDocument();
      expect(screen.getByText('高级开发工程师')).toBeInTheDocument();
      expect(screen.getByText('研发部')).toBeInTheDocument();
      expect(screen.getByText('SENIOR')).toBeInTheDocument();
      expect(screen.getByText('北京')).toBeInTheDocument();
    });
  });

  it('should display position history timeline', async () => {
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('历史记录')).toBeInTheDocument();
    });

    // Check position history entries
    expect(screen.getByText('开发工程师')).toBeInTheDocument();
    expect(screen.getByText('晋升')).toBeInTheDocument();
    expect(screen.getByText('入职')).toBeInTheDocument();
  });

  it('should display salary information', async () => {
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('¥18,000 - ¥22,000')).toBeInTheDocument();
      expect(screen.getByText('¥15,000 - ¥18,000')).toBeInTheDocument();
    });
  });

  it('should open create position change modal when clicking add button', async () => {
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('添加职位变更')).toBeInTheDocument();
    });

    const addButton = screen.getByText('添加职位变更');
    fireEvent.click(addButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建职位变更')).toBeInTheDocument();
    });

    // Check form fields
    expect(screen.getByLabelText('职位名称')).toBeInTheDocument();
    expect(screen.getByLabelText('部门')).toBeInTheDocument();
    expect(screen.getByLabelText('职级')).toBeInTheDocument();
    expect(screen.getByLabelText('生效日期')).toBeInTheDocument();
    expect(screen.getByLabelText('变更原因')).toBeInTheDocument();
  });

  it('should create position change successfully', async () => {
    const mockCreateFn = jest.fn().mockResolvedValue({ success: true, workflowId: 'wf-002' });
    mockUseCreatePositionChange.mockReturnValue({
      create: mockCreateFn,
      loading: false
    });

    render(<EmployeePositionHistoryPage />);
    
    // Open modal
    const addButton = screen.getByText('添加职位变更');
    fireEvent.click(addButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建职位变更')).toBeInTheDocument();
    });

    // Fill form
    fireEvent.change(screen.getByLabelText('职位名称'), { target: { value: '技术主管' } });
    fireEvent.change(screen.getByLabelText('部门'), { target: { value: '研发部' } });
    fireEvent.change(screen.getByLabelText('职级'), { target: { value: 'MANAGER' } });
    
    // Submit form
    const submitButton = screen.getByText('提交');
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockCreateFn).toHaveBeenCalledWith(expect.objectContaining({
        employeeId: 'emp-001',
        positionData: expect.objectContaining({
          positionTitle: '技术主管',
          department: '研发部',
          jobLevel: 'MANAGER'
        })
      }));
    });

    await waitFor(() => {
      expect(mockNotification.success).toHaveBeenCalledWith({
        message: '成功',
        description: '职位变更已提交，工作流ID: wf-002'
      });
    });
  });

  it('should display workflow history tab', async () => {
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('工作流历史')).toBeInTheDocument();
    });

    // Click workflow history tab
    const workflowTab = screen.getByText('工作流历史');
    fireEvent.click(workflowTab);

    await waitFor(() => {
      expect(screen.getByText('POSITION_CHANGE')).toBeInTheDocument();
      expect(screen.getByText('COMPLETED')).toBeInTheDocument();
    });
  });

  it('should show effective dates correctly', async () => {
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('2024-01-01')).toBeInTheDocument();
      expect(screen.getByText('2023-01-15')).toBeInTheDocument();
    });
  });

  it('should handle form validation in create modal', async () => {
    render(<EmployeePositionHistoryPage />);
    
    // Open modal
    const addButton = screen.getByText('添加职位变更');
    fireEvent.click(addButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建职位变更')).toBeInTheDocument();
    });

    // Try to submit empty form
    const submitButton = screen.getByText('提交');
    fireEvent.click(submitButton);

    // Should show validation errors
    await waitFor(() => {
      expect(screen.getByText('请输入职位名称')).toBeInTheDocument();
      expect(screen.getByText('请输入部门')).toBeInTheDocument();
      expect(screen.getByText('请选择生效日期')).toBeInTheDocument();
    });
  });

  it('should close modal when clicking cancel', async () => {
    render(<EmployeePositionHistoryPage />);
    
    // Open modal
    const addButton = screen.getByText('添加职位变更');
    fireEvent.click(addButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建职位变更')).toBeInTheDocument();
    });

    // Click cancel
    const cancelButton = screen.getByText('取消');
    fireEvent.click(cancelButton);

    // Modal should close
    await waitFor(() => {
      expect(screen.queryByText('创建职位变更')).not.toBeInTheDocument();
    });
  });

  it('should navigate back when clicking back button', async () => {
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('返回员工列表')).toBeInTheDocument();
    });

    const backButton = screen.getByText('返回员工列表');
    fireEvent.click(backButton);

    expect(mockRouter.back).toHaveBeenCalled();
  });

  it('should handle API errors gracefully', async () => {
    mockUseEmployee.mockReturnValue({
      employee: null,
      loading: false,
      error: 'Failed to load employee',
      refetch: jest.fn()
    });
    
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('加载失败')).toBeInTheDocument();
    });
  });

  it('should display retroactive indicator', async () => {
    const retroactiveHistory = [
      {
        ...mockPositionHistory[0],
        isRetroactive: true
      }
    ];

    mockUsePositionTimeline.mockReturnValue({
      positionTimeline: retroactiveHistory,
      loading: false,
      refetch: jest.fn()
    });

    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('追溯')).toBeInTheDocument();
    });
  });

  it('should show employment type information', async () => {
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('全职')).toBeInTheDocument(); // FULL_TIME translated
    });
  });

  it('should handle create position change failure', async () => {
    const mockCreateFn = jest.fn().mockRejectedValue(new Error('Creation failed'));
    mockUseCreatePositionChange.mockReturnValue({
      create: mockCreateFn,
      loading: false
    });

    render(<EmployeePositionHistoryPage />);
    
    // Open modal and fill form
    const addButton = screen.getByText('添加职位变更');
    fireEvent.click(addButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建职位变更')).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText('职位名称'), { target: { value: '技术主管' } });
    fireEvent.change(screen.getByLabelText('部门'), { target: { value: '研发部' } });
    
    const submitButton = screen.getByText('提交');
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockNotification.error).toHaveBeenCalledWith({
        message: '失败',
        description: '职位变更提交失败'
      });
    });
  });

  it('should refresh data when refetch is called', async () => {
    const mockRefetch = jest.fn();
    mockUseEmployee.mockReturnValue({
      employee: mockEmployee,
      loading: false,
      refetch: mockRefetch
    });

    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('刷新')).toBeInTheDocument();
    });

    const refreshButton = screen.getByText('刷新');
    fireEvent.click(refreshButton);

    expect(mockRefetch).toHaveBeenCalled();
  });

  it('should display workflow details in history tab', async () => {
    render(<EmployeePositionHistoryPage />);
    
    // Click workflow history tab
    const workflowTab = screen.getByText('工作流历史');
    fireEvent.click(workflowTab);

    await waitFor(() => {
      expect(screen.getByText('开发工程师 → 高级开发工程师')).toBeInTheDocument();
      expect(screen.getByText('2024-01-01')).toBeInTheDocument();
    });
  });

  it('should handle empty position history', async () => {
    mockUsePositionTimeline.mockReturnValue({
      positionTimeline: [],
      loading: false,
      refetch: jest.fn()
    });

    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      expect(screen.getByText('暂无职位历史记录')).toBeInTheDocument();
    });
  });

  it('should display salary range formatting', async () => {
    render(<EmployeePositionHistoryPage />);
    
    await waitFor(() => {
      // Check that salary ranges are formatted with currency
      expect(screen.getByText('¥18,000 - ¥22,000')).toBeInTheDocument();
    });
  });

  it('should show loading state during position change creation', async () => {
    mockUseCreatePositionChange.mockReturnValue({
      create: jest.fn().mockImplementation(() => new Promise(() => {})), // Never resolves
      loading: true
    });

    render(<EmployeePositionHistoryPage />);
    
    // Open modal
    const addButton = screen.getByText('添加职位变更');
    fireEvent.click(addButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建职位变更')).toBeInTheDocument();
    });

    // Submit button should show loading state
    const submitButton = screen.getByText('提交');
    expect(submitButton.closest('button')).toHaveClass('ant-btn-loading');
  });
});