import { defineConfig, devices } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { SERVICE_PORTS } from './src/shared/config/ports';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const PROJECT_ROOT = path.resolve(__dirname, '..');
const DEV_JWT_PATH = path.join(PROJECT_ROOT, '.cache', 'dev.jwt');

const sanitizeJwt = (rawToken?: string | null): string | undefined => {
  if (!rawToken) return undefined;
  const trimmed = rawToken.trim();
  if (!trimmed) return undefined;
  const token = trimmed.toLowerCase().startsWith('bearer ') ? trimmed.slice(7).trim() : trimmed;
  if (token.split('.').length !== 3) {
    console.warn('⚠️  PW_JWT 格式无效，请重新生成 dev.jwt');
    return undefined;
  }
  return token;
};

let playwrightJwt = sanitizeJwt(process.env.PW_JWT);

if (!playwrightJwt) {
  try {
    const tokenFromFile = fs.readFileSync(DEV_JWT_PATH, 'utf-8');
    playwrightJwt = sanitizeJwt(tokenFromFile);
  } catch (_error) {
    console.warn('⚠️  PW_JWT 未设置，且未能从 .cache/dev.jwt 读取令牌');
  }
}

if (playwrightJwt) {
  process.env.PW_JWT = playwrightJwt;
} else {
  delete process.env.PW_JWT;
}

// 允许通过环境变量跳过 webServer（前端已在本机运行时避免重复启动）
const SKIP_SERVER = process.env.PW_SKIP_SERVER === '1';
// 允许通过 VITE_PORT_FRONTEND_DEV + PW_BASE_URL 改变端口，避免与现有 dev server 冲突
const FRONTEND_URL = process.env.PW_BASE_URL || `http://localhost:${SERVICE_PORTS.FRONTEND_DEV}`;
const SAVE_HAR = process.env.E2E_SAVE_HAR === '1';
// 行业最佳实践：参数化计划号，统一证据目录命名
// E2E_PLAN_ID 优先：将 HAR/报告落盘到 logs/plan<id>/，例如 E2E_PLAN_ID=254 -> logs/plan254/
// 兼容旧参数 E2E_PLAN=240BT 的 BT/B 后缀策略（不推荐，保留兼容）
const PLAN_ID = process.env.E2E_PLAN_ID && /^\d+$/.test(process.env.E2E_PLAN_ID) ? process.env.E2E_PLAN_ID : '';
const LEGACY_PLAN = (process.env.E2E_PLAN || '').toUpperCase();
const HAR_BASE_DIR = PLAN_ID ? path.resolve(__dirname, '..', 'logs', `plan${PLAN_ID}`) 
  : path.resolve(__dirname, '..', 'logs', 'plan240', (LEGACY_PLAN === '240BT' ? 'BT' : 'B'));
const HAR_DIR = HAR_BASE_DIR;
const TS = Date.now();
if (SAVE_HAR) {
  try {
    fs.mkdirSync(HAR_DIR, { recursive: true });
  } catch {
    /* ignore fs errors */
  }
}

export default defineConfig({
  // 全局测试超时：2分钟
  timeout: 120_000,
  testDir: path.resolve(__dirname, 'tests/e2e'),
  fullyParallel: process.env.E2E_STRICT === '1' ? false : true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.E2E_STRICT === '1' ? 1 : (process.env.CI ? 2 : 4),
  reporter: 'html',
  expect: { timeout: 15_000 },
  
  use: {
    baseURL: FRONTEND_URL, // 允许通过 PW_BASE_URL 覆盖
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 15_000,
    navigationTimeout: 30_000,
    // 按项目生成 HAR（用于 240B 证据）；文件名含时间戳避免覆盖
    ...(SAVE_HAR ? { recordHar: { path: path.join(HAR_DIR, `network-har-default-${TS}.har`), mode: 'minimal' as const } } : {}),
  },

  projects: [
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        ...(SAVE_HAR ? { recordHar: { path: path.join(HAR_DIR, `network-har-chromium-${TS}.har`), mode: 'minimal' as const } } : {}),
      },
    },
    {
      name: 'firefox',
      use: { 
        ...devices['Desktop Firefox'],
        ...(SAVE_HAR ? { recordHar: { path: path.join(HAR_DIR, `network-har-firefox-${TS}.har`), mode: 'minimal' as const } } : {}),
      },
    },
    // Webkit暂时禁用 - WSL环境缺少系统依赖
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },
  ],

  webServer: SKIP_SERVER ? undefined : {
    command: 'npm run dev',
    url: `http://localhost:${SERVICE_PORTS.FRONTEND_DEV}`,
    reuseExistingServer: true,
    // 前端开发服务器启动等待：2分钟
    timeout: 120 * 1000,
  },
});
