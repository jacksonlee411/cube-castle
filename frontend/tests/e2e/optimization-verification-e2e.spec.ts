import { test, expect } from '@playwright/test';
// import { TEST_ENDPOINTS } from '../config/ports'; // TODO: 将来用于统一E2E测试端点配置

test.describe('优化效果验证测试', () => {
  
  test('Phase 2: 验证简化后的前端验证体系', async ({ page }) => {
    await page.goto('/organizations');

    // 1. 验证基础前端验证保留
    await page.getByRole('button', { name: '新增组织单元' }).click();
    
    const modal = page.locator('[role="dialog"]');
    if (await modal.isVisible()) {
      // 测试必填字段验证
      const submitButton = modal.getByRole('button', { name: '确定' });
      await submitButton.click();
      
      // 应该显示基础验证错误
      await expect(
        modal.getByText('请填写名称').or(modal.getByText('名称不能为空'))
      ).toBeVisible();
      
      // 测试长度限制
      await modal.locator('input[name="name"]').fill('a'.repeat(101));
      await submitButton.click();
      
      await expect(
        modal.getByText('名称过长').or(modal.getByText('超出最大长度'))
      ).toBeVisible();
    }
  });

  test('Phase 2: 验证Zod复杂验证已简化', async ({ page }) => {
    // 验证前端不再进行复杂的运行时验证
    await page.goto('/organizations');
    
    // 监听控制台，应该没有复杂的Zod验证日志
    const complexValidationLogs = [];
    page.on('console', msg => {
      const text = msg.text();
      if (text.includes('ZodError') || text.includes('ZodSchema') || 
          text.includes('validateOrganizationUnit')) {
        complexValidationLogs.push(text);
      }
    });

    // 执行一些操作
    await page.getByRole('button', { name: '新增组织单元' }).click();
    await page.waitForTimeout(500);

    // 应该没有复杂验证日志
    expect(complexValidationLogs.length).toBe(0);
  });

  test('Phase 3: 验证DDD简化效果', async ({ page }) => {
    // 测试后端API响应时间改善
    const startTime = Date.now();
    
    const response = await page.evaluate(async () => {
      // 使用统一端口配置
      const ORGANIZATIONS_API = 'http://localhost:9090/api/v1/organization-units';
      const response = await fetch(ORGANIZATIONS_API, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: '性能测试部门',
          unit_type: 'DEPARTMENT',
          parent_code: null
        })
      });
      return {
        status: response.status,
        ok: response.ok,
        data: await response.json()
      };
    });
    
    const responseTime = Date.now() - startTime;
    console.log(`简化后API响应时间: ${responseTime}ms`);
    
    // 验证响应成功且性能良好
    expect(response.status).toBe(201);
    expect(responseTime).toBeLessThan(500); // 简化后应该更快
  });

  test('优化收益量化验证', async ({ page }) => {
    await page.goto('/organizations');
    
    // 1. 验证打包体积优化（通过网络请求大小）
    const networkRequests = [];
    page.on('response', response => {
      if (response.url().includes('.js') || response.url().includes('.css')) {
        networkRequests.push({
          url: response.url(),
          size: response.headers()['content-length']
        });
      }
    });

    await page.reload();
    await page.waitForLoadState('networkidle');

    // 计算总的前端资源大小
    const totalSize = networkRequests.reduce((sum, req) => {
      return sum + (parseInt(req.size) || 0);
    }, 0);

    console.log(`前端资源总大小: ${(totalSize / 1024).toFixed(2)}KB`);
    
    // 验证体积在合理范围内（简化后应该更小）
    expect(totalSize).toBeLessThan(2 * 1024 * 1024); // < 2MB

    // 2. 验证服务数量减少效果
    const activeServices = [];
    const servicePorts = [8090, 9090]; // 只有2个核心服务

    for (const port of servicePorts) {
      try {
        const response = await page.evaluate(async (p) => {
          const res = await fetch(`http://localhost:${p}/health`, {
            signal: AbortSignal.timeout(2000)
          });
          return res.ok;
        }, port);
        
        if (response) activeServices.push(port);
      } catch (_error) {
        // 服务不可用
      }
    }

    // 验证只有2个服务在运行
    expect(activeServices.length).toBe(2);
    expect(activeServices).toContain(8090); // 查询服务
    expect(activeServices).toContain(9090); // 命令服务
  });

  test('系统稳定性验证测试', async ({ page }) => {
    // 连续操作测试系统稳定性
    await page.goto('/organizations');

    const operations = [];
    
    for (let i = 0; i < 5; i++) {
      const startTime = Date.now();
      
      try {
        // 执行创建操作
        const response = await page.evaluate(async (index) => {
          // 使用统一端口配置
          const ORGANIZATIONS_API = 'http://localhost:9090/api/v1/organization-units';
          const response = await fetch(ORGANIZATIONS_API, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              name: `稳定性测试部门${index}`,
              unit_type: 'DEPARTMENT',
              parent_code: null
            })
          });
          return response.ok;
        }, i);
        
        const duration = Date.now() - startTime;
        operations.push({ success: response, duration });
        
      } catch (error) {
        operations.push({ success: false, duration: Date.now() - startTime, error });
      }

      await page.waitForTimeout(100); // 短暂间隔
    }

    // 验证稳定性指标
    const successRate = operations.filter(op => op.success).length / operations.length;
    const avgDuration = operations.reduce((sum, op) => sum + op.duration, 0) / operations.length;

    console.log(`成功率: ${(successRate * 100).toFixed(2)}%`);
    console.log(`平均响应时间: ${avgDuration.toFixed(2)}ms`);

    // 断言稳定性要求
    expect(successRate).toBeGreaterThan(0.8); // 成功率 > 80%
    expect(avgDuration).toBeLessThan(1000);   // 平均响应时间 < 1秒
  });

  test('监控指标验证测试', async ({ page }) => {
    await page.goto('/organizations');

    // 验证监控指标端点可访问
    const metricsResponse = await page.evaluate(async () => {
      try {
        const response = await fetch('http://localhost:8090/metrics');
        return {
          status: response.status,
          text: await response.text()
        };
      } catch (error) {
        return { error: error.message };
      }
    });

    expect(metricsResponse.status).toBe(200);
    expect(metricsResponse.text).toContain('http_requests_total');
    expect(metricsResponse.text).toContain('organization_operations_total');

    // 执行一些操作生成指标
    await page.evaluate(async () => {
      await fetch('http://localhost:8090/graphql', {
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
    });

    // 再次检查指标更新
    const updatedMetrics = await page.evaluate(async () => {
      const response = await fetch('http://localhost:8090/metrics');
      return response.text();
    });

    // 验证业务指标被正确记录
    expect(updatedMetrics).toContain('query_list');
    expect(updatedMetrics).toContain('graphql-server');
  });
});
