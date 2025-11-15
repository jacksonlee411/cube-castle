import { test, expect } from '@playwright/test';
import { setupAuth } from './auth-setup';
import { waitForPageReady } from './utils/waitPatterns';

const ORG_CODE = process.env.E2E_ORG_CODE || '1000000';

test.describe('组织详情 Smoke', () => {
  test('可以打开组织详情页面（Temporal）', async ({ page }) => {
    await setupAuth(page);
    const url = `/organizations/${ORG_CODE}/temporal`;
    await page.goto(url, { waitUntil: 'networkidle' });
    await waitForPageReady(page, { timeout: 20000 });

    // 验证页面骨架（Master-Detail 容器）可见，代表组织详情已打开
    await expect(page.getByTestId('temporal-master-detail-view')).toBeVisible({ timeout: 20000 });
  });
});
