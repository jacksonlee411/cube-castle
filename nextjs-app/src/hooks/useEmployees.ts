// src/hooks/useEmployees.ts
import { useState, useEffect, useCallback } from 'react';
import { useQuery, useMutation, useSubscription, gql } from '@apollo/client';
import { apolloClient } from '@/lib/graphql-client';
import { 
  GET_EMPLOYEES, 
  GET_EMPLOYEE, 
  GET_POSITION_TIMELINE,
  CREATE_POSITION_CHANGE,
  VALIDATE_POSITION_CHANGE,
  EMPLOYEE_POSITION_CHANGED 
} from '@/lib/graphql-queries';
import { restApiClient, handleApiError } from '@/lib/rest-api-client';
import { toast } from 'react-hot-toast';
import {
  GraphQLEmployee,
  GraphQLEmployeeFilters,
  GraphQLEmployeesResponse,
  GraphQLEmployeeResponse,
  GraphQLPositionChangeInput,
  GraphQLPositionChangeValidation,
  GraphQLPositionChangeResult,
  GraphQLEmployeePositionChangedPayload,
  GraphQLPositionTimelineResponse,
  GraphQLApprovePositionChangeResult,
  GraphQLRejectPositionChangeResult
} from '@/types/graphql';

// Re-export for backward compatibility
export type EmployeeFilters = GraphQLEmployeeFilters;
export type Employee = GraphQLEmployee;
export type CreatePositionChangeInput = GraphQLPositionChangeInput;

// Hook for fetching employees with filters and pagination (using REST API)
export const useEmployees = (
  filters?: EmployeeFilters,
  pageSize: number = 20
) => {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [hasNextPage, setHasNextPage] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalCount, setTotalCount] = useState(0);

  const fetchEmployees = async (page: number = 1, isLoadMore: boolean = false) => {
    setLoading(true);
    setError(null);
    
    try {
      // Build query parameters
      const queryParams = new URLSearchParams();
      queryParams.append('page', page.toString());
      queryParams.append('page_size', pageSize.toString());
      
      // Add filters
      if (filters?.search) queryParams.append('search', filters.search);
      if (filters?.department) queryParams.append('department', filters.department);
      if (filters?.status) queryParams.append('status', filters.status);
      if (filters?.employmentType) queryParams.append('employment_type', filters.employmentType);
      
      const response = await fetch(`/api/employees?${queryParams.toString()}`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const data = await response.json();

      if (data && data.employees) {
        const newEmployees = data.employees.map((emp: any) => ({
          id: emp.id,
          employeeId: emp.employee_number,
          legalName: `${emp.first_name} ${emp.last_name}`,
          preferredName: emp.first_name,
          email: emp.email,
          status: emp.status?.toUpperCase() || 'ACTIVE',
          hireDate: emp.hire_date,
          currentPosition: undefined, // Will be populated if needed
        }));

        if (isLoadMore) {
          setEmployees(prev => [...prev, ...newEmployees]);
        } else {
          setEmployees(newEmployees);
        }
        
        setTotalCount(data.total_count || 0);
        setHasNextPage(data.pagination?.has_next || false);
      } else {
        throw new Error('Failed to fetch employees');
      }
    } catch (err: any) {
      setError(err);
      handleApiError(err, '获取员工列表');
    } finally {
      setLoading(false);
    }
  };

  // Initial load and when filters change
  useEffect(() => {
    setCurrentPage(1);
    fetchEmployees(1, false);
  }, [filters, pageSize]);

  const loadMore = async () => {
    if (hasNextPage && !loading) {
      const nextPage = currentPage + 1;
      setCurrentPage(nextPage);
      await fetchEmployees(nextPage, true);
    }
  };

  const refresh = () => {
    setCurrentPage(1);
    setEmployees([]);
    fetchEmployees(1, false);
  };

  return {
    employees,
    loading,
    error,
    hasNextPage,
    loadMore,
    refresh,
    totalCount,
  };
};

// Hook for fetching a single employee with REST API
export const useEmployee = (employeeId: string) => {
  const [employee, setEmployee] = useState<GraphQLEmployee | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchEmployee = async () => {
    if (!employeeId) return;
    
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/employees/${employeeId}`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const emp = await response.json();
      
      if (emp) {
        setEmployee({
          id: emp.id,
          employeeId: emp.employee_number,
          legalName: `${emp.first_name} ${emp.last_name}`,
          preferredName: emp.first_name,
          email: emp.email,
          status: emp.status?.toUpperCase() || 'ACTIVE',
          hireDate: emp.hire_date,
          currentPosition: undefined, // Will be populated if needed
        });
      } else {
        throw new Error('Failed to fetch employee');
      }
    } catch (err: any) {
      setError(err);
      handleApiError(err, '获取员工信息');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchEmployee();
  }, [employeeId]);

  const refetch = () => {
    fetchEmployee();
  };

  return {
    employee,
    loading,
    error,
    refetch,
    isUsingRestFallback: true, // Always true since we're using REST API
  };
};

// Hook for creating position changes
export const useCreatePositionChange = () => {
  const [createPositionChange, { loading, error }] = useMutation<{ createPositionChange: GraphQLPositionChangeResult }>(CREATE_POSITION_CHANGE);
  const [validatePositionChange] = useMutation<{ validatePositionChange: GraphQLPositionChangeValidation }>(VALIDATE_POSITION_CHANGE);

  const create = async (input: CreatePositionChangeInput) => {
    try {
      // First validate the position change
      const validationResult = await validatePositionChange({
        variables: {
          employeeId: input.employeeId,
          effectiveDate: input.effectiveDate,
        },
      });

      const validation = validationResult.data?.validatePositionChange;
      
      if (!validation?.isValid) {
        const errors = validation?.errors ?? [];
        const errorMessages = errors.map(err => err.message).join(', ');
        throw new Error(`验证失败: ${errorMessages}`);
      }

      // Show warnings if any
      if (validation?.warnings && validation.warnings.length > 0) {
        validation.warnings.forEach(warning => {
        toast(warning.message, { icon: '⚠️' });
        });
      }

      // Create position change
      const result = await createPositionChange({
        variables: { input },
        refetchQueries: [GET_EMPLOYEES, GET_EMPLOYEE],
      });

      const data = result.data?.createPositionChange;
      
      if (data?.errors && data.errors.length > 0) {
        const errorMessages = data.errors.map(err => err.message).join(', ');
        throw new Error(errorMessages);
      }

      toast.success(
        data?.workflowId 
          ? `职位变更已提交，工作流 ${data.workflowId} 已启动，请等待审批。` 
          : '职位变更已完成。',
        {
          duration: 4000,
        }
      );

      return {
        success: true,
        positionHistory: data?.positionHistory,
        workflowId: data?.workflowId,
      };
    } catch (err: any) {
      toast.error(
        `职位变更失败: ${err.message || '创建职位变更时发生错误'}`,
        {
          duration: 5000,
        }
      );
      
      return {
        success: false,
        error: err.message,
      };
    }
  };

  return {
    create,
    loading,
    error,
  };
};

// Hook for employee filters with debouncing
export const useEmployeeFilters = (initialFilters?: EmployeeFilters) => {
  const [filters, setFilters] = useState<EmployeeFilters>(initialFilters || {});
  const [debouncedFilters, setDebouncedFilters] = useState<EmployeeFilters>(filters);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedFilters(filters);
    }, 300); // 300ms debounce

    return () => clearTimeout(timer);
  }, [filters]);

  const updateFilter = (key: keyof EmployeeFilters, value: string | undefined) => {
    setFilters(prev => ({
      ...prev,
      [key]: value || undefined,
    }));
  };

  const clearFilters = () => {
    setFilters({});
  };

  return {
    filters: debouncedFilters,
    updateFilter,
    clearFilters,
    hasActiveFilters: Object.values(filters).some(value => value !== undefined && value !== ''),
  };
};

// Hook for employee position timeline
export const usePositionTimeline = (employeeId: string, maxEntries: number = 20) => {
  const { data, loading, error, refetch } = useQuery(GET_POSITION_TIMELINE, {
    variables: { employeeId, maxEntries },
    skip: !employeeId,
    errorPolicy: 'all',
  });

  // Subscribe to position changes for real-time updates
  useSubscription(EMPLOYEE_POSITION_CHANGED, {
    variables: { employeeId },
    skip: !employeeId,
    onData: ({ data: subscriptionData }) => {
      if (subscriptionData?.data?.employeePositionChanged) {
        // Refetch timeline when position changes
        refetch();
        
        toast.success(
          '职位历史已更新',
          {
            duration: 3000,
          }
        );
      }
    },
  });

  return {
    positionTimeline: data?.employee?.positionTimeline || [],
    loading,
    error,
    refetch,
  };
};

// Hook for employee statistics
export const useEmployeeStats = () => {
  const { employees, loading } = useEmployees();

  const stats = {
    total: employees.length,
    active: employees.filter(emp => emp.status === 'ACTIVE').length,
    inactive: employees.filter(emp => emp.status === 'INACTIVE').length,
    pending: employees.filter(emp => emp.status === 'PENDING').length,
    byDepartment: employees.reduce((acc, emp) => {
      const dept = emp.currentPosition?.department || '未分配';
      acc[dept] = (acc[dept] || 0) + 1;
      return acc;
    }, {} as Record<string, number>),
    byEmploymentType: employees.reduce((acc, emp) => {
      const type = emp.currentPosition?.employmentType || '未设置';
      acc[type] = (acc[type] || 0) + 1;
      return acc;
    }, {} as Record<string, number>),
  };

  return {
    stats,
    loading,
  };
};