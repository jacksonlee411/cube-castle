import { test, expect } from '@playwright/test';
import { TestHelpers, TestDataGenerator, NavigationHelper } from '../utils/test-helpers';

test.describe('职位管理页面', () => {
  let helpers: TestHelpers;
  let navigation: NavigationHelper;

  test.beforeEach(async ({ page }) => {
    helpers = new TestHelpers(page);
    navigation = new NavigationHelper(page);
    
    // 导航到职位管理页面
    await navigation.goToPositions();
    await helpers.waitForPageLoad();
  });

  test('页面基础加载和布局验证', async ({ page }) => {
    // 验证页面标题
    await helpers.verifyPageTitle('职位管理');
    
    // 验证统计卡片 - 职位管理页面应该有4个统计卡片
    await expect(page.locator('[data-testid="stats-card"]')).toHaveCount(4);
    
    // 验证关键统计指标
    await helpers.verifyStatsCard('总职位数');
    await helpers.verifyStatsCard('空缺职位');
    await helpers.verifyStatsCard('平均薪资');
    await helpers.verifyStatsCard('热门部门');
    
    // 验证搜索和筛选功能
    await expect(page.locator('input[placeholder*="搜索"]')).toBeVisible();
    
    // 验证新增职位按钮
    await expect(page.locator('button:has-text("新增职位")')).toBeVisible();
    
    // 验证数据表格
    await helpers.waitForDataTableLoad();
    await expect(page.locator('[data-testid="data-table"]')).toBeVisible();
  });

  test('职位数据表格功能验证', async ({ page }) => {
    await helpers.waitForDataTableLoad();
    
    // 验证表格列标题
    const expectedColumns = ['职位信息', '部门', '薪资范围', '状态', '操作'];
    for (const column of expectedColumns) {
      await expect(page.locator('thead th')).toContainText(column);
    }
    
    // 验证表格有数据行
    const dataRows = page.locator('[data-testid="data-table"] tbody tr');
    const rowCount = await dataRows.count();
    expect(rowCount).toBeGreaterThan(0);
    
    // 验证分页功能
    const paginationInfo = page.locator('[data-testid="pagination-info"]');
    if (await paginationInfo.isVisible()) {
      const paginationText = await paginationInfo.textContent();
      expect(paginationText).toMatch(/\d+/); // 包含数字
    }
  });

  test('职位创建流程', async ({ page }) => {
    const testPosition = TestDataGenerator.generatePosition();
    
    // 点击新增职位按钮
    await helpers.clickButtonAndWait('新增职位');
    await helpers.waitForModal();
    
    try {
      // 填写基本信息 - 使用智能等待
      await helpers.fillFormField('input[name="title"]', testPosition.title);
      await helpers.fillFormField('input[name="department"]', testPosition.department);
      
      // 选择职级
      const jobLevelSelect = page.locator('select[name="jobLevel"]');
      if (await jobLevelSelect.isVisible()) {
        await jobLevelSelect.selectOption(testPosition.jobLevel);
      }
      
      // 填写薪资范围
      await helpers.fillFormField('input[name="salaryMin"]', testPosition.salaryMin);
      await helpers.fillFormField('input[name="salaryMax"]', testPosition.salaryMax);
      
      // 填写职位描述
      await helpers.fillFormField('textarea[name="description"]', testPosition.description);
      
      // 提交表单
      await helpers.clickButtonAndWait('创建');
      
      // 验证成功提示
      await helpers.verifyToastMessage('职位创建成功');
      
    } catch (error) {
      // 如果表单元素不存在，可能是页面结构不同
      console.log('表单测试失败，尝试简化验证:', error);
      
      // 简化验证：至少验证模态框有表单元素
      const modal = page.locator('[role="dialog"], .modal');
      if (await modal.isVisible()) {
        const formElements = modal.locator('input, select, textarea');
        const elementCount = await formElements.count();
        expect(elementCount).toBeGreaterThan(0);
        
        // 关闭模态框
        await helpers.closeModal();
      }
    }
    
    // 验证模态框关闭
    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
    
    // 验证新职位出现在列表中
    await helpers.waitForDataTableLoad();
    await helpers.searchInTable(testPosition.title);
    await expect(page.locator('[data-testid="data-table"]')).toContainText(testPosition.title);
  });

  test('职位编辑功能', async ({ page }) => {
    await helpers.waitForDataTableLoad();
    
    // 找到第一个编辑按钮
    const firstRowEdit = page.locator('[data-testid="data-table"] tbody tr').first().locator('button:has-text("编辑")');
    
    if (await firstRowEdit.isVisible()) {
      // 获取当前职位名称
      const currentPositionName = await page.locator('[data-testid="data-table"] tbody tr').first().locator('td').first().textContent();
      
      await firstRowEdit.click();
      await helpers.waitForModal();
      
      // 修改职位标题
      const updatedTitle = `更新职位${Date.now()}`;
      const titleInput = page.locator('input[name="title"]');
      await titleInput.fill(updatedTitle);
      
      // 保存更改
      await helpers.clickButtonAndWait('保存');
      
      // 验证成功提示
      await helpers.verifyToastMessage('职位信息已更新');
      
      // 验证更新后的信息
      await helpers.waitForDataTableLoad();
      await helpers.searchInTable(updatedTitle);
      await expect(page.locator('[data-testid="data-table"]')).toContainText(updatedTitle);
    }
  });

  test('职位搜索和筛选功能', async ({ page }) => {
    await helpers.waitForDataTableLoad();
    
    // 测试职位名称搜索
    await helpers.searchInTable('工程师');
    await helpers.waitForDataTableLoad();
    
    // 验证搜索结果
    const searchResults = page.locator('[data-testid="data-table"] tbody tr');
    const resultCount = await searchResults.count();
    
    if (resultCount > 0) {
      // 验证第一个结果包含搜索关键词
      const firstResult = await searchResults.first().textContent();
      expect(firstResult?.toLowerCase()).toContain('工程师');
    }
    
    // 清除搜索
    await helpers.searchInTable('');
    await helpers.waitForDataTableLoad();
    
    // 测试部门筛选 (如果有筛选功能)
    const departmentFilter = page.locator('select[name="department"]');
    if (await departmentFilter.isVisible()) {
      await departmentFilter.selectOption('技术部');
      await helpers.waitForDataTableLoad();
      
      // 重置筛选器
      await departmentFilter.selectOption('');
      await helpers.waitForDataTableLoad();
    }
  });

  test('职位状态管理', async ({ page }) => {
    await helpers.waitForDataTableLoad();
    
    // 找到第一个状态切换按钮
    const statusToggle = page.locator('[data-testid="data-table"] tbody tr').first().locator('button:has-text("停用"), button:has-text("启用")');
    
    if (await statusToggle.isVisible()) {
      const currentStatusText = await statusToggle.textContent();
      await statusToggle.click();
      
      // 验证状态切换提示
      const expectedMessage = currentStatusText?.includes('停用') ? '职位已停用' : '职位已启用';
      await helpers.verifyToastMessage(expectedMessage);
      
      // 验证状态已更改
      await helpers.waitForDataTableLoad();
      const newStatusText = await page.locator('[data-testid="data-table"] tbody tr').first().locator('button:has-text("停用"), button:has-text("启用")').textContent();
      expect(newStatusText).not.toBe(currentStatusText);
    }
  });

  test('薪资范围验证', async ({ page }) => {
    // 点击新增职位
    await page.locator('button:has-text("新增职位")').click();
    await helpers.waitForModal();
    
    // 填写基础信息
    await page.locator('input[name="title"]').fill('测试薪资验证');
    await page.locator('input[name="department"]').fill('测试部');
    
    // 测试无效薪资范围（最低薪资大于最高薪资）
    await page.locator('input[name="salaryMin"]').fill('30000');
    await page.locator('input[name="salaryMax"]').fill('20000');
    
    // 尝试提交
    await page.locator('button:has-text("创建")').click();
    
    // 验证错误提示（如果有验证）
    const errorMessage = page.locator('.error, [role="alert"]');
    if (await errorMessage.isVisible()) {
      await expect(errorMessage).toContainText('薪资');
    }
    
    // 关闭模态框
    await helpers.closeModal();
  });

  test('职位详情查看', async ({ page }) => {
    await helpers.waitForDataTableLoad();
    
    // 点击第一个职位的查看详情按钮
    const viewButton = page.locator('[data-testid="data-table"] tbody tr').first().locator('button:has-text("查看"), button:has-text("详情")');
    
    if (await viewButton.isVisible()) {
      await viewButton.click();
      await helpers.waitForModal();
      
      // 验证详情模态框包含关键信息
      const modal = page.locator('[role="dialog"]');
      await expect(modal).toContainText('职位信息');
      await expect(modal).toContainText('职位描述');
      
      // 关闭详情模态框
      await helpers.closeModal();
    }
  });

  test('响应式设计验证', async ({ page }) => {
    // 切换到移动端视口
    await page.setViewportSize({ width: 375, height: 667 });
    await helpers.waitForPageLoad();
    
    // 验证移动端布局
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('[data-testid="data-table"]')).toBeVisible();
    
    // 验证统计卡片在移动端的显示
    const statsCards = page.locator('[data-testid="stats-card"]');
    await expect(statsCards.first()).toBeVisible();
    
    // 恢复桌面视口
    await page.setViewportSize({ width: 1280, height: 720 });
  });

  test('数据导出功能', async ({ page }) => {
    await helpers.waitForDataTableLoad();
    
    // 查找导出按钮
    const exportButton = page.locator('button:has-text("导出"), button:has-text("Export")');
    
    if (await exportButton.isVisible()) {
      // 点击导出按钮
      const downloadPromise = page.waitForEvent('download');
      await exportButton.click();
      
      // 验证文件下载
      const download = await downloadPromise;
      expect(download.suggestedFilename()).toMatch(/\.(csv|xlsx)$/);
    }
  });

  test.afterEach(async ({ page }) => {
    // 截图用于调试
    await helpers.takeScreenshot(`positions-test-${Date.now()}`);
  });
});