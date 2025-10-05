import { test, expect } from '@playwright/test';
import { setupAuth } from './auth-setup';
import { ensurePwJwt, getPwJwt } from './utils/authToken';

test.describe('回归测试和兼容性验证', () => {
  
  test.beforeEach(async ({ page }) => {
    await setupAuth(page);
  });

  test('关键功能回归测试', async ({ page }) => {
    await page.goto('/organizations');

    // 1. 验证基本UI组件正常工作
    await expect(page.getByText('组织架构管理')).toBeVisible();
    await expect(page.getByRole('button', { name: '新增组织单元' })).toBeVisible();
    await expect(page.getByRole('button', { name: '导入数据' })).toBeVisible();
    await expect(page.getByRole('button', { name: '导出报告' })).toBeVisible();

    // 2. 验证数据加载功能
    await page.waitForTimeout(2000);
    
    const hasData = await page.locator('table tbody tr').count() > 0;
    const hasNoDataMessage = await page.getByText('暂无组织数据').isVisible();
    
    expect(hasData || hasNoDataMessage).toBe(true);

    // 3. 验证Canvas组件样式兼容性
    const canvasComponents = await page.locator('[class*="css-"]').count();
    expect(canvasComponents).toBeGreaterThan(0);
  });

  test('API兼容性测试', async ({ page }) => {
    // 验证新的统一GraphQL接口向下兼容
    
    // 1. 测试GraphQL查询
    const token = await ensurePwJwt();
    const tenantId = process.env.PW_TENANT_ID ?? '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';

    if (!token && !getPwJwt()) {
      throw new Error('无法获取 GraphQL 验证所需的 RS256 JWT 令牌');
    }

    const graphqlResult = await page.evaluate(async ({ token, tenantId }) => {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json'
      };
      if (tenantId) {
        headers['X-Tenant-ID'] = tenantId;
      }
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch('http://localhost:8090/graphql', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          query: `query ($page: Int!, $size: Int!) {
            organizations(pagination: { page: $page, pageSize: $size }) {
              data {
                code
                name
                unitType
                status
                level
                parentCode
                createdAt
                updatedAt
              }
              pagination {
                total
              }
            }
          }`,
          variables: { page: 1, size: 50 }
        })
      });
      return response.json();
    }, { token: token ?? getPwJwt() ?? '', tenantId });

    expect(graphqlResult).toHaveProperty(['data', 'organizations', 'data']);

    // 2. 测试REST API兼容性（如果保留）
    const restHealth = await page.evaluate(async () => {
      try {
        const response = await fetch('http://localhost:9090/health');
        return {
          status: response.status,
          ok: response.ok,
          data: await response.json()
        };
      } catch (error) {
        return { error: (error as Error).message };
      }
    });

    expect(restHealth.status).toBe(200);
    expect(restHealth.data).toHaveProperty('status', 'healthy');
    expect(restHealth.data).toHaveProperty('service', 'organization-command-service');
  });

  test('数据迁移验证测试', async ({ page }) => {
    // 验证重构后数据完整性
    await page.goto('/organizations');

    // 1. 验证已知的测试数据存在
    const knownOrganizations = [
      '高谷集团',
      '技术部'
    ];

    for (const orgName of knownOrganizations) {
      // 尝试查找组织数据
      const orgElement = page.getByText(orgName);
      
      if (await orgElement.isVisible()) {
        console.log(`✓ 找到组织: ${orgName}`);
        
        // 验证相关数据字段完整
        const row = page.locator(`tr:has-text("${orgName}")`);
        const cells = await row.locator('td').allTextContents();
        
        expect(cells.length).toBeGreaterThan(3); // 至少有编码、名称、类型、状态
      }
    }

    // 2. 验证数据结构字段完整性
    const sampleData = await page.evaluate(async ({ token, tenantId }) => {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json'
      };
      if (tenantId) {
        headers['X-Tenant-ID'] = tenantId;
      }
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch('http://localhost:8090/graphql', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          query: `query ($page: Int!, $size: Int!) {
            organizations(pagination: { page: $page, pageSize: $size }) {
              data {
                code
                name
                unitType
                status
                level
                parentCode
              }
            }
          }`,
          variables: { page: 1, size: 1 }
        })
      });
      const result = await response.json();
      return result.data?.organizations?.data?.[0];
    }, authContext);

    if (sampleData) {
      expect(sampleData).toHaveProperty('code');
      expect(sampleData).toHaveProperty('name'); 
      expect(sampleData).toHaveProperty('unitType');
      expect(sampleData).toHaveProperty('status');
      expect(sampleData).toHaveProperty('level');
      // parentCode 可以为 null，所以使用 toHaveProperty
      expect('parentCode' in sampleData).toBe(true);
    }
  });

  test('跨浏览器兼容性验证', async ({ page, browserName }) => {
    await page.goto('/organizations');
    
    console.log(`测试浏览器: ${browserName}`);

    // 1. 验证基本功能在所有浏览器中正常工作
    await expect(page.getByText('组织架构管理')).toBeVisible();

    // 2. 验证JavaScript功能
    const jsTest = await page.evaluate(() => {
      // 测试现代JavaScript特性
      try {
        const testData = { test: 'value' };
        const { test } = testData; // 解构赋值
        return Promise.resolve(test === 'value');
      } catch (_error) {
        return false;
      }
    });

    expect(jsTest).toBe(true);

    // 3. 验证Canvas Kit在不同浏览器中的渲染
    const canvasStyles = await page.locator('body').getAttribute('style');
    
    if (canvasStyles) {
      expect(canvasStyles).toContain('--cnvs-');
    }

    // 4. 验证网络请求在不同浏览器中正常工作
    const apiTest = await page.evaluate(async () => {
      try {
        const response = await fetch('http://localhost:8090/health');
        return response.ok;
      } catch (_error) {
        return false;
      }
    });

    expect(apiTest).toBe(true);
  });

  test('性能回归测试', async ({ page }) => {
    // 验证优化后性能不劣于重构前

    const startTime = Date.now();
    await page.goto('/organizations');
    await expect(page.getByText('组织架构管理')).toBeVisible();
    const loadTime = Date.now() - startTime;

    console.log(`页面加载时间: ${loadTime}ms`);

    // 1. 页面加载性能不应劣化
    expect(loadTime).toBeLessThan(5000); // 5秒内加载完成

    // 2. 测试内存使用情况
    const memoryUsage = await page.evaluate(() => {
      // @ts-expect-error - performance.memory is not in standard types but exists in Chrome
      return performance.memory ? {
        // @ts-expect-error - performance.memory is not in standard types but exists in Chrome
        usedJSHeapSize: performance.memory.usedJSHeapSize,
        // @ts-expect-error - performance.memory is not in standard types but exists in Chrome
        totalJSHeapSize: performance.memory.totalJSHeapSize
      } : null;
    });

    if (memoryUsage) {
      console.log(`内存使用: ${(memoryUsage.usedJSHeapSize / 1024 / 1024).toFixed(2)}MB`);
      
      // 内存使用应该在合理范围内
      expect(memoryUsage.usedJSHeapSize).toBeLessThan(100 * 1024 * 1024); // < 100MB
    }

    // 3. 测试API响应性能
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
    expect(apiTime).toBeLessThan(2000); // API响应 < 2秒
  });

  test('错误边界和异常处理测试', async ({ page }) => {
    await page.goto('/organizations');

    // 1. 测试网络中断处理
    await page.route('**/*', route => route.abort());
    
    // 刷新页面触发网络错误（允许导航失败，但应展示错误状态而非白屏）
    await page.reload().catch(() => {});

    // 应该显示友好的错误信息而不是白屏
    const errorText = await page.textContent('body');
    expect(errorText).not.toBe('');
    
    // 恢复网络
    await page.unroute('**/*');

    // 2. 测试API错误处理
    await page.route('**/graphql', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Server Error' })
      });
    });

    await page.reload();
    
    // 应该有错误提示而不是崩溃
    await expect(
      page.getByText('加载失败').or(page.getByText('服务器错误'))
    ).toBeVisible();

    await page.unroute('**/graphql');

    // 3. 测试JavaScript错误处理
    const jsErrors = [];
    page.on('pageerror', error => jsErrors.push(error));

    // 触发一些操作
    await page.getByRole('button', { name: '新增组织单元' }).click();
    await page.waitForTimeout(1000);

    // 不应该有未捕获的JavaScript错误
    expect(jsErrors.length).toBe(0);
  });
});
