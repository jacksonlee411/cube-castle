import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, act } from '@testing-library/react'
import {
  useUpdateJobFamilyGroup,
  useUpdateJobFamily,
  useUpdateJobRole,
  useUpdateJobLevel,
  type UpdateJobFamilyGroupInput,
  type UpdateJobFamilyInput,
  type UpdateJobRoleInput,
  type UpdateJobLevelInput,
} from '../useJobCatalogMutations'
import { JobCatalogStatus } from '@/generated/graphql-types'

const mocks = vi.hoisted(() => ({
  requestMock: vi.fn(),
  loggerMock: {
    mutation: vi.fn(),
    error: vi.fn(),
  },
}))

vi.mock('@/shared/api', () => ({
  unifiedRESTClient: {
    request: mocks.requestMock,
  },
}))

vi.mock('@/shared/utils/logger', () => ({
  logger: mocks.loggerMock,
}))

const createClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: Infinity,
      },
      mutations: {
        retry: false,
      },
    },
  })

const createWrapper = (client: QueryClient) => {
  const Wrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  )
  return Wrapper
}

beforeEach(() => {
  mocks.requestMock.mockReset()
  mocks.loggerMock.mutation.mockClear()
  mocks.loggerMock.error.mockClear()
})

afterEach(() => {
  vi.clearAllTimers()
})

describe('useJobCatalogMutations REST integration', () => {
  it('updates job family group via REST endpoint and invalidates cache', async () => {
    const queryClient = createClient()
    try {
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
      const wrapper = createWrapper(queryClient)
      mocks.requestMock.mockResolvedValue({ success: true, data: null })

      const { result } = renderHook(() => useUpdateJobFamilyGroup(), { wrapper })

      const input: UpdateJobFamilyGroupInput = {
        code: 'PROF',
        recordId: 'rec-prof',
        name: '专业技术类',
        status: JobCatalogStatus.ACTIVE,
        effectiveDate: '2025-01-01',
        description: '描述',
      }

      await act(async () => {
        await result.current.mutateAsync(input)
      })

      expect(mocks.requestMock).toHaveBeenCalledWith('/job-family-groups/PROF', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'If-Match': 'rec-prof' },
        body: JSON.stringify({
          name: '专业技术类',
          status: JobCatalogStatus.ACTIVE,
          effectiveDate: '2025-01-01',
          description: '描述',
        }),
      })
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['jobCatalog', 'groups'], exact: false })
    } finally {
      queryClient.clear()
    }
  })

  it('updates job family and retains parent group reference', async () => {
    const queryClient = createClient()
    try {
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
      const wrapper = createWrapper(queryClient)
      mocks.requestMock.mockResolvedValue({ success: true, data: null })

      const { result } = renderHook(() => useUpdateJobFamily(), { wrapper })

      const input: UpdateJobFamilyInput = {
        code: 'PROF-SALES',
        recordId: 'rec-family',
        jobFamilyGroupCode: 'PROF',
        name: '销售序列',
        status: JobCatalogStatus.ACTIVE,
        effectiveDate: '2025-01-10',
        description: '描述',
      }

      await act(async () => {
        await result.current.mutateAsync(input)
      })

      expect(mocks.requestMock).toHaveBeenCalledWith('/job-families/PROF-SALES', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'If-Match': 'rec-family' },
        body: JSON.stringify({
          jobFamilyGroupCode: 'PROF',
          name: '销售序列',
          status: JobCatalogStatus.ACTIVE,
          effectiveDate: '2025-01-10',
          description: '描述',
        }),
      })
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['jobCatalog', 'families'], exact: false })
    } finally {
      queryClient.clear()
    }
  })

  it('updates job role with derived family code', async () => {
    const queryClient = createClient()
    try {
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
      const wrapper = createWrapper(queryClient)
      mocks.requestMock.mockResolvedValue({ success: true, data: null })

      const { result } = renderHook(() => useUpdateJobRole(), { wrapper })

      const input: UpdateJobRoleInput = {
        code: 'PROF-SALES-MGR',
        recordId: 'rec-role',
        jobFamilyCode: 'PROF-SALES',
        name: '销售经理',
        status: JobCatalogStatus.ACTIVE,
        effectiveDate: '2025-02-01',
        description: '负责销售',
      }

      await act(async () => {
        await result.current.mutateAsync(input)
      })

      expect(mocks.requestMock).toHaveBeenCalledWith('/job-roles/PROF-SALES-MGR', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'If-Match': 'rec-role' },
        body: JSON.stringify({
          jobFamilyCode: 'PROF-SALES',
          name: '销售经理',
          status: JobCatalogStatus.ACTIVE,
          effectiveDate: '2025-02-01',
          description: '负责销售',
        }),
      })
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['jobCatalog', 'roles'], exact: false })
    } finally {
      queryClient.clear()
    }
  })

  it('updates job level with role linkage and rank persistence', async () => {
    const queryClient = createClient()
    try {
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
      const wrapper = createWrapper(queryClient)
      mocks.requestMock.mockResolvedValue({ success: true, data: null })

      const { result } = renderHook(() => useUpdateJobLevel(), { wrapper })

      const input: UpdateJobLevelInput = {
        code: 'PROF-SALES-MGR-L3',
        recordId: 'rec-level',
        jobRoleCode: 'PROF-SALES-MGR',
        name: '高级销售经理',
        status: JobCatalogStatus.ACTIVE,
        effectiveDate: '2025-03-01',
        description: '关键岗位',
        levelRank: 3,
      }

      await act(async () => {
        await result.current.mutateAsync(input)
      })

      expect(mocks.requestMock).toHaveBeenCalledWith('/job-levels/PROF-SALES-MGR-L3', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'If-Match': 'rec-level' },
        body: JSON.stringify({
          jobRoleCode: 'PROF-SALES-MGR',
          name: '高级销售经理',
          status: JobCatalogStatus.ACTIVE,
          effectiveDate: '2025-03-01',
          description: '关键岗位',
          levelRank: 3,
        }),
      })
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['jobCatalog', 'levels'], exact: false })
    } finally {
      queryClient.clear()
    }
  })

  it('surfaces REST validation errors through createQueryError', async () => {
    const queryClient = createClient()
    try {
      const wrapper = createWrapper(queryClient)
      mocks.requestMock.mockResolvedValue({
        success: false,
        error: {
          message: '编码重复',
          code: 'VALIDATION_ERROR',
          details: { field: 'code' },
        },
        requestId: 'req-1',
      })

      const { result } = renderHook(() => useUpdateJobFamilyGroup(), { wrapper })

      const input: UpdateJobFamilyGroupInput = {
        code: 'DUPL',
        recordId: 'rec-dupl',
        name: '重复编码',
        status: JobCatalogStatus.ACTIVE,
        effectiveDate: '2025-01-01',
      }

      await act(async () => {
        await expect(result.current.mutateAsync(input)).rejects.toMatchObject({
          message: '编码重复',
          code: 'VALIDATION_ERROR',
          requestId: 'req-1',
        })
      })
    } finally {
      queryClient.clear()
    }
  })
})
