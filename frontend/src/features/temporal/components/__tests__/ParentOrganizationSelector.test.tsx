import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { ParentOrganizationSelector } from '../ParentOrganizationSelector'

const graphqlOk = (data: unknown) => ({ ok: true, json: () => Promise.resolve(data) })

function mockFetchWithTokenAndGraphQL(graphqlData: unknown) {
  vi.spyOn(global, 'fetch').mockImplementation((input: any) => {
    const url = typeof input === 'string' ? input : (input?.url || '')
    if (url.includes('/auth/dev-token')) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve({ accessToken: 'test-token', expiresIn: 3600 }) } as any)
    }
    return Promise.resolve(graphqlOk({ data: graphqlData }) as any)
  })
}

describe('ParentOrganizationSelector', () => {
  beforeEach(() => {
    (global as any).__SCOPES__ = ['org:read']
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
    const itemBtn = await screen.findByTestId('combobox-item-1000000#Root')
    fireEvent.click(itemBtn)
    expect(onChange).toHaveBeenCalledWith('1000000')
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

    const itemBtn = await screen.findByTestId('combobox-item-B#B')
    fireEvent.click(itemBtn)
    // 触发循环错误，不调用 onChange
    expect(onChange).not.toHaveBeenCalled()
    expect(onErr).toHaveBeenCalled()
    expect(screen.getByTestId('form-field').getAttribute('data-error') || '').toMatch(/循环依赖/)
  })

  it('gates by PBAC: disabled when missing org:read', async () => {
    ;(global as any).__SCOPES__ = [] // 移除权限
    const organizations = { organizations: { data: [], pagination: { total: 0, page: 1, pageSize: 500 } } }
    mockFetchWithTokenAndGraphQL(organizations)
    render(<ParentOrganizationSelector currentCode="X" effectiveDate="2025-09-17" onChange={() => {}} />)
    // 组件应禁用并显示权限错误
    await waitFor(() => expect(screen.getByTestId('form-field').getAttribute('data-error')).toContain('权限'))
    const input = screen.getByTestId('combobox-input') as HTMLInputElement
    expect(input.disabled).toBe(true)
  })
})
