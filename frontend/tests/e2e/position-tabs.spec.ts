import { test, expect, type Page } from '@playwright/test';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';
import {
  POSITION_FIXTURE_CODE,
  POSITION_GRAPHQL_FIXTURES as GRAPHQL_FIXTURES,
  POSITIONS_QUERY_NAME,
  POSITION_DETAIL_QUERY_NAME,
  VACANT_POSITIONS_QUERY_NAME,
  POSITION_HEADCOUNT_STATS_QUERY_NAME,
} from './utils/positionFixtures';
import { waitForGraphQL, waitForPageReady } from './utils/waitPatterns';
const POSITION_ASSIGNMENTS_QUERY_NAME = 'PositionAssignments';

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

    if (query.includes(POSITION_ASSIGNMENTS_QUERY_NAME)) {
      // 返回一个最小的 assignments 数据用于渲染
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            positionAssignments: {
              data: GRAPHQL_FIXTURES.positionDetail.data.positionAssignments.data,
              pagination: {
                total: 2,
                page: 1,
                pageSize: 50,
                hasNext: false,
                hasPrevious: false,
              },
              totalCount: 2,
            },
          },
        }),
      });
    }

    if (query.includes('auditHistory(') || query.includes('TemporalEntityAuditHistory')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { auditHistory: [] } }),
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

  // 拦截 /auth/dev-token，避免真实令牌获取带来的不确定等待
  await page.route('**/auth/dev-token', async route => {
    if (route.request().method() !== 'POST') return route.continue();
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        accessToken: FAKE_RS256_JWT,
        tokenType: 'Bearer',
        expiresIn: 3600,
      }),
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
    // 网络抓取可选（在 CI 中通过 HAR 记录）
    await seedAuth(page);
    await stubGraphQL(page);

    await page.goto(POSITION_DETAIL_PATH);
    // 等待 GraphQL 详情查询返回，确保详情容器可渲染
    await waitForGraphQL(page, POSITION_DETAIL_QUERY_NAME).catch(() => {});
    await waitForPageReady(page);
    // 如应用触发令牌获取，先等待 dev-token 返回（无则忽略）
    await page.waitForResponse(r => r.url().includes('/auth/dev-token'), { timeout: 5000 }).catch(() => {});

    // DOM 就绪断言（使用统一 SSoT 选择器，避免文本渲染时序/字体影响）
    await expect(page.getByTestId(temporalEntitySelectors.position.temporalPageWrapper)).toBeVisible();
    await expect(page.getByTestId(temporalEntitySelectors.position.temporalPage)).toBeVisible();

    // 概览页签默认展示（以 overviewCard 的可见性作为断言，不依赖具体文案）
    await expect(page.getByTestId(temporalEntitySelectors.position.overviewCard)).toBeVisible();

    const clickTabByKey = async (key: string) => {
      const tab = page.getByTestId(temporalEntitySelectors.position.tabId(key));
      await tab.click();
      await expect(tab).toHaveAttribute('aria-selected', 'true');
    };

    // 任职记录
    await clickTabByKey('assignments');
    await expect(page.getByText('导出 CSV')).toBeVisible({ timeout: 10000 });

    // 调动记录
    await clickTabByKey('transfers');
    await expect(page.getByRole('heading', { name: '调动记录' })).toBeVisible({ timeout: 10000 });

    // 时间线
    await clickTabByKey('timeline');
    await expect(page.getByRole('heading', { name: '时间线事件' })).toBeVisible({ timeout: 10000 });

    // 版本历史
    await clickTabByKey('versions');
    await expect(page.getByTestId(temporalEntitySelectors.position.versionToolbar)).toBeVisible({ timeout: 10000 });
    await expect(page.getByTestId(temporalEntitySelectors.position.versionList)).toBeVisible({ timeout: 10000 });

    await clickTabByKey('audit');
    await expect(page.getByRole('heading', { name: '审计历史' })).toBeVisible({ timeout: 10000 });
    // 网络请求计数另由 HAR/CI 工具汇总
  });

  test('Mock 模式下隐藏写入按钮', async ({ page }) => {
    test.skip(process.env.VITE_POSITIONS_MOCK_MODE !== 'true', '需在 VITE_POSITIONS_MOCK_MODE=true 环境下运行');

    await seedAuth(page);
    await stubGraphQL(page);
    await page.goto(POSITION_DETAIL_PATH);
    await waitForPageReady(page);
    await waitForGraphQL(page, POSITION_DETAIL_QUERY_NAME);

    await expect(page.getByTestId(temporalEntitySelectors.position.temporalPageWrapper)).toBeVisible();
    await expect(page.getByTestId(temporalEntitySelectors.position.mockBanner!)).toBeVisible();
    await expect(page.getByTestId(temporalEntitySelectors.position.editButton!)).toBeHidden();
    await expect(page.getByTestId(temporalEntitySelectors.position.createVersionButton!)).toBeHidden();
  });
});
