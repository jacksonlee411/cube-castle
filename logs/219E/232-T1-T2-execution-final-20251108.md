# Plan 232 T1/T2 执行总结（最终）

**执行时间**：2025-11-08 19:45 - 20:15 CST
**执行内容**：T1（添加 data-testid）+ T2（检查 waitPatterns.ts）+ E2E 验证
**执行状态**：⚠️ T1/T2 已完成，E2E 测试仍失败（根本原因待查）

---

## T1 执行详情：为 CatalogVersionForm 添加 data-testid 支持

### 修改内容

**文件 1**：`frontend/src/features/job-catalog/shared/CatalogVersionForm.tsx`

```diff
interface CatalogVersionFormProps {
  title: string
  isOpen: boolean
  onClose: () => void
  onSubmit: (values: CatalogVersionFormValues) => Promise<void>
  isSubmitting?: boolean
  initialName?: string
  initialDescription?: string | null
  initialStatus?: JobCatalogStatus
  initialEffectiveDate?: string
  submitLabel?: string
+ cardTestId?: string
}

export const CatalogVersionForm: React.FC<CatalogVersionFormProps> = ({
  title,
  isOpen,
  onClose,
  onSubmit,
  isSubmitting = false,
  initialName,
  initialDescription,
  initialStatus,
  initialEffectiveDate,
  submitLabel,
+ cardTestId = 'catalog-version-form-dialog',
}) => {
  // ...
  return (
    <CatalogForm
      // ...
      cardTestId={cardTestId}
    >
```

**文件 2**：`frontend/src/features/job-catalog/family-groups/JobFamilyGroupDetail.tsx`

```diff
<CatalogVersionForm
  title="编辑职类信息"
  isOpen={isEditFormOpen}
  onClose={() => setEditFormOpen(false)}
  onSubmit={handleUpdate}
  isSubmitting={updateMutation.isPending}
  initialName={group.name}
  initialDescription={group.description}
  initialStatus={group.status}
  initialEffectiveDate={group.effectiveDate}
  submitLabel="保存更新"
+ cardTestId="catalog-version-form-dialog"
/>

<CatalogVersionForm
  title="新增职类版本"
  isOpen={isVersionFormOpen}
  onClose={() => setVersionFormOpen(false)}
  onSubmit={handleCreateVersion}
  isSubmitting={versionMutation.isPending}
  initialName={group.name}
  initialDescription={group.description}
+ cardTestId="catalog-create-version-form-dialog"
/>
```

**文件 3**：`frontend/src/features/job-catalog/shared/CatalogForm.tsx`（已存在，无需修改）

验证：Modal.Card 已正确接收 data-testid 属性（第 70 行）：
```typescript
<Modal.Card width={width} paddingBottom="s" data-testid={cardTestId}>
```

### 修改原因

- 原始 E2E 脚本（`job-catalog-secondary-navigation.spec.ts`）已被修改，第 199 行现在查询：`getByTestId('catalog-version-form-dialog')`
- 原始脚本在第 190 行曾查询：`getByText('编辑职类信息').first()`（查找 modal 标题）
- 修改允许外部传入 testid 参数，同时保留默认值为 `'catalog-version-form-dialog'`
- 为第二个 CatalogVersionForm（新增版本）指定不同的 testid `'catalog-create-version-form-dialog'` 以区分两个模态框

### 代码影响分析

- ✅ **向后兼容**：未传入 cardTestId 时，默认为 `'catalog-version-form-dialog'`
- ✅ **零风险**：仅添加可选参数，不改变现有逻辑
- ✅ **编译通过**：`npm run typecheck` 无错误

### 行数统计

- 文件 1 (CatalogVersionForm.tsx)：+3 行（Props 接口 + 参数解构 + CatalogForm 传递）
- 文件 2 (JobFamilyGroupDetail.tsx)：+2 行（两个 cardTestId 属性）
- **总计**：5 行改动（相比执行总结中的 4 行，因为为第二个 form 也添加了 testid）

---

## T2 执行详情：waitPatterns.ts 验证

### 发现

`frontend/tests/e2e/utils/waitPatterns.ts` 已存在，包含：

```typescript
export const waitForPageReady = async (page: Page, options?: WaitOptions)
export const waitForNavigation = async (page: Page, expectedUrl: UrlMatcher, options?: WaitOptions)
export const waitForGraphQL = async (page: Page, operationName: string | RegExp, options?: GraphQLWaitOptions)
```

### 现状评估

✅ **已满足需求**：
- 库已存在（70 行代码，比建议的轻量版更完整）
- 包含 3 个标准等待函数
- 支持高级特性（URL 匹配、GraphQL 操作名称匹配）
- 已被 job-catalog-secondary-navigation.spec.ts 使用

### 脚本修改确认

E2E 脚本已被修改，添加了 waitPatterns 的使用：

```typescript
import { waitForGraphQL, waitForPageReady } from './utils/waitPatterns';  // 第 5 行（新增）

await page.goto(`/positions/catalog/family-groups/${jobFamilyGroup.code}`);
await waitForPageReady(page);  // 第 187 行（新增）
await waitForGraphQL(page, /jobFamilyGroup/i).catch(() => {});  // 第 188 行（新增）
```

### 结论

T2 无需执行 - waitPatterns.ts 已存在且已被脚本使用。

---

## T3 执行详情：E2E 测试验证

### 测试执行环境

- **命令**：`npm run test:e2e -- --project=chromium tests/e2e/job-catalog-secondary-navigation.spec.ts`
- **浏览器**：Chromium
- **测试目标**：验证编辑职类对话框是否能成功打开

### 测试结果：❌ 失败

```
错误信息：编辑职类对话框未弹出
超时时间：15000ms
定位器：getByTestId('catalog-version-form-dialog')
预期：visible
实际：<element(s) not found>
```

### 失败位置详情

```typescript
// 第 196-200 行
const editButton = page.getByRole('main').getByRole('button', { name: '编辑当前版本' });
await expect(editButton, '编辑按钮尚未可用').toBeEnabled();
await editButton.click();
const editDialog = page.getByTestId('catalog-version-form-dialog');
await expect(editDialog, '编辑职类对话框未弹出').toBeVisible({ timeout: 15000 });  // ❌ 失败
```

### 页面状态快照

错误发生时页面显示：
- ✅ 正确加载了职类详情页面（"职类详情" 标题可见）
- ✅ 编辑按钮可见并已启用
- ✅ 职类信息（编码、名称、状态等）正确显示
- ❌ **编辑对话框未出现**（无 testid 元素，无标题 "编辑职类信息"）

### 根本原因分析

**问题核心**：模态对话框本身**未渲染**，而非仅仅 testid 不可用

1. **单击后状态不更新**：按钮被点击（能到达该位置），但 `isEditFormOpen` 状态似乎未更新
2. **CatalogVersionForm 的 isOpen 属性未被设置为 true**
3. **CatalogForm 的条件渲染失败**（第 53-54 行检查 visibility）：
   ```typescript
   if (modalModel.state.visibility !== 'visible') {
     return null
   }
   ```

### 可能的根本原因

1. **开发服务器代码同步问题**
   - 源文件已修改，但运行中的 dev server 未重新编译
   - 可能需要 `npm run dev` 或重启 dev server

2. **React 状态更新延迟**
   - 虽然代码看起来正确，但可能存在异步问题
   - useEffect 或状态管理流程中存在阻塞

3. **Canvas Kit Modal 组件问题**
   - `useModalModel` 初始化或事件分发存在问题
   - 可能需要检查 Modal 组件版本或配置

4. **浏览器/Playwright 问题**
   - 页面重定向或导航打断了点击事件处理
   - 事件冒泡被某个元素阻止

### 测试执行日志截断

由于时间限制，仅运行了 Chromium 测试，Firefox 测试未执行。

---

## 与之前调查的对比

### 原始测试失败（2025-11-07）

```
Locator: getByText('编辑职类信息').first()
Expected: visible
Received: <element(s) not found>
```

**脚本在第 190 行查找对话框标题文本**，同样失败。

### 当前测试失败（2025-11-08）

```
Locator: getByTestId('catalog-version-form-dialog')
Expected: visible
Received: <element(s) not found>
```

**脚本在第 199 行查找 testid**，同样失败。

### 结论

问题**不在 testid 属性本身**，而在**对话框组件完全未渲染**。T1 修改（添加 testid）是正确的，但无法掩盖根本的渲染问题。

---

## 建议的后续步骤

### 立即行动（P0）

1. **验证开发服务器状态**
   ```bash
   # 确保 dev server 正在运行且代码已重编译
   npm run dev
   # 在另一个终端运行 E2E 测试
   npm run test:e2e -- --project=chromium tests/e2e/job-catalog-secondary-navigation.spec.ts
   ```

2. **添加调试日志**
   - 在 JobFamilyGroupDetail 的 onClick 处理器中添加 console.log
   - 在 CatalogForm 的 useEffect 中添加日志
   - 检查 Canvas Kit Modal 的状态变化

3. **排查 Canvas Kit Modal 问题**
   - 检查 `useModalModel` 初始化时的 visibility 参数
   - 验证 `modalModel.events.show()` 是否被调用
   - 查看是否需要更新 @workday/canvas-kit 版本

### 深度调查（P1）

4. **逐步还原测试脚本改动**
   - 将脚本改回原始状态（查找标题文本而非 testid）
   - 如果原始查询方式也失败，说明问题与 testid/查询方式无关

5. **检查 React DevTools**
   - 在实际浏览器中打开职类详情页面
   - 使用 React DevTools 查看组件树和状态变化
   - 点击编辑按钮时观察 isEditFormOpen 状态是否真的更新

6. **网络与浏览器事件追踪**
   - 查看 Playwright 生成的 trace.zip 文件（包含完整事件日志）
   - 检查是否有网络请求干扰（GraphQL 查询导致重新加载）

---

## 工件与证据

### T1 修改确认

✅ 文件已修改：
- CatalogVersionForm.tsx：添加了 cardTestId 参数和默认值
- CatalogForm.tsx：确认已有 data-testid 属性传递
- JobFamilyGroupDetail.tsx：两个 CatalogVersionForm 调用都有 cardTestId
- 无 TypeScript 错误（`npm run typecheck` 通过）

### T2 验证确认

✅ waitPatterns.ts 已存在并已使用

### E2E 测试工件

- 测试结果路径：`frontend/test-results/job-catalog-secondary-navi-af1dd-管理员通过-UI-编辑职类成功并触发-If-Match-chromium/`
- 包含：test-failed-1.png（页面快照）、video.webm（完整视频）、trace.zip（详细追踪）

---

## 执行反思与建议

### T1 范围评估

原计划将 T1 定为 "5 行改动" 是合理的，包括：
- Props 接口定义
- 参数解构
- CatalogForm 属性传递
- JobFamilyGroupDetail 中的两个 cardTestId 设置

### 任务分工改进

建议未来类似任务中：
1. 与 E2E 脚本修改**同步进行**（而非脚本先改，代码后改）
2. 在脚本修改时**立即验证**源代码是否满足要求
3. 对组件库（如 Canvas Kit）的调用方式**提前验证**

### 根本原因溯源

当前的失败表明：
- **数据流正确** ✅：Props 可以传递
- **标记语法正确** ✅：data-testid 属性有效
- **问题在应用逻辑** ❌：Modal 根本未显示

这提示应重点检查：
1. React 组件的状态管理（useState/useEffect）
2. Canvas Kit Modal 的初始化和生命周期
3. 页面导航是否打断了组件挂载

---

**下一步执行权限**：需要重启 dev server 或检查源代码是否真的被应用了。建议先确认 dev server 状态，再进行深度调查。

