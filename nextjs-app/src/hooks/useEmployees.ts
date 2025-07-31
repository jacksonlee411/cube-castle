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

export interface EmployeeFilters {
  search?: string;
  department?: string;
  status?: string;
  employmentType?: string;
  managerId?: string;
  hiredAfter?: string;
  hiredBefore?: string;
}

export interface Employee {
  id: string;
  employeeId: string;
  legalName: string;
  preferredName?: string;
  email: string;
  status: string;
  hireDate: string;
  terminationDate?: string;
  currentPosition?: {
    positionTitle: string;
    department: string;
    jobLevel?: string;
    location?: string;
    employmentType: string;
    effectiveDate: string;
  };
}

export interface CreatePositionChangeInput {
  employeeId: string;
  positionData: {
    positionTitle: string;
    department: string;
    jobLevel?: string;
    location?: string;
    employmentType: string;
    reportsToEmployeeId?: string;
    minSalary?: number;
    maxSalary?: number;
    currency?: string;
  };
  effectiveDate: string;
  changeReason?: string;
  isRetroactive?: boolean;
}

// Hook for fetching employees with filters and pagination
export const useEmployees = (
  filters?: EmployeeFilters,
  pageSize: number = 20
) => {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [hasNextPage, setHasNextPage] = useState(true);
  const [cursor, setCursor] = useState<string | null>(null);

  const { data, loading, error, fetchMore } = useQuery(GET_EMPLOYEES, {
    variables: {
      filters,
      first: pageSize,
      after: cursor,
    },
    notifyOnNetworkStatusChange: true,
    errorPolicy: 'all',
  });

  useEffect(() => {
    if (data?.employees) {
      const newEmployees = data.employees.edges.map((edge: any) => edge.node);
      
      if (cursor === null) {
        // Initial load or filter change
        setEmployees(newEmployees);
      } else {
        // Load more
        setEmployees(prev => [...prev, ...newEmployees]);
      }
      
      setHasNextPage(data.employees.pageInfo.hasNextPage);
    }
  }, [data, cursor]);

  const loadMore = async () => {
    if (hasNextPage && !loading) {
      const result = await fetchMore({
        variables: {
          after: data?.employees.pageInfo.endCursor,
        },
      });
      
      setCursor(result.data.employees.pageInfo.endCursor);
    }
  };

  const refresh = () => {
    setCursor(null);
    setEmployees([]);
  };

  return {
    employees,
    loading,
    error,
    hasNextPage,
    loadMore,
    refresh,
    totalCount: data?.employees?.totalCount || 0,
  };
};

// Hook for fetching a single employee with REST API fallback
export const useEmployee = (employeeId: string) => {
  const [restFallback, setRestFallback] = useState(false);
  const [restData, setRestData] = useState<any>(null);
  const [restLoading, setRestLoading] = useState(false);
  const [restError, setRestError] = useState<any>(null);
  const [apolloInitialized, setApolloInitialized] = useState(false);

  // Check if Apollo client is properly initialized
  useEffect(() => {
    const checkApolloClient = async () => {
      try {
        // Test if Apollo client can make a simple query
        const result = await apolloClient.query({
          query: gql`query TestQuery { __typename }`,
          fetchPolicy: 'network-only',
          errorPolicy: 'ignore'
        });
        setApolloInitialized(true);
      } catch (error) {
        // Apollo client not ready, using REST API immediately
        setRestFallback(true);
        setApolloInitialized(false);
      }
    };

    checkApolloClient();
  }, []);

  const tryRestApiFallback = useCallback(async () => {
    if (!restFallback && employeeId) {
      setRestFallback(true);
      setRestLoading(true);
      try {
        const result = await restApiClient.getEmployee(employeeId);
        if (result.success) {
          setRestData({ employee: result.data });
          setRestError(null);
        } else {
          setRestError(new Error(result.error || 'REST API failed'));
          handleApiError(result, 'GraphQL服务不可用，REST备用服务也失败');
        }
      } catch (err) {
        setRestError(err);
        handleApiError(err, '获取員工信息');
      } finally {
        setRestLoading(false);
      }
    }
  }, [restFallback, employeeId]);

  // Use REST API immediately if Apollo is not initialized or if forced
  useEffect(() => {
    if (employeeId && (!apolloInitialized || restFallback)) {
      tryRestApiFallback();
    }
  }, [employeeId, apolloInitialized, restFallback, tryRestApiFallback]);

  const { data, loading, error, refetch } = useQuery(GET_EMPLOYEE, {
    variables: { id: employeeId },
    skip: !employeeId || restFallback || !apolloInitialized,
    errorPolicy: 'all',
    fetchPolicy: 'cache-first', // Use cache-first to avoid immediate network requests
    onError: async (graphQLError) => {
      // GraphQL query failed, falling back to REST API
      await tryRestApiFallback();
    },
  });

  // Subscribe to position changes for this employee (only works with GraphQL)
  useSubscription(EMPLOYEE_POSITION_CHANGED, {
    variables: { employeeId },
    skip: !employeeId || restFallback,
    onData: ({ data: subscriptionData }) => {
      if (subscriptionData?.data?.employeePositionChanged) {
        // Refetch employee data when position changes
        if (restFallback) {
          // Refresh REST data
          setRestLoading(true);
          restApiClient.getEmployee(employeeId).then(result => {
            if (result.success) {
              setRestData({ employee: result.data });
            }
            setRestLoading(false);
          });
        } else {
          refetch();
        }
        
        toast.success(
          '职位信息已更新',
          {
            duration: 3000,
          }
        );
      }
    },
  });

  const retry = async () => {
    if (restFallback) {
      // Try to switch back to GraphQL
      setRestFallback(false);
      setRestData(null);
      setRestError(null);
      refetch();
    } else {
      refetch();
    }
  };

  return {
    employee: restFallback ? restData?.employee : data?.employee,
    loading: restFallback ? restLoading : loading,
    error: restFallback ? restError : error,
    refetch: retry,
    isUsingRestFallback: restFallback,
  };
};

// Hook for creating position changes
export const useCreatePositionChange = () => {
  const [createPositionChange, { loading, error }] = useMutation(CREATE_POSITION_CHANGE);
  const [validatePositionChange] = useMutation(VALIDATE_POSITION_CHANGE);

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
        const errors = validation?.errors || [];
        const errorMessages = errors.map((err: any) => err.message).join(', ');
        throw new Error(`验证失败: ${errorMessages}`);
      }

      // Show warnings if any
      if (validation?.warnings && validation.warnings.length > 0) {
        validation.warnings.forEach((warning: any) => {
        toast.warning(warning.message);
        });
      }

      // Create position change
      const result = await createPositionChange({
        variables: { input },
        refetchQueries: [GET_EMPLOYEES, GET_EMPLOYEE],
      });

      const data = result.data?.createPositionChange;
      
      if (data?.errors && data.errors.length > 0) {
        const errorMessages = data.errors.map((err: any) => err.message).join(', ');
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