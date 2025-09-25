/**
 * 系统监控总览页面端到端验证
 * 目标：验证认证保护、数据渲染、刷新机制与轮询定时器配置
 */
import { test, expect } from '@playwright/test'
import { validateTestEnvironment } from './config/test-environment'
let BASE_URL: string

const DEFAULT_TENANT = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'

function base64UrlEncode(value: Record<string, unknown> | string): string {
  const input = typeof value === 'string' ? value : JSON.stringify(value)
  return Buffer.from(input)
    .toString('base64')
    .replace(/=/g, '')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
}

function mintDevToken(): string {
  const now = Math.floor(Date.now() / 1000)
  const header = base64UrlEncode({ alg: 'RS256', typ: 'JWT', kid: 'e2e-test-key' })
  const payload = base64UrlEncode({
    sub: 'e2e-monitoring-user',
    tenantId: DEFAULT_TENANT,
    exp: now + 2 * 60 * 60,
    iat: now,
    iss: 'cube-castle-dev',
    scope: ['system:monitor:read', 'system:ops:read', 'system:ops:write'],
  })
  const signature = base64UrlEncode('e2e-test-signature')
  return `${header}.${payload}.${signature}`
}

test.describe('系统监控总览 /dashboard', () => {
  test.beforeAll(async () => {
    const envValidation = await validateTestEnvironment()
    if (!envValidation.isValid) {
      envValidation.errors.forEach(error => console.error(`环境验证失败: ${error}`))
      throw new Error('前端服务不可用，终止测试')
    }
    BASE_URL = envValidation.frontendUrl
  })

  test('未认证用户访问会被重定向到登录页', async ({ page }) => {
    await page.goto(`${BASE_URL}/dashboard`)
    await expect(page).toHaveURL(/\/login\?redirect=%2Fdashboard$/)
    await expect(page.getByRole('heading', { name: '登录' })).toBeVisible()
  })

  test('认证后展示监控数据并支持刷新', async ({ page }) => {
    const accessToken = mintDevToken()

    await page.addInitScript(() => {
      const recorded: number[] = []
      const originalSetInterval = window.setInterval.bind(window)
      ;(window as unknown as { __MONITORING_INTERVALS__?: number[] }).__MONITORING_INTERVALS__ = recorded
      window.setInterval = function (handler: TimerHandler, timeout?: number, ...args: unknown[]) {
        if (typeof timeout === 'number') {
          recorded.push(timeout)
        }
        return originalSetInterval(handler, timeout, ...(args as []))
      }
    })

    await page.route('**/.well-known/jwks.json', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          keys: [
            {
              kty: 'RSA',
              kid: 'e2e-test-key',
              use: 'sig',
              alg: 'RS256',
              n: 'test-key-modulus',
              e: 'AQAB',
            },
          ],
        }),
      })
    })

    await page.route('**/auth/dev-token', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            token: accessToken,
            tenantId: DEFAULT_TENANT,
            roles: ['ADMIN'],
            expiresAt: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString(),
          },
        }),
      })
    })

    const mockMetrics = {
      totalOrganizations: 13,
      currentRecords: 40,
      futureRecords: 5,
      historicalRecords: 18,
      duplicateCurrentCount: 0,
      missingCurrentCount: 0,
      timelineOverlapCount: 0,
      inconsistentFlagCount: 0,
      orphanRecordCount: 0,
      healthScore: 96.8,
      alertLevel: 'HEALTHY',
      lastCheckTime: new Date().toISOString(),
    }

    await page.route('**/api/v1/operational/health', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            status: mockMetrics.alertLevel,
            healthScore: mockMetrics.healthScore,
            summary: {
              totalOrganizations: mockMetrics.totalOrganizations,
              currentRecords: mockMetrics.currentRecords,
              futureRecords: mockMetrics.futureRecords,
              historicalRecords: mockMetrics.historicalRecords,
            },
            issues: {
              duplicateCurrentCount: mockMetrics.duplicateCurrentCount,
              missingCurrentCount: mockMetrics.missingCurrentCount,
              timelineOverlapCount: mockMetrics.timelineOverlapCount,
              inconsistentFlagCount: mockMetrics.inconsistentFlagCount,
              orphanRecordCount: mockMetrics.orphanRecordCount,
            },
            lastCheckTime: mockMetrics.lastCheckTime,
          },
        }),
      })
    })

    await page.route('**/api/v1/operational/metrics', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true, data: mockMetrics }),
      })
    })

    await page.route('**/api/v1/operational/alerts', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: { alertCount: 0, alerts: [] },
        }),
      })
    })

    await page.route('**/api/v1/operational/rate-limit/stats', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            totalRequests: 120,
            blockedRequests: 3,
            activeClients: 8,
            lastReset: new Date(Date.now() - 60_000).toISOString(),
            blockRate: '2.50%',
          },
        }),
      })
    })

    await page.goto(`${BASE_URL}/dashboard`, { waitUntil: 'domcontentloaded' })
    await expect(page).toHaveURL(/\/login\?redirect=%2Fdashboard$/)

    await page.getByRole('button', { name: '重新获取开发令牌并继续' }).click()
    await page.waitForURL(/\/dashboard$/, { timeout: 15_000 })
    await expect(page.getByText('系统监控总览')).toBeVisible()
    await expect(page.getByText('健康概览')).toBeVisible()
    await expect(page.getByText('一致性问题')).toBeVisible()
    await expect(page.getByText('告警', { exact: true }).first()).toBeVisible()
    await expect(page.getByText('限流统计')).toBeVisible()

    await expect(page.locator('text=/加载监控数据失败/')).toHaveCount(0)

    const healthCardText = await page.locator('text=/健康分/i').first().textContent()
    expect(healthCardText?.includes('-')).toBeFalsy()

    const intervals = await page.evaluate(() => {
      const win = window as unknown as { __MONITORING_INTERVALS__?: number[] }
      return win.__MONITORING_INTERVALS__ || []
    })
    expect(intervals.some(timeout => timeout >= 29_000 && timeout <= 31_000)).toBeTruthy()

    const timestampLocator = page.locator('text=/最后更新:/')
    await expect(timestampLocator).not.toHaveText(/-\s*$/)

    await page.waitForTimeout(1200)
    const beforeRefresh = await timestampLocator.textContent()

    const refreshDone = page.waitForResponse(response =>
      response.url().endsWith('/api/v1/operational/metrics') && response.status() === 200
    )
    await page.getByRole('button', { name: '刷新' }).click()
    await refreshDone

    await expect(async () => {
      const afterRefresh = await timestampLocator.textContent()
      expect(afterRefresh).not.toBe(beforeRefresh)
    }).toPass({ timeout: 5000 })
  })
})
