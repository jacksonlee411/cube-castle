import { test, expect } from '@playwright/test';
import { setupAuth } from './auth-setup';

const hasAuthToken = Boolean(process.env.PW_JWT);
test.skip(!hasAuthToken, '需要 RS256 JWT 令牌运行受保护路由测试');

test.describe('重构后架构完整性验证', () => {
  
  test.beforeEach(async ({ page }) => {
    await setupAuth(page);
    await page.goto('/organizations');
  });

  test('Phase 1: 服务合并验证 - 双核心服务架构', async ({ page }) => {
    // 验证命令服务(9090端口)可访问性
    const commandResponse = await page.evaluate(async () => {
      try {
        const response = await fetch('http://localhost:9090/health');
        return { status: response.status, ok: response.ok };
      } catch (error) {
        return { error: error.message };
      }
    });

    // 验证查询服务(8090端口)可访问性
    const queryResponse = await page.evaluate(async () => {
      try {
        const response = await fetch('http://localhost:8090/health');
        return { status: response.status, ok: response.ok };
      } catch (error) {
        return { error: error.message };
      }
    });

    // 断言服务健康状态
    expect(commandResponse.status).toBe(200);
    expect(queryResponse.status).toBe(200);
  });

  test('Phase 1: GraphQL统一查询接口验证', async ({ page }) => {
    const token = process.env.PW_JWT;
    test.skip(!token, '缺少 PW_JWT 令牌，无法验证 GraphQL 接口');

    const tenantId = process.env.PW_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';

    const graphqlResult = await page.evaluate(async ({ authToken, tenant }) => {
      // 使用相对路径，通过 Vite dev server 代理到后端
      const response = await fetch('/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authToken}`,
          'X-Tenant-ID': tenant,
        },
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
                page
                pageSize
                hasNext
              }
            }
          }`,
          variables: { page: 1, size: 5 },
        }),
      });

      const body = await response.json();
      return {
        status: response.status,
        ok: response.ok,
        body,
      };
    }, { authToken: token, tenant: tenantId });

    expect(graphqlResult.status).toBe(200);
    expect(graphqlResult.ok).toBeTruthy();
    expect(graphqlResult.body).toHaveProperty('data.organizations.data');
    expect(Array.isArray(graphqlResult.body.data.organizations.data)).toBeTruthy();
  });

  test('Phase 1: 冗余服务移除验证', async ({ page }) => {
    // 验证移除的服务不再响应
    const removedServices = [
      'http://localhost:8091',  // organization-api-gateway
      'http://localhost:8092',  // organization-api-server  
      'http://localhost:8093',  // organization-query
      'http://localhost:8094'   // organization-sync-service
    ];

    for (const serviceUrl of removedServices) {
      const response = await page.evaluate(async (url) => {
        try {
          const response = await fetch(`${url}/health`, { 
            signal: AbortSignal.timeout(2000) 
          });
          return { reachable: true, status: response.status };
        } catch (error) {
          return { reachable: false, error: error.name };
        }
      }, serviceUrl);

      // 期望服务不可达（已移除）
      expect(response.reachable).toBe(false);
    }
  });
});
