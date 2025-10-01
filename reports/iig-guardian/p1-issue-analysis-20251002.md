# P1 问题根因分析（2025-10-02）

## 问题1: 业务流程页面加载超时

### 症状
- Playwright 测试访问 `/organizations` 页面
- 等待 `'组织架构管理'` 文本出现
- 120秒超时，文本未找到

### 根因
**前端认证状态缺失** - Playwright 测试未在浏览器 localStorage 中设置认证 token

### 技术细节

1. **路由配置正确** (`App.tsx:41-50`):
   ```tsx
   <Route path="/organizations" element={
     <RequireAuth>
       <Suspense fallback={<SuspenseLoader />}>
         <OrganizationDashboard />
       </Suspense>
     </RequireAuth>
   } />
   ```

2. **组件包含目标文本** (`OrganizationDashboard.tsx:30`):
   ```tsx
   <Heading size="large">组织架构管理</Heading>
   ```

3. **问题点**: `RequireAuth` 组件检查认证状态
   - 前端应用从 localStorage 读取 token
   - Playwright HTTP 头部认证（`playwright.config.ts:25-28`）不影响客户端存储
   - 未认证用户被重定向到登录页或显示空白页

### 解决方案

**方案A: 为 Playwright 测试设置 localStorage**

创建全局设置（`tests/e2e/global-setup.ts`）:
```typescript
import { chromium } from '@playwright/test';

async function globalSetup() {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  
  // 设置认证 token 到 localStorage
  await page.context().addInitScript(() => {
    const token = process.env.PW_JWT;
    if (token) {
      localStorage.setItem('oauth_token', JSON.stringify({
        access_token: token,
        token_type: 'Bearer',
        expires_in: 86400
      }));
    }
  });
  
  await browser.close();
}

export default globalSetup;
```

**方案B: 在每个测试的 beforeEach 中设置认证**

修改 `business-flow-e2e.spec.ts`:
```typescript
test.beforeEach(async ({ page }) => {
  // 先设置认证
  await page.addInitScript(() => {
    const token = process.env.PW_JWT;
    if (token) {
      localStorage.setItem('oauth_token', JSON.stringify({
        access_token: token,
        token_type: 'Bearer',
        expires_in: 86400
      }));
    }
  });
  
  await page.goto('/organizations');
  await expect(page.getByText('组织架构管理')).toBeVisible();
});
```

**推荐**: 方案A（全局设置）+ 确保所有测试都能复用

---

## 问题2: 测试页面交互元素缺失

### 症状
- `basic-functionality-test.spec.ts:60` 测试失败
- `hasButtons = 0`（期望 > 0）

### 可能根因
1. 测试页面路由不存在或配置错误
2. 测试页面组件未加载
3. 同样的认证问题导致组件未渲染

### 需要检查
- 测试页面路由配置
- 测试页面组件实现
- 是否也需要认证

---

## 优先级

1. **高优**: 修复认证问题（影响所有需要认证的测试）
2. **中优**: 排查测试页面问题（影响较小）

---

**分析时间**: 2025-10-02 00:05 UTC+8
**责任团队**: 前端团队 + QA
