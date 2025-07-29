// tests/unit/pages/admin/graph-sync.test.tsx
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import GraphSyncAdminPage from '../../../../src/pages/admin/graph-sync';
import { FULL_GRAPH_SYNC, SYNC_DEPARTMENT, SYNC_EMPLOYEE_TO_GRAPH } from '../../../../src/lib/graphql-queries';
import '@testing-library/jest-dom';

// Mock antd notification
const mockNotification = {
  success: jest.fn(),
  error: jest.fn(),
  warning: jest.fn(),
  info: jest.fn(),
};

jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  notification: mockNotification,
}));

// Mock dayjs
jest.mock('dayjs', () => {
  const mockDayjs = jest.fn(() => ({
    format: jest.fn(() => '2025-01-27 15:30:00')
  }));
  return mockDayjs;
});

const successfulFullSyncMock = {
  request: {
    query: FULL_GRAPH_SYNC,
  },
  result: {
    data: {
      fullGraphSync: {
        success: true,
        syncedEmployees: 150,
        syncedPositions: 200,
        syncedRelationships: 85,
        errors: []
      }
    }
  }
};

const failedFullSyncMock = {
  request: {
    query: FULL_GRAPH_SYNC,
  },
  result: {
    data: {
      fullGraphSync: {
        success: false,
        syncedEmployees: 120,
        syncedPositions: 150,
        syncedRelationships: 60,
        errors: ['Connection timeout', 'Invalid data format in record 45']
      }
    }
  }
};

const successfulDeptSyncMock = {
  request: {
    query: SYNC_DEPARTMENT,
    variables: {
      department: '研发部'
    }
  },
  result: {
    data: {
      syncDepartment: {
        success: true,
        syncedEmployees: 25,
        syncedPositions: 30,
        syncedRelationships: 15,
        errors: []
      }
    }
  }
};

const successfulEmployeeSyncMock = {
  request: {
    query: SYNC_EMPLOYEE_TO_GRAPH,
    variables: {
      employeeId: 'EMP001'
    }
  },
  result: {
    data: {
      syncEmployeeToGraph: true
    }
  }
};

describe('GraphSyncAdminPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render the graph sync admin page with correct title', () => {
    render(
      <MockedProvider mocks={[]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    expect(screen.getByText('图数据库同步管理')).toBeInTheDocument();
    expect(screen.getByText('管理员工数据与Neo4j图数据库的同步操作')).toBeInTheDocument();
  });

  it('should display sync status information', () => {
    render(
      <MockedProvider mocks={[]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    expect(screen.getByText('同步状态')).toBeInTheDocument();
    expect(screen.getByText('待开始')).toBeInTheDocument(); // Initial idle status
  });

  it('should display sync operation buttons', () => {
    render(
      <MockedProvider mocks={[]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    expect(screen.getByText('完整同步')).toBeInTheDocument();
    expect(screen.getByText('部门同步')).toBeInTheDocument();
    expect(screen.getByText('单个员工同步')).toBeInTheDocument();
  });

  it('should execute full sync successfully', async () => {
    render(
      <MockedProvider mocks={[successfulFullSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const fullSyncButton = screen.getByText('完整同步');
    fireEvent.click(fullSyncButton);

    // Should show loading state
    await waitFor(() => {
      expect(screen.getByText('同步中...')).toBeInTheDocument();
    });

    // Should show success results
    await waitFor(() => {
      expect(mockNotification.success).toHaveBeenCalledWith({
        message: '完整同步成功',
        description: '已同步 150 个员工，200 个职位，85 个关系'
      });
    });

    // Should display sync results
    await waitFor(() => {
      expect(screen.getByText('150')).toBeInTheDocument(); // Synced employees
      expect(screen.getByText('200')).toBeInTheDocument(); // Synced positions
      expect(screen.getByText('85')).toBeInTheDocument(); // Synced relationships
    });
  });

  it('should handle full sync failure', async () => {
    render(
      <MockedProvider mocks={[failedFullSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const fullSyncButton = screen.getByText('完整同步');
    fireEvent.click(fullSyncButton);

    await waitFor(() => {
      expect(mockNotification.error).toHaveBeenCalledWith({
        message: '同步失败',
        description: '同步过程中遇到 2 个错误'
      });
    });

    // Should display error details
    await waitFor(() => {
      expect(screen.getByText('Connection timeout')).toBeInTheDocument();
      expect(screen.getByText('Invalid data format in record 45')).toBeInTheDocument();
    });
  });

  it('should execute department sync successfully', async () => {
    render(
      <MockedProvider mocks={[successfulDeptSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const deptSyncButton = screen.getByText('部门同步');
    fireEvent.click(deptSyncButton);

    // Should show department selection modal
    await waitFor(() => {
      expect(screen.getByText('选择部门')).toBeInTheDocument();
    });

    // Select department
    const departmentSelect = screen.getByDisplayValue('选择部门');
    fireEvent.mouseDown(departmentSelect);
    
    await waitFor(() => {
      const researchOption = screen.getByText('研发部');
      fireEvent.click(researchOption);
    });

    // Click sync button in modal
    const syncButton = screen.getByText('开始同步');
    fireEvent.click(syncButton);

    await waitFor(() => {
      expect(mockNotification.success).toHaveBeenCalledWith({
        message: '部门同步成功',
        description: '已同步 25 个员工，30 个职位，15 个关系'
      });
    });
  });

  it('should execute single employee sync successfully', async () => {
    render(
      <MockedProvider mocks={[successfulEmployeeSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const empSyncButton = screen.getByText('单个员工同步');
    fireEvent.click(empSyncButton);

    // Should show employee selection modal
    await waitFor(() => {
      expect(screen.getByText('选择员工')).toBeInTheDocument();
    });

    // Enter employee ID
    const employeeInput = screen.getByPlaceholderText('输入员工ID');
    fireEvent.change(employeeInput, { target: { value: 'EMP001' } });

    // Click sync button in modal
    const syncButton = screen.getByText('开始同步');
    fireEvent.click(syncButton);

    await waitFor(() => {
      expect(mockNotification.success).toHaveBeenCalledWith({
        message: '员工同步成功',
        description: '员工 EMP001 已成功同步到图数据库'
      });
    });
  });

  it('should display sync progress during operation', async () => {
    render(
      <MockedProvider mocks={[successfulFullSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const fullSyncButton = screen.getByText('完整同步');
    fireEvent.click(fullSyncButton);

    // Should show progress indicator
    await waitFor(() => {
      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });
  });

  it('should show last sync time after successful sync', async () => {
    render(
      <MockedProvider mocks={[successfulFullSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const fullSyncButton = screen.getByText('完整同步');
    fireEvent.click(fullSyncButton);

    await waitFor(() => {
      expect(screen.getByText('最后同步时间')).toBeInTheDocument();
      expect(screen.getByText('2025-01-27 15:30:00')).toBeInTheDocument();
    });
  });

  it('should disable buttons during sync operation', async () => {
    render(
      <MockedProvider mocks={[successfulFullSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const fullSyncButton = screen.getByText('完整同步');
    const deptSyncButton = screen.getByText('部门同步');
    const empSyncButton = screen.getByText('单个员工同步');

    fireEvent.click(fullSyncButton);

    // All buttons should be disabled during sync
    await waitFor(() => {
      expect(fullSyncButton.closest('button')).toBeDisabled();
      expect(deptSyncButton.closest('button')).toBeDisabled();
      expect(empSyncButton.closest('button')).toBeDisabled();
    });
  });

  it('should show warning alert before full sync', async () => {
    render(
      <MockedProvider mocks={[successfulFullSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const fullSyncButton = screen.getByText('完整同步');
    fireEvent.click(fullSyncButton);

    // Should show confirmation modal
    await waitFor(() => {
      expect(screen.getByText('确认完整同步')).toBeInTheDocument();
      expect(screen.getByText('完整同步将会重新同步所有员工数据，这可能需要几分钟时间。确定要继续吗？')).toBeInTheDocument();
    });

    // Confirm sync
    const confirmButton = screen.getByText('确认');
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(screen.getByText('同步中...')).toBeInTheDocument();
    });
  });

  it('should cancel sync confirmation', async () => {
    render(
      <MockedProvider mocks={[]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const fullSyncButton = screen.getByText('完整同步');
    fireEvent.click(fullSyncButton);

    await waitFor(() => {
      expect(screen.getByText('确认完整同步')).toBeInTheDocument();
    });

    // Cancel sync
    const cancelButton = screen.getByText('取消');
    fireEvent.click(cancelButton);

    // Modal should close and no sync should happen
    await waitFor(() => {
      expect(screen.queryByText('确认完整同步')).not.toBeInTheDocument();
    });

    expect(screen.getByText('待开始')).toBeInTheDocument(); // Status should remain idle
  });

  it('should display sync statistics', async () => {
    render(
      <MockedProvider mocks={[successfulFullSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    expect(screen.getByText('同步统计')).toBeInTheDocument();
    expect(screen.getByText('已同步员工')).toBeInTheDocument();
    expect(screen.getByText('已同步职位')).toBeInTheDocument();
    expect(screen.getByText('已同步关系')).toBeInTheDocument();
    expect(screen.getByText('错误数量')).toBeInTheDocument();
  });

  it('should show error details in expandable table', async () => {
    render(
      <MockedProvider mocks={[failedFullSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const fullSyncButton = screen.getByText('完整同步');
    fireEvent.click(fullSyncButton);

    await waitFor(() => {
      expect(screen.getByText('确认完整同步')).toBeInTheDocument();
    });

    const confirmButton = screen.getByText('确认');
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(screen.getByText('错误详情')).toBeInTheDocument();
    });

    // Should show error table
    expect(screen.getByText('Connection timeout')).toBeInTheDocument();
    expect(screen.getByText('Invalid data format in record 45')).toBeInTheDocument();
  });

  it('should handle GraphQL errors gracefully', async () => {
    const errorMock = {
      request: {
        query: FULL_GRAPH_SYNC,
      },
      error: new Error('Network error'),
    };

    render(
      <MockedProvider mocks={[errorMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const fullSyncButton = screen.getByText('完整同步');
    fireEvent.click(fullSyncButton);

    await waitFor(() => {
      expect(screen.getByText('确认完整同步')).toBeInTheDocument();
    });

    const confirmButton = screen.getByText('确认');
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(mockNotification.error).toHaveBeenCalledWith({
        message: '同步失败',
        description: '网络错误，请稍后重试'
      });
    });
  });

  it('should validate employee ID input', async () => {
    render(
      <MockedProvider mocks={[]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const empSyncButton = screen.getByText('单个员工同步');
    fireEvent.click(empSyncButton);

    await waitFor(() => {
      expect(screen.getByText('选择员工')).toBeInTheDocument();
    });

    // Try to sync without entering employee ID
    const syncButton = screen.getByText('开始同步');
    fireEvent.click(syncButton);

    // Should show validation error
    await waitFor(() => {
      expect(screen.getByText('请输入员工ID')).toBeInTheDocument();
    });
  });

  it('should close modal when clicking cancel in department sync', async () => {
    render(
      <MockedProvider mocks={[]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    const deptSyncButton = screen.getByText('部门同步');
    fireEvent.click(deptSyncButton);

    await waitFor(() => {
      expect(screen.getByText('选择部门')).toBeInTheDocument();
    });

    const cancelButton = screen.getByText('取消');
    fireEvent.click(cancelButton);

    await waitFor(() => {
      expect(screen.queryByText('选择部门')).not.toBeInTheDocument();
    });
  });

  it('should show appropriate status icons', async () => {
    render(
      <MockedProvider mocks={[successfulFullSyncMock]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    // Initial status should show clock icon
    expect(screen.getByLabelText('clock-circle')).toBeInTheDocument();

    const fullSyncButton = screen.getByText('完整同步');
    fireEvent.click(fullSyncButton);

    await waitFor(() => {
      expect(screen.getByText('确认完整同步')).toBeInTheDocument();
    });

    const confirmButton = screen.getByText('确认');
    fireEvent.click(confirmButton);

    // During sync should show loading icon
    await waitFor(() => {
      expect(screen.getByLabelText('loading')).toBeInTheDocument();
    });

    // After success should show check icon
    await waitFor(() => {
      expect(screen.getByLabelText('check-circle')).toBeInTheDocument();
    });
  });

  it('should display sync operation instructions', () => {
    render(
      <MockedProvider mocks={[]} addTypename={false}>
        <GraphSyncAdminPage />
      </MockedProvider>
    );
    
    expect(screen.getByText('操作说明')).toBeInTheDocument();
    expect(screen.getByText('同步所有员工数据到图数据库')).toBeInTheDocument();
    expect(screen.getByText('同步指定部门的员工数据')).toBeInTheDocument();
    expect(screen.getByText('同步单个员工数据')).toBeInTheDocument();
  });
});