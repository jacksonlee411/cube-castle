import { test, expect, Page } from '@playwright/test'
import path from 'node:path'
import { mkdir } from 'node:fs/promises'
import { TOKEN_STORAGE_KEY } from '@/shared/api/auth'
import { E2E_CONFIG, validateTestEnvironment } from './config/test-environment'
import { updateCachedJwt } from './utils/authToken'

const shouldCapture = process.env.PW_CAPTURE_LAYOUT === 'true'
const OUTPUT_DIR = process.env.PW_CAPTURE_OUTPUT ?? 'artifacts/layout'
const TENANT_ID = process.env.PW_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
const COMMAND_BASE_URL = E2E_CONFIG.COMMAND_API_URL.replace(/\/$/, '').replace(/\/api\/v1$/, '')

type TokenOptions = {
  roles: string[]
  userId: string
  duration?: string
}

const ADMIN_SCOPES = [
  'job-catalog:read',
  'job-catalog:update',
  'job-catalog:write',
  'position:read',
  'position:write',
  'org:read',
  'org:write',
] as const

const captureIfEnabled = async (page: Page, fileName: string) => {
  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(500)
  const filePath = path.join(OUTPUT_DIR, `${fileName}.png`)
  await page.screenshot({ path: filePath, fullPage: true })
  console.info(`布局截图已生成：${filePath}`)
}

const setAuthStorage = async (page: Page, token: string, scopes: readonly string[] = ADMIN_SCOPES) => {
  await page.addInitScript(
    ({ storageKey, accessToken, tenant, scopeList }) => {
      ;(window as typeof window & { __SCOPES__?: string[] }).__SCOPES__ = Array.from(scopeList)
      window.localStorage.setItem(
        storageKey,
        JSON.stringify({
          accessToken,
          tokenType: 'Bearer',
          expiresIn: 8 * 60 * 60,
          issuedAt: Date.now(),
          scope: Array.from(scopeList).join(' '),
        }),
      )
      window.localStorage.setItem('cube-castle-tenant-id', tenant)
    },
    { storageKey: TOKEN_STORAGE_KEY, accessToken: token, tenant: TENANT_ID, scopeList: scopes },
  )
}

const mintToken = async (
  request: import('@playwright/test').APIRequestContext,
  options: TokenOptions,
): Promise<string> => {
  const response = await request.post(`${COMMAND_BASE_URL}/auth/dev-token`, {
    data: {
      tenantId: TENANT_ID,
      roles: options.roles,
      userId: options.userId,
      duration: options.duration ?? '2h',
    },
    headers: { 'Content-Type': 'application/json' },
  })

  expect(response.ok(), '获取开发 JWT 失败').toBeTruthy()
  const json = (await response.json()) as { data?: { token?: string }; token?: string; accessToken?: string }
  const token = json?.data?.token ?? json?.token ?? json?.accessToken
  if (!token) {
    throw new Error('开发令牌响应缺少 token 字段')
  }
  return token
}

const fetchFirstJobFamilyGroupCode = async (
  request: import('@playwright/test').APIRequestContext,
  token: string,
) => {
  const query = `
    query FirstJobFamilyGroup($includeInactive: Boolean) {
      jobFamilyGroups(includeInactive: $includeInactive) {
        code
      }
    }
  `

  const response = await request.post(E2E_CONFIG.GRAPHQL_API_URL, {
    data: { query, variables: { includeInactive: true } },
    headers: {
      Authorization: `Bearer ${token}`,
      'X-Tenant-ID': TENANT_ID,
      'Content-Type': 'application/json',
    },
  })

  expect(response.ok(), 'GraphQL 查询失败').toBeTruthy()
  const json = (await response.json()) as { data?: { jobFamilyGroups?: Array<{ code: string }> } }
  return json.data?.jobFamilyGroups?.[0]?.code ?? null
}

test.describe('Job Catalog 布局基线截图', () => {
  test.describe.configure({ mode: 'serial' })
  test.skip(!shouldCapture, '设置 PW_CAPTURE_LAYOUT=true 以启用布局截图')

  let adminToken: string
  let jobFamilyGroupCode: string | null = null

  test.beforeAll(async ({ request }) => {
    await mkdir(OUTPUT_DIR, { recursive: true })

    const env = await validateTestEnvironment()
    test.skip(!env.isValid, env.errors.join('; '))

    const [commandHealth, graphqlHealth] = await Promise.all([
      request.get(E2E_CONFIG.COMMAND_HEALTH_URL),
      request.get(E2E_CONFIG.GRAPHQL_HEALTH_URL),
    ])

    test.skip(!commandHealth.ok(), `命令服务不可用: ${E2E_CONFIG.COMMAND_HEALTH_URL}`)
    test.skip(!graphqlHealth.ok(), `查询服务不可用: ${E2E_CONFIG.GRAPHQL_HEALTH_URL}`)

    adminToken = await mintToken(request, { roles: ['ADMIN', 'USER'], userId: 'job-catalog-layout' })
    updateCachedJwt(adminToken)

    jobFamilyGroupCode = await fetchFirstJobFamilyGroupCode(request, adminToken)
  })

  test('职位列表页面布局', async ({ page }) => {
    await setAuthStorage(page, adminToken)
    await page.goto('/positions')
    await expect(page.getByTestId('position-dashboard')).toBeVisible()
    await captureIfEnabled(page, 'positions-list')
  })

  test('职类管理列表布局', async ({ page }) => {
    await setAuthStorage(page, adminToken)
    await page.goto('/positions/catalog/family-groups')
    await expect(page.getByRole('heading', { name: '职类管理' })).toBeVisible()
    await captureIfEnabled(page, 'job-family-groups-list')
  })

  test('职类详情布局', async ({ page }) => {
    test.skip(!jobFamilyGroupCode, '暂无职类数据可用于详情截图')
    await setAuthStorage(page, adminToken)
    await page.goto(`/positions/catalog/family-groups/${jobFamilyGroupCode}`)
    await expect(page.getByRole('heading', { name: '职类详情' })).toBeVisible()
    await captureIfEnabled(page, 'job-family-group-detail')
  })
})
