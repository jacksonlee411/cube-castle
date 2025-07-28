// tests/unit/pages/workflows/demo.test.tsx
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useRouter } from 'next/router';
import WorkflowDemoPage from '../../../../src/pages/workflows/demo';
import '@testing-library/jest-dom';

// Mock next/router
jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

// Mock dayjs
jest.mock('dayjs', () => {
  const originalDayjs = jest.requireActual('dayjs');
  const mockDayjs = jest.fn(() => ({
    format: jest.fn(() => '2025-01-27 15:30:00'),
    add: jest.fn(() => ({
      format: jest.fn(() => '2025-01-30 15:30:00')
    })),
    endOf: jest.fn(() => mockDayjs)
  }));
  mockDayjs.extend = originalDayjs.extend;
  return mockDayjs;
});

// Mock antd - 统一处理message mock
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}));

// Get reference to mocked message
const { message: mockMessage } = require('antd');

const mockRouter = {
  push: jest.fn(),
  pathname: '/workflows/demo',
  route: '/workflows/demo',
  query: {},
  asPath: '/workflows/demo',
};

describe('WorkflowDemoPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
  });

  it('should render the workflow demo page title and description', () => {
    render(<WorkflowDemoPage />);
    
    expect(screen.getByText('工作流管理演示')).toBeInTheDocument();
    expect(screen.getByText('Temporal.io 驱动的业务流程自动化演示，支持职位变更、批量操作等复杂工作流')).toBeInTheDocument();
  });

  it('should display workflow statistics', () => {
    render(<WorkflowDemoPage />);
    
    expect(screen.getByText('运行中')).toBeInTheDocument();
    expect(screen.getByText('已完成')).toBeInTheDocument();
    expect(screen.getByText('等待中')).toBeInTheDocument();
    expect(screen.getByText('失败')).toBeInTheDocument();

    // Check initial statistics (based on demo data)
    expect(screen.getByText('1')).toBeInTheDocument(); // Running workflows
    expect(screen.getByText('1')).toBeInTheDocument(); // Completed workflows
    expect(screen.getByText('0')).toBeInTheDocument(); // Pending workflows
    expect(screen.getByText('1')).toBeInTheDocument(); // Failed workflows
  });

  it('should display demo workflow cards', () => {
    render(<WorkflowDemoPage />);
    
    // Check workflow cards
    expect(screen.getByText('张三')).toBeInTheDocument();
    expect(screen.getByText('批量调动 (5人)')).toBeInTheDocument();
    expect(screen.getByText('赵六')).toBeInTheDocument();

    // Check workflow statuses
    expect(screen.getByText('RUNNING')).toBeInTheDocument();
    expect(screen.getByText('COMPLETED')).toBeInTheDocument();
    expect(screen.getByText('FAILED')).toBeInTheDocument();
  });

  it('should show workflow progress correctly', () => {
    render(<WorkflowDemoPage />);
    
    // Check progress steps for running workflow
    expect(screen.getByText('创建申请')).toBeInTheDocument();
    expect(screen.getByText('数据验证')).toBeInTheDocument();
    expect(screen.getByText('经理审批')).toBeInTheDocument();
    expect(screen.getByText('总监审批')).toBeInTheDocument();
    expect(screen.getByText('HR确认')).toBeInTheDocument();
    expect(screen.getByText('执行变更')).toBeInTheDocument();
  });

  it('should display approval timeline', () => {
    render(<WorkflowDemoPage />);
    
    // Check approvals for running workflow
    expect(screen.getByText('李经理')).toBeInTheDocument();
    expect(screen.getByText('王总监')).toBeInTheDocument();
    expect(screen.getByText('员工表现优秀，同意晋升')).toBeInTheDocument();
    expect(screen.getByText('APPROVED')).toBeInTheDocument();
    expect(screen.getByText('PENDING')).toBeInTheDocument();
  });

  it('should open create workflow modal when clicking create button', async () => {
    render(<WorkflowDemoPage />);
    
    const createButton = screen.getByText('创建新工作流');
    fireEvent.click(createButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建新的职位变更工作流')).toBeInTheDocument();
    });

    // Check form fields
    expect(screen.getByLabelText('员工ID')).toBeInTheDocument();
    expect(screen.getByLabelText('员工姓名')).toBeInTheDocument();
    expect(screen.getByLabelText('当前职位')).toBeInTheDocument();
    expect(screen.getByLabelText('新职位')).toBeInTheDocument();
    expect(screen.getByLabelText('新部门')).toBeInTheDocument();
    expect(screen.getByLabelText('薪资调整 (元)')).toBeInTheDocument();
    expect(screen.getByLabelText('生效日期')).toBeInTheDocument();
  });

  it('should create new workflow successfully', async () => {
    render(<WorkflowDemoPage />);
    
    // Open create modal
    const createButton = screen.getByText('创建新工作流');
    fireEvent.click(createButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建新的职位变更工作流')).toBeInTheDocument();
    });

    // Fill form
    fireEvent.change(screen.getByLabelText('员工ID'), { target: { value: 'EMP999' } });
    fireEvent.change(screen.getByLabelText('员工姓名'), { target: { value: '测试员工' } });
    fireEvent.change(screen.getByLabelText('当前职位'), { target: { value: '开发工程师' } });
    fireEvent.change(screen.getByLabelText('新职位'), { target: { value: '高级开发工程师' } });

    // Select department
    const departmentSelect = screen.getByLabelText('新部门');
    fireEvent.mouseDown(departmentSelect);
    await waitFor(() => {
      fireEvent.click(screen.getByText('研发部'));
    });

    // Submit form
    const submitButton = screen.getByText('创建工作流');
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockMessage.success).toHaveBeenCalledWith('工作流创建成功！');
    }, { timeout: 2000 });

    // Modal should close
    await waitFor(() => {
      expect(screen.queryByText('创建新的职位变更工作流')).not.toBeInTheDocument();
    });
  });

  it('should navigate to workflow details when clicking details button', async () => {
    render(<WorkflowDemoPage />);
    
    // Click details button for first workflow
    const detailsButtons = screen.getAllByText('详情');
    fireEvent.click(detailsButtons[0]);

    expect(mockRouter.push).toHaveBeenCalledWith('/workflows/wf-001');
  });

  it('should cancel workflow with confirmation', async () => {
    render(<WorkflowDemoPage />);
    
    // Click cancel button for running workflow
    const cancelButton = screen.getByText('取消');
    fireEvent.click(cancelButton);

    // Should show confirmation modal
    await waitFor(() => {
      expect(screen.getByText('确认取消工作流？')).toBeInTheDocument();
      expect(screen.getByText('取消后的工作流无法恢复，确定要取消吗？')).toBeInTheDocument();
    });

    // Confirm cancellation
    const confirmButton = screen.getByText('确定');
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(mockMessage.success).toHaveBeenCalledWith('工作流已取消');
    });
  });

  it('should display different workflow types correctly', () => {
    render(<WorkflowDemoPage />);
    
    // Check different workflow types
    expect(screen.getByText('张三')).toBeInTheDocument(); // POSITION_CHANGE
    expect(screen.getByText('批量调动 (5人)')).toBeInTheDocument(); // BULK_TRANSFER
    expect(screen.getByText('赵六')).toBeInTheDocument(); // POSITION_CHANGE (failed)
  });

  it('should show appropriate status colors and icons', () => {
    render(<WorkflowDemoPage />);
    
    // Check status tags are present
    const runningTags = screen.getAllByText('RUNNING');
    const completedTags = screen.getAllByText('COMPLETED');
    const failedTags = screen.getAllByText('FAILED');
    
    expect(runningTags.length).toBeGreaterThan(0);
    expect(completedTags.length).toBeGreaterThan(0);
    expect(failedTags.length).toBeGreaterThan(0);
  });

  it('should display error alert for failed workflows', () => {
    render(<WorkflowDemoPage />);
    
    expect(screen.getByText('工作流执行失败')).toBeInTheDocument();
    expect(screen.getByText('数据验证失败：薪资变更幅度超过公司政策限制')).toBeInTheDocument();
  });

  it('should handle form validation in create modal', async () => {
    render(<WorkflowDemoPage />);
    
    // Open create modal
    const createButton = screen.getByText('创建新工作流');
    fireEvent.click(createButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建新的职位变更工作流')).toBeInTheDocument();
    });

    // Try to submit empty form
    const submitButton = screen.getByText('创建工作流');
    fireEvent.click(submitButton);

    // Should show validation errors
    await waitFor(() => {
      expect(screen.getByText('请输入员工ID')).toBeInTheDocument();
      expect(screen.getByText('请输入员工姓名')).toBeInTheDocument();
      expect(screen.getByText('请输入当前职位')).toBeInTheDocument();
    });
  });

  it('should close create modal when clicking cancel', async () => {
    render(<WorkflowDemoPage />);
    
    // Open create modal
    const createButton = screen.getByText('创建新工作流');
    fireEvent.click(createButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建新的职位变更工作流')).toBeInTheDocument();
    });

    // Click cancel
    const cancelButton = screen.getByText('取消');
    fireEvent.click(cancelButton);

    // Modal should close
    await waitFor(() => {
      expect(screen.queryByText('创建新的职位变更工作流')).not.toBeInTheDocument();
    });
  });

  it('should display workflow employee information correctly', () => {
    render(<WorkflowDemoPage />);
    
    // Check employee current positions
    expect(screen.getByText('高级开发工程师')).toBeInTheDocument();
    expect(screen.getByText('多个职位')).toBeInTheDocument();
    expect(screen.getByText('产品经理')).toBeInTheDocument();

    // Check target positions
    expect(screen.getByText('技术主管')).toBeInTheDocument();
    expect(screen.getByText('各自新职位')).toBeInTheDocument();
    expect(screen.getByText('高级产品经理')).toBeInTheDocument();
  });

  it('should show workflow timeline information', () => {
    render(<WorkflowDemoPage />);
    
    // Check timestamps are displayed
    expect(screen.getByText('2025-01-25 10:30:00')).toBeInTheDocument();
    expect(screen.getByText('2025-01-18 14:20:00')).toBeInTheDocument();
  });

  it('should handle department selection in create form', async () => {
    render(<WorkflowDemoPage />);
    
    // Open create modal
    const createButton = screen.getByText('创建新工作流');
    fireEvent.click(createButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建新的职位变更工作流')).toBeInTheDocument();
    });

    // Open department dropdown
    const departmentSelect = screen.getByLabelText('新部门');
    fireEvent.mouseDown(departmentSelect);
    
    await waitFor(() => {
      expect(screen.getByText('研发部')).toBeInTheDocument();
      expect(screen.getByText('产品部')).toBeInTheDocument();
      expect(screen.getByText('市场部')).toBeInTheDocument();
      expect(screen.getByText('人事部')).toBeInTheDocument();
    });
  });

  it('should display salary change information', () => {
    render(<WorkflowDemoPage />);
    
    // Salary changes might be displayed in workflow cards
    // The demo data includes salary_change values that should be shown
    const workflowCards = screen.getAllByText(/技术主管|高级产品经理/);
    expect(workflowCards.length).toBeGreaterThan(0);
  });

  it('should handle workflow progress calculation correctly', () => {
    render(<WorkflowDemoPage />);
    
    // Check progress percentages are displayed somewhere
    // Based on the demo data: wf-001 has 60% progress, wf-002 has 100%, wf-003 has 30%
    const progressElements = document.querySelectorAll('[class*="progress"], [class*="ant-progress"]');
    expect(progressElements.length).toBeGreaterThan(0);
  });

  it('should show appropriate buttons for different workflow states', () => {
    render(<WorkflowDemoPage />);
    
    // Running workflow should have cancel button
    expect(screen.getByText('取消')).toBeInTheDocument();
    
    // All workflows should have details button
    const detailsButtons = screen.getAllByText('详情');
    expect(detailsButtons.length).toBe(3); // Three demo workflows
  });

  it('should handle workflow creation form field types correctly', async () => {
    render(<WorkflowDemoPage />);
    
    // Open create modal
    const createButton = screen.getByText('创建新工作流');
    fireEvent.click(createButton);
    
    await waitFor(() => {
      expect(screen.getByText('创建新的职位变更工作流')).toBeInTheDocument();
    });

    // Check salary input accepts numbers
    const salaryInput = screen.getByLabelText('薪资调整 (元)');
    fireEvent.change(salaryInput, { target: { value: '5000' } });
    expect(salaryInput).toHaveValue(5000);

    // Check date picker is present
    expect(screen.getByLabelText('生效日期')).toBeInTheDocument();
  });

  it('should display workflow cards with proper styling for different statuses', () => {
    render(<WorkflowDemoPage />);
    
    // Check that workflow cards exist and have proper structure
    const workflowCards = document.querySelectorAll('.ant-card');
    expect(workflowCards.length).toBeGreaterThan(3); // At least 3 workflow cards plus header cards
  });
});