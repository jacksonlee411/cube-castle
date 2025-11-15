import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useTemporalEntityDetail } from '../useTemporalEntityDetail'

// Mock graphql enterprise adapter (organization path)
vi.mock('@/shared/api', async () => {
  return {
    graphqlEnterpriseAdapter: {
      request: vi.fn(),
    },
  }
})

// Mock organization document import to avoid real gql text dependency
vi.mock('@/shared/hooks/useEnterpriseOrganizations', async () => {
  return {
    ORGANIZATION_BY_CODE_DOCUMENT: 'MOCK_ORG_DOC',
  }
})

// Mock position detail hook (position path)
vi.mock('@/shared/hooks/useEnterprisePositions', async () => {
  return {
    usePositionDetail: vi.fn().mockImplementation((_code: string, _opts?: { enabled?: boolean }) => {
      return {
        data: {
          position: {
            code: _code,
            recordId: 'pos-r1',
            title: 'Position A',
            organizationCode: '1000001',
            organizationName: 'Org 1',
            status: 'ACTIVE',
            effectiveDate: '2025-01-01',
            endDate: null,
          },
          versions: [],
          timeline: [],
          assignments: [],
          currentAssignment: null,
          transfers: [],
        },
        isLoading: false,
        isError: false,
        error: null,
        refetch: vi.fn(),
      }
    }),
  }
})

const { graphqlEnterpriseAdapter } = await import('@/shared/api')

describe('useTemporalEntityDetail', () => {
  let client: QueryClient
  const wrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  )

  beforeEach(() => {
    client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    vi.clearAllMocks()
  })

  afterEach(() => {
    client.clear()
  })

  it('returns organization record when GraphQL succeeds', async () => {
    ;(graphqlEnterpriseAdapter.request as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: true,
      data: {
        organization: {
          code: '1000001',
          name: 'Org 1',
          status: 'ACTIVE',
          parentCode: null,
          recordId: 'org-r1',
          effectiveDate: '2024-01-01',
          endDate: null,
        },
      },
      requestId: 'req-1',
      timestamp: new Date().toISOString(),
    })

    const { result } = renderHook(
      () => useTemporalEntityDetail('organization', '1000001', { enabled: true, asOfDate: '2024-12-31' }),
      { wrapper },
    )

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(result.current.data?.record?.entityType).toBe('organization')
    expect(result.current.data?.record?.code).toBe('1000001')
    expect(result.current.data?.record?.status).toBe('ACTIVE')
  })

  it('bubbles organization error when GraphQL fails', async () => {
    ;(graphqlEnterpriseAdapter.request as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: false,
      error: { code: 'INTERNAL', message: 'boom' },
      requestId: 'req-2',
      timestamp: new Date().toISOString(),
    })

    const { result } = renderHook(
      () => useTemporalEntityDetail('organization', '1000002', { enabled: true }),
      { wrapper },
    )

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })
    expect(result.current.error).toBeTruthy()
  })

  it('returns position detail mapped record when position hook provides data', async () => {
    const { result } = renderHook(
      () => useTemporalEntityDetail('position', 'P9000001', { enabled: true, includeDeleted: false }),
      { wrapper },
    )

    await waitFor(() => {
      // position 路径返回的是合并后的数据对象，不一定包含 react-query 的 isSuccess
      expect(result.current.data?.record).toBeTruthy()
    })
    expect(result.current.data?.record?.entityType).toBe('position')
    expect(result.current.data?.record?.code).toBe('P9000001')
    expect(result.current.data?.position?.title).toBe('Position A')
  })
})
