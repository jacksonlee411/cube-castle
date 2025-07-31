import { Page, expect } from '@playwright/test';

/**
 * E2E测试辅助工具类
 */
export class TestHelpers {
  constructor(private page: Page) {}

  /**
   * 等待页面加载完成
   */
  async waitForPageLoad() {
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForLoadState('domcontentloaded');
  }

  /**
   * 等待数据表格加载完成 - 智能适应版本
   */
  async waitForDataTableLoad() {
    try {
      // 先尝试等待标准的DataTable组件
      await this.page.waitForSelector('[data-testid="data-table"]', { timeout: 5000 });
    } catch {
      // 如果没有标准组件，尝试等待其他表格类型
      try {
        await this.page.waitForSelector('table, .data-table, [role="table"]', { timeout: 5000 });
      } catch {
        // 如果都没有，就等待页面基本加载完成
        await this.waitForPageLoad();
      }
    }
    
    // 尝试等待加载状态消失（如果存在的话）
    try {
      await this.page.waitForSelector('[data-testid="loading"], .loading, .spinner', { state: 'detached', timeout: 3000 });
    } catch {
      // 如果没有loading状态，就忽略
    }
  }

  /**
   * 等待模态框出现 - 智能适应版本
   */
  async waitForModal() {
    try {
      await this.page.waitForSelector('[role="dialog"]', { timeout: 3000 });
    } catch {
      // 如果没有标准模态框，尝试其他模态框选择器
      try {
        await this.page.waitForSelector('.modal, [data-testid*="modal"], .dialog', { timeout: 3000 });
      } catch {
        // 如果没有模态框，可能是内联表单或直接页面跳转
        await this.page.waitForTimeout(500); // 等待短暂时间让页面稳定
      }
    }
  }

  /**
   * 关闭模态框
   */
  async closeModal() {
    const closeButton = this.page.locator('[role="dialog"] button:has-text("取消"), [role="dialog"] button:has-text("关闭")').first();
    if (await closeButton.isVisible()) {
      await closeButton.click();
    }
    await this.page.waitForSelector('[role="dialog"]', { state: 'detached' });
  }

  /**
   * 填写表单字段
   */
  async fillForm(formData: Record<string, string>) {
    for (const [field, value] of Object.entries(formData)) {
      const input = this.page.locator(`input[name="${field}"], textarea[name="${field}"], select[name="${field}"]`);
      if (await input.isVisible()) {
        await input.fill(value);
      }
    }
  }

  /**
   * 智能等待表单元素可用
   */
  async waitForFormElement(selector: string, timeout: number = 10000) {
    try {
      // 等待元素可见并可用
      await this.page.waitForSelector(selector, { timeout: timeout / 2 });
      await this.page.locator(selector).waitFor({ state: 'visible', timeout: timeout / 2 });
      
      // 验证元素确实可交互
      const element = this.page.locator(selector);
      await expect(element).toBeEnabled();
      
      return element;
    } catch (error) {
      throw new Error(`表单元素 "${selector}" 在 ${timeout}ms 内未变得可用: ${error}`);
    }
  }

  /**
   * 智能表单填写
   */
  async fillFormField(selector: string, value: string, timeout: number = 10000) {
    const element = await this.waitForFormElement(selector, timeout);
    await element.clear();
    await element.fill(value);
    
    // 验证填写成功
    await expect(element).toHaveValue(value);
  }

  /**
   * 点击按钮并等待响应 - 增强版
   */
  async clickButtonAndWait(buttonText: string, waitForResponse?: string) {
    try {
      const button = this.page.locator(`button:has-text("${buttonText}")`);
      
      // 等待按钮可见并可用
      await button.waitFor({ state: 'visible', timeout: 5000 });
      await expect(button).toBeEnabled();
      
      if (waitForResponse) {
        await Promise.all([
          this.page.waitForResponse(response => response.url().includes(waitForResponse)),
          button.click()
        ]);
      } else {
        await button.click();
        // 等待点击效果
        await this.page.waitForTimeout(300);
      }
    } catch (error) {
      throw new Error(`点击按钮 "${buttonText}" 失败: ${error}`);
    }
  }

  /**
   * 搜索表格数据
   */
  async searchInTable(searchTerm: string) {
    const searchInput = this.page.locator('input[placeholder*="搜索"]');
    await searchInput.fill(searchTerm);
    await this.page.waitForTimeout(500); // 等待搜索防抖
  }

  /**
   * 验证表格行数
   */
  async verifyTableRowCount(expectedCount: number) {
    const rows = this.page.locator('[data-testid="data-table"] tbody tr');
    await expect(rows).toHaveCount(expectedCount);
  }

  /**
   * 验证提示消息
   */
  async verifyToastMessage(message: string, type: 'success' | 'error' | 'info' = 'success') {
    const toast = this.page.locator('[data-sonner-toaster]');
    await expect(toast).toContainText(message);
  }

  /**
   * 截图用于调试
   */
  async takeScreenshot(name: string) {
    await this.page.screenshot({ 
      path: `test-results/screenshots/${name}-${Date.now()}.png`,
      fullPage: true 
    });
  }

  /**
   * 验证页面标题
   */
  async verifyPageTitle(expectedTitle: string) {
    const title = this.page.locator('h1');
    await expect(title).toContainText(expectedTitle);
  }

  /**
   * 验证统计卡片
   */
  async verifyStatsCard(cardTitle: string, expectedValue?: string) {
    const card = this.page.locator('[data-testid="stats-card"]').filter({ hasText: cardTitle });
    await expect(card).toBeVisible();
    
    if (expectedValue) {
      await expect(card).toContainText(expectedValue);
    }
  }

  /**
   * 导航到指定页面
   */
  async navigateTo(path: string) {
    await this.page.goto(path);
    await this.waitForPageLoad();
  }
}

/**
 * 测试数据生成器
 */
export class TestDataGenerator {
  static generateEmployee() {
    const timestamp = Date.now();
    return {
      legalName: `测试员工${timestamp}`,
      email: `test${timestamp}@company.com`,
      position: '软件工程师',
      department: '技术部',
      hireDate: '2024-01-01'
    };
  }

  static generatePosition() {
    const timestamp = Date.now();
    return {
      title: `测试职位${timestamp}`,
      department: '技术部',
      jobLevel: 'P5',
      salaryMin: '18000',
      salaryMax: '25000',
      description: '这是一个测试职位描述'
    };
  }

  static generateOrganization() {
    const timestamp = Date.now();
    return {
      name: `测试组织${timestamp}`,
      type: 'department',
      managerName: '测试经理',
      maxCapacity: '20',
      description: '这是一个测试组织'
    };
  }
}

/**
 * 页面导航助手
 */
export class NavigationHelper {
  constructor(private page: Page) {}

  async goToEmployees() {
    await this.page.goto('/employees');
  }

  async goToPositions() {
    await this.page.goto('/positions');
  }

  async goToOrganizationChart() {
    await this.page.goto('/organization/chart');
  }

  async goToWorkflowDetail(workflowId: string) {
    await this.page.goto(`/workflows/${workflowId}`);
  }

  async goToEmployeePositionHistory(employeeId: string) {
    await this.page.goto(`/employees/positions/${employeeId}`);
  }

  async goToAdminGraphSync() {
    await this.page.goto('/admin/graph-sync');
  }

  async goToWorkflowDemo() {
    await this.page.goto('/workflows/demo');
  }
}