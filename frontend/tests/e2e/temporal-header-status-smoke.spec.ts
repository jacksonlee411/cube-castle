import { test, expect } from '@playwright/test';
import { setupAuth } from './auth-setup';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

const ORG_CODE = process.env.E2E_ORG_CODE || '1000001';

test('Temporal header shows page and optional status (smoke)', async ({ page }) => {
  // 认证会话，避免跳转登录页
  await setupAuth(page);

  // 使用 baseURL + 相对路径
  await page.goto(`/organizations/${ORG_CODE}/temporal`);

  // 稳定断言：主容器可见（而非依赖文案）
  await expect(page.getByTestId(temporalEntitySelectors.page.wrapper)).toBeVisible({ timeout: 10_000 });

  // 可选弱断言：版本页签可见（使用 testid）
  // await expect(page.getByTestId(temporalEntitySelectors.position.tabId('versions'))).toBeVisible({ timeout: 10_000 });

  // 可选弱断言：存在“当前状态”文案时校验其可见（不强制）
  const maybeStatus = page.getByText('当前状态').first();
  if (await maybeStatus.isVisible({ timeout: 1000 }).catch(() => false)) {
    await expect(maybeStatus).toBeVisible();
  }
});
