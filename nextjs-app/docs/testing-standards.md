# Cube Castle 前端测试规范

## 1. 测试原则与目标

### 1.1 核心原则
- **质量优先**: 测试质量比覆盖率数字更重要
- **真实性**: 尽可能模拟真实用户使用场景
- **可维护性**: 测试代码应该易于理解和维护
- **渐进式**: 逐步提升测试覆盖率和质量标准
- **批判性思维**: 持续评估和改进测试策略

### 1.2 测试目标
- 确保代码功能正确性
- 防止回归错误
- 提供代码重构的安全保障
- 作为代码使用文档
- 提升开发者信心

## 2. 测试层次与策略

### 2.1 测试金字塔
```
        E2E Tests (10%)
      ┌─────────────────┐
     Integration Tests (20%)
    ┌─────────────────────┐
   Unit Tests (70%)
  ┌─────────────────────────┐
```

### 2.2 各层测试职责

#### 单元测试 (Unit Tests)
- **目标**: 测试单个函数、组件的独立功能
- **位置**: `tests/unit/`
- **命名**: `*.test.{ts,tsx}`
- **重点**: 逻辑正确性、边界条件、错误处理

#### 集成测试 (Integration Tests)
- **目标**: 测试模块间交互、API集成
- **位置**: `tests/integration/`
- **重点**: 数据流、服务集成、跨模块协作

#### 端到端测试 (E2E Tests)
- **目标**: 模拟真实用户操作流程
- **位置**: `tests/e2e/`
- **工具**: Playwright (推荐) 或 Cypress
- **重点**: 用户关键路径、业务流程

## 3. 技术栈与工具配置

### 3.1 核心工具
- **测试框架**: Jest
- **React测试**: @testing-library/react
- **DOM环境**: jsdom
- **类型支持**: @types/jest
- **覆盖率**: jest内置coverage

### 3.2 配置文件
- `jest.config.js`: Jest主配置
- `jest.setup.js`: 全局测试设置
- `tests/setup/`: 测试环境配置

### 3.3 Mock策略
```typescript
// 优先级顺序
1. 尽量使用真实实现
2. 使用测试专用的轻量级实现
3. Mock外部依赖
4. 最后才Mock内部模块
```

## 4. 代码质量标准

### 4.1 覆盖率目标
```yaml
阶段性目标:
  Phase 1 (当前): 1% - 基础设施建立
  Phase 2 (1个月): 30% - 核心功能覆盖
  Phase 3 (3个月): 60% - 主要模块覆盖
  Phase 4 (6个月): 80% - 生产就绪标准

最低要求:
  statements: 80%
  branches: 75%
  functions: 80%
  lines: 80%
```

### 4.2 代码质量检查
- **ESLint**: 无警告通过
- **TypeScript**: 严格类型检查
- **Prettier**: 代码格式统一
- **测试代码**: 同样需要遵循质量标准

## 5. 测试编写规范

### 5.1 文件结构
```
tests/
├── unit/                 # 单元测试
│   ├── components/       # 组件测试
│   ├── hooks/           # 自定义Hook测试
│   ├── lib/             # 工具库测试
│   └── utils/           # 辅助函数测试
├── integration/         # 集成测试
│   ├── api/             # API集成测试
│   └── pages/           # 页面集成测试
├── e2e/                 # 端到端测试
├── setup/               # 测试环境配置
├── fixtures/            # 测试数据
└── __mocks__/          # Mock文件
```

### 5.2 命名规范

#### 测试文件命名
```typescript
// 组件测试
ComponentName.test.tsx

// 函数测试  
functionName.test.ts

// Hook测试
useHookName.test.ts

// 页面测试
page-name.test.tsx
```

#### 测试用例命名
```typescript
describe('ComponentName', () => {
  // 中文描述，更清晰的业务语义
  it('应该正确渲染用户信息', () => {});
  it('当加载中时应该显示loading状态', () => {});
  it('点击按钮时应该调用相应的回调函数', () => {});
});
```

### 5.3 测试结构模式

#### AAA模式 (Arrange-Act-Assert)
```typescript
it('应该正确计算总价', () => {
  // Arrange - 准备测试数据
  const items = [
    { price: 100, quantity: 2 },
    { price: 50, quantity: 3 }
  ];
  
  // Act - 执行被测试的操作
  const total = calculateTotal(items);
  
  // Assert - 验证结果
  expect(total).toBe(350);
});
```

#### Given-When-Then模式
```typescript
it('当用户提交有效表单时应该成功创建用户', () => {
  // Given - 给定初始条件
  const validUserData = {
    name: '张三',
    email: 'zhangsan@example.com'
  };
  
  // When - 当执行某个操作
  render(<UserForm onSubmit={mockSubmit} />);
  fireEvent.click(screen.getByRole('button', { name: '提交' }));
  
  // Then - 那么应该得到预期结果
  expect(mockSubmit).toHaveBeenCalledWith(validUserData);
});
```

## 6. Mock与模拟策略

### 6.1 Mock优先级原则
1. **避免Mock**: 优先使用真实实现
2. **Mock外部依赖**: 网络请求、第三方库
3. **Mock复杂组件**: UI库组件可适度简化
4. **Mock时间**: 确保测试结果可预测

### 6.2 Mock实施规范

#### 网络请求Mock
```typescript
// 使用MSW进行API Mock (推荐)
import { rest } from 'msw';
import { setupServer } from 'msw/node';

const server = setupServer(
  rest.get('/api/users', (req, res, ctx) => {
    return res(ctx.json({ users: mockUsers }));
  })
);

// 或使用Jest Mock (简单场景)
global.fetch = jest.fn();
```

#### 组件Mock规范
```typescript
// 正确的组件Mock
jest.mock('@/components/ComplexChart', () => {
  return function MockComplexChart({ data, title }: any) {
    return (
      <div data-testid="complex-chart">
        <h3>{title}</h3>
        <div>Data points: {data?.length || 0}</div>
      </div>
    );
  };
});

// 避免过度简化
jest.mock('@/components/ImportantComponent', () => () => <div />); // ❌ 太简化
```

### 6.3 Mock数据管理
```typescript
// tests/fixtures/users.ts
export const mockUsers = [
  {
    id: 'user-1',
    name: '张三',
    email: 'zhangsan@example.com',
    // 包含边界情况的测试数据
  }
];

// 支持数据构建器模式
export const createMockUser = (overrides = {}) => ({
  id: `user-${Math.random()}`,
  name: '测试用户',
  email: 'test@example.com',
  ...overrides
});
```

## 7. 组件测试最佳实践

### 7.1 测试重点
- **渲染正确性**: 确保组件能正确渲染
- **属性传递**: 验证props正确传递和使用
- **用户交互**: 测试点击、输入等用户操作
- **状态变化**: 验证组件状态正确更新
- **错误边界**: 测试异常情况处理

### 7.2 查询优先级
```typescript
// 推荐查询优先级 (按可访问性)
1. getByRole()        // 最推荐，符合可访问性
2. getByLabelText()   // 表单元素
3. getByText()        // 文本内容
4. getByDisplayValue() // 表单值
5. getByAltText()     // 图片alt文本
6. getByTestId()      // 最后选择，仅用于无其他选择时
```

### 7.3 组件测试模板
```typescript
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ComponentName } from '@/components/ComponentName';

// Mock外部依赖
jest.mock('@/lib/api-client');

describe('ComponentName', () => {
  // 通用测试数据
  const defaultProps = {
    title: '测试标题',
    onSubmit: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('渲染测试', () => {
    it('应该正确渲染基本内容', () => {
      render(<ComponentName {...defaultProps} />);
      
      expect(screen.getByText('测试标题')).toBeInTheDocument();
    });

    it('应该正确处理loading状态', () => {
      render(<ComponentName {...defaultProps} loading />);
      
      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });
  });

  describe('用户交互', () => {
    it('点击提交按钮应该调用onSubmit', async () => {
      const user = userEvent.setup();
      render(<ComponentName {...defaultProps} />);
      
      await user.click(screen.getByRole('button', { name: '提交' }));
      
      expect(defaultProps.onSubmit).toHaveBeenCalled();
    });
  });

  describe('错误处理', () => {
    it('应该显示错误信息', () => {
      render(<ComponentName {...defaultProps} error="测试错误" />);
      
      expect(screen.getByText('测试错误')).toBeInTheDocument();
    });
  });
});
```

## 8. API与集成测试

### 8.1 API测试重点
- **成功路径**: 正常API调用和响应
- **错误处理**: HTTP错误码、网络异常
- **数据验证**: 请求参数和响应格式
- **认证授权**: Token处理、权限验证

### 8.2 集成测试策略
```typescript
// API客户端测试
describe('REST API Client', () => {
  beforeEach(() => {
    fetchMock.resetMocks();
  });

  it('应该成功获取用户列表', async () => {
    fetchMock.mockResponseOnce(JSON.stringify({
      users: [{ id: 1, name: '张三' }]
    }));

    const result = await apiClient.getUsers();

    expect(result.success).toBe(true);
    expect(result.data.users).toHaveLength(1);
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/users'),
      expect.objectContaining({ method: 'GET' })
    );
  });
});
```

## 9. 性能与可访问性测试

### 9.1 性能测试
```typescript
// 组件性能测试
it('应该在合理时间内渲染大量数据', () => {
  const startTime = performance.now();
  const largeDataSet = Array.from({ length: 1000 }, (_, i) => ({
    id: i,
    name: `用户${i}`
  }));

  render(<UserList users={largeDataSet} />);

  const endTime = performance.now();
  expect(endTime - startTime).toBeLessThan(100); // 100ms内完成渲染
});
```

### 9.2 可访问性测试
```typescript
import { axe, toHaveNoViolations } from 'jest-axe';

expect.extend(toHaveNoViolations);

it('应该没有可访问性问题', async () => {
  const { container } = render(<ComponentName {...defaultProps} />);
  const results = await axe(container);
  
  expect(results).toHaveNoViolations();
});
```

## 10. 持续集成与质量门禁

### 10.1 CI/CD流程
```yaml
# .github/workflows/test.yml
name: Test Suite
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm ci
      - run: npm run lint
      - run: npm run type-check
      - run: npm test -- --coverage
      - run: npm run test:e2e
```

### 10.2 质量门禁标准
- **代码覆盖率**: 不低于设定阈值
- **测试通过率**: 100%
- **ESLint检查**: 无警告
- **TypeScript**: 无类型错误
- **构建成功**: 无构建错误

## 11. 测试债务管理

### 11.1 技术债务识别
当前项目存在的测试技术债务：

1. **覆盖率过低** (1.28%)
   - 影响: 无法有效防止回归错误
   - 优先级: 高
   - 改进计划: 逐步提升至80%

2. **过度依赖Mock**
   - 影响: 无法发现真实环境问题
   - 优先级: 中
   - 改进计划: 增加集成测试，减少Mock使用

3. **组件测试深度不足**
   - 影响: UI交互问题可能遗漏
   - 优先级: 中
   - 改进计划: 完善用户交互测试

### 11.2 债务偿还策略
1. **每次功能开发必须包含测试**
2. **每个Sprint分配20%时间用于测试改进**
3. **定期进行测试质量审查**
4. **建立测试指标监控dashboard**

## 12. 团队协作与规范

### 12.1 Code Review要求
- **测试覆盖**: 新代码必须有测试覆盖
- **测试质量**: 测试代码需要Review
- **Mock合理性**: 评估Mock策略是否合适
- **边界情况**: 确保覆盖边界和异常情况

### 12.2 文档维护
- **测试规范更新**: 随项目演进更新
- **最佳实践分享**: 定期团队分享
- **问题记录**: 记录测试中遇到的问题和解决方案

## 13. 监控与改进

### 13.1 测试指标监控
- **覆盖率趋势**: 跟踪覆盖率变化
- **测试执行时间**: 监控测试性能
- **失败率统计**: 分析测试稳定性
- **代码质量指标**: ESLint、TypeScript错误数

### 13.2 持续改进机制
- **月度测试质量回顾**
- **季度测试策略评估**
- **年度测试技术栈升级**
- **持续学习最佳实践**

---

## 版本信息
- 文档版本: v1.0
- 创建日期: 2025-01-31
- 最后更新: 2025-01-31
- 维护者: 前端开发团队

## 附录

### A. 常用测试工具链接
- [Jest官方文档](https://jestjs.io/)
- [Testing Library](https://testing-library.com/)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)

### B. 测试最佳实践参考
- [Testing Best Practices](https://github.com/goldbergyoni/javascript-testing-best-practices)
- [React Testing Patterns](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)