import { test, expect } from '@playwright/test';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

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
  const base = (process.env.PW_BASE_URL || '').replace(/\/+$/, '');
  const hasServer =
    (!!base && (await ping(`${base}/health`))) ||
    (await ping('/', 1000));
  test.skip(!hasServer, 'Server not available; skipping smoke test');

  await page.goto('/organizations/new/temporal');

  // 创建模式标题或提示文案
  const createTitle = page.locator('text=创建新组织').first();
  await expect(createTitle).toBeVisible({ timeout: 5000 });

  // 存在表单主体
  const form = page.getByTestId(temporalEntitySelectors.organization.form);
  await expect(form).toBeVisible();
});
