# 🚨 接口定义冻结令 (Interface Definition Freeze)

**生效日期**: 2025-09-07  
**冻结级别**: **S级** - 立即生效  
**适用范围**: 所有组织相关接口定义  

## 📋 冻结原因

基于重复代码消除计划的发现，项目中存在**55个重复的组织接口定义**，冗余度高达87%，已达到不可维护状态。

## 🛑 严格禁止事项

### 1. 禁止新增组织接口
- ❌ 不得创建新的 `Organization*` 接口
- ❌ 不得在组件内部定义临时接口
- ❌ 不得复制现有接口到新文件

### 2. 禁止新增重复Hook
- ❌ 不得创建新的组织相关Hook
- ✅ 只允许使用: `useEnterpriseOrganizations`, `useOrganizationList`

### 3. 禁止新增API客户端
- ❌ 不得创建新的API客户端实现
- ✅ 只允许使用: `unified-client.ts`

## ✅ 强制要求

### 必须复用现有接口
```typescript
// ✅ 正确 - 复用现有接口
import { OrganizationUnit } from '@/shared/types/organization';

// ❌ 错误 - 创建新接口
interface MyOrganizationData { ... }
```

### 必须使用允许的Hook
```typescript
// ✅ 正确 - 使用允许的Hook
import { useEnterpriseOrganizations } from '@/shared/hooks';

// ❌ 错误 - 创建新Hook
const useMyOrganizations = () => { ... }
```

### 必须使用统一客户端
```typescript
// ✅ 正确 - 使用统一客户端
import { organizationAPI } from '@/shared/api/unified-client';

// ❌ 错误 - 创建新客户端
const myOrganizationAPI = { ... }
```

## 🔍 强制检查机制

### ESLint规则
- 配置文件: `.eslintrc.interface-freeze.json`
- 违规检查: 提交时自动阻止
- 错误级别: **Error** (阻止构建)

### CI/CD门禁
- Pull Request自动检查
- 违规时拒绝合并
- 需要架构师review才能豁免

## 🆘 例外申请流程

### 申请条件
1. 确实无法使用现有接口满足需求
2. 已尝试扩展现有接口但不可行
3. 新功能对业务价值重大

### 申请流程
1. 在GitHub Issue中详细说明需求
2. 架构师review和批准
3. 更新此冻结规则文档
4. 在代码中添加详细注释说明

## 📈 解冻条件

### 自动解冻触发条件
- 接口定义数量减少到10个以下
- 冗余度降低到10%以下
- 完成Phase 2类型系统重构

### 手动解冻
- 架构师评估项目状态
- 确认技术债务已清理
- 更新开发规范和最佳实践

## ⚠️ 违规后果

### 开发阶段
- ESLint报错，无法构建
- Pre-commit Hook阻止提交
- CI/CD管道失败

### 代码审查
- PR自动标记为违规
- 需要修复后才能合并
- 记录违规次数

### 项目影响
- 技术债务持续积累
- 维护成本指数增长
- 项目可维护性崩溃

---

**🚨 此冻结令为项目生存关键措施，严格执行零容忍政策！**

最后更新: 2025-09-07  
更新人: Claude Code Agent  
下次review: Phase 1完成后