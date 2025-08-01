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
    const response = await fetch(url);
    
    if (!response.ok) {
      const errorMessage = `HTTP ${response.status}: ${response.statusText}`;
      console.error('❌ SWR Fetcher: HTTP错误', response.status, response.statusText);
      
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
    
    const data = await response.json();
    console.log('✅ SWR Fetcher: 成功获取数据', data.employees?.length || 0, '个员工');
    
    return data;
  } catch (error) {
    console.error('💥 SWR Fetcher: 请求失败', error);
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
  
  // Enhanced SWR configuration with production-grade features
  const { data, error, isLoading, mutate } = useSWR<EmployeesResponse>(
    url, 
    fetcher,
    {
      // 数据同步策略
      revalidateOnFocus: true,           // 窗口聚焦时重新验证
      revalidateOnReconnect: true,       // 网络重连时重新验证  
      revalidateIfStale: true,           // 数据过期时重新验证
      refreshInterval: 30000,            // 30秒自动刷新
      
      // 缓存和去重策略
      dedupingInterval: 5000,            // 5秒内去重相同请求
      focusThrottleInterval: 5000,       // 聚焦节流间隔
      
      // 错误重试策略
      errorRetryCount: 3,                // 最多重试3次
      errorRetryInterval: 5000,          // 重试间隔5秒
      shouldRetryOnError: (error) => {
        // 对于客户端错误(4xx)不重试，对于服务器错误(5xx)重试
        if (error.message.includes('HTTP 4')) return false;
        return true;
      },
      
      // 成功回调
      onSuccess: (data) => {
        const count = data?.employees?.length || 0;
        console.log('🎉 SWR Success: 成功加载', count, '个员工');
        
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
  
  // Enhanced data transformation with memoization and error handling
  const employees = React.useMemo(() => {
    if (!data?.employees || !Array.isArray(data.employees)) {
      console.log('📊 No valid employees data');
      return [];
    }

    console.log('🔄 Transforming', data.employees.length, 'employees');
    
    try {
      return data.employees.map((emp: any) => ({
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
  }, [data?.employees]);
  
  console.log('📊 最终员工数据:', employees.length, '个员工');
  
  // Enhanced return with proper error handling
  return React.useMemo(() => ({
    employees,
    totalCount: data?.total_count || 0,
    isLoading: !!isLoading,
    isError: !!error,
    error: error || null,
    mutate,
  }), [employees, data?.total_count, isLoading, error, mutate]);
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