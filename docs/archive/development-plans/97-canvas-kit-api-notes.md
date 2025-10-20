# 97号 Canvas Kit v13.2.x API 调研笔记

- **版本确认**：项目依赖在 `frontend/package.json` 中锁定为 `@workday/canvas-kit-react@^13.2.15`，实际安装版本为 `13.2.18`（见 `npm list @workday/canvas-kit-react` 输出）。
- **PrimaryButton 变更**：`frontend/node_modules/@workday/canvas-kit-react/button/lib/PrimaryButton.tsx` 定义的 `PrimaryButtonProps` 仅暴露 `variant?: 'inverse'`，原先的 `'primary'|'secondary'` 语义已被移除，需要通过按钮类型（Primary/Secondary/Tertiary）区分样式，而非 `variant` 字符串。
- **Select 组合式用法**：`frontend/node_modules/@workday/canvas-kit-react/select/lib/Select.tsx` 显示 `Select` 现为容器组件，必须搭配 `Select.Input`、`Select.Popper`、`Select.Card` 等子组件；容器本身未暴露 `disabled`，需通过模型或为 `Select.Input` 传入对应属性后拦截交互。
- **Modal/Popup 关闭流程**：`frontend/node_modules/@workday/canvas-kit-react/modal/lib/Modal.tsx` 与 `dialog/lib/Dialog.tsx` 均未暴露 `onClose` 属性，关闭逻辑依赖 `model.events.hide()` 及子组件（如 `Modal.CloseIcon`、`Modal.CloseButton`），需要在业务层手动订阅 `model.state.visibility` 并调用外部 `onClose`。
- **FormField 错误态**：`frontend/node_modules/@workday/canvas-kit-react/form-field/lib/hooks/useFormFieldModel.tsx` 将 `error` 收敛为 `'error' | 'alert' | undefined`，对应输入组件（见 `useFormFieldInput.tsx`）会透传 `aria-invalid` 与 `error` 枚举，不再接受布尔值；需要在表单层使用模型或自定义封装保持兼容。
