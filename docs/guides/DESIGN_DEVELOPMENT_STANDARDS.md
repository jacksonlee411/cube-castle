# Cube Castle 项目设计和开发规范文档

## 📋 概述

本文档制定了 Cube Castle 项目的设计和开发标准，确保代码质量、用户体验一致性和维护效率。所有团队成员必须严格遵循本规范。

## 🎯 Canvas Kit v13 图标使用规范

### 核心原则

1. **Canvas Kit 优先**: 所有图标必须优先使用 Canvas Kit v13 的 SystemIcon 组件
2. **严禁使用 Emoji**: 禁止在任何 UI 组件中使用 emoji 图标
3. **语义明确**: 图标选择必须符合其语义含义
4. **一致性**: 相同功能在不同组件中使用相同图标

### 图标映射标准

#### 通用操作图标
```tsx
import { 
  editIcon,        // 编辑操作
  trashIcon,       // 删除操作  
  checkIcon,       // 确认/成功状态
  xIcon,          // 取消/失败状态
  addIcon,        // 新增操作
  refreshIcon,    // 刷新操作
  gearIcon,       // 设置/配置
  searchIcon,     // 搜索功能
  filterIcon,     // 筛选功能
  infoIcon        // 信息提示
} from '@workday/canvas-system-icons-web';
```

#### 时间相关图标
```tsx
import {
  clockIcon,         // 时间/时态管理
  calendarIcon,      // 日期/计划
  timelineAllIcon,   // 时间线显示
  documentIcon       // 历史记录
} from '@workday/canvas-system-icons-web';
```

#### 状态指示图标
```tsx
import {
  checkCircleIcon,      // 启用状态
  exclamationIcon,      // 警告状态
  exclamationCircleIcon // 错误状态
} from '@workday/canvas-system-icons-web';
```

### 使用示例

#### ✅ 正确使用
```tsx
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { editIcon } from '@workday/canvas-system-icons-web';

// 正确的图标使用
<SystemIcon icon={editIcon} size={16} color={colors.blueberry600} />
```

#### ❌ 错误使用
```tsx
// 错误：使用emoji
<span>✏️</span>

// 错误：使用文字替代图标的地方  
<span>编辑</span> // 应该使用SystemIcon

// 错误：混合使用
<span>📅 计划</span> // 应该统一使用Canvas Kit
```

## 🎨 UI 组件设计规范

### 组件结构标准

1. **FormField 组件**: 使用 Canvas Kit v13 复合组件模式
```tsx
<FormField>
  <FormField.Label>标签名称</FormField.Label>
  <FormField.Field>
    <TextInput />
  </FormField.Field>
</FormField>
```

2. **Modal 组件**: 使用 useModalModel 钩子模式
```tsx
const model = useModalModel();

<Modal model={model}>
  <Modal.Overlay>
    <Modal.Card>
      <Modal.CloseIcon onClick={model.events.hide} />
      <Modal.Heading>标题</Modal.Heading>
      <Modal.Body>内容</Modal.Body>
    </Modal.Card>
  </Modal.Overlay>
</Modal>
```

### 语义化文本规范

当Canvas Kit图标库无法满足语义表达需求时，采用以下策略：

1. **使用描述性文本**: 用简洁的中文词汇替代emoji
2. **保持一致性**: 相同概念在项目中使用统一的文字表达
3. **避免歧义**: 确保文字表达清晰明确

#### 标准文字映射
```
✅ -> "启用" 或 "成功"
❌ -> "失败" 或 "错误" 
📅 -> "计划" 或 "日期"
⏰ -> "时间" 或 "当前"
🔄 -> "刷新" 或 "更新"
⚙️ -> "设置" 或 "配置"
📋 -> "详情" 或 "列表"
```

## 💼 时态管理规范

### 类型安全要求

1. **统一字符串类型**: 所有日期时间字段使用 string 类型
2. **类型转换工具**: 使用 TemporalConverter 工具类处理日期转换
3. **零 TypeScript 错误**: 构建过程中不允许任何 TypeScript 错误

### 时态组件命名

```tsx
// 时态相关组件命名规范
TemporalNavbar          // 时态导航栏
TemporalTable          // 时态数据表格  
TemporalSettings       // 时态设置
TemporalStatusSelector // 时态状态选择器
```

## 🔧 代码质量标准

### 导入组织

```tsx
// 1. React 相关
import React from 'react';

// 2. Canvas Kit 组件
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { SystemIcon } from '@workday/canvas-kit-react/icon';

// 3. Canvas Kit 图标
import { editIcon, trashIcon } from '@workday/canvas-system-icons-web';

// 4. 项目内部组件
import { CustomComponent } from '../components/CustomComponent';
```

### 错误处理

```tsx
// 统一错误显示格式
<Text color={colors.cinnamon600}>
  <SystemIcon icon={exclamationIcon} size={16} color={colors.cinnamon600} />
  {errorMessage}
</Text>
```

### 加载状态

```tsx
// 统一加载状态显示
<Text color={colors.blueberry600}>
  加载中...
</Text>
```

## 📝 文档规范

### 组件文档

每个组件必须包含：
1. 功能描述注释
2. Props 接口定义
3. 使用示例
4. 相关的 Canvas Kit 依赖说明

```tsx
/**
 * 时态导航栏组件
 * 提供时态模式切换、时间点选择等核心功能
 * 
 * @param showAdvancedSettings - 是否显示高级设置
 * @param compact - 是否紧凑模式
 * @param onModeChange - 模式切换回调
 */
export interface TemporalNavbarProps {
  showAdvancedSettings?: boolean;
  compact?: boolean;
  onModeChange?: (mode: TemporalMode) => void;
}
```

## 🚀 性能优化规范

### Canvas Kit 优化

1. **按需导入图标**:
```tsx
// ✅ 正确：按需导入
import { editIcon } from '@workday/canvas-system-icons-web';

// ❌ 错误：全量导入
import * as icons from '@workday/canvas-system-icons-web';
```

2. **图标尺寸标准化**:
```tsx
// 标准图标尺寸
size={16}  // 小图标，用于按钮和行内显示
size={20}  // 中等图标，用于卡片标题
size={24}  // 大图标，用于页面标题
```

### 内存管理

1. **避免内存泄漏**: 正确清理事件监听器和定时器
2. **合理使用缓存**: 避免过度缓存导致内存占用过高

## ✅ 验收标准

### 代码审查检查项

- [ ] 无任何 emoji 图标使用
- [ ] 所有图标使用 Canvas Kit SystemIcon
- [ ] TypeScript 编译零错误
- [ ] 组件遵循 Canvas Kit v13 API 规范
- [ ] 统一的错误和加载状态显示
- [ ] 完整的组件文档注释

### 测试要求

1. **单元测试**: 所有新组件必须有对应的单元测试
2. **E2E 测试**: 关键业务流程必须有端到端测试覆盖
3. **类型测试**: 验证 TypeScript 类型定义正确性

## 🔄 规范更新流程

1. **提议**: 通过 Issue 提出规范修改建议
2. **讨论**: 团队成员充分讨论可行性和影响
3. **实施**: 更新文档并通知所有开发者
4. **监督**: 在代码审查中强制执行新规范

## 📚 参考资源

- [Canvas Kit v13 官方文档](https://workday.github.io/canvas-kit/)
- [Canvas Kit 图标库](https://github.com/Workday/canvas-kit/tree/master/modules/icon)
- [TypeScript 最佳实践](https://www.typescriptlang.org/docs/)
- [React 设计模式](https://reactpatterns.com/)

---

**最后更新**: 2025-08-16  
**版本**: v1.0  
**负责人**: Cube Castle 开发团队

> 本规范是确保项目质量和一致性的重要文档，所有团队成员必须严格遵循。如有疑问或建议，请及时与团队沟通。