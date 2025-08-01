import React from 'react';
import useSWR from 'swr';
import { toast } from 'react-hot-toast';
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

// Enhanced SWR fetcher function with improved error handling
const fetcher = async (url: string) => {
  console.log('🚀 SWR Fetcher: 开始获取数据', url);
  
  try {
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache',
      },
    });
    
    if (!response.ok) {
      const errorMessage = `HTTP ${response.status}: ${response.statusText}`;
      console.error('❌ SWR Fetcher: HTTP错误', response.status, response.statusText);
      
      // Log response details for debugging
      const responseText = await response.text().catch(() => 'Unable to read response');
      console.error('🔍 Response details:', responseText.substring(0, 500));
      
      // 提供用户友好的错误信息
      if (response.status >= 500) {
        throw new Error('服务器暂时不可用，请稍后重试');
      } else if (response.status === 404) {
        throw new Error('请求的资源未找到');
      } else if (response.status === 403) {
        throw new Error('没有权限访问此资源');
      } else {
        throw new Error(errorMessage);
      }
    }
    
    const contentType = response.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
      console.error('❌ SWR Fetcher: 非JSON响应', contentType);
      throw new Error('服务器返回了无效的数据格式');
    }
    
    const data = await response.json();
    console.log('✅ SWR Fetcher: 成功获取数据', {
      hasEmployees: !!data.employees,
      employeesCount: data.employees?.length || 0,
      totalCount: data.total_count,
      dataKeys: Object.keys(data || {})
    });
    
    return data;
  } catch (error) {
    console.error('💥 SWR Fetcher: 请求失败', {
      error: error instanceof Error ? error.message : error,
      url,
      timestamp: new Date().toISOString()
    });
    throw error;
  }
};

// Employee API interfaces
interface EmployeesResponse {
  employees: any[];
  total_count: number;
  pagination?: {
    page: number;
    page_size: number;
    total_pages: number;
    has_next: boolean;
    has_prev: boolean;
  };
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

// Production-grade SWR hook with enhanced configuration
export function useEmployeesSWR(options: UseEmployeesOptions = {}): UseEmployeesReturn {
  const { page = 1, pageSize = 50, search, department } = options;
  
  // URL construction with memoization for performance
  const url = React.useMemo(() => {
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('page_size', pageSize.toString());
    if (search) params.append('search', search);
    if (department) params.append('department', department);
    return `/api/employees?${params.toString()}`;
  }, [page, pageSize, search, department]);
  
  console.log('🔗 SWR URL:', url);
  console.log('🔧 SWR Hook Initialization:', {
    url,
    page,
    pageSize,
    search,
    department
  });

  // Enhanced SWR configuration with production-grade features
  const { data, error, isLoading, mutate } = useSWR<EmployeesResponse>(
    url, 
    fetcher, // CRITICAL: Enable local fetcher for data fetching
    {
      // 数据同步策略 - 与全局配置一致
      revalidateOnFocus: true,           // 窗口聚焦时重新验证
      revalidateOnReconnect: true,       // 网络重连时重新验证  
      revalidateIfStale: true,           // 数据过期时重新验证
      revalidateOnMount: true,           // 挂载时重新验证
      refreshInterval: 0,                // 禁用自动刷新，与全局一致
      
      // 缓存和去重策略 - 与全局配置协调
      dedupingInterval: 0,               // 与全局一致：禁用去重，强制每次都获取
      focusThrottleInterval: 0,          // 与全局一致：禁用聚焦节流，立即执行
      
      // 强制数据获取设置
      suspense: false,                   // 不使用suspense模式
      shouldRetryOnError: true,          // 错误时重试
      
      // 错误重试策略 - 与全局配置协调
      errorRetryCount: 3,                // 本地hook更激进的重试策略
      errorRetryInterval: 1000,          // 更快的重试间隔
      
      // 强制初始数据获取
      fallbackData: undefined,           // 明确设置为undefined
      keepPreviousData: false,           // 不保留旧数据，强制重新获取
      
      // 成功回调
      onSuccess: (data) => {
        const count = data?.employees?.length || 0;
        console.log('🎉 SWR Success: 成功加载', count, '个员工');
        console.log('🔍 Success data details:', {
          hasData: !!data,
          dataKeys: data ? Object.keys(data) : [],
          employeesCount: count,
          totalCount: data?.total_count
        });
        
        // 仅在数据加载成功且有数据时显示成功提示
        if (count > 0 && !isLoading) {
          // 避免过于频繁的成功提示
          setTimeout(() => {
            console.log('📊 数据已更新');
          }, 100);
        }
      },
      
      // 错误回调
      onError: (error) => {
        console.error('❌ SWR Error:', error.message);
        console.error('🔍 Error details:', {
          errorType: typeof error,
          errorMessage: error.message,
          errorStack: error.stack?.substring(0, 200),
          url,
          timestamp: new Date().toISOString()
        });
        
        // 显示用户友好的错误提示
        toast.error(`数据加载失败: ${error.message}`, {
          duration: 4000,
          position: 'top-right',
        });
      },
      
      // 加载状态回调
      onLoadingSlow: () => {
        console.warn('⏳ SWR: 数据加载较慢');
        toast.loading('正在加载员工数据...', {
          duration: 2000,
        });
      },
      
      // 慢加载阈值
      loadingTimeout: 3000,              // 3秒后触发慢加载提示
    }
  );
  
  // 🔥 CRITICAL FIX: Direct data fetch if SWR fails
  const [directData, setDirectData] = React.useState<EmployeesResponse | null>(null);
  const [directLoading, setDirectLoading] = React.useState(false);
  const [directError, setDirectError] = React.useState<Error | null>(null);
  
  React.useEffect(() => {
    // More aggressive fallback - check earlier and more frequently
    const fallbackTimer = setTimeout(async () => {
      if (!data && !error && !directData && !isLoading) {
        console.log('🔥 SWR未在500ms内触发，启用直接数据获取');
        setDirectLoading(true);
        setDirectError(null);
        
        try {
          const response = await fetch(url, {
            headers: {
              'Content-Type': 'application/json',
              'Cache-Control': 'no-cache',
            },
          });
          if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
          }
          const fetchedData = await response.json();
          console.log('🔥 直接获取成功:', {
            employeesCount: fetchedData.employees?.length || 0,
            totalCount: fetchedData.total_count,
            url
          });
          setDirectData(fetchedData);
          
          // Show success toast for direct fetch
          toast.success(`数据直接获取成功: ${fetchedData.employees?.length || 0} 个员工`, {
            duration: 3000,
            position: 'top-right',
          });
        } catch (err) {
          console.error('🔥 直接获取失败:', err);
          setDirectError(err as Error);
        } finally {
          setDirectLoading(false);
        }
      }
    }, 500); // Reduced from 2000ms to 500ms for faster fallback
    
    // Clear direct data if SWR starts working
    if (data) {
      console.log('✅ SWR数据到达，清除直接数据回退');
      setDirectData(null);
      setDirectError(null);
      setDirectLoading(false);
      clearTimeout(fallbackTimer);
    }
    
    return () => clearTimeout(fallbackTimer);
  }, [url, data, error, directData, isLoading]);
  
  // CRITICAL FIX: More aggressive SWR triggering
  React.useEffect(() => {
    const timer = setTimeout(() => {
      if (!data && !isLoading && !error && !directData && !directLoading) {
        console.log('🚨 SWR未自动触发，强制执行mutate');
        mutate().then((result) => {
          console.log('🚨 强制mutate结果:', !!result);
        }).catch((error) => {
          console.error('🚨 强制mutate失败:', error);
        });
      }
    }, 200); // Very early check - 200ms
    
    return () => clearTimeout(timer);
  }, [data, isLoading, error, mutate, directData, directLoading]);
  
  // Additional trigger on component mount
  React.useEffect(() => {
    console.log('🚀 Hook挂载，立即尝试数据获取');
    // Immediate attempt
    setTimeout(() => {
      if (!data && !isLoading) {
        console.log('🚀 立即触发mutate');
        mutate();
      }
    }, 50); // Very immediate
  }, []); // Only on mount
  
  // Enhanced data transformation with memoization and error handling
  const employees = React.useMemo(() => {
    // Use direct data as fallback if SWR data is not available
    const activeData = data || directData;
    
    console.log('🔍 SWR Data Analysis:', {
      hasData: !!activeData,
      dataType: typeof activeData,
      dataKeys: activeData ? Object.keys(activeData) : [],
      hasEmployees: !!activeData?.employees,
      employeesType: typeof activeData?.employees,
      employeesLength: Array.isArray(activeData?.employees) ? activeData.employees.length : 'not-array',
      totalCount: activeData?.total_count,
      rawData: activeData ? JSON.stringify(activeData).substring(0, 200) + '...' : 'null',
      usingDirectData: !data && !!directData
    });

    if (!activeData?.employees || !Array.isArray(activeData.employees)) {
      console.log('📊 No valid employees data');
      return [];
    }

    console.log('🔄 Transforming', activeData.employees.length, 'employees');
    
    try {
      return activeData.employees.map((emp: any) => ({
        id: emp.id || '',
        employeeId: emp.employee_number || '',
        legalName: `${emp.first_name || ''} ${emp.last_name || ''}`.trim(),
        preferredName: emp.first_name || null,
        email: emp.email || '',
        phone: emp.phone_number || undefined,
        status: emp.status?.toLowerCase() === 'active' ? 'active' : 'inactive',
        hireDate: emp.hire_date || '',
        department: emp.department || '未分配部门',
        position: emp.position || '未设置职位',
        managerName: emp.manager_name || null,
      })) as Employee[];
    } catch (transformError) {
      console.error('❌ 员工数据转换失败:', transformError);
      toast.error('数据格式错误，请联系管理员');
      return [];
    }
  }, [data?.employees, directData?.employees]);
  
  console.log('📊 最终员工数据:', employees.length, '个员工');
  
  // Enhanced return with proper error handling and direct data fallback
  return React.useMemo(() => {
    const activeData = data || directData;
    const activeError = error || directError;
    const activeLoading = isLoading || directLoading;
    
    return {
      employees,
      totalCount: activeData?.total_count || 0,
      isLoading: activeLoading,
      isError: !!activeError,
      error: activeError || null,
      mutate,
    };
  }, [employees, data?.total_count, directData?.total_count, isLoading, directLoading, error, directError, mutate]);
}

// Enhanced hook for single employee with production-grade caching
export function useEmployeeSWR(employeeId: string) {
  const { data, error, isLoading, mutate } = useSWR(
    employeeId ? `/api/employees/${employeeId}` : null,
    fetcher,
    {
      revalidateOnFocus: true,         // 聚焦时重新验证
      revalidateOnReconnect: true,     // 重连时重新验证
      refreshInterval: 60000,          // 60秒自动刷新 (单个员工数据变化较少)
      dedupingInterval: 10000,         // 10秒去重间隔 (单个员工查询频率较低)
      errorRetryCount: 2,              // 最多重试2次
      errorRetryInterval: 3000,        // 3秒重试间隔
      
      onError: (error) => {
        console.error('❌ 单个员工数据加载失败:', error);
        toast.error(`员工信息加载失败: ${error.message}`);
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

// Enhanced hook for employee statistics with intelligent caching
export function useEmployeeStatsSWR() {
  const { employees, isLoading, isError } = useEmployeesSWR({ 
    pageSize: 100,  // Backend limit is 100, not 1000
  });
  
  // Memoized statistics calculation for performance
  const stats = React.useMemo(() => ({
    total: employees.length,
    active: employees.filter(emp => emp.status === 'active').length,
    inactive: employees.filter(emp => emp.status === 'inactive').length,
    pending: employees.filter(emp => emp.status === 'pending').length,
    departments: new Set(employees.map(emp => emp.department).filter(Boolean)).size,
  }), [employees]);
  
  // Memoized department distribution for charts
  const departmentData = React.useMemo(() => {
    const departmentMap = employees.reduce((acc, emp) => {
      if (emp.department) {
        acc.set(emp.department, (acc.get(emp.department) || 0) + 1);
      }
      return acc;
    }, new Map());

    return Array.from(departmentMap.entries()).map(([department, count]) => ({
      label: department,
      value: count,
      color: `hsl(${Math.abs(department.split('').reduce((a: number, b: string) => a + b.charCodeAt(0), 0)) % 360}, 70%, 60%)`
    }));
  }, [employees]);
  
  return {
    stats,
    departmentData,
    isLoading,
    isError,
  };
}