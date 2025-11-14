import { test, expect } from '@playwright/test';

async function ping(url: string, timeout = 1500): Promise<boolean> {
  try {
    const res = await fetch(url, { method: 'GET', signal: AbortSignal.timeout(timeout) });
    return res.ok;
  } catch {
    return false;
  }
}

test('Temporal header shows page and optional status (smoke)', async ({ page }) => {
  test.slow();
  const hasServer =
    (await ping('http://localhost:8090/health')) ||
    (await ping('http://localhost:3000', 1000));
  test.skip(!hasServer, 'Server not available; skipping smoke test');

  // 使用一个典型组织详情路径；具体可由 CI 注入种子数据
  await page.goto('http://localhost:3000/organizations/1000001/temporal');
  // 验证页面加载基本元素（标题或标签）
  const title = page.locator('text=组织详情').first();
  const tabs = page.locator('text=版本历史').first();
  await expect(title.or(tabs)).toBeVisible({ timeout: 5000 });

  // 如果存在“当前状态”，可选验证其出现（不强制）
  const maybeStatus = page.locator('text=当前状态').first();
  if (await maybeStatus.isVisible({ timeout: 1000 }).catch(() => false)) {
    await expect(maybeStatus).toBeVisible();
  }
});

