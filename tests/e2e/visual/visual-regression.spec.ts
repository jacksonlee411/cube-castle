// tests/visual/visual-regression.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Visual Regression Tests', () => {
  // Configure test to use consistent browser settings
  test.use({
    viewport: { width: 1280, height: 720 },
    deviceScaleFactor: 1,
  });

  test.beforeEach(async ({ page }) => {
    // Disable animations and transitions for consistent screenshots
    await page.addStyleTag({
      content: `
        *, *::before, *::after {
          animation-duration: 0s !important;
          animation-delay: 0s !important;
          transition-duration: 0s !important;
          transition-delay: 0s !important;
        }
      `
    });
  });

  test.describe('Homepage Visual Tests', () => {
    test('should match homepage layout', async ({ page }) => {
      await page.goto('/');
      
      // Wait for content to load
      await page.waitForSelector('h1:has-text("员工模型管理系统")');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('homepage-full.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match homepage hero section', async ({ page }) => {
      await page.goto('/');
      
      await page.waitForSelector('h1:has-text("员工模型管理系统")');
      
      // Screenshot of hero section only
      const heroSection = page.locator('[data-testid="hero-section"]').first();
      await expect(heroSection).toHaveScreenshot('homepage-hero.png');
    });

    test('should match feature cards section', async ({ page }) => {
      await page.goto('/');
      
      await page.waitForSelector('text=核心功能模块');
      
      // Screenshot of feature cards
      const featureCards = page.locator('[data-testid="feature-cards"]').first();
      await expect(featureCards).toHaveScreenshot('homepage-features.png');
    });

    test('should match statistics section', async ({ page }) => {
      await page.goto('/');
      
      await page.waitForSelector('text=测试覆盖率');
      
      // Screenshot of statistics section
      const statsSection = page.locator('[data-testid="stats-section"]').first();
      await expect(statsSection).toHaveScreenshot('homepage-stats.png');
    });

    test('should match technology stack section', async ({ page }) => {
      await page.goto('/');
      
      await page.waitForSelector('text=技术架构');
      
      // Screenshot of technology stack
      const techStack = page.locator('[data-testid="tech-stack"]').first();
      await expect(techStack).toHaveScreenshot('homepage-tech-stack.png');
    });
  });

  test.describe('Employee Management Visual Tests', () => {
    test('should match employees list page', async ({ page }) => {
      await page.goto('/employees');
      
      // Wait for table to load
      await page.waitForSelector('.ant-table-tbody tr');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('employees-list.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match employee table header', async ({ page }) => {
      await page.goto('/employees');
      
      await page.waitForSelector('.ant-table-thead');
      
      // Screenshot of table header
      const tableHeader = page.locator('.ant-table-thead');
      await expect(tableHeader).toHaveScreenshot('employees-table-header.png');
    });

    test('should match employee search and filters', async ({ page }) => {
      await page.goto('/employees');
      
      await page.waitForSelector('[data-testid="search-filters"]');
      
      // Screenshot of search and filter section
      const searchFilters = page.locator('[data-testid="search-filters"]');
      await expect(searchFilters).toHaveScreenshot('employees-search-filters.png');
    });

    test('should match add employee modal', async ({ page }) => {
      await page.goto('/employees');
      
      // Open add employee modal
      await page.click('text=添加员工');
      await page.waitForSelector('.ant-modal');
      
      // Screenshot of modal
      const modal = page.locator('.ant-modal');
      await expect(modal).toHaveScreenshot('employees-add-modal.png');
    });

    test('should match employee actions dropdown', async ({ page }) => {
      await page.goto('/employees');
      
      // Wait for table to load
      await page.waitForSelector('.ant-table-tbody tr');
      
      // Click actions dropdown on first employee
      await page.click('.ant-table-tbody tr:first-child [data-testid="employee-actions"]');
      await page.waitForSelector('.ant-dropdown-menu');
      
      // Screenshot of dropdown menu
      const dropdown = page.locator('.ant-dropdown-menu');
      await expect(dropdown).toHaveScreenshot('employees-actions-dropdown.png');
    });
  });

  test.describe('Organization Chart Visual Tests', () => {
    test('should match organization chart page', async ({ page }) => {
      await page.goto('/organization/chart');
      
      // Wait for organization data to load
      await page.waitForSelector('text=组织概览');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('organization-chart.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match organization overview cards', async ({ page }) => {
      await page.goto('/organization/chart');
      
      await page.waitForSelector('[data-testid="organization-overview"]');
      
      // Screenshot of overview section
      const overview = page.locator('[data-testid="organization-overview"]');
      await expect(overview).toHaveScreenshot('organization-overview.png');
    });

    test('should match department filter section', async ({ page }) => {
      await page.goto('/organization/chart');
      
      await page.waitForSelector('[data-testid="department-filters"]');
      
      // Screenshot of filter section
      const filters = page.locator('[data-testid="department-filters"]');
      await expect(filters).toHaveScreenshot('organization-filters.png');
    });

    test('should match department cards layout', async ({ page }) => {
      await page.goto('/organization/chart');
      
      await page.waitForSelector('[data-testid="department-cards"]');
      
      // Screenshot of department cards
      const departmentCards = page.locator('[data-testid="department-cards"]');
      await expect(departmentCards).toHaveScreenshot('organization-departments.png');
    });
  });

  test.describe('SAM Dashboard Visual Tests', () => {
    test('should match SAM dashboard layout', async ({ page }) => {
      await page.goto('/sam/dashboard');
      
      // Wait for dashboard to load
      await page.waitForSelector('text=SAM 智能分析仪表板');
      await page.waitForTimeout(2000); // Wait for charts to render
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('sam-dashboard.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match alert level indicator', async ({ page }) => {
      await page.goto('/sam/dashboard');
      
      await page.waitForSelector('[data-testid="alert-level"]');
      
      // Screenshot of alert level section
      const alertLevel = page.locator('[data-testid="alert-level"]');
      await expect(alertLevel).toHaveScreenshot('sam-alert-level.png');
    });

    test('should match organization health metrics', async ({ page }) => {
      await page.goto('/sam/dashboard');
      
      await page.waitForSelector('[data-testid="health-metrics"]');
      
      // Screenshot of health metrics
      const healthMetrics = page.locator('[data-testid="health-metrics"]');
      await expect(healthMetrics).toHaveScreenshot('sam-health-metrics.png');
    });

    test('should match risk assessment section', async ({ page }) => {
      await page.goto('/sam/dashboard');
      
      await page.waitForSelector('[data-testid="risk-assessment"]');
      
      // Screenshot of risk assessment
      const riskAssessment = page.locator('[data-testid="risk-assessment"]');
      await expect(riskAssessment).toHaveScreenshot('sam-risk-assessment.png');
    });

    test('should match recommendations panel', async ({ page }) => {
      await page.goto('/sam/dashboard');
      
      await page.waitForSelector('[data-testid="recommendations"]');
      
      // Screenshot of recommendations
      const recommendations = page.locator('[data-testid="recommendations"]');
      await expect(recommendations).toHaveScreenshot('sam-recommendations.png');
    });

    test('should match chart components', async ({ page }) => {
      await page.goto('/sam/dashboard');
      
      // Wait for charts to load
      await page.waitForSelector('canvas');
      await page.waitForTimeout(1000);
      
      // Screenshot each chart type
      const charts = page.locator('[data-testid="chart-container"]');
      const chartCount = await charts.count();
      
      for (let i = 0; i < chartCount; i++) {
        const chart = charts.nth(i);
        await expect(chart).toHaveScreenshot(`sam-chart-${i}.png`);
      }
    });
  });

  test.describe('Workflow Demo Visual Tests', () => {
    test('should match workflow demo page', async ({ page }) => {
      await page.goto('/workflows/demo');
      
      // Wait for workflow data to load
      await page.waitForSelector('text=工作流管理演示');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('workflows-demo.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match workflow statistics cards', async ({ page }) => {
      await page.goto('/workflows/demo');
      
      await page.waitForSelector('[data-testid="workflow-stats"]');
      
      // Screenshot of statistics
      const stats = page.locator('[data-testid="workflow-stats"]');
      await expect(stats).toHaveScreenshot('workflows-stats.png');
    });

    test('should match workflow cards layout', async ({ page }) => {
      await page.goto('/workflows/demo');
      
      await page.waitForSelector('[data-testid="workflow-cards"]');
      
      // Screenshot of workflow cards
      const workflowCards = page.locator('[data-testid="workflow-cards"]');
      await expect(workflowCards).toHaveScreenshot('workflows-cards.png');
    });

    test('should match create workflow modal', async ({ page }) => {
      await page.goto('/workflows/demo');
      
      // Open create workflow modal
      await page.click('text=创建新工作流');
      await page.waitForSelector('.ant-modal');
      
      // Screenshot of modal
      const modal = page.locator('.ant-modal');
      await expect(modal).toHaveScreenshot('workflows-create-modal.png');
    });

    test('should match workflow progress indicators', async ({ page }) => {
      await page.goto('/workflows/demo');
      
      await page.waitForSelector('[data-testid="workflow-progress"]');
      
      // Screenshot of progress section
      const progress = page.locator('[data-testid="workflow-progress"]').first();
      await expect(progress).toHaveScreenshot('workflows-progress.png');
    });
  });

  test.describe('Workflow Details Visual Tests', () => {
    test('should match workflow details page', async ({ page }) => {
      await page.goto('/workflows/wf-001');
      
      // Wait for workflow details to load
      await page.waitForSelector('text=工作流详情');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('workflow-details.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match employee information section', async ({ page }) => {
      await page.goto('/workflows/wf-001');
      
      await page.waitForSelector('[data-testid="employee-info"]');
      
      // Screenshot of employee info
      const employeeInfo = page.locator('[data-testid="employee-info"]');
      await expect(employeeInfo).toHaveScreenshot('workflow-employee-info.png');
    });

    test('should match workflow progress steps', async ({ page }) => {
      await page.goto('/workflows/wf-001');
      
      await page.waitForSelector('[data-testid="workflow-steps"]');
      
      // Screenshot of progress steps
      const steps = page.locator('[data-testid="workflow-steps"]');
      await expect(steps).toHaveScreenshot('workflow-steps.png');
    });

    test('should match approval timeline', async ({ page }) => {
      await page.goto('/workflows/wf-001');
      
      await page.waitForSelector('[data-testid="approval-timeline"]');
      
      // Screenshot of approval timeline
      const timeline = page.locator('[data-testid="approval-timeline"]');
      await expect(timeline).toHaveScreenshot('workflow-timeline.png');
    });
  });

  test.describe('Employee Position History Visual Tests', () => {
    test('should match position history page', async ({ page }) => {
      await page.goto('/employees/positions/emp-001');
      
      // Wait for position data to load
      await page.waitForSelector('text=职位历史');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('position-history.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match current position card', async ({ page }) => {
      await page.goto('/employees/positions/emp-001');
      
      await page.waitForSelector('[data-testid="current-position"]');
      
      // Screenshot of current position
      const currentPosition = page.locator('[data-testid="current-position"]');
      await expect(currentPosition).toHaveScreenshot('position-current.png');
    });

    test('should match position timeline', async ({ page }) => {
      await page.goto('/employees/positions/emp-001');
      
      await page.waitForSelector('[data-testid="position-timeline"]');
      
      // Screenshot of timeline
      const timeline = page.locator('[data-testid="position-timeline"]');
      await expect(timeline).toHaveScreenshot('position-timeline.png');
    });

    test('should match create position change modal', async ({ page }) => {
      await page.goto('/employees/positions/emp-001');
      
      // Open create modal
      await page.click('text=添加职位变更');
      await page.waitForSelector('.ant-modal');
      
      // Screenshot of modal
      const modal = page.locator('.ant-modal');
      await expect(modal).toHaveScreenshot('position-create-modal.png');
    });
  });

  test.describe('Admin Graph Sync Visual Tests', () => {
    test('should match graph sync admin page', async ({ page }) => {
      await page.goto('/admin/graph-sync');
      
      // Wait for sync interface to load
      await page.waitForSelector('text=图数据库同步管理');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('admin-graph-sync.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match sync status section', async ({ page }) => {
      await page.goto('/admin/graph-sync');
      
      await page.waitForSelector('[data-testid="sync-status"]');
      
      // Screenshot of sync status
      const syncStatus = page.locator('[data-testid="sync-status"]');
      await expect(syncStatus).toHaveScreenshot('admin-sync-status.png');
    });

    test('should match sync operations section', async ({ page }) => {
      await page.goto('/admin/graph-sync');
      
      await page.waitForSelector('[data-testid="sync-operations"]');
      
      // Screenshot of operations
      const operations = page.locator('[data-testid="sync-operations"]');
      await expect(operations).toHaveScreenshot('admin-sync-operations.png');
    });

    test('should match sync statistics', async ({ page }) => {
      await page.goto('/admin/graph-sync');
      
      await page.waitForSelector('[data-testid="sync-statistics"]');
      
      // Screenshot of statistics
      const statistics = page.locator('[data-testid="sync-statistics"]');
      await expect(statistics).toHaveScreenshot('admin-sync-statistics.png');
    });
  });

  test.describe('Responsive Design Visual Tests', () => {
    test('should match mobile layout - homepage', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/');
      
      await page.waitForSelector('h1:has-text("员工模型管理系统")');
      
      await expect(page).toHaveScreenshot('mobile-homepage.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match tablet layout - employees', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 });
      await page.goto('/employees');
      
      await page.waitForSelector('.ant-table-tbody tr');
      
      await expect(page).toHaveScreenshot('tablet-employees.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match mobile layout - SAM dashboard', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/sam/dashboard');
      
      await page.waitForSelector('text=SAM 智能分析仪表板');
      await page.waitForTimeout(2000); // Wait for charts
      
      await expect(page).toHaveScreenshot('mobile-sam-dashboard.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });
  });

  test.describe('Theme and Color Scheme Tests', () => {
    test('should match dark theme if available', async ({ page }) => {
      await page.goto('/');
      
      // If dark theme toggle exists, test it
      const themeToggle = page.locator('[data-testid="theme-toggle"]');
      if (await themeToggle.isVisible()) {
        await themeToggle.click();
        await page.waitForTimeout(500);
        
        await expect(page).toHaveScreenshot('homepage-dark-theme.png', {
          fullPage: true,
          animations: 'disabled'
        });
      }
    });

    test('should match high contrast mode if available', async ({ page }) => {
      await page.goto('/');
      
      // Apply high contrast styles
      await page.addStyleTag({
        content: `
          body { filter: contrast(150%) brightness(1.2); }
        `
      });
      
      await page.waitForSelector('h1:has-text("员工模型管理系统")');
      
      await expect(page).toHaveScreenshot('homepage-high-contrast.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });
  });

  test.describe('Print Layout Tests', () => {
    test('should match print layout - SAM report', async ({ page }) => {
      await page.goto('/sam/dashboard');
      
      // Wait for content to load
      await page.waitForSelector('text=SAM 智能分析仪表板');
      await page.waitForTimeout(2000);
      
      // Emulate print media
      await page.emulateMedia({ media: 'print' });
      
      await expect(page).toHaveScreenshot('sam-dashboard-print.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match print layout - employee list', async ({ page }) => {
      await page.goto('/employees');
      
      await page.waitForSelector('.ant-table-tbody tr');
      
      // Emulate print media
      await page.emulateMedia({ media: 'print' });
      
      await expect(page).toHaveScreenshot('employees-print.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });
  });

  test.describe('Error State Visual Tests', () => {
    test('should match 404 error page', async ({ page }) => {
      await page.goto('/non-existent-page');
      
      // Wait for 404 page to load
      await page.waitForSelector('text=404');
      
      await expect(page).toHaveScreenshot('error-404.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });

    test('should match error states in components', async ({ page }) => {
      // Mock API to return errors
      await page.route('**/api/**', route => {
        route.fulfill({
          status: 500,
          body: JSON.stringify({ error: 'Internal Server Error' })
        });
      });
      
      await page.goto('/employees');
      
      // Wait for error state to appear
      await page.waitForSelector('.ant-notification');
      
      await expect(page).toHaveScreenshot('employees-error-state.png', {
        fullPage: true,
        animations: 'disabled'
      });
    });
  });

  test.describe('Loading State Visual Tests', () => {
    test('should match loading states', async ({ page }) => {
      // Delay API responses
      await page.route('**/api/**', async route => {
        await new Promise(resolve => setTimeout(resolve, 1000));
        await route.continue();
      });
      
      const loadingPromise = page.goto('/sam/dashboard');
      
      // Screenshot loading state
      await page.waitForSelector('.ant-spin');
      await expect(page.locator('.ant-spin').first()).toHaveScreenshot('loading-spinner.png');
      
      await loadingPromise;
    });
  });
});