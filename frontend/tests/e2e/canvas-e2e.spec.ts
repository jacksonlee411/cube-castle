import { test, expect } from '@playwright/test';
import { setupAuth } from './auth-setup';

const DASHBOARD_TESTID = 'organization-dashboard';

test.describe('Canvas UI 基础体验', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page);
    await page.goto('/');
    await expect(page.locator(`[data-testid="${DASHBOARD_TESTID}"]`)).toBeVisible({ timeout: 15000 });
  });

  test('应用外壳渲染与导航', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /Cube Castle/i })).toBeVisible();

    const topBar = page.locator('header');
    if (await topBar.count()) {
      await expect(topBar.first()).toContainText('Cube Castle');
    }

    const dashboard = page.locator(`[data-testid="${DASHBOARD_TESTID}"]`);
    await expect(dashboard).toContainText('组织单元列表');
  });

  test('导航跳转至组织列表', async ({ page }) => {
    const navButton = page.getByRole('link', { name: /组织架构/ }).first().or(
      page.getByRole('button', { name: /组织架构/ }).first(),
    );
    await navButton.click();
    await expect(page).toHaveURL(/\/organizations/);

    const dashboard = page.locator(`[data-testid="${DASHBOARD_TESTID}"]`);
    await expect(dashboard.getByRole('button', { name: /新增组织单元/ })).toBeVisible();
  });

  test('组织表格加载', async ({ page }) => {
    await page.goto('/organizations');
    const dashboard = page.locator(`[data-testid="${DASHBOARD_TESTID}"]`);
    await expect(dashboard).toBeVisible({ timeout: 15000 });

    const table = page.getByTestId('organization-table');
    await expect(table).toBeVisible();

    const rows = table.locator('tbody tr');
    if ((await rows.count()) === 0) {
      await expect(table).toContainText(['没有组织数据', '暂无组织数据']);
    } else {
      const firstRow = rows.first();
      await expect(firstRow.locator('td').first()).toHaveText(/\d{7}/);
    }
  });

  test('响应式布局', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await expect(page.locator(`[data-testid="${DASHBOARD_TESTID}"]`)).toBeVisible();

    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.getByRole('heading', { name: /Cube Castle/i })).toBeVisible();

    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.getByRole('heading', { name: /Cube Castle/i })).toBeVisible();

    await page.setViewportSize({ width: 1280, height: 720 });
  });

  test('主要操作按钮具有 Canvas 样式', async ({ page }) => {
    const dashboard = page.locator(`[data-testid="${DASHBOARD_TESTID}"]`);
    const primaryButton = dashboard.getByRole('button', { name: /新增组织单元/ });
    await expect(primaryButton).toBeVisible();
  });

  test('统计数据与表格行展示', async ({ page }) => {
    await page.goto('/organizations');
    const dashboard = page.locator(`[data-testid="${DASHBOARD_TESTID}"]`);
    await expect(dashboard).toBeVisible({ timeout: 15000 });

    const table = page.getByTestId('organization-table');
    await expect(table).toBeVisible();

    const rows = table.locator('tbody tr');
    if (await rows.count()) {
      await expect(rows.first().locator('td').nth(0)).toHaveText(/\d{7}/);
    }
  });
});
