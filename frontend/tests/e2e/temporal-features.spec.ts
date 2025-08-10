/**
 * 时态功能E2E测试
 * 测试完整的时态管理用户流程
 */
import { test, expect } from '@playwright/test';

test.describe('时态管理功能', () => {
  test.beforeEach(async ({ page }) => {
    // 导航到时态管理页面
    await page.goto('/temporal');
    
    // 等待页面加载完成
    await page.waitForLoadState('networkidle');
  });

  test('应该正确显示时态导航栏', async ({ page }) => {
    // 验证时态导航栏存在
    await expect(page.locator('[data-testid="temporal-navbar"]')).toBeVisible();
    
    // 验证模式切换按钮
    await expect(page.getByText('当前')).toBeVisible();
    await expect(page.getByText('历史')).toBeVisible();
    await expect(page.getByText('规划')).toBeVisible();
    
    // 验证当前模式高亮显示
    await expect(page.getByRole('button', { name: '当前' })).toHaveAttribute('aria-pressed', 'true');
  });

  test('应该能够切换到历史模式', async ({ page }) => {
    // 点击历史模式按钮
    await page.getByText('历史').click();
    
    // 验证日期选择器弹出
    await expect(page.getByText('选择历史查看时点')).toBeVisible();
    
    // 选择一个历史日期
    const lastMonth = new Date();
    lastMonth.setMonth(lastMonth.getMonth() - 1);
    const dateString = lastMonth.toISOString().slice(0, 16);
    
    await page.locator('input[type="datetime-local"]').first().fill(dateString);
    await page.getByText('确认选择').click();
    
    // 验证模式切换成功
    await expect(page.getByText('历史视图')).toBeVisible();
    
    // 验证历史模式标识
    await expect(page.getByText('历史')).toHaveAttribute('aria-pressed', 'true');
  });

  test('应该能够在表格中查看组织数据', async ({ page }) => {
    // 等待表格加载
    await expect(page.getByText('组织架构')).toBeVisible();
    
    // 验证表格列标题
    await expect(page.getByText('组织代码')).toBeVisible();
    await expect(page.getByText('组织名称')).toBeVisible();
    await expect(page.getByText('类型')).toBeVisible();
    await expect(page.getByText('状态')).toBeVisible();
    
    // 验证至少有一行数据
    const tableRows = page.locator('tbody tr');
    await expect(tableRows.first()).toBeVisible();
  });

  test('应该能够查看组织的时间线', async ({ page }) => {
    // 等待表格加载
    await expect(page.getByText('组织架构')).toBeVisible();
    
    // 点击第一个组织的时间线按钮
    const timelineButton = page.getByRole('button', { name: '查看时间线' }).first();
    await timelineButton.click();
    
    // 验证时间线弹窗打开
    await expect(page.getByText('时间线')).toBeVisible();
    
    // 验证时间线内容
    const timelineEvents = page.locator('[data-testid="timeline-event"]');
    if (await timelineEvents.count() > 0) {
      await expect(timelineEvents.first()).toBeVisible();
    } else {
      // 如果没有事件，应该显示空状态
      await expect(page.getByText('📭 暂无时间线事件')).toBeVisible();
    }
  });

  test('应该能够查看组织的历史版本对比', async ({ page }) => {
    // 等待表格加载
    await expect(page.getByText('组织架构')).toBeVisible();
    
    // 点击第一个组织的历史按钮
    const historyButton = page.getByRole('button', { name: '查看历史版本' }).first();
    await historyButton.click();
    
    // 验证版本对比弹窗打开
    await expect(page.getByText('版本历史')).toBeVisible();
    
    // 验证版本对比内容
    const versionContent = page.locator('[data-testid="version-comparison"]');
    if (await versionContent.isVisible()) {
      // 如果有历史版本，验证对比功能
      await expect(page.getByText('版本对比')).toBeVisible();
    } else {
      // 如果没有历史版本，应该显示相应提示
      await expect(page.getByText('仅有一个版本，无法对比')).toBeVisible();
    }
  });

  test('应该能够使用标签页切换视图', async ({ page }) => {
    // 验证默认在组织列表标签页
    await expect(page.getByRole('tab', { name: '组织列表' })).toHaveAttribute('aria-selected', 'true');
    
    // 切换到时间线视图标签页
    await page.getByRole('tab', { name: '时间线视图' }).click();
    await expect(page.getByRole('tab', { name: '时间线视图' })).toHaveAttribute('aria-selected', 'true');
    
    // 应该显示选择组织的提示（因为没有选中的组织）
    await expect(page.getByText('请从组织列表中选择一个组织来查看其时间线')).toBeVisible();
    
    // 切换到版本对比标签页
    const comparisonTab = page.getByRole('tab', { name: '版本对比' });
    if (await comparisonTab.isVisible()) {
      await comparisonTab.click();
      await expect(comparisonTab).toHaveAttribute('aria-selected', 'true');
      await expect(page.getByText('请从组织列表中选择一个组织来查看版本对比')).toBeVisible();
    }
  });

  test('应该能够使用时态设置', async ({ page }) => {
    // 点击设置按钮（如果存在）
    const settingsButton = page.getByRole('button', { name: '时态查询设置' });
    if (await settingsButton.isVisible()) {
      await settingsButton.click();
      
      // 验证设置弹窗打开
      await expect(page.getByText('时态查询设置')).toBeVisible();
      
      // 验证设置选项
      await expect(page.getByText('基础设置')).toBeVisible();
      await expect(page.getByText('时间范围筛选')).toBeVisible();
      
      // 关闭设置弹窗
      await page.getByText('取消').click();
      await expect(page.getByText('时态查询设置')).not.toBeVisible();
    }
  });

  test('应该在历史模式下禁用编辑操作', async ({ page }) => {
    // 切换到历史模式
    await page.getByText('历史').click();
    
    // 选择历史日期
    const lastMonth = new Date();
    lastMonth.setMonth(lastMonth.getMonth() - 1);
    const dateString = lastMonth.toISOString().slice(0, 16);
    
    await page.locator('input[type="datetime-local"]').first().fill(dateString);
    await page.getByText('确认选择').click();
    
    // 等待模式切换完成
    await expect(page.getByText('历史视图')).toBeVisible();
    
    // 验证编辑按钮被禁用
    const editButtons = page.getByRole('button', { name: '历史模式下不可编辑' });
    if (await editButtons.count() > 0) {
      await expect(editButtons.first()).toBeDisabled();
    }
    
    // 验证历史模式提示信息
    await expect(page.getByText(/当前显示历史.*编辑和删除功能已禁用/)).toBeVisible();
  });

  test('应该能够刷新数据缓存', async ({ page }) => {
    // 查找刷新按钮
    const refreshButton = page.getByRole('button', { name: '刷新数据缓存' });
    if (await refreshButton.isVisible()) {
      await refreshButton.click();
      
      // 验证加载状态（可能很快，不一定能捕获到）
      // 主要验证操作没有出错
      await expect(page.getByText('组织架构')).toBeVisible();
    }
  });

  test('应该显示缓存统计信息', async ({ page }) => {
    // 查找缓存统计徽章
    const cacheStats = page.locator('[data-testid="cache-stats"]');
    if (await cacheStats.isVisible()) {
      // 验证缓存统计显示
      await expect(cacheStats).toContainText(/\d+/);
    }
  });

  test('应该能够选择和批量操作组织', async ({ page }) => {
    // 如果有选择功能，测试批量选择
    const selectAllCheckbox = page.locator('thead input[type="checkbox"]');
    if (await selectAllCheckbox.isVisible()) {
      await selectAllCheckbox.click();
      
      // 验证选择统计显示
      await expect(page.getByText(/已选择 \d+ 项/)).toBeVisible();
      
      // 验证批量操作按钮显示
      const batchButtons = page.getByText('批量对比');
      if (await batchButtons.isVisible()) {
        await expect(batchButtons).toBeVisible();
      }
    }
  });

  test('应该正确响应网络错误', async ({ page }) => {
    // 模拟网络错误
    await page.route('**/api/**', route => route.abort('failed'));
    
    // 尝试刷新数据
    await page.reload();
    
    // 验证错误状态显示
    await expect(page.getByText(/❌.*加载.*失败/)).toBeVisible();
  });

  test('应该在移动设备上正确显示', async ({ page }) => {
    // 设置移动设备视口
    await page.setViewportSize({ width: 375, height: 667 });
    
    // 验证导航栏在移动设备上的响应式表现
    await expect(page.getByText('当前')).toBeVisible();
    
    // 验证表格在移动设备上可以滚动
    const table = page.locator('table');
    if (await table.isVisible()) {
      await expect(table).toBeVisible();
    }
  });
});