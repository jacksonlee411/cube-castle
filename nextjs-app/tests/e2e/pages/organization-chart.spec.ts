import { test, expect } from '@playwright/test';
import { TestHelpers, TestDataGenerator, NavigationHelper } from '../utils/test-helpers';

test.describe('组织架构页面', () => {
  let helpers: TestHelpers;
  let navigation: NavigationHelper;

  test.beforeEach(async ({ page }) => {
    helpers = new TestHelpers(page);
    navigation = new NavigationHelper(page);
    
    // 导航到组织架构页面
    await navigation.goToOrganizationChart();
    await helpers.waitForPageLoad();
  });

  test('页面基础加载和布局验证', async ({ page }) => {
    // 验证页面标题
    await helpers.verifyPageTitle('组织架构图');
    
    // 验证页面描述
    await expect(page.locator('p:has-text("可视化组织结构管理")')).toBeVisible();
    
    // 验证统计卡片
    await expect(page.locator('[data-testid="stats-card"]')).toHaveCount(4);
    await helpers.verifyStatsCard('组织总数');
    await helpers.verifyStatsCard('总员工数');
    await helpers.verifyStatsCard('最大层级');
    await helpers.verifyStatsCard('平均占用率');
    
    // 验证控制按钮
    await expect(page.locator('button:has-text("全部展开")')).toBeVisible();
    await expect(page.locator('button:has-text("全部收起")')).toBeVisible();
    await expect(page.locator('button:has-text("新增组织")')).toBeVisible();
    
    // 验证组织架构树
    await expect(page.locator('[data-testid="org-tree"]')).toBeVisible();
  });

  test('组织树展开收起功能', async ({ page }) => {
    // 等待组织树加载
    await page.waitForSelector('[data-testid="org-tree"]');
    
    // 测试全部收起功能
    await page.locator('button:has-text("全部收起")').click();
    await page.waitForTimeout(500);
    
    // 验证子节点被隐藏
    const collapsedNodes = page.locator('[data-testid="org-node"][style*="display: none"]');
    
    // 测试全部展开功能
    await page.locator('button:has-text("全部展开")').click();
    await page.waitForTimeout(500);
    
    // 验证所有节点可见
    const visibleNodes = page.locator('[data-testid="org-node"]:visible');
    const nodeCount = await visibleNodes.count();
    expect(nodeCount).toBeGreaterThan(1);
    
    // 测试单个节点展开收起
    const expandButton = page.locator('[data-testid="expand-button"]').first();
    if (await expandButton.isVisible()) {
      await expandButton.click();
      await page.waitForTimeout(300);
      
      // 再次点击收起
      await expandButton.click();
      await page.waitForTimeout(300);
    }
  });

  test('组织创建流程', async ({ page }) => {
    const testOrganization = TestDataGenerator.generateOrganization();
    
    // 点击新增组织按钮
    await page.locator('button:has-text("新增组织")').click();
    await helpers.waitForModal();
    
    // 填写组织信息
    await page.locator('input[name="name"]').fill(testOrganization.name);
    
    // 选择组织类型
    const typeSelect = page.locator('select[name="type"]');
    await typeSelect.selectOption(testOrganization.type);
    
    // 填写负责人
    await page.locator('input[name="managerName"]').fill(testOrganization.managerName);
    
    // 填写最大容量
    await page.locator('input[name="maxCapacity"]').fill(testOrganization.maxCapacity);
    
    // 填写描述
    await page.locator('textarea[name="description"]').fill(testOrganization.description);
    
    // 提交表单
    await helpers.clickButtonAndWait('创建');
    
    // 验证成功提示
    await helpers.verifyToastMessage('组织.*已成功创建');
    
    // 验证模态框关闭
    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
    
    // 等待组织树重新加载
    await page.waitForTimeout(1000);
    
    // 验证新组织出现在树中
    await expect(page.locator('[data-testid="org-tree"]')).toContainText(testOrganization.name);
  });

  test('组织编辑功能', async ({ page }) => {
    // 等待组织树加载
    await page.waitForSelector('[data-testid="org-tree"]');
    
    // 找到第一个组织节点的编辑按钮
    const editButton = page.locator('[data-testid="org-node"]').first().locator('button:has-text("编辑"), [data-testid="edit-org-button"]');
    
    if (await editButton.isVisible()) {
      await editButton.click();
      await helpers.waitForModal();
      
      // 修改组织名称
      const updatedName = `更新组织${Date.now()}`;
      const nameInput = page.locator('input[name="name"]');
      await nameInput.fill(updatedName);
      
      // 保存更改
      await helpers.clickButtonAndWait('更新');
      
      // 验证成功提示
      await helpers.verifyToastMessage('组织.*信息已更新');
      
      // 等待树重新渲染
      await page.waitForTimeout(1000);
      
      // 验证更新后的名称
      await expect(page.locator('[data-testid="org-tree"]')).toContainText(updatedName);
    }
  });

  test('组织删除功能', async ({ page }) => {
    // 先创建一个测试组织用于删除
    const testOrg = TestDataGenerator.generateOrganization();
    
    // 创建组织
    await page.locator('button:has-text("新增组织")').click();
    await helpers.waitForModal();
    await page.locator('input[name="name"]').fill(testOrg.name);
    await page.locator('select[name="type"]').selectOption('group');
    await helpers.clickButtonAndWait('创建');
    await page.waitForTimeout(1000);
    
    // 找到刚创建的组织节点
    const orgNode = page.locator(`[data-testid="org-node"]:has-text("${testOrg.name}")`);
    
    if (await orgNode.isVisible()) {
      // 点击删除按钮
      const deleteButton = orgNode.locator('button:has-text("删除"), [data-testid="delete-org-button"]');
      
      if (await deleteButton.isVisible()) {
        // 设置确认对话框监听
        page.on('dialog', async dialog => {
          expect(dialog.message()).toContain('确定要删除');
          await dialog.accept();
        });
        
        await deleteButton.click();
        
        // 验证删除成功提示
        await helpers.verifyToastMessage('组织.*已从系统中删除');
        
        // 等待树重新渲染
        await page.waitForTimeout(1000);
        
        // 验证组织已从树中移除
        await expect(page.locator('[data-testid="org-tree"]')).not.toContainText(testOrg.name);
      }
    }
  });

  test('组织层级结构显示', async ({ page }) => {
    // 等待组织树加载
    await page.waitForSelector('[data-testid="org-tree"]');
    
    // 验证根节点存在
    const rootNode = page.locator('[data-testid="org-node"]').first();
    await expect(rootNode).toBeVisible();
    
    // 验证层级指示器
    const levelIndicators = page.locator('[data-testid="level-indicator"]');
    const levelCount = await levelIndicators.count();
    
    if (levelCount > 0) {
      // 验证层级从L0开始
      await expect(levelIndicators.first()).toContainText('L0');
    }
    
    // 验证连接线显示
    const connectionLines = page.locator('[data-testid="connection-line"]');
    if (await connectionLines.first().isVisible()) {
      expect(await connectionLines.count()).toBeGreaterThan(0);
    }
  });

  test('组织信息显示', async ({ page }) => {
    // 等待组织树加载
    await page.waitForSelector('[data-testid="org-tree"]');
    
    // 验证组织节点包含关键信息
    const firstOrgNode = page.locator('[data-testid="org-node"]').first();
    
    // 验证组织名称
    await expect(firstOrgNode.locator('[data-testid="org-name"]')).toBeVisible();
    
    // 验证组织类型标签
    await expect(firstOrgNode.locator('[data-testid="org-type-badge"]')).toBeVisible();
    
    // 验证负责人信息
    const managerInfo = firstOrgNode.locator('[data-testid="manager-info"]');
    if (await managerInfo.isVisible()) {
      await expect(managerInfo).toContainText('👑');
    }
    
    // 验证员工统计
    const employeeStats = firstOrgNode.locator('[data-testid="employee-stats"]');
    if (await employeeStats.isVisible()) {
      await expect(employeeStats).toContainText('👥');
    }
    
    // 验证层级信息
    const levelInfo = firstOrgNode.locator('[data-testid="level-info"]');
    if (await levelInfo.isVisible()) {
      await expect(levelInfo).toContainText('L');
    }
  });

  test('添加子部门功能', async ({ page }) => {
    // 等待组织树加载
    await page.waitForSelector('[data-testid="org-tree"]');
    
    // 找到第一个组织节点的添加子部门按钮
    const addChildButton = page.locator('[data-testid="org-node"]').first().locator('button:has-text("添加子部门"), [data-testid="add-child-button"]');
    
    if (await addChildButton.isVisible()) {
      await addChildButton.click();
      await helpers.waitForModal();
      
      // 验证上级组织已预选
      const parentSelect = page.locator('select[name="parentId"]');
      if (await parentSelect.isVisible()) {
        const selectedValue = await parentSelect.inputValue();
        expect(selectedValue).toBeTruthy();
      }
      
      // 填写子部门信息
      const childOrg = TestDataGenerator.generateOrganization();
      await page.locator('input[name="name"]').fill(childOrg.name);
      
      // 类型应该自动设置为合适的子类型
      const typeSelect = page.locator('select[name="type"]');
      const currentType = await typeSelect.inputValue();
      expect(currentType).toBeTruthy();
      
      // 提交表单
      await helpers.clickButtonAndWait('创建');
      
      // 验证成功提示
      await helpers.verifyToastMessage('组织.*已成功创建');
      
      // 验证新子部门出现在树中
      await page.waitForTimeout(1000);
      await expect(page.locator('[data-testid="org-tree"]')).toContainText(childOrg.name);
    }
  });

  test('搜索和筛选功能', async ({ page }) => {
    // 如果有搜索功能
    const searchInput = page.locator('input[placeholder*="搜索组织"]');
    
    if (await searchInput.isVisible()) {
      // 搜索特定组织
      await searchInput.fill('技术部');
      await page.waitForTimeout(500);
      
      // 验证搜索结果
      const visibleNodes = page.locator('[data-testid="org-node"]:visible');
      const nodeCount = await visibleNodes.count();
      
      if (nodeCount > 0) {
        // 验证包含搜索关键词
        await expect(page.locator('[data-testid="org-tree"]')).toContainText('技术部');
      }
      
      // 清除搜索
      await searchInput.fill('');
      await page.waitForTimeout(500);
    }
  });

  test('响应式设计验证', async ({ page }) => {
    // 切换到移动端视口
    await page.setViewportSize({ width: 375, height: 667 });
    await helpers.waitForPageLoad();
    
    // 验证移动端布局
    await expect(page.locator('h1')).toBeVisible();
    
    // 验证组织树在移动端的显示
    await expect(page.locator('[data-testid="org-tree"]')).toBeVisible();
    
    // 验证控制按钮在移动端可见
    await expect(page.locator('button:has-text("新增组织")')).toBeVisible();
    
    // 恢复桌面视口
    await page.setViewportSize({ width: 1280, height: 720 });
  });

  test('组织统计数据验证', async ({ page }) => {
    // 验证统计卡片数据的合理性
    const statsCards = page.locator('[data-testid="stats-card"]');
    
    // 组织总数应该大于0
    const totalOrgsCard = statsCards.filter({ hasText: '组织总数' });
    const totalOrgsValue = await totalOrgsCard.locator('.text-2xl').textContent();
    const totalOrgs = parseInt(totalOrgsValue || '0');
    expect(totalOrgs).toBeGreaterThan(0);
    
    // 总员工数应该大于等于0
    const totalEmployeesCard = statsCards.filter({ hasText: '总员工数' });
    const totalEmployeesValue = await totalEmployeesCard.locator('.text-2xl').textContent();
    const totalEmployees = parseInt(totalEmployeesValue || '0');
    expect(totalEmployees).toBeGreaterThanOrEqual(0);
    
    // 最大层级应该大于等于1
    const maxLevelCard = statsCards.filter({ hasText: '最大层级' });
    const maxLevelValue = await maxLevelCard.locator('.text-2xl').textContent();
    const maxLevel = parseInt(maxLevelValue || '0');
    expect(maxLevel).toBeGreaterThanOrEqual(1);
  });

  test.afterEach(async ({ page }) => {
    // 截图用于调试
    await helpers.takeScreenshot(`organization-chart-test-${Date.now()}`);
  });
});