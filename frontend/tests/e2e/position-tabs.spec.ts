import { test, expect, type Page } from '@playwright/test';
import {
  POSITION_FIXTURE_CODE,
  POSITION_GRAPHQL_FIXTURES as GRAPHQL_FIXTURES,
  POSITIONS_QUERY_NAME,
  POSITION_DETAIL_QUERY_NAME,
  VACANT_POSITIONS_QUERY_NAME,
  POSITION_HEADCOUNT_STATS_QUERY_NAME,
} from './utils/positionFixtures';

const FAKE_RS256_JWT = 'eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJwbGF5d3JpZ2h0LXRlc3QifQ.signature';
const POSITION_CODE = POSITION_FIXTURE_CODE;
const POSITION_DETAIL_PATH = `/positions/${POSITION_CODE}`;

async function stubGraphQL(page: Page) {
  await page.route('**/graphql', async route => {
    const request = route.request();
    let body: { query?: string } | undefined;
    try {
      body = request.postDataJSON();
    } catch (_error) {
      return route.continue();
    }

    const query = body?.query;
    if (!query) {
      return route.continue();
    }

    if (query.includes(POSITIONS_QUERY_NAME)) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(GRAPHQL_FIXTURES.positions),
      });
    }

    if (query.includes(POSITION_DETAIL_QUERY_NAME)) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(GRAPHQL_FIXTURES.positionDetail),
      });
    }

    if (query.includes(VACANT_POSITIONS_QUERY_NAME)) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(GRAPHQL_FIXTURES.vacantPositions),
      });
    }

    if (query.includes(POSITION_HEADCOUNT_STATS_QUERY_NAME)) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(GRAPHQL_FIXTURES.headcountStats),
      });
    }

    return route.continue();
  });
}

async function seedAuth(page: Page) {
  await page.route('**/.well-known/jwks.json', route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ keys: [{ kid: 'test-key', kty: 'RSA', use: 'sig' }] }),
      });
    });

  await page.addInitScript(({ token }) => {
      const issuedAt = Date.now();
      window.localStorage.setItem(
        'cubeCastleOauthToken',
        JSON.stringify({
          accessToken: token,
          tokenType: 'Bearer',
          expiresIn: 8 * 60 * 60,
          issuedAt,
        }),
      );
      window.localStorage.setItem('cube-castle-tenant-id', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9');
    }, { token: FAKE_RS256_JWT });
}

test.describe('职位详情多页签体验', () => {
  test('六个页签可切换且展示对应内容', async ({ page }) => {
    await seedAuth(page);
    await stubGraphQL(page);

    await page.goto(POSITION_DETAIL_PATH);

    await expect(page.getByTestId('position-temporal-page')).toBeVisible();
    await expect(page.getByText(`职位详情：${POSITION_CODE}`)).toBeVisible();

    // 概览页签默认展示
    await expect(page.getByTestId('position-overview-card')).toContainText(`职位编码：${POSITION_CODE}`);

    const clickTab = async (label: string) => {
      await page.getByText(label, { exact: true }).click();
    };

    await clickTab('任职记录');
    await expect(page.getByText('任职历史')).toBeVisible();
    await expect(page.getByText('张三')).toBeVisible();

    await clickTab('调动记录');
    await expect(page.getByText('调动记录')).toBeVisible();
    await expect(page.getByText('业务线整合')).toBeVisible();

    await clickTab('时间线');
    await expect(page.getByText('时间线事件')).toBeVisible();
    await expect(page.getByText('岗位创建')).toBeVisible();

    await clickTab('版本历史');
    await expect(page.getByTestId('position-version-toolbar')).toBeVisible();
    await expect(page.getByTestId('position-version-list')).toBeVisible();

    await clickTab('审计历史');
    await expect(page.getByText('当前版本缺少 recordId，无法加载审计历史。')).toBeVisible();
  });

  test('Mock 模式下隐藏写入按钮', async ({ page }) => {
    test.skip(process.env.VITE_POSITIONS_MOCK_MODE !== 'true', '需在 VITE_POSITIONS_MOCK_MODE=true 环境下运行');

    await seedAuth(page);
    await stubGraphQL(page);
    await page.goto(POSITION_DETAIL_PATH);

    await expect(page.getByTestId('position-temporal-page')).toBeVisible();
    await expect(page.getByTestId('position-mock-banner')).toBeVisible();
    await expect(page.getByTestId('position-edit-button')).toBeHidden();
    await expect(page.getByTestId('position-version-button')).toBeHidden();
  });
});
