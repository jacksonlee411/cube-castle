import { test, expect } from '@playwright/test';
import { setupAuth } from './auth-setup';

test.describe('认证修复验证', () => {

  test.beforeEach(async ({ page }) => {
    // 设置认证信息到 localStorage（确保 RequireAuth 可以通过验证）
    await setupAuth(page);

    // 导航到组织管理页面
    await page.goto('/organizations');

    // 等待页面加载完成
    await expect(page.getByText('组织架构管理')).toBeVisible({ timeout: 10000 });
  });

  test('验证页面可以成功加载', async ({ page }) => {
    console.log('✅ P1 问题已修复：页面成功加载，不再超时');
    console.log('✅ 认证设置有效：RequireAuth 通过验证');

    // 验证页面URL正确
    expect(page.url()).toContain('/organizations');

    // 验证页面标题存在
    await expect(page.getByText('组织架构管理')).toBeVisible();
  });
});
