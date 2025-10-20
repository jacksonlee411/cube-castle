// @vitest-environment jsdom
import React from 'react'
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, it, beforeEach, expect, vi } from 'vitest'
const mocks = vi.hoisted(() => ({
  mockUseJobFamilyGroups: vi.fn(),
  mockUseJobFamilies: vi.fn(),
  mockUseJobRoles: vi.fn(),
  mockUseJobLevels: vi.fn(),
  mockUseCreateJobFamilyGroup: vi.fn(),
  mockUseCreateJobFamilyGroupVersion: vi.fn(),
  mockUseUpdateJobFamilyGroup: vi.fn(),
  mockUseCreateJobFamilyVersion: vi.fn(),
  mockUseUpdateJobFamily: vi.fn(),
  mockUseCreateJobRoleVersion: vi.fn(),
  mockUseUpdateJobRole: vi.fn(),
  mockUseCreateJobLevelVersion: vi.fn(),
  mockUseUpdateJobLevel: vi.fn(),
}))

let JobFamilyGroupList: React.ComponentType
let JobFamilyGroupDetail: React.ComponentType
let JobFamilyDetail: React.ComponentType
let JobRoleDetail: React.ComponentType
let JobLevelDetail: React.ComponentType

vi.mock('@/shared/auth/hooks', () => ({
  useAuth: () => ({ hasPermission: () => true }),
}))

vi.mock('@/shared/hooks/useJobCatalog', () => ({
  useJobFamilyGroups: (...args: unknown[]) => mocks.mockUseJobFamilyGroups(...args),
  useJobFamilies: (...args: unknown[]) => mocks.mockUseJobFamilies(...args),
  useJobRoles: (...args: unknown[]) => mocks.mockUseJobRoles(...args),
  useJobLevels: (...args: unknown[]) => mocks.mockUseJobLevels(...args),
}))

vi.mock('@/shared/hooks/useJobCatalogMutations', () => ({
  useCreateJobFamilyGroup: () => mocks.mockUseCreateJobFamilyGroup(),
  useCreateJobFamilyGroupVersion: () => mocks.mockUseCreateJobFamilyGroupVersion(),
  useUpdateJobFamilyGroup: () => mocks.mockUseUpdateJobFamilyGroup(),
  useCreateJobFamily: vi.fn(),
  useCreateJobRole: vi.fn(),
  useCreateJobLevel: vi.fn(),
  useCreateJobFamilyVersion: () => mocks.mockUseCreateJobFamilyVersion(),
  useCreateJobRoleVersion: () => mocks.mockUseCreateJobRoleVersion(),
  useCreateJobLevelVersion: () => mocks.mockUseCreateJobLevelVersion(),
  useUpdateJobFamily: () => mocks.mockUseUpdateJobFamily(),
  useUpdateJobRole: () => mocks.mockUseUpdateJobRole(),
  useUpdateJobLevel: () => mocks.mockUseUpdateJobLevel(),
}))

vi.mock('../shared/CatalogForm', () => ({
  CatalogForm: ({ isOpen, onSubmit, children, submitLabel = '提交', isSubmitting = false, errorMessage }: any) => {
    if (!isOpen) {
      return null
    }
    return (
      <form data-testid="mock-catalog-form" onSubmit={onSubmit}>
        {children}
        {errorMessage ? <div data-testid="mock-catalog-error">{errorMessage}</div> : null}
        <button type="submit" disabled={isSubmitting}>
          {submitLabel}
        </button>
      </form>
    )
  },
}))

vi.mock('../shared/CatalogForm.tsx', () => ({
  CatalogForm: ({ isOpen, onSubmit, children, submitLabel = '提交', isSubmitting = false, errorMessage }: any) => {
    if (!isOpen) {
      return null
    }
    return (
      <form data-testid="mock-catalog-form" onSubmit={onSubmit}>
        {children}
        {errorMessage ? <div data-testid="mock-catalog-error">{errorMessage}</div> : null}
        <button type="submit" disabled={isSubmitting}>
          {submitLabel}
        </button>
      </form>
    )
  },
}))

beforeAll(async () => {
  const listModule = await import('../family-groups/JobFamilyGroupList')
  const detailModule = await import('../family-groups/JobFamilyGroupDetail')
  const familyDetailModule = await import('../families/JobFamilyDetail')
  const roleDetailModule = await import('../roles/JobRoleDetail')
  const levelDetailModule = await import('../levels/JobLevelDetail')
  JobFamilyGroupList = listModule.JobFamilyGroupList as React.ComponentType
  JobFamilyGroupDetail = detailModule.JobFamilyGroupDetail as React.ComponentType
  JobFamilyDetail = familyDetailModule.JobFamilyDetail as React.ComponentType
  JobRoleDetail = roleDetailModule.JobRoleDetail as React.ComponentType
  JobLevelDetail = levelDetailModule.JobLevelDetail as React.ComponentType
})

describe('Job Catalog pages', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.mockUseCreateJobFamilyGroup.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseCreateJobFamilyGroupVersion.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseUpdateJobFamilyGroup.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseCreateJobFamilyVersion.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseUpdateJobFamily.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseCreateJobRoleVersion.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseUpdateJobRole.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseCreateJobLevelVersion.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseUpdateJobLevel.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
  })

  it('renders job family group list and allows creating new entries', async () => {
    const mutateAsync = vi.fn().mockResolvedValue(undefined)
    mocks.mockUseCreateJobFamilyGroup.mockReturnValue({ mutateAsync, isPending: false })
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

    render(
      <MemoryRouter>
        <JobFamilyGroupList />
      </MemoryRouter>,
    )

    expect(screen.getByText('职类管理')).toBeInTheDocument()
    expect(screen.getByText('专业技术类')).toBeInTheDocument()

    fireEvent.change(screen.getByPlaceholderText('输入关键字搜索'), { target: { value: '销售' } })
    await waitFor(() => expect(screen.getByText('暂无数据')).toBeInTheDocument())

    fireEvent.change(screen.getByPlaceholderText('输入关键字搜索'), { target: { value: '' } })
    await waitFor(() => expect(screen.getByText('专业技术类')).toBeInTheDocument())

    fireEvent.click(screen.getByText('新增职类'))

    const createForm = await screen.findByTestId('mock-catalog-form')

    fireEvent.change(screen.getByPlaceholderText('例如：PROF'), { target: { value: '' } })
    fireEvent.click(within(createForm).getByText('确认创建'))
    expect(await screen.findByText('职类编码需为 4-6 位大写字母')).toBeInTheDocument()

    fireEvent.change(screen.getByPlaceholderText('例如：PROF'), { target: { value: 'SALE' } })
    fireEvent.click(within(createForm).getByText('确认创建'))
    expect(await screen.findByText('请输入职类名称')).toBeInTheDocument()
  })

  it('renders job family group detail and opens version form', async () => {
    mocks.mockUseCreateJobFamilyGroupVersion.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseJobFamilyGroups.mockReturnValue({
      data: [
        {
          code: 'PROF',
          name: '专业技术类',
          status: 'ACTIVE',
          effectiveDate: '2025-01-01',
          endDate: null,
          description: '专家序列',
          recordId: 'uuid-1',
        },
      ],
      isLoading: false,
    })

    render(
      <MemoryRouter initialEntries={['/positions/catalog/family-groups/PROF']}>
        <Routes>
          <Route path="/positions/catalog/family-groups/:code" element={<JobFamilyGroupDetail />} />
        </Routes>
      </MemoryRouter>,
    )

    expect(screen.getByText('职类详情')).toBeInTheDocument()
    expect(screen.getByText('专业技术类')).toBeInTheDocument()
    expect(screen.getByText('uuid-1')).toBeInTheDocument()

    fireEvent.click(screen.getByText('新增版本'))

    const versionForm = await screen.findByTestId('mock-catalog-form')

    fireEvent.click(within(versionForm).getByText('提交'))
    expect(await screen.findByText('请选择生效日期')).toBeInTheDocument()
  })

  it('allows updating job family group with prefilled version form', async () => {
    const mutateAsync = vi.fn().mockResolvedValue(undefined)
    mocks.mockUseUpdateJobFamilyGroup.mockReturnValue({ mutateAsync, isPending: false })
    mocks.mockUseJobFamilyGroups.mockReturnValue({
      data: [
        {
          code: 'PROF',
          name: '专业技术类',
          status: 'ACTIVE',
          effectiveDate: '2025-01-01',
          endDate: null,
          description: '专家序列',
          recordId: 'uuid-1',
        },
      ],
      isLoading: false,
    })

    render(
      <MemoryRouter initialEntries={['/positions/catalog/family-groups/PROF']}>
        <Routes>
          <Route path="/positions/catalog/family-groups/:code" element={<JobFamilyGroupDetail />} />
        </Routes>
      </MemoryRouter>,
    )

    const editButton = screen.getByText('编辑当前版本')
    fireEvent.click(editButton)

    const editForm = await screen.findByTestId('mock-catalog-form')
    const nameInput = within(editForm).getByPlaceholderText('版本名称') as HTMLInputElement
    expect(nameInput.value).toBe('专业技术类')
    fireEvent.change(nameInput, { target: { value: '专业技术类（更新）' } })

    const dateInput = within(editForm).getByDisplayValue('2025-01-01') as HTMLInputElement
    expect(dateInput.value).toBe('2025-01-01')

    fireEvent.click(within(editForm).getByText('保存更新'))

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledTimes(1)
      expect(mutateAsync).toHaveBeenCalledWith({
        code: 'PROF',
        recordId: 'uuid-1',
        name: '专业技术类（更新）',
        status: 'ACTIVE',
        effectiveDate: '2025-01-01',
        description: '专家序列',
      })
    })
  })

  it('allows updating job family with group code preserved', async () => {
    const mutateAsync = vi.fn().mockResolvedValue(undefined)
    mocks.mockUseUpdateJobFamily.mockReturnValue({ mutateAsync, isPending: false })
    mocks.mockUseJobFamilies.mockReturnValue({
      data: [
        {
          code: 'PROF-SALES',
          name: '销售序列',
          status: 'ACTIVE',
          effectiveDate: '2025-01-10',
          endDate: null,
          description: '销售岗位集合',
          recordId: 'family-1',
          groupCode: 'PROF',
        },
      ],
      isLoading: false,
    })

    render(
      <MemoryRouter initialEntries={['/positions/catalog/families/PROF-SALES']}>
        <Routes>
          <Route path="/positions/catalog/families/:code" element={<JobFamilyDetail />} />
        </Routes>
      </MemoryRouter>,
    )

    fireEvent.click(screen.getByText('编辑当前版本'))

    const editForm = await screen.findByTestId('mock-catalog-form')
    const nameInput = within(editForm).getByPlaceholderText('版本名称') as HTMLInputElement
    expect(nameInput.value).toBe('销售序列')

    fireEvent.change(nameInput, { target: { value: '销售序列（更新）' } })
    fireEvent.click(within(editForm).getByText('保存更新'))

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith({
        code: 'PROF-SALES',
        recordId: 'family-1',
        jobFamilyGroupCode: 'PROF',
        name: '销售序列（更新）',
        status: 'ACTIVE',
        effectiveDate: '2025-01-10',
        description: '销售岗位集合',
      })
    })
  })

  it('allows updating job role with derived family code', async () => {
    const mutateAsync = vi.fn().mockResolvedValue(undefined)
    mocks.mockUseUpdateJobRole.mockReturnValue({ mutateAsync, isPending: false })
    mocks.mockUseJobRoles.mockReturnValue({
      data: [
        {
          code: 'PROF-SALES-MGR',
          name: '销售经理',
          status: 'ACTIVE',
          effectiveDate: '2025-02-01',
          endDate: null,
          description: '负责销售团队',
          recordId: 'role-1',
          familyCode: 'PROF-SALES',
        },
      ],
      isLoading: false,
    })

    render(
      <MemoryRouter initialEntries={['/positions/catalog/roles/PROF-SALES-MGR']}>
        <Routes>
          <Route path="/positions/catalog/roles/:code" element={<JobRoleDetail />} />
        </Routes>
      </MemoryRouter>,
    )

    fireEvent.click(screen.getByText('编辑当前版本'))

    const editForm = await screen.findByTestId('mock-catalog-form')
    const nameInput = within(editForm).getByPlaceholderText('版本名称') as HTMLInputElement
    expect(nameInput.value).toBe('销售经理')

    fireEvent.click(within(editForm).getByText('保存更新'))

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith({
        code: 'PROF-SALES-MGR',
        recordId: 'role-1',
        jobFamilyCode: 'PROF-SALES',
        name: '销售经理',
        status: 'ACTIVE',
        effectiveDate: '2025-02-01',
        description: '负责销售团队',
      })
    })
  })

  it('allows updating job level with role code and rank persisted', async () => {
    const mutateAsync = vi.fn().mockResolvedValue(undefined)
    mocks.mockUseUpdateJobLevel.mockReturnValue({ mutateAsync, isPending: false })
    mocks.mockUseJobLevels.mockReturnValue({
      data: [
        {
          code: 'PROF-SALES-MGR-L3',
          name: '高级销售经理',
          status: 'ACTIVE',
          effectiveDate: '2025-03-01',
          endDate: null,
          description: '关键岗位',
          recordId: 'level-1',
          roleCode: 'PROF-SALES-MGR',
          levelRank: 3,
        },
      ],
      isLoading: false,
    })

    render(
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
      </MemoryRouter>,
    )

    fireEvent.click(screen.getByText('编辑当前版本'))

    const editForm = await screen.findByTestId('mock-catalog-form')
    const statusSelect = within(editForm).getByTestId('canvas-select') as HTMLSelectElement
    expect(statusSelect.value).toBe('ACTIVE')

    fireEvent.click(within(editForm).getByText('保存更新'))

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith({
        code: 'PROF-SALES-MGR-L3',
        recordId: 'level-1',
        jobRoleCode: 'PROF-SALES-MGR',
        levelRank: 3,
        name: '高级销售经理',
        status: 'ACTIVE',
        effectiveDate: '2025-03-01',
        description: '关键岗位',
      })
    })
  })
})
