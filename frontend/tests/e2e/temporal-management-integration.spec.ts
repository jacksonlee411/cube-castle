/**
 * 时态管理组件集成测试
 * 测试时态管理主从视图组件的完整功能
 * 🎯 使用动态环境配置替代硬编码端口
 */

import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';
import { E2E_CONFIG, validateTestEnvironment } from './config/test-environment';

let FRONTEND_URL: string;
const COMMAND_API_URL = E2E_CONFIG.COMMAND_API_URL; // 命令服务（REST）
const GRAPHQL_API_URL = E2E_CONFIG.GRAPHQL_API_URL; // 查询服务（GraphQL）
const GRAPHQL_HEADERS = { 'Content-Type': 'application/json' } as const;
const TEST_ORG_CODE = '1000056';

const ORGANIZATION_VERSIONS_QUERY = `
  query OrganizationVersions($code: String!) {
    organizationVersions(code: $code) {
      code
      name
      unitType
      status
      effectiveDate
      endDate
      recordId
    }
  }
`;

const ORGANIZATION_AS_OF_QUERY = `
  query OrganizationAsOf($code: String!, $asOfDate: String!) {
    organization(code: $code, asOfDate: $asOfDate) {
      code
      name
      unitType
      status
      effectiveDate
      endDate
      recordId
    }
  }
`;

const GRAPHQL_HEALTH_QUERY = 'query GraphQLHealth { __typename }';

async function graphQLRequest(page: Page, query: string, variables: Record<string, unknown> = {}) {
  const response = await page.request.post(GRAPHQL_API_URL, {
    data: { query, variables },
    headers: GRAPHQL_HEADERS,
  });
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  if (body.errors) {
    throw new Error(`GraphQL errors: ${JSON.stringify(body.errors)}`);
  }
  return body.data;
}

test.describe('时态管理系统集成测试', () => {
  
  test.beforeAll(async () => {
    const envValidation = await validateTestEnvironment();
    if (!envValidation.isValid) {
      console.error('🚨 测试环境验证失败:', envValidation.errors);
      throw new Error('测试环境不可用');
    }
    FRONTEND_URL = envValidation.frontendUrl;
    console.log(`✅ 使用前端基址: ${FRONTEND_URL}`);
  });
  
  test.beforeEach(async ({ page }) => {
    // 确保命令服务健康
    const restHealthResponse = await page.request.get(`${COMMAND_API_URL}/health`);
    expect(restHealthResponse.ok()).toBeTruthy();

    // 确认 GraphQL 查询服务可用
    const data = await graphQLRequest(page, GRAPHQL_HEALTH_QUERY);
    expect(data.__typename).toBe('Query');
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
    const scenarios: Array<{ name: string; maxTime: number; exec: () => Promise<import('@playwright/test').APIResponse>; }> = [
      {
        name: '健康检查',
        maxTime: 150,
        exec: () => page.request.get(`${COMMAND_API_URL}/health`),
      },
      {
        name: 'GraphQL 组织版本查询',
        maxTime: 800,
        exec: () => page.request.post(GRAPHQL_API_URL, {
          data: {
            query: ORGANIZATION_VERSIONS_QUERY,
            variables: { code: TEST_ORG_CODE },
          },
          headers: GRAPHQL_HEADERS,
        }),
      },
    ];

    for (const scenario of scenarios) {
      const started = Date.now();
      const response = await scenario.exec();
      const finished = Date.now();
      const responseTime = finished - started;

      expect(response.ok()).toBeTruthy();
      expect(responseTime).toBeLessThan(scenario.maxTime);

      console.log(`${scenario.name} 响应时间: ${responseTime}ms (限制: ${scenario.maxTime}ms)`);
    }
  });

  test('时态数据一致性验证', async ({ page }) => {
    const data = await graphQLRequest(page, ORGANIZATION_VERSIONS_QUERY, { code: TEST_ORG_CODE });
    const versions = data.organizationVersions as Array<Record<string, string | null>>;

    expect(Array.isArray(versions)).toBeTruthy();
    expect(versions.length).toBeGreaterThan(0);

    for (const version of versions) {
      expect(version.code).toBe(TEST_ORG_CODE);
      expect(version.name).toBeTruthy();
      expect(version.unitType).toBeTruthy();
      expect(version.status).toBeTruthy();
      expect(version.effectiveDate).toBeTruthy();
      expect(version.recordId).toBeTruthy();
    }

    const currentRecords = versions.filter(version => version.endDate === null);
    expect(currentRecords.length).toBeLessThanOrEqual(1);

    const effectiveTimestamps = versions.map(version => new Date(version.effectiveDate as string).getTime());
    const sorted = [...effectiveTimestamps].sort((a, b) => b - a);
    expect(effectiveTimestamps).toEqual(sorted);
  });

  test('缓存机制验证', async ({ page }) => {
    const variables = { code: TEST_ORG_CODE, asOfDate: '2025-08-12' };

    const startTime1 = Date.now();
    const response1 = await page.request.post(GRAPHQL_API_URL, {
      data: { query: ORGANIZATION_AS_OF_QUERY, variables },
      headers: GRAPHQL_HEADERS,
    });
    const time1 = Date.now() - startTime1;
    expect(response1.ok()).toBeTruthy();
    const body1 = await response1.json();
    expect(body1.errors).toBeUndefined();

    const startTime2 = Date.now();
    const response2 = await page.request.post(GRAPHQL_API_URL, {
      data: { query: ORGANIZATION_AS_OF_QUERY, variables },
      headers: GRAPHQL_HEADERS,
    });
    const time2 = Date.now() - startTime2;
    expect(response2.ok()).toBeTruthy();
    const body2 = await response2.json();
    expect(body2.errors).toBeUndefined();

    expect(body1.data).toEqual(body2.data);
    console.log(`GraphQL 缓存验证 - 首次: ${time1}ms, 二次: ${time2}ms`);
  });

  test('错误处理和边界情况', async ({ page }) => {
    const invalidOrgResponse = await page.request.post(GRAPHQL_API_URL, {
      data: {
        query: ORGANIZATION_AS_OF_QUERY,
        variables: { code: 'INVALID999', asOfDate: '2025-08-12' },
      },
      headers: GRAPHQL_HEADERS,
    });
    expect(invalidOrgResponse.ok()).toBeTruthy();
    const invalidOrgBody = await invalidOrgResponse.json();
    expect(invalidOrgBody.errors ?? null).toBeNull();
    expect(invalidOrgBody.data.organization).toBeNull();

    const invalidDateResponse = await page.request.post(GRAPHQL_API_URL, {
      data: {
        query: ORGANIZATION_AS_OF_QUERY,
        variables: { code: TEST_ORG_CODE, asOfDate: 'invalid-date' },
      },
      headers: GRAPHQL_HEADERS,
    });
    const invalidDateBody = await invalidDateResponse.json();
    const hasErrors = Array.isArray(invalidDateBody.errors) && invalidDateBody.errors.length > 0;
    const hasNullData = invalidDateBody?.data?.organization === null;
    expect(hasErrors || hasNullData).toBeTruthy();

    const invalidEventResponse = await page.request.post(
      `${COMMAND_API_URL}/organization-units/${TEST_ORG_CODE}/events`,
      {
        data: {
          eventType: 'INVALID_EVENT',
          recordId: '00000000-0000-0000-0000-000000000000',
          changeReason: 'Playwright invalid event',
          effectiveDate: '2025-01-01',
        },
        headers: GRAPHQL_HEADERS,
      }
    );
    expect([400, 422]).toContain(invalidEventResponse.status());
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
    const scenarios: Array<{ name: string; maxTime: number; exec: () => Promise<import('@playwright/test').APIResponse>; }> = [
      {
        name: '健康检查',
        maxTime: 150,
        exec: () => page.request.get(`${COMMAND_API_URL}/health`),
      },
      {
        name: 'GraphQL 版本列表',
        maxTime: 800,
        exec: () => page.request.post(GRAPHQL_API_URL, {
          data: {
            query: ORGANIZATION_VERSIONS_QUERY,
            variables: { code: TEST_ORG_CODE },
          },
          headers: GRAPHQL_HEADERS,
        }),
      },
      {
        name: 'GraphQL asOf 查询',
        maxTime: 800,
        exec: () => page.request.post(GRAPHQL_API_URL, {
          data: {
            query: ORGANIZATION_AS_OF_QUERY,
            variables: { code: TEST_ORG_CODE, asOfDate: '2025-08-12' },
          },
          headers: GRAPHQL_HEADERS,
        }),
      },
    ];

    for (const scenario of scenarios) {
      const started = Date.now();
      const response = await scenario.exec();
      const finished = Date.now();
      const responseTime = finished - started;

      expect(response.ok()).toBeTruthy();
      expect(responseTime).toBeLessThan(scenario.maxTime);

      console.log(`${scenario.name}: ${responseTime}ms (限制: ${scenario.maxTime}ms)`);
    }
  });
});
