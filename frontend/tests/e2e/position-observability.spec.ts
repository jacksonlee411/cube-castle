import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';
import { setupAuth } from './auth-setup';
import { v4 as uuidv4 } from 'uuid';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

// Enforce per-test timeout 2 minutes
test.setTimeout(120_000);

// Write logs to repository root logs/plan240/D (single source of truth)
const OBS_DIR = path.resolve(process.cwd(), '..', 'logs', 'plan240', 'D');

const ensureDir = (dir: string) => {
  try {
    fs.mkdirSync(dir, { recursive: true });
  } catch {
    // ignore
  }
};

test.describe('Position Observability (Plan 240D)', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page);
  });

  test('collects [OBS] logs and writes to plan240/D', async ({ page, request }, testInfo) => {
    ensureDir(OBS_DIR);
    const browserName = testInfo.project.name ?? 'browser';
    const logFile = path.join(OBS_DIR, `obs-position-observability-${browserName}.log`);
    const collected: string[] = [];

    page.on('console', (msg) => {
      try {
        const text = msg.text();
        if (text.includes('[OBS] ')) {
          collected.push(text);
        }
      } catch {
        // ignore
      }
    });

    // Support two modes:
    // - If PW_POSITION_CODE is provided, use it and skip creating any data (read-only validation)
    // - Otherwise, create a fresh position and a version to deterministically exercise events
    const providedCode = process.env.PW_POSITION_CODE ?? '';
    const strictVersionFlow = !providedCode;
    let positionCode = providedCode;

    if (!positionCode) {
      const TEST_ID = `OBS-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
      const createPayload = {
        title: `观测用职位-${TEST_ID}`,
        jobFamilyGroupCode: 'OPER',
        jobFamilyCode: 'OPER-OPS',
        jobRoleCode: 'OPER-OPS-MGR',
        jobLevelCode: 'S1',
        organizationCode: '1000000',
        positionType: 'REGULAR',
        employmentType: 'FULL_TIME',
        headcountCapacity: 1.0,
        effectiveDate: '2025-01-01',
        operationReason: `E2E-OBS ${TEST_ID}`,
      };
      const token = await page.evaluate(() => {
        const stored = localStorage.getItem('cubeCastleOauthToken');
        if (!stored) return null;
        try {
          const parsed = JSON.parse(stored);
          return parsed.accessToken as string | null;
        } catch {
          return null;
        }
      });
      if (!token) {
        throw new Error('缺少访问令牌，无法创建观测用职位');
      }
      const tenantId = process.env.PW_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';
      const base = process.env.PW_BASE_URL?.replace(/\/+$/, '') || '';
      const createResp = await request.post(`${base}/api/v1/positions`, {
        headers: {
          Authorization: `Bearer ${token}`,
          'X-Tenant-ID': tenantId,
          'Content-Type': 'application/json',
          'X-Idempotency-Key': `obs-${uuidv4()}`,
        },
        data: createPayload,
      });
      if (createResp.status() === 422) {
        test.skip(true, '职位创建依赖参考数据缺失（Job Catalog/Org），跳过观测验证');
      }
      expect(createResp.status()).toBe(201);
      const created = await createResp.json();
      positionCode = created?.data?.code as string;

      // Create a new version to ensure version.select/export are available
      const versionPayload = {
        ...createPayload,
        title: `观测用职位-版本-${TEST_ID}`,
        effectiveDate: '2025-02-01',
        operationReason: `E2E-OBS version ${TEST_ID}`,
      };
      const versionResp = await request.post(`${base}/api/v1/positions/${positionCode}/versions`, {
        headers: {
          Authorization: `Bearer ${token}`,
          'X-Tenant-ID': tenantId,
          'Content-Type': 'application/json',
          'X-Idempotency-Key': `obs-ver-${uuidv4()}`,
        },
        data: versionPayload,
      });
      expect([201, 409]).toContain(versionResp.status());
    }

    // Navigate to detail to trigger hydration and tab events with fresh data
    await page.goto(`/positions/${positionCode}`);
    // Click to another tab to emit tab.change
    await page.getByTestId(temporalEntitySelectors.position.tabId('timeline')).click({ timeout: 15000 }).catch(() => {});
    // Attempt export to exercise export events (ignore failures)
    // Switch to versions tab first
    await page.getByTestId(temporalEntitySelectors.position.tabId('versions')).click({ timeout: 15000 }).catch(() => {});
    // Select first version row (if present) to emit version.select
    const firstVersionRow = page.locator('[data-testid^="temporal-position-version-row-"]').first();
    if (await firstVersionRow.isVisible().catch(() => false)) {
      await firstVersionRow.click().catch(() => {});
    }
    const exportBtn = page.getByTestId(temporalEntitySelectors.position.versionExportButton || 'temporal-position-version-export-button');
    // Ensure export button is ready before clicking (avoid no-op click)
    try {
      await expect(exportBtn).toBeVisible({ timeout: 10_000 });
      await expect(exportBtn).toBeEnabled({ timeout: 10_000 });
    } catch {
      // ignore: some environments may not have versions yet
    }
    // Concurrently wait for an export.* OBS message to avoid race conditions
    const waitObsExport = page.waitForEvent('console', {
      timeout: 8_000,
      predicate: (msg) => {
        const t = msg.text?.() ?? '';
        return t.includes('[OBS] position.version.export.start') ||
               t.includes('[OBS] position.version.export.done')  ||
               t.includes('[OBS] position.version.export.error');
      },
    }).catch(() => null);
    if (await exportBtn.isVisible().catch(() => false)) {
      await exportBtn.click().catch(() => {});
      // try to await export.* quickly; ignore timeout
      await waitObsExport;
    }

    // Small dwell to allow initial marks/console
    await page.waitForTimeout(500);

    // Persist any collected logs for CI artifact
    try {
      const payload = collected.join('\n') + (collected.length ? '\n' : '');
      fs.writeFileSync(logFile, payload, 'utf8');
    } catch {
      // ignore
    }

    // Basic assertions: at least one hydrate.done appears
    const hasHydrateDone = collected.some(line => line.includes('[OBS] position.hydrate.done'));
    expect(hasHydrateDone, '应捕获到至少一条 [OBS] position.hydrate.done 事件；请确认已设置 VITE_OBS_ENABLED=true 且访问了有效职位详情页（可用 PW_POSITION_CODE 指定）').toBeTruthy();
    // Optional but recommended assertions for this plan:
    const hasTabChange = collected.some(line => line.includes('[OBS] position.tab.change'));
    expect(hasTabChange, '应捕获到 [OBS] position.tab.change 事件').toBeTruthy();
    const hasVersionSelect = collected.some(line => line.includes('[OBS] position.version.select'));
    const hasExport = collected.some(line => /\[OBS] position\.version\.export\.(start|done|error)/.test(line));
    if (strictVersionFlow) {
      expect(hasVersionSelect, '应捕获到 [OBS] position.version.select 事件（已创建新版本并尝试点击）').toBeTruthy();
      expect(hasExport, '应捕获到 [OBS] position.version.export.* 事件（已尝试导出）').toBeTruthy();
    } else {
      // 在 PW_POSITION_CODE 模式下，这两类事件为尽力尝试（不做强制断言）
      console.log(`ℹ️ 使用现有职位 ${positionCode}：version.select=${hasVersionSelect}, export.*=${hasExport}`);
    }
    expect(fs.existsSync(OBS_DIR)).toBeTruthy();
  });
});
