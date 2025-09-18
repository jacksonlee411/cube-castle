import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { ParentOrganizationSelector } from '../ParentOrganizationSelector'
import { authManager } from '../../../../shared/api/auth'

function mockFetchWithTokenAndGraphQL(graphqlData: unknown) {
  vi.spyOn(global, 'fetch').mockImplementation((input: any) => {
    const url = typeof input === 'string' ? input : (input?.url || '')
    if (url.includes('/.well-known/jwks.json')) {
      return Promise.resolve(
        new Response(
          JSON.stringify({ keys: [{ kty: 'RSA', kid: 'test', alg: 'RS256', n: 'test', e: 'AQAB' }] }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        )
      )
    }
    if (url.includes('/auth/dev-token')) {
      return Promise.resolve(
        new Response(
          JSON.stringify({ accessToken: 'test-token', expiresIn: 3600 }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        )
      )
    }
    return Promise.resolve(
      new Response(
        JSON.stringify({ data: graphqlData }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      )
    )
  })
}

describe('ParentOrganizationSelector', () => {
  beforeEach(() => {
    (global as any).__SCOPES__ = ['org:read']
    vi.spyOn(authManager, 'getAccessToken').mockResolvedValue('test-token')
  })
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('loads candidates and allows selection (calls onChange with code)', async () => {
    const organizations = {
      organizations: {
        data: [
          { code: '1000000', name: 'Root', unitType: 'DEPARTMENT', parentCode: '0', level: 0, effectiveDate: '2025-01-01', endDate: null, isFuture: false },
          { code: '1000001', name: 'Self', unitType: 'DEPARTMENT', parentCode: '1000000', level: 1, effectiveDate: '2025-01-01', endDate: null, isFuture: false },
          { code: '1000002', name: 'Dept Two', unitType: 'DEPARTMENT', parentCode: '1000001', level: 2, effectiveDate: '2025-01-01', endDate: null, isFuture: false }
        ],
        pagination: { total: 2, page: 1, pageSize: 500 }
      }
    }
    mockFetchWithTokenAndGraphQL(organizations)

    const onChange = vi.fn()
    render(<ParentOrganizationSelector currentCode="1000001" effectiveDate="2025-09-15" onChange={onChange} />)

    // 等待加载结束并渲染候选项（自组织会被过滤）
    const itemBtn = await screen.findByTestId('combobox-item-1000000')
    fireEvent.click(itemBtn)
    await waitFor(() => expect(onChange).toHaveBeenCalledWith('1000000'))
  })

  it('detects cycle and reports error via onValidationError', async () => {
    const organizations = {
      organizations: {
        data: [
          { code: 'A', name: 'A', unitType: 'DEPARTMENT', parentCode: '0', level: 1, effectiveDate: '2025-01-01', endDate: null, isFuture: false },
          { code: 'B', name: 'B', unitType: 'DEPARTMENT', parentCode: 'A', level: 2, effectiveDate: '2025-01-01', endDate: null, isFuture: false }
        ],
        pagination: { total: 2, page: 1, pageSize: 500 }
      }
    }
    mockFetchWithTokenAndGraphQL(organizations)
    const onErr = vi.fn()
    const onChange = vi.fn()
    render(<ParentOrganizationSelector currentCode="A" effectiveDate="2025-09-16" onChange={onChange} onValidationError={onErr} />)

    const itemBtn = await screen.findByTestId('combobox-item-B')
    fireEvent.click(itemBtn)
    await waitFor(() => expect(onChange).not.toHaveBeenCalled())
    await waitFor(() => expect(onErr).toHaveBeenCalled())
    expect(screen.getByText(/循环依赖/)).toBeTruthy()
  })

  it('gates by PBAC: disabled when missing org:read', async () => {
    ;(global as any).__SCOPES__ = [] // 移除权限
    const organizations = { organizations: { data: [], pagination: { total: 0, page: 1, pageSize: 500 } } }
    mockFetchWithTokenAndGraphQL(organizations)
    render(<ParentOrganizationSelector currentCode="X" effectiveDate="2025-09-17" onChange={() => {}} />)
    // 组件应禁用并显示权限错误
    const errors = await screen.findAllByText('您没有权限查看组织列表')
    expect(errors.length).toBeGreaterThan(0)
    const input = screen.getByTestId('combobox-input') as HTMLInputElement
    expect(input.disabled).toBe(true)
  })
})
