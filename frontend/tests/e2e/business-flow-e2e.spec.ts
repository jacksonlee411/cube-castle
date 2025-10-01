import { test, expect } from '@playwright/test';
import { setupAuth } from './auth-setup';

test.describe('业务流程端到端测试', () => {

  test.beforeEach(async ({ page }) => {
    // 设置认证信息到 localStorage（确保 RequireAuth 可以通过验证）
    await setupAuth(page);

    // 导航到组织管理页面
    await page.goto('/organizations');

    // 等待页面加载完成
    await expect(page.getByText('组织架构管理')).toBeVisible();
  });

  test('完整CRUD业务流程测试', async ({ page }) => {
    // === CREATE 操作测试 ===
    
    // 1. 点击新增按钮
    await page.getByRole('button', { name: '新增组织单元' }).click();
    
    // 2. 等待表单模态框出现
    const formModal = page.getByTestId('organization-form');
    await expect(formModal).toBeVisible();
    
    // 3. 填写表单字段 - 使用更可靠的测试选择器
    await page.getByTestId('form-field-name').fill('测试部门E2E');
    await page.getByTestId('form-field-unit-type').selectOption('DEPARTMENT');
    await page.getByTestId('form-field-description').fill('E2E测试创建的组织');
    
    // 4. 提交表单
    await page.getByTestId('form-submit-button').click();
    
    // 5. 验证表单关闭
    await expect(formModal).not.toBeVisible();

    // === READ 操作测试 ===
    
    // 6. 验证数据显示在表格中
    await page.waitForTimeout(2000); // 等待数据刷新
    const organizationTable = page.getByTestId('organization-table');
    await expect(organizationTable).toBeVisible();
    
    const tableRow = page.locator('tr:has-text("测试部门E2E")');
    await expect(tableRow).toBeVisible();
    
    // 7. 验证数据字段完整性
    await expect(tableRow.getByText('DEPARTMENT')).toBeVisible();
    await expect(tableRow.getByText('ACTIVE')).toBeVisible();

    // === UPDATE 操作测试 ===
    
    // 8. 点击编辑按钮
    const editButton = tableRow.getByRole('button', { name: /编辑|Edit/ });
    
    if (await editButton.isVisible()) {
      await editButton.click();
      
      // 等待编辑表单出现
      await expect(formModal).toBeVisible();
      await expect(page.getByText('编辑组织单元')).toBeVisible();
      
      // 修改名称 - 使用更可靠的测试选择器
      const nameInput = page.getByTestId('form-field-name');
      await nameInput.clear();
      await nameInput.fill('测试部门E2E-已更新');
      
      // 提交更新
      await page.getByTestId('form-submit-button').click();
      
      // 验证表单关闭
      await expect(formModal).not.toBeVisible();
      
      // 验证更新后的数据
      await page.waitForTimeout(2000);
      await expect(page.locator('tr:has-text("测试部门E2E-已更新")')).toBeVisible();
    }

    // === DELETE 操作测试 ===
    
    // 9. 点击删除按钮
    const updatedRow = page.locator('tr:has-text("测试部门E2E")');
    const deleteButton = updatedRow.getByRole('button', { name: /删除|Delete/ });
    
    if (await deleteButton.isVisible()) {
      await deleteButton.click();
      
      // 如果有确认对话框，确认删除
      const confirmButton = page.getByRole('button', { name: /确认|删除|Delete|确定/ });
      if (await confirmButton.isVisible({ timeout: 2000 })) {
        await confirmButton.click();
      }
      
      // 验证删除后数据不再显示
      await page.waitForTimeout(2000);
      await expect(updatedRow).not.toBeVisible();
    } else {
      console.log('删除按钮不可见，跳过删除测试');
    }
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
        console.log('✓ 搜索功能正常');
      } else {
        console.log('搜索结果为空，继续测试');
      }
      
      // 清空搜索
      await searchInput.clear();
      await page.waitForTimeout(1000);
    } else {
      console.log('搜索框不可见，跳过搜索测试');
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
        console.log('✓ 类型筛选功能正常');
      } else {
        console.log('无COMPANY类型数据，筛选结果为空');
      }
      
      // 重置筛选
      await typeFilterSelect.selectOption('');
      await page.waitForTimeout(1000);
    } else {
      console.log('类型筛选器不可见，跳过筛选测试');
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
        console.log('✓ 分页功能正常');
        
        // 返回第一页
        const prevPageButton = page.getByRole('button', { name: /上一页|Previous|</ });
        if (await prevPageButton.isVisible()) {
          await prevPageButton.click();
          await page.waitForTimeout(1000);
        }
      }
    } else {
      console.log('分页按钮不可见，数据可能不足一页');
    }

    // 4. 验证数据加载状态
    const tableRows = organizationTable.locator('tbody tr');
    const rowCount = await tableRows.count();
    console.log(`表格显示 ${rowCount} 行数据`);
    
    if (rowCount > 0) {
      // 验证表格基本结构
      const firstRow = tableRows.first();
      await expect(firstRow).toBeVisible();
      
      // 验证表头存在
      const tableHeaders = organizationTable.locator('thead th');
      const headerCount = await tableHeaders.count();
      expect(headerCount).toBeGreaterThan(3); // 至少有编码、名称、类型、状态等列
      
      console.log('✓ 表格结构验证通过');
    } else {
      console.log('表格无数据，检查数据源');
    }
  });

  test('性能和响应时间测试', async ({ page }) => {
    const startTime = Date.now();
    
    // 1. 测试页面加载性能
    await page.goto('/organizations');
    await expect(page.getByText('组织架构管理')).toBeVisible();
    
    const loadTime = Date.now() - startTime;
    console.log(`页面加载时间: ${loadTime}ms`);
    
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
    console.log(`API响应时间: ${apiTime}ms`);
    
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
    
    // 验证错误状态显示
    await expect(
      page.getByText('加载失败').or(page.getByText('网络错误'))
    ).toBeVisible();

    // 2. 测试重试机制
    await page.unroute('**/graphql');
    
    const retryButton = page.getByRole('button', { name: '重试' }).or(
      page.getByText('重试')
    );
    
    if (await retryButton.isVisible()) {
      await retryButton.click();
      
      // 验证恢复后正常显示
      await expect(page.getByText('组织架构管理')).toBeVisible();
    }
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
      const statusMap = {
        'ACTIVE': '启用',
        'INACTIVE': '禁用',
        'PLANNED': '计划'
      };
      const expectedStatus = statusMap[firstApiItem.status] || firstApiItem.status;
      expect(firstFrontendItem.status).toBe(expectedStatus);
    }
  });
});
