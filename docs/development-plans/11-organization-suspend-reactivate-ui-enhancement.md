# 11. 组织详情页面停用/重新启用按钮增强方案

## 1. 方案概述

### 1.1 目标
在组织详情页面增加"停用"和"重新启用"按钮，提供直观的组织状态管理功能，符合主流HCM系统的交互模式和企业级用户的操作习惯。

### 1.2 业务价值
- **提升操作效率**：用户无需离开详情页面即可完成组织状态变更
- **增强用户体验**：提供符合主流HCM系统标准的直观交互方式
- **降低操作风险**：通过状态判断和确认机制减少误操作

### 1.3 设计原则
- **符合现有布局**：与"插入新版本"、"修改记录"等按钮保持一致的视觉风格
- **状态驱动显示**：基于当前组织状态智能显示相应操作按钮
- **权限控制**：遵循现有权限体系，确保操作安全性
- **API契约遵循**：使用已有的`/suspend`和`/activate`端点
- **多租户安全**：通过统一客户端自动处理X-Tenant-ID头部
- **并发控制**：实现乐观锁机制防止版本冲突

## 2. 现状分析

### 2.1 当前按钮布局分析
**位置**：组织详情页面右侧区域（`TemporalMasterDetailView.tsx:637-646`）
**现有按钮**：
- 刷新按钮（SecondaryButton）：位于页面头部右侧

**表单操作按钮**：在`InlineNewVersionForm`组件中
- 插入新版本（PrimaryButton）
- 修改记录（SecondaryButton）
- 作废版本（TertiaryButton）

### 2.2 现有状态管理
**状态模型**：
- `ACTIVE`：启用状态，正常运行
- `INACTIVE`：停用状态，等价于停用/暂停
- `PLANNED`：计划状态，未来生效

**可用操作**（`statusUtils.ts:79-90`）：
- `ACTIVE` → 可执行 `UPDATE`, `SUSPEND`
- `INACTIVE` → 可执行 `REACTIVATE`
- `PLANNED` → 可执行 `UPDATE`

### 2.3 API支持现状
**已有端点**：
- `POST /api/v1/organization-units/{code}/suspend`：停用组织
- `POST /api/v1/organization-units/{code}/activate`：重新启用组织

**权限要求**：
- 停用操作：`org:suspend`
- 启用操作：`org:activate`

**多租户支持**：
- 强制要求`X-Tenant-ID`头部（通过`unifiedRESTClient`自动处理）
- 数据完全隔离，确保跨租户安全性

### 2.4 Canvas Kit兼容性现状
**当前版本**：Canvas Kit v13.2.15（与`frontend/package.json`一致）
**可用图标**：
- `pauseIcon` - 用于停用操作
- `playIcon` - 用于启用操作
- `checkCircleIcon` - 用于确认状态
- `exclamationCircleIcon` - 用于警告提示

## 3. 设计方案

### 3.1 按钮布局设计

#### 3.1.1 位置选择
**方案A（推荐）**：页面头部操作区域
- 位置：页面头部右侧，与刷新按钮并列
- 优点：操作优先级高，用户易发现，符合主流HCM系统习惯
- 布局：`[停用/启用] [刷新]`

**方案B**：表单操作区域
- 位置：右侧表单中，与其他操作按钮并列
- 缺点：与其他版本管理操作混合，语义不够清晰

#### 3.1.2 视觉设计
**按钮样式**：
- **停用按钮**：`SecondaryButton`，variant="inverse"，图标：`pauseIcon`（来自`@workday/canvas-system-icons-web`）
- **启用按钮**：`PrimaryButton`，图标：`playIcon`（来自`@workday/canvas-system-icons-web`）
- **兼容性确认**：已核实 Canvas Kit v13 `SecondaryButton` 原生支持 `variant="inverse"`，可直接应用。

**尺寸**：与现有刷新按钮保持一致（默认尺寸）

**示例代码**：
```typescript
import { pauseIcon, playIcon } from '@workday/canvas-system-icons-web';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react';

// 停用按钮
<SecondaryButton
  variant="inverse"
  icon={pauseIcon}
  onClick={handleSuspend}
>
  停用
</SecondaryButton>

// 启用按钮
<PrimaryButton
  icon={playIcon}
  onClick={handleActivate}
>
  启用
</PrimaryButton>
```

### 3.2 显示逻辑设计

#### 3.2.1 按钮显示规则

```typescript
interface ButtonDisplayLogic {
  showSuspendButton: boolean;    // 显示停用按钮
  showActivateButton: boolean;   // 显示启用按钮
  isReadonly: boolean;           // 是否只读模式
  hasSuspendPermission: boolean; // 是否拥有停用权限
  hasActivatePermission: boolean;// 是否拥有启用权限
}

// 核心判断逻辑
const getButtonDisplay = (
  currentStatus: OrganizationStatus,
  isReadonly: boolean,
  permissions: string[]
): ButtonDisplayLogic => {
  const hasSuspendPermission = permissions.includes('org:suspend');
  const hasActivatePermission = permissions.includes('org:activate');

  if (isReadonly) {
    return {
      showSuspendButton: false,
      showActivateButton: false,
      isReadonly,
      hasSuspendPermission,
      hasActivatePermission
    };
  }

  switch (currentStatus) {
    case 'ACTIVE':
      return {
        showSuspendButton: hasSuspendPermission,
        showActivateButton: false,
        isReadonly,
        hasSuspendPermission,
        hasActivatePermission
      };
    case 'INACTIVE':
      return {
        showSuspendButton: false,
        showActivateButton: hasActivatePermission,
        isReadonly,
        hasSuspendPermission,
        hasActivatePermission
      };
    case 'PLANNED':
    default:
      return {
        showSuspendButton: false,
        showActivateButton: false,
        isReadonly,
        hasSuspendPermission,
        hasActivatePermission
      };
  }
};
```

#### 3.2.2 状态判断细则

| 当前状态 | 显示停用按钮 | 显示启用按钮 | 备注 |
|---------|-------------|-------------|------|
| ACTIVE  | ✅ | ❌ | 允许停用 |
| INACTIVE | ❌ | ✅ | 允许重新启用 |
| PLANNED | ❌ | ❌ | 计划状态不支持状态变更 |

**权限控制**：
- 停用按钮：需要`org:suspend`权限
- 启用按钮：需要`org:activate`权限
- 只读模式：隐藏所有操作按钮

### 3.3 交互流程设计

#### 3.3.1 停用操作流程
1. **点击停用按钮** → 弹出带表单的确认对话框（包含必填的停用原因与生效日期输入）
2. **确认对话框内容**：
   ```
   确认停用组织？

   组织名称：[组织名称]
   组织编码：[组织编码]
   当前状态：启用

   停用原因（必填）：[输入框]
   生效日期（必填）：[日期选择器，默认当天]

   停用后该组织将变为非活跃状态，可通过重新启用恢复。

   [取消] [确认停用]
   ```
3. **确认操作**（验证必填输入）→ 调用API → 刷新页面状态 → 显示成功提示

#### 3.3.2 启用操作流程
1. **点击启用按钮** → 弹出带表单的确认对话框（包含必填的启用原因与生效日期输入）
2. **确认对话框内容**：
   ```
   确认重新启用组织？

   组织名称：[组织名称]
   组织编码：[组织编码]
   当前状态：停用

   启用原因（必填）：[输入框]
   生效日期（必填）：[日期选择器，默认当天]

   启用后该组织将恢复正常运行状态。

   [取消] [确认启用]
   ```
3. **确认操作**（验证必填输入）→ 调用API → 刷新页面状态 → 显示成功提示

### 3.4 错误处理

#### 3.4.1 错误处理原则
- **统一入口**：复用共享的 `handleAPIError`，返回一条可展示的中文文案。
- **最小分支**：按钮侧只区分“成功/失败”，失败场景中如检测到版本冲突提示用户刷新重试，其余保持通用错误提示。
- **反馈路径**：通过组件的 `onSuccess` / `onError` 回调，把结果交回页面容器决定是否刷新视图。

#### 3.4.2 并发冲突处理
- **版本冲突检测**：通过 `If-Match` 头部启用乐观锁控制，后端返回 412 时视为并发冲突。
- **自动刷新机制**：冲突时调用现有 Hook 内部的 `invalidateQueries`，确保状态与后端对齐。
- **用户提示**：保持单一提示语，提示用户“数据已更新，请刷新后重试”。

### 3.5 评审发现与优化建议

**评审发现**
- 停用/启用流程确认弹窗缺少对 `operationReason`、`effectiveDate` 两个契约必填字段的采集与校验，直接调用将导致接口校验失败。
- `useSuspendOrganization`、`useActivateOrganization` 当前仍以临时字段 `reason` 调用 API，且统一 REST 客户端丢弃响应头，无法满足契约要求的 `Idempotency-Key` 及 `ETag`/`If-Match` 机制。
- 乐观锁链路尚未回传最新 `ETag`，前端后续操作无法携带版本标识，实际并发冲突仍然存在风险。
- Canvas Kit `SecondaryButton` 支持 `variant="inverse"` 已经过版本确认，可继续按计划使用。

**优化建议**
- 在停用/启用弹窗中新增必填的操作原因与生效日期输入控件，并在提交前完成校验，确保请求体与契约一致。
- 重构 `useSuspendOrganization`、`useActivateOrganization` Hook，改用契约字段名提交，并在内部生成 `Idempotency-Key`，同时支持传入/传出 `If-Match` 所需的 `ETag`。
- 扩展 `unifiedRESTClient` 暴露响应头信息，借此从响应中提取最新 `ETag`，通过新增的 `onETagChange` 回调反馈给页面容器，形成完整的乐观锁闭环。
- 在捕获 412 冲突后结合 React Query 进行缓存失效与重新加载，同时向用户提示刷新数据，避免脏数据继续覆盖。

## 4. 技术实现方案

### 4.1 组件修改范围

#### 4.1.1 主要修改文件
1. **`TemporalMasterDetailView.tsx`**：添加按钮组件和状态逻辑
2. **新建 `SuspendActivateButtons.tsx`**：独立的按钮组件
3. **`statusUtils.ts`**：扩展状态工具函数（如需要）

#### 4.1.2 组件设计

**新建按钮组件**：
```typescript
// SuspendActivateButtons.tsx
interface SuspendActivateButtonsProps {
  currentStatus: OrganizationStatus;
  organizationCode: string;
  organizationName: string;
  currentETag?: string;
  isReadonly?: boolean;
  permissions: string[];
  defaultEffectiveDate?: string;
  onStatusChange: (newStatus: OrganizationStatus) => void;
  onETagChange?: (etag: string | null) => void;
  onError: (error: string) => void;
  onSuccess: (message: string) => void;
}

export const SuspendActivateButtons: React.FC<SuspendActivateButtonsProps>
```

- 组件内部包含停用/启用原因输入框与生效日期选择器，提交前执行必填校验。
- 成功后将最新状态与 `ETag` 通过 `onStatusChange`、`onETagChange` 回传，供页面容器刷新缓存并保存并发控制上下文。

### 4.2 API集成

#### 4.2.1 契约对齐要求
- **请求体字段**：`operationReason`、`effectiveDate` 为后端契约必填字段（参见 `docs/api/openapi.yaml:2076`），前端需强制采集并按原字段名提交，禁止继续使用现有 Hook 中的临时 `reason` 字段。
- **Idempotency-Key**：契约要求提供幂等键 Header。当前统一 REST 客户端与相关 Hook 均未自动生成/附加该 Header，需要在实现阶段补齐（建议基于现有 Idempotency 工具或新增 `generateIdempotencyKey()`）。
- **ETag / If-Match**：契约要求响应返回最新 `ETag` 与请求端附带 `If-Match`。现有客户端会丢弃响应头，Hook 也未设置 `If-Match`；必须扩展统一 REST 客户端以暴露响应头，并在 Hook 中传入最新 ETag 实现乐观锁。

#### 4.2.2 前端调用策略
- **Hook 调整**：扩展 `useSuspendOrganization`、`useActivateOrganization` 接口签名，要求提供 `operationReason`、`effectiveDate`，并在内部映射到契约字段。
- **Header 支持**：在 Hook 中生成并传递 `Idempotency-Key`，同时接收上一笔操作返回的 `ETag` 并以 `If-Match` 头发送；更新统一 REST 客户端以返回 `{ data, headers }` 或类同结构供调用方读取。
- **组件职责**：按钮组件负责管理输入表单及校验，将 `organizationCode`、`operationReason`、`effectiveDate`、`etag`（若存在）传给 Hook，禁止将 `operationReason` 标记为可选。
- **错误分流**：继续复用 `handleAPIError`，并在收到 412 时透出“数据已更新，请刷新后重试”提示，同时触发 React Query 失效逻辑。

### 4.3 状态管理

#### 4.3.1 状态更新流程（React Query集成）

设计重点：

- **缓存策略**：沿用 React Query 既有失效方案（`useSuspendOrganization` 内已处理列表与详情缓存刷新），组件只需在成功回调中触发 `onStatusChange` 以刷新局部状态。
- **消息提示**：统一调用页面层的 `showSuccess` / `showError`。组件内部仅在 Promise 结果后根据成功/失败回调。
- **ETag 管理**：Hook 成功后需将响应头中的最新 `ETag` 返回给调用方，以便后续操作携带 `If-Match`。
- **并发冲突**：使用 `If-Match` 后，412 会由 Hook 捕获并转为“数据已被其他用户修改”消息；组件接收回调后执行刷新逻辑。

#### 4.3.2 乐观更新策略
1. **版本控制**：每次操作前从最近一次成功响应缓存中获取最新 `ETag`
2. **并发检测**：调用 Hook 时传入 `If-Match`；若返回 412，统一提示用户刷新后重试
3. **自动恢复**：冲突时自动刷新数据并提示用户
4. **状态一致性**：确保 UI 状态与后端数据完全同步

### 4.4 确认对话框设计

#### 4.4.1 使用Canvas Kit v13 Modal组件
```typescript
import { Modal } from '@workday/canvas-kit-react/modal';
import { Flex } from '@workday/canvas-kit-react/layout';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react';

interface ConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  content: React.ReactNode;
  isSubmitting?: boolean;
}

const ConfirmationModal: React.FC<ConfirmModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  title,
  content,
  isSubmitting
}) => {
  if (!isOpen) return null;

  return (
    <Modal onTargetChange={onClose}>
      <Modal.Overlay>
        <Modal.Card>
          <Modal.CloseIcon />
          <Modal.Heading>{title}</Modal.Heading>
          <Modal.Body>{content}</Modal.Body>
          <Flex gap="s" marginTop="l">
            <SecondaryButton
              onClick={onClose}
              disabled={isSubmitting}
            >
              取消
            </SecondaryButton>
            <PrimaryButton
              onClick={onConfirm}
              disabled={isSubmitting}
              loading={isSubmitting}
            >
              确认
            </PrimaryButton>
          </Flex>
        </Modal.Card>
      </Modal.Overlay>
    </Modal>
  );
};
```

## 5. 用户体验优化

### 5.1 主流HCM系统对标

根据主流HCM系统的最佳实践：

#### 5.1.1 按钮状态设计
- **可操作状态**：按钮正常显示，具有明确的操作提示
- **不可操作状态**：按钮禁用，并提供禁用原因说明
- **权限不足**：隐藏按钮或显示无权限提示

#### 5.1.2 操作反馈
- **即时反馈**：按钮点击后立即显示加载状态
- **操作结果**：成功/失败消息明确显示
- **状态同步**：页面状态与实际数据保持同步

### 5.2 可访问性增强
```typescript
// 增强的可访问性实现
<SecondaryButton
  variant="inverse"
  icon={pauseIcon}
  aria-label={`停用组织 ${organizationName}`}
  aria-describedby="suspend-help-text"
  disabled={!canSuspend || isSubmitting}
  onClick={handleSuspend}
>
  停用
</SecondaryButton>
<Text id="suspend-help-text" visuallyHidden>
  点击后将显示确认对话框，组织将被标记为非活跃状态
</Text>
```

- **键盘导航**：支持Tab键切换，Enter/Space激活
- **屏幕阅读器**：完整的ARIA标签和描述
- **颜色对比度**：Canvas Kit组件默认符合WCAG AA标准
- **加载状态**：按钮loading属性自动处理可访问性

### 5.3 响应式设计
- **移动端适配**：按钮在小屏幕设备上保持可用性
- **触摸友好**：按钮大小符合触摸操作标准

## 6. 测试策略

### 6.1 单元测试
- **按钮显示逻辑**：测试各种状态下的按钮显示规则
- **API调用**：模拟API调用成功/失败场景
- **权限控制**：测试不同权限组合下的按钮行为

### 6.2 集成测试
- **端到端流程**：测试完整的停用/启用操作流程
- **状态同步**：验证操作后页面状态正确更新
- **错误处理**：测试各种错误场景的处理

### 6.3 用户体验测试
- **可用性测试**：验证按钮位置和操作流程的直观性
- **性能测试**：确保操作响应时间在可接受范围内

## 7. 实施计划

### 7.1 开发阶段
1. **第一阶段**（2天）：
   - 创建`SuspendActivateButtons`组件
   - 实现基本的显示逻辑和API调用

2. **第二阶段**（1天）：
   - 集成到`TemporalMasterDetailView`组件
   - 实现确认对话框

3. **第三阶段**（1天）：
   - 完善错误处理和用户反馈
   - 添加单元测试

### 7.2 测试阶段
1. **功能测试**（1天）：验证基本功能正确性
2. **集成测试**（1天）：验证与现有系统的集成
3. **用户体验测试**（0.5天）：收集用户反馈并优化

### 7.3 部署阶段
1. **代码审查**：确保代码质量和规范性
2. **文档更新**：更新用户手册和API文档
3. **生产部署**：分阶段部署到生产环境

## 8. 风险评估

### 8.1 技术风险
- **API兼容性**：统一客户端缺乏 `Idempotency-Key` 与 `ETag` 支持，需新增能力后方可满足契约 - **风险程度：中**
- **状态同步**：通过 React Query 缓存失效确保一致性 - **风险程度：低**
- **Canvas Kit 兼容**：现版 `@workday/canvas-kit-react@13.2.15` 已覆盖所需组件 - **风险程度：极低**
- **多租户安全**：统一客户端自动处理租户头 - **风险程度：极低**
- **并发冲突**：通过 If-Match + 412 控制冲突，仍需提示用户刷新 - **风险程度：低**

### 8.2 用户体验风险
- **操作混淆**：用户可能混淆停用与删除操作 - **风险程度：中**
- **权限困惑**：用户可能不理解为什么某些按钮不可见 - **风险程度：低**

### 8.3 缓解措施
- **清晰的操作提示**：通过确认对话框明确操作含义
- **完善的错误提示**：提供明确的权限和状态说明
- **渐进式部署**：先在测试环境充分验证再生产部署

## 9. 成功标准

### 9.1 功能标准
- ✅ 按钮在不同状态下正确显示/隐藏
- ✅ 停用/启用操作成功率 > 99%
- ✅ 操作后状态同步正确率 100%

### 9.2 性能标准
- ✅ 按钮响应时间 < 200ms
- ✅ API调用响应时间 < 2s
- ✅ 页面状态更新时间 < 1s

### 9.3 用户体验标准
- ✅ 用户操作路径减少至少50%
- ✅ 用户满意度调查 > 4.5/5
- ✅ 操作错误率 < 1%

## 10. 总结

本方案通过在组织详情页面添加状态驱动的停用/启用按钮，显著提升了用户的操作效率和体验。设计充分考虑了主流HCM系统的交互模式，与现有界面风格保持一致，同时确保了操作的安全性和可靠性。

通过合理的权限控制、清晰的操作流程和完善的错误处理，该功能将为企业用户提供更加直观、高效的组织架构管理体验。

---

*文档版本：v2.1*
*创建日期：2025-09-21*
*更新日期：2025-09-21*
*状态：已优化，可实施*

## 更新记录

### v2.1 (2025-09-21)
- 补充评审发现：明确前端必须采集并提交 `operationReason`、`effectiveDate`，并调整按钮交互流程展示。
- 指出当前 Hook/客户端缺失 `Idempotency-Key`、`ETag`/`If-Match` 支持，要求在实施时扩展并返回响应头。
- 更新组件接口说明，新增 `currentETag`、`onETagChange` 等属性，强调表单校验与并发控制职责。
- 记录 Canvas Kit `variant="inverse"` 兼容性验证结果。

### v2.0 (2025-09-21)
- 修正 Canvas Kit 图标引用（pauseIcon/playIcon 替代不存在的图标）
- 明确使用 unifiedRESTClient 处理多租户安全
- 增加并发控制和乐观锁机制（ETag/If-Match 已生效）
- 简化错误处理与状态管理描述，聚焦复用现有 Hook
- 更新 Canvas Kit v13 Modal API 使用方式
- 增强可访问性实现细节
- 同步 API 契约，启用 `ETag` / `If-Match` + 412 并发防护
