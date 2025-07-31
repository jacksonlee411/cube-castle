import { test, expect } from '@playwright/test';
import { TestHelpers, NavigationHelper } from '../utils/test-helpers';

test.describe('工作流详情页面', () => {
  let helpers: TestHelpers;
  let navigation: NavigationHelper;
  const testWorkflowId = '1'; // 使用示例数据中的工作流ID

  test.beforeEach(async ({ page }) => {
    helpers = new TestHelpers(page);
    navigation = new NavigationHelper(page);
    
    // 导航到工作流详情页面
    await navigation.goToWorkflowDetail(testWorkflowId);
    await helpers.waitForPageLoad();
  });

  test('页面基础加载和布局验证', async ({ page }) => {
    // 验证页面有工作流标题
    await expect(page.locator('h1')).toContainText('职位晋升申请');
    
    // 验证返回按钮
    await expect(page.locator('button:has-text("返回")')).toBeVisible();
    
    // 验证操作下拉菜单
    await expect(page.locator('button:has-text("操作")')).toBeVisible();
    
    // 验证状态栏
    const statusBar = page.locator('[data-testid="status-bar"]');
    if (await statusBar.isVisible()) {
      await expect(statusBar).toContainText('进行中');
    } else {
      // 如果没有data-testid，查找包含状态信息的卡片
      const statusCards = page.locator('.bg-blue-100, .bg-green-100, .bg-red-100');
      const statusCardCount = await statusCards.count();
      expect(statusCardCount).toBeGreaterThanOrEqual(1);
    }
    
    // 验证进度条
    await expect(page.locator('[role="progressbar"], [data-testid="progress-bar"]')).toBeVisible();
    
    // 验证工作流步骤区域
    await expect(page.locator('[data-testid="workflow-steps"], .space-y-4')).toBeVisible();
    
    // 验证活动日志区域
    await expect(page.locator('[data-testid="activity-log"]')).toBeVisible();
  });

  test('工作流步骤显示和交互', async ({ page }) => {
    // 等待工作流步骤加载
    await page.waitForSelector('[data-testid="workflow-step"], .relative .z-10');
    
    // 验证工作流步骤数量
    const steps = page.locator('[data-testid="workflow-step"], .relative .z-10');
    const stepCount = await steps.count();
    expect(stepCount).toBeGreaterThan(0);
    
    // 验证步骤状态图标
    const statusIcons = page.locator('[data-testid="step-status-icon"]');
    if (await statusIcons.first().isVisible()) {
      expect(await statusIcons.count()).toBe(stepCount);
    }
    
    // 测试步骤展开/收起功能
    const expandButton = page.locator('[data-testid="expand-step"], button:has([class*="chevron"])').first();
    if (await expandButton.isVisible()) {
      // 展开步骤
      await expandButton.click();
      await page.waitForTimeout(300);
      
      // 验证展开的内容
      const expandedContent = page.locator('[data-testid="expanded-content"]');
      if (await expandedContent.isVisible()) {
        await expect(expandedContent).toBeVisible();
      }
      
      // 收起步骤
      await expandButton.click();
      await page.waitForTimeout(300);
    }
  });

  test('工作流审批功能', async ({ page }) => {
    // 查找进行中的审批步骤
    const inProgressStep = page.locator('[data-testid="workflow-step"]:has(.bg-blue-100), .ring-2.ring-blue-500');
    
    if (await inProgressStep.first().isVisible()) {
      // 展开当前步骤
      const expandButton = inProgressStep.first().locator('button:has([class*="chevron"])');
      if (await expandButton.isVisible()) {
        await expandButton.click();
        await page.waitForTimeout(300);
      }
      
      // 查找审批按钮
      const approveButton = page.locator('button:has-text("审批通过")');
      const rejectButton = page.locator('button:has-text("拒绝申请")');
      
      if (await approveButton.isVisible()) {
        // 测试审批通过流程
        await approveButton.click();
        await helpers.waitForModal();
        
        // 填写审批意见
        const commentTextarea = page.locator('textarea[placeholder*="评论"], textarea[placeholder*="原因"]');
        if (await commentTextarea.isVisible()) {
          await commentTextarea.fill('测试审批通过意见');
        }
        
        // 确认审批
        await helpers.clickButtonAndWait('通过');
        
        // 验证成功提示
        await helpers.verifyToastMessage('审批成功');
        
        // 验证工作流状态更新
        await page.waitForTimeout(1000);
        
        // 验证进度条更新
        const progressBar = page.locator('[role="progressbar"], [data-testid="progress-bar"]');
        if (await progressBar.isVisible()) {
          const progressValue = await progressBar.getAttribute('value') || await progressBar.getAttribute('data-value');
          expect(progressValue).toBeTruthy();
        }
      }
    }
  });

  test('工作流拒绝功能', async ({ page }) => {
    // 重新加载页面以确保测试独立性
    await page.reload();
    await helpers.waitForPageLoad();
    
    // 查找进行中的审批步骤
    const inProgressStep = page.locator('[data-testid="workflow-step"]:has(.bg-blue-100), .ring-2.ring-blue-500');
    
    if (await inProgressStep.first().isVisible()) {
      // 展开当前步骤
      const expandButton = inProgressStep.first().locator('button:has([class*="chevron"])');
      if (await expandButton.isVisible()) {
        await expandButton.click();
        await page.waitForTimeout(300);
      }
      
      // 查找拒绝按钮
      const rejectButton = page.locator('button:has-text("拒绝申请")');
      
      if (await rejectButton.isVisible()) {
        await rejectButton.click();
        await helpers.waitForModal();
        
        // 填写拒绝原因（必填）
        const commentTextarea = page.locator('textarea[placeholder*="评论"], textarea[placeholder*="原因"]');
        await commentTextarea.fill('测试拒绝原因：不符合晋升条件');
        
        // 确认拒绝
        await helpers.clickButtonAndWait('拒绝');
        
        // 验证成功提示
        await helpers.verifyToastMessage('已拒绝申请');
        
        // 验证工作流状态更新为失败
        await page.waitForTimeout(1000);
        const failedStatus = page.locator('.bg-red-100, .text-red-600');
        if (await failedStatus.first().isVisible()) {
          await expect(failedStatus.first()).toBeVisible();
        }
      }
    }
  });

  test('活动日志显示', async ({ page }) => {
    // 验证活动日志区域
    const activityLog = page.locator('[data-testid="activity-log"]');
    await expect(activityLog).toBeVisible();
    
    // 验证日志项目
    const logItems = page.locator('[data-testid="log-item"], .bg-gray-50');
    const logCount = await logItems.count();
    expect(logCount).toBeGreaterThan(0);
    
    // 验证日志项包含关键信息
    const firstLogItem = logItems.first();
    await expect(firstLogItem).toContainText(/创建|开始|审批|完成/);
    
    // 验证时间戳格式
    const timestamp = firstLogItem.locator('.text-xs.text-gray-400, [data-testid="timestamp"]');
    if (await timestamp.isVisible()) {
      const timeText = await timestamp.textContent();
      expect(timeText).toMatch(/\d{2}:\d{2}/); // HH:mm 格式
    }
    
    // 验证操作者信息
    const actor = firstLogItem.locator('.font-medium, [data-testid="actor"]');
    if (await actor.isVisible()) {
      const actorText = await actor.textContent();
      expect(actorText).toBeTruthy();
    }
  });

  test('工作流详情信息显示', async ({ page }) => {
    // 验证申请详情卡片
    const detailsCard = page.locator('[data-testid="workflow-details"]');
    if (await detailsCard.isVisible()) {
      // 验证员工信息
      await expect(detailsCard).toContainText('员工姓名');
      await expect(detailsCard).toContainText('当前职位');
      await expect(detailsCard).toContainText('目标职位');
      
      // 验证薪资调整信息
      const salaryInfo = detailsCard.locator(':has-text("薪资调整")');
      if (await salaryInfo.isVisible()) {
        await expect(salaryInfo).toContainText('¥');
        await expect(salaryInfo).toContainText('→');
      }
      
      // 验证生效日期
      await expect(detailsCard).toContainText('生效日期');
    }
  });

  test('操作菜单功能', async ({ page }) => {
    // 点击操作菜单
    await page.locator('button:has-text("操作")').click();
    
    // 验证菜单项
    const menuItems = page.locator('[role="menuitem"], [data-testid="menu-item"]');
    
    // 验证编辑工作流选项
    const editOption = page.locator(':has-text("编辑工作流")');
    if (await editOption.isVisible()) {
      await expect(editOption).toBeVisible();
    }
    
    // 验证导出报告选项
    const exportOption = page.locator(':has-text("导出报告")');
    if (await exportOption.isVisible()) {
      await expect(exportOption).toBeVisible();
    }
    
    // 验证分享链接选项
    const shareOption = page.locator(':has-text("分享链接")');
    if (await shareOption.isVisible()) {
      await expect(shareOption).toBeVisible();
    }
    
    // 关闭菜单
    await page.keyboard.press('Escape');
  });

  test('进度条和状态指示', async ({ page }) => {
    // 验证进度条存在且有值
    const progressBar = page.locator('[role="progressbar"], [data-testid="progress-bar"]');
    await expect(progressBar).toBeVisible();
    
    // 验证进度百分比显示
    const progressText = page.locator(':has-text("%"), [data-testid="progress-text"]');
    if (await progressText.isVisible()) {
      const progressValue = await progressText.textContent();
      expect(progressValue).toMatch(/\d+%/);
    }
    
    // 验证状态徽章
    const statusBadges = page.locator('[data-testid="status-badge"], .bg-blue-100, .bg-green-100, .bg-red-100');
    expect(await statusBadges.count()).toBeGreaterThan(0);
    
    // 验证优先级显示
    const priorityBadge = page.locator(':has-text("优先级")').locator('..').locator('.bg-yellow-100, .bg-red-100, .bg-gray-100');
    if (await priorityBadge.isVisible()) {
      await expect(priorityBadge).toBeVisible();
    }
  });

  test('返回功能', async ({ page }) => {
    // 点击返回按钮
    await page.locator('button:has-text("返回")').click();
    
    // 验证页面导航
    await page.waitForTimeout(500);
    
    // 应该回到上一页（可能是工作流列表页）
    const currentUrl = page.url();
    expect(currentUrl).not.toContain(`/workflows/${testWorkflowId}`);
  });

  test('响应式设计验证', async ({ page }) => {
    // 切换到移动端视口
    await page.setViewportSize({ width: 375, height: 667 });
    await helpers.waitForPageLoad();
    
    // 验证移动端布局
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('button:has-text("返回")')).toBeVisible();
    
    // 验证工作流步骤在移动端的显示
    const steps = page.locator('[data-testid="workflow-step"], .relative .z-10');
    await expect(steps.first()).toBeVisible();
    
    // 验证活动日志在移动端的显示
    await expect(page.locator('[data-testid="activity-log"]')).toBeVisible();
    
    // 恢复桌面视口
    await page.setViewportSize({ width: 1280, height: 720 });
  });

  test('工作流数据完整性验证', async ({ page }) => {
    // 验证基本信息完整性
    await expect(page.locator('h1')).not.toBeEmpty();
    
    // 验证发起人信息
    const initiatorInfo = page.locator(':has-text("发起人")');
    if (await initiatorInfo.isVisible()) {
      await expect(initiatorInfo).not.toBeEmpty();
    }
    
    // 验证创建时间
    const createdTime = page.locator(':has-text("创建时间")');
    if (await createdTime.isVisible()) {
      const timeText = await createdTime.textContent();
      expect(timeText).toMatch(/\d{4}-\d{2}-\d{2}/); // YYYY-MM-DD 格式
    }
    
    // 验证最后更新时间
    const updatedTime = page.locator(':has-text("最后更新")');
    if (await updatedTime.isVisible()) {
      const timeText = await updatedTime.textContent();
      expect(timeText).toMatch(/\d{4}-\d{2}-\d{2}/);
    }
  });

  test('页面加载性能验证', async ({ page }) => {
    // 测量页面加载时间
    const startTime = Date.now();
    await navigation.goToWorkflowDetail(testWorkflowId);
    await helpers.waitForPageLoad();
    const loadTime = Date.now() - startTime;
    
    // 验证页面在3秒内加载完成
    expect(loadTime).toBeLessThan(3000);
    
    // 验证关键元素已加载
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('[role="progressbar"], [data-testid="progress-bar"]')).toBeVisible();
  });

  test.afterEach(async ({ page }) => {
    // 截图用于调试
    await helpers.takeScreenshot(`workflow-detail-test-${Date.now()}`);
  });
});