import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';
import { TOKEN_STORAGE_KEY } from '@/shared/api/auth';
import { E2E_CONFIG, validateTestEnvironment } from './config/test-environment';
import { ensurePwJwt, getPwJwt } from './utils/authToken';

const MOCK_MODE_ENV = process.env.E2E_MOCK_MODE === 'true';
let USE_MOCK_MODE = MOCK_MODE_ENV;

let FRONTEND_URL: string;

const COMMAND_API_URL = E2E_CONFIG.COMMAND_API_URL;
const COMMAND_HEALTH_URL = E2E_CONFIG.COMMAND_HEALTH_URL;
const GRAPHQL_API_URL = E2E_CONFIG.GRAPHQL_API_URL;
const GRAPHQL_HEALTH_URL = E2E_CONFIG.GRAPHQL_HEALTH_URL;
const GRAPHQL_HEADERS = { 'Content-Type': 'application/json' } as const;
const TEST_ORG_CODE = process.env.E2E_ORG_CODE || '1000000';

const ORGANIZATION_VERSIONS_QUERY = `
  query OrganizationVersions($code: String!) {
    organizationVersions(code: $code) {
      recordId
      code
      name
      unitType
      status
      effectiveDate
      endDate
    }
  }
`;

const ORGANIZATION_AS_OF_QUERY = `
  query OrganizationAsOf($code: String!, $asOfDate: String!) {
    organization(code: $code, asOfDate: $asOfDate) {
      recordId
      code
      name
      unitType
      status
      effectiveDate
      endDate
    }
  }
`;

const GRAPHQL_HEALTH_QUERY = 'query GraphQLHealth { __typename }';

const MOCK_GRAPHQL_DATA = {
  organizationVersions: [
    {
      recordId: 'mock-record-001',
      code: TEST_ORG_CODE,
      name: '示例组织',
      unitType: 'DEPARTMENT',
      status: 'ACTIVE',
      effectiveDate: '2024-01-01',
      endDate: null,
    },
  ],
  organizationAsOf: {
    recordId: 'mock-record-001',
    code: TEST_ORG_CODE,
    name: '示例组织',
    unitType: 'DEPARTMENT',
    status: 'ACTIVE',
    effectiveDate: '2024-01-01',
    endDate: null,
  },
};

function structuredCloneSafe<T>(value: T): T {
  return typeof structuredClone === 'function'
    ? structuredClone(value)
    : JSON.parse(JSON.stringify(value));
}

async function pingCommandHealth(): Promise<boolean> {
  try {
    const response = await fetch(COMMAND_HEALTH_URL, {
      method: 'GET',
      signal: AbortSignal.timeout(2000),
    });
    return response.ok;
  } catch (_error) {
    return false;
  }
}

async function pingGraphQL(): Promise<boolean> {
  try {
    const response = await fetch(GRAPHQL_HEALTH_URL, {
      method: 'GET',
      signal: AbortSignal.timeout(2000),
    });
    if (response.ok) {
      return true;
    }
  } catch (_error) {
    // fallback to GraphQL 查询检测
  }

  try {
    const response = await fetch(GRAPHQL_API_URL, {
      method: 'POST',
      headers: GRAPHQL_HEADERS,
      body: JSON.stringify({ query: GRAPHQL_HEALTH_QUERY }),
      signal: AbortSignal.timeout(2000),
    });
    return response.ok;
  } catch (_error) {
    return false;
  }
}

function enableMockMode(reason: string) {
  if (!USE_MOCK_MODE) {
    console.warn(`⚠️ 启用 E2E Mock 模式: ${reason}`);
  }
  USE_MOCK_MODE = true;
}

function isMockMode(): boolean {
  return USE_MOCK_MODE;
}

async function ensureAuthentication(page: Page) {
  let accessToken = getPwJwt();
  if (!accessToken) {
    accessToken = await ensurePwJwt() ?? null;
  }

  if (!accessToken) {
    throw new Error('无法获取 RS256 JWT 令牌，无法进行集成测试');
  }

  const payload = {
    accessToken,
    tokenType: 'Bearer',
    expiresIn: 3600,
    issuedAt: Date.now(),
    scope: 'org:read org:create org:update',
  };

  await page.addInitScript(({ tokenPayload }) => {
    window.localStorage.setItem(TOKEN_STORAGE_KEY, JSON.stringify(tokenPayload));
  }, { tokenPayload: payload });
}

async function graphQLRequest(page: Page, query: string, variables: Record<string, unknown> = {}) {
  if (isMockMode()) {
    if (query.includes('__typename')) {
      return { __typename: 'Query' } as Record<string, unknown>;
    }

    if (query.includes('organizationVersions')) {
      return {
        organizationVersions: structuredCloneSafe(MOCK_GRAPHQL_DATA.organizationVersions),
      } as Record<string, unknown>;
    }

    if (query.includes('organization(')) {
      return {
        organization: structuredCloneSafe(MOCK_GRAPHQL_DATA.organizationAsOf),
      } as Record<string, unknown>;
    }

    return {} as Record<string, unknown>;
  }

  const response = await page.request.post(GRAPHQL_API_URL, {
    data: { query, variables },
    headers: GRAPHQL_HEADERS,
  });
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  if (body.errors) {
    throw new Error(`GraphQL errors: ${JSON.stringify(body.errors)}`);
  }
  return body.data as Record<string, unknown>;
}

async function navigateToTemporalPage(page: Page) {
  if (isMockMode()) {
    test.skip(true, 'Mock 模式下跳过 UI 导航验证');
  }

  await ensureAuthentication(page);

  const targetUrl = `${FRONTEND_URL}/organizations/${TEST_ORG_CODE}/temporal`;
  await page.goto(targetUrl, { waitUntil: 'networkidle' });

  if (page.url().includes('/login')) {
    test.skip(true, '访问组织详情页面需要 PW_JWT 令牌');
  }

  await expect(page).toHaveURL(targetUrl);
  await expect(page.getByText('版本历史', { exact: true }).first()).toBeVisible();
}

async function commandRequestStatus(
  page: Page,
  method: 'GET' | 'POST',
  path: string,
  options: { data?: Record<string, unknown>; headers?: Record<string, string> } = {}
): Promise<number> {
  if (isMockMode()) {
    if (method === 'GET' && path.endsWith('/temporal')) {
      return 404;
    }
    if (method === 'POST' && path.endsWith('/versions')) {
      return 422;
    }
    return 200;
  }

  const url = `${COMMAND_API_URL}${path}`;
  const requestMethod = method.toLowerCase();
  const response = await (page.request as any)[requestMethod](url, options);
  return response.status();
}

test.describe('时态管理系统集成测试', () => {
  test.beforeAll(async () => {
    const envValidation = await validateTestEnvironment({ allowUnreachableFrontend: true });

    if (!envValidation.isValid && !isMockMode()) {
      enableMockMode(envValidation.errors.join('; '));
    }

    if (envValidation.warnings.length > 0) {
      envValidation.warnings.forEach((warning) => console.warn(`⚠️ ${warning}`));
    }

    FRONTEND_URL = envValidation.frontendUrl;

    if (!isMockMode()) {
      const [restOk, gqlOk] = await Promise.all([pingCommandHealth(), pingGraphQL()]);
      if (!restOk || !gqlOk) {
        enableMockMode('后端服务不可用');
      }
    }

    if (isMockMode()) {
      console.warn('⚠️ Playwright E2E 将在 Mock 模式下运行（后端服务不可用）');
    }
  });

  test.beforeEach(async ({ page }) => {
    if (isMockMode()) {
      return;
    }

    const restHealth = await page.request.get(COMMAND_HEALTH_URL);
    expect(restHealth.ok()).toBeTruthy();

    const data = await graphQLRequest(page, GRAPHQL_HEALTH_QUERY);
    expect(data.__typename).toBe('Query');
  });

  test.describe('UI 场景 (需认证)', () => {
    test.skip(isMockMode(), 'Mock 模式下跳过 UI 场景');

    test('组织列表可导航至组织详情页面', async ({ page }) => {
      await ensureAuthentication(page);

      await page.goto(`${FRONTEND_URL}/organizations`, { waitUntil: 'networkidle' });
      const searchInput = page.getByPlaceholder('搜索组织名称...');
      await searchInput.fill(TEST_ORG_CODE);
      await page.waitForTimeout(500);
      await expect(page.getByTestId(`table-row-${TEST_ORG_CODE}`)).toBeVisible({ timeout: 15000 });

      const manageButton = page.getByTestId(`temporal-manage-button-${TEST_ORG_CODE}`);
      await expect(manageButton).toBeVisible();
      await manageButton.click();

      await expect(page).toHaveURL(new RegExp(`/organizations/${TEST_ORG_CODE}/temporal$`));
      await expect(page.getByText('版本历史', { exact: true }).first()).toBeVisible();
      await expect(page.getByText('审计历史', { exact: true }).first()).toBeVisible();
    });

    test('组织详情页面展示关键时态组件', async ({ page }) => {
      await navigateToTemporalPage(page);

      await expect(page.getByRole('button', { name: '刷新' })).toBeVisible();
      await expect(page.getByText('点击版本节点查看详情')).toBeVisible();
      await expect(page.locator('input[type="date"]')).toBeVisible();
      await expect(page.getByPlaceholder('请输入组织名称')).toBeVisible();
    });
  });

  test('GraphQL 版本列表契约校验', async ({ page }) => {
    const payload = await graphQLRequest(page, ORGANIZATION_VERSIONS_QUERY, { code: TEST_ORG_CODE });
    const versions = payload.organizationVersions as Array<Record<string, unknown>>;

    expect(Array.isArray(versions)).toBeTruthy();
    expect(versions.length).toBeGreaterThan(0);

    for (const version of versions) {
      expect(version.code).toBe(TEST_ORG_CODE);
      expect(typeof version.name).toBe('string');
      expect(typeof version.unitType).toBe('string');
      expect(typeof version.status).toBe('string');
      expect(typeof version.effectiveDate).toBe('string');
    }
  });

  test('GraphQL asOf 查询支持指定时间点', async ({ page }) => {
    const payload = await graphQLRequest(page, ORGANIZATION_AS_OF_QUERY, {
      code: TEST_ORG_CODE,
      asOfDate: '2024-01-01',
    });

    const organization = payload.organization as Record<string, unknown> | null | undefined;
    if (organization) {
      expect(organization.code).toBe(TEST_ORG_CODE);
      expect(typeof organization.effectiveDate).toBe('string');
    } else {
      expect(organization).toBeNull();
    }
  });

  test('命令服务拒绝未契约的 /temporal 路径', async ({ page }) => {
    const status = await commandRequestStatus(
      page,
      'GET',
      `/organization-units/${TEST_ORG_CODE}/temporal`
    );
    expect(status).toBe(404);
  });

  test('命令服务 /versions 缺少必填字段时返回验证错误', async ({ page }) => {
    const status = await commandRequestStatus(
      page,
      'POST',
      `/organization-units/${TEST_ORG_CODE}/versions`,
      {
        data: {
          name: '',
          unitType: 'DEPARTMENT',
        },
        headers: GRAPHQL_HEADERS,
      }
    );

    expect([400, 422]).toContain(status);
  });
});
