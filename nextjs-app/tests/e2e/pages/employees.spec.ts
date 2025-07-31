import { test, expect } from '@playwright/test';
import { TestHelpers, TestDataGenerator, NavigationHelper } from '../utils/test-helpers';

test.describe('员工管理页面', () => {
  let helpers: TestHelpers;
  let navigation: NavigationHelper;

  test.beforeEach(async ({ page }) => {
    helpers = new TestHelpers(page);
    navigation = new NavigationHelper(page);
    
    // 导航到员工管理页面
    await navigation.goToEmployees();
    await helpers.waitForPageLoad();
  });

  test('页面基础加载和布局验证', async ({ page }) => {
    // 验证页面标题
    await helpers.verifyPageTitle('员工管理');
    
    // 验证统计卡片
    await expect(page.locator('[data-testid="stats-card"]')).toHaveCount(4);
    
    // 验证搜索框存在
    await expect(page.locator('input[placeholder*="搜索"]')).toBeVisible();
    
    // 验证新增员工按钮
    await expect(page.locator('button:has-text("新增员工")')).toBeVisible();
    
    // 验证数据表格加载
    await helpers.waitForDataTableLoad();
    await expect(page.locator('[data-testid="data-table"]')).toBeVisible();
  });

  test('数据表格功能验证', async ({ page }) => {
    await helpers.waitForDataTableLoad();
    
    // 验证表格列标题
    const expectedColumns = ['员工信息', '联系方式', '职位信息', '入职时间', '操作'];
    for (const column of expectedColumns) {
      await expect(page.locator('th')).toContainText(column);
    }
    
    // 验证分页功能
    const paginationNext = page.locator('button[aria-label="Go to next page"]');
    if (await paginationNext.isVisible() && await paginationNext.isEnabled()) {
      await paginationNext.click();
      await helpers.waitForDataTableLoad();
    }
    
    // 验证搜索功能
    await helpers.searchInTable('张');
    await helpers.waitForDataTableLoad();
    
    // 清除搜索
    await helpers.searchInTable('');
    await helpers.waitForDataTableLoad();
  });

  test('员工创建流程', async ({ page }) => {
    const testEmployee = TestDataGenerator.generateEmployee();
    
    // 点击新增员工按钮
    await page.locator('button:has-text("新增员工")').click();
    await helpers.waitForModal();
    
    // 填写员工信息
    await page.locator('input[name="legalName"]').fill(testEmployee.legalName);
    await page.locator('input[name="email"]').fill(testEmployee.email);
    await page.locator('input[name="hireDate"]').fill(testEmployee.hireDate);
    
    // 选择职位和部门 (如果有下拉选择)
    const positionSelect = page.locator('select[name="position"]');
    if (await positionSelect.isVisible()) {
      await positionSelect.selectOption(testEmployee.position);
    }
    
    // 提交表单
    await helpers.clickButtonAndWait('创建');
    
    // 验证成功提示
    await helpers.verifyToastMessage('员工创建成功');
    
    // 验证模态框关闭
    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
    
    // 验证新员工出现在列表中
    await helpers.waitForDataTableLoad();
    await helpers.searchInTable(testEmployee.legalName);
    await expect(page.locator('[data-testid="data-table"]')).toContainText(testEmployee.legalName);
  });

  test('员工信息编辑', async ({ page }) => {
    await helpers.waitForDataTableLoad();
    
    // 点击第一行的编辑按钮
    const firstRowEdit = page.locator('[data-testid="data-table"] tbody tr').first().locator('button:has-text("编辑")');
    if (await firstRowEdit.isVisible()) {
      await firstRowEdit.click();
      await helpers.waitForModal();
      
      // 修改员工信息
      const updatedName = `更新员工${Date.now()}`;
      await page.locator('input[name="legalName"]').fill(updatedName);
      
      // 保存更改
      await helpers.clickButtonAndWait('保存');
      
      // 验证成功提示
      await helpers.verifyToastMessage('员工信息已更新');
      
      // 验证更新后的信息
      await helpers.waitForDataTableLoad();
      await helpers.searchInTable(updatedName);
      await expect(page.locator('[data-testid="data-table"]')).toContainText(updatedName);
    }
  });

  test('员工搜索和筛选', async ({ page }) => {
    await helpers.waitForDataTableLoad();
    
    // 测试搜索功能
    await helpers.searchInTable('张');
    await helpers.waitForDataTableLoad();
    
    // 验证搜索结果
    const tableRows = page.locator('[data-testid="data-table"] tbody tr');
    const rowCount = await tableRows.count();
    
    if (rowCount > 0) {
      // 验证搜索结果包含关键词
      const firstRowText = await tableRows.first().textContent();
      expect(firstRowText).toContain('张');
    }
    
    // 清除搜索
    await helpers.searchInTable('');
    await helpers.waitForDataTableLoad();
  });

  test('表格排序功能', async ({ page }) => {
    await helpers.waitForDataTableLoad();
    
    // 点击员工姓名列进行排序
    const nameColumnHeader = page.locator('th:has-text("员工信息")');
    if (await nameColumnHeader.isVisible()) {
      await nameColumnHeader.click();
      await helpers.waitForDataTableLoad();
      
      // 再次点击切换排序顺序
      await nameColumnHeader.click();
      await helpers.waitForDataTableLoad();
    }
  });

  test('响应式设计验证', async ({ page }) => {
    // 测试移动端视口
    await page.setViewportSize({ width: 375, height: 667 });
    await helpers.waitForPageLoad();
    
    // 验证页面在小屏幕上正常显示
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('[data-testid="data-table"]')).toBeVisible();
    
    // 恢复桌面视口
    await page.setViewportSize({ width: 1280, height: 720 });
  });

  test('页面性能验证', async ({ page }) => {
    // 测量页面加载性能
    const startTime = Date.now();
    await navigation.goToEmployees();
    await helpers.waitForDataTableLoad();
    const loadTime = Date.now() - startTime;
    
    // 验证页面在3秒内加载完成
    expect(loadTime).toBeLessThan(3000);
    
    // 验证关键元素可见
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('[data-testid="data-table"]')).toBeVisible();
  });

  test.afterEach(async ({ page }) => {
    // 清理测试数据或重置状态
    await helpers.takeScreenshot(`employees-test-${Date.now()}`);
  });
});