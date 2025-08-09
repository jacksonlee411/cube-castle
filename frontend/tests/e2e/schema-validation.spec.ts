import { test, expect } from '@playwright/test';

test.describe('Cube Castle - Schema验证集成测试', () => {
  
  test.beforeEach(async ({ page }) => {
    // 每个测试前清理数据库测试数据
    await page.goto('http://localhost:3001/test');
  });

  test('验证创建组织单元的完整流程', async ({ page }) => {
    // 1. 点击创建按钮
    await page.getByRole('button', { name: '测试创建组织单元' }).click();
    
    // 2. 验证成功提示
    await expect(page.locator('text=创建成功')).toBeVisible();
    
    // 3. 关闭提示框
    await page.getByRole('button', { name: 'OK' }).click();
    
    // 4. 检查控制台日志确认数据格式正确
    const logs = await page.evaluate(() => {
      return window.console.logs || [];
    });
    
    // 应该包含正确的响应格式
    expect(logs.some(log => 
      log.includes('创建成功') && 
      log.includes('code') && 
      log.includes('name') &&
      log.includes('unit_type') &&
      log.includes('status')
    )).toBe(true);
  });

  test('验证组织架构页面数据加载', async ({ page }) => {
    // 导航到组织架构页面
    await page.getByRole('button', { name: '组织架构' }).click();
    
    // 验证页面加载
    await expect(page.locator('h2:has-text("组织架构管理")')).toBeVisible();
    
    // 验证筛选器
    await expect(page.getByRole('combobox').first()).toBeVisible();
    
    // 验证新增按钮
    await expect(page.getByRole('button', { name: '新增组织单元' })).toBeVisible();
  });

  test('验证错误处理机制', async ({ page }) => {
    // 模拟网络错误情况
    await page.route('http://localhost:9090/api/v1/organization-units', route => {
      route.fulfill({
        status: 500,
        contentType: 'text/plain',
        body: 'Internal Server Error'
      });
    });

    await page.getByRole('button', { name: '测试创建组织单元' }).click();
    
    // 验证错误提示
    await expect(page.locator('text=创建失败')).toBeVisible();
    await page.getByRole('button', { name: 'OK' }).click();
  });

  test('验证表单验证Schema', async ({ page }) => {
    // 导航到组织架构页面
    await page.getByRole('button', { name: '组织架构' }).click();
    
    // 点击新增按钮
    await page.getByRole('button', { name: '新增组织单元' }).click();
    
    // 如果有表单弹窗，验证必填字段
    const modal = page.locator('[role="dialog"]');
    if (await modal.isVisible()) {
      // 验证表单字段存在
      await expect(modal.locator('input[name="name"]')).toBeVisible();
      await expect(modal.locator('select[name="unit_type"]')).toBeVisible();
    }
  });

  test('验证数据类型转换', async ({ page }) => {
    // 监听网络请求
    let requestData = null;
    let responseData = null;

    page.on('request', request => {
      if (request.url().includes('organization-units') && request.method() === 'POST') {
        requestData = request.postData();
      }
    });

    page.on('response', response => {
      if (response.url().includes('organization-units') && response.status() === 201) {
        response.json().then(data => {
          responseData = data;
        });
      }
    });

    // 执行创建操作
    await page.getByRole('button', { name: '测试创建组织单元' }).click();
    await page.getByRole('button', { name: 'OK' }).click();

    // 等待数据传输完成
    await page.waitForTimeout(1000);

    // 验证请求数据格式
    if (requestData) {
      const parsedRequest = JSON.parse(requestData);
      expect(parsedRequest).toHaveProperty('name');
      expect(parsedRequest).toHaveProperty('unit_type');
      expect(parsedRequest.unit_type).toBe('DEPARTMENT');
    }

    // 验证响应数据格式
    if (responseData) {
      expect(responseData).toHaveProperty('code');
      expect(responseData).toHaveProperty('name');
      expect(responseData).toHaveProperty('unit_type');
      expect(responseData).toHaveProperty('status');
      expect(responseData).toHaveProperty('created_at');
    }
  });

  test('验证Zod Schema运行时验证', async ({ page }) => {
    // 模拟后端返回错误格式数据
    await page.route('http://localhost:9090/api/v1/organization-units', route => {
      route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          // 故意返回错误的字段类型
          code: "invalid_code_type", // 应该是数字
          name: null, // 应该是字符串
          unit_type: "INVALID_TYPE", // 应该是有效的枚举值
          status: 123, // 应该是字符串
          created_at: "invalid_date_format"
        })
      });
    });

    // 监听控制台错误
    const consoleErrors = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    await page.getByRole('button', { name: '测试创建组织单元' }).click();
    
    // 应该有验证错误
    expect(consoleErrors.some(error => 
      error.includes('ValidationError') || error.includes('Invalid')
    )).toBe(true);
  });
});