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
const FRONTEND_URL = process.env.PW_BASE_URL || `http://localhost:${SERVICE_PORTS.FRONTEND_DEV}`;

export default defineConfig({
  // 全局测试超时：2分钟
  timeout: 120_000,
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 2 : 4,
  reporter: 'html',
  expect: { timeout: 10_000 },
  
  use: {
    baseURL: FRONTEND_URL, // 允许通过 PW_BASE_URL 覆盖
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 15_000,
    navigationTimeout: 30_000,
    // 为所有请求注入认证头（后端强制要求）
    extraHTTPHeaders: {
      'Authorization': process.env.PW_JWT ? `Bearer ${process.env.PW_JWT}` : '',
      'X-Tenant-ID': process.env.PW_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    },
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
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
    reuseExistingServer: !process.env.CI,
    // 前端开发服务器启动等待：2分钟
    timeout: 120 * 1000,
  },
});
