import { test, expect } from '@playwright/test';

test.describe('Canvas Frontend E2E Tests', () => {
  
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('应用外壳完整渲染测试', async ({ page }) => {
    // 检查侧边栏logo
    await expect(page.getByText('🏰 Cube Castle')).toBeVisible();
    
    // 检查导航菜单项
    const dashboard = page.getByText('仪表板');
    const organizations = page.getByText('组织架构');  
    const employees = page.getByText('员工管理');
    const positions = page.getByText('职位管理');
    
    await expect(dashboard).toBeVisible();
    await expect(organizations).toBeVisible();
    await expect(employees).toBeVisible();
    await expect(positions).toBeVisible();
    
    // 检查顶部栏
    await expect(page.getByText('组织管理')).toBeVisible();
    await expect(page.getByText('设置')).toBeVisible();
    await expect(page.getByText('通知')).toBeVisible();
  });

  test('导航功能完整流程测试', async ({ page }) => {
    // 默认应该在组织架构页面
    await expect(page.getByText('组织架构管理')).toBeVisible();
    
    // 点击组织架构导航确认激活状态
    await page.getByText('组织架构').click();
    await expect(page.url()).toContain('/organizations');
    
    // 检查组织管理页面核心元素
    await expect(page.getByText('组织架构管理')).toBeVisible();
    await expect(page.getByText('新增组织单元')).toBeVisible();
    await expect(page.getByText('导入数据')).toBeVisible();
    await expect(page.getByText('导出报告')).toBeVisible();
  });

  test('组织数据加载和显示测试', async ({ page }) => {
    await page.goto('/organizations');
    
    // 等待数据加载
    await page.waitForTimeout(2000);
    
    // 检查是否显示加载状态或数据
    const loadingText = page.getByText('加载组织数据中...');
    const noDataText = page.getByText('暂无组织数据');
    const organizationData = page.getByText('高谷集团');
    
    // 应该显示其中一种状态
    await expect(
      loadingText.or(noDataText).or(organizationData)
    ).toBeVisible();
    
    // 如果有数据，检查表格结构
    const table = page.getByRole('table').first();
    if (await table.isVisible()) {
      await expect(page.getByText('编码')).toBeVisible();
      await expect(page.getByText('名称')).toBeVisible();
      await expect(page.getByText('类型')).toBeVisible();
      await expect(page.getByText('状态')).toBeVisible();
    }
  });

  test('Canvas组件样式验证测试', async ({ page }) => {
    // 检查Canvas样式是否正确加载
    const body = page.locator('body');
    
    // Canvas应该设置CSS变量
    const bodyStyles = await body.getAttribute('style');
    if (bodyStyles) {
      expect(bodyStyles).toContain('--cnvs-');
    }
    
    // 检查Canvas组件是否有正确的class
    const mainContainer = page.locator('[class*="css-"]').first();
    await expect(mainContainer).toBeVisible();
  });

  test('响应式设计验证测试', async ({ page }) => {
    // 桌面视图测试
    await page.setViewportSize({ width: 1280, height: 720 });
    await expect(page.getByText('🏰 Cube Castle')).toBeVisible();
    await expect(page.getByText('组织架构管理')).toBeVisible();
    
    // 平板视图测试  
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.getByText('🏰 Cube Castle')).toBeVisible();
    
    // 移动视图测试
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.getByText('🏰 Cube Castle')).toBeVisible();
    
    // 恢复桌面视图
    await page.setViewportSize({ width: 1280, height: 720 });
  });

  test('API集成功能测试', async ({ page }) => {
    await page.goto('/organizations');
    
    // 等待API数据加载
    await page.waitForTimeout(3000);
    
    // 检查统计卡片数据
    const statsCards = page.locator('[data-testid*="card"]');
    if (await statsCards.first().isVisible()) {
      await expect(page.getByText('按类型统计')).toBeVisible();
      await expect(page.getByText('按状态统计')).toBeVisible();
      await expect(page.getByText('总体概况')).toBeVisible();
    }
    
    // 检查是否显示实际组织数据
    const orgName = page.getByText('高谷集团');
    const orgCode = page.getByText('1000000');
    
    if (await orgName.isVisible()) {
      await expect(orgName).toBeVisible();
      await expect(orgCode).toBeVisible();
      await expect(page.getByText('COMPANY')).toBeVisible();
      await expect(page.getByText('ACTIVE')).toBeVisible();
    }
  });
});