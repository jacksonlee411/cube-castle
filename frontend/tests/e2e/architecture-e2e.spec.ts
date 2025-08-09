import { test, expect } from '@playwright/test';

test.describe('重构后架构完整性验证', () => {
  
  test.beforeEach(async ({ page }) => {
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
    // 验证GraphQL端点
    const graphqlResponse = await page.evaluate(async () => {
      try {
        const response = await fetch('http://localhost:8090/graphql', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            query: `{
              organizations {
                code
                name
                unitType
                status
              }
            }`
          })
        });
        return await response.json();
      } catch (error) {
        return { error: error.message };
      }
    });

    expect(graphqlResponse).toHaveProperty('data');
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