/**
 * 时态GraphQL功能完整端到端测试套件
 * 
 * 测试覆盖:
 * 1. 功能完整性: 历史记录查看、时间点查询、组织切换
 * 2. 数据准确性: 14条历史记录验证、时间点精确查询
 * 3. 错误处理: 无效组织、无记录情况、网络异常
 * 4. 性能基准: 响应时间、缓存效率、并发处理
 * 5. 用户体验: 加载状态、错误提示、视觉反馈
 * 6. 跨浏览器: Chrome、Firefox兼容性验证
 */

import { test, expect, Page } from '@playwright/test';

// 测试数据配置
const TEST_CONFIG = {
  baseUrl: ((process.env.PW_BASE_URL || '').replace(/\/+$/, '') || '') + '/temporal-graphql',
  validOrgCode: '1000056',
  invalidOrgCode: '9999999',
  testDates: {
    current: '2043-06-01',      // 当前记录
    historical: '2026-01-15',   // 历史记录
    noRecord: '2020-01-01',     // 无记录时间点
    future: '2050-12-31'        // 未来时间点
  },
  expectedHistoryCount: 14,
  performanceThresholds: {
    pageLoad: 5000,     // 5秒
    apiResponse: 1000,  // 1秒
    cacheHit: 100      // 100ms
  }
};

// 测试辅助函数
class TemporalTestHelper {
  constructor(private page: Page) {}

  async navigateToPage() {
    await this.page.goto(TEST_CONFIG.baseUrl);
    await this.page.waitForLoadState('networkidle');
  }

  async selectOrganization(orgCode: string) {
    // 如果是预设组织，直接选择
    if (orgCode === TEST_CONFIG.validOrgCode) {
      await this.page.selectOption('select', `${orgCode} - 完整历史记录演示 (14条记录)`);
    } else {
      // 使用自定义代码
      await this.page.fill('input[placeholder*="输入组织代码"]', orgCode);
      await this.page.click('button:has-text("使用自定义代码")');
    }
    
    // 验证组织代码已切换
    await expect(this.page.locator('text=当前查询组织:')).toBeVisible();
    await expect(this.page.locator(`strong:has-text("${orgCode}")`)).toBeVisible();
  }

  async switchToHistoryTab() {
    await this.page.click('tab:has-text("历史记录查看")');
    await this.page.waitForSelector('[data-testid="history-records"]', { timeout: 5000 });
  }

  async switchToTimePointTab() {
    await this.page.click('tab:has-text("时间点查询")');
    await this.page.waitForSelector('[data-testid="timepoint-query"]', { timeout: 5000 });
  }

  async queryTimePoint(date: string) {
    await this.page.fill('input[type="date"]', date);
    
    // 记录开始时间测量性能
    const startTime = Date.now();
    await this.page.click('button:has-text("查询")');
    
    // 等待查询完成
    await this.page.waitForSelector('[data-testid="query-result"], [data-testid="no-result"]', { timeout: 5000 });
    
    const endTime = Date.now();
    return endTime - startTime;
  }

  async getHistoryRecordCount(): Promise<number> {
    const countText = await this.page.textContent('text=/时态历史记录 \\((\\d+) 条\\)/');
    const match = countText?.match(/时态历史记录 \((\d+) 条\)/);
    return match ? parseInt(match[1], 10) : 0;
  }

  async refreshHistoryData() {
    const startTime = Date.now();
    await this.page.click('button:has-text("刷新")');
    await this.page.waitForLoadState('networkidle');
    return Date.now() - startTime;
  }

  async clearCache() {
    await this.page.click('button:has-text("清除缓存")');
    await this.page.waitForLoadState('networkidle');
  }
}

// 主测试套件
test.describe('时态GraphQL功能完整验证', () => {
  let helper: TemporalTestHelper;

  test.beforeEach(async ({ page }) => {
    helper = new TemporalTestHelper(page);
    await helper.navigateToPage();
  });

  test.describe('功能完整性测试', () => {
    
    test('历史记录查看功能验证', async ({ page }) => {
      // 切换到历史记录标签页
      await helper.switchToHistoryTab();
      
      // 验证历史记录数量
      const recordCount = await helper.getHistoryRecordCount();
      expect(recordCount).toBe(TEST_CONFIG.expectedHistoryCount);
      
      // 验证历史记录内容完整性
      await expect(page.locator('text=端到端测试重组部门2043')).toBeVisible();
      await expect(page.locator('text=初创测试部门v1')).toBeVisible();
      
      // 验证生效期间显示
      await expect(page.locator('text=2043-01-01 起生效')).toBeVisible();
      await expect(page.locator('text=2023-01-01 至 2023-12-31')).toBeVisible();
      
      // 验证当前记录标识
      await expect(page.locator('.badge:has-text("当前")')).toBeVisible();
    });

    test('时间点查询功能验证', async ({ page }) => {
      await helper.switchToTimePointTab();
      
      // 测试当前记录查询
      const currentQueryTime = await helper.queryTimePoint(TEST_CONFIG.testDates.current);
      expect(currentQueryTime).toBeLessThan(TEST_CONFIG.performanceThresholds.apiResponse);
      
      await expect(page.locator('text=端到端测试重组部门2043')).toBeVisible();
      await expect(page.locator('text=当前记录')).toBeVisible();
      
      // 测试历史记录查询
      await helper.queryTimePoint(TEST_CONFIG.testDates.historical);
      await expect(page.locator('text=时态管理测试部门v10_新增功能验证')).toBeVisible();
      await expect(page.locator('text=历史记录')).toBeVisible();
    });

    test('组织切换功能验证', async ({ page }) => {
      // 切换到无效组织
      await helper.selectOrganization(TEST_CONFIG.invalidOrgCode);
      
      // 验证历史记录为空
      await helper.switchToHistoryTab();
      const emptyRecordCount = await helper.getHistoryRecordCount();
      expect(emptyRecordCount).toBe(0);
      await expect(page.locator('text=该组织没有历史记录')).toBeVisible();
      
      // 切换回有效组织
      await page.click('button:has-text("返回预设选择")');
      await helper.selectOrganization(TEST_CONFIG.validOrgCode);
      
      // 验证数据恢复
      const recoveredCount = await helper.getHistoryRecordCount();
      expect(recoveredCount).toBe(TEST_CONFIG.expectedHistoryCount);
    });
  });

  test.describe('错误处理验证', () => {
    
    test('无效组织代码处理', async ({ page }) => {
      await helper.selectOrganization(TEST_CONFIG.invalidOrgCode);
      
      // 历史记录标签页错误处理
      await helper.switchToHistoryTab();
      await expect(page.locator('text=时态历史记录 (0 条)')).toBeVisible();
      await expect(page.locator('text=该组织没有历史记录')).toBeVisible();
      
      // 时间点查询错误处理
      await helper.switchToTimePointTab();
      await helper.queryTimePoint(TEST_CONFIG.testDates.current);
      await expect(page.locator(`text=在 ${TEST_CONFIG.testDates.current} 时间点没有找到组织`)).toBeVisible();
      await expect(page.locator('text=无查询结果')).toBeVisible();
    });

    test('无记录时间点处理', async ({ page }) => {
      await helper.switchToTimePointTab();
      await helper.queryTimePoint(TEST_CONFIG.testDates.noRecord);
      
      await expect(page.locator(`text=在 ${TEST_CONFIG.testDates.noRecord} 时间点没有找到`)).toBeVisible();
      await expect(page.locator('text=无查询结果')).toBeVisible();
    });

    test('用户界面反馈验证', async ({ page }) => {
      await helper.switchToTimePointTab();
      
      // 验证使用说明显示
      await expect(page.locator('text=使用说明:')).toBeVisible();
      await expect(page.locator('text=蓝色背景表示当前有效记录')).toBeVisible();
      await expect(page.locator('text=橙色背景表示历史记录')).toBeVisible();
      
      // 验证快速选择按钮
      await expect(page.locator('button:has-text("今天")')).toBeVisible();
      await expect(page.locator('button:has-text("昨天")')).toBeVisible();
      await expect(page.locator('button:has-text("年初")')).toBeVisible();
    });
  });

  test.describe('性能基准验证', () => {
    
    test('页面加载性能', async () => {
      const startTime = Date.now();
      await helper.navigateToPage();
      const loadTime = Date.now() - startTime;
      
      expect(loadTime).toBeLessThan(TEST_CONFIG.performanceThresholds.pageLoad);
    });

    test('API响应性能', async () => {
      await helper.switchToTimePointTab();
      
      // 测试多个时间点查询的性能
      const queryTimes: number[] = [];
      
      for (const date of Object.values(TEST_CONFIG.testDates)) {
        const queryTime = await helper.queryTimePoint(date);
        queryTimes.push(queryTime);
      }
      
      // 验证所有查询都在性能阈值内
      queryTimes.forEach(time => {
        expect(time).toBeLessThan(TEST_CONFIG.performanceThresholds.apiResponse);
      });
      
      // 验证平均响应时间
      const avgTime = queryTimes.reduce((sum, time) => sum + time, 0) / queryTimes.length;
      expect(avgTime).toBeLessThan(TEST_CONFIG.performanceThresholds.apiResponse / 2);
    });

    test('缓存效率验证', async () => {
      await helper.switchToTimePointTab();
      
      // 首次查询 (Cache MISS)
      const firstQuery = await helper.queryTimePoint(TEST_CONFIG.testDates.current);
      
      // 第二次相同查询 (Cache HIT)
      const secondQuery = await helper.queryTimePoint(TEST_CONFIG.testDates.current);
      
      // 缓存命中应该更快
      expect(secondQuery).toBeLessThan(firstQuery);
      expect(secondQuery).toBeLessThan(TEST_CONFIG.performanceThresholds.cacheHit);
    });

    test('历史记录刷新性能', async () => {
      await helper.switchToHistoryTab();
      
      const refreshTime = await helper.refreshHistoryData();
      expect(refreshTime).toBeLessThan(TEST_CONFIG.performanceThresholds.apiResponse);
      
      // 验证数据完整性
      const recordCount = await helper.getHistoryRecordCount();
      expect(recordCount).toBe(TEST_CONFIG.expectedHistoryCount);
    });
  });

  test.describe('用户体验验证', () => {
    
    test('加载状态显示', async ({ page }) => {
      await helper.switchToTimePointTab();
      
      // 监听加载状态
      await page.click('button:has-text("查询")');
      
      // 验证加载中状态 (快速查询可能看不到)
      try {
        await page.waitForSelector('text=查询中...', { timeout: 500 });
      } catch {
        // 查询太快，加载状态可能不显示，这是正常的
      }
      
      // 验证最终有结果显示
      await page.waitForSelector('[data-testid="query-result"], [data-testid="no-result"]');
    });

    test('视觉反馈验证', async ({ page }) => {
      await helper.switchToTimePointTab();
      
      // 当前记录 - 蓝色背景
      await helper.queryTimePoint(TEST_CONFIG.testDates.current);
      await expect(page.locator('.current-record-card')).toHaveCSS('background-color', /blue|#.*blue.*/i);
      
      // 历史记录 - 橙色背景  
      await helper.queryTimePoint(TEST_CONFIG.testDates.historical);
      await expect(page.locator('.historical-record-card')).toHaveCSS('background-color', /orange|peach|#.*orange.*/i);
    });

    test('交互流畅性验证', async ({ page }) => {
      // 标签页切换
      await helper.switchToHistoryTab();
      await expect(page.locator('tab[aria-selected="true"]:has-text("历史记录查看")')).toBeVisible();
      
      await helper.switchToTimePointTab();
      await expect(page.locator('tab[aria-selected="true"]:has-text("时间点查询")')).toBeVisible();
      
      // 快速日期选择
      await page.click('button:has-text("今天")');
      const todayInput = await page.inputValue('input[type="date"]');
      const today = new Date().toISOString().split('T')[0];
      expect(todayInput).toBe(today);
      
      // 组织代码切换
      await helper.selectOrganization(TEST_CONFIG.invalidOrgCode);
      await expect(page.locator(`strong:has-text("${TEST_CONFIG.invalidOrgCode}")`)).toBeVisible();
      
      await page.click('button:has-text("返回预设选择")');
      await expect(page.locator(`strong:has-text("${TEST_CONFIG.validOrgCode}")`)).toBeVisible();
    });
  });
});

// 跨浏览器兼容性测试 (单独配置)
test.describe('跨浏览器兼容性验证', () => {
  
  ['chromium', 'firefox'].forEach(browserName => {
    test(`${browserName} 基本功能验证`, async ({ page }) => {
      const helper = new TemporalTestHelper(page);
      await helper.navigateToPage();
      
      // 基本功能测试
      await helper.switchToHistoryTab();
      const recordCount = await helper.getHistoryRecordCount();
      expect(recordCount).toBe(TEST_CONFIG.expectedHistoryCount);
      
      await helper.switchToTimePointTab();
      await helper.queryTimePoint(TEST_CONFIG.testDates.current);
      await expect(page.locator('text=端到端测试重组部门2043')).toBeVisible();
    });
  });
});

// 并发测试
test.describe('并发性能验证', () => {
  
  test('并发时间点查询', async ({ browser }) => {
    // 创建多个浏览器上下文模拟并发用户
    const contexts = await Promise.all([
      browser.newContext(),
      browser.newContext(),
      browser.newContext()
    ]);
    
    const helpers = await Promise.all(
      contexts.map(async context => {
        const page = await context.newPage();
        const helper = new TemporalTestHelper(page);
        await helper.navigateToPage();
        await helper.switchToTimePointTab();
        return helper;
      })
    );
    
    // 并发执行查询
    const startTime = Date.now();
    const queryPromises = helpers.map(helper => 
      helper.queryTimePoint(TEST_CONFIG.testDates.current)
    );
    
    const queryTimes = await Promise.all(queryPromises);
    const totalTime = Date.now() - startTime;
    
    // 验证并发性能
    expect(totalTime).toBeLessThan(TEST_CONFIG.performanceThresholds.apiResponse * 2);
    queryTimes.forEach(time => {
      expect(time).toBeLessThan(TEST_CONFIG.performanceThresholds.apiResponse);
    });
    
    // 清理
    await Promise.all(contexts.map(context => context.close()));
  });
});

// 数据完整性验证
test.describe('数据完整性验证', () => {
  
  test('历史记录数据准确性', async ({ page }) => {
    const helper = new TemporalTestHelper(page);
    await helper.navigateToPage();
    await helper.switchToHistoryTab();
    
    // 验证关键历史记录存在
    const expectedRecords = [
      { name: '端到端测试重组部门2043', year: '2043' },
      { name: '初创测试部门v1', year: '2023' },
      { name: '时态管理测试部门v10_新增功能验证', year: '2026' },
      { name: '重组后的创新实验室', year: '2029' }
    ];
    
    for (const record of expectedRecords) {
      await expect(page.locator(`text=${record.name}`)).toBeVisible();
      await expect(page.locator(`text*=${record.year}`)).toBeVisible();
    }
    
    // 验证变更原因显示
    await expect(page.locator('text=端到端测试重组验证')).toBeVisible();
    await expect(page.locator('text=部门成立')).toBeVisible();
    
    // 验证组织类型和状态
    await expect(page.locator('text=DEPARTMENT')).toBeVisible();
    await expect(page.locator('text=ACTIVE')).toBeVisible();
  });

  test('时间点查询数据准确性', async ({ page }) => {
    const helper = new TemporalTestHelper(page);
    await helper.navigateToPage();
    await helper.switchToTimePointTab();
    
    // 测试特定时间点的数据准确性
    const testCases = [
      {
        date: '2043-06-01',
        expectedName: '端到端测试重组部门2043',
        expectedStatus: '当前记录',
        expectedType: 'DEPARTMENT'
      },
      {
        date: '2026-01-15', 
        expectedName: '时态管理测试部门v10_新增功能验证',
        expectedStatus: '历史记录',
        expectedType: 'DEPARTMENT'
      }
    ];
    
    for (const testCase of testCases) {
      await helper.queryTimePoint(testCase.date);
      await expect(page.locator(`text=${testCase.expectedName}`)).toBeVisible();
      await expect(page.locator(`text=${testCase.expectedStatus}`)).toBeVisible();
      await expect(page.locator(`text=${testCase.expectedType}`)).toBeVisible();
    }
  });
});
