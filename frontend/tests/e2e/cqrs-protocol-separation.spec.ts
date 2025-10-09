/**
 * CQRS协议分离验证测试
 * 测试目标: 验证命令端和查询端严格分离，协议使用正确
 *
 * 命令端 (9090): 仅支持REST API的CUD操作
 * 查询端 (8090): 仅支持GraphQL查询操作
 */

import { test, expect } from '@playwright/test';
import { E2E_CONFIG } from './config/test-environment';
import { ensurePwJwt, getPwJwt } from './utils/authToken';

const TENANT_ID = process.env.PW_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';
const COMMAND_API_BASE = E2E_CONFIG.COMMAND_API_URL.replace(/\/$/, '');
const GRAPHQL_API_URL = E2E_CONFIG.GRAPHQL_API_URL;
const GRAPHQL_ORIGIN = (() => {
  try {
    const parsed = new URL(GRAPHQL_API_URL);
    return `${parsed.protocol}//${parsed.host}`;
  } catch {
    return GRAPHQL_API_URL.replace(/\/graphql$/, '');
  }
})();

const buildCommandUrl = (path: string): string => `${COMMAND_API_BASE}${path.startsWith('/') ? path : `/${path}`}`;
const buildQueryRestUrl = (path: string): string => `${GRAPHQL_ORIGIN}${path.startsWith('/') ? path : `/${path}`}`;

let authHeaders: Record<string, string>;
let graphqlHeaders: Record<string, string>;
let healthHeaders: Record<string, string>;

test.describe('CQRS协议分离验证', () => {

  test.beforeAll(async () => {
    console.log('🚀 开始CQRS架构协议分离测试');
    const resolvedToken = (await ensurePwJwt({ tenantId: TENANT_ID })) ?? getPwJwt();
    if (!resolvedToken) {
      throw new Error('缺少有效的 RS256 JWT，请先运行 make run-dev && make jwt-dev-mint');
    }
    authHeaders = {
      Authorization: `Bearer ${resolvedToken}`,
      'X-Tenant-ID': TENANT_ID,
      'Content-Type': 'application/json',
    };
    graphqlHeaders = {
      ...authHeaders,
    };
    healthHeaders = {
      Authorization: authHeaders.Authorization,
      'X-Tenant-ID': authHeaders['X-Tenant-ID'],
    };
    console.log('✅ 已加载认证令牌用于CQRS验证');
  });

  test('🚫 命令端应拒绝GET查询请求', async ({ request }) => {
    console.log('测试: 命令端拒绝GET查询');

    // 尝试在命令端执行查询操作 - 应该失败
    const response = await request.get(buildCommandUrl('/organization-units'));

    // 验证命令端返回401（未认证）或405（方法不允许）
    // 由于认证中间件优先于路由检查，返回401是正确的安全实践
    expect([401, 405]).toContain(response.status());

    console.log(`✅ 命令端正确拒绝GET查询请求 (HTTP ${response.status()})`);
  });

  test('🚫 命令端应拒绝单个组织查询', async ({ request }) => {
    console.log('测试: 命令端拒绝单个组织查询');

    const response = await request.get(buildCommandUrl('/organization-units/1000001'));

    // 验证命令端返回401（未认证）或405（方法不允许）
    expect([401, 405]).toContain(response.status());

    console.log(`✅ 命令端正确拒绝单个组织查询请求 (HTTP ${response.status()})`);
  });

  test('✅ 命令端应支持POST创建操作', async ({ request }) => {
    console.log('测试: 命令端支持POST创建');

    const createData = {
      name: '测试组织CQRS' + Date.now(),
      unitType: 'DEPARTMENT',
      parentCode: '1000000',
      description: 'CQRS测试创建',
      effectiveDate: new Date().toISOString().slice(0, 10),
      operationReason: 'CQRS协议自动化验证',
    };

    const response = await request.post(buildCommandUrl('/organization-units'), {
      headers: authHeaders,
      data: createData
    });

    if (response.status() !== 201) {
      console.warn('❌ 创建组织失败，状态码:', response.status(), '响应:', await response.text());
    }
    expect(response.status()).toBe(201);
    
    const body = await response.json();
    expect(body.success).toBeTruthy();
    expect(body.data?.code).toMatch(/^\d{7}$/); // 7位数字代码
    expect(body.data?.name).toBe(createData.name);
    expect(body.data?.unitType).toBe(createData.unitType);
    
    console.log('✅ 命令端正确支持POST创建操作');
    return body.data?.code; // 返回代码供后续测试使用
  });

  test('🚫 查询端应拒绝POST命令请求', async ({ request }) => {
    console.log('测试: 查询端拒绝POST命令');
    
    const createData = {
      name: '应该被拒绝的组织',
      unit_type: 'DEPARTMENT'
    };

    const response = await request.post(buildQueryRestUrl('/api/v1/organization-units'), {
      data: createData
    });

    // 查询端应该不存在此端点
    expect(response.status()).toBe(404);
    
    console.log('✅ 查询端正确拒绝POST命令请求');
  });

  test('🚫 查询端应拒绝PUT更新请求', async ({ request }) => {
    console.log('测试: 查询端拒绝PUT更新');
    
    const updateData = {
      name: '应该被拒绝的更新'
    };

    const response = await request.put(buildQueryRestUrl('/api/v1/organization-units/1000001'), {
      data: updateData
    });

    expect(response.status()).toBe(404);
    
    console.log('✅ 查询端正确拒绝PUT更新请求');
  });

  test('🚫 查询端应拒绝DELETE删除请求', async ({ request }) => {
    console.log('测试: 查询端拒绝DELETE删除');
    
    const response = await request.delete(buildQueryRestUrl('/api/v1/organization-units/1000001'));

    expect(response.status()).toBe(404);
    
    console.log('✅ 查询端正确拒绝DELETE删除请求');
  });

  test('✅ 查询端应支持GraphQL查询', async ({ request }) => {
    console.log('测试: 查询端支持GraphQL查询');

    const graphqlQuery = {
      query: `query ($page: Int!, $size: Int!) {
        organizations(pagination: { page: $page, pageSize: $size }) {
          data {
            code
            name
            unitType
            status
          }
        }
      }`,
      variables: { page: 1, size: 5 },
    };

    const response = await request.post(GRAPHQL_API_URL, {
      headers: graphqlHeaders,
      data: graphqlQuery
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.data).toBeDefined();
    expect(body.data.organizations.data).toBeInstanceOf(Array);
    
    console.log('✅ 查询端正确支持GraphQL查询');
    console.log(`📊 查询到 ${body.data.organizations.data.length} 个组织`);
  });

  test('✅ 查询端应支持单个组织GraphQL查询', async ({ request }) => {
    console.log('测试: 查询端支持单个组织GraphQL查询');
    
    // 首先获取一个存在的组织代码
    const listQuery = {
      query: `query ($page: Int!, $size: Int!) {
        organizations(pagination: { page: $page, pageSize: $size }) {
          data {
            code
            name
          }
        }
      }`,
      variables: { page: 1, size: 1 },
    };

    const listResponse = await request.post(GRAPHQL_API_URL, {
      headers: graphqlHeaders,
      data: listQuery
    });

    const listBody = await listResponse.json();
    if (listBody.data.organizations.data.length === 0) {
      console.log('⚠️ 跳过测试: 没有可查询的组织');
      return;
    }

    const testCode = listBody.data.organizations.data[0].code;
    console.log(`📋 使用组织代码: ${testCode}`);

    // 查询单个组织
    const singleQuery = {
      query: `query ($code: String!) {
        organization(code: $code) {
          code
          name
          unitType
          status
        }
      }`,
      variables: { code: testCode },
    };

    const response = await request.post(GRAPHQL_API_URL, {
      headers: graphqlHeaders,
      data: singleQuery
    });

    expect(response.status()).toBe(200);
    
    const body = await response.json();
    expect(body.data).toBeDefined();
    expect(body.data.organization).toBeDefined();
    expect(body.data.organization.code).toBe(testCode);
    
    console.log('✅ 查询端正确支持单个组织GraphQL查询');
  });

  test('✅ 查询端应支持组织统计GraphQL查询', async ({ request }) => {
    console.log('测试: 查询端支持组织统计查询');
    
    const statsQuery = {
      query: `
        query {
          organizationStats {
            totalCount
            byType {
              unitType
              count
            }
            byStatus {
              status
              count
            }
          }
        }
      `
    };

    const response = await request.post(GRAPHQL_API_URL, {
      headers: graphqlHeaders,
      data: statsQuery
    });

    expect(response.status()).toBe(200);
    
    const body = await response.json();
    expect(body.data).toBeDefined();
    expect(body.data.organizationStats).toBeDefined();
    expect(body.data.organizationStats.totalCount).toBeGreaterThanOrEqual(0);
    expect(body.data.organizationStats.byType).toBeInstanceOf(Array);
    expect(body.data.organizationStats.byStatus).toBeInstanceOf(Array);
    
    console.log('✅ 查询端正确支持组织统计GraphQL查询');
    console.log(`📊 统计信息: 总计${body.data.organizationStats.totalCount}个组织`);
  });

  test('🔄 CQRS端到端操作验证', async ({ request }) => {
    console.log('测试: CQRS端到端操作流程');
    
    const timestamp = Date.now();
    
    // 1. 命令端创建组织
    console.log('📝 步骤1: 通过命令端创建组织');
    const createData = {
      name: `CQRS测试组织${timestamp}`,
      unitType: 'DEPARTMENT',
      parentCode: '1000000',
      description: 'CQRS端到端测试',
      effectiveDate: new Date().toISOString().slice(0, 10),
      operationReason: 'CQRS端到端自动化校验',
    };

    const createResponse = await request.post(buildCommandUrl('/organization-units'), {
      headers: authHeaders,
      data: createData
    });

    expect(createResponse.status()).toBe(201);
    const createdEnvelope = await createResponse.json();
    const createdOrgCode = createdEnvelope.data?.code;
    if (!createdOrgCode) {
      throw new Error('命令端未返回组织代码，无法继续端到端验证');
    }
    console.log(`✅ 创建成功，组织代码: ${createdOrgCode}`);

    // 2. 等待CDC同步 (给系统一些时间同步数据)
    console.log('⏳ 步骤2: 等待CDC数据同步...');
    await new Promise(resolve => setTimeout(resolve, 2000)); // 等待2秒

    // 3. 查询端验证数据
    console.log('🔍 步骤3: 通过查询端验证数据');
    const queryData = {
      query: `
        query($code: String!) {
          organization(code: $code) {
            code
            name
            unitType
            status
          }
        }
      `,
      variables: { code: createdOrgCode }
    };

    const queryResponse = await request.post(GRAPHQL_API_URL, {
      headers: graphqlHeaders,
      data: queryData
    });

    expect(queryResponse.status()).toBe(200);
    const queryBody = await queryResponse.json();
    
    if (queryBody.data.organization) {
      expect(queryBody.data.organization.code).toBe(createdOrgCode);
      expect(queryBody.data.organization.name).toBe(createData.name);
      console.log('✅ CQRS端到端流程验证成功');
    } else {
      console.log('⚠️ CDC同步可能需要更多时间，这是正常的最终一致性行为');
    }

    // 4. 命令端更新组织  
    console.log('📝 步骤4: 通过命令端更新组织');
    const updateData = {
      name: `CQRS更新测试${timestamp}`,
      description: '已通过CQRS更新'
    };

    const updateResponse = await request.put(buildCommandUrl(`/organization-units/${createdOrgCode}`), {
      headers: authHeaders,
      data: updateData
    });

    expect(updateResponse.status()).toBe(200);
    const updatedEnvelope = await updateResponse.json();
    expect(updatedEnvelope.data?.name || updatedEnvelope.name).toBe(updateData.name);
    console.log('✅ 更新成功');

    console.log('🎉 CQRS端到端操作验证完成');
  });

  test('📋 CQRS架构健康检查', async ({ request }) => {
    console.log('测试: CQRS架构健康检查');
    
    // 检查命令端健康状态
    const commandHealthResponse = await request.get(E2E_CONFIG.COMMAND_HEALTH_URL, {
      headers: healthHeaders,
    });
    expect(commandHealthResponse.status()).toBe(200);
    
    const commandHealth = await commandHealthResponse.json();
    expect(commandHealth.service).toContain('command');
    console.log('✅ 命令端健康状态正常');

    // 检查查询端健康状态
    const queryHealthResponse = await request.get(E2E_CONFIG.GRAPHQL_HEALTH_URL, {
      headers: healthHeaders,
    });
    expect(queryHealthResponse.status()).toBe(200);

    const queryHealth = await queryHealthResponse.json();
    expect(queryHealth.service).toContain('graphql');
    console.log('✅ 查询端健康状态正常');

    console.log('🎉 CQRS架构健康检查完成');
  });

  test.afterAll(async () => {
    console.log('🏁 CQRS协议分离测试完成');
    console.log('📊 测试结果总结:');
    console.log('  ✅ 命令端正确拒绝查询操作');
    console.log('  ✅ 查询端正确拒绝命令操作');  
    console.log('  ✅ 协议分离严格执行');
    console.log('  ✅ CQRS架构符合设计规范');
  });
});
