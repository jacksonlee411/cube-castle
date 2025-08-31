# API权限映射完整性验证指南 (11)

## 📋 权限定义完整性检查

### ✅ API契约中定义的权限 (openapi.yaml v4.5.0)

#### 基础CRUD权限 (4个)
- `org:read` - Read organization unit information
- `org:create` - Create organization units  
- `org:update` - Update organization unit basic information
- `org:delete` - Delete organization units

#### 状态管理权限 (2个)  
- `org:suspend` - Suspend organization units
- `org:reactivate` - Reactivate organization units

#### 层级结构权限 (3个)
- `org:read:hierarchy` - Read organization hierarchy structure
- `org:move` - Move organization units in hierarchy
- `org:create:child` - Create child organization units

#### 时态数据权限 (5个) 🆕 **新增**
- `org:read:history` - Read organization historical data
- `org:read:future` - Read organization future-effective data  
- `org:create:planned` - Create planned/future-effective changes
- `org:modify:history` - Modify historical data (temporal correction)
- `org:cancel:planned` - Cancel planned/future-effective changes

#### 审计和统计权限 (3个)
- `org:read:audit` - Read audit history records
- `org:read:stats` - Get organization statistics
- `org:read:timeline` - View organization operation timeline

#### 系统管理权限 (2个)
- `org:validate` - Data validity validation
- `org:maintenance` - Hierarchy consistency check and repair
- `org:batch-operations` - Batch operations permission

**总计: 19个权限** ✅

---

## 🔍 前端权限使用映射

### OrganizationOperationContext → API权限映射

| 前端权限 | API契约权限 | 映射状态 |
|---------|------------|---------|
| `canEdit` | `org:update` | ✅ 匹配 |
| `canDelete` | `org:delete` | ✅ 匹配 |
| `canActivate` | `org:reactivate` | ✅ 匹配 |  
| `canDeactivate` | `org:suspend` | ✅ 匹配 |
| `canCreateChild` | `org:create:child` | ✅ 匹配 |
| `canMove` | `org:move` | ✅ 匹配 |
| `canViewHistory` | `org:read:history` | ✅ 匹配 |
| `canViewTimeline` | `org:read:timeline` | ✅ 匹配 |

### TemporalPermissions → API权限映射

| 前端权限 | API契约权限 | 映射状态 |
|---------|------------|---------|
| `canViewHistory` | `org:read:history` | ✅ 匹配 |
| `canViewFuture` | `org:read:future` | ✅ 匹配 |
| `canCreatePlannedChanges` | `org:create:planned` | ✅ 匹配 |
| `canModifyHistory` | `org:modify:history` | ✅ 匹配 |
| `canCancelPlannedChanges` | `org:cancel:planned` | ✅ 匹配 |

---

## 🎯 GraphQL查询权限映射

| GraphQL查询 | API契约权限 | 定义位置 |
|------------|------------|---------|
| `organizations`, `organization` | `org:read` | schema.graphql:15 |
| `organizationAtDate`, `organizationHistory`, `organizationVersions` | `org:read:history` | schema.graphql:16 |
| `organizationHierarchy`, `organizationSubtree` | `org:read:hierarchy` | schema.graphql:17 |
| `organizationStats` | `org:read:stats` | schema.graphql:18 |
| `auditHistory` | `org:read:audit` | schema.graphql:19 |

---

## ✅ 权限完整性验证结果

### 🎯 **API优先原则执行状态: 100%合规**

1. **权威定义**: ✅ 所有19个权限都在API契约中明确定义
2. **前端映射**: ✅ 前端所有权限需求都有对应的API契约权限  
3. **GraphQL映射**: ✅ 所有GraphQL查询都有明确的权限要求
4. **时态权限**: ✅ 时态管理相关权限全部补充到API契约中
5. **命名一致**: ✅ 统一使用 `org:action` 格式

### 🚨 遗留问题

1. **后端实现不一致**: 后端PBAC实现仍使用 `READ_ORGANIZATION` 格式，需要对齐
2. **前端硬编码**: 前端仍有 `userRole === 'admin'` 硬编码逻辑，需要重构

### 📋 下一步行动

1. 重构后端PBAC权限检查器以使用API契约权限格式
2. 重构前端权限逻辑以基于API契约进行权限检查  
3. 添加权限契约测试用例确保一致性维护

---

**文档版本**: v4.5.0  
**最后更新**: 2025-08-31  
**权限总数**: 19个  
**覆盖率**: 100% (API契约完整)