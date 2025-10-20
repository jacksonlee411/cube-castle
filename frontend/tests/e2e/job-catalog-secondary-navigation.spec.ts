import { test, expect, Page } from '@playwright/test';
import { TOKEN_STORAGE_KEY } from '@/shared/api/auth';
import { E2E_CONFIG, validateTestEnvironment } from './config/test-environment';
import { updateCachedJwt } from './utils/authToken';

const TENANT_ID = process.env.PW_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';
const COMMAND_BASE_URL = E2E_CONFIG.COMMAND_API_URL.replace(/\/$/, '').replace(/\/api\/v1$/, '');
const COMMAND_API_URL = E2E_CONFIG.COMMAND_API_URL.replace(/\/$/, '');

type JobFamilyGroupResponse = {
  data?: {
    recordId?: string;
    RecordID?: string;
  };
  success?: boolean;
};

type TokenOptions = {
  roles: string[];
  userId: string;
  duration?: string;
};

const generateCode = (): string => {
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  let code = 'QA';
  while (code.length < 4) {
    code += alphabet[Math.floor(Math.random() * alphabet.length)];
  }
  return code;
};

const createJobFamilyPayload = (code: string, nameSuffix: string) => ({
  code,
  name: `E2E 职类 ${nameSuffix}`,
  status: 'ACTIVE',
  effectiveDate: new Date().toISOString().slice(0, 10),
  description: `Playwright E2E 初始化 (${nameSuffix})`,
});

const ADMIN_SCOPES = [
  'job-catalog:read',
  'job-catalog:update',
  'job-catalog:write',
  'position:read',
  'position:write',
  'org:read',
  'org:write',
] as const;

const setAuthStorage = async (page: Page, token: string, scopes: readonly string[] = ADMIN_SCOPES) => {
  await page.addInitScript(
    ({ storageKey, accessToken, tenant, scopeList }) => {
      (window as typeof window & { __SCOPES__?: string[] }).__SCOPES__ = Array.from(scopeList);
      window.localStorage.setItem(
        storageKey,
        JSON.stringify({
          accessToken,
          tokenType: 'Bearer',
          expiresIn: 8 * 60 * 60,
          issuedAt: Date.now(),
          scope: Array.from(scopeList).join(' '),
        }),
      );
      window.localStorage.setItem('cube-castle-tenant-id', tenant);
    },
    { storageKey: TOKEN_STORAGE_KEY, accessToken: token, tenant: TENANT_ID, scopeList: scopes },
  );
};

const mintToken = async (
  request: import('@playwright/test').APIRequestContext,
  options: TokenOptions,
): Promise<string> => {
  const response = await request.post(`${COMMAND_BASE_URL}/auth/dev-token`, {
    data: {
      tenantId: TENANT_ID,
      roles: options.roles,
      userId: options.userId,
      duration: options.duration ?? '2h',
    },
    headers: { 'Content-Type': 'application/json' },
  });

  expect(response.ok(), '获取开发 JWT 失败').toBeTruthy();
  const json = (await response.json()) as { data?: { token?: string }; token?: string; accessToken?: string };
  const token = json?.data?.token ?? json?.token ?? json?.accessToken;
  if (!token) {
    throw new Error('开发令牌响应缺少 token 字段');
  }
  return token;
};

const createJobFamilyGroup = async (
  request: import('@playwright/test').APIRequestContext,
  token: string,
) => {
  const code = generateCode();
  const payload = createJobFamilyPayload(code, '初始化');
  const response = await request.post(`${COMMAND_API_URL}/job-family-groups`, {
    data: payload,
    headers: {
      Authorization: `Bearer ${token}`,
      'X-Tenant-ID': TENANT_ID,
      'Content-Type': 'application/json',
    },
  });

  expect(response.status(), '创建职类失败').toBe(201);
  const body = (await response.json()) as JobFamilyGroupResponse;
  const recordId = body.data?.recordId ?? body.data?.RecordID;
  if (!recordId) {
    throw new Error('创建职类响应缺少 recordId');
  }

  return { code, recordId, name: payload.name };
};

const fetchJobFamilyGroupName = async (
  request: import('@playwright/test').APIRequestContext,
  token: string,
  code: string,
) => {
  const query = `
    query JobFamilyGroups($includeInactive: Boolean) {
      jobFamilyGroups(includeInactive: $includeInactive) {
        code
        name
      }
    }
  `;

  const response = await request.post(E2E_CONFIG.GRAPHQL_API_URL, {
    data: { query, variables: { includeInactive: true } },
    headers: {
      Authorization: `Bearer ${token}`,
      'X-Tenant-ID': TENANT_ID,
      'Content-Type': 'application/json',
    },
  });

  expect(response.ok(), 'GraphQL 查询失败').toBeTruthy();
  const json = (await response.json()) as {
    data?: { jobFamilyGroups?: Array<{ code: string; name: string }> };
    errors?: unknown;
  };
  if (json.errors) {
    throw new Error(`GraphQL 返回错误: ${JSON.stringify(json.errors)}`);
  }
  const target = json.data?.jobFamilyGroups?.find(item => item.code === code);
  return target?.name ?? null;
};

test.describe.serial('职位管理二级导航（真实后端权限验证）', () => {
  let adminToken: string;
  let viewerToken: string;
  let jobFamilyGroup: { code: string; recordId: string; name: string };

  test.beforeAll(async ({ request }) => {
    const env = await validateTestEnvironment();
    test.skip(!env.isValid, env.errors.join('; '));

    const [commandHealth, graphqlHealth] = await Promise.all([
      request.get(E2E_CONFIG.COMMAND_HEALTH_URL),
      request.get(E2E_CONFIG.GRAPHQL_HEALTH_URL),
    ]);

    test.skip(!commandHealth.ok(), `命令服务不可用: ${E2E_CONFIG.COMMAND_HEALTH_URL}`);
    test.skip(!graphqlHealth.ok(), `查询服务不可用: ${E2E_CONFIG.GRAPHQL_HEALTH_URL}`);

    adminToken = await mintToken(request, { roles: ['ADMIN', 'USER'], userId: 'job-catalog-admin' });
    updateCachedJwt(adminToken);

    viewerToken = await mintToken(request, { roles: ['USER'], userId: 'job-catalog-viewer' });
    jobFamilyGroup = await createJobFamilyGroup(request, adminToken);
  });

  test('管理员通过 UI 编辑职类成功并触发 If-Match', async ({ page, request }) => {
    await setAuthStorage(page, adminToken);

    await expect
      .poll(async () => fetchJobFamilyGroupName(request, adminToken, jobFamilyGroup.code))
      .toBe(jobFamilyGroup.name);

    await page.goto(`/positions/catalog/family-groups/${jobFamilyGroup.code}`);
    await expect(page.getByRole('heading', { name: '职类详情' })).toBeVisible();
    await expect(page.getByText(jobFamilyGroup.name)).toBeVisible();

  await page.getByRole('button', { name: '编辑当前版本' }).click();
  await expect(page.getByRole('heading', { name: '编辑职类信息' })).toBeVisible();

  const newName = `${jobFamilyGroup.name}-已更新`;

  const nameInput = page.getByPlaceholder('版本名称');
  await nameInput.fill(newName);

  const [response] = await Promise.all([
    page.waitForResponse(res => res.url().includes(`/api/v1/job-family-groups/${jobFamilyGroup.code}`) && res.request().method() === 'PUT'),
    page.getByRole('button', { name: /保存更新|提交|确认/ }).click(),
  ]);

    expect(response.status()).toBe(200);

  await expect
    .poll(async () => fetchJobFamilyGroupName(request, adminToken, jobFamilyGroup.code))
    .toBe(newName);
    jobFamilyGroup = { ...jobFamilyGroup, name: newName };
  });

  test('普通用户更新职类遭遇 403 禁止访问', async ({ request }) => {
    const response = await request.put(`${COMMAND_API_URL}/job-family-groups/${jobFamilyGroup.code}`, {
      data: {
        name: `${jobFamilyGroup.name}-403`,
        status: 'ACTIVE',
        effectiveDate: new Date().toISOString().slice(0, 10),
      },
      headers: {
        Authorization: `Bearer ${viewerToken}`,
        'X-Tenant-ID': TENANT_ID,
        'Content-Type': 'application/json',
        'If-Match': jobFamilyGroup.recordId,
      },
    });

    expect(response.status()).toBe(403);
  });

  test('If-Match 不匹配返回 412 Precondition Failed', async ({ request }) => {
    const response = await request.put(`${COMMAND_API_URL}/job-family-groups/${jobFamilyGroup.code}`, {
      data: {
        name: `${jobFamilyGroup.name}-412`,
        status: 'ACTIVE',
        effectiveDate: new Date().toISOString().slice(0, 10),
      },
      headers: {
        Authorization: `Bearer ${adminToken}`,
        'X-Tenant-ID': TENANT_ID,
        'Content-Type': 'application/json',
        'If-Match': `${jobFamilyGroup.recordId}-stale`,
      },
    });

    expect(response.status()).toBe(412);
  });
});
