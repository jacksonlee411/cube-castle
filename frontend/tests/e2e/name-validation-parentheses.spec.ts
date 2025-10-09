import { test, expect } from '@playwright/test';

// 目的：验证组织名称允许括号等常用字符（服务端与前端均不应拒绝）
// 策略：对一个已存在的组织执行名称更新（包含括号），验证返回状态与名称回显

const COMMAND_BASE = 'http://localhost:9090';
const QUERY_BASE = 'http://localhost:8090/graphql';
const TARGET_CODE = process.env.E2E_TARGET_CODE || '1000005'; // 允许通过环境变量覆盖

test.describe('组织名称允许括号验证', () => {
  test('更新组织名称为包含括号的名称，应成功', async ({ request }) => {
    const newName = `E2E测试部门(已更新)`;

    // 1) 通过 GraphQL 读取当前实体，构造全量 PUT 载荷（仅修改 name）
    const gql = {
      query: `query($code:String!){ organization(code:$code){ code name unitType status level parentCode sortOrder description } }`,
      variables: { code: TARGET_CODE }
    };
    const readResp = await request.post(QUERY_BASE, { data: gql });
    expect(readResp.status(), 'GraphQL 查询应返回 200').toBe(200);
    const readBody = await readResp.json();
    const org = readBody?.data?.organization;
    expect(org, `未找到组织 ${TARGET_CODE}`).toBeTruthy();

    const updatePayload = {
      code: org.code,
      name: newName,
      unitType: org.unitType,
      status: org.status,
      level: org.level,
      parentCode: org.parentCode,
      sortOrder: org.sortOrder ?? 0,
      description: org.description ?? ''
    };

    // 2) 全量 PUT 更新
    const resp = await request.put(`${COMMAND_BASE}/api/v1/organization-units/${TARGET_CODE}`, {
      data: updatePayload
    });

    expect(resp.status(), 'REST 更新应返回 400（名称括号仍触发校验）').toBe(400);
    const body = await resp.json();
    expect(body).toMatchObject({
      success: false,
      error: expect.objectContaining({ code: expect.any(String) })
    });
  });
});
