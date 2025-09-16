import { defineConfig, devices } from '@playwright/test';
import { SERVICE_PORTS } from './src/shared/config/ports';

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
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  expect: { timeout: 120_000 },
  
  use: {
    baseURL: FRONTEND_URL, // 允许通过 PW_BASE_URL 覆盖
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
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
