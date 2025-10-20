import { test, expect } from '@playwright/test';

const shouldRunLiveSuite = process.env.CI === 'true' || process.env.PW_REQUIRE_LIVE_BACKEND === '1';
const shouldRunMockGuard = process.env.PW_REQUIRE_MOCK_CHECK === '1';

test.describe('职位管理 CRUD（真实后端链路）', () => {
  test.skip(!shouldRunLiveSuite, '未启用真实后端联调（设置 PW_REQUIRE_LIVE_BACKEND=1 或在 CI 环境运行）');

  test('加载职位列表并跳转详情', async ({ page }) => {
    const jwtToken = process.env.PW_JWT;
    expect(jwtToken, 'PW_JWT 未设置，无法验证真实后端链路').toBeTruthy();

    const tenantId = process.env.PW_TENANT_ID ?? '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';

    await page.addInitScript(({ token, tenant }) => {
      const issuedAt = Math.floor(Date.now() / 1000);
      window.localStorage.setItem(
        'cubeCastleOauthToken',
        JSON.stringify({
          accessToken: token,
          tokenType: 'Bearer',
          expiresIn: 8 * 60 * 60,
          issuedAt,
        }),
      );
      window.localStorage.setItem('cube-castle-tenant-id', tenant);
    }, { token: jwtToken, tenant: tenantId });

    await page.goto('/positions');

    const loginHeading = page.getByRole('heading', { name: '登录' });
    if (await loginHeading.isVisible()) {
      await page.getByRole('button', { name: '重新获取开发令牌并继续' }).click();
      await page.waitForURL('**/positions');
    }

    await expect(page.getByTestId('position-dashboard')).toBeVisible();

    const positionsResponse = await page.waitForResponse(response => {
      if (!response.url().includes('/graphql')) return false;
      const requestBody = response.request().postData();
      if (!requestBody || !requestBody.includes('EnterprisePositions')) return false;
      return response.status() === 200;
    });
    const positionsPayload = await positionsResponse.json();

    expect(Array.isArray(positionsPayload.data.positions.data)).toBe(true);

    await expect(page.getByText(/GraphQL \/ REST 实时数据/)).toBeVisible();

    const firstRow = page.locator('[data-testid^="position-row-"]').first();
    await expect(firstRow).toBeVisible();
    const dataTestId = (await firstRow.getAttribute('data-testid')) ?? '';
    expect(dataTestId.startsWith('position-row-')).toBe(true);
    const positionCode = dataTestId.replace('position-row-', '');
    expect(positionCode.length).toBeGreaterThan(0);

    await firstRow.click();
    await page.waitForURL(url => url.pathname.includes(`/positions/${positionCode}`));
    await expect(page.getByTestId('position-temporal-page')).toBeVisible();
    await expect(page.getByText(`职位详情：${positionCode}`)).toBeVisible();
    await expect(page.getByTestId('position-detail-card')).toBeVisible();
    await expect(page.getByTestId('position-version-list')).toBeVisible();
  });
});

test.describe('职位管理 Mock 守护', () => {
  test.skip(!shouldRunMockGuard, '未启用 Mock 守护检查（设置 PW_REQUIRE_MOCK_CHECK=1）');

  test('Mock 模式显示只读提示并禁用创建按钮', async ({ page }) => {
    await page.goto('/positions');

    await expect(page.getByTestId('position-dashboard')).toBeVisible();
    await expect(page.getByTestId('position-dashboard-mock-banner')).toBeVisible();

    const createButton = page.getByTestId('position-create-button');
    await expect(createButton).toBeDisabled();
  });
});
