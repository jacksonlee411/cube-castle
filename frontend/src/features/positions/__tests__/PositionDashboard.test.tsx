// @vitest-environment jsdom
import { render, screen, fireEvent } from '@testing-library/react'
import { beforeEach, vi, type Mock } from 'vitest'
import { PositionDashboard } from '../PositionDashboard'
import type { PositionRecord, PositionsQueryResult } from '@/shared/types/positions'

const navigateMock = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => navigateMock,
  }
})

vi.mock('@/shared/hooks/useEnterprisePositions', () => ({
  useEnterprisePositions: vi.fn(),
}))

vi.mock('@/shared/hooks/usePositionMutations', () => ({
  useTransferPosition: vi.fn(() => ({
    mutateAsync: vi.fn(),
    isPending: false,
    isSuccess: false,
    error: null,
  })),
}))

vi.mock('../components/PositionVacancyBoard', () => ({
  PositionVacancyBoard: () => <div data-testid="position-vacancy-board" />,
}))

vi.mock('../components/PositionTransferDialog', () => ({
  PositionTransferDialog: () => <div data-testid="position-transfer-dialog" />,
}))

vi.mock('../components/PositionHeadcountDashboard', () => ({
  PositionHeadcountDashboard: () => <div data-testid="position-headcount-dashboard" />,
}))

const { useEnterprisePositions } = await import('@/shared/hooks/useEnterprisePositions')
const mockedUseEnterprisePositions = useEnterprisePositions as unknown as Mock

const samplePosition: PositionRecord = {
  code: 'P9000001',
  title: '物业保洁员',
  jobFamilyGroupCode: 'OPER',
  jobFamilyGroupName: 'OPER',
  jobFamilyCode: 'OPER-OPS',
  jobFamilyName: 'OPER-OPS',
  jobRoleCode: 'OPER-OPS-CLEAN',
  jobRoleName: 'OPER-OPS-CLEAN',
  jobLevelCode: 'S1',
  jobLevelName: 'S1',
  organizationCode: '2000010',
  organizationName: '上海虹桥商务区物业项目',
  positionType: 'REGULAR',
  employmentType: 'FULL_TIME',
  headcountCapacity: 8,
  headcountInUse: 6,
  availableHeadcount: 2,
  gradeLevel: undefined,
  reportsToPositionCode: 'P2000008',
  status: 'FILLED',
  effectiveDate: '2024-01-01',
  endDate: undefined,
  isCurrent: true,
  isFuture: false,
  createdAt: '2024-01-01T00:00:00.000Z',
  updatedAt: '2024-01-01T00:00:00.000Z',
}

const positionsQueryResult: PositionsQueryResult = {
  positions: [samplePosition],
  pagination: {
    total: 1,
    page: 1,
    pageSize: 100,
    hasNext: false,
    hasPrevious: false,
  },
  totalCount: 1,
  timestamp: '2025-10-16T00:00:00.000Z',
}

beforeEach(() => {
  mockedUseEnterprisePositions.mockReset()
  navigateMock.mockReset()

  mockedUseEnterprisePositions.mockReturnValue({
    data: positionsQueryResult,
    isLoading: false,
    isError: false,
  })
})

describe('PositionDashboard（Stage 1 数据接入）', () => {
  it('渲染职位列表与统计信息', () => {
    render(<PositionDashboard />)

    expect(screen.getByTestId('position-dashboard')).toBeInTheDocument()
    expect(screen.getByText('职位管理（Stage 1 数据接入）')).toBeInTheDocument()
    expect(screen.getByText('岗位总数')).toBeInTheDocument()
    expect(screen.getByTestId('position-row-P9000001')).toBeInTheDocument()
    expect(screen.getAllByText('物业保洁员')[0]).toBeInTheDocument()
    expect(screen.getByTestId('position-vacancy-board')).toBeInTheDocument()
    expect(screen.getByTestId('position-headcount-dashboard')).toBeInTheDocument()
    expect(screen.getByTestId('position-create-button')).toBeInTheDocument()
  })

  it('点击职位行时跳转到详情页', () => {
    render(<PositionDashboard />)

    const row = screen.getByTestId('position-row-P9000001')
    fireEvent.click(row)

    expect(navigateMock).toHaveBeenCalledWith('/positions/P9000001')
  })

  it('点击创建职位按钮跳转到新建页面', () => {
    render(<PositionDashboard />)

    fireEvent.click(screen.getByTestId('position-create-button'))
    expect(navigateMock).toHaveBeenCalledWith('/positions/new')
  })
})
