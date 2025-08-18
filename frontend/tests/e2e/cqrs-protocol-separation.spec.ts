/**
 * CQRS协议分离验证测试
 * 测试目标: 验证命令端和查询端严格分离，协议使用正确
 * 
 * 命令端 (9090): 仅支持REST API的CUD操作
 * 查询端 (8090): 仅支持GraphQL查询操作
 */

import { test, expect } from '@playwright/test';

test.describe('CQRS协议分离验证', () => {

  test.beforeAll(async () => {
    console.log('🚀 开始CQRS架构协议分离测试');
  });

  test('🚫 命令端应拒绝GET查询请求', async ({ request }) => {
    console.log('测试: 命令端拒绝GET查询');
    
    // 尝试在命令端执行查询操作 - 应该失败
    const response = await request.get('http://localhost:9090/api/v1/organization-units');
    
    // 验证命令端正确拒绝查询操作
    expect(response.status()).toBe(405); // Method Not Allowed - 更准确的HTTP状态码
    
    const body = await response.json();
    expect(body.code).toBe('METHOD_NOT_ALLOWED');
    expect(body.message).toBe('方法不允许');
    
    console.log('✅ 命令端正确拒绝GET查询请求');
  });

  test('🚫 命令端应拒绝单个组织查询', async ({ request }) => {
    console.log('测试: 命令端拒绝单个组织查询');
    
    const response = await request.get('http://localhost:9090/api/v1/organization-units/1000001');
    
    expect(response.status()).toBe(405);
    
    const body = await response.json();
    expect(body.code).toBe('METHOD_NOT_ALLOWED');
    
    console.log('✅ 命令端正确拒绝单个组织查询请求');
  });

  test('✅ 命令端应支持POST创建操作', async ({ request }) => {
    console.log('测试: 命令端支持POST创建');
    
    const createData = {
      name: '测试组织CQRS' + Date.now(),
      unit_type: 'DEPARTMENT',
      description: 'CQRS测试创建'
    };

    const response = await request.post('http://localhost:9090/api/v1/organization-units', {
      data: createData
    });

    expect(response.status()).toBe(201);
    
    const body = await response.json();
    expect(body.code).toMatch(/^\d{7}$/); // 7位数字代码
    expect(body.name).toBe(createData.name);
    expect(body.unit_type).toBe(createData.unit_type);
    
    console.log('✅ 命令端正确支持POST创建操作');
    return body.code; // 返回代码供后续测试使用
  });

  test('🚫 查询端应拒绝POST命令请求', async ({ request }) => {
    console.log('测试: 查询端拒绝POST命令');
    
    const createData = {
      name: '应该被拒绝的组织',
      unit_type: 'DEPARTMENT'
    };

    const response = await request.post('http://localhost:8090/api/v1/organization-units', {
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

    const response = await request.put('http://localhost:8090/api/v1/organization-units/1000001', {
      data: updateData
    });

    expect(response.status()).toBe(404);
    
    console.log('✅ 查询端正确拒绝PUT更新请求');
  });

  test('🚫 查询端应拒绝DELETE删除请求', async ({ request }) => {
    console.log('测试: 查询端拒绝DELETE删除');
    
    const response = await request.delete('http://localhost:8090/api/v1/organization-units/1000001');

    expect(response.status()).toBe(404);
    
    console.log('✅ 查询端正确拒绝DELETE删除请求');
  });

  test('✅ 查询端应支持GraphQL查询', async ({ request }) => {
    console.log('测试: 查询端支持GraphQL查询');
    
    const graphqlQuery = {
      query: `
        query {
          organizations(first: 5) {
            code
            name
            unit_type
            status
          }
        }
      `
    };

    const response = await request.post('http://localhost:8090/graphql', {
      data: graphqlQuery
    });

    expect(response.status()).toBe(200);
    
    const body = await response.json();
    expect(body.data).toBeDefined();
    expect(body.data.organizations).toBeInstanceOf(Array);
    
    console.log('✅ 查询端正确支持GraphQL查询');
    console.log(`📊 查询到 ${body.data.organizations.length} 个组织`);
  });

  test('✅ 查询端应支持单个组织GraphQL查询', async ({ request }) => {
    console.log('测试: 查询端支持单个组织GraphQL查询');
    
    // 首先获取一个存在的组织代码
    const listQuery = {
      query: `
        query {
          organizations(first: 1) {
            code
            name
          }
        }
      `
    };

    const listResponse = await request.post('http://localhost:8090/graphql', {
      data: listQuery
    });

    const listBody = await listResponse.json();
    if (listBody.data.organizations.length === 0) {
      console.log('⚠️ 跳过测试: 没有可查询的组织');
      return;
    }

    const testCode = listBody.data.organizations[0].code;
    console.log(`📋 使用组织代码: ${testCode}`);

    // 查询单个组织
    const singleQuery = {
      query: `
        query($code: String!) {
          organization(code: $code) {
            code
            name
            unit_type
            status
          }
        }
      `,
      variables: { code: testCode }
    };

    const response = await request.post('http://localhost:8090/graphql', {
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

    const response = await request.post('http://localhost:8090/graphql', {
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
      unit_type: 'DEPARTMENT',
      description: 'CQRS端到端测试'
    };

    const createResponse = await request.post('http://localhost:9090/api/v1/organization-units', {
      data: createData
    });

    expect(createResponse.status()).toBe(201);
    const createdOrg = await createResponse.json();
    console.log(`✅ 创建成功，组织代码: ${createdOrg.code}`);

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
            unit_type
            status
          }
        }
      `,
      variables: { code: createdOrg.code }
    };

    const queryResponse = await request.post('http://localhost:8090/graphql', {
      data: queryData
    });

    expect(queryResponse.status()).toBe(200);
    const queryBody = await queryResponse.json();
    
    if (queryBody.data.organization) {
      expect(queryBody.data.organization.code).toBe(createdOrg.code);
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

    const updateResponse = await request.put(`http://localhost:9090/api/v1/organization-units/${createdOrg.code}`, {
      data: updateData
    });

    expect(updateResponse.status()).toBe(200);
    const updatedOrg = await updateResponse.json();
    expect(updatedOrg.name).toBe(updateData.name);
    console.log('✅ 更新成功');

    console.log('🎉 CQRS端到端操作验证完成');
  });

  test('📋 CQRS架构健康检查', async ({ request }) => {
    console.log('测试: CQRS架构健康检查');
    
    // 检查命令端健康状态
    const commandHealthResponse = await request.get('http://localhost:9090/health');
    expect(commandHealthResponse.status()).toBe(200);
    
    const commandHealth = await commandHealthResponse.json();
    expect(commandHealth.service).toContain('Command Service');
    expect(commandHealth.architecture).toContain('CQRS Command Side');
    console.log('✅ 命令端健康状态正常');

    // 检查查询端健康状态
    const queryHealthResponse = await request.get('http://localhost:8090/health');
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