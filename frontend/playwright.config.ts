import { defineConfig, devices } from '@playwright/test';
import { SERVICE_PORTS } from './src/shared/config/ports';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  
  use: {
    baseURL: `http://localhost:${SERVICE_PORTS.FRONTEND_DEV}`, // 使用统一端口配置
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
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

  webServer: {
    command: 'npm run dev',
    url: `http://localhost:${SERVICE_PORTS.FRONTEND_DEV}`, // 使用统一端口配置
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
  },
});