# Cube Castle v2.0 UI组件库开发指南

**版本**: v2.0.0-alpha.1  
**更新时间**: 2025-07-31  
**适用范围**: 现代化UI组件库迁移

## 🎯 概述

Cube Castle v2.0 采用全新的现代化UI组件库架构，完全移除Ant Design，迁移至基于 shadcn/ui + Radix UI + Tailwind CSS 的现代组件系统。

## 🏗️ 技术架构

### 组件分层架构
```
┌─────────────────────────────────────┐
│ Page-level Compositions             │ ← 页面级组合组件
├─────────────────────────────────────┤
│ Custom Business Components          │ ← 业务组件
├─────────────────────────────────────┤
│ shadcn/ui Components               │ ← 设计系统实现
├─────────────────────────────────────┤
│ Radix UI Primitives                │ ← 无头组件基础
└─────────────────────────────────────┘
```

### 核心技术栈
- **无头组件**: Radix UI Primitives
- **设计系统**: shadcn/ui
- **样式系统**: Tailwind CSS 3.4+
- **图标库**: Lucide React
- **类型安全**: TypeScript 5.5+

## 📦 组件库使用指南

### 基础组件导入
```typescript
// 基础UI组件
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

// 图标
import { User, Settings, Plus } from 'lucide-react'

// 表单组件
import { Form, FormControl, FormField, FormItem, FormLabel } from '@/components/ui/form'
```

### 常用组件示例

#### 按钮组件
```typescript
import { Button } from '@/components/ui/button'

// 基础用法
<Button>默认按钮</Button>
<Button variant="outline">边框按钮</Button>
<Button variant="ghost">幽灵按钮</Button>
<Button size="sm">小尺寸</Button>

// 带图标
<Button>
  <Plus className="mr-2 h-4 w-4" />
  添加用户
</Button>
```

#### 表单组件
```typescript
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Form, FormControl, FormField, FormItem, FormLabel } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

const formSchema = z.object({
  username: z.string().min(2, "用户名至少2个字符"),
  email: z.string().email("请输入有效的邮箱地址")
})

function MyForm() {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      email: ""
    }
  })

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormLabel>用户名</FormLabel>
              <FormControl>
                <Input placeholder="请输入用户名" {...field} />
              </FormControl>
            </FormItem>
          )}
        />
        <Button type="submit">提交</Button>
      </form>
    </Form>
  )
}
```

#### 数据表格
```typescript
import { DataTable } from '@/components/ui/data-table'
import { ColumnDef } from '@tanstack/react-table'

const columns: ColumnDef<User>[] = [
  {
    accessorKey: "name",
    header: "姓名",
  },
  {
    accessorKey: "email", 
    header: "邮箱",
  },
  {
    id: "actions",
    header: "操作",
    cell: ({ row }) => (
      <Button variant="ghost" size="sm">
        编辑
      </Button>
    ),
  },
]

<DataTable columns={columns} data={users} />
```

## 🎨 样式系统

### Tailwind CSS 使用
```typescript
// 响应式设计
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">

// 状态变化
<Button className="hover:bg-blue-600 focus:ring-2 focus:ring-blue-500">

// 深色模式支持
<div className="bg-white dark:bg-gray-800 text-gray-900 dark:text-white">
```

### CSS Variables 主题定制
```css
:root {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --primary: 222.2 47.4% 11.2%;
  --primary-foreground: 210 40% 98%;
}

.dark {
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
}
```

## 🔄 迁移指南

### 从 Ant Design 迁移

#### 常用组件映射表
| Ant Design | 现代化组件 | 导入路径 |
|------------|-----------|----------|
| `<Button>` | `<Button>` | `@/components/ui/button` |
| `<Input>` | `<Input>` | `@/components/ui/input` |
| `<Card>` | `<Card>` | `@/components/ui/card` |
| `<Table>` | `<DataTable>` | `@/components/ui/data-table` |
| `<Form>` | `<Form>` + React Hook Form | `@/components/ui/form` |
| `<Select>` | `<Select>` | `@/components/ui/select` |
| `<Modal>` | `<Dialog>` | `@/components/ui/dialog` |
| `<Tooltip>` | `<Tooltip>` | `@/components/ui/tooltip` |

#### 迁移示例

**之前 (Ant Design)**:
```typescript
import { Button, Input, Form, message } from 'antd'
import { UserOutlined } from '@ant-design/icons'

<Form onFinish={onFinish}>
  <Form.Item name="username" rules={[{ required: true }]}>
    <Input prefix={<UserOutlined />} placeholder="用户名" />
  </Form.Item>
  <Form.Item>
    <Button type="primary" htmlType="submit">
      提交
    </Button>
  </Form.Item>
</Form>
```

**现在 (现代化组件)**:
```typescript
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Form, FormControl, FormField, FormItem } from '@/components/ui/form'
import { User } from 'lucide-react'
import { toast } from 'sonner'

<Form {...form}>
  <form onSubmit={form.handleSubmit(onSubmit)}>
    <FormField
      name="username"
      render={({ field }) => (
        <FormItem>
          <FormControl>
            <div className="relative">
              <User className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
              <Input className="pl-10" placeholder="用户名" {...field} />
            </div>
          </FormControl>
        </FormItem>
      )}
    />
    <Button type="submit">提交</Button>
  </form>
</Form>
```

## 🛠️ 开发工作流

### 创建新组件
1. 在 `src/components/ui/` 目录下创建组件文件
2. 使用 TypeScript 和 Radix UI primitives
3. 添加 Tailwind CSS 样式
4. 导出组件和类型定义

### 组件开发模板
```typescript
// src/components/ui/my-component.tsx
"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface MyComponentProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "destructive"
  size?: "default" | "sm" | "lg"
}

const MyComponent = React.forwardRef<HTMLDivElement, MyComponentProps>(
  ({ className, variant = "default", size = "default", ...props }, ref) => {
    return (
      <div
        className={cn(
          "base-styles",
          {
            "variant-styles": variant === "default",
            "size-styles": size === "default",
          },
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
MyComponent.displayName = "MyComponent"

export { MyComponent }
```

## 📋 最佳实践

### 1. 类型安全
- 使用 TypeScript 严格模式
- 为组件属性定义准确的类型
- 使用 Zod 进行运行时验证

### 2. 可访问性
- 使用 Radix UI 获得内置可访问性
- 添加适当的 ARIA 标签
- 确保键盘导航支持

### 3. 性能优化
- 使用 React.forwardRef 避免不必要的重渲染
- 合理使用 React.memo
- 按需导入组件和工具函数

### 4. 样式管理
- 优先使用 Tailwind 工具类
- 使用 CSS Variables 进行主题定制
- 保持样式的一致性和可维护性

## 🚨 注意事项

### 当前限制 (v2.0.0-alpha.1)
- 部分复杂组件仍在重构中
- 某些页面功能暂时不可用
- 建议仅在开发环境使用

### 常见问题

#### Q: 如何处理深色模式？
A: 使用 `next-themes` 和 Tailwind 的 `dark:` 前缀类。

#### Q: 如何自定义主题颜色？
A: 修改 `tailwind.config.js` 中的颜色定义和 CSS Variables。

#### Q: 如何处理表单验证？
A: 使用 React Hook Form + Zod 的组合，提供类型安全的验证。

## 🔗 相关资源

- **shadcn/ui 文档**: https://ui.shadcn.com/
- **Radix UI 文档**: https://www.radix-ui.com/
- **Tailwind CSS 文档**: https://tailwindcss.com/
- **Lucide React 图标**: https://lucide.dev/
- **React Hook Form**: https://react-hook-form.com/

---

**维护团队**: Cube Castle 前端开发团队  
**更新频率**: 随技术栈演进定期更新  
**反馈渠道**: 技术团队内部讨论