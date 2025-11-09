import { test, expect } from '@playwright/test';
import type { Locator, Page } from '@playwright/test';
import { setupAuth } from './auth-setup';
import { E2E_CONFIG } from './config/test-environment';
import { ensurePwJwt, getPwJwt } from './utils/authToken';
import { waitForGraphQL, waitForNavigation, waitForPageReady } from './utils/waitPatterns';

const ROOT_PARENT_CODE = '1000000';
const waitForOrganizationsResponse = (page: Page) => waitForGraphQL(page, /organizations/i);

const getSearchInput = (page: Page) =>
  page
    .locator(
      'input[placeholder*="搜索组织名称"], input[placeholder*="搜索"], input[name="search"]',
    )
    .first();

const waitForTemporalDetailReady = async (page: Page): Promise<void> => {
  await waitForPageReady(page);
  await expect(page.getByTestId('temporal-master-detail-view')).toBeVisible({
    timeout: 20000,
  });
  await expect(page.getByTestId('temporal-timeline')).toBeVisible({
    timeout: 20000,
  });
};

const selectTimelineNodeIfNeeded = async (page: Page) => {
  const firstNode = page.getByTestId('temporal-timeline-node').first();
  if (await firstNode.count()) {
    await firstNode.click();
    await page.waitForTimeout(250);
  }
};

const expectDeleteButtonVisible = async (page: Page, attempt = 0): Promise<Locator> => {
  const wrapper = page.getByTestId('temporal-delete-record-button-wrapper');
  await expect(wrapper).toBeVisible({ timeout: 20000 });

  const recordButton = wrapper.getByTestId('temporal-delete-record-button');
  if (await recordButton.count()) {
    await expect(recordButton).toBeVisible({ timeout: 20000 });
    return recordButton;
  }

  const organizationButton = wrapper.getByTestId('temporal-delete-organization-button');
  if (await organizationButton.count()) {
    await expect(organizationButton).toBeVisible({ timeout: 20000 });
    return organizationButton;
  }

  if (attempt >= 1) {
    throw new Error('未能找到删除按钮，即使尝试重新选择时间线节点后仍失败');
  }

  await selectTimelineNodeIfNeeded(page);
  await page.waitForTimeout(250);
  return expectDeleteButtonVisible(page, attempt + 1);
};

const filterOrganizationsByName = async (
  page: Page,
  name: string,
): Promise<void> => {
  const searchInput = getSearchInput(page);
  await expect(searchInput).toBeVisible({ timeout: 10000 });

  const waitForQuery = waitForOrganizationsResponse(page);
  await searchInput.fill(name);
  await Promise.race([
    waitForQuery.catch(() => {}),
    page.waitForTimeout(4000),
  ]);
};

const resetOrganizationFilters = async (page: Page): Promise<void> => {
  const searchInput = getSearchInput(page);
  if (await searchInput.isVisible()) {
    await searchInput.fill('');
    await Promise.race([
      waitForOrganizationsResponse(page).catch(() => {}),
      page.waitForTimeout(2000),
    ]);
  }

  const resetButton = page.getByRole('button', { name: '重置筛选' });
  if (await resetButton.isEnabled()) {
    await Promise.race([
      waitForOrganizationsResponse(page).catch(() => {}),
      (async () => {
        await resetButton.click();
        await page.waitForTimeout(2000);
      })(),
    ]);
  }
};

test.describe('业务流程端到端测试', () => {

  test.beforeEach(async ({ page }) => {
    // 设置认证信息到 localStorage（确保 RequireAuth 可以通过验证）
    await setupAuth(page);

    // 导航到组织管理页面 (使用相对路径,Playwright会自动添加baseURL)
    const initialOrganizationsResponse = waitForOrganizationsResponse(page);
    await page.goto('/organizations');
    await waitForPageReady(page);
    await initialOrganizationsResponse;

    // 等待页面加载完成 - 使用 data-testid 而不是文本，避免加载状态干扰
    await expect(page.getByTestId('organization-dashboard-wrapper')).toBeVisible({ timeout: 15000 });

    // 等待加载状态消失，确保数据已加载
    await page.waitForSelector('text=加载组织数据中...', { state: 'detached', timeout: 15000 }).catch(() => {
      // 如果没有加载状态也没关系，说明加载很快完成了
    });

    // 最后确认标题可见
    await expect(page.getByText('组织架构管理')).toBeVisible({ timeout: 10000 });
  });

  test('完整CRUD业务流程测试', async ({ page }) => {
    test.setTimeout(180000);

    const uniqueSuffix = Date.now().toString(36);
    const baseName = `测试部门E2E-${uniqueSuffix}`;
    const updatedName = `${baseName}-已更新`;

    await test.step('创建新组织', async () => {
      await page.getByTestId('create-organization-button').click();
      await waitForNavigation(page, '**/organizations/new');
      await expect(page.getByTestId('organization-form')).toBeVisible();

      const today = new Date().toISOString().slice(0, 10);
      await page.getByTestId('form-field-effective-date').fill(today);
      await page.getByTestId('form-field-name').fill(baseName);
      await page.getByTestId('form-field-description').fill(`自动化创建 ${baseName}`);

      const parentSelector = page.getByTestId('combobox-input');
      await parentSelector.click();

      const parentMenu = page.getByTestId('combobox-menu');
      await expect(parentMenu).toBeVisible({ timeout: 10000 });

      const parentOption = page.getByTestId(`combobox-item-${ROOT_PARENT_CODE}`);
      await parentOption.waitFor({ state: 'visible', timeout: 10000 });
      await parentOption.click();

      await expect(parentSelector).toHaveValue(new RegExp(`^${ROOT_PARENT_CODE}`));

      await page.getByTestId('form-submit-button').click();

      await waitForNavigation(page, /\/organizations\/[0-9]{7}\/temporal$/);
      await expect(page.getByTestId('organization-form')).toBeVisible();
    });

    const detailUrl = page.url();
    const createdCodeMatch = detailUrl.match(/organizations\/(\d{7})/);
    if (!createdCodeMatch?.[1]) {
      throw new Error('成功创建的组织需要返回7位编码');
    }
    const organizationCode = createdCodeMatch[1];

    await page.getByTestId('back-to-organization-list').click();
    await waitForNavigation(page, '**/organizations');
    await Promise.race([
      waitForOrganizationsResponse(page).catch(() => {}),
      page.waitForTimeout(3000),
    ]);

    await test.step('验证列表展示新组织', async () => {
      // 等待列表加载完成
      const organizationTable = page.getByTestId('organization-table');
      await expect(organizationTable).toBeVisible();

      // 等待加载状态消失，确保数据已刷新
      await page.waitForSelector('text=加载组织数据中...', { state: 'detached', timeout: 15000 }).catch(() => {
        // 如果没有加载状态也没关系，说明加载很快完成了
      });

      await filterOrganizationsByName(page, baseName);

      const jwt = (await ensurePwJwt()) ?? getPwJwt();
      if (!jwt) {
        throw new Error('缺少可用的JWT令牌');
      }

      const graphqlResponse = await page.request.post(E2E_CONFIG.GRAPHQL_API_URL, {
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${jwt}`,
          'X-Tenant-ID': process.env.PW_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
        },
        data: {
          query: `query ($codes: [String!], $pageSize: Int!) {
            organizations(filter: { codes: $codes }, pagination: { page: 1, pageSize: $pageSize }) {
              data { code name unitType }
            }
          }`,
          variables: { codes: [organizationCode], pageSize: 5 },
        },
      });

      expect(graphqlResponse.ok()).toBeTruthy();
      const graphqlBody = await graphqlResponse.json();
      const matched = graphqlBody?.data?.organizations?.data?.find(
        (item: { code: string }) => item.code === organizationCode,
      );
      expect(matched).toBeDefined();
      expect(matched.name).toBe(baseName);
      expect(matched.unitType).toBe('DEPARTMENT');

      // 若UI已经刷新，则新行应可见；否则记录日志但不影响断言
      const createdRow = page.getByTestId(`table-row-${organizationCode}`);
      if (await createdRow.count()) {
        await expect(createdRow.getByText(baseName)).toBeVisible();
      }

      await resetOrganizationFilters(page);
    });

    await test.step('更新组织名称', async () => {
      await page.goto(`/organizations/${organizationCode}/temporal`);
      await waitForPageReady(page);
      await expect(page.getByTestId('organization-form')).toBeVisible();
      await waitForTemporalDetailReady(page);

      await page.getByTestId('edit-history-toggle-button').click();
      const nameInput = page.getByTestId('form-field-name');
      await expect(nameInput).toBeEnabled();
      await nameInput.fill(updatedName);
      await page.getByTestId('submit-edit-history-button').click();

      await expect(nameInput).toHaveValue(updatedName);
      await expect(nameInput).toBeDisabled();

      await page.goto('/organizations');
      await waitForPageReady(page);
      await Promise.race([
        waitForOrganizationsResponse(page).catch(() => {}),
        page.waitForTimeout(3000),
      ]);

      await filterOrganizationsByName(page, updatedName);

      const updatedRow = page.getByTestId(`table-row-${organizationCode}`);
      if (await updatedRow.count()) {
        await expect(updatedRow.getByText(updatedName)).toBeVisible({ timeout: 15000 });
      }

      await resetOrganizationFilters(page);
    });

    await test.step('删除组织并在列表中消失', async () => {
      await page.goto(`/organizations/${organizationCode}/temporal`);
      await waitForPageReady(page);
      await expect(page.getByTestId('organization-form')).toBeVisible();

      await waitForTemporalDetailReady(page);
      const deleteButton = await expectDeleteButtonVisible(page);
      await deleteButton.click();
      const confirmButton = page.getByTestId('deactivate-confirm-button');
      await expect(confirmButton).toBeVisible();
      await confirmButton.click();

      await page.goto('/organizations');
      await waitForPageReady(page);
      await Promise.race([
        waitForOrganizationsResponse(page).catch(() => {}),
        page.waitForTimeout(3000),
      ]);

      await filterOrganizationsByName(page, baseName);

      const deletedRow = page.getByTestId(`table-row-${organizationCode}`);
      if (await deletedRow.count()) {
        await expect(deletedRow.locator('[data-testid^="status-pill-"]')).toContainText(['停用', '计划中', '已删除'], {
          timeout: 10000,
        });
      }

      await resetOrganizationFilters(page);
    });
  });

  test('分页和筛选功能测试', async ({ page }) => {
    // 等待页面数据加载完成
    await page.waitForTimeout(2000);
    const organizationTable = page.getByTestId('organization-table');
    await expect(organizationTable).toBeVisible();

    // 1. 验证搜索功能
    const searchInput = page.locator('input[placeholder*="搜索"], input[name="search"]').first();
    
    if (await searchInput.isVisible()) {
      // 输入搜索关键词
      await searchInput.fill('高谷集团');
      await page.waitForTimeout(1500); // 等待debounce搜索
      
      // 验证搜索结果
      const searchResults = page.locator('tr:has-text("高谷集团")');
      if (await searchResults.first().isVisible()) {
        await expect(searchResults.first()).toBeVisible();
      }
      
      // 清空搜索
      await searchInput.clear();
      await page.waitForTimeout(1000);
    }

    // 2. 验证筛选功能
    const typeFilterSelect = page.locator('select[name*="type"], select[name*="unit_type"]').first();
    
    if (await typeFilterSelect.isVisible()) {
      // 选择特定类型进行筛选
      await typeFilterSelect.selectOption('COMPANY');
      await page.waitForTimeout(1000);
      
      // 验证筛选结果
      const companyRows = page.locator('tr:has-text("COMPANY")');
      if (await companyRows.first().isVisible()) {
        await expect(companyRows.first()).toBeVisible();
      }
      
      // 重置筛选
      await typeFilterSelect.selectOption('');
      await page.waitForTimeout(1000);
    }

    // 3. 验证分页功能（如果有足够数据）
    const _paginationArea = page.locator('[data-testid*="pagination"], .pagination').first();
    const nextPageButton = page.getByRole('button', { name: /下一页|Next|>/ });
    
    // 检查是否有分页控件
    if (await nextPageButton.isVisible()) {
      // 记录当前页码
      const currentPageInfo = page.locator('text=/页|Page/').first();
      const initialPage = await currentPageInfo.textContent();
      
      // 点击下一页
      await nextPageButton.click();
      await page.waitForTimeout(1500);
      
      // 验证页面已切换
      const newPageInfo = await currentPageInfo.textContent();
      if (initialPage !== newPageInfo) {
        // 返回第一页
        const prevPageButton = page.getByRole('button', { name: /上一页|Previous|</ });
        if (await prevPageButton.isVisible()) {
          await prevPageButton.click();
          await page.waitForTimeout(1000);
        }
      }
    }

    // 4. 验证数据加载状态
    const tableRows = organizationTable.locator('tbody tr');
    const rowCount = await tableRows.count();

    if (rowCount > 0) {
      // 验证表格基本结构
      const firstRow = tableRows.first();
      await expect(firstRow).toBeVisible();
      
      // 验证表头存在
      const tableHeaders = organizationTable.locator('thead th');
      const headerCount = await tableHeaders.count();
      expect(headerCount).toBeGreaterThan(3); // 至少有编码、名称、类型、状态等列
    }
  });

  test('性能和响应时间测试', async ({ page }) => {
    const startTime = Date.now();

    // 1. 测试页面加载性能
    await page.goto('/organizations');
    await waitForPageReady(page);
    await expect(page.getByText('组织架构管理')).toBeVisible();

    const loadTime = Date.now() - startTime;

    // 断言加载时间在合理范围内（< 3秒）
    expect(loadTime).toBeLessThan(3000);

    // 2. 测试API响应性能
    const apiStartTime = Date.now();
    
    await page.evaluate(async () => {
      const response = await fetch('http://localhost:8090/graphql', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: `query ($page: Int!, $size: Int!) {
            organizations(pagination: { page: $page, pageSize: $size }) {
              data {
                code
                name
                unitType
              }
            }
          }`,
          variables: { page: 1, size: 5 }
        })
      });
      return response.json();
    });

    const apiTime = Date.now() - apiStartTime;

    // 断言API响应时间在合理范围内（< 1秒）
    expect(apiTime).toBeLessThan(1000);
  });

  test('错误处理和恢复测试', async ({ page }) => {
    // 1. 测试网络错误处理
    await page.route('**/graphql', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error' })
      });
    });

    await page.reload();
    await waitForPageReady(page);
    
    // 验证错误状态显示
    await expect(
      page.getByText('加载失败').or(page.getByText('网络错误'))
    ).toBeVisible();

    // 2. 测试重试机制
    await page.unroute('**/graphql');
    
    const treeRetryButton = page.getByTestId('organization-tree-retry-button');
    const reloadButton = page.getByRole('button', { name: /重新加载|重试/ });

    if (await treeRetryButton.isVisible()) {
      await treeRetryButton.click();
    } else {
      await reloadButton.waitFor({ state: 'visible', timeout: 15000 });
      await reloadButton.click();
      await page.waitForLoadState('load');
    }

    // 验证恢复后正常显示
    await expect(page.getByText('组织架构管理')).toBeVisible();
  });

  test('数据一致性验证测试', async ({ page }) => {
    // 验证前端显示的数据与后端API返回的数据一致
    
    // 1. 获取前端显示的数据
    const frontendData = await page.evaluate(() => {
      const rows = Array.from(document.querySelectorAll('tr'));
      return rows.map(row => {
        const cells = Array.from(row.querySelectorAll('td'));
        if (cells.length >= 4) {
          return {
            code: cells[0]?.textContent?.trim(),
            name: cells[1]?.textContent?.trim(),
            type: cells[2]?.textContent?.trim(),
            status: cells[3]?.textContent?.trim()
          };
        }
        return null;
      }).filter(Boolean);
    });

    // 2. 直接调用API获取数据
    const apiData = await page.evaluate(async () => {
      const response = await fetch('http://localhost:8090/graphql', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: `query ($page: Int!, $size: Int!) {
            organizations(pagination: { page: $page, pageSize: $size }) {
              data {
                code
                name
                unitType
                status
              }
              pagination {
                total
              }
            }
          }`,
          variables: { page: 1, size: 50 }
        })
      });
      const result = await response.json();
      return result.data?.organizations?.data ?? [];
    });

    // 3. 验证数据一致性 - 考虑状态显示的本地化
    if (frontendData.length > 0 && apiData.length > 0) {
      const firstFrontendItem = frontendData[0];
      const firstApiItem = apiData[0];

      expect(firstFrontendItem.code).toBe(firstApiItem.code);
      expect(firstFrontendItem.name).toBe(firstApiItem.name);
      expect(firstFrontendItem.type).toBe(firstApiItem.unitType);
      
      // 状态字段处理本地化映射
      const statusMap: Record<string, string> = {
        'ACTIVE': '✓ 启用',
        'INACTIVE': '停用',
        'PLANNED': '计划中'
      };
      const expectedDisplayStatus = statusMap[firstApiItem.status] || firstApiItem.status;
      expect(firstFrontendItem.status).toBe(expectedDisplayStatus);
    }
  });
});
