import { test, expect } from '@playwright/test';

async function ping(url: string, timeout = 1500): Promise<boolean> {
  try {
    const res = await fetch(url, { method: 'GET', signal: AbortSignal.timeout(timeout) });
    return res.ok;
  } catch {
    return false;
  }
}

test('Temporal edit form creation page renders (smoke)', async ({ page }) => {
  test.slow();
  const hasServer =
    (await ping('http://localhost:8090/health')) ||
    (await ping('http://localhost:3000', 1000));
  test.skip(!hasServer, 'Server not available; skipping smoke test');

  await page.goto('http://localhost:3000/organizations/new/temporal');

  // 创建模式标题或提示文案
  const createTitle = page.locator('text=创建新组织').first();
  await expect(createTitle).toBeVisible({ timeout: 5000 });

  // 存在表单主体
  const form = page.locator('[data-testid="organization-form"]').first();
  await expect(form).toBeVisible();
});

