/**
 * Playwright E2E 测试认证设置辅助函数
 *
 * 用途：为测试页面注入 localStorage 认证信息，确保 RequireAuth 组件可以正常验证
 *
 * 使用方式：
 * ```typescript
 * import { setupAuth } from './auth-setup';
 *
 * test.beforeEach(async ({ page }) => {
 *   await setupAuth(page);
 *   await page.goto('/organizations');
 * });
 * ```
 */

import { Page } from '@playwright/test';

export async function setupAuth(page: Page): Promise<void> {
  // 从环境变量获取 JWT token
  const token = process.env.PW_JWT;
  const tenantId = process.env.PW_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';

  if (!token) {
    console.warn('⚠️  PW_JWT 环境变量未设置，测试可能无法访问受保护路由');
    return;
  }

  // 先导航到基础URL建立上下文,然后注入localStorage
  // 这确保localStorage在正确的域下设置
  await page.goto('/');

  // 直接在页面上下文中设置 localStorage
  await page.evaluate((authData) => {
    // 设置 OAuth token（前端 authManager 期望的键名和格式）
    // 参考：frontend/src/shared/api/auth.ts:327 - localStorage.getItem('cube_castle_oauth_token')
    localStorage.setItem('cube_castle_oauth_token', JSON.stringify({
      accessToken: authData.token,
      tokenType: 'Bearer',
      expiresIn: 86400, // 24小时有效期（秒）
      issuedAt: Date.now() // 当前时间戳
    }));

    // 设置租户信息（如果前端需要）
    localStorage.setItem('tenant_id', authData.tenantId);
  }, { token, tenantId });

  console.log('✅ 认证设置已注入 localStorage');
}

/**
 * 清除认证信息（用于测试登出场景）
 */
export async function clearAuth(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.removeItem('cube_castle_oauth_token');
    localStorage.removeItem('tenant_id');
  });

  console.log('✅ 认证信息已清除');
}
