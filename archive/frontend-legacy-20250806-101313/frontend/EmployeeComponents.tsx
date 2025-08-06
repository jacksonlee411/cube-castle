// 员工管理前端组件 - 8位编码优化版
// 版本: v1.0 Optimized
// 创建日期: 2025-08-05
// 基于: 8位编码员工管理API成功实现
// 架构: React + TypeScript + 零转换编码系统

import React, { useState, useEffect } from 'react';

// 8位编码员工类型定义
interface Employee {
  code: string;
  organization_code: string;
  primary_position_code?: string;
  employee_type: 'FULL_TIME' | 'PART_TIME' | 'CONTRACTOR' | 'INTERN';
  employment_status: 'ACTIVE' | 'TERMINATED' | 'ON_LEAVE' | 'PENDING_START';
  first_name: string;
  last_name: string;
  email: string;
  personal_email?: string;
  phone_number?: string;
  hire_date: string;
  termination_date?: string;
  personal_info?: string;
  employee_details?: string;
  tenant_id: string;
  created_at: string;
  updated_at: string;
}

interface EmployeeWithRelations extends Employee {
  organization?: {
    code: string;
    name: string;
    unit_type: string;
  };
  primary_position?: {
    code: string;
    position_type: string;
    status: string;
    details: string;
  };
  all_positions?: Array<{
    position_code: string;
    assignment_type: string;
    status: string;
    start_date: string;
    end_date?: string;
  }>;
  manager?: {
    code: string;
    first_name: string;
    last_name: string;
    email: string;
    employee_type: string;
  };
  direct_reports?: Array<{
    code: string;
    first_name: string;
    last_name: string;
    email: string;
    employee_type: string;
  }>;
}

interface EmployeeListResponse {
  employees: Employee[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

interface EmployeeStats {
  total_employees: number;
  active_employees: number;
  recent_hires_30days: number;
  by_type: Record<string, number>;
  by_status: Record<string, number>;
  by_organization: Record<string, number>;
}

// API客户端类 - 8位编码专用
class EmployeeAPI {
  private baseURL: string;

  constructor(baseURL: string = 'http://localhost:8084') {
    this.baseURL = baseURL;
  }

  // 验证8位员工编码格式
  private validateEmployeeCode(code: string): boolean {
    return /^[0-9]{8}$/.test(code) && 
           parseInt(code) >= 10000000 && 
           parseInt(code) <= 99999999;
  }

  // 验证7位组织编码格式
  private validateOrganizationCode(code: string): boolean {
    return /^[0-9]{7}$/.test(code) && 
           parseInt(code) >= 1000000 && 
           parseInt(code) <= 9999999;
  }

  // 验证7位职位编码格式
  private validatePositionCode(code: string): boolean {
    return /^[0-9]{7}$/.test(code) && 
           parseInt(code) >= 1000000 && 
           parseInt(code) <= 9999999;
  }

  // 获取员工列表
  async getAll(params?: {
    employee_type?: string;
    employment_status?: string;
    organization_code?: string;
    page?: number;
    page_size?: number;
  }): Promise<EmployeeListResponse> {
    const searchParams = new URLSearchParams();
    if (params?.employee_type) searchParams.set('employee_type', params.employee_type);
    if (params?.employment_status) searchParams.set('employment_status', params.employment_status);
    if (params?.organization_code) searchParams.set('organization_code', params.organization_code);
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.page_size) searchParams.set('page_size', params.page_size.toString());

    const response = await fetch(`${this.baseURL}/api/v1/employees?${searchParams}`);
    if (!response.ok) {
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  // 通过8位编码获取员工
  async getByCode(code: string, options?: {
    with_organization?: boolean;
    with_position?: boolean;
    with_all_positions?: boolean;
    with_manager?: boolean;
    with_direct_reports?: boolean;
  }): Promise<EmployeeWithRelations> {
    if (!this.validateEmployeeCode(code)) {
      throw new Error(`Invalid employee code: ${code}. Must be 8 digits (10000000-99999999).`);
    }

    const searchParams = new URLSearchParams();
    if (options?.with_organization) searchParams.set('with_organization', 'true');
    if (options?.with_position) searchParams.set('with_position', 'true');
    if (options?.with_all_positions) searchParams.set('with_all_positions', 'true');
    if (options?.with_manager) searchParams.set('with_manager', 'true');
    if (options?.with_direct_reports) searchParams.set('with_direct_reports', 'true');

    const response = await fetch(`${this.baseURL}/api/v1/employees/${code}?${searchParams}`);
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Employee not found: ${code}`);
      }
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  // 创建员工
  async create(employee: {
    organization_code: string;
    primary_position_code?: string;
    employee_type: string;
    employment_status?: string;
    first_name: string;
    last_name: string;
    email: string;
    personal_email?: string;
    phone_number?: string;
    hire_date: string;
    personal_info?: Record<string, any>;
    employee_details?: Record<string, any>;
  }): Promise<Employee> {
    if (!this.validateOrganizationCode(employee.organization_code)) {
      throw new Error('Invalid organization code: must be 7 digits (1000000-9999999)');
    }

    if (employee.primary_position_code && !this.validatePositionCode(employee.primary_position_code)) {
      throw new Error('Invalid position code: must be 7 digits (1000000-9999999)');
    }

    const response = await fetch(`${this.baseURL}/api/v1/employees`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(employee),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API error: ${response.status} ${errorText}`);
    }
    return response.json();
  }

  // 更新员工
  async update(code: string, updates: {
    organization_code?: string;
    primary_position_code?: string;
    employment_status?: string;
    email?: string;
    personal_email?: string;
    phone_number?: string;
    termination_date?: string;
    personal_info?: Record<string, any>;
    employee_details?: Record<string, any>;
  }): Promise<Employee> {
    if (!this.validateEmployeeCode(code)) {
      throw new Error('Invalid employee code: must be 8 digits (10000000-99999999)');
    }

    const response = await fetch(`${this.baseURL}/api/v1/employees/${code}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API error: ${response.status} ${errorText}`);
    }
    return response.json();
  }

  // 删除员工
  async delete(code: string): Promise<void> {
    if (!this.validateEmployeeCode(code)) {
      throw new Error('Invalid employee code: must be 8 digits (10000000-99999999)');
    }

    const response = await fetch(`${this.baseURL}/api/v1/employees/${code}`, {
      method: 'DELETE',
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API error: ${response.status} ${errorText}`);
    }
  }

  // 获取统计信息
  async getStats(): Promise<EmployeeStats> {
    const response = await fetch(`${this.baseURL}/api/v1/employees/stats`);
    if (!response.ok) {
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  // 健康检查
  async healthCheck(): Promise<{
    status: string;
    timestamp: string;
    service: string;
    version: string;
    features: string[];
  }> {
    const response = await fetch(`${this.baseURL}/health`);
    if (!response.ok) {
      throw new Error(`Health check failed: ${response.status}`);
    }
    return response.json();
  }
}

// React Hook - 员工数据管理
export const useEmployees = (apiBaseURL?: string) => {
  const [api] = useState(() => new EmployeeAPI(apiBaseURL));
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [stats, setStats] = useState<EmployeeStats | null>(null);

  // 获取员工列表
  const fetchEmployees = async (params?: {
    employee_type?: string;
    employment_status?: string;
    organization_code?: string;
    page?: number;
    page_size?: number;
  }) => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.getAll(params);
      setEmployees(response.employees);
      return response;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 获取单个员工
  const fetchEmployeeByCode = async (code: string, options?: {
    with_organization?: boolean;
    with_position?: boolean;
    with_all_positions?: boolean;
    with_manager?: boolean;
    with_direct_reports?: boolean;
  }) => {
    setLoading(true);
    setError(null);
    try {
      const employee = await api.getByCode(code, options);
      return employee;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 创建员工
  const createEmployee = async (employee: {
    organization_code: string;
    primary_position_code?: string;
    employee_type: string;
    employment_status?: string;
    first_name: string;
    last_name: string;
    email: string;
    personal_email?: string;
    phone_number?: string;
    hire_date: string;
    personal_info?: Record<string, any>;
    employee_details?: Record<string, any>;
  }) => {
    setLoading(true);
    setError(null);
    try {
      const newEmployee = await api.create(employee);
      // 刷新列表
      await fetchEmployees();
      return newEmployee;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 更新员工
  const updateEmployee = async (code: string, updates: {
    organization_code?: string;
    primary_position_code?: string;
    employment_status?: string;
    email?: string;
    personal_email?: string;
    phone_number?: string;
    termination_date?: string;
    personal_info?: Record<string, any>;
    employee_details?: Record<string, any>;
  }) => {
    setLoading(true);
    setError(null);
    try {
      const updatedEmployee = await api.update(code, updates);
      // 刷新列表
      await fetchEmployees();
      return updatedEmployee;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 删除员工
  const deleteEmployee = async (code: string) => {
    setLoading(true);
    setError(null);
    try {
      await api.delete(code);
      // 刷新列表
      await fetchEmployees();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 获取统计信息
  const fetchStats = async () => {
    try {
      const statsData = await api.getStats();
      setStats(statsData);
      return statsData;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    }
  };

  return {
    employees,
    loading,
    error,
    stats,
    fetchEmployees,
    fetchEmployeeByCode,
    createEmployee,
    updateEmployee,
    deleteEmployee,
    fetchStats,
    api
  };
};

// React组件 - 员工选择器
export const EmployeeSelector: React.FC<{
  onSelect: (employee: Employee) => void;
  filter?: { employee_type?: string; employment_status?: string; organization_code?: string };
  placeholder?: string;
  apiBaseURL?: string;
}> = ({ onSelect, filter = {}, placeholder = "选择员工", apiBaseURL }) => {
  const { employees, loading, error, fetchEmployees } = useEmployees(apiBaseURL);
  const [selectedCode, setSelectedCode] = useState<string>('');

  useEffect(() => {
    fetchEmployees(filter);
  }, [filter]);

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const code = event.target.value;
    setSelectedCode(code);
    
    const selected = employees.find(emp => emp.code === code);
    if (selected) {
      onSelect(selected);
    }
  };

  const parseDetails = (details?: string) => {
    try {
      return details ? JSON.parse(details) : {};
    } catch {
      return {};
    }
  };

  return (
    <div className="employee-selector">
      <select 
        value={selectedCode} 
        onChange={handleChange}
        disabled={loading}
        style={{
          padding: '8px 12px',
          border: '1px solid #ddd',
          borderRadius: '4px',
          fontSize: '14px',
          minWidth: '350px'
        }}
      >
        <option value="">{loading ? '加载中...' : placeholder}</option>
        {employees.map(emp => {
          const details = parseDetails(emp.employee_details);
          return (
            <option key={emp.code} value={emp.code}>
              {emp.code} - {emp.first_name}{emp.last_name} ({details.title || emp.employee_type}) - {emp.employment_status}
            </option>
          );
        })}
      </select>
      {error && (
        <div style={{ color: 'red', fontSize: '12px', marginTop: '4px' }}>
          {error}
        </div>
      )}
    </div>
  );
};

// React组件 - 员工表格
export const EmployeeTable: React.FC<{
  filter?: { employee_type?: string; employment_status?: string; organization_code?: string };
  onRowClick?: (employee: Employee) => void;
  onEdit?: (employee: Employee) => void;
  onDelete?: (employee: Employee) => void;
  apiBaseURL?: string;
}> = ({ filter = {}, onRowClick, onEdit, onDelete, apiBaseURL }) => {
  const { employees, loading, error, fetchEmployees, stats, fetchStats, deleteEmployee } = useEmployees(apiBaseURL);

  useEffect(() => {
    fetchEmployees(filter);
    fetchStats();
  }, [filter]);

  const parseDetails = (details?: string) => {
    try {
      return details ? JSON.parse(details) : {};
    } catch {
      return {};
    }
  };

  const handleDelete = async (employee: Employee) => {
    if (window.confirm(`确定要删除员工 ${employee.first_name}${employee.last_name} (${employee.code}) 吗？`)) {
      try {
        await deleteEmployee(employee.code);
        if (onDelete) onDelete(employee);
      } catch (err) {
        alert(`删除失败: ${err}`);
      }
    }
  };

  if (loading) {
    return <div style={{ padding: '20px', textAlign: 'center' }}>加载中...</div>;
  }

  if (error) {
    return <div style={{ padding: '20px', color: 'red' }}>错误: {error}</div>;
  }

  return (
    <div className="employee-table">
      {stats && (
        <div style={{ marginBottom: '20px', padding: '15px', backgroundColor: '#f8f9fa', borderRadius: '8px' }}>
          <h4 style={{ margin: '0 0 10px 0' }}>👥 员工统计</h4>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '15px' }}>
            <div>
              <strong>总计:</strong> {stats.total_employees} 名员工<br/>
              <strong>活跃:</strong> {stats.active_employees} 人<br/>
              <strong>新入职:</strong> {stats.recent_hires_30days} 人(30天内)
            </div>
            <div>
              <strong>按类型:</strong><br/>
              全职: {stats.by_type.FULL_TIME || 0}<br/>
              兼职: {stats.by_type.PART_TIME || 0}<br/>
              合同工: {stats.by_type.CONTRACTOR || 0}<br/>
              实习生: {stats.by_type.INTERN || 0}
            </div>
            <div>
              <strong>按状态:</strong><br/>
              在职: {stats.by_status.ACTIVE || 0}<br/>
              离职: {stats.by_status.TERMINATED || 0}<br/>
              休假: {stats.by_status.ON_LEAVE || 0}<br/>
              待入职: {stats.by_status.PENDING_START || 0}
            </div>
            <div>
              <strong>按组织:</strong><br/>
              {Object.entries(stats.by_organization).map(([org, count]) => (
                <div key={org}>{org}: {count}</div>
              ))}
            </div>
          </div>
        </div>
      )}
      
      <table style={{ width: '100%', borderCollapse: 'collapse', border: '1px solid #ddd', backgroundColor: 'white' }}>
        <thead>
          <tr style={{ backgroundColor: '#f8f9fa' }}>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>编码</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>姓名</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>职位</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>类型</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>状态</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>组织</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>入职日期</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>操作</th>
          </tr>
        </thead>
        <tbody>
          {employees.map(emp => {
            const details = parseDetails(emp.employee_details);
            return (
              <tr 
                key={emp.code}
                onClick={() => onRowClick?.(emp)}
                style={{ 
                  cursor: onRowClick ? 'pointer' : 'default',
                  backgroundColor: onRowClick ? 'transparent' : undefined
                }}
                onMouseEnter={(e) => {
                  if (onRowClick) e.currentTarget.style.backgroundColor = '#f8f9fa';
                }}
                onMouseLeave={(e) => {
                  if (onRowClick) e.currentTarget.style.backgroundColor = 'transparent';
                }}
              >
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  <code style={{ 
                    backgroundColor: '#e8f5e8', 
                    padding: '4px 6px', 
                    borderRadius: '4px',
                    color: '#2e7d32',
                    fontWeight: 'bold'
                  }}>
                    {emp.code}
                  </code>
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd', fontWeight: '500' }}>
                  {emp.first_name}{emp.last_name}
                  <br/>
                  <small style={{ color: '#666' }}>{emp.email}</small>
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  {details.title || '未设置职位名称'}
                  {emp.primary_position_code && (
                    <br/>
                    <small style={{ color: '#666' }}>#{emp.primary_position_code}</small>
                  )}
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  <span style={{
                    padding: '4px 8px',
                    borderRadius: '12px',
                    fontSize: '11px',
                    fontWeight: '500',
                    backgroundColor: emp.employee_type === 'FULL_TIME' ? '#e8f5e8' : 
                                 emp.employee_type === 'PART_TIME' ? '#fff3e0' : 
                                 emp.employee_type === 'CONTRACTOR' ? '#f3e5f5' : '#e3f2fd',
                    color: emp.employee_type === 'FULL_TIME' ? '#2e7d32' : 
                           emp.employee_type === 'PART_TIME' ? '#ef6c00' : 
                           emp.employee_type === 'CONTRACTOR' ? '#7b1fa2' : '#1565c0'
                  }}>
                    {emp.employee_type}
                  </span>
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  <span style={{
                    padding: '4px 8px',
                    borderRadius: '12px',
                    fontSize: '11px',
                    fontWeight: '500',
                    backgroundColor: emp.employment_status === 'ACTIVE' ? '#d4edda' : 
                                 emp.employment_status === 'TERMINATED' ? '#f8d7da' : 
                                 emp.employment_status === 'ON_LEAVE' ? '#fff3cd' : '#e2e3e5',
                    color: emp.employment_status === 'ACTIVE' ? '#155724' : 
                           emp.employment_status === 'TERMINATED' ? '#721c24' : 
                           emp.employment_status === 'ON_LEAVE' ? '#856404' : '#495057'
                  }}>
                    {emp.employment_status}
                  </span>
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  <code style={{ 
                    backgroundColor: '#f3e5f5', 
                    padding: '2px 4px', 
                    borderRadius: '2px', 
                    color: '#7b1fa2',
                    fontSize: '12px'
                  }}>
                    {emp.organization_code}
                  </code>
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  {new Date(emp.hire_date).toLocaleDateString('zh-CN')}
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  <div style={{ display: 'flex', gap: '8px' }}>
                    {onEdit && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          onEdit(emp);
                        }}
                        style={{
                          padding: '4px 8px',
                          fontSize: '12px',
                          border: '1px solid #007bff',
                          backgroundColor: 'white',
                          color: '#007bff',
                          borderRadius: '4px',
                          cursor: 'pointer'
                        }}
                      >
                        编辑
                      </button>
                    )}
                    {onDelete && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDelete(emp);
                        }}
                        style={{
                          padding: '4px 8px',
                          fontSize: '12px',
                          border: '1px solid #dc3545',
                          backgroundColor: 'white',
                          color: '#dc3545',
                          borderRadius: '4px',
                          cursor: 'pointer'
                        }}
                      >
                        删除
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>

      {employees.length === 0 && (
        <div style={{ 
          padding: '40px', 
          textAlign: 'center', 
          color: '#666',
          backgroundColor: 'white',
          border: '1px solid #ddd',
          borderTop: 'none'
        }}>
          暂无员工数据
        </div>
      )}
    </div>
  );
};

// React组件 - 员工创建表单
export const EmployeeCreateForm: React.FC<{
  onSuccess?: (employee: Employee) => void;
  onCancel?: () => void;
  apiBaseURL?: string;
}> = ({ onSuccess, onCancel, apiBaseURL }) => {
  const { createEmployee, loading, error } = useEmployees(apiBaseURL);
  const [formData, setFormData] = useState({
    organization_code: '',
    primary_position_code: '',
    employee_type: 'FULL_TIME',
    employment_status: 'ACTIVE',
    first_name: '',
    last_name: '',
    email: '',
    personal_email: '',
    phone_number: '',
    hire_date: new Date().toISOString().split('T')[0], // 今天日期
    title: '',
    level: '',
    salary: '',
    age: '',
    gender: '',
    address: ''
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const personal_info = {
      age: formData.age ? parseInt(formData.age) : undefined,
      gender: formData.gender || undefined,
      address: formData.address || undefined
    };

    const employee_details = {
      title: formData.title,
      level: formData.level || undefined,
      salary: formData.salary ? parseInt(formData.salary) : undefined
    };

    try {
      const employee = await createEmployee({
        organization_code: formData.organization_code,
        primary_position_code: formData.primary_position_code || undefined,
        employee_type: formData.employee_type,
        employment_status: formData.employment_status,
        first_name: formData.first_name,
        last_name: formData.last_name,
        email: formData.email,
        personal_email: formData.personal_email || undefined,
        phone_number: formData.phone_number || undefined,
        hire_date: formData.hire_date,
        personal_info: Object.keys(personal_info).some(key => personal_info[key as keyof typeof personal_info] !== undefined) ? personal_info : undefined,
        employee_details: Object.keys(employee_details).some(key => employee_details[key as keyof typeof employee_details] !== undefined) ? employee_details : undefined
      });
      
      if (onSuccess) onSuccess(employee);
      
      // 重置表单
      setFormData({
        organization_code: '',
        primary_position_code: '',
        employee_type: 'FULL_TIME',
        employment_status: 'ACTIVE',
        first_name: '',
        last_name: '',
        email: '',
        personal_email: '',
        phone_number: '',
        hire_date: new Date().toISOString().split('T')[0],
        title: '',
        level: '',
        salary: '',
        age: '',
        gender: '',
        address: ''
      });
    } catch (err) {
      // 错误已通过hook处理
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ 
      maxWidth: '700px', 
      padding: '20px', 
      border: '1px solid #ddd', 
      borderRadius: '8px',
      backgroundColor: 'white'
    }}>
      <h3 style={{ marginTop: 0 }}>👤 创建新员工</h3>
      
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '15px', marginBottom: '15px' }}>
        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            组织编码 (7位) *
          </label>
          <input
            type="text"
            value={formData.organization_code}
            onChange={(e) => setFormData({...formData, organization_code: e.target.value})}
            placeholder="1000000"
            pattern="[0-9]{7}"
            required
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>

        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            主要职位编码 (7位)
          </label>
          <input
            type="text"
            value={formData.primary_position_code}
            onChange={(e) => setFormData({...formData, primary_position_code: e.target.value})}
            placeholder="1000001"
            pattern="[0-9]{7}"
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '15px', marginBottom: '15px' }}>
        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            姓 *
          </label>
          <input
            type="text"
            value={formData.first_name}
            onChange={(e) => setFormData({...formData, first_name: e.target.value})}
            placeholder="张"
            required
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>

        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            名 *
          </label>
          <input
            type="text"
            value={formData.last_name}
            onChange={(e) => setFormData({...formData, last_name: e.target.value})}
            placeholder="三"
            required
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>
      </div>

      <div style={{ marginBottom: '15px' }}>
        <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
          邮箱 *
        </label>
        <input
          type="email"
          value={formData.email}
          onChange={(e) => setFormData({...formData, email: e.target.value})}
          placeholder="zhang.san@company.com"
          required
          style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
        />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '15px', marginBottom: '15px' }}>
        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            员工类型 *
          </label>
          <select
            value={formData.employee_type}
            onChange={(e) => setFormData({...formData, employee_type: e.target.value})}
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          >
            <option value="FULL_TIME">全职</option>
            <option value="PART_TIME">兼职</option>
            <option value="CONTRACTOR">合同工</option>
            <option value="INTERN">实习生</option>
          </select>
        </div>

        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            就业状态 *
          </label>
          <select
            value={formData.employment_status}
            onChange={(e) => setFormData({...formData, employment_status: e.target.value})}
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          >
            <option value="ACTIVE">在职</option>
            <option value="PENDING_START">待入职</option>
            <option value="ON_LEAVE">休假</option>
            <option value="TERMINATED">离职</option>
          </select>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '15px', marginBottom: '15px' }}>
        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            入职日期 *
          </label>
          <input
            type="date"
            value={formData.hire_date}
            onChange={(e) => setFormData({...formData, hire_date: e.target.value})}
            required
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>

        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            个人邮箱
          </label>
          <input
            type="email"
            value={formData.personal_email}
            onChange={(e) => setFormData({...formData, personal_email: e.target.value})}
            placeholder="zhang.san@gmail.com"
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>

        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            手机号码
          </label>
          <input
            type="tel"
            value={formData.phone_number}
            onChange={(e) => setFormData({...formData, phone_number: e.target.value})}
            placeholder="13800138000"
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>
      </div>

      <div style={{ marginBottom: '15px' }}>
        <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
          职位名称
        </label>
        <input
          type="text"
          value={formData.title}
          onChange={(e) => setFormData({...formData, title: e.target.value})}
          placeholder="高级软件工程师"
          style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
        />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '15px', marginBottom: '15px' }}>
        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            职级
          </label>
          <input
            type="text"
            value={formData.level}
            onChange={(e) => setFormData({...formData, level: e.target.value})}
            placeholder="P6"
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>

        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            薪资
          </label>
          <input
            type="number"
            value={formData.salary}
            onChange={(e) => setFormData({...formData, salary: e.target.value})}
            placeholder="25000"
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>

        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            年龄
          </label>
          <input
            type="number"
            min="18"
            max="65"
            value={formData.age}
            onChange={(e) => setFormData({...formData, age: e.target.value})}
            placeholder="30"
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '15px', marginBottom: '15px' }}>
        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            性别
          </label>
          <select
            value={formData.gender}
            onChange={(e) => setFormData({...formData, gender: e.target.value})}
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          >
            <option value="">请选择</option>
            <option value="M">男</option>
            <option value="F">女</option>
          </select>
        </div>

        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            地址
          </label>
          <input
            type="text"
            value={formData.address}
            onChange={(e) => setFormData({...formData, address: e.target.value})}
            placeholder="北京市朝阳区"
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>
      </div>

      {error && (
        <div style={{ 
          padding: '10px', 
          backgroundColor: '#f8d7da', 
          color: '#721c24', 
          borderRadius: '4px', 
          marginBottom: '15px',
          fontSize: '14px'
        }}>
          {error}
        </div>
      )}

      <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            disabled={loading}
            style={{
              padding: '10px 20px',
              border: '1px solid #6c757d',
              backgroundColor: 'white',
              color: '#6c757d',
              borderRadius: '4px',
              cursor: 'pointer'
            }}
          >
            取消
          </button>
        )}
        <button
          type="submit"
          disabled={loading}
          style={{
            padding: '10px 20px',
            border: 'none',
            backgroundColor: loading ? '#6c757d' : '#007bff',
            color: 'white',
            borderRadius: '4px',
            cursor: loading ? 'not-allowed' : 'pointer'
          }}
        >
          {loading ? '创建中...' : '创建员工'}
        </button>
      </div>
    </form>
  );
};

// 导出类型和组件
export type { Employee, EmployeeWithRelations, EmployeeListResponse, EmployeeStats };
export { EmployeeAPI };