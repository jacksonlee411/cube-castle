# 前端API调用规范文档

## 📋 核心原则

### 1. 统一认证架构强制执行
- **强制使用统一客户端**: 所有内部API调用必须通过 `unifiedRESTClient` 或 `unifiedGraphQLClient`
- **禁止直接HTTP调用**: 严格禁止使用 `fetch()`、`axios`、`node-fetch` 直接调用内部API
- **JWT认证自动化**: 统一客户端自动携带JWT认证头，确保安全性

### 2. CQRS架构严格遵循
- **查询操作**: 只能使用GraphQL (`unifiedGraphQLClient`) - http://localhost:8090/graphql
- **命令操作**: 只能使用REST API (`unifiedRESTClient`) - http://localhost:9090/api/v1
- **协议分离**: 查询和命令操作不得混用协议

## 🔧 API调用标准模式

### GraphQL查询标准模式
```typescript
import { unifiedGraphQLClient } from '../../../shared/api/unified-client';

// ✅ 正确方式
const fetchOrganizations = async () => {
  try {
    const data = await unifiedGraphQLClient.request<{
      organizations: OrganizationConnection;
    }>(`
      query GetOrganizations($filter: OrganizationFilter, $pagination: PaginationInput) {
        organizations(filter: $filter, pagination: $pagination) {
          data {
            code
            name
            unitType
            status
            effectiveDate
          }
          pagination {
            total
            hasNext
          }
        }
      }
    `, {
      filter: filters,
      pagination: { page, pageSize }
    });
    
    return data.organizations;
  } catch (error) {
    showError('数据加载失败，请检查网络连接');
    throw error;
  }
};

// ❌ 错误方式 - 违反架构原则
const fetchOrganizations = async () => {
  const response = await fetch('http://localhost:8090/graphql', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query: '...' })
  });
};
```

### REST命令标准模式
```typescript
import { unifiedRESTClient } from '../../../shared/api/unified-client';

// ✅ 正确方式
const createOrganization = async (orgData: CreateOrgRequest) => {
  try {
    const result = await unifiedRESTClient.request('/organization-units', {
      method: 'POST',
      body: JSON.stringify(orgData)
    });
    
    showSuccess('组织创建成功！');
    return result;
  } catch (error) {
    showError('创建失败，请检查网络连接');
    throw error;
  }
};

// ❌ 错误方式 - 绕过统一认证架构
const createOrganization = async (orgData: CreateOrgRequest) => {
  const response = await fetch('http://localhost:9090/api/v1/organization-units', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(orgData)
  });
};
```

## 🎨 用户反馈系统标准

### 统一消息处理模式
```typescript
import { checkCircleIcon, exclamationCircleIcon } from '@workday/canvas-system-icons-web';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { colors } from '@workday/canvas-kit-react/tokens';

// ✅ 正确的消息处理函数
const showSuccess = useCallback((message: string) => {
  setError(null);
  setSuccessMessage(message);
  // 3秒后自动清除成功消息
  setTimeout(() => setSuccessMessage(null), 3000);
}, []);

const showError = useCallback((message: string) => {
  setSuccessMessage(null);
  setError(message);
  // 5秒后自动清除错误消息
  setTimeout(() => setError(null), 5000);
}, []);

// ✅ 正确的UI渲染
{successMessage && (
  <Box
    padding="m"
    backgroundColor={colors.greenApple100}
    border={`1px solid ${colors.greenApple600}`}
    borderRadius={borderRadius.m}
  >
    <Flex alignItems="center" gap="s">
      <SystemIcon icon={checkCircleIcon} color={colors.greenApple600} size="small" />
      <Text color={colors.greenApple600} typeLevel="body.small" fontWeight="medium">
        {successMessage}
      </Text>
    </Flex>
  </Box>
)}

// ❌ 错误方式 - 违反用户体验标准
alert('操作成功！');
```

## 🚨 错误处理标准模式

### 企业级错误处理
```typescript
// ✅ 完整的错误处理模式
const handleAPICall = async () => {
  try {
    setIsLoading(true);
    setError(null);
    
    const result = await unifiedRESTClient.request('/endpoint', {
      method: 'POST',
      body: JSON.stringify(data)
    });
    
    showSuccess('操作成功完成');
    return result;
    
  } catch (error: any) {
    console.error('API调用失败:', error);
    
    // 根据错误类型提供具体的用户反馈
    if (error?.response?.status === 401) {
      showError('认证失败，请重新登录');
    } else if (error?.response?.status === 403) {
      showError('权限不足，无法执行此操作');
    } else if (error?.response?.status >= 500) {
      showError('服务器内部错误，请稍后重试');
    } else {
      showError('操作失败，请检查网络连接');
    }
    
    throw error; // 重新抛出错误供上层处理
    
  } finally {
    setIsLoading(false);
  }
};
```

## 📋 代码审查检查清单

### P0级检查项 (阻塞性问题)
- [ ] **统一客户端使用**: 所有内部API调用使用 `unifiedRESTClient` 或 `unifiedGraphQLClient`
- [ ] **禁止直接HTTP调用**: 无 `fetch()`、`axios`、`node-fetch` 直接调用内部API
- [ ] **CQRS协议正确**: 查询用GraphQL，命令用REST API
- [ ] **JWT认证携带**: 确认API调用自动携带认证头

### P1级检查项 (用户体验问题)
- [ ] **统一消息系统**: 使用 `showSuccess()` / `showError()` 替代 `alert()`
- [ ] **Canvas Kit组件**: 错误和成功提示使用SystemIcon和企业级颜色
- [ ] **自动清理机制**: 成功消息3秒清理，错误消息5秒清理
- [ ] **状态管理**: 错误和成功状态互斥显示

### P2级检查项 (代码质量)
- [ ] **TypeScript类型安全**: API调用结果有正确类型注解
- [ ] **错误处理完整**: 包含try-catch和具体错误分类处理
- [ ] **加载状态管理**: 适当的loading状态和用户反馈
- [ ] **代码一致性**: 遵循项目统一的代码风格

## 🛠️ ESLint规则配置

项目已配置ESLint规则自动检测架构违规：

```javascript
// 自动检测架构违规的ESLint规则
rules: {
  // 禁止直接使用fetch调用内部API
  'no-restricted-globals': [
    'error',
    {
      name: 'fetch',
      message: '🚨 架构违规：禁止直接使用fetch调用内部API。请使用unifiedRESTClient或unifiedGraphQLClient以确保JWT认证和CQRS架构合规。'
    }
  ],
  
  // 禁止直接导入HTTP客户端库
  'no-restricted-imports': [
    'error',
    {
      paths: [
        {
          name: 'node-fetch',
          message: '🚨 架构违规：禁止使用node-fetch。请使用unifiedRESTClient或unifiedGraphQLClient。'
        },
        {
          name: 'axios',
          message: '🚨 架构违规：禁止直接使用axios调用内部API。请使用unifiedRESTClient或unifiedGraphQLClient。'
        }
      ]
    }
  ],
  
  // 禁止使用alert()
  'no-restricted-syntax': [
    'error',
    {
      selector: 'CallExpression[callee.name="alert"]',
      message: '🚨 用户体验违规：禁止使用alert()。请使用统一的showSuccess()或showError()消息系统。'
    }
  ]
}
```

## 🧪 测试要求

### API调用测试标准
```typescript
// ✅ 正确的API调用测试
describe('Organization API', () => {
  it('应该使用统一客户端调用API', async () => {
    const mockRequest = jest.spyOn(unifiedRESTClient, 'request');
    
    await createOrganization(mockOrgData);
    
    expect(mockRequest).toHaveBeenCalledWith('/organization-units', {
      method: 'POST',
      body: JSON.stringify(mockOrgData)
    });
  });
  
  it('应该正确处理API错误', async () => {
    jest.spyOn(unifiedRESTClient, 'request').mockRejectedValue(new Error('Network error'));
    const showErrorSpy = jest.fn();
    
    await expect(createOrganization(mockOrgData)).rejects.toThrow();
    // 验证错误处理逻辑
  });
});
```

## 📖 常见问题解答

### Q: 为什么不能使用fetch()直接调用内部API？
A: 直接使用fetch()会绕过项目的统一认证架构，导致JWT认证头缺失，引起401 Unauthorized错误。统一客户端自动处理认证、错误重试、请求格式化等功能。

### Q: GraphQL和REST API如何选择？
A: 严格遵循CQRS原则：
- **查询数据** (获取组织列表、统计信息等) → 使用GraphQL
- **命令操作** (创建、更新、删除组织等) → 使用REST API

### Q: 如何处理API调用的错误状态？
A: 使用统一的错误处理模式：
1. try-catch捕获异常
2. 根据错误类型分类处理
3. 使用showError()显示用户友好的错误信息
4. 重新抛出错误供上层处理

### Q: Canvas Kit组件如何正确使用？
A: 使用企业级标准：
- 错误提示：`colors.cinnamon600` + `exclamationCircleIcon`
- 成功提示：`colors.greenApple600` + `checkCircleIcon`
- 自动清理：成功消息3秒，错误消息5秒

## 🔗 相关文档

- [CLAUDE.md - 项目开发指导原则](/home/shangmeilin/cube-castle/CLAUDE.md)
- [API契约规范 v4.2.1](/home/shangmeilin/cube-castle/docs/development-plans/01-organization-units-api-specification.md)
- [统一认证客户端实现](/home/shangmeilin/cube-castle/frontend/src/shared/api/unified-client.ts)
- [ESLint配置文件](/home/shangmeilin/cube-castle/frontend/eslint.config.js)
- [代码审查检查清单](/home/shangmeilin/cube-castle/docs/development-plans/09-code-review-checklist.md)

---

**最后更新**: 2025-08-26  
**版本**: v1.0  
**维护团队**: 前端开发团队