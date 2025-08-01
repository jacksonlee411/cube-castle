import useSWR from 'swr';
import { logger } from '@/lib/logger';

// Employee interface for the new SWR-based hooks
export interface Employee {
  id: string;
  employeeId: string;
  legalName: string;
  preferredName?: string | null;
  email: string;
  phone?: string;
  status: 'active' | 'inactive' | 'pending';
  hireDate: string;
  department?: string;
  position?: string;
  managerId?: string;
  managerName?: string | null;
  avatar?: string;
}

// SWR fetcher function with comprehensive logging and monitoring
const fetcher = async (url: string) => {
  const startTime = Date.now();
  console.log('🚀 SWR Fetcher: 开始获取数据', url);
  
  try {
    const response = await fetch(url);
    
    if (!response.ok) {
      const error = new Error(`HTTP ${response.status}: ${response.statusText}`);
      const duration = Date.now() - startTime;
      
      console.error('❌ SWR Fetcher: HTTP错误', response.status, response.statusText);
      logger.trackSWRRequest(url, false, duration, error);
      throw error;
    }
    
    const data = await response.json();
    const duration = Date.now() - startTime;
    
    console.log('✅ SWR Fetcher: 成功获取数据', data.employees?.length || 0, '个员工');
    logger.trackSWRRequest(url, true, duration);
    
    return data;
  } catch (error) {
    const duration = Date.now() - startTime;
    logger.trackSWRRequest(url, false, duration, error as Error);
    throw error;
  }
};

// Employee API interfaces
interface EmployeesResponse {
  employees: any[];
  total_count: number;
  page: number;
  page_size: number;
}

interface UseEmployeesOptions {
  page?: number;
  pageSize?: number;
  search?: string;
  department?: string;
}

interface UseEmployeesReturn {
  employees: Employee[];
  totalCount: number;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  mutate: () => Promise<any>;
}

// Main hook for employees data using SWR with intelligent caching
export function useEmployeesSWR(options: UseEmployeesOptions = {}): UseEmployeesReturn {
  const { page = 1, pageSize = 50, search, department } = options;
  
  // Build query parameters
  const params = new URLSearchParams();
  params.append('page', page.toString());
  params.append('page_size', pageSize.toString());
  if (search) params.append('search', search);
  if (department) params.append('department', department);
  
  const url = `/api/employees?${params.toString()}`;
  
  // Intelligent caching strategy based on data characteristics
  const getCachingStrategy = () => {
    // Real-time data for searches and filters (shorter cache)
    if (search || department) {
      return {
        dedupingInterval: 2000,     // 2 seconds deduplication
        refreshInterval: 30000,     // Refresh every 30s for filtered data
        revalidateOnFocus: true,    // Revalidate on focus for search results
        revalidateOnMount: true,    // Always fresh data for searches
      };
    }
    
    // Static data for main employee list (longer cache)
    return {
      dedupingInterval: 10000,    // 10 seconds deduplication
      refreshInterval: 300000,    // Refresh every 5 minutes for static data
      revalidateOnFocus: false,   // Don't revalidate on focus for static data
      revalidateOnMount: false,   // Use cached data when available
    };
  };
  
  const cachingStrategy = getCachingStrategy();
  
  // Use SWR for data fetching with optimized configuration
  const { data, error, isLoading, mutate } = useSWR<EmployeesResponse>(
    url, 
    fetcher,
    {
      // Dynamic caching strategy
      ...cachingStrategy,
      
      // Common configuration
      revalidateOnReconnect: true,  // Refetch when reconnecting
      errorRetryCount: 3,           // Retry failed requests 3 times
      errorRetryInterval: 1000,     // Wait 1s between retries
      
      // Performance optimization: Use background refresh
      revalidateIfStale: true,      // Revalidate stale data
      shouldRetryOnError: true,     // Retry on network errors
      
      // SWR callback hooks with enhanced logging
      onSuccess: (data) => {
        console.log('🎉 SWR Success: 数据加载成功', data.employees?.length || 0, '个员工');
        console.log('📊 缓存策略:', search || department ? '实时刷新' : '长期缓存');
      },
      onError: (error) => {
        console.error('💥 SWR Error: 数据加载失败', error.message);
      },
      onLoadingSlow: (key, config) => {
        console.warn('⏳ SWR Loading Slow: 请求超时', key);
        logger.warn('SWR', key, 'Request taking longer than expected');
      },
      
      // Performance optimization: Smart refresh
      refreshWhenHidden: false,     // Don't refresh when tab is hidden
      refreshWhenOffline: false,    // Don't refresh when offline
      
      // Advanced caching: Compare function for smart updates
      compare: (a, b) => {
        // Only update if employee count or data actually changed
        if (!a || !b) return false;
        return a.employees?.length === b.employees?.length && 
               JSON.stringify(a.employees?.[0]) === JSON.stringify(b.employees?.[0]);
      }
    }
  );
  
  // Transform API data to Employee interface
  const employees: Employee[] = data?.employees?.map((emp: any) => ({
    id: emp.id,
    employeeId: emp.employee_number,
    legalName: `${emp.first_name} ${emp.last_name}`,
    preferredName: emp.first_name || null,
    email: emp.email,
    phone: emp.phone_number || undefined,
    status: emp.status?.toLowerCase() === 'active' ? 'active' : 'inactive',
    hireDate: emp.hire_date,
    department: emp.department || '未分配部门',
    position: emp.position || '未设置职位',
    managerName: emp.manager_name || null,
  })) || [];
  
  return {
    employees,
    totalCount: data?.total_count || 0,
    isLoading,
    isError: !!error,
    error: error || null,
    mutate,
  };
}

// Hook for single employee with SWR and optimized caching
export function useEmployeeSWR(employeeId: string) {
  const { data, error, isLoading, mutate } = useSWR(
    employeeId ? `/api/employees/${employeeId}` : null,
    fetcher,
    {
      // Long-term caching for individual employee data (less frequently changing)
      dedupingInterval: 30000,      // 30 seconds deduplication
      refreshInterval: 600000,      // Refresh every 10 minutes
      revalidateOnFocus: false,     // Don't revalidate on focus for individual employee
      revalidateOnReconnect: true,  // Revalidate on reconnect
      revalidateOnMount: false,     // Use cache when available
      
      // Error handling
      errorRetryCount: 3,
      errorRetryInterval: 2000,     // Longer retry interval for individual requests
      
      // Performance optimization
      refreshWhenHidden: false,
      refreshWhenOffline: false,
      revalidateIfStale: true,
      
      // Enhanced logging
      onSuccess: (data) => {
        console.log('🎉 SWR Success: 单个员工数据加载成功', data.id);
        console.log('📊 缓存策略: 长期缓存 (10分钟)');
      },
      onError: (error) => {
        console.error('💥 SWR Error: 单个员工数据加载失败', error.message);
      },
      onLoadingSlow: (key, config) => {
        console.warn('⏳ SWR Loading Slow: 单个员工请求超时', key);
        logger.warn('SWR', key, 'Individual employee request slow');
      },
    }
  );
  
  return {
    employee: data,
    isLoading,
    isError: !!error,
    error,
    mutate,
  };
}

// Hook for employee statistics using SWR with aggressive caching
export function useEmployeeStatsSWR() {
  const { employees, isLoading, isError } = useEmployeesSWR({ 
    pageSize: 1000,  // Get more data for accurate statistics
  });
  
  // Use SWR for caching computed statistics
  const statsData = useSWR(
    employees.length > 0 ? 'employee-stats' : null,
    () => {
      console.log('📊 计算员工统计数据', employees.length, '个员工');
      
      const stats = {
        total: employees.length,
        active: employees.filter(emp => emp.status === 'active').length,
        inactive: employees.filter(emp => emp.status === 'inactive').length,
        pending: employees.filter(emp => emp.status === 'pending').length,
        departments: new Set(employees.map(emp => emp.department).filter(Boolean)).size,
      };
      
      // Department distribution for charts
      const departmentData = Array.from(
        employees.reduce((acc, emp) => {
          if (emp.department) {
            acc.set(emp.department, (acc.get(emp.department) || 0) + 1);
          }
          return acc;
        }, new Map())
      ).map(([department, count]) => ({
        label: department,
        value: count,
        color: `hsl(${Math.random() * 360}, 70%, 60%)`
      }));
      
      return { stats, departmentData };
    },
    {
      // Aggressive caching for statistics (computed data)
      dedupingInterval: 60000,     // 1 minute deduplication
      refreshInterval: 900000,     // Refresh every 15 minutes
      revalidateOnFocus: false,    // Don't revalidate on focus
      revalidateOnReconnect: false, // Don't revalidate on reconnect
      revalidateOnMount: false,    // Use cached computation
      
      // No retries for computed data
      errorRetryCount: 0,
      
      // Performance optimization
      refreshWhenHidden: false,
      refreshWhenOffline: false,
      
      onSuccess: (data) => {
        console.log('📊 统计数据计算完成:', data.stats);
      }
    }
  );
  
  return {
    stats: statsData.data?.stats || {
      total: 0,
      active: 0,
      inactive: 0,
      pending: 0,
      departments: 0,
    },
    departmentData: statsData.data?.departmentData || [],
    isLoading: isLoading || statsData.isLoading,
    isError: isError || statsData.error,
  };
}