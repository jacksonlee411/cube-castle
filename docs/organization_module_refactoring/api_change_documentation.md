# API变更文档 - 业务ID系统实施

**文档版本**: 1.0  
**变更日期**: 2025年8月4日  
**影响版本**: API v1.0 → v1.1  
**变更类型**: 破坏性变更 + 功能增强  

## 变更概述

本次变更将Cube Castle CoreHR API从UUID主导的系统迁移到业务ID主导的系统，旨在提高用户体验和API可用性。

### 主要变更目标
- 🎯 **用户友好**: 使用简单数字ID替代复杂UUID
- 🔄 **向后兼容**: 保留UUID支持以维护现有集成
- 📊 **性能优化**: 改善查询性能和缓存效率
- 🛡️ **安全增强**: 实施业务ID验证和错误处理

## 破坏性变更 (Breaking Changes)

### 1. 主键字段变更

#### 员工管理 API

**变更前 (v1.0)**:
```json
{
  "id": "e60891dc-7d20-444b-9002-22419238d499",
  "employee_number": "EMP001",
  "first_name": "张",
  "last_name": "三"
}
```

**变更后 (v1.1)**:
```json
{
  "id": "1",                                       // ✨ 业务ID作为主键
  "uuid": "e60891dc-7d20-444b-9002-22419238d499", // 🔒 UUID仅在请求时包含
  "first_name": "张",
  "last_name": "三"
}
```

**影响分析**:
- ❌ **破坏性**: `id`字段数据类型从UUID变为数字字符串
- ❌ **破坏性**: 移除`employee_number`字段
- ⚠️ **警告**: 现有基于UUID的查询需要更新

#### 组织管理 API

**变更前 (v1.0)**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "code": "TECH",
  "name": "技术部",
  "status": "active"
}
```

**变更后 (v1.1)**:
```json
{
  "id": "100000",                                  // ✨ 业务ID作为主键
  "uuid": "550e8400-e29b-41d4-a716-446655440000", // 🔒 UUID仅在请求时包含
  "name": "技术部",
  "unit_type": "DEPARTMENT",                       // ✨ 新增字段
  "status": "ACTIVE"                               // 🔄 枚举值变更
}
```

**影响分析**:
- ❌ **破坏性**: `id`字段数据类型从UUID变为数字字符串
- ❌ **破坏性**: 移除`code`字段
- ❌ **破坏性**: `status`枚举值变更 (`active` → `ACTIVE`)
- ✅ **新增**: `unit_type`字段提供更精确的组织类型

### 2. URL路径参数变更

#### 员工详情查询

**变更前**:
```
GET /api/v1/corehr/employees/{employee_id}
# employee_id: UUID格式
```

**变更后**:
```
GET /api/v1/corehr/employees/{employee_id}
# employee_id: 业务ID格式 (1-99999999)
```

**向后兼容方案**:
```
GET /api/v1/corehr/employees/{employee_id}?uuid_lookup=true
# 仍支持UUID查询，但需要显式启用
```

#### 组织详情查询

**变更前**:
```
GET /api/v1/corehr/organizations/{organization_id}
# organization_id: UUID格式
```

**变更后**:
```
GET /api/v1/corehr/organizations/{organization_id}
# organization_id: 业务ID格式 (100000-999999)
```

### 3. 关联字段变更

**所有关联字段从UUID改为业务ID**:

```json
// 变更前
{
  "position_id": "a4c2f6d8-2b1a-4c8d-9b7e-0e1f7a2d8c3b",
  "organization_id": "550e8400-e29b-41d4-a716-446655440000",
  "manager_id": "e60891dc-7d20-444b-9002-22419238d499"
}

// 变更后
{
  "position_id": "1000000",      // 职位业务ID
  "organization_id": "100000",   // 组织业务ID
  "manager_id": "2"              // 员工业务ID
}
```

## 非破坏性变更 (Non-Breaking Changes)

### 1. 新增字段

#### UUID包含机制
```
GET /api/v1/corehr/employees/1?include_uuid=true
```

**响应示例**:
```json
{
  "id": "1",
  "uuid": "e60891dc-7d20-444b-9002-22419238d499", // ✨ 可选包含UUID
  "first_name": "张",
  "last_name": "三"
}
```

#### 组织层级计算字段
```json
{
  "id": "100000",
  "name": "技术部",
  "employee_count": 15,          // ✨ 自动计算员工数量
  "unit_type": "DEPARTMENT"      // ✨ 明确的组织类型
}
```

### 2. 增强的错误处理

#### 业务ID格式验证

**请求示例**:
```
GET /api/v1/corehr/employees/invalid_id
```

**错误响应**:
```json
{
  "error": "VALIDATION_ERROR",
  "message": "Invalid business ID format",
  "details": {
    "field": "employee_id",
    "expected_format": "1-99999999 (string format)",
    "provided_value": "invalid_id"
  },
  "validation_errors": [
    {
      "field": "employee_id",
      "message": "Must be a string representation of number between 1-99999999",
      "code": "INVALID_BUSINESS_ID_FORMAT"
    }
  ],
  "timestamp": "2025-08-04T10:30:00Z",
  "request_id": "req_12345"
}
```

### 3. 新增业务ID管理端点

#### 业务ID生成API
```
POST /api/v1/corehr/business-ids/generate
Content-Type: application/json

{
  "entity_type": "employee",
  "count": 1
}
```

**响应**:
```json
{
  "generated_ids": ["12345"],
  "entity_type": "employee",
  "range": {
    "min": 1,
    "max": 99999999
  }
}
```

## 业务ID规则和约束

### ID范围定义

| 实体类型 | ID范围 | 格式规则 | 示例 |
|---------|--------|----------|------|
| 员工 (Employee) | 1-99999999 | 数字字符串，无前导零 | "1", "12345" |
| 组织 (Organization) | 100000-999999 | 6位数字字符串 | "100000", "123456" |
| 职位 (Position) | 1000000-9999999 | 7位数字字符串 | "1000000", "1234567" |

### 验证规则

#### 员工ID验证
```regex
^[1-9][0-9]{0,7}$
```
- 不能以0开头
- 长度1-8位
- 仅包含数字

#### 组织ID验证
```regex
^[1-9][0-9]{5}$
```
- 不能以0开头
- 固定6位长度
- 仅包含数字

#### 职位ID验证
```regex
^[1-9][0-9]{6}$
```
- 不能以0开头
- 固定7位长度
- 仅包含数字

## 迁移策略

### 阶段1: 数据准备 (1-2天)
1. **数据库更新**: 为所有现有记录生成业务ID
2. **索引创建**: 在业务ID字段上创建索引
3. **约束添加**: 实施业务ID唯一性约束

### 阶段2: API更新 (3-5天)
1. **端点修改**: 更新所有CRUD端点支持业务ID
2. **验证实施**: 添加业务ID格式验证
3. **错误处理**: 实现标准化错误响应

### 阶段3: 兼容性支持 (2-3天)
1. **UUID查询**: 保留UUID查询支持（通过查询参数）
2. **响应格式**: 实现可选UUID包含机制
3. **文档更新**: 更新API文档和示例

### 阶段4: 测试验证 (2-3天)
1. **单元测试**: 验证所有业务ID相关功能
2. **集成测试**: 测试端到端工作流
3. **性能测试**: 验证查询性能改善

## 客户端迁移指南

### 立即需要的变更

#### 1. 更新数据模型
```javascript
// 变更前
interface Employee {
  id: string; // UUID
  employee_number: string;
  first_name: string;
  last_name: string;
}

// 变更后
interface Employee {
  id: string; // 业务ID (数字字符串)
  uuid?: string; // 可选UUID
  first_name: string;
  last_name: string;
}
```

#### 2. 更新API调用
```javascript
// 变更前
const employee = await fetch(`/api/v1/corehr/employees/${uuid}`);

// 变更后
const employee = await fetch(`/api/v1/corehr/employees/${businessId}`);

// 或者保持UUID兼容性（临时方案）
const employee = await fetch(`/api/v1/corehr/employees/${uuid}?uuid_lookup=true`);
```

#### 3. 更新表单和UI
```javascript
// 变更前：显示复杂UUID
<span>员工ID: {employee.id}</span>

// 变更后：显示简洁业务ID
<span>员工编号: {employee.id}</span>
```

### 推荐的迁移步骤

1. **第1周**: 更新数据模型和类型定义
2. **第2周**: 修改API调用逻辑，使用业务ID
3. **第3周**: 更新UI显示和用户交互
4. **第4周**: 移除UUID依赖，完成迁移

## 风险评估和缓解措施

### 高风险项
- **数据一致性**: 业务ID生成可能出现冲突
  - 缓解: 实施严格的ID生成算法和数据库约束
- **现有集成中断**: 第三方系统可能依赖UUID
  - 缓解: 保留UUID兼容查询6个月以上

### 中等风险项
- **性能影响**: 新的查询模式可能影响性能
  - 缓解: 预先创建适当索引，进行性能测试
- **用户培训需求**: 用户需要适应新的ID格式
  - 缓解: 提供详细文档和培训材料

### 低风险项
- **数据迁移复杂性**: 一次性数据转换
  - 缓解: 充分测试迁移脚本，设置回滚方案

## 回滚计划

### 紧急回滚 (< 4小时)
1. 恢复API端点到UUID模式
2. 重置数据库架构到迁移前状态
3. 通知所有相关团队和客户

### 计划性回滚 (1-2天)
1. 数据清理和一致性检查
2. 逐步恢复原有API行为
3. 更新文档和通知

## 成功指标

### 技术指标
- API响应时间改善 ≥ 20%
- 数据库查询性能提升 ≥ 30%
- 客户端错误率降低 ≥ 50%

### 用户体验指标
- 用户界面操作效率提升 ≥ 40%
- 支持请求减少 ≥ 60%
- 用户满意度提升 ≥ 25%

### 运维指标
- 系统监控告警减少 ≥ 30%
- 日志分析效率提升 ≥ 50%
- 问题定位时间缩短 ≥ 40%

## 联系信息

### 技术支持
- **主要联系人**: 技术架构团队
- **邮箱**: architecture@cubecastle.com
- **文档更新**: 本文档将持续更新至迁移完成

### 反馈渠道
- **GitHub Issues**: cube-castle/api-feedback
- **内部Slack**: #api-changes
- **邮件列表**: api-updates@cubecastle.com

---

**文档状态**: 最终版本  
**最后更新**: 2025年8月4日  
**下次审核**: 2025年8月11日