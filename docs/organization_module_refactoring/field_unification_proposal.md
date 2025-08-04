# 字段统一设计方案 - 业务ID与UUID分离

**方案编号**: FIELD-UNIFY-2025-08-04  
**提案日期**: 2025年8月4日  
**提案人员**: Claude Code Assistant  
**优先级**: 高  

## 问题分析

### 当前字段使用状况

通过分析现有API文档和实际数据结构，发现以下问题：

#### 1. 员工管理模块
**当前结构**:
```json
{
  "id": "e60891dc-7d20-444b-9002-22419238d499",  // UUID作为主键
  "tenant_id": "00000000-0000-0000-0000-000000000000",
  "first_name": "Test",
  "last_name": "Employee", 
  "employee_type": "FULL_TIME"
}
```

**问题**: 缺少人类可读的员工编号（如EMP001）

#### 2. 组织管理模块
**当前结构**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",  // UUID作为主键
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  "unit_type": "DEPARTMENT",
  "name": "技术部"
}
```

**问题**: 同样缺少人类可读的组织编码（如DEPT001）

#### 3. 字段映射混乱
- Neo4j中使用`id`字段存储UUID
- 某些查询使用`employee_id`字段  
- 缺乏统一的字段命名规范

## 统一字段设计方案

### 核心设计原则

1. **双字段系统**: 
   - `uuid`: 系统内部使用的UUID，保证全局唯一性
   - `id`: 业务层面的人类可读ID，便于用户识别和操作

2. **统一命名规范**:
   - 所有实体都采用相同的字段模式
   - API兼容性考虑向后兼容

3. **数据库层面支持**:
   - PostgreSQL和Neo4j都支持双字段存储
   - 建立适当的索引和约束

### 建议的统一字段结构

#### 员工实体 (Employee)
```json
{
  // 系统字段
  "uuid": "e60891dc-7d20-444b-9002-22419238d499",  // 系统UUID
  "id": "EMP001",                                   // 业务ID
  "tenant_id": "00000000-0000-0000-0000-000000000000",
  
  // 业务字段  
  "first_name": "张",
  "last_name": "三",
  "email": "zhangsan@company.com",
  "employee_type": "FULL_TIME",
  "employment_status": "ACTIVE",
  "hire_date": "2020-01-15T00:00:00Z",
  
  // 关联字段
  "department_id": "DEPT001",          // 业务ID引用
  "position_id": "POS001",             // 业务ID引用
  "manager_id": "EMP002",              // 业务ID引用
  
  // 系统字段
  "created_at": "2025-08-04T08:05:20.187517545+08:00",
  "updated_at": "2025-08-04T08:05:20.187517585+08:00"
}
```

#### 组织实体 (Organization)
```json
{
  // 系统字段
  "uuid": "550e8400-e29b-41d4-a716-446655440000",  // 系统UUID
  "id": "DEPT001",                                  // 业务ID
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  
  // 业务字段
  "unit_type": "DEPARTMENT", 
  "name": "技术部",
  "description": "负责技术开发和维护",
  "status": "ACTIVE",
  
  // 关联字段
  "parent_id": "COMP001",              // 业务ID引用
  "manager_id": "EMP001",              // 业务ID引用
  
  // 计算字段
  "level": 1,
  "employee_count": 15,
  
  // 系统字段
  "created_at": "2024-01-01T00:00:00.000Z",
  "updated_at": "2024-01-01T00:00:00.000Z"
}
```

#### 职位实体 (Position)
```json
{
  // 系统字段
  "uuid": "pos-uuid-here",             // 系统UUID
  "id": "POS001",                      // 业务ID
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  
  // 业务字段
  "title": "高级软件工程师", 
  "job_level": "P6",
  "position_type": "TECHNICAL",
  "status": "ACTIVE",
  
  // 关联字段
  "department_id": "DEPT001",          // 业务ID引用
  
  // 容量管理
  "max_capacity": 5,
  "current_count": 3,
  
  // 系统字段
  "created_at": "2024-01-01T00:00:00.000Z",
  "updated_at": "2024-01-01T00:00:00.000Z"
}
```

### 业务ID生成规则

#### ID前缀规范
- **EMP**: 员工 (Employee) - EMP001, EMP002, ...
- **DEPT**: 部门 (Department) - DEPT001, DEPT002, ...
- **COMP**: 公司 (Company) - COMP001
- **POS**: 职位 (Position) - POS001, POS002, ...
- **TEAM**: 团队 (Team) - TEAM001, TEAM002, ...

#### 生成策略
1. **自动递增**: 基于租户的自动递增编号
2. **格式化**: 前缀 + 零填充数字 (如: EMP001, DEPT0001)
3. **唯一性检查**: 租户内唯一性约束
4. **可配置**: 允许租户自定义前缀和格式

## API文档调整方案

### 需要更新的API端点

#### 1. 员工相关API
```http
GET /api/v1/queries/employees
```
**新增查询参数**:
- `search_by`: 支持按`uuid`或`id`搜索
- `include_uuid`: 是否在响应中包含UUID字段

**响应格式调整**:
```json
{
  "employees": [
    {
      "uuid": "e60891dc-7d20-444b-9002-22419238d499",  // 新增
      "id": "EMP001",                                   // 新增 
      "tenant_id": "00000000-0000-0000-0000-000000000000",
      "first_name": "张",
      "last_name": "三",
      // ... 其他字段保持不变
    }
  ]
}
```

#### 2. 组织相关API
```http
GET /api/v1/corehr/organizations
```
**响应格式调整**:
```json
{
  "organizations": [
    {
      "uuid": "550e8400-e29b-41d4-a716-446655440000",  // 新增
      "id": "DEPT001",                                  // 新增
      "tenant_id": "00000000-0000-0000-0000-000000000001", 
      "unit_type": "DEPARTMENT",
      "name": "技术部",
      // ... 其他字段保持不变
    }
  ]
}
```

#### 3. 新增业务ID管理端点
```http
POST /api/v1/commands/generate-business-id
```
**请求体**:
```json
{
  "entity_type": "employee",  // employee, organization, position
  "tenant_id": "tenant-uuid",
  "custom_prefix": "EMP"      // 可选，自定义前缀
}
```

**响应**:
```json
{
  "business_id": "EMP001",
  "entity_type": "employee",
  "generated_at": "2025-08-04T08:00:00Z"
}
```

### API版本兼容性

#### 向后兼容策略
1. **渐进式迁移**: 新字段作为可选字段添加
2. **双字段支持**: API同时支持UUID和业务ID查询
3. **响应格式**: 默认包含业务ID，UUID为可选
4. **查询参数**: 支持通过`?include_uuid=true`包含UUID

#### 迁移时间表
- **Phase 1** (1-2周): 添加业务ID字段，保持UUID作为主键
- **Phase 2** (2-3周): 更新前端代码使用业务ID
- **Phase 3** (4周后): 逐步迁移查询逻辑优先使用业务ID

## 数据库设计调整

### PostgreSQL表结构调整

#### 员工表 (employees)
```sql
ALTER TABLE employees 
  ADD COLUMN business_id VARCHAR(20) UNIQUE NOT NULL;
  
CREATE INDEX idx_employees_business_id ON employees(tenant_id, business_id);
CREATE INDEX idx_employees_uuid ON employees(id);  -- 保持UUID索引

-- 添加约束
ALTER TABLE employees 
  ADD CONSTRAINT uk_employees_business_id_tenant 
  UNIQUE (tenant_id, business_id);
```

#### 组织表 (organization_units)
```sql
ALTER TABLE organization_units 
  ADD COLUMN business_id VARCHAR(20) UNIQUE NOT NULL;
  
CREATE INDEX idx_org_business_id ON organization_units(tenant_id, business_id);
CREATE INDEX idx_org_uuid ON organization_units(id);

-- 添加约束
ALTER TABLE organization_units 
  ADD CONSTRAINT uk_org_business_id_tenant 
  UNIQUE (tenant_id, business_id);
```

#### 关联字段调整
```sql
-- 添加业务ID关联字段
ALTER TABLE employees 
  ADD COLUMN department_business_id VARCHAR(20),
  ADD COLUMN position_business_id VARCHAR(20),
  ADD COLUMN manager_business_id VARCHAR(20);

-- 创建外键约束
ALTER TABLE employees 
  ADD CONSTRAINT fk_employees_department_business 
  FOREIGN KEY (tenant_id, department_business_id) 
  REFERENCES organization_units(tenant_id, business_id);
```

### Neo4j结构调整

#### 节点属性更新
```cypher
// 为现有员工节点添加业务ID
MATCH (e:Employee)
WHERE e.id IS NOT NULL 
SET e.business_id = 'EMP' + RIGHT('000' + toString(e.sequence_number), 3),
    e.uuid = e.id;

// 为现有组织节点添加业务ID  
MATCH (o:Organization)
WHERE o.id IS NOT NULL
SET o.business_id = 'DEPT' + RIGHT('000' + toString(o.sequence_number), 3),
    o.uuid = o.id;
```

#### 查询优化
```cypher
// 创建业务ID索引
CREATE INDEX employee_business_id FOR (e:Employee) ON (e.tenant_id, e.business_id);
CREATE INDEX organization_business_id FOR (o:Organization) ON (o.tenant_id, o.business_id);
```

## 实施计划

### Phase 1: 数据库结构调整 (1周)
- [ ] 修改PostgreSQL表结构，添加业务ID字段
- [ ] 为现有数据生成业务ID  
- [ ] 更新Neo4j节点属性
- [ ] 创建必要的索引和约束
- [ ] 验证数据完整性

### Phase 2: 后端API调整 (1-2周)
- [ ] 更新Go结构体定义
- [ ] 修改CQRS命令和查询处理器
- [ ] 更新仓储层查询逻辑
- [ ] 添加业务ID生成服务
- [ ] 更新API响应格式
- [ ] 编写兼容性测试

### Phase 3: 前端适配 (1周)
- [ ] 更新TypeScript接口定义
- [ ] 修改API客户端调用
- [ ] 更新UI显示逻辑
- [ ] 添加业务ID搜索功能
- [ ] 验证用户体验

### Phase 4: 文档与测试 (0.5周)
- [ ] 更新API文档
- [ ] 添加集成测试
- [ ] 性能测试和优化
- [ ] 用户培训文档

## 风险评估与缓解

### 技术风险
1. **数据迁移风险**: 
   - 缓解: 分批迁移，保持回滚能力
   - 备份策略: 完整数据库备份
   
2. **性能影响**: 
   - 缓解: 预建索引，查询优化
   - 监控: API响应时间监控

3. **兼容性问题**:
   - 缓解: 渐进式部署，双字段支持
   - 测试: 全面的回归测试

### 业务风险
1. **用户适应性**:
   - 缓解: 渐进式UI变更，用户培训
   - 支持: 提供新旧ID对照表

2. **数据一致性**:
   - 缓解: 事务性操作，一致性检查
   - 监控: 自动化数据校验

## 预期收益

### 用户体验改善
- ✅ 更直观的员工和组织识别
- ✅ 更友好的搜索和筛选功能  
- ✅ 更清晰的关联关系展示

### 系统维护改善  
- ✅ 统一的字段命名规范
- ✅ 减少字段映射混乱
- ✅ 更好的数据可读性

### 性能优化
- ✅ 优化的索引策略
- ✅ 更高效的查询性能
- ✅ 减少UUID比较开销

---

**方案状态**: 等待用户确认  
**预计实施周期**: 3-4周  
**技术复杂度**: 中等  
**业务影响**: 正面，提升用户体验