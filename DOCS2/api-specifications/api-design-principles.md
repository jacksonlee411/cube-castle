# API设计原则与标准

**版本**: v1.1  
**创建日期**: 2025-08-04  
**更新日期**: 2025-08-06  
**适用范围**: Cube Castle项目所有API  
**状态**: 标准实施中 + 前端集成优化

## 🎯 总体原则

### 1. 一致性原则
- 所有API遵循相同的设计模式
- 统一的命名约定和数据格式
- 标准化的错误处理和响应结构
- **前后端数据格式统一** 🆕

### 2. 简洁性原则
- API设计简单直观，易于理解和使用
- 避免不必要的复杂性和冗余
- 优先选择简单有效的解决方案
- **前端组件友好的数据结构** 🆕

### 3. 可扩展性原则
- 设计支持未来功能扩展
- 向后兼容性保证
- 模块化和松耦合架构
- **前端状态管理兼容** 🆕

### 4. 性能优先原则
- 响应时间和吞吐量优化
- 合理的缓存策略
- 资源使用效率最大化
- **前端渲染优化支持** 🆕

## 🏗️ RESTful设计标准

### HTTP方法使用规范
```yaml
GET:    查询资源，无副作用，幂等
POST:   创建资源，有副作用，非幂等
PUT:    完整更新资源，有副作用，幂等
PATCH:  部分更新资源，有副作用，幂等
DELETE: 删除资源，有副作用，幂等
```

### 资源路径设计 (优化版)
```yaml
正确设计 (基于编码系统，Person Name版更新):
  - /api/v1/employees              # 8位编码员工集合 ✅ 新增
  - /api/v1/employees/{employee_code}    # 特定员工 (8位编码) ✅ 新增
  - /api/v1/employees/{employee_code}/positions  # 员工的职位分配 ✅ 新增
  - /api/v1/positions              # 7位编码职位集合
  - /api/v1/positions/{code}       # 特定职位 (7位编码)
  - /api/v1/positions/{code}/incumbents  # 职位的在职员工
  - /api/v1/organization-units     # 7位编码组织单元
  - /api/v1/organization-units/{code}    # 特定组织 (7位编码)

编码路径示例:
  - /api/v1/employees/10000001     # 8位员工编码查询 ✅ 新增
  - /api/v1/positions/1000001      # 7位职位编码查询
  - /api/v1/organization-units/1000000  # 7位组织编码查询

性能优化路径:
  - 直接编码查询，无转换开销
  - 数字字符串路径参数，高效解析
  - RESTful语义清晰，支持缓存

统一命名规范 (Person Name版):
  - 员工路径参数: {employee_code} (8位)
  - 职位路径参数: {code} (7位) ← 保持原规定
  - 组织路径参数: {code} (7位) ← 保持原规定

编码冲突处理:
  - 通过API路径区分实体类型 (/employees vs /positions vs /organization-units)
  - 上下文明确，避免7位编码冲突混淆
  - 字段名称明确标识 (employee_code vs position_code vs organization_code)

错误设计 (应避免):
  - /api/v1/getEmployees           # 动词不应出现在路径中
  - /api/v1/employee               # 应使用复数形式
  - /api/v1/employees-list         # 避免冗余描述
  - /api/v1/employees/uuid-format  # 避免复杂UUID路径
```

### HTTP状态码标准
```yaml
成功响应:
  200: OK - 成功获取资源
  201: Created - 成功创建资源
  204: No Content - 成功执行但无返回内容

客户端错误:
  400: Bad Request - 请求参数错误
  401: Unauthorized - 未认证
  403: Forbidden - 权限不足
  404: Not Found - 资源不存在
  409: Conflict - 资源冲突
  422: Unprocessable Entity - 数据验证失败

服务器错误:
  500: Internal Server Error - 服务器内部错误
  502: Bad Gateway - 网关错误
  503: Service Unavailable - 服务不可用
```

## 🎨 前端集成优化 🆕

### Vite + Canvas Kit 集成标准

#### API响应适配Canvas Kit组件
```typescript
// 组织统计数据响应格式 - 适配Canvas Kit组件
{
  "by_type": {
    "COMPANY": 1,
    "DEPARTMENT": 8,
    "TEAM": 17
  },
  "by_status": {
    "ACTIVE": 25,
    "INACTIVE": 1
  },
  "total_count": 26
}

// 前端Canvas Kit组件使用
const StatsCard: React.FC<{ title: string; stats: Record<string, number> }> = ({ title, stats }) => {
  return (
    <Card height="100%">
      <Card.Heading>{title}</Card.Heading>
      <Card.Body>
        {Object.entries(stats).map(([key, value]) => (
          <Box key={key} paddingY="xs">
            <Text>{key}: {value}</Text>
          </Box>
        ))}
      </Card.Body>
    </Card>
  );
};
```

#### React Query状态管理集成
```typescript
// API客户端类型安全集成
interface OrganizationStatsResponse {
  by_type: Record<string, number>;
  by_status: Record<string, number>;
  total_count: number;
}

// React Query Hook
export const useOrganizationStats = () => {
  return useQuery<OrganizationStatsResponse>({
    queryKey: ['organization', 'stats'],
    queryFn: () => organizationApi.getStats(),
    staleTime: 5 * 60 * 1000, // 5分钟缓存
  });
};
```

#### TypeScript类型定义标准
```typescript
// 共享类型定义 - 前后端一致
export interface OrganizationUnit {
  code: string;                    // 7位组织编码
  name: string;                    // 组织名称
  unit_type: 'COMPANY' | 'DEPARTMENT' | 'TEAM';
  status: 'ACTIVE' | 'INACTIVE' | 'PLANNED';
  parent_code?: string;            // 父组织编码
  level: number;                   // 组织层级
  tenant_id: string;               // 租户ID
  created_at: string;              // ISO时间戳
  updated_at: string;              // ISO时间戳
}

// 列表响应类型
export interface OrganizationListResponse {
  organizations: OrganizationUnit[];
  total_count: number;
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}
```

### 前端性能优化集成

#### 分页加载优化
```yaml
前端分页策略:
  - 默认页面大小: 20 (适配Canvas Kit Table性能)
  - 虚拟滚动: 大数据集支持 (>100条记录)
  - 预加载: 下一页数据预取
  - 缓存策略: 5分钟本地缓存

API响应格式:
  data: []              # 当前页数据
  pagination:
    page: 1             # 当前页码
    page_size: 20       # 每页大小
    total: 156          # 总记录数
    total_pages: 8      # 总页数
    has_next: true      # 是否有下一页
    has_prev: false     # 是否有上一页
```

#### 实时数据同步
```typescript
// WebSocket集成标准
interface RealtimeEvent {
  event_type: 'organization_created' | 'organization_updated' | 'organization_deleted';
  entity_type: 'organization_unit';
  entity_code: string;
  data: OrganizationUnit | null;
  timestamp: string;
}

// React Query实时更新
const useRealtimeOrganizations = () => {
  const queryClient = useQueryClient();
  
  useEffect(() => {
    const eventSource = new EventSource('/api/v1/organizations/events');
    
    eventSource.onmessage = (event) => {
      const data: RealtimeEvent = JSON.parse(event.data);
      
      // 更新本地缓存
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
    };
  }, [queryClient]);
};
```

## 📊 数据格式标准

### 请求格式 (优化版)
```json
// POST /api/v1/employees - 8位编码员工创建 (Person Name版)
{
  "organization_code": "1000000",       // 7位组织编码 (必需)
  "primary_position_code": "1000001",   // 7位主要职位编码 (可选)
  "employee_type": "FULL_TIME",         // 员工类型
  "employment_status": "ACTIVE",        // 就业状态 (默认)
  
  // Person Name 简化字段组
  "person_name": "张三",                // 完整姓名 (必填)
  "first_name": "张",                   // 姓 (可选)
  "last_name": "三",                    // 名 (可选)
  
  "email": "zhang.san@company.com",     // 工作邮箱 (必需)
  "personal_email": "zhang.san@gmail.com", // 个人邮箱 (可选)
  "phone_number": "13800138000",        // 手机号码 (可选)
  "hire_date": "2025-08-05",            // 入职日期 (必需)
  
  "personal_info": {                    // 个人信息 (可选)
    "age": 28,
    "gender": "M",
    "address": "北京市朝阳区"
  },
  "employee_details": {                 // 员工详情 (可选)
    "title": "高级软件工程师",
    "level": "P6",
    "salary": 25000
  }
}

// POST /api/v1/positions - 7位编码职位创建 (保持原规定)
{
  "organization_code": "1000000",       // 7位组织编码 (必需)
  "manager_position_code": "1000001",   // 7位管理职位编码 (可选)
  "position_type": "FULL_TIME",         // 优化的职位类型
  "job_profile_id": "uuid",             // 保留UUID用于外部集成
  "status": "OPEN",                     // 默认状态
  "budgeted_fte": 1.0,                  // 预算FTE
  "details": {                          // 多态配置
    "title": "高级软件工程师",
    "salary_range": {
      "min": 60000,
      "max": 90000,
      "currency": "CNY"
    },
    "benefits": ["health_insurance", "annual_leave"],
    "work_schedule": "9_to_5",
    "remote_allowed": true
  }
}

// PUT /api/v1/employees/{employee_code} - 员工更新 (8位编码) 
{
  "employment_status": "ON_LEAVE",      // 状态更新
  "person_name": "张三（更新）",        // 完整姓名更新
  "phone_number": "13800138888",        // 联系方式更新
  "employee_details": {                 // 部分更新支持
    "title": "资深软件工程师",
    "level": "P7",
    "salary": 30000
  }
}
```

### 响应格式 (优化版)
```json
// 员工单个资源响应 - 8位编码系统 (Person Name版)
{
  "employee_code": "10000001",          // 8位员工编码
  "organization_code": "1000000",       // 7位组织编码
  "primary_position_code": "1000001",   // 7位主要职位编码
  "employee_type": "FULL_TIME",
  "employment_status": "ACTIVE",
  
  // Person Name 简化字段组
  "person_name": "张三",                // 完整姓名 (主要显示)
  "first_name": "张",                   // 姓 (可选)
  "last_name": "三",                    // 名 (可选)
  
  "email": "zhang.san@company.com",
  "personal_email": "zhang.san@gmail.com",
  "phone_number": "13800138000",
  "hire_date": "2025-08-05",
  
  "personal_info": "{\"age\": 28, \"gender\": \"M\"}",     // JSON string
  "employee_details": "{\"title\": \"高级软件工程师\", \"level\": \"P6\"}", // JSON string
  
  "tenant_id": "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9",
  "created_at": "2025-08-05T00:00:00Z",
  "updated_at": "2025-08-05T00:00:00Z"
}

// 职位单个资源响应 - 7位编码系统 (保持原规定)
{
  "code": "1000001",                    // 7位职位编码
  "organization_code": "1000000",       // 7位组织编码
  "manager_position_code": "1000002",   // 7位管理职位编码
  "position_type": "FULL_TIME",
  "job_profile_id": "uuid",             // 外部系统集成保留UUID
  "status": "OPEN",
  "budgeted_fte": 1.0,
  "details": {
    "title": "高级软件工程师",
    "salary_range": {
      "min": 60000,
      "max": 90000,
      "currency": "CNY"
    }
  },
  "created_at": "2025-08-05T00:00:00Z",
  "updated_at": "2025-08-05T00:00:00Z"
}

// 集合资源响应 - 统一格式
{
  "employees": [                        // 员工实体名称复数形式
    // 员工资源数组
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,                    // 优化默认页面大小
    "total": 150,
    "total_pages": 8
  }
}

// 错误响应 - 增强版 (员工编码验证)
{
  "error": {
    "code": "INVALID_EMPLOYEE_CODE",
    "message": "无效的员工编码格式",
    "details": {
      "field": "employee_code",
      "value": "123",
      "constraint": "must be 8 digits (10000000-99999999)",
      "expected_format": "8-digit numeric code"
    },
    "timestamp": "2025-08-05T00:00:00Z",
    "request_id": "req_12345678"
  }
}
```

### 字段命名约定
```yaml
风格: snake_case (后端) / camelCase (前端)
时间: ISO 8601格式 (2025-08-04T00:00:00Z)
布尔值: true/false
枚举: 大写字母 (ACTIVE, INACTIVE)
标识符字段: 见标识符命名标准
```

### 标识符命名标准 ⭐ (优化版 v2.0)
```yaml
实体编码位数分配 (修订版):
  组织单元: 7位数字 (1000000-9999999) ✅ 已优化实施
  员工: 8位数字 (10000000-99999999) ✅ 已优化实施 (Person Name版)
  职位: 7位数字 (1000000-9999999) ← 保持原规定
  作业档案: 5位数字 (10000-99999) ← 保持原规定

编码冲突问题识别:
  ⚠️ 问题: 组织单元和职位都使用7位编码 (1000000-9999999)
  ⚠️ 影响: 用户无法区分 "1000001" 是组织还是职位
  ⚠️ 建议: 考虑通过前缀或上下文区分，或未来优化时调整

优化后的标识符架构建议:
  主键系统: 编码直接作为数据库主键，消除UUID转换
  查询性能: 数字索引性能优于UUID索引
  零转换: 无需业务ID↔UUID映射，简化架构
  响应速度: 基于成功经验预期40-60%性能提升

统一命名规范 (v2.0更新):
  - 主实体字段: 使用 "{entity}_code" 格式
  - 员工实体: "employee_code": "10000001" (8位) ✅ 新标准
  - 组织单元: "code": "1000001" (7位) ← 保持原规定
  - 职位实体: "code": "1000001" (7位) ← 保持原规定

关系引用:
  - 使用 "{entity}_code" 格式
  - 示例: "organization_code": "1000000" (组织关系 - 7位)
  - 示例: "position_code": "1000001" (职位关系 - 7位)
  - 示例: "primary_position_code": "1000002" (主要职位关系 - 7位)
  - 示例: "manager_position_code": "1000003" (管理职位关系 - 7位)
  - 示例: "employee_code": "10000001" (员工关系 - 8位)

内部标识符 (过渡期):
  - UUID仅在遗留系统中保留
  - 新系统完全基于编码架构
  - 逐步迁移现有UUID系统

设计原则:
  - 业务语义清晰: "编码"比"ID"更直观
  - 用户认知简单: 通过字段名和上下文区分实体类型
  - 性能优先: 直接主键查询，最佳数据库性能
  - 独立扩展: 各实体编码位数按原规定保持
  - 行业标准兼容: 符合企业级HR系统惯例

成功案例验证:
  - 7位组织编码: 已实现60%性能提升 ✅
  - 职位管理优化: 基于7位编码，预期性能提升 🔄
  - 架构一致性: 统一的编码设计模式 ✅
```

## 🔒 安全设计标准

### 认证和授权
```yaml
认证方式:
  - JWT Bearer Token
  - Header: "Authorization: Bearer <token>"

授权控制:
  - 基于角色的访问控制 (RBAC)
  - 资源级权限检查
  - 租户隔离

安全头部:
  - X-Tenant-ID: 租户标识
  - X-Request-ID: 请求追踪ID
  - X-API-Version: API版本
```

### 数据验证
```yaml
输入验证:
  - 所有输入参数必须验证
  - 使用白名单验证方法
  - 防范SQL注入和XSS攻击

输出过滤:
  - 敏感信息过滤
  - 数据脱敏处理
  - 最小权限原则
```

## ⚡ 性能设计标准

### 响应时间目标
```yaml
API类型: 目标响应时间
简单查询: < 100ms
复杂查询: < 500ms
数据创建: < 200ms
数据更新: < 200ms
数据删除: < 100ms
```

### 缓存策略
```yaml
缓存层级:
  - 应用层缓存: 5-30分钟
  - 数据库查询缓存: 1-5分钟
  - CDN缓存: 24小时

缓存策略:
  - 读多写少: 长期缓存
  - 实时性要求高: 短期缓存
  - 静态数据: 永久缓存
```

### 分页设计
```yaml
默认参数:
  - page_size: 50 (默认)
  - max_page_size: 100 (最大)
  - page: 1 (起始页)

响应格式:
  data: [] # 数据数组
  pagination:
    page: 当前页码
    page_size: 每页大小
    total: 总记录数
    total_pages: 总页数
```

## 🔄 版本管理策略

### 版本化方案
```yaml
URL版本化: /api/v1/positions (推荐)
Header版本化: X-API-Version: v1 (备选)
参数版本化: ?version=v1 (不推荐)
```

### 版本兼容性
```yaml
向后兼容原则:
  - 新增字段: 兼容
  - 删除字段: 需要版本升级
  - 修改字段类型: 需要版本升级
  - 修改字段含义: 需要版本升级

弃用策略:
  - 提前3个月通知
  - 提供迁移指南
  - 支持并行版本
```

## 🧪 测试标准

### API测试覆盖率
```yaml
单元测试: ≥90%
集成测试: ≥80%
端到端测试: 核心场景100%
性能测试: 所有API端点
安全测试: 所有API端点
```

### 测试用例设计
```yaml
正常场景:
  - 有效数据测试
  - 边界值测试
  - 典型用例测试

异常场景:
  - 无效参数测试
  - 权限验证测试
  - 资源不存在测试
  - 并发访问测试

性能场景:
  - 负载测试
  - 压力测试
  - 并发测试
```

## 📚 文档标准

### API文档要求
```yaml
必须包含:
  - 完整的端点描述
  - 请求/响应示例
  - 参数说明和验证规则
  - 错误码和处理说明
  - 认证和权限要求

可选包含:
  - SDK和代码示例
  - 集成指南
  - 最佳实践
  - 常见问题解答
```

### OpenAPI规范
```yaml
格式: OpenAPI 3.0+
工具: Swagger UI
生成: 自动从代码注解生成
维护: 与代码同步更新
```

## 🔍 监控和日志

### API监控指标
```yaml
业务指标:
  - 请求量 (QPS)
  - 响应时间 (P50, P95, P99)
  - 错误率
  - 可用性 (SLA)

技术指标:
  - CPU使用率
  - 内存使用率
  - 数据库连接数
  - 缓存命中率
```

### 日志规范
```yaml
日志级别:
  ERROR: 系统错误，需要立即处理
  WARN: 警告信息，需要关注
  INFO: 一般信息，正常业务流程
  DEBUG: 调试信息，开发环境使用

日志格式:
  timestamp: ISO 8601格式
  level: 日志级别
  message: 日志消息
  context: 上下文信息 (request_id, user_id, tenant_id)
```

## 🚨 错误处理标准

### 错误分类
```yaml
客户端错误 (4xx):
  - 请求格式错误
  - 参数验证失败
  - 权限不足
  - 资源不存在

服务端错误 (5xx):
  - 系统内部错误
  - 数据库连接失败
  - 第三方服务不可用
  - 超时错误
```

### 错误响应格式
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "用户友好的错误描述",
    "details": {
      "field": "出错的字段",
      "value": "错误的值",
      "constraint": "约束类型"
    },
    "trace_id": "请求追踪ID"
  }
}
```

## 📋 API设计检查清单

### 设计阶段
- [ ] 遵循RESTful设计原则
- [ ] 使用标准HTTP方法和状态码
- [ ] 设计合理的资源路径
- [ ] 定义清晰的数据模型
- [ ] 考虑向后兼容性

### 实现阶段
- [ ] 实现完整的输入验证
- [ ] 添加适当的错误处理
- [ ] 实施安全认证和授权
- [ ] 优化性能和响应时间
- [ ] 添加缓存策略

### 测试阶段
- [ ] 编写全面的单元测试
- [ ] 执行集成测试
- [ ] 进行性能测试
- [ ] 实施安全测试
- [ ] 验证文档准确性

### 发布阶段
- [ ] 更新API文档
- [ ] 配置监控和告警
- [ ] 准备回滚计划
- [ ] 通知相关团队
- [ ] 收集使用反馈

## 🔄 持续改进

### 定期评估
```yaml
评估频率: 每季度
评估内容:
  - API使用情况分析
  - 性能指标评估
  - 用户反馈收集
  - 安全性审查

改进措施:
  - 性能优化
  - 功能增强
  - 文档完善
  - 工具升级
```

### 团队培训
```yaml
新员工培训:
  - API设计原则
  - 开发最佳实践
  - 工具使用指南
  - 代码审查标准

持续学习:
  - 技术分享会
  - 外部培训
  - 行业最佳实践
  - 工具和技术更新
```

---

**制定者**: 系统架构师  
**审核者**: 技术委员会  
**生效日期**: 2025-08-04  
**下次审查**: 2025-11-04