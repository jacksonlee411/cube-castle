import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';
import { setupAuth } from './auth-setup';
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
import { installNetworkCapture } from './utils/networkCapture';

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

test.describe('职位生命周期视图', () => {
  test('展示任职与调动历史', async ({ page }) => {
    const teardownCapture = await installNetworkCapture(page, 'position-lifecycle');
    await stubGraphQL(page);
    await setupAuth(page);

  const positionsQueryReady = waitForGraphQL(page, POSITIONS_QUERY_NAME);
  await page.goto('/positions');
  await positionsQueryReady;
  await waitForPageReady(page);

    await expect(page.getByRole('heading', { name: '职位管理（Stage 1 数据接入）' })).toBeVisible();
    const row = page.getByTestId(temporalEntitySelectors.position.row(POSITION_FIXTURE_CODE));
    await expect(row).toBeVisible();
    await row.click();
    // 等待详情查询与页面容器可见
    await Promise.race([waitForGraphQL(page, POSITION_DETAIL_QUERY_NAME), page.waitForTimeout(500)]);
    await expect(page.getByTestId(temporalEntitySelectors.position.temporalPageWrapper)).toBeVisible();

    const detailCard = page.getByTestId(temporalEntitySelectors.position.detailCard);
    await expect(detailCard).toContainText('生命周期演示岗位');
    await expect(detailCard).toContainText('当前任职');
    await expect(detailCard).toContainText('张三');
    await expect(detailCard).toContainText('夜班负责人');
    await expect(detailCard).toContainText('任职历史');
    await expect(detailCard).toContainText('李四');
    await expect(detailCard).toContainText('调动记录');
    await expect(detailCard).toContainText('ORG-A → ORG-B');
    await expect(detailCard).toContainText('业务线整合');
    await expect(detailCard.getByRole('button', { name: '发起职位转移' })).toBeVisible();

    const vacancyBoard = page.getByTestId('position-vacancy-board');
    await expect(vacancyBoard).toContainText('空缺职位看板');
    await expect(vacancyBoard).toContainText('P-VAC-001');
    await expect(vacancyBoard).toContainText('缺编演示组织');
    await expect(vacancyBoard).toContainText('空缺职位数');

    const headcountDashboard = page.getByTestId('position-headcount-dashboard');
    await expect(headcountDashboard).toContainText('职位编制统计');
    await expect(headcountDashboard).toContainText('总编制');
    await expect(headcountDashboard).toContainText('5');
    await expect(headcountDashboard).toContainText('已占用');
    await expect(headcountDashboard).toContainText('3');
    await expect(page.getByTestId('headcount-level-table')).toContainText('S1');
    await expect(page.getByTestId('headcount-type-table')).toContainText('REGULAR');
    await expect(page.getByTestId('headcount-family-table')).toContainText('OPER-OPS');
    await teardownCapture();
  });
});
