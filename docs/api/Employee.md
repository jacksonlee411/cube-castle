# Employee API Documentation

## Overview

员工数据模型和API接口文档，包含传统REST API和现代SWR数据获取模式的使用指南。

**版本**: v2.0.0-alpha.2  
**最后更新**: 2025年8月1日

## 数据模型

### Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | 员工唯一标识符 | [optional] [default to undefined]
**employee_number** | **string** | 员工编号 | [default to undefined]
**first_name** | **string** | 名 | [default to undefined]
**last_name** | **string** | 姓 | [default to undefined]
**email** | **string** | 邮箱地址 | [default to undefined]
**phone_number** | **string** | 电话号码 | [optional] [default to undefined]
**hire_date** | **string** | 入职日期 (ISO 8601格式) | [default to undefined]
**position_id** | **string** | 职位ID | [optional] [default to undefined]
**organization_id** | **string** | 组织架构ID | [optional] [default to undefined]
**status** | **string** | 员工状态 (active/inactive/terminated) | [optional] [default to StatusEnum_Active]
**created_at** | **string** | 创建时间 (ISO 8601格式) | [optional] [default to undefined]
**updated_at** | **string** | 更新时间 (ISO 8601格式) | [optional] [default to undefined]

### TypeScript Interface

```typescript
interface Employee {
  id?: string;
  employee_number: string;
  first_name: string;
  last_name: string;
  email: string;
  phone_number?: string;
  hire_date: string;
  position_id?: string;
  organization_id?: string;
  status?: 'active' | 'inactive' | 'terminated';
  created_at?: string;
  updated_at?: string;
}
```

## SWR数据获取模式 🆕

### useEmployeesSWR Hook

现代化数据获取钩子，提供智能缓存、错误处理和性能监控。

```typescript
import { useEmployeesSWR } from '@/hooks/useEmployeesSWR';

// 获取员工列表
function EmployeeList() {
  const { 
    data: employees, 
    error, 
    isLoading, 
    refresh,
    isEmpty,
    isError 
  } = useEmployeesSWR.useEmployees();

  if (isLoading) return <div>加载中...</div>;
  if (isError) return <div>加载失败: {error.message}</div>;
  if (isEmpty) return <div>暂无员工数据</div>;

  return (
    <div>
      {employees.map(employee => (
        <EmployeeCard key={employee.id} employee={employee} />
      ))}
      <button onClick={refresh}>刷新数据</button>
    </div>
  );
}
```

### 单个员工详情

```typescript
import { useEmployeesSWR } from '@/hooks/useEmployeesSWR';

function EmployeeDetail({ employeeId }: { employeeId: string }) {
  const { 
    data: employee, 
    error, 
    isLoading 
  } = useEmployeesSWR.useEmployee(employeeId);

  if (isLoading) return <div>加载员工信息...</div>;
  if (error) return <div>员工信息加载失败</div>;

  return (
    <div>
      <h1>{employee.first_name} {employee.last_name}</h1>
      <p>员工编号: {employee.employee_number}</p>
      <p>邮箱: {employee.email}</p>
      <p>入职日期: {new Date(employee.hire_date).toLocaleDateString()}</p>
    </div>
  );
}
```

### 员工统计数据

```typescript
import { useEmployeesSWR } from '@/hooks/useEmployeesSWR';

function EmployeeStats() {
  const { 
    data: stats, 
    isLoading 
  } = useEmployeesSWR.useEmployeeStats();

  if (isLoading) return <div>加载统计数据...</div>;

  return (
    <div>
      <div>总员工数: {stats?.total || 0}</div>
      <div>在职员工: {stats?.active || 0}</div>
      <div>离职员工: {stats?.inactive || 0}</div>
      <div>本月新入职: {stats?.newHires || 0}</div>
    </div>
  );
}
```

## 缓存策略配置

### 智能缓存设置

```typescript
// 员工列表 - 中期缓存策略
{
  dedupingInterval: 10000,      // 10秒去重间隔
  refreshInterval: 300000,      // 5分钟后台刷新
  revalidateOnFocus: false,     // 焦点时不重新验证
  revalidateOnReconnect: true   // 网络重连时验证
}

// 员工详情 - 短期缓存策略
{
  dedupingInterval: 5000,       // 5秒去重间隔
  refreshInterval: 60000,       // 1分钟后台刷新
  revalidateOnFocus: true,      // 焦点时重新验证
  revalidateOnReconnect: true   // 网络重连时验证
}

// 统计数据 - 长期缓存策略
{
  dedupingInterval: 60000,      // 1分钟去重间隔
  refreshInterval: 900000,      // 15分钟后台刷新
  revalidateOnFocus: false,     // 焦点时不验证
  revalidateOnReconnect: true   // 网络重连时验证
}
```

## 传统REST API

### 基础用法 (已废弃，建议使用SWR)

```typescript
import { Employee } from 'cube-castle-api';

// ⚠️ 传统模式 - 不推荐
const instance: Employee = {
    id: "emp_123",
    employee_number: "E001",
    first_name: "张",
    last_name: "三",
    email: "zhang.san@company.com",
    phone_number: "+86 138 0013 8000",
    hire_date: "2024-01-15T00:00:00Z",
    position_id: "pos_456",
    organization_id: "org_789",
    status: "active",
    created_at: "2024-01-15T09:00:00Z",
    updated_at: "2024-01-15T09:00:00Z",
};
```

## 性能指标

### SWR架构优势

- **缓存命中率**: 70%+ (减少网络请求)
- **响应时间**: 首次加载500ms → 200ms
- **重复访问**: 提升50-70%加载速度
- **用户体验**: 后台自动数据更新
- **开发效率**: 50%代码量减少

### 监控集成

```typescript
// 自动性能监控
import { SWRMonitoring } from '@/components/ui/swr-monitoring';

function AdminPanel() {
  return (
    <div>
      <h1>系统监控</h1>
      <SWRMonitoring />
    </div>
  );
}
```

## 错误处理

### 统一错误处理策略

```typescript
const { data, error } = useEmployeesSWR.useEmployees();

// 错误类型判断
if (error) {
  switch (error.status) {
    case 401:
      // 未授权 - 重定向登录
      router.push('/login');
      break;
    case 403:
      // 禁止访问 - 显示权限提示
      toast.error('没有访问权限');
      break;
    case 404:
      // 资源不存在
      toast.error('员工信息不存在');
      break;
    case 500:
      // 服务器错误
      toast.error('服务器错误，请稍后重试');
      break;
    default:
      toast.error('未知错误');
  }
}
```

## 最佳实践

### 1. 数据获取模式选择

```typescript
// ✅ 推荐：使用SWR钩子
const { data, error, isLoading } = useEmployeesSWR.useEmployees();

// ❌ 不推荐：传统useEffect
useEffect(() => {
  fetchEmployees().then(setEmployees);
}, []);
```

### 2. 条件数据获取

```typescript
// 条件获取数据
const { data } = useEmployeesSWR.useEmployee(
  shouldFetch ? employeeId : null
);
```

### 3. 数据预加载

```typescript
// 预加载关联数据
useEffect(() => {
  if (employee?.position_id) {
    // 预加载职位信息
    mutate(`/api/positions/${employee.position_id}`);
  }
}, [employee]);
```

### 4. 实时数据更新

```typescript
// 手动触发数据更新
const handleEmployeeUpdate = async (updatedEmployee) => {
  // 乐观更新
  mutate('/api/employees', 
    employees => employees.map(emp => 
      emp.id === updatedEmployee.id ? updatedEmployee : emp
    ), 
    false
  );
  
  // 发送更新请求
  await updateEmployee(updatedEmployee);
  
  // 重新验证数据
  mutate('/api/employees');
};
```

## 迁移指南

### 从传统模式迁移到SWR

1. **替换useEffect数据获取**
```typescript
// 旧代码
useEffect(() => {
  setLoading(true);
  fetchEmployees()
    .then(setEmployees)
    .catch(setError)
    .finally(() => setLoading(false));
}, []);

// 新代码
const { data: employees, error, isLoading } = useEmployeesSWR.useEmployees();
```

2. **简化状态管理**
```typescript
// 旧代码
const [employees, setEmployees] = useState([]);
const [loading, setLoading] = useState(false);
const [error, setError] = useState(null);

// 新代码
const { data: employees, error, isLoading } = useEmployeesSWR.useEmployees();
```

3. **启用智能缓存**
```typescript
// 自动缓存管理，无需手动处理
const { data, mutate } = useEmployeesSWR.useEmployees();

// 手动刷新数据
const refreshData = () => mutate();
```

---

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

**文档维护**: 开发团队  
**技术支持**: [SWR架构实施方案](../architecture/swr_architecture_implementation.md)
