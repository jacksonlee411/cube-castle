import { test, expect } from '@playwright/test';
import { setupAuth } from './auth-setup';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

test.describe('Organization Create Flow', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page);
    await page.goto('/organizations');
    await expect(page.getByTestId(temporalEntitySelectors.organization.dashboard)).toBeVisible({ timeout: 15000 });
  });

  test('allows selecting parent organization before submitting create request', async ({ page }) => {
    await page.getByTestId(temporalEntitySelectors.organization.createButton).click();
    await expect(page).toHaveURL(/\/organizations\/new$/);
    await expect(page.getByTestId(temporalEntitySelectors.organization.form)).toBeVisible();

    const parentInput = page.getByTestId('combobox-input');
    await parentInput.click();

    const parentMenu = page.getByTestId('combobox-menu');
    await expect(parentMenu).toBeVisible();

    const firstItem = parentMenu.locator('[data-testid^="combobox-item-"]').first();
    await expect(firstItem).toBeVisible();
    const parentCode = await firstItem.getAttribute('data-testid');
    expect(parentCode).toBeTruthy();

    await firstItem.click();
    await expect(parentInput).not.toHaveValue('');

    // 填写必填项，确认表单可提交
    const nameInput = page.getByTestId('form-field-name');
    await nameInput.fill(`自动化创建-${Date.now()}`);
    await page.getByTestId('form-field-effective-date').fill(new Date().toISOString().slice(0, 10));

    const submitButton = page.getByTestId('form-submit-button');
    await expect(submitButton).toBeEnabled();

    // 不实际提交，回到列表
    await page.goto('/organizations');
    await expect(page.getByTestId(temporalEntitySelectors.organization.dashboard)).toBeVisible();
  });
});
