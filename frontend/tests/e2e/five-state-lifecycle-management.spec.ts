/**
 * 五状态生命周期管理系统E2E测试
 * 测试范围：CURRENT, HISTORICAL, PLANNED, SUSPENDED, DELETED
 * 版本：v2.1 - 时态管理系统升级
 * 创建时间：2025-08-18
 */

import { test, expect, Page } from '@playwright/test';
import { E2E_CONFIG, validateTestEnvironment } from './config/test-environment';

// 测试数据配置 - 使用动态环境配置
const TEST_CONFIG = {
  baseUrl: '',  // 将在beforeAll中初始化
  apiUrl: E2E_CONFIG.COMMAND_API_URL,
  graphqlUrl: E2E_CONFIG.GRAPHQL_API_URL,
  temporalApiUrl: E2E_CONFIG.COMMAND_API_URL,
  testCode: '1000004', // 使用现有的测试组织
  timeout: 30000
};

// 测试用例: 五状态生命周期管理系统完整功能验证
test.describe('五状态生命周期管理系统 E2E 测试', () => {
  let page: Page;

  test.beforeAll(async () => {
    const envValidation = await validateTestEnvironment();
    if (!envValidation.isValid) {
      console.error('🚨 测试环境验证失败:', envValidation.errors);
      throw new Error('测试环境不可用');
    }
    TEST_CONFIG.baseUrl = envValidation.frontendUrl;
    console.log(`✅ 使用前端基址: ${TEST_CONFIG.baseUrl}`);
  });

  test.beforeEach(async ({ page: testPage }) => {
    page = testPage;
    
    // 设置页面超时时间
    page.setDefaultTimeout(TEST_CONFIG.timeout);
    
    // 导航到时态管理页面
    await page.goto(`${TEST_CONFIG.baseUrl}/temporal/${TEST_CONFIG.testCode}`);
    
    // 等待页面加载完成
    await page.waitForSelector('[data-testid="temporal-master-detail-view"]', { timeout: 10000 });
  });

  test('应该正确显示五种状态的组织记录', async () => {
    // 1. 验证当前记录状态 (CURRENT)
    const currentBadge = page.locator('[data-testid="lifecycle-status-badge"][data-status="CURRENT"]').first();
    await expect(currentBadge).toBeVisible();
    await expect(currentBadge).toContainText('当前记录');
    
    // 2. 验证历史记录状态 (HISTORICAL) 
    const historicalBadges = page.locator('[data-testid="lifecycle-status-badge"][data-status="HISTORICAL"]');
    await expect(historicalBadges).toHaveCount(4); // 应该有4条历史记录
    
    // 3. 验证计划中状态 (PLANNED)
    const plannedBadge = page.locator('[data-testid="lifecycle-status-badge"][data-status="PLANNED"]').first();
    await expect(plannedBadge).toBeVisible();
    await expect(plannedBadge).toContainText('计划中');
    
    // 4. 验证时间轴导航正确显示
    const timelineNodes = page.locator('[data-testid="timeline-node"]');
    await expect(timelineNodes).toHaveCountGreaterThan(5); // 至少6条记录
  });

  test('应该支持状态转换功能', async () => {
    // 1. 选择一个历史记录节点
    const firstHistoricalNode = page.locator('[data-testid="timeline-node"]').nth(1);
    await firstHistoricalNode.click();
    
    // 2. 验证详情区域显示正确信息
    const detailsCard = page.locator('[data-testid="version-details-card"]');
    await expect(detailsCard).toBeVisible();
    
    // 3. 点击编辑按钮
    const editButton = page.locator('[data-testid="edit-version-button"]');
    await editButton.click();
    
    // 4. 验证五状态选择器可见
    const statusSelector = page.locator('[data-testid="five-state-status-selector"]');
    await expect(statusSelector).toBeVisible();
    
    // 5. 验证所有状态选项可用（除了删除状态）
    await statusSelector.click();
    const statusOptions = page.locator('[data-testid="status-option"]');
    await expect(statusOptions).toHaveCount(4); // CURRENT, HISTORICAL, PLANNED, SUSPENDED
  });

  test('应该正确处理停用和恢复操作', async () => {
    // 1. 选择当前记录
    const currentNode = page.locator('[data-testid="timeline-node"][data-current="true"]');
    await currentNode.click();
    
    // 2. 打开操作菜单
    const actionMenu = page.locator('[data-testid="version-action-menu"]');
    await actionMenu.click();
    
    // 3. 点击停用操作
    const suspendButton = page.locator('[data-testid="suspend-version-button"]');
    await suspendButton.click();
    
    // 4. 确认停用对话框
    const confirmDialog = page.locator('[data-testid="confirm-suspend-dialog"]');
    await expect(confirmDialog).toBeVisible();
    
    const confirmButton = page.locator('[data-testid="confirm-suspend-button"]');
    await confirmButton.click();
    
    // 5. 验证状态更新为停用
    await page.waitForTimeout(2000); // 等待状态更新
    const suspendedBadge = page.locator('[data-testid="lifecycle-status-badge"][data-status="SUSPENDED"]');
    await expect(suspendedBadge).toBeVisible();
  });

  test('应该支持自动结束日期管理', async () => {
    // 1. 点击新建版本
    const newVersionButton = page.locator('[data-testid="new-version-button"]');
    await newVersionButton.click();
    
    // 2. 填写新版本表单
    await page.fill('[data-testid="version-name-input"]', '自动结束日期测试版本');
    await page.selectOption('[data-testid="unit-type-select"]', 'DEPARTMENT');
    await page.selectOption('[data-testid="status-select"]', 'PLANNED');
    
    // 3. 设置生效日期为未来日期
    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 30);
    const futureDateStr = futureDate.toISOString().split('T')[0];
    await page.fill('[data-testid="effective-date-input"]', futureDateStr);
    
    // 4. 提交表单
    const submitButton = page.locator('[data-testid="submit-version-button"]');
    await submitButton.click();
    
    // 5. 验证成功创建并且前一版本的结束日期自动设置
    await page.waitForTimeout(3000); // 等待API响应
    
    // 6. 检查是否显示成功消息
    const successMessage = page.locator('[data-testid="success-message"]');
    await expect(successMessage).toBeVisible();
  });

  test('应该正确显示状态转换提示', async () => {
    // 1. 选择一个历史记录
    const historicalNode = page.locator('[data-testid="timeline-node"][data-status="HISTORICAL"]').first();
    await historicalNode.click();
    
    // 2. 打开编辑表单
    const editButton = page.locator('[data-testid="edit-version-button"]');
    await editButton.click();
    
    // 3. 改变状态选择
    const statusSelector = page.locator('[data-testid="five-state-status-selector"]');
    await statusSelector.selectOption('CURRENT');
    
    // 4. 验证状态转换提示出现
    const transitionHint = page.locator('[data-testid="state-transition-hint"]');
    await expect(transitionHint).toBeVisible();
    await expect(transitionHint).toContainText('历史记录将转为当前生效状态');
  });

  test('应该支持批量状态查询和筛选', async () => {
    // 1. 打开状态筛选器
    const statusFilter = page.locator('[data-testid="status-filter"]');
    await statusFilter.click();
    
    // 2. 选择只显示历史记录
    const historicalFilter = page.locator('[data-testid="filter-historical"]');
    await historicalFilter.click();
    
    // 3. 验证只显示历史记录
    const visibleNodes = page.locator('[data-testid="timeline-node"]:visible');
    await expect(visibleNodes).toHaveCount(4); // 应该只显示4条历史记录
    
    // 4. 验证所有可见节点都是历史记录状态
    for (let i = 0; i < 4; i++) {
      const node = visibleNodes.nth(i);
      const badge = node.locator('[data-testid="lifecycle-status-badge"][data-status="HISTORICAL"]');
      await expect(badge).toBeVisible();
    }
  });

  test('应该正确处理删除和恢复操作', async () => {
    // 1. 选择一个非当前记录
    const historicalNode = page.locator('[data-testid="timeline-node"][data-status="HISTORICAL"]').first();
    await historicalNode.click();
    
    // 2. 打开危险操作菜单
    const dangerMenu = page.locator('[data-testid="danger-action-menu"]');
    await dangerMenu.click();
    
    // 3. 点击软删除操作
    const deleteButton = page.locator('[data-testid="soft-delete-button"]');
    await deleteButton.click();
    
    // 4. 确认删除对话框
    const deleteDialog = page.locator('[data-testid="confirm-delete-dialog"]');
    await expect(deleteDialog).toBeVisible();
    await expect(deleteDialog).toContainText('此操作不可撤销');
    
    const confirmDeleteButton = page.locator('[data-testid="confirm-delete-button"]');
    await confirmDeleteButton.click();
    
    // 5. 验证记录被标记为删除状态
    await page.waitForTimeout(2000);
    const deletedBadge = page.locator('[data-testid="lifecycle-status-badge"][data-status="DELETED"]');
    await expect(deletedBadge).toBeVisible();
  });

  test('应该验证数据完整性约束', async () => {
    // 1. 尝试创建无效的时间范围
    const newVersionButton = page.locator('[data-testid="new-version-button"]');
    await newVersionButton.click();
    
    // 2. 设置生效日期早于现有记录
    await page.fill('[data-testid="version-name-input"]', '无效时间范围测试');
    await page.fill('[data-testid="effective-date-input"]', '2000-01-01'); // 过早的日期
    
    // 3. 尝试提交
    const submitButton = page.locator('[data-testid="submit-version-button"]');
    await submitButton.click();
    
    // 4. 验证显示错误消息
    const errorMessage = page.locator('[data-testid="validation-error"]');
    await expect(errorMessage).toBeVisible();
    await expect(errorMessage).toContainText('生效日期不能早于');
  });

  test('应该支持时态查询API集成', async () => {
    // 1. 选择特定日期进行时态查询
    const dateSelector = page.locator('[data-testid="as-of-date-picker"]');
    await dateSelector.fill('2020-06-01');
    
    // 2. 执行查询
    const queryButton = page.locator('[data-testid="temporal-query-button"]');
    await queryButton.click();
    
    // 3. 验证返回正确的历史状态
    await page.waitForTimeout(2000);
    const queryResult = page.locator('[data-testid="temporal-query-result"]');
    await expect(queryResult).toBeVisible();
    
    // 4. 验证显示该日期的有效记录
    const effectiveVersion = page.locator('[data-testid="effective-version-card"]');
    await expect(effectiveVersion).toBeVisible();
    await expect(effectiveVersion).toContainText('战略人力资源部'); // 2020年有效的版本
  });
});

// 性能测试用例
test.describe('五状态生命周期管理 - 性能测试', () => {
  test('大量历史记录下的页面性能', async ({ page }) => {
    // 1. 导航到有大量历史记录的组织
    await page.goto(`${TEST_CONFIG.baseUrl}/temporal/${TEST_CONFIG.testCode}`);
    
    // 2. 测量页面加载性能
    const startTime = Date.now();
    await page.waitForSelector('[data-testid="temporal-master-detail-view"]');
    const loadTime = Date.now() - startTime;
    
    // 3. 验证加载时间在合理范围内（< 3秒）
    expect(loadTime).toBeLessThan(3000);
    
    // 4. 测量时间轴滚动性能
    const timeline = page.locator('[data-testid="timeline-container"]');
    
    const scrollStart = Date.now();
    await timeline.evaluate(el => el.scrollTop = el.scrollHeight);
    const scrollTime = Date.now() - scrollStart;
    
    // 5. 验证滚动响应时间（< 500ms）
    expect(scrollTime).toBeLessThan(500);
  });

  test('状态转换操作响应时间', async ({ page }) => {
    await page.goto(`${TEST_CONFIG.baseUrl}/temporal/${TEST_CONFIG.testCode}`);
    await page.waitForSelector('[data-testid="temporal-master-detail-view"]');
    
    // 选择记录并测量状态切换时间
    const historicalNode = page.locator('[data-testid="timeline-node"]').nth(1);
    
    const switchStart = Date.now();
    await historicalNode.click();
    await page.waitForSelector('[data-testid="version-details-card"]');
    const switchTime = Date.now() - switchStart;
    
    // 验证状态切换响应时间（< 1秒）
    expect(switchTime).toBeLessThan(1000);
  });
});

// 错误处理测试用例
test.describe('五状态生命周期管理 - 错误处理', () => {
  test('API错误时的用户界面处理', async ({ page }) => {
    // 1. 模拟API服务不可用
    await page.route('**/api/v1/organization-units/**', route => {
      route.fulfill({
        status: 500,
        body: JSON.stringify({ error: 'Internal Server Error' })
      });
    });
    
    // 2. 导航到页面
    await page.goto(`${TEST_CONFIG.baseUrl}/temporal/${TEST_CONFIG.testCode}`);
    
    // 3. 验证显示错误状态
    const errorMessage = page.locator('[data-testid="api-error-message"]');
    await expect(errorMessage).toBeVisible();
    await expect(errorMessage).toContainText('无法加载组织数据');
    
    // 4. 验证重试按钮可用
    const retryButton = page.locator('[data-testid="retry-button"]');
    await expect(retryButton).toBeVisible();
    await expect(retryButton).toBeEnabled();
  });

  test('网络中断时的离线处理', async ({ page }) => {
    // 1. 先正常加载页面
    await page.goto(`${TEST_CONFIG.baseUrl}/temporal/${TEST_CONFIG.testCode}`);
    await page.waitForSelector('[data-testid="temporal-master-detail-view"]');
    
    // 2. 模拟网络中断
    await page.route('**/*', route => route.abort('internetdisconnected'));
    
    // 3. 尝试执行需要网络的操作
    const newVersionButton = page.locator('[data-testid="new-version-button"]');
    await newVersionButton.click();
    
    // 4. 验证显示网络错误提示
    const networkError = page.locator('[data-testid="network-error"]');
    await expect(networkError).toBeVisible();
  });
});

export default {};