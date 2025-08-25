# 前端代码审查检查清单

## 🎯 使用说明

本检查清单适用于所有前端代码审查，确保代码符合项目架构原则和质量标准。审查者应逐项检查，发现问题必须要求修复后方可合并。

## 🚨 P0级检查项 (阻塞性问题 - 必须通过)

### API调用架构合规性
- [ ] **统一客户端使用**: 所有内部API调用使用 `unifiedRESTClient` 或 `unifiedGraphQLClient`
  - 检查点：搜索 `fetch(`、`axios(`、导入`node-fetch`
  - 违规示例：`fetch('http://localhost:9090/api/...')`
  - 正确示例：`await unifiedRESTClient.request('/endpoint', options)`

- [ ] **CQRS协议正确性**: 
  - 查询操作 → GraphQL (`unifiedGraphQLClient`)
  - 命令操作 → REST API (`unifiedRESTClient`)
  - 检查点：确认API调用选择正确协议

- [ ] **JWT认证架构**: 
  - 无直接设置Authorization头 (统一客户端自动处理)
  - 无认证绕过代码
  - 检查点：搜索 `Authorization:`、`Bearer`

### 导入和依赖合规性
- [ ] **Canvas Kit v13标准**: 
  - 导入路径正确：`@workday/canvas-kit-react/`
  - 无废弃API使用
  - SystemIcon正确导入和使用

- [ ] **类型导入**: 
  - TypeScript接口正确导入
  - 无`any`类型滥用
  - 泛型类型正确使用

## 🟡 P1级检查项 (用户体验问题)

### 用户反馈系统
- [ ] **统一消息系统**: 
  - 使用 `showSuccess()` / `showError()` 替代 `alert()`
  - 检查点：搜索 `alert(`
  - 自动清理机制：成功3秒，错误5秒

- [ ] **企业级视觉标准**: 
  - 错误提示：`colors.cinnamon600` + `exclamationCircleIcon`
  - 成功提示：`colors.greenApple600` + `checkCircleIcon` 
  - 状态互斥：错误和成功状态不同时显示

- [ ] **加载状态管理**: 
  - API调用期间显示loading状态
  - 禁用相关交互元素
  - 提供用户反馈

### 错误处理完整性
- [ ] **try-catch覆盖**: 所有API调用包含完整错误处理
- [ ] **错误分类处理**: 
  - 401 → "认证失败，请重新登录"
  - 403 → "权限不足，无法执行此操作"  
  - 5xx → "服务器内部错误，请稍后重试"
  - 其他 → "操作失败，请检查网络连接"

- [ ] **错误重新抛出**: catch块中适当重新抛出错误供上层处理

## 🔵 P2级检查项 (代码质量)

### TypeScript类型安全
- [ ] **类型注解完整**: API调用结果有明确类型声明
- [ ] **接口定义**: 复杂数据结构使用interface定义
- [ ] **联合类型**: 状态字段使用联合类型 (`'ACTIVE' | 'INACTIVE'`)
- [ ] **泛型使用**: GraphQL查询使用泛型类型参数

### 代码结构和命名
- [ ] **变量命名**: 
  - camelCase命名规范
  - 语义化命名
  - 无拼写错误

- [ ] **函数设计**: 
  - 单一职责原则
  - 合理的函数长度 (<50行)
  - 纯函数优先

- [ ] **状态管理**: 
  - 最小化状态数量
  - 状态更新逻辑清晰
  - useCallback和useMemo适当使用

### 性能优化
- [ ] **依赖数组**: useEffect、useCallback、useMemo的依赖数组正确
- [ ] **不必要渲染**: 避免不必要的组件重新渲染
- [ ] **内存泄漏**: 适当的清理逻辑 (定时器、事件监听器)

## 📋 具体审查步骤

### 第一步：自动化检查
```bash
# 运行ESLint检查架构违规
npm run lint

# 运行TypeScript编译检查
npm run typecheck

# 运行测试套件
npm test
```

### 第二步：手动代码审查
1. **文件级审查**:
   - 检查导入语句合规性
   - 验证导出内容适当性
   - 确认文件结构清晰

2. **组件级审查**:
   - Props类型定义完整
   - 状态管理合理
   - 生命周期使用正确

3. **函数级审查**:
   - API调用模式正确
   - 错误处理完整
   - 返回值类型正确

### 第三步：功能验证
- [ ] **手动测试**: 在浏览器中验证功能正常工作
- [ ] **错误场景**: 测试网络错误、权限错误等异常情况
- [ ] **用户体验**: 确认加载状态、成功/错误反馈正常

## 🚨 常见违规模式识别

### 架构违规代码模式
```typescript
// ❌ 违规模式1: 直接fetch调用
const response = await fetch('http://localhost:9090/api/v1/organization-units', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(data)
});

// ❌ 违规模式2: alert()使用
alert('操作成功！');

// ❌ 违规模式3: 手动设置认证头
headers: {
  'Authorization': `Bearer ${token}`,
  'Content-Type': 'application/json'
}

// ❌ 违规模式4: 混用协议
// 查询操作使用REST API，命令操作使用GraphQL
```

### 正确代码模式
```typescript
// ✅ 正确模式1: 统一客户端使用
const result = await unifiedRESTClient.request('/organization-units', {
  method: 'POST',
  body: JSON.stringify(data)
});

// ✅ 正确模式2: 统一消息系统
showSuccess('操作成功！');

// ✅ 正确模式3: 完整错误处理
try {
  const result = await apiCall();
  showSuccess('操作成功');
} catch (error) {
  console.error('API调用失败:', error);
  showError('操作失败，请检查网络连接');
  throw error;
}
```

## 📊 审查质量标准

### 通过标准
- **P0级检查项**: 100%通过 (0个违规)
- **P1级检查项**: 95%以上通过 (≤1个轻微问题)
- **P2级检查项**: 90%以上通过 (可接受少量代码风格问题)

### 阻塞条件
以下情况必须阻止代码合并：
- 任何P0级架构违规问题
- TypeScript编译错误
- ESLint错误级别问题
- 测试套件失败
- 手动功能测试失败

## 🔧 审查工具配置

### VSCode配置
```json
{
  "eslint.validate": ["javascript", "typescript", "typescriptreact"],
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "typescript.preferences.includePackageJsonAutoImports": "off"
}
```

### Git Hook配置
项目已配置Pre-commit hook自动检查：
- ESLint架构违规检查
- TypeScript编译验证
- 基础代码格式检查

## 📋 审查记录模板

### Pull Request审查评论模板
```markdown
## 代码审查结果

### ✅ 通过项
- [ ] P0级架构合规检查通过
- [ ] TypeScript编译无错误  
- [ ] ESLint检查通过
- [ ] 手动功能测试正常

### ⚠️ 需要修复的问题
- [ ] 问题1: 描述 (P0/P1/P2级)
- [ ] 问题2: 描述 (P0/P1/P2级)

### 💡 改进建议
- 建议1: 描述
- 建议2: 描述

### 审查结论
- [ ] ✅ 批准合并
- [ ] ⚠️ 需要修复后重新审查  
- [ ] ❌ 拒绝合并 (重大问题)
```

---

**文档版本**: v1.0  
**最后更新**: 2025-08-26  
**适用范围**: 所有前端代码Pull Request审查  
**维护团队**: 前端开发团队