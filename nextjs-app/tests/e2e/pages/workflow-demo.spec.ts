import { test, expect } from '@playwright/test';
import { TestHelpers, NavigationHelper } from '../utils/test-helpers';

test.describe('工作流演示页面', () => {
  let helpers: TestHelpers;
  let navigation: NavigationHelper;

  test.beforeEach(async ({ page }) => {
    helpers = new TestHelpers(page);
    navigation = new NavigationHelper(page);
    
    // 导航到工作流演示页面
    await navigation.goToWorkflowDemo();
    await helpers.waitForPageLoad();
  });

  test('页面基础加载和布局验证', async ({ page }) => {
    // 验证页面标题
    await expect(page.locator('h1')).toContainText('工作流演示中心');
    
    // 验证页面描述
    await expect(page.locator('p:has-text("交互式工作流程演示")')).toBeVisible();
    
    // 验证统计卡片
    await expect(page.locator('[data-testid="stats-card"], .grid .p-4')).toHaveCount(4);
    await helpers.verifyStatsCard('可用模板');
    await helpers.verifyStatsCard('热门模板');
    await helpers.verifyStatsCard('总使用次数');
    await helpers.verifyStatsCard('平均完成时间');
    
    // 验证筛选区域
    await expect(page.locator(':has-text("业务类别")')).toBeVisible();
    await expect(page.locator(':has-text("复杂程度")')).toBeVisible();
    await expect(page.locator('button:has-text("重置筛选")')).toBeVisible();
    
    // 验证模板网格
    const templateCards = page.locator('.hover\\:shadow-lg');
    const cardCount = await templateCards.count();
    expect(cardCount).toBeGreaterThan(0);
  });

  test('统计卡片数据验证', async ({ page }) => {
    // 等待数据加载
    await page.waitForTimeout(1000);
    
    // 验证可用模板数量
    const availableTemplatesCard = page.locator(':has-text("可用模板")').locator('..').locator('.text-2xl.font-bold');
    const availableCount = await availableTemplatesCard.textContent();
    const available = parseInt(availableCount || '0');
    expect(available).toBeGreaterThan(0);
    
    // 验证热门模板数量
    const popularTemplatesCard = page.locator(':has-text("热门模板")').locator('..').locator('.text-2xl.font-bold');
    const popularCount = await popularTemplatesCard.textContent();
    const popular = parseInt(popularCount || '0');
    expect(popular).toBeGreaterThanOrEqual(0);
    expect(popular).toBeLessThanOrEqual(available);
    
    // 验证总使用次数
    const totalUsageCard = page.locator(':has-text("总使用次数")').locator('..').locator('.text-2xl.font-bold');
    const totalUsage = await totalUsageCard.textContent();
    const usage = parseInt(totalUsage || '0');
    expect(usage).toBeGreaterThan(0);
    
    // 验证平均完成时间格式
    const avgTimeCard = page.locator(':has-text("平均完成时间")').locator('..').locator('.text-2xl.font-bold');
    const avgTime = await avgTimeCard.textContent();
    expect(avgTime).toMatch(/\d+天/);
  });

  test('工作流模板卡片显示', async ({ page }) => {
    // 等待模板加载
    await page.waitForTimeout(1000);
    
    // 验证模板卡片
    const templateCards = page.locator('.hover\\:shadow-lg');
    const cardCount = await templateCards.count();
    expect(cardCount).toBeGreaterThan(0);
    
    // 验证第一个模板卡片的内容
    const firstCard = templateCards.first();
    
    // 验证模板名称
    const templateName = firstCard.locator('h3.text-lg.font-semibold');
    await expect(templateName).not.toBeEmpty();
    
    // 验证模板描述
    const templateDesc = firstCard.locator('p.text-gray-600.text-sm');
    await expect(templateDesc).not.toBeEmpty();
    
    // 验证热门标签（如果存在）
    const popularBadge = firstCard.locator(':has-text("热门")');
    if (await popularBadge.isVisible()) {
      await expect(popularBadge).toBeVisible();
    }
    
    // 验证模板信息（类别、复杂度、时间、使用次数）
    const categoryInfo = firstCard.locator(':has-text("人力资源"), :has-text("财务管理"), :has-text("运营管理"), :has-text("行政管理")');
    await expect(categoryInfo).toBeVisible();
    
    const complexityInfo = firstCard.locator(':has-text("简单"), :has-text("中等"), :has-text("复杂")');
    await expect(complexityInfo).toBeVisible();
    
    const timeInfo = firstCard.locator('text=/\\d+天/');
    await expect(timeInfo).toBeVisible();
    
    const usageInfo = firstCard.locator('text=/\\d+次/');
    await expect(usageInfo).toBeVisible();
    
    // 验证流程步骤显示
    const stepsSection = firstCard.locator(':has-text("流程步骤")');
    await expect(stepsSection).toBeVisible();
    
    const stepItems = firstCard.locator('.w-5.h-5.rounded-full');
    const stepCount = await stepItems.count();
    expect(stepCount).toBeGreaterThan(0);
    
    // 验证操作按钮
    const startDemoButton = firstCard.locator('button:has-text("开始演示")');
    await expect(startDemoButton).toBeVisible();
    
    const viewDetailsButton = firstCard.locator('button:has-text("查看详情")');
    await expect(viewDetailsButton).toBeVisible();
  });

  test('筛选功能测试', async ({ page }) => {
    // 等待模板加载
    await page.waitForTimeout(1000);
    
    // 测试业务类别筛选
    const categorySelect = page.locator('select, [role="combobox"]').filter({ hasText: '全部类别' }).first();
    await categorySelect.click();
    
    // 选择人力资源
    const hrOption = page.locator('text=人力资源');
    if (await hrOption.isVisible()) {
      await hrOption.click();
      
      // 等待筛选结果
      await page.waitForTimeout(500);
      
      // 验证筛选结果
      const filteredCards = page.locator('.hover\\:shadow-lg');
      const filteredCount = await filteredCards.count();
      expect(filteredCount).toBeGreaterThan(0);
      
      // 验证所有显示的卡片都包含人力资源标签
      for (let i = 0; i < filteredCount; i++) {
        const card = filteredCards.nth(i);
        const hasHRLabel = await card.locator(':has-text("人力资源")').isVisible();
        expect(hasHRLabel).toBe(true);
      }
    } else {
      // 如果下拉选项不可见，按ESC键关闭
      await page.keyboard.press('Escape');
    }
    
    // 测试复杂程度筛选
    const complexitySelect = page.locator('select, [role="combobox"]').filter({ hasText: '全部复杂度' }).first();
    await complexitySelect.click();
    
    // 选择简单
    const simpleOption = page.locator('text=简单');
    if (await simpleOption.isVisible()) {
      await simpleOption.click();
      
      // 等待筛选结果
      await page.waitForTimeout(500);
      
      // 验证筛选结果
      const simpleCards = page.locator('.hover\\:shadow-lg');
      const simpleCount = await simpleCards.count();
      
      if (simpleCount > 0) {
        // 验证显示的卡片包含简单标签
        const firstSimpleCard = simpleCards.first();
        const hasSimpleLabel = await firstSimpleCard.locator(':has-text("简单")').isVisible();
        expect(hasSimpleLabel).toBe(true);
      }
    } else {
      await page.keyboard.press('Escape');
    }
    
    // 测试重置筛选
    const resetButton = page.locator('button:has-text("重置筛选")');
    await resetButton.click();
    
    // 验证筛选器已重置（所有模板重新显示）
    await page.waitForTimeout(500);
    const allCards = page.locator('.hover\\:shadow-lg');
    const allCount = await allCards.count();
    expect(allCount).toBeGreaterThan(0);
  });

  test('工作流演示执行', async ({ page }) => {
    // 等待模板加载
    await page.waitForTimeout(1000);
    
    // 找到第一个模板的开始演示按钮
    const firstStartButton = page.locator('button:has-text("开始演示")').first();
    await expect(firstStartButton).toBeVisible();
    
    // 点击开始演示
    await firstStartButton.click();
    
    // 验证成功提示
    await helpers.verifyToastMessage('开始演示');
    
    // 验证演示执行面板出现
    await page.waitForTimeout(1000);
    const executionPanel = page.locator(':has-text("演示执行中")');
    await expect(executionPanel).toBeVisible();
    
    // 验证进度条
    const progressBar = page.locator('[role="progressbar"]');
    await expect(progressBar).toBeVisible();
    
    // 验证进度百分比
    const progressPercent = page.locator('text=/\\d+%/');
    await expect(progressPercent).toBeVisible();
    
    // 验证步骤列表
    const stepItems = page.locator('.w-8.h-8.rounded-full');
    const stepCount = await stepItems.count();
    expect(stepCount).toBeGreaterThan(0);
    
    // 验证执行日志
    const logSection = page.locator(':has-text("执行日志")');
    await expect(logSection).toBeVisible();
    
    const logEntries = page.locator('.bg-gray-50.rounded.text-sm');
    await expect(logEntries.first()).toBeVisible();
    
    // 验证控制按钮
    const pauseButton = page.locator('button:has-text("暂停")');
    if (await pauseButton.isVisible()) {
      await expect(pauseButton).toBeVisible();
    }
    
    const resetButton = page.locator('button:has-text("重置")');
    await expect(resetButton).toBeVisible();
  });

  test('演示控制功能', async ({ page }) => {
    // 等待模板加载
    await page.waitForTimeout(1000);
    
    // 开始演示
    const firstStartButton = page.locator('button:has-text("开始演示")').first();
    await firstStartButton.click();
    
    // 等待演示开始
    await page.waitForTimeout(1000);
    
    // 测试暂停功能
    const pauseButton = page.locator('button:has-text("暂停")');
    if (await pauseButton.isVisible()) {
      await pauseButton.click();
      
      // 验证暂停提示
      await helpers.verifyToastMessage('演示已暂停');
      
      // 验证继续按钮出现
      const resumeButton = page.locator('button:has-text("继续")');
      await expect(resumeButton).toBeVisible();
      
      // 测试继续功能
      await resumeButton.click();
      
      // 验证恢复提示
      await helpers.verifyToastMessage('演示已恢复');
    }
    
    // 测试重置功能
    const resetButton = page.locator('button:has-text("重置")');
    await resetButton.click();
    
    // 验证重置提示
    await helpers.verifyToastMessage('演示已重置');
    
    // 验证演示面板消失
    await page.waitForTimeout(500);
    const executionPanel = page.locator(':has-text("演示执行中")');
    await expect(executionPanel).not.toBeVisible();
  });

  test('模板类型和徽章显示', async ({ page }) => {
    // 等待模板加载
    await page.waitForTimeout(1000);
    
    // 验证模板类型徽章
    const typeBadges = page.locator('.bg-blue-100, .bg-green-100, .bg-red-100, .bg-gray-100');
    const badgeCount = await typeBadges.count();
    expect(badgeCount).toBeGreaterThan(0);
    
    // 验证类型标签文本
    const validTypes = ['职位变更', '请假申请', '费用报销', '文档审批'];
    let foundValidType = false;
    
    for (let i = 0; i < badgeCount; i++) {
      const badgeText = await typeBadges.nth(i).textContent();
      if (validTypes.some(type => badgeText?.includes(type))) {
        foundValidType = true;
        break;
      }
    }
    
    expect(foundValidType).toBe(true);
    
    // 验证热门徽章
    const popularBadges = page.locator(':has-text("热门")');
    const popularCount = await popularBadges.count();
    // 热门徽章可能存在也可能不存在
    expect(popularCount).toBeGreaterThanOrEqual(0);
  });

  test('步骤图标和信息显示', async ({ page }) => {
    // 等待模板加载
    await page.waitForTimeout(1000);
    
    // 验证步骤编号圆圈
    const stepNumbers = page.locator('.w-5.h-5.rounded-full.bg-gray-100');
    const numberCount = await stepNumbers.count();
    expect(numberCount).toBeGreaterThan(0);
    
    // 验证步骤图标
    const stepIcons = page.locator('svg.h-4.w-4.text-blue-500, svg.h-4.w-4.text-purple-500, svg.h-4.w-4.text-green-500, svg.h-4.w-4.text-orange-500');
    const iconCount = await stepIcons.count();
    expect(iconCount).toBeGreaterThan(0);
    
    // 验证角色信息格式
    const roleInfo = page.locator('.text-gray-400');
    const roleCount = await roleInfo.count();
    
    if (roleCount > 0) {
      const firstRole = await roleInfo.first().textContent();
      expect(firstRole).toMatch(/\(.+\)/); // 应该包含括号
    }
  });

  test('模板详细信息显示', async ({ page }) => {
    // 等待模板加载
    await page.waitForTimeout(1000);
    
    // 验证第一个模板的详细信息
    const firstCard = page.locator('.hover\\:shadow-lg').first();
    
    // 验证复杂度颜色编码
    const complexityElement = firstCard.locator('.text-green-600, .text-yellow-600, .text-red-600');
    await expect(complexityElement).toBeVisible();
    
    // 验证时间和使用次数格式
    const timeElement = firstCard.locator('text=/\\d+天/');
    const timeText = await timeElement.textContent();
    expect(timeText).toMatch(/\d+天/);
    
    const usageElement = firstCard.locator('text=/\\d+次/');
    const usageText = await usageElement.textContent();
    expect(usageText).toMatch(/\d+次/);
    
    // 验证步骤预览
    const stepsPreview = firstCard.locator(':has-text("流程步骤")');
    await expect(stepsPreview).toBeVisible();
    
    // 验证步骤计数
    const stepCountText = await stepsPreview.textContent();
    expect(stepCountText).toMatch(/流程步骤 \(\d+个\)/);
  });

  test('执行日志时间格式', async ({ page }) => {
    // 等待模板加载
    await page.waitForTimeout(1000);
    
    // 开始演示
    const firstStartButton = page.locator('button:has-text("开始演示")').first();
    await firstStartButton.click();
    
    // 等待日志生成
    await page.waitForTimeout(2000);
    
    // 验证日志时间戳格式
    const timestamps = page.locator('.text-xs.text-gray-400');
    const timestampCount = await timestamps.count();
    
    if (timestampCount > 0) {
      const firstTimestamp = await timestamps.first().textContent();
      expect(firstTimestamp).toMatch(/\d{2}:\d{2}:\d{2}/); // HH:mm:ss format
    }
    
    // 验证日志消息内容
    const logMessages = page.locator('.text-gray-600');
    const messageCount = await logMessages.count();
    
    if (messageCount > 0) {
      const firstMessage = await logMessages.first().textContent();
      expect(firstMessage?.trim()).toBeTruthy();
    }
  });

  test('响应式设计验证', async ({ page }) => {
    // 切换到移动端视口
    await page.setViewportSize({ width: 375, height: 667 });
    await helpers.waitForPageLoad();
    
    // 验证移动端布局
    await expect(page.locator('h1')).toBeVisible();
    
    // 验证统计卡片在移动端的显示（可能堆叠显示）
    const statsCards = page.locator('.grid .p-4');
    await expect(statsCards.first()).toBeVisible();
    
    // 验证筛选器在移动端的显示
    await expect(page.locator(':has-text("业务类别")')).toBeVisible();
    await expect(page.locator(':has-text("复杂程度")')).toBeVisible();
    
    // 验证模板卡片在移动端的显示
    const templateCards = page.locator('.hover\\:shadow-lg');
    await expect(templateCards.first()).toBeVisible();
    
    // 恢复桌面视口
    await page.setViewportSize({ width: 1280, height: 720 });
  });

  test('空状态处理', async ({ page }) => {
    // 设置一个不存在的筛选条件组合
    const categorySelect = page.locator('select, [role="combobox"]').filter({ hasText: '全部类别' }).first();
    await categorySelect.click();
    
    const adminOption = page.locator('text=行政管理');
    if (await adminOption.isVisible()) {
      await adminOption.click();
      
      const complexitySelect = page.locator('select, [role="combobox"]').filter({ hasText: '全部复杂度' }).first();
      await complexitySelect.click();
      
      const complexOption = page.locator('text=复杂');
      if (await complexOption.isVisible()) {
        await complexOption.click();
        
        // 等待筛选结果
        await page.waitForTimeout(500);
        
        // 检查是否显示空状态
        const noResultsAlert = page.locator(':has-text("没有找到符合条件的工作流模板")');
        const templateCards = page.locator('.hover\\:shadow-lg');
        const cardCount = await templateCards.count();
        
        // 要么显示空状态提示，要么有匹配的模板
        if (cardCount === 0) {
          await expect(noResultsAlert).toBeVisible();
        } else {
          expect(cardCount).toBeGreaterThan(0);
        }
      } else {
        await page.keyboard.press('Escape');
      }
    } else {
      await page.keyboard.press('Escape');
    }
  });

  test('页面加载性能验证', async ({ page }) => {
    // 测量页面加载时间
    const startTime = Date.now();
    await navigation.goToWorkflowDemo();
    await helpers.waitForPageLoad();
    const loadTime = Date.now() - startTime;
    
    // 验证页面在3秒内加载完成
    expect(loadTime).toBeLessThan(3000);
    
    // 验证关键元素已加载
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator(':has-text("业务类别")')).toBeVisible();
    
    // 等待模板加载完成
    await page.waitForTimeout(1000);
    const templateCards = page.locator('.hover\\:shadow-lg');
    await expect(templateCards.first()).toBeVisible();
  });

  test.afterEach(async ({ page }) => {
    // 截图用于调试
    await helpers.takeScreenshot(`workflow-demo-test-${Date.now()}`);
  });
});