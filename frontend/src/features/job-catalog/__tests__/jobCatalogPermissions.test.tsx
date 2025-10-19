// @vitest-environment jsdom
import React from 'react'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, it, expect, beforeEach, beforeAll, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

const mocks = vi.hoisted(() => ({
  hasPermission: vi.fn(),
  mockUseJobFamilyGroups: vi.fn(),
  mockUseJobFamilies: vi.fn(),
  mockUseJobRoles: vi.fn(),
  mockUseJobLevels: vi.fn(),
}))

let JobFamilyGroupList: React.ComponentType
let JobFamilyGroupDetail: React.ComponentType
let JobFamilyDetail: React.ComponentType
let JobRoleDetail: React.ComponentType
let JobLevelDetail: React.ComponentType

vi.mock('@/shared/auth/hooks', () => ({
  useAuth: () => ({
    hasPermission: mocks.hasPermission,
  }),
}))

vi.mock('@/shared/hooks/useJobCatalog', () => ({
  useJobFamilyGroups: (...args: unknown[]) => mocks.mockUseJobFamilyGroups(...args),
  useJobFamilies: (...args: unknown[]) => mocks.mockUseJobFamilies(...args),
  useJobRoles: (...args: unknown[]) => mocks.mockUseJobRoles(...args),
  useJobLevels: (...args: unknown[]) => mocks.mockUseJobLevels(...args),
}))

beforeAll(async () => {
  const listModule = await import('../family-groups/JobFamilyGroupList')
  const groupDetailModule = await import('../family-groups/JobFamilyGroupDetail')
  const familyDetailModule = await import('../families/JobFamilyDetail')
  const roleDetailModule = await import('../roles/JobRoleDetail')
  const levelDetailModule = await import('../levels/JobLevelDetail')

  JobFamilyGroupList = listModule.JobFamilyGroupList as React.ComponentType
  JobFamilyGroupDetail = groupDetailModule.JobFamilyGroupDetail as React.ComponentType
  JobFamilyDetail = familyDetailModule.JobFamilyDetail as React.ComponentType
  JobRoleDetail = roleDetailModule.JobRoleDetail as React.ComponentType
  JobLevelDetail = levelDetailModule.JobLevelDetail as React.ComponentType
})

describe('Job Catalog permissions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.hasPermission.mockReturnValue(false)
    mocks.mockUseJobFamilyGroups.mockReset()
    mocks.mockUseJobFamilies.mockReset()
    mocks.mockUseJobRoles.mockReset()
    mocks.mockUseJobLevels.mockReset()
  })

  const createClient = () =>
    new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })

  it('hides creation actions on list when lacks write permission', () => {
    mocks.hasPermission.mockImplementation(scope => scope !== 'job-catalog:create')
    mocks.mockUseJobFamilyGroups.mockReturnValue({
      data: [
        {
          code: 'PROF',
          name: '专业技术类',
          status: 'ACTIVE',
          effectiveDate: '2025-01-01',
          endDate: null,
          description: '描述',
          recordId: 'r1',
        },
      ],
      isLoading: false,
    })

    const client = createClient()
    try {
      render(
        <QueryClientProvider client={client}>
          <MemoryRouter>
            <JobFamilyGroupList />
          </MemoryRouter>
        </QueryClientProvider>,
      )

      expect(screen.queryByText('新增职类')).not.toBeInTheDocument()
    } finally {
      client.clear()
    }

  })

  it('hides edit and version actions on detail pages without update permission', () => {
    mocks.hasPermission.mockImplementation(scope => scope !== 'job-catalog:update')
    mocks.mockUseJobFamilyGroups.mockReturnValue({
      data: [
        {
          code: 'PROF',
          name: '专业技术类',
          status: 'ACTIVE',
          effectiveDate: '2025-01-01',
          endDate: null,
          description: '描述',
          recordId: 'uuid-1',
        },
      ],
      isLoading: false,
    })

    const client = createClient()
    try {
      render(
        <QueryClientProvider client={client}>
          <MemoryRouter initialEntries={['/positions/catalog/family-groups/PROF']}>
            <Routes>
              <Route path="/positions/catalog/family-groups/:code" element={<JobFamilyGroupDetail />} />
            </Routes>
          </MemoryRouter>
        </QueryClientProvider>,
      )

      expect(screen.queryByText('新增版本')).not.toBeInTheDocument()
      expect(screen.queryByText('编辑当前版本')).not.toBeInTheDocument()
    } finally {
      client.clear()
    }
  })

  it('respects update permission for nested detail hierarchy', () => {
    mocks.hasPermission.mockReturnValue(false)
    mocks.mockUseJobFamilies.mockReturnValue({
      data: [
        {
          code: 'PROF-SALES',
          groupCode: 'PROF',
          name: '销售序列',
          status: 'ACTIVE',
          effectiveDate: '2025-01-10',
          endDate: null,
          description: '描述',
          recordId: 'fam-1',
        },
      ],
      isLoading: false,
    })

    const client = createClient()
    try {
      render(
        <QueryClientProvider client={client}>
          <MemoryRouter initialEntries={['/positions/catalog/families/PROF-SALES']}>
            <Routes>
              <Route path="/positions/catalog/families/:code" element={<JobFamilyDetail />} />
            </Routes>
          </MemoryRouter>
        </QueryClientProvider>,
      )

      expect(screen.queryByText('编辑当前版本')).not.toBeInTheDocument()
      expect(screen.queryByText('新增版本')).not.toBeInTheDocument()
    } finally {
      client.clear()
    }
  })

  it('requires update scope for job roles', () => {
    mocks.hasPermission.mockReturnValue(false)
    mocks.mockUseJobRoles.mockReturnValue({
      data: [
        {
          code: 'PROF-SALES-MGR',
          familyCode: 'PROF-SALES',
          name: '销售经理',
          status: 'ACTIVE',
          effectiveDate: '2025-02-01',
          endDate: null,
          description: '描述',
          recordId: 'role-1',
        },
      ],
      isLoading: false,
    })

    const client = createClient()
    try {
      render(
        <QueryClientProvider client={client}>
          <MemoryRouter initialEntries={['/positions/catalog/roles/PROF-SALES-MGR']}>
            <Routes>
              <Route path="/positions/catalog/roles/:code" element={<JobRoleDetail />} />
            </Routes>
          </MemoryRouter>
        </QueryClientProvider>,
      )

      expect(screen.queryByText('编辑当前版本')).not.toBeInTheDocument()
      expect(screen.queryByText('新增版本')).not.toBeInTheDocument()
    } finally {
      client.clear()
    }
  })

  it('requires update scope for job levels and enforces role context', () => {
    mocks.hasPermission.mockReturnValue(false)
    mocks.mockUseJobLevels.mockReturnValue({
      data: [
        {
          code: 'PROF-SALES-MGR-L3',
          roleCode: 'PROF-SALES-MGR',
          name: '高级销售经理',
          status: 'ACTIVE',
          effectiveDate: '2025-03-01',
          endDate: null,
          description: '描述',
          recordId: 'level-1',
          levelRank: 3,
        },
      ],
      isLoading: false,
    })

    const client = createClient()
    try {
      render(
        <QueryClientProvider client={client}>
          <MemoryRouter
            initialEntries={[
              {
                pathname: '/positions/catalog/levels/PROF-SALES-MGR-L3',
                state: { roleCode: 'PROF-SALES-MGR' },
              },
            ]}
          >
            <Routes>
              <Route path="/positions/catalog/levels/:code" element={<JobLevelDetail />} />
            </Routes>
          </MemoryRouter>
        </QueryClientProvider>,
      )

      expect(screen.queryByText('编辑当前版本')).not.toBeInTheDocument()
      expect(screen.queryByText('新增版本')).not.toBeInTheDocument()
    } finally {
      client.clear()
    }
  })
})
