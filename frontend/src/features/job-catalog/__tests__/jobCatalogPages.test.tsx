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
}))

let JobFamilyGroupList: React.ComponentType
let JobFamilyGroupDetail: React.ComponentType

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
  useCreateJobFamilyVersion: vi.fn(),
  useCreateJobRoleVersion: vi.fn(),
  useCreateJobLevelVersion: vi.fn(),
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
  JobFamilyGroupList = listModule.JobFamilyGroupList as React.ComponentType
  JobFamilyGroupDetail = detailModule.JobFamilyGroupDetail as React.ComponentType
})

describe('Job Catalog pages', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.mockUseCreateJobFamilyGroup.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseCreateJobFamilyGroupVersion.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
    mocks.mockUseUpdateJobFamilyGroup.mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(undefined), isPending: false })
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
        name: '专业技术类（更新）',
        status: 'ACTIVE',
        effectiveDate: '2025-01-01',
        description: '专家序列',
      })
    })
  })
})
