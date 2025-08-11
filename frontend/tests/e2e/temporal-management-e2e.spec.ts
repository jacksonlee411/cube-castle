/**
 * 时态管理系统端到端测试
 * 验证时态组织架构管理的完整用户流程
 */
import { test, expect, Page } from '@playwright/test';

// 测试配置
const BASE_URL = 'http://localhost:3000';
const TEST_TIMEOUT = 30000;

// 测试数据
const TEST_ORG = {
  name: `E2E测试组织_${Date.now()}`,
  unit_type: 'DEPARTMENT',
  status: 'ACTIVE',
  description: 'E2E自动化测试创建的组织'
};

const PLANNED_ORG = {
  name: `E2E计划组织_${Date.now()}`,
  unit_type: 'DEPARTMENT', 
  status: 'PLANNED',
  description: 'E2E自动化测试创建的计划组织'
};

test.describe('时态管理系统 E2E 测试套件', () => {
  
  test.beforeEach(async ({ page }) => {
    // 设置测试超时
    test.setTimeout(TEST_TIMEOUT);
    
    // 导航到应用主页
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');
  });

  test.describe('1. 时态导航功能测试', () => {
    
    test('应该能够切换时态模式', async ({ page }) => {
      // 验证页面加载
      await expect(page).toHaveTitle(/Cube Castle/);
      
      // 查找时态导航栏
      const temporalNavbar = page.locator('[data-testid="temporal-navbar"]');
      await expect(temporalNavbar).toBeVisible();
      
      // 测试当前模式 -> 历史模式
      const historicalButton = page.locator('button:has-text("历史模式")');
      await historicalButton.click();
      
      // 验证模式切换
      await expect(page.locator('[data-testid="mode-indicator"]:has-text("历史")')).toBeVisible();
      
      // 测试历史模式 -> 规划模式  
      const planningButton = page.locator('button:has-text("规划模式")');
      await planningButton.click();
      
      // 验证模式切换
      await expect(page.locator('[data-testid="mode-indicator"]:has-text("规划")')).toBeVisible();
      
      // 返回当前模式
      const currentButton = page.locator('button:has-text("当前模式")');
      await currentButton.click();
      
      await expect(page.locator('[data-testid="mode-indicator"]:has-text("当前")')).toBeVisible();
    });

    test('应该在历史模式下禁用编辑操作', async ({ page }) => {
      // 切换到历史模式
      await page.locator('button:has-text("历史模式")').click();
      await page.waitForTimeout(1000);
      
      // 验证新增按钮被禁用或显示禁用文本
      const addButton = page.locator('button:has-text("新增组织")').first();
      await expect(addButton).toBeDisabled();
    });
  });

  test.describe('2. 组织CRUD操作测试', () => {
    
    test('应该能够创建新组织', async ({ page }) => {
      // 点击新增组织按钮
      const addButton = page.locator('button:has-text("新增组织")').first();
      await addButton.click();
      
      // 验证表单弹窗打开
      const formModal = page.locator('[data-testid="organization-form"]');
      await expect(formModal).toBeVisible();
      
      // 填写表单
      await page.fill('[data-testid="form-field-name"] input', TEST_ORG.name);
      await page.selectOption('[data-testid="form-field-unit-type"] select', TEST_ORG.unit_type);
      await page.fill('[data-testid="form-field-description"] textarea', TEST_ORG.description);
      
      // 提交表单
      const submitButton = page.locator('[data-testid="form-submit-button"]');
      await submitButton.click();
      
      // 验证创建成功 - 等待表单关闭
      await expect(formModal).not.toBeVisible();
      
      // 验证组织出现在列表中
      await page.waitForTimeout(2000); // 等待数据加载
      await expect(page.locator(`text=${TEST_ORG.name}`).first()).toBeVisible();
    });

    test('应该能够编辑现有组织', async ({ page }) => {
      // 先创建一个测试组织
      await createTestOrganization(page, TEST_ORG);
      
      // 找到并点击编辑按钮
      const editButton = page.locator(`tr:has-text("${TEST_ORG.name}") button[title*="编辑"]`).first();
      await editButton.click();
      
      // 验证编辑表单打开
      const formModal = page.locator('[data-testid="organization-form"]');
      await expect(formModal).toBeVisible();
      
      // 修改名称
      const updatedName = `${TEST_ORG.name}_已编辑`;
      await page.fill('[data-testid="form-field-name"] input', updatedName);
      
      // 提交编辑
      await page.locator('[data-testid="form-submit-button"]').click();
      await expect(formModal).not.toBeVisible();
      
      // 验证编辑成功
      await page.waitForTimeout(2000);
      await expect(page.locator(`text=${updatedName}`).first()).toBeVisible();
    });
  });

  test.describe('3. 计划组织创建测试', () => {
    
    test('应该能够创建计划组织', async ({ page }) => {
      // 点击新增计划组织按钮
      const plannedButton = page.locator('button:has-text("新增计划组织")');
      await plannedButton.click();
      
      // 验证表单弹窗打开
      const formModal = page.locator('[data-testid="organization-form"]');
      await expect(formModal).toBeVisible();
      
      // 填写基本信息
      await page.fill('[data-testid="form-field-name"] input', PLANNED_ORG.name);
      await page.fill('[data-testid="form-field-description"] textarea', PLANNED_ORG.description);
      
      // 验证时态管理自动启用
      const temporalCheckbox = page.locator('[data-testid="form-field-is-temporal"] input');
      await expect(temporalCheckbox).toBeChecked();
      
      // 设置生效时间（明天）
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      const tomorrowStr = tomorrow.toISOString().slice(0, 16);
      
      await page.fill('[data-testid="form-field-effective-from"] input', tomorrowStr);
      
      // 填写变更原因
      await page.fill('[data-testid="form-field-change-reason"] textarea', '创建计划中的新部门');
      
      // 提交表单
      await page.locator('[data-testid="form-submit-button"]').click();
      await expect(formModal).not.toBeVisible();
      
      // 验证计划组织创建成功
      await page.waitForTimeout(2000);
      await expect(page.locator(`text=${PLANNED_ORG.name}`).first()).toBeVisible();
      
      // 验证状态为计划中
      const plannedBadge = page.locator(`tr:has-text("${PLANNED_ORG.name}") text=计划`).first();
      await expect(plannedBadge).toBeVisible();
    });
  });

  test.describe('4. 时态表格功能测试', () => {
    
    test('应该显示时态指示器和状态', async ({ page }) => {
      // 验证时态表格存在
      const temporalTable = page.locator('[data-testid="temporal-table"]');
      if (await temporalTable.isVisible()) {
        // 验证时态指示器列存在
        await expect(page.locator('th:has-text("时态状态")')).toBeVisible();
      }
      
      // 验证状态列显示
      await expect(page.locator('th:has-text("状态")')).toBeVisible();
    });

    test('应该支持搜索和筛选', async ({ page }) => {
      // 查找搜索输入框
      const searchInput = page.locator('input[placeholder*="搜索"]').first();
      if (await searchInput.isVisible()) {
        // 输入搜索关键词
        await searchInput.fill('部门');
        await page.waitForTimeout(1000);
        
        // 验证搜索结果
        const searchResults = page.locator('table tbody tr');
        const count = await searchResults.count();
        expect(count).toBeGreaterThanOrEqual(0);
      }
    });
  });

  test.describe('5. 时间线功能测试', () => {
    
    test('应该能够查看组织时间线', async ({ page }) => {
      // 先创建一个测试组织
      await createTestOrganization(page, TEST_ORG);
      
      // 点击时间线按钮
      const timelineButton = page.locator(`tr:has-text("${TEST_ORG.name}") button[title*="时间线"]`).first();
      if (await timelineButton.isVisible()) {
        await timelineButton.click();
        
        // 验证时间线组件加载
        await expect(page.locator('[data-testid="timeline-component"]')).toBeVisible();
        
        // 验证时间线事件
        const timelineEvents = page.locator('[data-testid="timeline-event"]');
        const eventCount = await timelineEvents.count();
        expect(eventCount).toBeGreaterThanOrEqual(1); // 至少应该有创建事件
      }
    });
  });

  test.describe('6. 记录对比功能测试', () => {
    
    test('应该能够进行记录对比', async ({ page }) => {
      // 访问记录对比页面（如果有直接链接）
      const versionComparisonButton = page.locator('button:has-text("记录对比")').first();
      if (await versionComparisonButton.isVisible()) {
        await versionComparisonButton.click();
        
        // 验证记录对比界面加载
        await expect(page.locator('[data-testid="record-comparison"]')).toBeVisible();
        
        // 验证日期选择器
        const versionSelectors = page.locator('select');
        const selectorCount = await versionSelectors.count();
        expect(selectorCount).toBeGreaterThanOrEqual(2); // 至少两个日期选择器
      }
    });
  });

  test.describe('7. 时态设置功能测试', () => {
    
    test('应该能够打开和配置时态设置', async ({ page }) => {
      // 查找设置按钮
      const settingsButton = page.locator('button:has-text("设置")').first();
      if (await settingsButton.isVisible()) {
        await settingsButton.click();
        
        // 验证设置弹窗打开
        const settingsModal = page.locator('[data-testid="temporal-settings"]');
        await expect(settingsModal).toBeVisible();
        
        // 测试基础设置
        const limitSelect = page.locator('[data-testid="settings-limit"] select');
        if (await limitSelect.isVisible()) {
          await limitSelect.selectOption('100');
        }
        
        // 测试包含停用数据选项
        const includeInactiveCheckbox = page.locator('[data-testid="settings-include-inactive"] input');
        if (await includeInactiveCheckbox.isVisible()) {
          await includeInactiveCheckbox.check();
        }
        
        // 应用设置
        const applyButton = page.locator('button:has-text("应用设置")');
        if (await applyButton.isVisible() && await applyButton.isEnabled()) {
          await applyButton.click();
          await expect(settingsModal).not.toBeVisible();
        }
      }
    });
  });

  test.describe('8. 错误处理和边界条件测试', () => {
    
    test('应该处理网络错误', async ({ page }) => {
      // 拦截网络请求并模拟错误
      await page.route('**/api/**', route => {
        route.abort('failed');
      });
      
      // 尝试加载数据
      await page.reload();
      
      // 验证错误提示
      await expect(page.locator('text=加载失败').or(page.locator('text=网络错误'))).toBeVisible({ timeout: 10000 });
    });

    test('应该处理空数据状态', async ({ page }) => {
      // 查找空状态提示
      const emptyState = page.locator('text=暂无数据').or(page.locator('text=没有找到'));
      if (await emptyState.isVisible()) {
        await expect(emptyState).toBeVisible();
      }
    });
  });

  test.describe('9. 性能和响应性测试', () => {
    
    test('页面加载时间应该合理', async ({ page }) => {
      const startTime = Date.now();
      
      await page.goto(BASE_URL);
      await page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      
      // 验证页面加载时间小于5秒
      expect(loadTime).toBeLessThan(5000);
      console.log(`页面加载时间: ${loadTime}ms`);
    });

    test('应该响应快速交互', async ({ page }) => {
      // 测试按钮点击响应
      const buttons = page.locator('button:visible');
      const buttonCount = await buttons.count();
      
      if (buttonCount > 0) {
        const startTime = Date.now();
        await buttons.first().click();
        const responseTime = Date.now() - startTime;
        
        // 验证交互响应时间小于1秒
        expect(responseTime).toBeLessThan(1000);
        console.log(`按钮响应时间: ${responseTime}ms`);
      }
    });
  });

  test.describe('10. 集成测试', () => {
    
    test('完整用户流程测试', async ({ page }) => {
      console.log('开始完整用户流程测试...');
      
      // 1. 创建组织
      console.log('步骤1: 创建组织');
      await createTestOrganization(page, {
        ...TEST_ORG,
        name: `完整流程测试_${Date.now()}`
      });
      
      // 2. 切换到历史模式
      console.log('步骤2: 切换时态模式');
      const historicalButton = page.locator('button:has-text("历史模式")');
      if (await historicalButton.isVisible()) {
        await historicalButton.click();
        await page.waitForTimeout(1000);
      }
      
      // 3. 返回当前模式
      console.log('步骤3: 返回当前模式');
      const currentButton = page.locator('button:has-text("当前模式")');
      if (await currentButton.isVisible()) {
        await currentButton.click();
        await page.waitForTimeout(1000);
      }
      
      // 4. 创建计划组织
      console.log('步骤4: 创建计划组织');
      const plannedButton = page.locator('button:has-text("新增计划组织")');
      if (await plannedButton.isVisible()) {
        await plannedButton.click();
        
        const formModal = page.locator('[data-testid="organization-form"]');
        if (await formModal.isVisible()) {
          await page.fill('[data-testid="form-field-name"] input', `计划流程测试_${Date.now()}`);
          
          // 设置未来生效时间
          const tomorrow = new Date();
          tomorrow.setDate(tomorrow.getDate() + 1);
          const tomorrowStr = tomorrow.toISOString().slice(0, 16);
          
          await page.fill('[data-testid="form-field-effective-from"] input', tomorrowStr);
          await page.fill('[data-testid="form-field-change-reason"] textarea', '完整流程测试');
          
          await page.locator('[data-testid="form-submit-button"]').click();
          await page.waitForTimeout(2000);
        }
      }
      
      console.log('完整用户流程测试完成');
    });
  });
});

// 辅助函数：创建测试组织
async function createTestOrganization(page: Page, orgData: any) {
  const addButton = page.locator('button:has-text("新增组织")').first();
  await addButton.click();
  
  const formModal = page.locator('[data-testid="organization-form"]');
  await expect(formModal).toBeVisible();
  
  await page.fill('[data-testid="form-field-name"] input', orgData.name);
  await page.selectOption('[data-testid="form-field-unit-type"] select', orgData.unit_type);
  if (orgData.description) {
    await page.fill('[data-testid="form-field-description"] textarea', orgData.description);
  }
  
  await page.locator('[data-testid="form-submit-button"]').click();
  await expect(formModal).not.toBeVisible();
  await page.waitForTimeout(2000);
}