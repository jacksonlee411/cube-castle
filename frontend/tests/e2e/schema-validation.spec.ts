import { test, expect } from '@playwright/test';
import { E2E_CONFIG, validateTestEnvironment } from './config/test-environment';
import { ensurePwJwt, getPwJwt } from './utils/authToken';

const TENANT_ID = process.env.PW_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';

let commandApiUrl: string;
let graphqlUrl: string;

const authHeaders = (token: string) => ({
  Authorization: `Bearer ${token}`,
  'X-Tenant-ID': TENANT_ID,
  'Content-Type': 'application/json',
});

test.describe('Schema & 契约验证', () => {
  test.beforeAll(async () => {
    const env = await validateTestEnvironment({ allowUnreachableFrontend: true });
    if (!env.isValid) {
      throw new Error(`测试环境不可用: ${env.errors.join(', ')}`);
    }
    commandApiUrl = E2E_CONFIG.COMMAND_API_URL.replace(/\/$/, '');
    graphqlUrl = E2E_CONFIG.GRAPHQL_API_URL;
  });

  test('REST 创建组织返回契约字段', async ({ request }) => {
    const token = (await ensurePwJwt()) ?? getPwJwt();
    expect(token, '缺少有效JWT').toBeTruthy();

    const payload = {
      name: `Schema验证-${Date.now()}`,
      unitType: 'DEPARTMENT',
      parentCode: '1000000',
      description: '契约验证自动化',
      effectiveDate: new Date().toISOString().slice(0, 10),
      operationReason: 'schema-validation',
    };

    const response = await request.post(`${commandApiUrl}/organization-units`, {
      headers: authHeaders(token!),
      data: payload,
    });

    expect(response.status()).toBe(201);
    const body = await response.json();

    expect(body).toMatchObject({
      success: true,
      data: expect.objectContaining({
        code: expect.any(String),
        name: payload.name,
        unitType: payload.unitType,
        status: 'ACTIVE',
        parentCode: payload.parentCode,
      }),
    });
  });

  test('REST 创建组织缺少字段时返回 400', async ({ request }) => {
    const token = (await ensurePwJwt()) ?? getPwJwt();
    expect(token).toBeTruthy();

    const response = await request.post(`${commandApiUrl}/organization-units`, {
      headers: authHeaders(token!),
      data: {
        unitType: 'DEPARTMENT',
        parentCode: '1000000',
      },
    });

    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body).toMatchObject({
      success: false,
      error: expect.objectContaining({ code: expect.any(String) }),
    });
  });

  test('GraphQL organizations 字段类型正确', async ({ request }) => {
    const token = (await ensurePwJwt()) ?? getPwJwt();
    expect(token).toBeTruthy();

    const response = await request.post(graphqlUrl, {
      headers: authHeaders(token!),
      data: {
        query: `query ($page:Int!, $size:Int!) {
          organizations(pagination:{page:$page,pageSize:$size}) {
            data {
              code
              name
              unitType
              status
            }
          }
        }`,
        variables: { page: 1, size: 5 },
      },
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    const list = body?.data?.organizations?.data ?? [];
    expect(Array.isArray(list)).toBe(true);
    if (list.length > 0) {
      const item = list[0];
      expect(item).toMatchObject({
        code: expect.any(String),
        name: expect.any(String),
        unitType: expect.any(String),
        status: expect.any(String),
      });
    }
  });

  test('GraphQL 错误响应携带 message', async ({ request }) => {
    const token = (await ensurePwJwt()) ?? getPwJwt();
    expect(token).toBeTruthy();

    const response = await request.post(graphqlUrl, {
      headers: authHeaders(token!),
      data: {
        query: `query {
          organizations(pagination:{page:1,pageSize:5}) {
            data {
              nonexistentField
            }
          }
        }`,
      },
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.errors?.[0]?.message).toContain('nonexistentField');
  });
});
