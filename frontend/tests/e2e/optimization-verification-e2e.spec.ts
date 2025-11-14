import { test, expect } from '@playwright/test';
import { setupAuth } from './auth-setup';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';
import { ensurePwJwt, getPwJwt } from './utils/authToken';
import { E2E_CONFIG } from './config/test-environment';
// import { TEST_ENDPOINTS } from '../config/ports'; // TODO: 将来用于统一E2E测试端点配置

const COMMAND_API_BASE = E2E_CONFIG.COMMAND_API_URL.replace(/\/$/, '');
const GRAPHQL_ORIGIN = (() => {
  try {
    const parsed = new URL(E2E_CONFIG.GRAPHQL_API_URL);
    return `${parsed.protocol}//${parsed.host}`;
  } catch {
    return E2E_CONFIG.GRAPHQL_API_URL.replace(/\/graphql$/, '');
  }
})();

const buildCommandEndpoint = (path: string): string => `${COMMAND_API_BASE}${path.startsWith('/') ? path : `/${path}`}`;
const buildQueryEndpoint = (path: string): string => `${GRAPHQL_ORIGIN}${path.startsWith('/') ? path : `/${path}`}`;

const buildAuthContext = async (): Promise<{ token: string; tenantId: string }> => {
  const token = await ensurePwJwt();
  const resolvedToken = token ?? getPwJwt() ?? '';
  if (!resolvedToken) {
    throw new Error('无法获取 RS256 JWT 令牌，无法完成优化验证测试');
  }
  return {
    token: resolvedToken,
    tenantId: process.env.PW_TENANT_ID ?? '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
  };
};

test.describe('优化效果验证测试', () => {
  
  test.beforeEach(async ({ page }) => {
    await setupAuth(page);
  });
  
  test('Phase 2: 验证简化后的前端验证体系', async ({ page }) => {
    await page.goto('/organizations');

    await page.getByTestId('create-organization-button').click();
    await page.waitForURL('**/organizations/new');

    const form = page.getByTestId(temporalEntitySelectors.organization.form);
    await expect(form).toBeVisible();

    // 空提交触发基础验证
    await page.getByTestId('form-submit-button').click();
    const nameError = page.getByText('组织名称是必填项');
    await expect(nameError).toBeVisible();

    await page.getByTestId('form-field-name').fill('验证前端表单');
    await page.getByTestId('form-submit-button').click();
    await expect(nameError).toHaveCount(0);

    // 直接返回列表，避免依赖禁用按钮
    await page.goto('/organizations');
    await expect(page).toHaveURL(/\/organizations/);
  });

  test('Phase 2: 验证Zod复杂验证已简化', async ({ page }) => {
    await page.goto('/organizations');

    const complexValidationLogs: string[] = [];
    page.on('console', (msg) => {
      const text = msg.text();
      if (
        text.includes('ZodError') ||
        text.includes('ZodSchema') ||
        text.includes('validateOrganizationUnit')
      ) {
        complexValidationLogs.push(text);
      }
    });

    await page.getByTestId('create-organization-button').click();
    await page.waitForURL('**/organizations/new');
    await page.getByTestId('form-field-name').focus();
    await page.waitForTimeout(500);

    expect(complexValidationLogs.length).toBe(0);

    await page.goto('/organizations');
    await expect(page).toHaveURL(/\/organizations/);
  });

  test('Phase 3: 验证DDD简化效果', async ({ page }) => {
    // 测试后端API响应时间改善
    const startTime = Date.now();
    
    const authContext = await buildAuthContext();

    const response = await page.evaluate(async ({ token, tenantId, commandEndpoint }) => {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json'
      };
      if (tenantId) {
        headers['X-Tenant-ID'] = tenantId;
      }
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch(commandEndpoint, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          name: '性能测试部门',
          unitType: 'DEPARTMENT',
          parentCode: '1000000',
          effectiveDate: new Date().toISOString().slice(0, 10),
          operationReason: 'E2E自动化验证'
        })
      });
      return {
        status: response.status,
        ok: response.ok,
        data: await response.json()
      };
    }, { ...authContext, commandEndpoint: buildCommandEndpoint('/organization-units') });
    
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
    
    // 基线：2025-11-08 调查记录 4.59 MB（含 source-map），上限放宽至 5 MB 以留出 10% 冗余
    // 参考 docs/reference/03-API-AND-TOOLS-GUIDE.md#e2e-前端资源体积基线
    expect(totalSize).toBeLessThan(5 * 1024 * 1024);

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
    const authContext = await buildAuthContext();
    const commandEndpoint = buildCommandEndpoint('/organization-units');

    const operations = [];
    
    for (let i = 0; i < 5; i++) {
      const startTime = Date.now();
      
      try {
        // 执行创建操作
        const response = await page.evaluate(async ({ index, token, tenantId, endpoint }) => {
          const headers: Record<string, string> = {
            'Content-Type': 'application/json'
          };
          if (tenantId) {
            headers['X-Tenant-ID'] = tenantId;
          }
          if (token) {
            headers['Authorization'] = `Bearer ${token}`;
          }

          const response = await fetch(endpoint, {
            method: 'POST',
            headers,
            body: JSON.stringify({
              name: `稳定性测试部门${index}`,
              unitType: 'DEPARTMENT',
              parentCode: '1000000',
              effectiveDate: new Date().toISOString().slice(0, 10),
              operationReason: 'E2E自动化稳定性验证'
            })
          });
          return response.ok;
        }, { index: i, ...authContext, endpoint: commandEndpoint });
        
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
    const authContext = await buildAuthContext();
    const metricsEndpoint = buildQueryEndpoint('/metrics');

    // 验证监控指标端点可访问
    const metricsResponse = await page.evaluate(async ({ token, tenantId, endpoint }) => {
      const headers: Record<string, string> = {};
      if (tenantId) {
        headers['X-Tenant-ID'] = tenantId;
      }
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      try {
        const response = await fetch(endpoint, {
          headers,
        });
        return {
          status: response.status,
          text: await response.text()
        };
      } catch (error) {
        return { error: (error as Error).message };
      }
    }, { ...authContext, endpoint: metricsEndpoint });

    expect(metricsResponse.status).toBe(200);
    expect(metricsResponse.text).toContain('go_gc_duration_seconds');

    // 执行一些操作生成指标
    await page.evaluate(async ({ token, tenantId, graphqlEndpoint }) => {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json'
      };
      if (tenantId) {
        headers['X-Tenant-ID'] = tenantId;
      }
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      await fetch(graphqlEndpoint, {
        method: 'POST',
        headers,
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
    }, { ...authContext, graphqlEndpoint: E2E_CONFIG.GRAPHQL_API_URL });

    // 再次检查指标更新
    const updatedMetrics = await page.evaluate(async ({ token, tenantId, endpoint }) => {
      const headers: Record<string, string> = {};
      if (tenantId) {
        headers['X-Tenant-ID'] = tenantId;
      }
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      const response = await fetch(endpoint, {
        headers,
      });
      return response.text();
    }, { ...authContext, endpoint: metricsEndpoint });

    // 验证业务指标被正确记录
    expect(updatedMetrics).toContain('organization_operations_total');
    expect(updatedMetrics).toContain('http_requests_total');
  });
});
