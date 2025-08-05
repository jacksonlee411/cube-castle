# 前端集成组件使用指南

## 📦 组织单元前端组件

### 🚀 快速开始

```bash
# 安装依赖
npm install react @types/react

# 导入组件
import { 
  OrganizationSelector, 
  OrganizationTable, 
  useOrganizationUnits,
  OrganizationAPI 
} from './OrganizationComponents';
```

## 🧩 组件说明

### 1. OrganizationAPI 类
```typescript
const api = new OrganizationAPI('http://localhost:8080');

// 获取组织列表
const response = await api.getAll({ 
  unit_type: 'DEPARTMENT', 
  status: 'ACTIVE' 
});

// 通过7位编码获取单个组织
const org = await api.getByCode('1000000');

// 获取统计信息
const stats = await api.getStats();

// 健康检查
const health = await api.healthCheck();
```

### 2. useOrganizationUnits Hook
```typescript
function MyComponent() {
  const { 
    organizations, 
    loading, 
    error, 
    fetchOrganizations,
    fetchStats 
  } = useOrganizationUnits();

  useEffect(() => {
    fetchOrganizations({ status: 'ACTIVE' });
  }, []);

  return (
    <div>
      {loading && <div>加载中...</div>}
      {error && <div>错误: {error}</div>}
      {organizations.map(org => (
        <div key={org.code}>{org.name}</div>
      ))}
    </div>
  );
}
```

### 3. OrganizationSelector 组件
```typescript
function MyForm() {
  const [selectedOrg, setSelectedOrg] = useState<OrganizationUnit | null>(null);

  return (
    <OrganizationSelector
      onSelect={setSelectedOrg}
      filter={{ unit_type: 'DEPARTMENT', status: 'ACTIVE' }}
      placeholder="请选择部门"
      apiBaseURL="http://localhost:8080"
    />
  );
}
```

### 4. OrganizationTable 组件  
```typescript
function OrgManagement() {
  const handleRowClick = (org: OrganizationUnit) => {
    console.log('选中组织:', org);
  };

  return (
    <OrganizationTable
      filter={{ status: 'ACTIVE' }}
      onRowClick={handleRowClick}
      apiBaseURL="http://localhost:8080"
    />
  );
}
```

## 🎯 特性说明

### ✅ 7位编码支持
- 自动验证7位数字编码格式
- 编码范围: 1000000-9999999
- 前端显示友好的编码格式

### ⚡ 高性能设计
- React Hook优化状态管理
- 智能缓存和错误处理
- 支持分页和过滤

### 🔧 完整功能
- **列表查询**: 支持类型和状态过滤
- **单个查询**: 通过7位编码精确查询
- **统计信息**: 实时统计数据展示
- **健康检查**: API服务状态监控

### 🎨 UI组件
- **选择器**: 下拉选择组织单元
- **表格**: 完整的组织数据展示
- **样式**: 内置样式，可自定义

## 📊 数据格式

### OrganizationUnit 类型
```typescript
interface OrganizationUnit {
  code: string;              // 7位编码
  name: string;              // 组织名称
  unit_type: string;         // 组织类型
  status: string;            // 状态
  level: number;             // 层级
  path: string;              // 路径
  sort_order: number;        // 排序
  parent_code?: string;      // 父级编码
  description?: string;      // 描述
  created_at: string;        // 创建时间
  updated_at: string;        // 更新时间
}
```

## 🚀 集成示例

### 完整应用示例
```typescript
import React from 'react';
import { OrganizationTable, OrganizationSelector } from './OrganizationComponents';

function App() {
  return (
    <div className="app">
      <h1>组织管理系统</h1>
      
      <div className="section">
        <h2>组织选择器</h2>
        <OrganizationSelector
          onSelect={(org) => console.log('选择:', org)}
          filter={{ status: 'ACTIVE' }}
        />
      </div>

      <div className="section">
        <h2>组织列表</h2>
        <OrganizationTable
          onRowClick={(org) => alert(`点击: ${org.name}`)}
        />
      </div>
    </div>
  );
}

export default App;
```

## 🔧 配置说明

### API基础URL配置
```typescript
// 开发环境
const api = new OrganizationAPI('http://localhost:8080');

// 生产环境  
const api = new OrganizationAPI('https://api.company.com');

// 使用环境变量
const api = new OrganizationAPI(process.env.REACT_APP_API_URL);
```

### 错误处理
```typescript
const { error } = useOrganizationUnits();

if (error) {
  // 处理错误
  console.error('API错误:', error);
}
```

## 📞 技术支持

- **API文档**: `/docs/api-docs/README.md`
- **性能指标**: `/docs/api-docs/METRICS.md`
- **示例代码**: `/frontend-test.html`

---

> 🎉 **7位编码组织单元前端组件已就绪！**  
> 支持React生态系统，提供完整的API集成方案