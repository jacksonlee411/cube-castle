/**
 * 简化的时态管理功能验证测试
 * 验证系统基本功能是否正常工作
 * 🎯 使用动态环境配置替代硬编码端口
 */
import { test, expect } from '@playwright/test';
import { validateTestEnvironment } from './config/test-environment';

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
  
  test('应用基础加载测试', async ({ page }) => {
    // 导航到应用
    const startTime = Date.now();
    await page.goto(BASE_URL);
    const loadTime = Date.now() - startTime;
    
    // 验证页面加载时间
    expect(loadTime).toBeLessThan(10000); // 10秒超时
    console.log(`页面加载时间: ${loadTime}ms`);
    
    // 等待页面加载完成
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    
    // 验证页面标题
    await expect(page).toHaveTitle(/Cube Castle/);
    
    // 截图记录
    await page.screenshot({ path: 'test-results/app-loaded.png' });
  });

  test('组织管理页面可访问', async ({ page }) => {
    // 导航到组织管理页面
    await page.goto(`${BASE_URL}/organizations`);
    await page.waitForLoadState('networkidle', { timeout: 15000 });
    
    // 查找页面内容
    const hasContent = await page.locator('h1, h2, [data-testid], .organization, .temporal').first().count();
    expect(hasContent).toBeGreaterThan(0);
    
    // 截图记录
    await page.screenshot({ path: 'test-results/organizations-page.png' });
  });

  test('测试页面功能验证', async ({ page }) => {
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
    
    // 查找可点击的按钮
    const buttons = page.locator('button:visible');
    const buttonCount = await buttons.count();
    
    if (buttonCount > 0) {
      const startTime = Date.now();
      await buttons.first().click();
      const responseTime = Date.now() - startTime;
      
      // 验证响应时间
      expect(responseTime).toBeLessThan(3000);
      console.log(`按钮响应时间: ${responseTime}ms`);
      
      // 等待可能的UI变化
      await page.waitForTimeout(1000);
    }
    
    // 截图记录
    await page.screenshot({ path: 'test-results/interaction-test.png' });
  });

  test('错误处理基础验证', async ({ page }) => {
    // 测试不存在的路由
    await page.goto(`${BASE_URL}/non-existent-route`);
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    
    // 应该有某种错误处理或重定向
    const url = page.url();
    console.log(`当前URL: ${url}`);
    
    // 截图记录
    await page.screenshot({ path: 'test-results/error-handling.png' });
  });
});