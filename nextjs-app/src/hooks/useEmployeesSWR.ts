import React from 'react';
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

// SWR fetcher function with simplified logging
const fetcher = async (url: string) => {
  console.log('🚀 SWR Fetcher: 开始获取数据', url);
  
  try {
    const response = await fetch(url);
    
    if (!response.ok) {
      const error = new Error(`HTTP ${response.status}: ${response.statusText}`);
      console.error('❌ SWR Fetcher: HTTP错误', response.status, response.statusText);
      throw error;
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

// Simplified SWR hook without complex memoization
export function useEmployeesSWR(options: UseEmployeesOptions = {}): UseEmployeesReturn {
  const { page = 1, pageSize = 50, search, department } = options;
  
  // Simple URL construction without memoization
  const params = new URLSearchParams();
  params.append('page', page.toString());
  params.append('page_size', pageSize.toString());
  if (search) params.append('search', search);
  if (department) params.append('department', department);
  
  const url = `/api/employees?${params.toString()}`;
  console.log('🔗 SWR URL:', url);
  
  // Use SWR with minimal configuration
  const { data, error, isLoading, mutate } = useSWR<EmployeesResponse>(
    url, 
    fetcher,
    {
      // Simple configuration without callbacks that might cause loops
      revalidateOnFocus: false,
      revalidateOnReconnect: true,
      revalidateIfStale: true,
      refreshInterval: 0, // Disable automatic refresh
      dedupingInterval: 2000,
    }
  );
  
  // Simple data transformation without memoization
  let employees: Employee[] = [];
  if (data?.employees && Array.isArray(data.employees)) {
    console.log('🔄 Transforming', data.employees.length, 'employees');
    employees = data.employees.map((emp: any) => ({
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
    }));
  } else {
    console.log('📊 No valid employees data');
  }
  
  console.log('📊 最终员工数据:', employees.length, '个员工');
  
  // Simple return without memoization
  return {
    employees,
    totalCount: data?.total_count || 0,
    isLoading: !!isLoading,
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
      revalidateOnFocus: false,
      revalidateOnReconnect: true,
      refreshInterval: 0,
      dedupingInterval: 5000,
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
    pageSize: 100,  // Backend limit is 100, not 1000
  });
  
  // Simple statistics calculation without SWR caching
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
  
  return {
    stats,
    departmentData,
    isLoading,
    isError,
  };
}