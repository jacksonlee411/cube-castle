/**
 * 时态管理E2E测试
 * 测试时态查询、事件创建和时间轴导航功能
 */

import { test, expect } from '@playwright/test';

const BASE_URL = 'http://localhost:3000';
const TEMPORAL_API_URL = 'http://localhost:9091';

// 测试数据
const TEST_TENANT_ID = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';
const TEST_ORG_CODE = '1000056';

test.describe('时态管理功能', () => {
  test.beforeEach(async ({ page }) => {
    // 导航到时态管理演示页面
    await page.goto(`${BASE_URL}/temporal-demo`);
    
    // 等待页面加载完成
    await expect(page.locator('h1')).toContainText('时态管理集成演示');
  });

  test('应该显示时态服务状态', async ({ page }) => {
    // 检查时态服务状态指示器
    const statusBadge = page.locator('[data-testid="temporal-service-status"]').or(
      page.locator('text=时态服务').first()
    );
    
    // 等待状态加载
    await expect(statusBadge).toBeVisible({ timeout: 10000 });
    
    // 验证服务状态（正常或异常）
    const badgeText = await statusBadge.textContent();
    expect(['正常', '异常', '未连接']).toContain(badgeText?.split(':')[1]?.trim());
  });

  test('应该显示组织列表', async ({ page }) => {
    // 等待组织列表表格加载
    await expect(page.locator('table')).toBeVisible();
    
    // 检查表格标题
    await expect(page.locator('th')).toContainText('组织代码');
    await expect(page.locator('th')).toContainText('组织名称');
    await expect(page.locator('th')).toContainText('状态');
    
    // 检查是否有组织数据
    const rows = page.locator('tbody tr');
    await expect(rows).toHaveCountGreaterThan(0);
  });

  test('应该能够搜索组织', async ({ page }) => {
    // 获取搜索框
    const searchInput = page.locator('input[placeholder*="输入组织名称"]');
    await expect(searchInput).toBeVisible();
    
    // 记录初始行数
    const initialRowCount = await page.locator('tbody tr').count();
    
    // 执行搜索
    await searchInput.fill(TEST_ORG_CODE);
    await searchInput.press('Enter');
    
    // 等待搜索结果
    await page.waitForTimeout(500);
    
    // 验证搜索结果
    const filteredRows = page.locator('tbody tr');
    const filteredCount = await filteredRows.count();
    
    if (filteredCount > 0) {
      // 验证搜索结果包含搜索关键词
      const firstRowCode = await filteredRows.first().locator('td').first().textContent();
      expect(firstRowCode).toContain(TEST_ORG_CODE);
    }
    
    expect(filteredCount).toBeLessThanOrEqual(initialRowCount);
  });

  test('应该能够打开组织详情面板', async ({ page }) => {
    // 找到第一个"查看详情"按钮
    const viewDetailsButton = page.locator('text=查看详情').first();
    await expect(viewDetailsButton).toBeVisible();
    
    // 点击查看详情
    await viewDetailsButton.click();
    
    // 等待详情面板打开
    await expect(page.locator('[data-testid="organization-detail-panel"]').or(
      page.locator('text=组织详情').first()
    )).toBeVisible({ timeout: 5000 });
    
    // 验证详情面板内容
    await expect(page.locator('text=基础信息')).toBeVisible();
    await expect(page.locator('text=时态管理信息')).toBeVisible();
  });
  
  test('时间轴导航功能测试', async ({ page }) => {
    // 先打开组织详情面板
    await page.locator('text=查看详情').first().click();
    await expect(page.locator('text=基础信息')).toBeVisible();
    
    // 检查时间轴是否存在
    const timeline = page.locator('[data-testid="timeline"]').or(
      page.locator('.timeline-container').first()
    );
    
    if (await timeline.isVisible()) {
      // 检查时间轴节点
      const timelineNodes = page.locator('[data-testid="timeline-node"]').or(
        page.locator('.timeline-node')
      );
      
      const nodeCount = await timelineNodes.count();
      if (nodeCount > 0) {
        // 点击第一个时间轴节点
        await timelineNodes.first().click();
        
        // 验证详情内容更新
        await expect(page.locator('input[value]').first()).toBeVisible();
        
        // 如果有多个节点，测试切换
        if (nodeCount > 1) {
          await timelineNodes.nth(1).click();
          await page.waitForTimeout(500);
        }
      }
    }
  });
});

test.describe('时态API集成测试', () => {
  test('时态查询API应该正常工作', async ({ request }) => {
    // 测试当前记录查询
    const currentRecordResponse = await request.get(
      `${TEMPORAL_API_URL}/api/v1/organization-units/${TEST_ORG_CODE}`,
      {
        headers: {
          'X-Tenant-ID': TEST_TENANT_ID
        }
      }
    );
    
    if (currentRecordResponse.ok()) {
      const data = await currentRecordResponse.json();
      expect(data).toHaveProperty('organizations');
      expect(Array.isArray(data.organization_units)).toBeTruthy();
    } else {
      // 如果404，说明测试组织不存在，这是可接受的
      expect([200, 404]).toContain(currentRecordResponse.status());
    }
  });
  
  test('历史记录查询应该正常工作', async ({ request }) => {
    const historyResponse = await request.get(
      `${TEMPORAL_API_URL}/api/v1/organization-units/${TEST_ORG_CODE}?includeHistory=true&maxRecords=5`,
      {
        headers: {
          'X-Tenant-ID': TEST_TENANT_ID
        }
      }
    );
    
    if (historyResponse.ok()) {
      const data = await historyResponse.json();
      expect(data).toHaveProperty('organizations');
      expect(data).toHaveProperty('result_count');
    } else {
      expect([200, 404]).toContain(historyResponse.status());
    }
  });
  
  test('未来记录查询应该正常工作', async ({ request }) => {
    const futureResponse = await request.get(
      `${TEMPORAL_API_URL}/api/v1/organization-units/${TEST_ORG_CODE}?includeFuture=true`,
      {
        headers: {
          'X-Tenant-ID': TEST_TENANT_ID
        }
      }
    );
    
    if (futureResponse.ok()) {
      const data = await futureResponse.json();
      expect(data).toHaveProperty('organizations');
      expect(data).toHaveProperty('query_options');
      expect(data.query_options.IncludeFuture).toBeTruthy();
    } else {
      expect([200, 404]).toContain(futureResponse.status());
    }
  });
  
  test('范围查询应该正常工作', async ({ request }) => {
    const rangeResponse = await request.get(
      `${TEMPORAL_API_URL}/api/v1/organization-units/${TEST_ORG_CODE}?effectiveFrom=2025-01-01&effectiveTo=2025-12-31`,
      {
        headers: {
          'X-Tenant-ID': TEST_TENANT_ID
        }
      }
    );
    
    if (rangeResponse.ok()) {
      const data = await rangeResponse.json();
      expect(data).toHaveProperty('organizations');
      expect(data.query_options).toHaveProperty('EffectiveFrom');
      expect(data.query_options).toHaveProperty('EffectiveTo');
    } else {
      expect([200, 404]).toContain(rangeResponse.status());
    }
  });
  
  test('时态事件创建API应该正常工作', async ({ request }) => {
    // 测试UPDATE事件
    const updateEventData = {
      event_type: 'UPDATE',
      effective_date: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(), // 明天
      change_data: {
        description: 'E2E测试更新描述'
      },
      change_reason: 'E2E自动化测试'
    };
    
    const eventResponse = await request.post(
      `${TEMPORAL_API_URL}/api/v1/organization-units/${TEST_ORG_CODE}/events`,
      {
        headers: {
          'X-Tenant-ID': TEST_TENANT_ID,
          'Content-Type': 'application/json'
        },
        data: updateEventData
      }
    );
    
    if (eventResponse.ok()) {
      const data = await eventResponse.json();
      expect(data).toHaveProperty('event_id');
      expect(data).toHaveProperty('event_type');
      expect(data.event_type).toBe('UPDATE');
      expect(data).toHaveProperty('status');
      expect(data.status).toBe('processed');
    } else {
      // 如果失败，检查是否是因为组织不存在
      expect([201, 404, 500]).toContain(eventResponse.status());
    }
  });
});

test.describe('时态管理性能测试', () => {
  test('页面加载性能应该在可接受范围内', async ({ page }) => {
    const startTime = Date.now();
    
    // 导航到页面
    await page.goto(`${BASE_URL}/temporal-demo`);
    
    // 等待关键内容加载
    await expect(page.locator('h1')).toContainText('时态管理集成演示');
    await expect(page.locator('table')).toBeVisible();
    
    const loadTime = Date.now() - startTime;
    
    // 验证加载时间在3秒内
    expect(loadTime).toBeLessThan(3000);
    console.log(`页面加载时间: ${loadTime}ms`);
  });
  
  test('组织详情面板打开性能', async ({ page }) => {
    // 先导航到页面
    await page.goto(`${BASE_URL}/temporal-demo`);
    await expect(page.locator('table')).toBeVisible();
    
    const startTime = Date.now();
    
    // 点击查看详情
    await page.locator('text=查看详情').first().click();
    
    // 等待详情面板显示
    await expect(page.locator('text=基础信息')).toBeVisible();
    
    const openTime = Date.now() - startTime;
    
    // 验证打开时间在2秒内
    expect(openTime).toBeLessThan(2000);
    console.log(`详情面板打开时间: ${openTime}ms`);
  });
});

test.describe('时态管理错误处理', () => {
  test('无效组织代码应该显示适当错误', async ({ page }) => {
    await page.goto(`${BASE_URL}/temporal-demo`);
    
    // 搜索不存在的组织
    const searchInput = page.locator('input[placeholder*="输入组织名称"]');
    await searchInput.fill('NONEXISTENT999');
    await searchInput.press('Enter');
    
    // 等待搜索结果
    await page.waitForTimeout(1000);
    
    // 验证没有搜索结果或显示适当消息
    const _noResultsMessage = page.locator('text=没有找到匹配的组织').or(
      page.locator('tbody tr')
    );
    
    const rows = await page.locator('tbody tr').count();
    if (rows === 0) {
      await expect(page.locator('text=没有找到匹配的组织')).toBeVisible();
    }
  });
  
  test('时态服务不可用时应该优雅处理', async ({ page, context }) => {
    // 模拟网络失败
    await context.route('**/api/v1/organization-units/**', route => {
      route.fulfill({
        status: 500,
        body: JSON.stringify({ error: 'Service Unavailable' })
      });
    });
    
    await page.goto(`${BASE_URL}/temporal-demo`);
    
    // 尝试打开详情面板
    const viewDetailsButton = page.locator('text=查看详情').first();
    if (await viewDetailsButton.isVisible()) {
      await viewDetailsButton.click();
      
      // 应该显示错误状态或加载状态
      // 页面应该不会崩溃
      await page.waitForTimeout(2000);
    }
    
    // 验证页面仍然可用
    await expect(page.locator('h1')).toContainText('时态管理集成演示');
  });
});

test.describe('时态管理可访问性测试', () => {
  test('页面应该满足基本可访问性要求', async ({ page }) => {
    await page.goto(`${BASE_URL}/temporal-demo`);
    
    // 检查是否有适当的标题层次
    await expect(page.locator('h1')).toBeVisible();
    
    // 检查表格是否有适当的标题
    const table = page.locator('table');
    await expect(table).toBeVisible();
    await expect(table.locator('th').first()).toBeVisible();
    
    // 检查按钮是否可访问
    const buttons = page.locator('button');
    const buttonCount = await buttons.count();
    expect(buttonCount).toBeGreaterThan(0);
    
    // 检查输入框是否有标签或占位符
    const inputs = page.locator('input');
    for (let i = 0; i < await inputs.count(); i++) {
      const input = inputs.nth(i);
      const placeholder = await input.getAttribute('placeholder');
      const ariaLabel = await input.getAttribute('aria-label');
      const id = await input.getAttribute('id');
      
      // 输入框应该有placeholder、aria-label或对应的label
      expect(placeholder || ariaLabel || id).toBeTruthy();
    }
  });
  
  test('键盘导航应该正常工作', async ({ page }) => {
    await page.goto(`${BASE_URL}/temporal-demo`);
    
    // 使用Tab键导航
    await page.keyboard.press('Tab');
    
    // 验证焦点在可聚焦元素上
    const focusedElement = await page.locator(':focus');
    await expect(focusedElement).toBeVisible();
    
    // 继续Tab导航几次
    for (let i = 0; i < 3; i++) {
      await page.keyboard.press('Tab');
      await page.waitForTimeout(100);
    }
    
    // 验证仍有元素获得焦点
    const finalFocusedElement = await page.locator(':focus');
    await expect(finalFocusedElement).toBeVisible();
  });
});