// tests/e2e/complete-pages.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Complete Frontend Pages E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app
    await page.goto('/');
  });

  test.describe('Homepage', () => {
    test('should display homepage with all sections', async ({ page }) => {
      // Check main title
      await expect(page.locator('h1')).toContainText('员工模型管理系统');
      
      // Check feature cards
      await expect(page.locator('[data-testid="feature-cards"]')).toBeVisible();
      
      // Check statistics
      await expect(page.locator('text=测试覆盖率')).toBeVisible();
      await expect(page.locator('text=89%')).toBeVisible();
      
      // Check technology stack section
      await expect(page.locator('text=技术架构')).toBeVisible();
      await expect(page.locator('text=Go + Ent ORM')).toBeVisible();
    });

    test('should navigate to employees page when clicking start button', async ({ page }) => {
      await page.click('text=开始使用');
      await expect(page).toHaveURL('/employees');
    });

    test('should navigate to feature pages when clicking feature cards', async ({ page }) => {
      // Test employee management card
      await page.click('[data-testid="employee-management-card"]');
      await expect(page).toHaveURL('/employees');
      
      await page.goBack();
      
      // Test organization chart card
      await page.click('[data-testid="organization-chart-card"]');
      await expect(page).toHaveURL('/organization/chart');
      
      await page.goBack();
      
      // Test SAM dashboard card
      await page.click('[data-testid="sam-dashboard-card"]');
      await expect(page).toHaveURL('/sam/dashboard');
      
      await page.goBack();
      
      // Test workflow demo card
      await page.click('[data-testid="workflow-demo-card"]');
      await expect(page).toHaveURL('/workflows/demo');
    });
  });

  test.describe('Employee Management Page', () => {
    test('should display employees list and perform CRUD operations', async ({ page }) => {
      await page.goto('/employees');
      
      // Check page title
      await expect(page.locator('h2')).toContainText('员工管理');
      
      // Check table is visible
      await expect(page.locator('.ant-table')).toBeVisible();
      
      // Test search functionality
      await page.fill('[placeholder="搜索员工姓名、ID、邮箱..."]', '张三');
      await page.keyboard.press('Enter');
      
      // Test add employee modal
      await page.click('text=添加员工');
      await expect(page.locator('.ant-modal')).toBeVisible();
      await expect(page.locator('text=添加新员工')).toBeVisible();
      
      // Fill form
      await page.fill('[data-testid="employee-id-input"]', 'EMP999');
      await page.fill('[data-testid="employee-name-input"]', '测试员工');
      await page.fill('[data-testid="employee-email-input"]', 'test@example.com');
      
      // Submit form
      await page.click('text=确定');
      
      // Verify success message
      await expect(page.locator('.ant-notification')).toContainText('员工创建成功');
    });

    test('should filter employees by department and status', async ({ page }) => {
      await page.goto('/employees');
      
      // Test department filter
      await page.click('[data-testid="department-filter"]');
      await page.click('text=研发部');
      
      // Test status filter
      await page.click('[data-testid="status-filter"]');
      await page.click('text=在职');
      
      // Verify filtered results
      await expect(page.locator('.ant-table-tbody tr')).toHaveCount(1, { timeout: 5000 });
    });

    test('should edit and delete employees', async ({ page }) => {
      await page.goto('/employees');
      
      // Click more actions on first employee
      await page.click('.ant-table-tbody tr:first-child [data-testid="employee-actions"]');
      
      // Test edit
      await page.click('text=编辑');
      await expect(page.locator('text=编辑员工信息')).toBeVisible();
      
      await page.fill('[data-testid="employee-name-input"]', '更新姓名');
      await page.click('text=确定');
      
      // Test delete with confirmation
      await page.click('.ant-table-tbody tr:first-child [data-testid="employee-actions"]');
      await page.click('text=删除');
      await expect(page.locator('text=确认删除')).toBeVisible();
      await page.click('text=确定');
    });
  });

  test.describe('Organization Chart Page', () => {
    test('should display organization structure and filter by department', async ({ page }) => {
      await page.goto('/organization/chart');
      
      // Check page title
      await expect(page.locator('h2')).toContainText('组织结构图');
      
      // Check organization overview
      await expect(page.locator('text=组织概览')).toBeVisible();
      await expect(page.locator('[data-testid="total-employees"]')).toBeVisible();
      
      // Test department filter
      await page.click('[data-testid="department-filter"]');
      await page.click('text=研发部');
      
      // Verify filtered results
      await expect(page.locator('[data-testid="department-section"]')).toContainText('研发部');
      
      // Test data refresh
      await page.click('text=刷新数据');
      await expect(page.locator('.ant-btn-loading')).toBeVisible();
    });

    test('should sync data to graph database', async ({ page }) => {
      await page.goto('/organization/chart');
      
      // Click sync button
      await page.click('text=同步到图数据库');
      
      // Wait for sync to complete
      await expect(page.locator('.ant-message')).toContainText('组织数据已同步到图数据库');
    });
  });

  test.describe('SAM Dashboard Page', () => {
    test('should display SAM analysis dashboard', async ({ page }) => {
      await page.goto('/sam/dashboard');
      
      // Check page title
      await expect(page.locator('h2')).toContainText('SAM 智能分析仪表板');
      
      // Check main sections
      await expect(page.locator('text=组织健康度')).toBeVisible();
      await expect(page.locator('text=风险评估')).toBeVisible();
      await expect(page.locator('text=人才分析')).toBeVisible();
      await expect(page.locator('text=战略建议')).toBeVisible();
      
      // Check charts are loaded
      await expect(page.locator('canvas')).toHaveCount(3, { timeout: 10000 });
      
      // Test filter functionality
      await page.click('[data-testid="department-filter"]');
      await page.click('text=研发部');
      
      // Test data refresh
      await page.click('text=刷新数据');
      await expect(page.locator('.ant-btn-loading')).toBeVisible();
    });

    test('should display risk alerts and recommendations', async ({ page }) => {
      await page.goto('/sam/dashboard');
      
      // Wait for data to load
      await page.waitForSelector('[data-testid="alert-level"]');
      
      // Check alert level indicator
      await expect(page.locator('[data-testid="alert-level"]')).toBeVisible();
      
      // Check risk assessment section
      await expect(page.locator('text=关键人员风险')).toBeVisible();
      await expect(page.locator('text=合规风险')).toBeVisible();
      
      // Check recommendations
      await expect(page.locator('[data-testid="recommendations-section"]')).toBeVisible();
    });

    test('should export analysis report', async ({ page }) => {
      await page.goto('/sam/dashboard');
      
      // Wait for data to load
      await page.waitForSelector('text=导出报告');
      
      // Test export functionality
      const downloadPromise = page.waitForEvent('download');
      await page.click('text=导出报告');
      const download = await downloadPromise;
      
      expect(download.suggestedFilename()).toMatch(/sam-analysis-report.*\.pdf/);
    });
  });

  test.describe('Workflow Demo Page', () => {
    test('should display workflow management interface', async ({ page }) => {
      await page.goto('/workflows/demo');
      
      // Check page title
      await expect(page.locator('h2')).toContainText('工作流管理演示');
      
      // Check workflow statistics
      await expect(page.locator('text=运行中')).toBeVisible();
      await expect(page.locator('text=已完成')).toBeVisible();
      await expect(page.locator('text=等待中')).toBeVisible();
      await expect(page.locator('text=失败')).toBeVisible();
      
      // Check workflow cards
      await expect(page.locator('[data-testid="workflow-card"]')).toHaveCount(3);
    });

    test('should create new workflow', async ({ page }) => {
      await page.goto('/workflows/demo');
      
      // Click create workflow button
      await page.click('text=创建新工作流');
      
      // Check modal is open
      await expect(page.locator('text=创建新的职位变更工作流')).toBeVisible();
      
      // Fill form
      await page.fill('[data-testid="employee-id-input"]', 'EMP888');
      await page.fill('[data-testid="employee-name-input"]', '新员工');
      await page.fill('[data-testid="current-position-input"]', '开发工程师');
      await page.fill('[data-testid="new-position-input"]', '高级开发工程师');
      
      // Select department
      await page.click('[data-testid="department-select"]');
      await page.click('text=研发部');
      
      // Submit form
      await page.click('text=创建工作流');
      
      // Verify success message
      await expect(page.locator('.ant-message')).toContainText('工作流创建成功');
    });

    test('should view workflow details', async ({ page }) => {
      await page.goto('/workflows/demo');
      
      // Click details button on first workflow
      await page.click('.ant-card:first-child text=详情');
      
      // Should navigate to workflow details page
      await expect(page).toHaveURL(/\/workflows\/wf-\d+/);
    });

    test('should cancel running workflow', async ({ page }) => {
      await page.goto('/workflows/demo');
      
      // Find running workflow and click cancel
      await page.click('[data-testid="workflow-cancel-button"]');
      
      // Confirm cancellation
      await expect(page.locator('text=确认取消工作流？')).toBeVisible();
      await page.click('text=确定');
      
      // Verify success message
      await expect(page.locator('.ant-message')).toContainText('工作流已取消');
    });
  });

  test.describe('Workflow Details Page', () => {
    test('should display workflow details and approval process', async ({ page }) => {
      await page.goto('/workflows/wf-001');
      
      // Check page title
      await expect(page.locator('h2')).toContainText('工作流详情');
      
      // Check employee information
      await expect(page.locator('text=员工信息')).toBeVisible();
      await expect(page.locator('[data-testid="employee-name"]')).toBeVisible();
      
      // Check position change details
      await expect(page.locator('text=职位变更详情')).toBeVisible();
      
      // Check workflow progress
      await expect(page.locator('[data-testid="workflow-progress"]')).toBeVisible();
      
      // Check approval timeline
      await expect(page.locator('text=审批进度')).toBeVisible();
      
      // Check workflow history
      await expect(page.locator('text=工作流历史')).toBeVisible();
    });

    test('should handle approval actions', async ({ page }) => {
      await page.goto('/workflows/wf-001');
      
      // Check if approval buttons are available
      const approveButton = page.locator('text=批准');
      if (await approveButton.isVisible()) {
        await approveButton.click();
        
        // Confirm approval
        await expect(page.locator('text=审批确认')).toBeVisible();
        await page.click('text=确认');
        
        // Verify success message
        await expect(page.locator('.ant-message')).toContainText('审批成功');
      }
    });
  });

  test.describe('Employee Position History Page', () => {
    test('should display position history and create new position change', async ({ page }) => {
      await page.goto('/employees/positions/emp-001');
      
      // Check page title
      await expect(page.locator('h2')).toContainText('职位历史');
      
      // Check current position section
      await expect(page.locator('text=当前职位')).toBeVisible();
      
      // Check history timeline
      await expect(page.locator('[data-testid="position-timeline"]')).toBeVisible();
      
      // Test create position change
      await page.click('text=添加职位变更');
      
      // Check modal is open
      await expect(page.locator('text=创建职位变更')).toBeVisible();
      
      // Fill form
      await page.fill('[data-testid="position-title-input"]', '技术主管');
      await page.fill('[data-testid="department-input"]', '研发部');
      await page.fill('[data-testid="job-level-input"]', 'MANAGER');
      
      // Submit form
      await page.click('text=提交');
      
      // Verify success message
      await expect(page.locator('.ant-notification')).toContainText('职位变更已提交');
    });

    test('should switch between history and workflow tabs', async ({ page }) => {
      await page.goto('/employees/positions/emp-001');
      
      // Click workflow history tab
      await page.click('text=工作流历史');
      
      // Check workflow content is visible
      await expect(page.locator('[data-testid="workflow-history"]')).toBeVisible();
      
      // Click back to position history tab
      await page.click('text=职位历史');
      
      // Check position history content is visible
      await expect(page.locator('[data-testid="position-timeline"]')).toBeVisible();
    });
  });

  test.describe('Admin Graph Sync Page', () => {
    test('should display graph sync management interface', async ({ page }) => {
      await page.goto('/admin/graph-sync');
      
      // Check page title
      await expect(page.locator('h2')).toContainText('图数据库同步管理');
      
      // Check sync status
      await expect(page.locator('text=同步状态')).toBeVisible();
      
      // Check sync operation buttons
      await expect(page.locator('text=完整同步')).toBeVisible();
      await expect(page.locator('text=部门同步')).toBeVisible();
      await expect(page.locator('text=单个员工同步')).toBeVisible();
      
      // Check statistics section
      await expect(page.locator('text=同步统计')).toBeVisible();
    });

    test('should perform full sync operation', async ({ page }) => {
      await page.goto('/admin/graph-sync');
      
      // Click full sync button
      await page.click('text=完整同步');
      
      // Confirm sync operation
      await expect(page.locator('text=确认完整同步')).toBeVisible();
      await page.click('text=确认');
      
      // Wait for sync to complete
      await expect(page.locator('text=同步中...')).toBeVisible();
      
      // Verify completion
      await expect(page.locator('.ant-notification')).toContainText('完整同步成功', { timeout: 30000 });
    });

    test('should perform department sync', async ({ page }) => {
      await page.goto('/admin/graph-sync');
      
      // Click department sync button
      await page.click('text=部门同步');
      
      // Select department
      await expect(page.locator('text=选择部门')).toBeVisible();
      await page.click('[data-testid="department-select"]');
      await page.click('text=研发部');
      
      // Start sync
      await page.click('text=开始同步');
      
      // Verify success
      await expect(page.locator('.ant-notification')).toContainText('部门同步成功');
    });
  });

  test.describe('Cross-page Navigation', () => {
    test('should navigate between all pages using navigation menu', async ({ page }) => {
      await page.goto('/');
      
      // Test navigation to each page
      const navigationTests = [
        { link: '员工管理', url: '/employees' },
        { link: '组织架构', url: '/organization/chart' },
        { link: 'SAM 智能分析', url: '/sam/dashboard' },
        { link: '工作流管理', url: '/workflows/demo' }
      ];
      
      for (const nav of navigationTests) {
        await page.click(`text=${nav.link}`);
        await expect(page).toHaveURL(nav.url);
        await page.goBack();
      }
    });

    test('should maintain state when navigating between pages', async ({ page }) => {
      // Set a filter on employees page
      await page.goto('/employees');
      await page.fill('[placeholder="搜索员工姓名、ID、邮箱..."]', '张三');
      
      // Navigate away and back
      await page.goto('/organization/chart');
      await page.goto('/employees');
      
      // Check if filter state is maintained (if implemented)
      // This depends on implementation details
    });
  });

  test.describe('Responsive Design', () => {
    test('should work correctly on mobile devices', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      
      await page.goto('/');
      
      // Check that mobile layout is applied
      await expect(page.locator('h1')).toBeVisible();
      
      // Test mobile navigation
      const menuButton = page.locator('[data-testid="mobile-menu-button"]');
      if (await menuButton.isVisible()) {
        await menuButton.click();
        await expect(page.locator('[data-testid="mobile-menu"]')).toBeVisible();
      }
    });

    test('should work correctly on tablet devices', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 });
      
      await page.goto('/employees');
      
      // Check that tablet layout is applied
      await expect(page.locator('.ant-table')).toBeVisible();
      
      // Test that table is responsive
      await expect(page.locator('.ant-table-scroll')).toBeVisible();
    });
  });

  test.describe('Performance and Loading', () => {
    test('should load pages within acceptable time limits', async ({ page }) => {
      const startTime = Date.now();
      await page.goto('/');
      const loadTime = Date.now() - startTime;
      
      // Page should load within 3 seconds
      expect(loadTime).toBeLessThan(3000);
      
      // Check that main content is visible
      await expect(page.locator('h1')).toBeVisible();
    });

    test('should handle slow network conditions', async ({ page }) => {
      // Simulate slow 3G network
      await page.route('**/*', route => {
        setTimeout(() => route.continue(), 100);
      });
      
      await page.goto('/sam/dashboard');
      
      // Should show loading states
      await expect(page.locator('.ant-spin')).toBeVisible();
      
      // Should eventually load content
      await expect(page.locator('text=SAM 智能分析仪表板')).toBeVisible({ timeout: 15000 });
    });
  });

  test.describe('Error Handling', () => {
    test('should handle API errors gracefully', async ({ page }) => {
      // Mock API failure
      await page.route('**/api/**', route => {
        route.fulfill({
          status: 500,
          body: JSON.stringify({ error: 'Internal Server Error' })
        });
      });
      
      await page.goto('/employees');
      
      // Should show error message
      await expect(page.locator('.ant-notification')).toContainText('获取员工数据失败');
    });

    test('should handle network failures', async ({ page }) => {
      // Mock network failure
      await page.route('**/api/**', route => {
        route.abort('failed');
      });
      
      await page.goto('/organization/chart');
      
      // Should show network error message
      await expect(page.locator('.ant-message')).toContainText('网络错误');
    });
  });
});