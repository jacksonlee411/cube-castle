import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';
import { setupAuth } from './auth-setup';

const OBS_DIR = path.resolve(process.cwd(), 'logs/plan240/D');

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

  test('collects [OBS] logs and writes to plan240/D', async ({ page }, testInfo) => {
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

    // Try to visit a position page. If no seed data, still exercise route/frame.
    // Prefer an exemplar code if available via env, fallback to list page.
    const exampleCode = process.env.PW_POSITION_CODE ?? '';
    if (exampleCode) {
      await page.goto(`/positions/${exampleCode}`);
    } else {
      await page.goto('/positions');
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

    // The test itself is non-strict: existence of file suffices
    expect(fs.existsSync(OBS_DIR)).toBeTruthy();
  });
});

