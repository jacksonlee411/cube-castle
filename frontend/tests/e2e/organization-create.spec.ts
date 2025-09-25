import { test, expect } from '@playwright/test';

type MockOrg = {
  code: string;
  name: string;
  unitType: string;
  parentCode?: string | null;
  level: number;
  effectiveDate: string;
  endDate?: string | null;
  isFuture: boolean;
};

type MockVersion = {
  recordId: string;
  code: string;
  name: string;
  unitType: string;
  status: string;
  level: number;
  path: string;
  effectiveDate: string;
  endDate: string | null;
  createdAt: string;
  updatedAt: string;
  parentCode?: string | null;
  description?: string | null;
};

const parentOrganizations: MockOrg[] = [
  {
    code: '1000000',
    name: '总部',
    unitType: 'DEPARTMENT',
    parentCode: null,
    level: 0,
    effectiveDate: '2024-01-01',
    endDate: null,
    isFuture: false,
  },
  {
    code: '1000001',
    name: '华东分部',
    unitType: 'DEPARTMENT',
    parentCode: '1000000',
    level: 1,
    effectiveDate: '2024-01-01',
    endDate: null,
    isFuture: false,
  },
];

const newOrganizationVersion: MockVersion = {
  recordId: '11111111-1111-1111-1111-111111111111',
  code: '1000999',
  name: '自动化测试组织',
  unitType: 'DEPARTMENT',
  status: 'ACTIVE',
  level: 2,
  path: '1000000/1000999',
  effectiveDate: '2030-01-01',
  endDate: null,
  createdAt: '2030-01-01T00:00:00.000Z',
  updatedAt: '2030-01-01T00:00:00.000Z',
  parentCode: '1000000',
  description: '自动化测试用例创建的组织',
};

const jwksResponse = {
  keys: [
    {
      kty: 'RSA',
      kid: 'test-key',
      alg: 'RS256',
      use: 'sig',
      n: 'test',
      e: 'AQAB',
    },
  ],
};

const rs256DevToken =
  'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.' +
  'eyJzdWIiOiJ0ZXN0IiwiZXhwIjoxODkzNDU2MDAwfQ.' +
  'c2lnbmF0dXJl';

const jsonHeaders = { 'content-type': 'application/json' } as const;

test.describe('Organization Create Flow', () => {
  test('allows selecting parent organization before submitting create request', async ({ page }) => {
    let capturedCreatePayload: Record<string, unknown> | null = null;

    await page.route('**/.well-known/jwks.json', async (route) => {
      await route.fulfill({
        status: 200,
        headers: jsonHeaders,
        body: JSON.stringify(jwksResponse),
      });
    });

    await page.route('**/auth/dev-token', async (route) => {
      await route.fulfill({
        status: 200,
        headers: jsonHeaders,
        body: JSON.stringify({
          accessToken: rs256DevToken,
          tokenType: 'Bearer',
          expiresIn: 3600,
        }),
      });
    });

    await page.route('**/graphql', async (route) => {
      const raw = route.request().postData();
      let query = '';
      if (raw) {
        try {
          const parsed = JSON.parse(raw);
          query = parsed.query ?? '';
        } catch {
          // 忽略解析异常
        }
      }

      const fulfill = (data: unknown) =>
        route.fulfill({
          status: 200,
          headers: jsonHeaders,
          body: JSON.stringify({ data }),
        });

      if (query.includes('GetValidParentOrganizations')) {
        fulfill({
          organizations: {
            data: parentOrganizations,
            pagination: { total: parentOrganizations.length, page: 1, pageSize: 500 },
          },
        });
        return;
      }

      if (query.includes('OrganizationVersions')) {
        fulfill({ organizationVersions: [newOrganizationVersion] });
        return;
      }

      if (query.includes('GetOrganization')) {
        fulfill({
          organization: {
            ...newOrganizationVersion,
            hierarchyDepth: 2,
            isCurrent: true,
          },
        });
        return;
      }

      if (query.includes('organizationHierarchy')) {
        fulfill({
          organizationHierarchy: {
            codePath: '1000000/1000999',
            namePath: '总部/自动化测试组织',
          },
        });
        return;
      }

      fulfill({});
    });

    await page.route('**/api/v1/organization-units', async (route) => {
      if (route.request().method() === 'POST') {
        const raw = route.request().postData();
        capturedCreatePayload = raw ? JSON.parse(raw) : null;
        await route.fulfill({
          status: 200,
          headers: jsonHeaders,
          body: JSON.stringify({ data: { code: '1000999' } }),
        });
        return;
      }

      await route.fulfill({
        status: 200,
        headers: jsonHeaders,
        body: JSON.stringify({ data: {} }),
      });
    });

    await page.route('**/api/v1/**', async (route) => {
      const method = route.request().method();
      if (method === 'GET' || method === 'POST') {
        await route.fulfill({
          status: 200,
          headers: jsonHeaders,
          body: JSON.stringify({ data: {} }),
        });
        return;
      }

      if (method === 'OPTIONS') {
        await route.fulfill({ status: 204, headers: {} });
        return;
      }

      await route.fulfill({ status: 200, headers: jsonHeaders, body: '{}' });
    });

    await page.goto('/organizations/new', { waitUntil: 'domcontentloaded' });

    const reloginButton = page.getByRole('button', { name: '重新获取开发令牌并继续' });
    if (await reloginButton.count()) {
      await reloginButton.click();
      const confirmButton = page.getByRole('button', { name: '继续访问' });
      if (await confirmButton.count()) {
        await confirmButton.click();
      }
      await page.waitForLoadState('networkidle');
    }

    await expect(page.getByText('新建组织 - 编辑组织信息')).toBeVisible();

    const dateInput = page.getByLabel('生效日期 *');
    await dateInput.fill('2030-01-01');

    const nameInput = page.getByPlaceholder('请输入组织名称');
    await nameInput.fill('自动化测试组织');

    const parentInput = page.getByTestId('combobox-input');
    await parentInput.click();

    await expect(page.getByTestId('combobox-item-1000000')).toBeVisible();
    await page.getByTestId('combobox-item-1000000').click();

    await expect(parentInput).toHaveValue('1000000 - 总部');

    await page.getByRole('button', { name: '创建组织' }).click();

    await expect.poll(() => capturedCreatePayload).toBeTruthy();

    expect(capturedCreatePayload).toMatchObject({
      effectiveDate: '2030-01-01',
      name: '自动化测试组织',
      parentCode: '1000000',
    });

    await page.waitForURL('**/organizations/1000999/temporal', { timeout: 10000 });

    await expect(page.getByText('组织详情 - 1000999 自动化测试组织')).toBeVisible();
  });
});
