# 标识符命名标准

**版本**: v1.1  
**创建日期**: 2025-08-05  
**最后更新**: 2025-08-05 (升级为7位数字)  
**适用范围**: Cube Castle项目所有实体标识符  
**状态**: 正式实施

## 📅 版本更新记录

### v1.1 (2025-08-05)
- **重要更新**: 标识符位数从6位升级到7位
- **新格式**: 1000000-9999999 (共900万个ID)
- **影响范围**: 所有实体类别的ID分配
- **向后兼容**: 支持现有6位ID的平滑升级

### v1.0 (2025-08-05)
- 初始版本，使用6位数字格式
- 确定code命名标准
- 建立UUID隐藏机制

## 🎯 标准概述

本标准定义了项目中所有实体标识符的命名规范，旨在提供一致、直观、业务友好的标识符设计，降低用户认知负担，提升系统易用性。

## 📋 核心原则

### 1. 业务语义优先
- 使用业务人员容易理解的术语
- 避免技术性过强的命名
- 符合行业惯例和用户预期

### 2. 认知简化
- 对外只暴露一种标识符类型
- 消除内部技术细节的复杂性
- 提供直观的API使用体验

### 3. 行业标准兼容
- 参考主流企业级系统的命名规范
- 符合HR行业的通用实践
- 便于与第三方系统集成

## 🏷️ 命名规范

### 主要标识符: `code`

#### 适用场景
- 所有对外API中的主要实体标识符
- 用户界面中显示的标识符
- 第三方系统集成时的引用标识符

#### 格式规范
```yaml
字段名: code
数据类型: VARCHAR(10)
格式: 7位数字
范围: 1000000-9999999
示例: "1000001", "1000023", "2345678"
```

#### 使用示例
```json
{
  "code": "1000001",
  "name": "技术部",
  "unit_type": "DEPARTMENT",
  "status": "ACTIVE"
}
```

### 关系引用标识符

#### 命名模式
使用 `{关联实体}_code` 格式来表示对其他实体的引用。

#### 格式规范
```yaml
父级关系: parent_code
部门关系: department_code  
经理关系: manager_code
职级关系: job_level_code
```

#### 使用示例
```json
{
  "code": "1000002",
  "parent_code": "1000001",    // 引用父级组织单元
  "manager_code": "2000001",   // 引用管理者
  "name": "前端开发组"
}
```

### 内部技术标识符

#### UUID使用原则
- **仅限内部使用**: UUID只在系统内部数据库和服务间使用
- **API完全隐藏**: 所有对外API不暴露UUID
- **数据库主键**: 继续使用UUID作为数据库主键
- **性能优化**: 利用UUID的唯一性和查询性能优势

#### 实现示例
```go
// 数据库模型 (内部)
type OrganizationUnit struct {
    ID           uuid.UUID `db:"id"`                    // 内部主键
    Code         string    `db:"code"`                  // 对外标识符
    ParentID     *uuid.UUID `db:"parent_id"`            // 内部关系
    // ...
}

// API响应模型 (对外)
type OrganizationResponse struct {
    Code       string  `json:"code"`                    // 只暴露code
    ParentCode *string `json:"parent_code,omitempty"`   // 关系也用code
    Name       string  `json:"name"`
    // ...
}
```

## 🎨 实体编码示例

### 组织架构类 (7位编码)
```yaml
组织单元: organization_units
  - 主标识符: code (7位)
  - 父级引用: parent_code  
  - 示例: "1000001", "1000002"
  - 范围: 1000000-9999999
```

### 人员管理类
```yaml
员工: employees (8位编码)
  - 主标识符: code
  - 部门引用: department_code (7位，引用组织单元)
  - 经理引用: manager_code (8位，引用其他员工)
  - 示例: "10000001", "10000002"
  - 范围: 10000000-99999999

职位: positions (7位编码)
  - 主标识符: code
  - 部门引用: department_code (7位，引用组织单元)
  - 示例: "1000001", "1000002"  
  - 范围: 1000000-9999999
```

### 系统配置类
```yaml
作业档案: job_profiles (5位编码)
  - 主标识符: code
  - 示例: "10001", "10002"
  - 范围: 10000-99999
```

## 🔢 实体编码位数策略

### 设计原则
- **独立设计**: 每种实体根据业务特点独立确定编码位数
- **避免耦合**: 不同实体间编码格式互不影响
- **按需分配**: 根据实体数量规模和业务需求确定位数
- **预留空间**: 为未来扩展预留充足的编码空间

### 实体编码规格

#### 组织单元 (Organization Units) - 7位编码
```yaml
位数: 7位数字
范围: 1000000-9999999
容量: 900万个ID
使用场景: 企业组织架构、部门层级
特点: 层级复杂，需要大量编码空间
```

#### 员工 (Employees) - 8位编码  
```yaml
位数: 8位数字
范围: 10000000-99999999
容量: 9000万个ID
使用场景: 员工档案管理
特点: 企业人员规模可能很大，需要充足的编码空间
```

#### 职位 (Positions) - 7位编码
```yaml
位数: 7位数字  
范围: 1000000-9999999
容量: 900万个ID
使用场景: 职位定义、岗位管理
特点: 职位种类和实例较多，需要较大编码空间
```

#### 作业档案 (Job Profiles) - 5位编码
```yaml
位数: 5位数字
范围: 10000-99999  
容量: 9万个ID
使用场景: 标准化职位描述
特点: 标准化程度高，数量相对可控
```

### 编码空间优势
- **精准分配**: 每种实体按实际需求分配编码空间
- **独立扩展**: 某个实体需要扩展时不影响其他实体
- **资源优化**: 避免编码空间浪费
- **维护简便**: 各实体编码规则独立维护

### 生成机制

#### 组织单元序列 (7位)
```sql
CREATE SEQUENCE org_code_seq 
    START WITH 1000000 
    INCREMENT BY 1 
    MAXVALUE 9999999;

CREATE OR REPLACE FUNCTION generate_org_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('org_code_seq')::text, 7, '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

#### 员工序列 (8位)
```sql
CREATE SEQUENCE employee_code_seq 
    START WITH 10000000 
    INCREMENT BY 1 
    MAXVALUE 99999999;

CREATE OR REPLACE FUNCTION generate_employee_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('employee_code_seq')::text, 8, '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

#### 职位序列 (7位)
```sql
CREATE SEQUENCE position_code_seq 
    START WITH 1000000 
    INCREMENT BY 1 
    MAXVALUE 9999999;

CREATE OR REPLACE FUNCTION generate_position_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('position_code_seq')::text, 7, '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

#### 作业档案序列 (5位)
```sql
CREATE SEQUENCE job_profile_code_seq 
    START WITH 10000 
    INCREMENT BY 1 
    MAXVALUE 99999;

CREATE OR REPLACE FUNCTION generate_job_profile_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('job_profile_code_seq')::text, 5, '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

### 并发安全
- 使用数据库序列确保并发安全
- 避免应用层生成ID的竞争条件
- 支持分布式部署环境

## 📐 API设计标准

### 端点路径
```yaml
正确设计:
  GET /api/v1/organizations/{code}
  POST /api/v1/organizations  
  PUT /api/v1/organizations/{code}
  DELETE /api/v1/organizations/{code}

错误设计:
  GET /api/v1/organizations/{id}          # 不要暴露内部ID
  GET /api/v1/organizations/{uuid}        # 不要暴露UUID
  GET /api/v1/organizations/{business_id} # 避免技术术语
```

### 查询参数
```yaml
正确设计:
  GET /api/v1/organizations?parent_code=1000001
  GET /api/v1/employees?department_code=2000001

错误设计:
  GET /api/v1/organizations?parent_id=uuid
  GET /api/v1/employees?department_business_id=2000001
```

### 请求体格式
```json
// 创建组织单元 - 正确
{
  "name": "新部门",
  "parent_code": "1000001",
  "unit_type": "DEPARTMENT"
}

// 创建组织单元 - 错误
{
  "name": "新部门", 
  "parent_business_id": "1000001",  // 避免business_id
  "parent_uuid": "uuid-string"     // 不要暴露UUID
}
```

## 🔄 数据转换机制

### API适配层
```go
// 服务层 - ID转换
type OrganizationService struct {
    repo   *OrganizationRepository
    mapper *IDMapper
}

func (s *OrganizationService) GetByCode(code string) (*Organization, error) {
    // code -> UUID 转换 (查询时)
    uuid := s.mapper.GetUUIDByCode(code)
    org := s.repo.FindByUUID(uuid)
    
    // UUID -> code 转换 (响应时)  
    return s.mapper.ToAPIResponse(org), nil
}
```

### 缓存策略
```go
// 双向映射缓存
type IDMapper struct {
    codeToUUID map[string]uuid.UUID
    uuidToCode map[uuid.UUID]string
    cache      *redis.Client
}

func (m *IDMapper) GetUUIDByCode(code string) uuid.UUID {
    // 优先从缓存获取
    if uuid, ok := m.codeToUUID[code]; ok {
        return uuid
    }
    
    // 缓存未命中，查询数据库
    uuid := m.repo.GetUUIDByCode(code)
    m.updateCache(code, uuid)
    return uuid
}
```

## 🧪 验证和测试

### 字段验证规则

#### 组织单元验证 (7位)
```yaml
code字段验证:
  - 格式: ^[0-9]{7}$ (7位数字)
  - 范围: 1000000-9999999
  - 唯一性: 全局唯一
  - 必填: true

parent_code字段验证:
  - 格式: ^[0-9]{7}$ (7位数字)  
  - 存在性: 必须引用存在的组织单元code
  - 循环检查: 防止父子关系循环
  - 必填: false (顶级节点可为空)
```

#### 员工验证 (8位)
```yaml
code字段验证:
  - 格式: ^[0-9]{8}$ (8位数字)
  - 范围: 10000000-99999999
  - 唯一性: 全局唯一
  - 必填: true

department_code字段验证:
  - 格式: ^[0-9]{7}$ (7位数字，引用组织单元)
  - 存在性: 必须引用存在的组织单元code
  - 必填: true

manager_code字段验证:
  - 格式: ^[0-9]{8}$ (8位数字，引用其他员工)
  - 存在性: 必须引用存在的员工code
  - 必填: false
```

#### 职位验证 (7位)
```yaml
code字段验证:
  - 格式: ^[0-9]{7}$ (7位数字)
  - 范围: 1000000-9999999
  - 唯一性: 全局唯一
  - 必填: true

department_code字段验证:
  - 格式: ^[0-9]{7}$ (7位数字，引用组织单元)
  - 存在性: 必须引用存在的组织单元code
  - 必填: true
```

#### 作业档案验证 (5位)
```yaml
code字段验证:
  - 格式: ^[0-9]{5}$ (5位数字)
  - 范围: 10000-99999
  - 唯一性: 全局唯一
  - 必填: true
```

### 测试用例设计
```yaml
正常场景:
  - 创建实体时自动生成code
  - 通过code查询实体
  - 通过parent_code建立关系
  - code的唯一性验证

异常场景:
  - 无效的code格式
  - 不存在的parent_code引用
  - code重复性冲突
  - 父子关系循环引用

边界场景:
  - code范围边界值测试
  - 序列耗尽处理
  - 大量并发创建测试
```

## 📊 性能考量

### 索引策略
```sql
-- 主要查询索引
CREATE UNIQUE INDEX idx_organizations_code ON organization_units(code);
CREATE INDEX idx_organizations_parent_code ON organization_units(parent_code);
CREATE INDEX idx_organizations_tenant_code ON organization_units(tenant_id, code);

-- 复合查询索引
CREATE INDEX idx_organizations_status_type ON organization_units(status, unit_type);
CREATE INDEX idx_organizations_path ON organization_units USING gin(path);
```

### 查询优化
```yaml
单表查询: 直接使用code索引，性能最优
关系查询: 通过parent_code索引，避免JOIN UUID
路径查询: 使用GIN索引支持层级路径搜索
分页查询: code字段支持稳定排序
```

## 🔍 监控和维护

### 关键指标
```yaml
业务指标:
  - code生成成功率: >99.9%
  - code查询响应时间: <50ms
  - 父子关系一致性: 100%
  - 缓存命中率: >95%

技术指标:
  - 序列使用率: <80% (告警阈值)
  - ID转换错误率: <0.1%
  - 数据库索引命中率: >98%
```

### 告警策略
```yaml
紧急告警:
  - code序列即将耗尽 (>90%)
  - code重复冲突
  - 父子关系循环检测

警告告警:
  - 缓存命中率下降 (<90%)
  - 查询响应时间超过阈值
  - ID转换失败率上升
```

## 🚀 迁移指南

### 从business_id迁移到code

#### 数据库迁移
```sql
-- 1. 添加新字段
ALTER TABLE organization_units ADD COLUMN code VARCHAR(10);

-- 2. 数据迁移 (保持原有业务ID值)
UPDATE organization_units 
SET code = business_id 
WHERE business_id IS NOT NULL;

-- 3. 添加约束
ALTER TABLE organization_units 
ADD CONSTRAINT uk_organizations_code UNIQUE (code);
ALTER TABLE organization_units 
ALTER COLUMN code SET NOT NULL;

-- 4. 删除旧字段 (确认无依赖后)
ALTER TABLE organization_units DROP COLUMN business_id;
```

#### API迁移
```yaml
第一阶段: 同时支持business_id和code
第二阶段: 标记business_id为deprecated  
第三阶段: 完全移除business_id支持
```

#### 前端迁移
```typescript
// 旧的类型定义
interface Organization {
  business_id: string;
  parent_business_id?: string;
  // ...
}

// 新的类型定义  
interface Organization {
  code: string;
  parent_code?: string;
  // ...
}
```

## 📚 最佳实践

### 开发指南
```yaml
DO (推荐做法):
  - 始终使用code作为对外标识符
  - API响应只包含code，隐藏UUID
  - 使用parent_code建立父子关系
  - 在业务逻辑中优先使用code

DON'T (避免做法):
  - 不要在API中暴露UUID
  - 不要使用business_id等技术术语
  - 不要在前端组件中处理UUID
  - 不要绕过ID映射机制直接使用UUID
```

### 代码审查清单
```yaml
设计审查:
  - [ ] 是否使用code作为主要标识符
  - [ ] 是否隐藏了内部UUID
  - [ ] 关系字段是否使用{entity}_code格式
  - [ ] API路径是否使用code参数

实现审查:
  - [ ] 是否实现了ID转换机制
  - [ ] 是否添加了适当的缓存策略
  - [ ] 是否包含必要的验证逻辑
  - [ ] 错误处理是否完整

测试审查:
  - [ ] 是否包含code格式验证测试
  - [ ] 是否测试了ID转换功能
  - [ ] 是否包含关系一致性测试
  - [ ] 性能测试是否覆盖查询场景
```

## 🔄 版本演进

### 当前版本: v1.0
```yaml
特性:
  - 基础code命名规范
  - UUID隐藏机制
  - ID转换和缓存策略
  - 基本验证和测试规范

限制:
  - 支持单一数字格式
  - 固定范围分配
  - 基础缓存策略
```

### 未来版本计划

### v1.1 (预计2025-Q4)
```yaml
增强特性:
  - 支持字母数字混合格式 (可选)
  - 智能编码分配算法
  - 更灵活的位数调整机制
  - 跨实体关系验证增强
```

#### v2.0 (预计2026-Q1)  
```yaml
重大升级:
  - 支持动态位数扩展
  - 多租户编码命名空间隔离
  - 智能编码生成算法
  - 完整的编码生命周期管理
  - 编码迁移和重构工具
```

## 📞 支持和反馈

### 技术支持
- **标准咨询**: 架构团队
- **实施指导**: 开发团队负责人  
- **问题反馈**: GitHub Issue
- **紧急支持**: 技术委员会

### 持续改进
- **反馈收集**: 每月开发者调研
- **标准评估**: 每季度架构评审
- **版本升级**: 根据项目需求和反馈

---

**制定者**: 系统架构师  
**审核者**: 技术委员会  
**批准者**: CTO  
**生效日期**: 2025-08-05  
**下次审查**: 2025-11-05