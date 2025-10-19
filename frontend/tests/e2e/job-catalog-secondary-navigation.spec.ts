import { test, expect } from '@playwright/test'
import { TOKEN_STORAGE_KEY } from '@/shared/api/auth'

type JobFamilyGroup = {
  code: string
  name: string
  status: string
  effectiveDate: string
  endDate: string | null
  description: string | null
  recordId: string
}

const TENANT_ID = 'test-tenant'
const AUTH_TOKEN = 'test-rs256-token'

const fulfillJson = async (route: import('@playwright/test').Route, body: unknown) => {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    headers: {
      'access-control-allow-origin': '*',
    },
    body: JSON.stringify(body),
  })
}

const setupAuthStorage = async (page: import('@playwright/test').Page) => {
  await page.addInitScript(({ tokenStorageKey, token, tenant }) => {
    localStorage.setItem(
      tokenStorageKey,
      JSON.stringify({
        accessToken: token,
        tokenType: 'Bearer',
        expiresIn: 8 * 60 * 60,
        issuedAt: Date.now(),
      }),
    )
    localStorage.setItem('cube-castle-tenant-id', tenant)
  }, {
    tokenStorageKey: TOKEN_STORAGE_KEY,
    token: AUTH_TOKEN,
    tenant: TENANT_ID,
  })
}

const registerJobCatalogMocks = async (page: import('@playwright/test').Page) => {
  const jobFamilyGroups: JobFamilyGroup[] = [
    {
      code: 'PROF',
      name: '专业技术类',
      status: 'ACTIVE',
      effectiveDate: '2025-01-01',
      endDate: null,
      description: '核心职类',
      recordId: 'record-prof',
    },
    {
      code: 'SALE',
      name: '销售序列',
      status: 'ACTIVE',
      effectiveDate: '2025-02-01',
      endDate: null,
      description: '销售团队',
      recordId: 'record-sale',
    },
  ]

  await page.route('**/graphql', async route => {
    const request = route.request()
    if (request.method() !== 'POST') {
      await route.continue()
      return
    }

    let query = ''
    try {
      const payload = JSON.parse(request.postData() ?? '{}') as { query?: string }
      query = payload.query ?? ''
    } catch (_error) {
      await fulfillJson(route, { data: {} })
      return
    }

    if (query.includes('JobFamilyGroups')) {
      await fulfillJson(route, {
        data: {
          jobFamilyGroups,
        },
      })
      return
    }

    await fulfillJson(route, { data: {} })
  })

  await page.route('**/api/v1/job-family-groups/**', async route => {
    const request = route.request()
    const url = new URL(request.url())

    if (request.method() === 'POST' && url.pathname.endsWith('/job-family-groups')) {
      let payload: any = {}
      try {
        payload = JSON.parse(request.postData() ?? '{}')
      } catch (_) {
        /* noop */
      }
      jobFamilyGroups.push({
        code: String(payload.code ?? 'NEW1'),
        name: String(payload.name ?? '新职类'),
        status: String(payload.status ?? 'ACTIVE'),
        effectiveDate: String(payload.effectiveDate ?? '2025-05-01'),
        endDate: payload.endDate ?? null,
        description: payload.description ?? null,
        recordId: `record-${Date.now()}`,
      })
      await fulfillJson(route, { success: true, data: null })
      return
    }

    if (request.method() === 'PUT') {
      const segments = url.pathname.split('/')
      const code = segments[segments.length - 1] ?? ''
      let payload: any = {}
      try {
        payload = JSON.parse(request.postData() ?? '{}')
      } catch (_) {
        /* noop */
      }
      const target = jobFamilyGroups.find(item => item.code === code)
      if (target) {
        target.name = String(payload.name ?? target.name)
        target.status = String(payload.status ?? target.status)
        target.effectiveDate = String(payload.effectiveDate ?? target.effectiveDate)
        target.description = payload.description ?? target.description
      }
      await fulfillJson(route, { success: true, data: null })
      return
    }

    await fulfillJson(route, { success: true, data: null })
  })
}

test.describe('职位管理二级导航（模拟后端）', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthStorage(page)
    await registerJobCatalogMocks(page)
  })

  test('展示职位管理子菜单并加载职类列表', async ({ page }) => {
    await page.goto('/positions')

    const navigationButton = page.getByRole('button', { name: '职位管理' })
    await expect(navigationButton).toBeVisible()

    const familyGroupButton = page.getByRole('button', { name: '职类管理' })
    await familyGroupButton.click()

    await expect(page).toHaveURL(/positions\/catalog\/family-groups/)
    await expect(page.getByRole('heading', { name: '职类管理' })).toBeVisible()
    await expect(page.getByText('专业技术类')).toBeVisible()
    await expect(page.getByText('销售序列')).toBeVisible()

    const searchBox = page.getByPlaceholder('输入关键字搜索')
    await searchBox.fill('财务')
    await expect(page.getByText('暂无数据')).toBeVisible()

    await searchBox.fill('销')
    await expect(page.getByText('销售序列')).toBeVisible()
  })

  test('支持新增与编辑职类（模拟成功响应）', async ({ page }) => {
    await page.goto('/positions/catalog/family-groups')

    await expect(page.getByRole('heading', { name: '职类管理' })).toBeVisible()

    await page.getByRole('button', { name: '新增职类' }).click()
    await expect(page.getByRole('heading', { name: '新增职类' })).toBeVisible()

    await page.getByPlaceholder('例如：PROF').fill('OPER')
    await page.getByPlaceholder('例如：专业技术类').fill('运营管理类')
    await page.getByLabel('生效日期').fill('2025-06-01')
    await page.getByPlaceholder('可选：维护该职类的说明').fill('由 Playwright 创建')

    const createRequestPromise = page.waitForRequest('**/api/v1/job-family-groups')
    await page.getByRole('button', { name: '确认创建' }).click()
    await createRequestPromise

    await expect(page.getByRole('heading', { name: '新增职类' })).not.toBeVisible()
    await expect(page.getByText('运营管理类')).toBeVisible()

    await page.getByText('专业技术类').click()
    await expect(page).toHaveURL(/family-groups\/PROF/)
    await expect(page.getByRole('heading', { name: '职类详情' })).toBeVisible()

    await page.getByRole('button', { name: '编辑当前版本' }).click()
    await expect(page.getByRole('heading', { name: '编辑职类信息' })).toBeVisible()

    const nameInput = page.getByPlaceholder('版本名称')
    await expect(nameInput).toHaveValue('专业技术类')
    await nameInput.fill('专业技术类（更新）')

    const updateRequestPromise = page.waitForRequest('**/api/v1/job-family-groups/PROF')
    await page.getByRole('button', { name: '保存更新' }).click()
    await updateRequestPromise

    await expect(page.getByRole('heading', { name: '编辑职类信息' })).not.toBeVisible()
    await expect(page.getByText('专业技术类（更新）')).toBeVisible()
  })
})
