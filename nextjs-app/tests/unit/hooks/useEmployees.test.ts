// tests/unit/hooks/useEmployees.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { useEmployees, useEmployee } from '@/hooks/useEmployees';

// 模拟GraphQL查询
const mockUseQuery = jest.fn();
const mockUseMutation = jest.fn();

jest.mock('@apollo/client', () => ({
  ...jest.requireActual('@apollo/client'),
  useQuery: (...args: any[]) => mockUseQuery(...args),
  useMutation: (...args: any[]) => mockUseMutation(...args),
  useSubscription: jest.fn(() => ({})),
}));

describe('useEmployees Hook', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // 设置默认的mock返回值
    mockUseQuery.mockReturnValue({
      data: {
        employees: {
          edges: [
            {
              node: {
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
            }
          ],
          pageInfo: {
            hasNextPage: false,
            endCursor: null
          },
          totalCount: 1
        }
      },
      loading: false,
      error: null,
      fetchMore: jest.fn(),
    });
  });

  it('正确初始化并返回员工数据', async () => {
    const { result } = renderHook(() => useEmployees());

    await waitFor(() => {
      expect(result.current.employees).toHaveLength(1);
      expect(result.current.employees[0].legalName).toBe('张三');
      expect(result.current.totalCount).toBe(1);
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBe(null);
    });
  });

  it('处理加载状态', () => {
    mockUseQuery.mockReturnValue({
      data: null,
      loading: true,
      error: null,
      fetchMore: jest.fn(),
    });

    const { result } = renderHook(() => useEmployees());

    expect(result.current.loading).toBe(true);
    expect(result.current.employees).toHaveLength(0);
  });

  it('处理错误状态', () => {
    const mockError = new Error('GraphQL error');
    mockUseQuery.mockReturnValue({
      data: null,
      loading: false,
      error: mockError,
      fetchMore: jest.fn(),
    });

    const { result } = renderHook(() => useEmployees());

    expect(result.current.error).toBe(mockError);
    expect(result.current.loading).toBe(false);
    expect(result.current.employees).toHaveLength(0);
  });

  it('使用过滤器参数', () => {
    const filters = {
      search: '张三',
      department: '技术部',
      status: 'ACTIVE'
    };

    renderHook(() => useEmployees(filters));

    expect(mockUseQuery).toHaveBeenCalledWith(
      expect.anything(), // GET_EMPLOYEES query
      {
        variables: {
          filters,
          first: 20,
          after: null,
        },
        notifyOnNetworkStatusChange: true,
        errorPolicy: 'all',
      }
    );
  });

  it('支持自定义页面大小', () => {
    const customPageSize = 50;
    renderHook(() => useEmployees({}, customPageSize));

    expect(mockUseQuery).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        variables: expect.objectContaining({
          first: customPageSize,
        }),
      })
    );
  });
});

describe('useEmployee Hook', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    mockUseQuery.mockReturnValue({
      data: {
        employee: {
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
      },
      loading: false,
      error: null,
      refetch: jest.fn(),
    });
  });

  it('正确返回单个员工数据', async () => {
    const { result } = renderHook(() => useEmployee('emp-1'));

    await waitFor(() => {
      expect(result.current.employee).toBeDefined();
      expect(result.current.employee.legalName).toBe('张三');
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBe(null);
    });
  });

  it('处理员工不存在的情况', () => {
    mockUseQuery.mockReturnValue({
      data: { employee: null },
      loading: false,
      error: null,
      refetch: jest.fn(),
    });

    const { result } = renderHook(() => useEmployee('non-existent-id'));

    expect(result.current.employee).toBe(null);
    expect(result.current.loading).toBe(false);
  });

  it('跳过查询当没有提供employeeId时', () => {
    renderHook(() => useEmployee(''));

    expect(mockUseQuery).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        skip: true, // 应该跳过查询
      })
    );
  });
});