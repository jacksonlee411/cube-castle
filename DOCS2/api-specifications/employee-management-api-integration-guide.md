# 员工管理API集成指南

**版本**: v2.0 Person Name Optimized  
**创建日期**: 2025-08-05  
**目标读者**: 前端开发者、第三方集成开发者  
**前置要求**: 熟悉RESTful API和JavaScript/TypeScript

## 📋 快速开始

### 1. 基础配置

```javascript
// API客户端配置
const EMPLOYEE_API_BASE = 'http://localhost:8084';
const API_VERSION = 'v1';

// 基础请求头
const DEFAULT_HEADERS = {
  'Content-Type': 'application/json',
  'Authorization': 'Bearer YOUR_JWT_TOKEN',
  'X-Tenant-ID': 'your-tenant-id'
};
```

### 2. 核心概念速览

```typescript
// 员工编码系统
interface EmployeeCoding {
  employee_code: string;        // 8位：10000001-99999999
  organization_code: string;    // 7位：1000000-9999999  
  primary_position_code?: string; // 7位：1000000-9999999
}

// Person Name简化设计
interface PersonName {
  person_name: string;         // 完整姓名（必填）
  first_name?: string;         // 姓（可选）
  last_name?: string;          // 名（可选）
}
```

## 🚀 常用操作示例

### 获取员工列表

```javascript
async function getEmployees(params = {}) {
  const queryParams = new URLSearchParams({
    page: params.page || 1,
    page_size: params.page_size || 20,
    ...params.filters
  });
  
  const response = await fetch(
    `${EMPLOYEE_API_BASE}/api/v1/employees?${queryParams}`,
    { headers: DEFAULT_HEADERS }
  );
  
  return response.json();
}

// 使用示例
const employees = await getEmployees({
  page: 1,
  page_size: 10,
  filters: {
    employee_type: 'FULL_TIME',
    employment_status: 'ACTIVE'
  }
});
```

### 获取单个员工

```javascript
async function getEmployee(employeeCode, options = {}) {
  const queryParams = new URLSearchParams();
  
  // 关联信息选项
  if (options.with_organization) queryParams.set('with_organization', 'true');
  if (options.with_position) queryParams.set('with_position', 'true');
  if (options.with_all_positions) queryParams.set('with_all_positions', 'true');
  
  const response = await fetch(
    `${EMPLOYEE_API_BASE}/api/v1/employees/${employeeCode}?${queryParams}`,
    { headers: DEFAULT_HEADERS }
  );
  
  if (!response.ok) {
    throw new Error(`Employee ${employeeCode} not found`);
  }
  
  return response.json();
}

// 使用示例
const employee = await getEmployee('10000001', {
  with_organization: true,
  with_position: true
});
```

### 创建员工

```javascript
async function createEmployee(employeeData) {
  // 验证必填字段
  if (!employeeData.person_name || !employeeData.email || !employeeData.hire_date) {
    throw new Error('person_name, email, hire_date are required');
  }
  
  const response = await fetch(
    `${EMPLOYEE_API_BASE}/api/v1/employees`,
    {
      method: 'POST',
      headers: DEFAULT_HEADERS,
      body: JSON.stringify(employeeData)
    }
  );
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error.message);
  }
  
  return response.json();
}

// 使用示例
const newEmployee = await createEmployee({
  organization_code: '1000000',
  primary_position_code: '1000001',
  employee_type: 'FULL_TIME',
  employment_status: 'ACTIVE',
  
  // Person Name字段
  person_name: '李四',
  first_name: '李',
  last_name: '四',
  
  email: 'li.si@company.com',
  personal_email: 'li.si@gmail.com',
  phone_number: '13800138001',
  hire_date: '2025-08-05',
  
  personal_info: {
    age: 30,
    gender: 'M'
  },
  employee_details: {
    title: '产品经理',
    level: 'P7'
  }
});
```

### 更新员工信息

```javascript
async function updateEmployee(employeeCode, updates) {
  const response = await fetch(
    `${EMPLOYEE_API_BASE}/api/v1/employees/${employeeCode}`,
    {
      method: 'PUT',
      headers: DEFAULT_HEADERS,
      body: JSON.stringify(updates)
    }
  );
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error.message);
  }
  
  return response.json();
}

// 使用示例
const updatedEmployee = await updateEmployee('10000001', {
  employment_status: 'ON_LEAVE',
  person_name: '张三（更新）',
  phone_number: '13800138888'
});
```

### 获取员工统计

```javascript
async function getEmployeeStats() {
  const response = await fetch(
    `${EMPLOYEE_API_BASE}/api/v1/employees/stats`,
    { headers: DEFAULT_HEADERS }
  );
  
  return response.json();
}

// 使用示例
const stats = await getEmployeeStats();
console.log(`总员工数: ${stats.total_employees}`);
console.log(`活跃员工: ${stats.active_employees}`);
```

## 🎨 React集成示例

### 员工管理Hook

```typescript
import { useState, useEffect } from 'react';

interface Employee {
  employee_code: string;
  organization_code: string;
  primary_position_code?: string;
  employee_type: 'FULL_TIME' | 'PART_TIME' | 'CONTRACTOR' | 'INTERN';
  employment_status: 'ACTIVE' | 'TERMINATED' | 'ON_LEAVE' | 'PENDING_START';
  person_name: string;
  first_name?: string;
  last_name?: string;
  email: string;
  personal_email?: string;
  phone_number?: string;
  hire_date: string;
  // ... 其他字段
}

export const useEmployees = () => {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchEmployees = async (params = {}) => {
    setLoading(true);
    setError(null);
    
    try {
      const data = await getEmployees(params);
      setEmployees(data.employees);
      return data;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const createEmployee = async (employeeData: any) => {
    setLoading(true);
    try {
      const newEmployee = await createEmployee(employeeData);
      await fetchEmployees(); // 刷新列表
      return newEmployee;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return {
    employees,
    loading,
    error,
    fetchEmployees,
    createEmployee
  };
};
```

### 员工列表组件

```typescript
import React from 'react';
import { useEmployees } from './useEmployees';

export const EmployeeList: React.FC = () => {
  const { employees, loading, error, fetchEmployees } = useEmployees();

  useEffect(() => {
    fetchEmployees();
  }, []);

  if (loading) return <div>加载中...</div>;
  if (error) return <div>错误: {error}</div>;

  return (
    <div>
      <h2>员工列表</h2>
      <table>
        <thead>
          <tr>
            <th>员工编码</th>
            <th>姓名</th>
            <th>邮箱</th>
            <th>类型</th>
            <th>状态</th>
          </tr>
        </thead>
        <tbody>
          {employees.map(emp => (
            <tr key={emp.employee_code}>
              <td>{emp.employee_code}</td>
              <td>
                <div>{emp.person_name}</div>
                {emp.first_name && emp.last_name && (
                  <small>{emp.first_name} {emp.last_name}</small>
                )}
              </td>
              <td>{emp.email}</td>
              <td>{emp.employee_type}</td>
              <td>{emp.employment_status}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
```

## 🔧 错误处理最佳实践

### 错误类型处理

```javascript
class EmployeeAPIError extends Error {
  constructor(response, errorData) {
    super(errorData.error.message);
    this.name = 'EmployeeAPIError';
    this.status = response.status;
    this.code = errorData.error.code;
    this.details = errorData.error.details;
  }
}

async function handleAPICall(apiCall) {
  try {
    return await apiCall();
  } catch (response) {
    if (response.status) {
      const errorData = await response.json();
      throw new EmployeeAPIError(response, errorData);
    }
    throw response; // 网络错误等
  }
}

// 使用示例
try {
  const employee = await handleAPICall(() => 
    getEmployee('invalid-code')
  );
} catch (error) {
  if (error instanceof EmployeeAPIError) {
    switch (error.status) {
      case 400:
        console.log('请求参数错误:', error.details);
        break;
      case 404:
        console.log('员工不存在');
        break;
      case 409:
        console.log('邮箱已存在');
        break;
      default:
        console.log('API错误:', error.message);
    }
  } else {
    console.log('网络错误:', error);
  }
}
```

### 表单验证辅助函数

```javascript
// 员工编码验证
export const validateEmployeeCode = (code) => {
  if (!code || code.length !== 8) {
    return '员工编码必须是8位数字';
  }
  if (!/^\d{8}$/.test(code)) {
    return '员工编码只能包含数字';
  }
  const codeNum = parseInt(code);
  if (codeNum < 10000000 || codeNum > 99999999) {
    return '员工编码必须在10000000-99999999范围内';
  }
  return null;
};

// Person Name验证
export const validatePersonName = (name) => {
  if (!name || name.trim().length === 0) {
    return '完整姓名不能为空';
  }
  if (name.length > 200) {
    return '姓名长度不能超过200字符';
  }
  return null;
};

// 邮箱验证
export const validateEmail = (email) => {
  if (!email) {
    return '邮箱不能为空';
  }
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) {
    return '邮箱格式不正确';
  }
  return null;
};

// 综合表单验证
export const validateEmployeeForm = (formData) => {
  const errors = {};
  
  const personNameError = validatePersonName(formData.person_name);
  if (personNameError) errors.person_name = personNameError;
  
  const emailError = validateEmail(formData.email);
  if (emailError) errors.email = emailError;
  
  if (!formData.hire_date) {
    errors.hire_date = '入职日期不能为空';
  }
  
  if (!formData.organization_code) {
    errors.organization_code = '组织编码不能为空';
  }
  
  return Object.keys(errors).length > 0 ? errors : null;
};
```

## 📊 性能优化建议

### 缓存策略

```javascript
class EmployeeCache {
  constructor(ttl = 5 * 60 * 1000) { // 5分钟TTL
    this.cache = new Map();
    this.ttl = ttl;
  }
  
  set(key, value) {
    this.cache.set(key, {
      value,
      timestamp: Date.now()
    });
  }
  
  get(key) {
    const item = this.cache.get(key);
    if (!item) return null;
    
    if (Date.now() - item.timestamp > this.ttl) {
      this.cache.delete(key);
      return null;
    }
    
    return item.value;
  }
  
  clear() {
    this.cache.clear();
  }
}

const employeeCache = new EmployeeCache();

// 带缓存的员工查询
async function getCachedEmployee(employeeCode) {
  const cacheKey = `employee_${employeeCode}`;
  const cached = employeeCache.get(cacheKey);
  
  if (cached) {
    return cached;
  }
  
  const employee = await getEmployee(employeeCode);
  employeeCache.set(cacheKey, employee);
  
  return employee;
}
```

### 批量操作

```javascript
// 批量获取员工
async function getBatchEmployees(employeeCodes) {
  const promises = employeeCodes.map(code => 
    getCachedEmployee(code).catch(err => ({
      employee_code: code,
      error: err.message
    }))
  );
  
  return Promise.all(promises);
}

// 使用示例
const employees = await getBatchEmployees([
  '10000001', '10000002', '10000003'
]);

employees.forEach(emp => {
  if (emp.error) {
    console.log(`员工 ${emp.employee_code} 获取失败: ${emp.error}`);
  } else {
    console.log(`员工: ${emp.person_name}`);
  }
});
```

## 🧪 测试示例

### 单元测试

```javascript
// Jest测试示例
describe('Employee API', () => {
  beforeEach(() => {
    fetch.resetMocks();
  });

  test('should get employee by code', async () => {
    const mockEmployee = {
      employee_code: '10000001',
      person_name: '张三',
      email: 'zhang.san@company.com'
    };
    
    fetch.mockResponseOnce(JSON.stringify(mockEmployee));
    
    const employee = await getEmployee('10000001');
    
    expect(employee).toEqual(mockEmployee);
    expect(fetch).toHaveBeenCalledWith(
      'http://localhost:8084/api/v1/employees/10000001?',
      expect.objectContaining({
        headers: DEFAULT_HEADERS
      })
    );
  });

  test('should handle employee not found', async () => {
    fetch.mockRejectOnce(new Response('Not Found', { status: 404 }));
    
    await expect(getEmployee('99999999')).rejects.toThrow('Employee 99999999 not found');
  });

  test('should validate employee code format', () => {
    expect(validateEmployeeCode('123')).toBe('员工编码必须是8位数字');
    expect(validateEmployeeCode('abcd1234')).toBe('员工编码只能包含数字');
    expect(validateEmployeeCode('10000001')).toBeNull();
  });
});
```

## 🔒 安全最佳实践

### JWT Token管理

```javascript
class TokenManager {
  constructor() {
    this.token = localStorage.getItem('jwt_token');
    this.refreshToken = localStorage.getItem('refresh_token');
  }
  
  setTokens(token, refreshToken) {
    this.token = token;
    this.refreshToken = refreshToken;
    localStorage.setItem('jwt_token', token);
    localStorage.setItem('refresh_token', refreshToken);
  }
  
  clearTokens() {
    this.token = null;
    this.refreshToken = null;
    localStorage.removeItem('jwt_token');
    localStorage.removeItem('refresh_token');
  }
  
  getAuthHeaders() {
    return this.token ? {
      'Authorization': `Bearer ${this.token}`
    } : {};
  }
  
  async refreshTokenIfNeeded() {
    // 实现token刷新logic
    if (this.isTokenExpiringSoon()) {
      await this.refreshAccessToken();
    }
  }
}

const tokenManager = new TokenManager();

// 带自动token刷新的API调用
async function authenticatedFetch(url, options = {}) {
  await tokenManager.refreshTokenIfNeeded();
  
  return fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      ...tokenManager.getAuthHeaders()
    }
  });
}
```

## 📋 集成检查清单

### 开发前准备
- [ ] 确认API基础地址和版本
- [ ] 获取有效的JWT Token
- [ ] 了解租户ID配置
- [ ] 阅读员工编码规范

### 基础功能实现
- [ ] 实现员工列表查询
- [ ] 实现单个员工查询
- [ ] 实现员工创建功能
- [ ] 实现员工更新功能
- [ ] 实现Person Name字段显示

### 高级功能实现
- [ ] 实现关联查询（组织、职位）
- [ ] 实现分页和筛选
- [ ] 实现统计信息展示
- [ ] 实现错误处理和验证
- [ ] 实现缓存和性能优化

### 测试和部署
- [ ] 编写单元测试
- [ ] 执行集成测试
- [ ] 进行性能测试
- [ ] 完成安全审查
- [ ] 准备生产环境配置

---

**📞 技术支持**:
- API文档: [员工管理API规范](./employee-management-api-specification.md)
- 基础地址: `http://localhost:8084`
- 健康检查: `http://localhost:8084/health`
- 版本: v2.0 Person Name Optimized