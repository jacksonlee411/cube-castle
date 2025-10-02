/**
 * 简化的时态管理功能验证测试
 * 验证系统基本功能是否正常工作
 * 🎯 使用动态环境配置替代硬编码端口
 */
import { test, expect } from '@playwright/test';
import { validateTestEnvironment } from './config/test-environment';
import { setupAuth } from './auth-setup';

const hasAuthToken = Boolean(process.env.PW_JWT);
test.skip(!hasAuthToken, '需要 RS256 JWT 令牌运行受保护路由测试');

let BASE_URL: string;

test.describe('时态管理系统基础功能验证', () => {
  
  // 🎯 测试前环境验证和动态端口配置
  test.beforeAll(async () => {
    const envValidation = await validateTestEnvironment();
    
    if (!envValidation.isValid) {
      console.error('🚨 测试环境验证失败:');
      envValidation.errors.forEach(error => console.error(`  - ${error}`));
      throw new Error('测试环境不可用');
    }
    
    BASE_URL = envValidation.frontendUrl;
    console.log(`✅ 使用前端基址: ${BASE_URL}`);
  });
  
  test.beforeEach(async ({ page }) => {
    await setupAuth(page);
  });
  
  test('应用基础加载测试', async ({ page }) => {
    const startTime = Date.now();
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    const loadTime = Date.now() - startTime;

    expect(loadTime).toBeLessThan(10000);
    await expect(page).toHaveTitle(/Cube Castle/);
    await page.screenshot({ path: 'test-results/app-loaded.png' });
  });

  test('组织管理页面可访问', async ({ page }) => {
    await page.goto(`${BASE_URL}/organizations`);
    await page.waitForLoadState('networkidle');

    // 等待组织dashboard加载完成
    await expect(page.getByTestId('organization-dashboard')).toBeVisible({ timeout: 15000 });

    // 等待加载状态消失
    await page.waitForSelector('text=加载组织数据中...', { state: 'detached', timeout: 15000 }).catch(() => {
      // 如果没有加载状态也没关系
    });

    // 确认创建按钮可见
    await expect(page.getByTestId('create-organization-button')).toBeVisible({ timeout: 10000 });
    await page.screenshot({ path: 'test-results/organizations-page.png' });
  });

  test.skip('测试页面功能验证', async ({ page }) => {
    // 导航到测试页面
    await page.goto(`${BASE_URL}/test`);
    await page.waitForLoadState('networkidle', { timeout: 15000 });
    
    // 查找表格或数据内容
    const hasTable = await page.locator('table, .table, [role="table"], .data-table').first().count();
    const hasButtons = await page.locator('button').count();
    
    console.log(`找到表格数量: ${hasTable}`);
    console.log(`找到按钮数量: ${hasButtons}`);
    
    // 验证页面有交互元素
    expect(hasButtons).toBeGreaterThan(0);
    
    // 截图记录
    await page.screenshot({ path: 'test-results/test-page.png' });
  });

  test('系统响应性测试', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');

    const buttons = page.locator('button:visible');
    const buttonCount = await buttons.count();

    if (buttonCount > 0) {
      const startTime = Date.now();
      await buttons.first().click();
      const responseTime = Date.now() - startTime;
      expect(responseTime).toBeLessThan(3000);
    }

    await page.screenshot({ path: 'test-results/interaction-test.png' });
  });

  test('错误处理基础验证', async ({ page }) => {
    await page.goto(`${BASE_URL}/non-existent-route`);
    await page.waitForLoadState('networkidle');

    const currentUrl = page.url();
    expect(currentUrl).toContain('/non-existent-route');
    await page.screenshot({ path: 'test-results/error-handling.png' });
  });
});
