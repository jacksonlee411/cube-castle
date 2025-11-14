import { test, expect } from "@playwright/test";
import temporalEntitySelectors from "@/shared/testids/temporalEntity";
import { ensurePwJwt } from "./utils/authToken";

const shouldRunLiveSuite =
  process.env.CI === "true" || process.env.PW_REQUIRE_LIVE_BACKEND === "1";
const shouldRunMockGuard = process.env.PW_REQUIRE_MOCK_CHECK === "1";

test.describe("职位管理 CRUD（真实后端链路）", () => {
  test.skip(
    !shouldRunLiveSuite,
    "未启用真实后端联调（设置 PW_REQUIRE_LIVE_BACKEND=1 或在 CI 环境运行）",
  );

  test("加载职位列表并跳转详情", async ({ page }) => {
    const tenantId =
      process.env.PW_TENANT_ID ?? "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9";

    const jwtToken = await ensurePwJwt({ tenantId });
    expect(
      jwtToken,
      "无法获取开发 JWT，请先运行 make jwt-dev-mint",
    ).toBeTruthy();

    await page.addInitScript(
      ({ token, tenant }) => {
        const issuedAt = Date.now();
        window.localStorage.setItem(
          "cubeCastleOauthToken",
          JSON.stringify({
            accessToken: token,
            tokenType: "Bearer",
            expiresIn: 8 * 60 * 60,
            issuedAt,
          }),
        );
        window.localStorage.setItem("cube-castle-tenant-id", tenant);
      },
      { token: jwtToken, tenant: tenantId },
    );

    const devTokenEndpoint = "**/auth/dev-token";
    await page.route(devTokenEndpoint, (route) => {
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ token: jwtToken }),
      });
    });

    await page.goto("/positions");

    const loginHeading = page.getByRole("heading", { name: "登录" });
    if (await loginHeading.isVisible()) {
      await page
        .getByRole("button", { name: "重新获取开发令牌并继续" })
        .click();
      await page.waitForURL("**/positions", { waitUntil: "networkidle" });
    }

    await page.unroute(devTokenEndpoint);

    await page
      .getByTestId(temporalEntitySelectors.position.dashboard)
      .waitFor({ state: "visible" });

    await expect(
      page.getByText(/当前页面依赖 GraphQL 查询服务与 REST 命令服务/),
    ).toBeVisible();

    const firstRow = page
      .locator(`[data-testid^="${temporalEntitySelectors.position.rowPrefix}"]`)
      .first();
    await expect(firstRow).toBeVisible();
    const dataTestId = (await firstRow.getAttribute("data-testid")) ?? "";
    const prefix = temporalEntitySelectors.position.rowPrefix;
    expect(dataTestId.startsWith(prefix)).toBe(true);
    const positionCode = dataTestId.replace(prefix, "");
    expect(positionCode.length).toBeGreaterThan(0);

    await firstRow.click();
    await page.waitForURL((url) =>
      url.pathname.includes(`/positions/${positionCode}`),
    );
    await expect(page.getByTestId(temporalEntitySelectors.position.temporalPage)).toBeVisible();
    await expect(page.getByText(`职位详情：${positionCode}`)).toBeVisible();
  });
});

test.describe("职位管理 Mock 守护", () => {
  test.skip(
    !shouldRunMockGuard,
    "未启用 Mock 守护检查（设置 PW_REQUIRE_MOCK_CHECK=1）",
  );

  test("Mock 模式显示只读提示并禁用创建按钮", async ({ page }) => {
    await page.goto("/positions");

    await expect(
      page.getByTestId(temporalEntitySelectors.position.dashboard),
    ).toBeVisible();
    await expect(
      page.getByTestId("position-dashboard-mock-banner"),
    ).toBeVisible();

    const createButton = page.getByTestId("position-create-button");
    await expect(createButton).toBeDisabled();
  });
});
