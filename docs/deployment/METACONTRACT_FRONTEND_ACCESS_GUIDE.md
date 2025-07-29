# 元合约编译器前端访问指南

## 概述

元合约编译器已经被集成为城堡项目的一个轻量级内部模块。本指南说明如何从前端页面访问和使用元合约编译器功能。

## 🌐 API 端点

**基础URL**: `http://localhost:8080/api/v1/metacontract`

### 核心API端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/projects` | GET | 获取项目列表 |
| `/projects` | POST | 创建新项目 |
| `/projects/{id}` | GET | 获取特定项目 |
| `/projects/{id}` | PUT | 更新项目 |
| `/projects/{id}` | DELETE | 删除项目 |
| `/projects/{id}/compile` | POST | 编译项目 |
| `/templates` | GET | 获取模板列表 |
| `/settings` | GET | 获取用户设置 |
| `/settings` | PUT | 更新用户设置 |

## 🖥️ 前端页面访问

### 1. 主编辑器页面

**URL**: `http://localhost:3000/metacontract-editor`

- 提供完整的元合约编辑器界面
- 支持语法高亮和代码补全
- 实时编译验证
- 模板支持

### 2. 项目管理页面

**URL**: `http://localhost:3000/metacontract-editor/projects`

- 项目列表管理
- 创建、编辑、删除项目
- 项目搜索和筛选

### 3. 特定项目编辑

**URL**: `http://localhost:3000/metacontract-editor/{projectId}`

- 编辑特定项目的元合约内容
- 自动保存功能
- 编译结果预览

## 📦 API 使用示例

### 使用 REST API 客户端

```typescript
import { restApiClient } from '@/lib/rest-api-client';

// 获取项目列表
const getProjects = async () => {
  const response = await restApiClient.getProjects({ limit: 10, offset: 0 });
  if (response.success) {
    console.log('Projects:', response.data.projects);
  }
};

// 创建新项目
const createProject = async () => {
  const projectData = {
    name: "新的元合约项目",
    description: "项目描述",
    content: `resource_name: example_entity
namespace: example.namespace
version: "1.0.0"

data_structure:
  fields:
    - name: id
      type: UUID
      constraints:
        primary_key: true
        required: true`
  };

  const response = await restApiClient.createProject(projectData);
  if (response.success) {
    console.log('Created project:', response.data);
  }
};

// 编译项目
const compileProject = async (projectId: string, content: string) => {
  const response = await restApiClient.compileProject(projectId, {
    content: content,
    preview: true
  });
  
  if (response.success) {
    console.log('Compilation result:', response.data);
  }
};

// 获取模板
const getTemplates = async () => {
  const response = await restApiClient.getTemplates('basic');
  if (response.success) {
    console.log('Templates:', response.data.templates);
  }
};
```

### 使用 React Hook

```typescript
import { useMetaContractEditor } from '@/hooks/useMetaContractEditor';

const MyComponent = () => {
  const {
    projects,
    currentProject,
    isLoading,
    createProject,
    updateProject,
    compileProject,
    loadProject
  } = useMetaContractEditor();

  const handleCreateProject = async () => {
    const project = await createProject({
      name: "新项目",
      content: "// 元合约内容"
    });
    console.log('Created:', project);
  };

  const handleCompile = async () => {
    if (currentProject) {
      const result = await compileProject(currentProject.id, currentProject.content);
      console.log('Compilation result:', result);
    }
  };

  return (
    <div>
      <button onClick={handleCreateProject}>创建项目</button>
      <button onClick={handleCompile}>编译</button>
      {/* 项目列表和编辑器 */}
    </div>
  );
};
```

## 🛠️ 开发设置

### 1. 启动后端服务

```bash
cd go-app
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动

### 2. 启动前端服务

```bash
cd nextjs-app
npm run dev
```

前端将在 `http://localhost:3000` 启动

### 3. 验证API连接

访问健康检查端点：
- 后端健康检查: `http://localhost:8080/health`
- API状态: `http://localhost:8080/api/v1/metacontract/templates`

## 🔧 配置说明

### 环境变量

在 `nextjs-app/.env.local` 中配置：

```bash
NEXT_PUBLIC_API_ENDPOINT=http://localhost:8080/api/v1
```

### 认证与权限

当前实现使用Mock认证：
- 默认租户ID: `00000000-0000-0000-0000-000000000000`
- 默认用户ID: `11111111-1111-1111-1111-111111111111`

生产环境需要实现实际的JWT认证。

## 📝 元合约语法示例

### 基础实体

```yaml
resource_name: employee
namespace: hr.employees
version: "1.0.0"

data_structure:
  fields:
    - name: id
      type: UUID
      constraints:
        primary_key: true
        required: true
    
    - name: employee_id
      type: String
      constraints:
        required: true
        unique: true
        max_length: 20
    
    - name: first_name
      type: String
      constraints:
        required: true
        max_length: 50
    
    - name: email
      type: String
      constraints:
        required: true
        unique: true
        format: email

security_model:
  access_control: rbac
  data_classification: confidential

temporal_behavior:
  temporality_paradigm: snapshot
  state_transition_model: discrete
```

## 🎯 功能特性

### 1. 实时编译
- 输入元合约YAML内容
- 实时语法验证
- 编译错误提示
- 生成代码预览

### 2. 模板系统
- 预置模板库
- 分类管理
- 快速项目初始化

### 3. 项目管理
- 多项目支持
- 版本控制
- 协作编辑

### 4. 用户设置
- 编辑器主题
- 字体大小
- 自动保存
- 快捷键绑定

## 🐛 故障排除

### 常见问题

1. **API连接失败**
   - 检查后端服务是否启动
   - 验证API端点配置
   - 查看浏览器开发者工具的网络请求

2. **编译错误**
   - 检查YAML语法
   - 验证字段约束
   - 查看编译器错误信息

3. **权限问题**
   - 确认租户ID设置
   - 检查用户认证状态

### 日志查看

- 后端日志：控制台输出
- 前端错误：浏览器开发者工具
- API请求：网络面板

## 📚 进一步开发

### 扩展功能

1. **实际数据库集成**
   - 替换Mock Repository
   - 实现数据持久化

2. **用户认证**
   - JWT Token验证
   - 角色权限管理

3. **实时协作**
   - WebSocket支持
   - 多用户编辑

4. **代码生成**
   - 完善Ent生成器
   - API路由生成
   - 业务逻辑生成

### 代码结构

```
cube-castle/
├── go-app/                          # 后端服务
│   ├── cmd/server/main.go          # 主服务器
│   ├── internal/metacontract/      # 编译器核心
│   ├── internal/metacontracteditor/ # 编辑器服务
│   └── internal/codegen/           # 代码生成器
├── nextjs-app/                     # 前端应用
│   ├── src/pages/metacontract-editor/ # 编辑器页面
│   ├── src/components/metacontract-editor/ # 编辑器组件
│   ├── src/lib/rest-api-client.ts  # API客户端
│   └── src/hooks/useMetaContractEditor.ts # React Hook
```

## 🚀 生产部署

### Docker部署

```bash
# 构建和启动服务
docker-compose up -d

# 访问应用
# 前端：http://localhost:3000
# 后端：http://localhost:8080
```

### 环境配置

生产环境需要配置：
- 数据库连接
- 认证服务
- 日志收集
- 监控告警

---

**最后更新**: 2024年当前日期
**版本**: v1.0.0