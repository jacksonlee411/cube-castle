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
import { TOKEN_STORAGE_KEY } from '@/shared/api/auth';
import { ensurePwJwt, getPwJwt, isJwtNearlyExpired } from './utils/authToken';

export async function setupAuth(page: Page): Promise<void> {
  const tenantId = process.env.PW_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';

  let token = getPwJwt();
  if (!token || isJwtNearlyExpired(token)) {
    token = await ensurePwJwt({ tenantId });
  }

  if (!token) {
    throw new Error('无法获取 RS256 开发令牌，请确认命令服务已启动并执行 make run-dev');
  }

  // 先导航到基础 URL 以便 localStorage 写入正确域
  await page.goto('/');

  // 仅通过 localStorage 注入认证信息，让前端统一客户端负责附加 Authorization / X-Tenant-ID
  await page.evaluate(({ tokenStorageKey, legacyKey, authData }) => {
    localStorage.setItem(tokenStorageKey, JSON.stringify({
      accessToken: authData.token,
      tokenType: 'Bearer',
      expiresIn: 86400,
      issuedAt: Date.now(),
    }));
    localStorage.removeItem(legacyKey);
    localStorage.setItem('tenant_id', authData.tenantId);
  }, {
    tokenStorageKey: TOKEN_STORAGE_KEY,
    legacyKey: ['cube', 'castle', 'oauth', 'token'].join('_'),
    authData: { token, tenantId },
  });

  console.log('✅ 认证设置已注入 localStorage');
}

/**
 * 清除认证信息（用于测试登出场景）
 */
export async function clearAuth(page: Page): Promise<void> {
  await page.evaluate(({ tokenStorageKey, legacyKey }) => {
    localStorage.removeItem(tokenStorageKey);
    localStorage.removeItem(legacyKey);
    localStorage.removeItem('tenant_id');
  }, {
    tokenStorageKey: TOKEN_STORAGE_KEY,
    legacyKey: ['cube', 'castle', 'oauth', 'token'].join('_')
  });

  console.log('✅ 认证信息已清除');
}
