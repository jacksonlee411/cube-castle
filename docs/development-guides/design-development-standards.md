# 设计与开发标准（前端补充）

> 声明：本文件为“前端 UI/组件规范补充”，权威规范以仓库根目录 `CLAUDE.md` 为准；与 API 契约无关。

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
```

