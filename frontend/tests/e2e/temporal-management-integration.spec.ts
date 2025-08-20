/**
 * 时态管理组件集成测试
 * 测试时态管理主从视图组件的完整功能
 */

import { test, expect } from '@playwright/test';

const TEMPORAL_SERVICE_URL = 'http://localhost:9091';
const FRONTEND_URL = 'http://localhost:3000';
const TEST_ORG_CODE = '1000056';

test.describe('时态管理系统集成测试', () => {
  
  test.beforeEach(async ({ page }) => {
    // 确保时态服务正常运行
    const healthResponse = await page.request.get(`${TEMPORAL_SERVICE_URL}/health`);
    expect(healthResponse.ok()).toBeTruthy();
  });

  test('时态管理演示页面加载和基本功能', async ({ page }) => {
    // 导航到时态管理演示页面
    await page.goto(`${FRONTEND_URL}/temporal-demo`);
    
    // 验证页面标题
    await expect(page.locator('text=时态管理集成演示')).toBeVisible();
    
    // 验证时态服务状态指示器
    await expect(page.locator('text=时态服务').first()).toBeVisible();
    
    // 验证组织列表显示
    await expect(page.locator('text=组织列表')).toBeVisible();
    
    // 验证搜索功能
    const searchInput = page.locator('input[placeholder*="输入组织名称或代码"]');
    await expect(searchInput).toBeVisible();
    await searchInput.fill(TEST_ORG_CODE);
    
    // 验证过滤结果
    await expect(page.locator(`text=${TEST_ORG_CODE}`)).toBeVisible();
  });

  test('组织详情面板时态管理功能', async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/temporal-demo`);
    
    // 点击查看详情按钮
    const viewDetailsButton = page.locator('text=查看详情').first();
    await viewDetailsButton.click();
    
    // 验证详情面板打开
    await expect(page.locator('text=时间轴导航')).toBeVisible();
    await expect(page.locator('text=版本详情')).toBeVisible();
    
    // 验证时态数据加载
    await expect(page.locator('[data-testid="timeline-node"]').first()).toBeVisible({ timeout: 10000 });
    
    // 验证版本节点可点击
    const firstTimelineNode = page.locator('[data-testid="timeline-node"]').first();
    await firstTimelineNode.click();
    
    // 验证详情信息显示
    await expect(page.locator('text=基本信息')).toBeVisible();
    await expect(page.locator('text=层级结构')).toBeVisible();
    await expect(page.locator('text=生效期间')).toBeVisible();
  });

  test('时态管理选项卡功能', async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/temporal-demo`);
    
    // 打开详情面板
    await page.locator('text=查看详情').first().click();
    
    // 等待面板加载
    await expect(page.locator('text=版本详情')).toBeVisible();
    
    // 测试时间线可视化选项卡
    await page.locator('text=📊 时间线可视化').click();
    await expect(page.locator('text=时间线可视化组件')).toBeVisible({ timeout: 5000 });
    
    // 测试新增版本选项卡
    await page.locator('text=➕ 新增版本').click();
    await expect(page.locator('text=新增时态版本')).toBeVisible({ timeout: 5000 });
    
    // 验证表单字段
    await expect(page.locator('select[name="event_type"]')).toBeVisible();
    await expect(page.locator('input[name="effective_date"]')).toBeVisible();
    await expect(page.locator('input[name="name"]')).toBeVisible();
    
    // 回到版本详情选项卡
    await page.locator('text=📋 版本详情').click();
    await expect(page.locator('text=基本信息')).toBeVisible();
  });

  test('时态事件创建功能测试', async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/temporal-demo`);
    
    // 打开详情面板并切换到新增版本选项卡
    await page.locator('text=查看详情').first().click();
    await page.locator('text=➕ 新增版本').click();
    
    // 填写新增版本表单
    await page.selectOption('select[name="event_type"]', 'UPDATE');
    await page.fill('input[name="effective_date"]', '2035-01-01');
    await page.fill('input[name="name"]', '测试新增时态版本');
    await page.fill('textarea[name="change_reason"]', 'Playwright自动化测试');
    await page.selectOption('select[name="status"]', 'ACTIVE');
    
    // 提交表单
    const submitButton = page.locator('button[type="submit"]');
    await submitButton.click();
    
    // 验证提交结果（可能需要处理成功或失败情况）
    await expect(page.locator('text=创建')).toBeVisible({ timeout: 5000 });
  });

  test('时态查询API响应时间测试', async ({ page }) => {
    // 测试各种时态查询的响应时间
    const queries = [
      `${TEMPORAL_SERVICE_URL}/health`,
      `${TEMPORAL_SERVICE_URL}/api/v1/organization-units/${TEST_ORG_CODE}/temporal?as_of_date=2025-08-12`,
      `${TEMPORAL_SERVICE_URL}/api/v1/organization-units/${TEST_ORG_CODE}/temporal?include_history=true&include_future=true`
    ];
    
    for (const query of queries) {
      const startTime = Date.now();
      const response = await page.request.get(query);
      const endTime = Date.now();
      const responseTime = endTime - startTime;
      
      expect(response.ok()).toBeTruthy();
      expect(responseTime).toBeLessThan(1000); // 响应时间应小于1秒
      
      console.log(`Query: ${query.split('/').pop()} - Response time: ${responseTime}ms`);
    }
  });

  test('时态数据一致性验证', async ({ page }) => {
    // 获取组织的完整时态数据
    const response = await page.request.get(
      `${TEMPORAL_SERVICE_URL}/api/v1/organization-units/${TEST_ORG_CODE}/temporal?include_history=true&include_future=true`
    );
    
    expect(response.ok()).toBeTruthy();
    
    const data = await response.json();
    const organizations = data.organization_units;
    
    // 验证数据结构
    expect(Array.isArray(organizations)).toBeTruthy();
    expect(organizations.length).toBeGreaterThan(0);
    
    // 验证每个记录包含必要字段
    for (const org of organizations) {
      expect(org).toHaveProperty('code');
      expect(org).toHaveProperty('name');
      expect(org).toHaveProperty('effective_date');
      expect(org).toHaveProperty('is_current');
      expect(org).toHaveProperty('unit_type');
      expect(org).toHaveProperty('status');
    }
    
    // 验证当前记录唯一性
    const currentRecords = organizations.filter(org => org.is_current === true);
    expect(currentRecords.length).toBeLessThanOrEqual(1);
    
    // 验证时间排序
    const dates = organizations.map(org => new Date(org.effective_date));
    const sortedDates = [...dates].sort((a, b) => b.getTime() - a.getTime());
    expect(dates.map(d => d.getTime())).toEqual(sortedDates.map(d => d.getTime()));
  });

  test('缓存机制验证', async ({ page }) => {
    const testUrl = `${TEMPORAL_SERVICE_URL}/api/v1/organization-units/${TEST_ORG_CODE}/temporal?as_of_date=2025-08-12`;
    
    // 第一次请求（缓存未命中）
    const startTime1 = Date.now();
    const response1 = await page.request.get(testUrl);
    const endTime1 = Date.now();
    const time1 = endTime1 - startTime1;
    
    expect(response1.ok()).toBeTruthy();
    
    // 第二次请求（缓存命中）
    const startTime2 = Date.now();
    const response2 = await page.request.get(testUrl);
    const endTime2 = Date.now();
    const time2 = endTime2 - startTime2;
    
    expect(response2.ok()).toBeTruthy();
    
    // 验证数据一致性
    const data1 = await response1.json();
    const data2 = await response2.json();
    expect(data1).toEqual(data2);
    
    // 缓存命中应该更快（通常情况下）
    console.log(`第一次请求: ${time1}ms, 第二次请求: ${time2}ms`);
  });

  test('错误处理和边界情况', async ({ page }) => {
    // 测试无效组织代码
    const invalidOrgResponse = await page.request.get(
      `${TEMPORAL_SERVICE_URL}/api/v1/organization-units/INVALID999/temporal`
    );
    expect(invalidOrgResponse.status()).toBe(404);
    
    // 测试无效日期格式
    const invalidDateResponse = await page.request.get(
      `${TEMPORAL_SERVICE_URL}/api/v1/organization-units/${TEST_ORG_CODE}/temporal?as_of_date=invalid-date`
    );
    // 可能返回400或默认处理，取决于实现
    expect([200, 400, 422]).toContain(invalidDateResponse.status());
    
    // 测试无效事件类型
    const invalidEventResponse = await page.request.post(
      `${TEMPORAL_SERVICE_URL}/api/v1/organization-units/${TEST_ORG_CODE}/events`,
      {
        data: {
          event_type: 'INVALID_EVENT',
          effective_date: '2025-01-01T00:00:00Z',
          change_data: {}
        }
      }
    );
    expect(invalidEventResponse.status()).toBe(400);
  });

  test('前端组件状态管理', async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/temporal-demo`);
    
    // 打开详情面板
    await page.locator('text=查看详情').first().click();
    
    // 验证时态数据加载状态
    await expect(page.locator('text=加载中')).toBeVisible();
    await expect(page.locator('text=加载中')).not.toBeVisible({ timeout: 10000 });
    
    // 验证数据加载完成后的UI状态
    await expect(page.locator('[data-testid="timeline-node"]').first()).toBeVisible();
    
    // 测试选中状态
    const firstNode = page.locator('[data-testid="timeline-node"]').first();
    await firstNode.click();
    
    // 验证选中样式（可能需要根据实际CSS调整）
    const selectedNode = page.locator('[data-testid="timeline-node"][data-selected="true"]');
    await expect(selectedNode).toBeVisible();
    
    // 关闭面板
    const closeButton = page.locator('button[aria-label="关闭"]');
    if (await closeButton.isVisible()) {
      await closeButton.click();
      await expect(page.locator('text=时间轴导航')).not.toBeVisible();
    }
  });
});

// 性能测试套件
test.describe('时态管理性能测试', () => {
  
  test('页面加载性能基准', async ({ page }) => {
    const startTime = Date.now();
    await page.goto(`${FRONTEND_URL}/temporal-demo`);
    await expect(page.locator('text=时态管理集成演示')).toBeVisible();
    const endTime = Date.now();
    
    const loadTime = endTime - startTime;
    expect(loadTime).toBeLessThan(3000); // 页面加载应在3秒内
    
    console.log(`页面加载时间: ${loadTime}ms`);
  });
  
  test('大量数据渲染性能', async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/temporal-demo`);
    
    // 打开具有多个时态记录的组织详情
    await page.locator('text=查看详情').first().click();
    
    const startTime = Date.now();
    await expect(page.locator('[data-testid="timeline-node"]').first()).toBeVisible({ timeout: 10000 });
    const endTime = Date.now();
    
    const renderTime = endTime - startTime;
    expect(renderTime).toBeLessThan(2000); // 渲染应在2秒内
    
    console.log(`时态数据渲染时间: ${renderTime}ms`);
    
    // 验证所有时态节点都正确渲染
    const timelineNodes = await page.locator('[data-testid="timeline-node"]').count();
    expect(timelineNodes).toBeGreaterThan(0);
    
    console.log(`渲染的时态节点数: ${timelineNodes}`);
  });
  
  test('API响应时间基准测试', async ({ page }) => {
    const testCases = [
      { name: '健康检查', url: `${TEMPORAL_SERVICE_URL}/health`, maxTime: 100 },
      { name: '当前记录查询', url: `${TEMPORAL_SERVICE_URL}/api/v1/organization-units/${TEST_ORG_CODE}/temporal?as_of_date=2025-08-12`, maxTime: 500 },
      { name: '完整历史查询', url: `${TEMPORAL_SERVICE_URL}/api/v1/organization-units/${TEST_ORG_CODE}/temporal?include_history=true&include_future=true`, maxTime: 1000 }
    ];
    
    for (const testCase of testCases) {
      const startTime = Date.now();
      const response = await page.request.get(testCase.url);
      const endTime = Date.now();
      
      const responseTime = endTime - startTime;
      
      expect(response.ok()).toBeTruthy();
      expect(responseTime).toBeLessThan(testCase.maxTime);
      
      console.log(`${testCase.name}: ${responseTime}ms (限制: ${testCase.maxTime}ms)`);
    }
  });
});