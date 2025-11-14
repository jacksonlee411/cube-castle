import { test, expect } from '@playwright/test';
import { validateTestEnvironment } from './config/test-environment';
import { setupAuth } from './auth-setup';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

const TEST_ORGANIZATION_CODE = '1000004';

let baseUrl: string;

test.describe('五状态生命周期管理系统', () => {
  test.beforeAll(async () => {
    const envValidation = await validateTestEnvironment();
    if (!envValidation.isValid) {
      throw new Error(`测试环境不可用: ${envValidation.errors.join(', ')}`);
    }
    baseUrl = envValidation.frontendUrl;
  });

  test.beforeEach(async ({ page }) => {
    await setupAuth(page);
    await page.goto(`${baseUrl}/organizations/${TEST_ORGANIZATION_CODE}/temporal`);
    await page.waitForSelector('[data-testid="temporal-master-detail-view"]', { timeout: 15_000 });
  });

  test('加载后显示时间轴及当前版本信息', async ({ page }) => {
    const timeline = page.locator('[data-testid="temporal-timeline"]');
    await expect(timeline).toBeVisible();

    const nodes = timeline.locator('[data-testid="temporal-timeline-node"]');
    await expect(nodes.first()).toBeVisible({ timeout: 15_000 });

    const nodeCount = await nodes.count();
    expect(nodeCount).toBeGreaterThan(0);

    const currentNode = nodes.filter({
      has: page.locator('[data-testid="temporal-lifecycle-badge"][data-lifecycle="CURRENT"]'),
    }).first();
    await expect(currentNode).toBeVisible();

    await expect(page.getByTestId(temporalEntitySelectors.organization.form)).toBeVisible();
  });

  test('支持选择时间轴节点并进入编辑模式', async ({ page }) => {
    const nodes = page.locator('[data-testid="temporal-timeline-node"]');
    await nodes.first().click();
    await expect(nodes.first()).toHaveAttribute('data-current', 'true');

    const editButton = page.locator('[data-testid="edit-history-toggle-button"]');
    await expect(editButton).toBeVisible();
    await editButton.click();

    const submitButton = page.locator('[data-testid="submit-edit-history-button"]');
    await expect(submitButton).toBeVisible();

    const nameInput = page.locator('[data-testid="form-field-name"]');
    await expect(nameInput).toBeEditable();
  });

  test('可以启动插入新版本流程并触发校验', async ({ page }) => {
    const insertButton = page.locator('[data-testid="start-insert-version-button"]');
    await expect(insertButton).toBeVisible();
    await insertButton.click();

    await expect(page.getByRole('heading', { name: '插入新版本记录' })).toBeVisible();

    await page.fill('[data-testid="form-field-name"]', '');
    await page.fill('[data-testid="form-field-effective-date"]', '');

    const submitButton = page.locator('[data-testid="submit-edit-history-button"]');
    await submitButton.click();

    const errorMessage = page.locator('[data-testid="temporal-form-error"]');
    await expect(errorMessage).toBeVisible();
    await expect(errorMessage).toContainText('错误项');
  });

  test('可切换到审计历史并显示记录提示', async ({ page }) => {
    const nodes = page.locator('[data-testid="temporal-timeline-node"]');
    await nodes.first().click();

    await page.getByText('审计历史', { exact: true }).click();

    const debugInfo = page.locator('text=🔍 调试信息');
    await expect(debugInfo).toBeVisible();
  });
});
