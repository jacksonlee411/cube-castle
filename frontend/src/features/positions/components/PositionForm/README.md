# PositionForm 组件说明

## 作用

`PositionForm` 封装职位创建 / 编辑 / 时态版本新增三种模式，统一处理：

- 字段状态管理与校验（`validation.ts`）
- 请求参数构建（`payload.ts`）
- 岗位字典级联选择（依赖 `usePositionCatalogOptions`）
- 成功 / 失败消息反馈（`useMessages`）

## 目录结构

- `index.tsx`：表单容器，负责 Mutation 调用与状态切换
- `FormFields.tsx`：渲染字段与表单布局
- `types.ts`：表单状态 / 错误类型定义
- `validation.ts`：必填项与格式校验
- `payload.ts`：构建创建 / 更新 / 版本请求体
- `PositionFormFields.stories.tsx`：Storybook 示例（正常态 / 字典不可用 / 错误提示）
- `__tests__/`：Vitest 覆盖 payload 与校验逻辑

## 使用方式

```tsx
import { PositionForm } from '@/features/positions/components/PositionForm';

<PositionForm
  mode="create"
  onSuccess={({ code }) => console.log('created', code)}
  onCancel={() => setVisible(false)}
/>;
```

> 编辑 (`mode="edit"`) 与版本 (`mode="version"`) 需传入 `position` 数据，以便预填字段。

## Storybook

运行 `npm --prefix frontend run storybook`，查看 `Positions/PositionForm/Fields` 分类的以下场景：

- `Default`：正常渲染并可编辑
- `CatalogUnavailable`：字典加载失败时的下拉备选
- `WithValidationErrors`：演示错误提示样式

## 测试

```bash
npm --prefix frontend run test -- --run src/features/positions/components/PositionForm
```

## 注意事项

- 新增字段时需同步更新 `createInitialState`、`validation.ts` 与 `payload.ts`
- 若岗位字典接口扩展，请在 `usePositionCatalogOptions` 中补充映射逻辑
- Mutation Hook 位于 `shared/hooks/usePositionMutations.ts`，统一处理缓存刷新
