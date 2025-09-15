import { test, expect } from '@playwright/test'

const COMMAND_BASE = 'http://localhost:9090'
const DEFAULT_TENANT = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'

async function mintDevToken() {
  const resp = await fetch(`${COMMAND_BASE}/auth/dev-token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      userId: 'dev-user',
      tenantId: DEFAULT_TENANT,
      roles: ['ADMIN', 'USER'],
      duration: '2h'
    })
  })
  expect(resp.ok).toBeTruthy()
  const body = await resp.json()
  expect(body.success).toBeTruthy()
  const token = body.data?.token as string
  expect(typeof token).toBe('string')
  return token
}

test.describe('Operational Monitoring API', () => {
  test('health/metrics/alerts/rate-limit require auth and tenant', async () => {
    // 401 未认证
    const noAuth = await fetch(`${COMMAND_BASE}/api/v1/operational/metrics`)
    expect(noAuth.status).toBe(401)

    // 生成开发令牌
    const token = await mintDevToken()

    // 403 缺少租户头
    const noTenant = await fetch(`${COMMAND_BASE}/api/v1/operational/metrics`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(noTenant.status === 401 || noTenant.status === 403).toBeTruthy()

    // 200 正常
    const headers = { Authorization: `Bearer ${token}`, 'X-Tenant-ID': DEFAULT_TENANT }

    const healthResp = await fetch(`${COMMAND_BASE}/api/v1/operational/health`, { headers })
    expect(healthResp.ok).toBeTruthy()
    const health = await healthResp.json()
    expect(health.success).toBeTruthy()
    expect(health.data).toHaveProperty('status')
    expect(health.data).toHaveProperty('healthScore')

    const metricsResp = await fetch(`${COMMAND_BASE}/api/v1/operational/metrics`, { headers })
    expect(metricsResp.ok).toBeTruthy()
    const metrics = await metricsResp.json()
    expect(metrics.success).toBeTruthy()
    expect(metrics.data).toHaveProperty('totalOrganizations')
    expect(metrics.data).toHaveProperty('currentRecords')

    const alertsResp = await fetch(`${COMMAND_BASE}/api/v1/operational/alerts`, { headers })
    expect(alertsResp.ok).toBeTruthy()
    const alerts = await alertsResp.json()
    expect(alerts.success).toBeTruthy()
    expect(alerts.data).toHaveProperty('alertCount')

    const rateResp = await fetch(`${COMMAND_BASE}/api/v1/operational/rate-limit/stats`, { headers })
    expect(rateResp.ok).toBeTruthy()
    const rate = await rateResp.json()
    expect(rate.success).toBeTruthy()
    expect(rate.data).toHaveProperty('totalRequests')
  })
})

