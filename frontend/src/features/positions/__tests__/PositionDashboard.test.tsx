// @vitest-environment jsdom
import { render, screen, fireEvent } from '@testing-library/react'
import { beforeEach, vi, type Mock } from 'vitest'
import { PositionDashboard } from '../PositionDashboard'
import type {
  PositionRecord,
  PositionDetailResult,
  PositionsQueryResult,
  VacantPositionsQueryResult,
} from '@/shared/types/positions'

vi.mock('@/shared/hooks/useEnterprisePositions', () => ({
  useEnterprisePositions: vi.fn(),
  usePositionDetail: vi.fn(),
  useVacantPositions: vi.fn(),
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

const { useEnterprisePositions, usePositionDetail, useVacantPositions } = await import('@/shared/hooks/useEnterprisePositions')
const mockedUseEnterprisePositions = useEnterprisePositions as unknown as Mock
const mockedUsePositionDetail = usePositionDetail as unknown as Mock
const mockedUseVacantPositions = useVacantPositions as unknown as Mock

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

const positionDetailResult: PositionDetailResult = {
  position: samplePosition,
  timeline: [
    {
      id: 'rec-001',
      status: 'FILLED',
      title: '岗位填充',
      effectiveDate: '2024-03-01',
      changeReason: '张三入职',
    },
  ],
  currentAssignment: {
    assignmentId: 'assign-001',
    positionCode: samplePosition.code,
    positionRecordId: 'pos-rec-001',
    employeeId: 'emp-001',
    employeeName: '张三',
    employeeNumber: 'E001',
    assignmentType: 'PRIMARY',
    assignmentStatus: 'ACTIVE',
    fte: 1,
    startDate: '2024-03-01',
    endDate: undefined,
    isCurrent: true,
    notes: '夜班负责人',
    createdAt: '2024-03-01T00:00:00.000Z',
    updatedAt: '2024-03-02T00:00:00.000Z',
  },
  assignments: [
    {
      assignmentId: 'assign-001',
      positionCode: samplePosition.code,
      positionRecordId: 'pos-rec-001',
      employeeId: 'emp-001',
      employeeName: '张三',
      employeeNumber: 'E001',
      assignmentType: 'PRIMARY',
      assignmentStatus: 'ACTIVE',
      fte: 1,
      startDate: '2024-03-01',
      endDate: undefined,
      isCurrent: true,
      notes: '夜班负责人',
      createdAt: '2024-03-01T00:00:00.000Z',
      updatedAt: '2024-03-02T00:00:00.000Z',
    },
    {
      assignmentId: 'assign-000',
      positionCode: samplePosition.code,
      positionRecordId: 'pos-rec-000',
      employeeId: 'emp-000',
      employeeName: '李四',
      employeeNumber: 'E000',
      assignmentType: 'PRIMARY',
      assignmentStatus: 'ENDED',
      fte: 1,
      startDate: '2023-01-01',
      endDate: '2024-02-28',
      isCurrent: false,
      notes: '内部调岗',
      createdAt: '2023-01-01T00:00:00.000Z',
      updatedAt: '2024-02-28T00:00:00.000Z',
    },
  ],
  transfers: [
    {
      transferId: 'transfer-001',
      positionCode: samplePosition.code,
      fromOrganizationCode: '1001000',
      toOrganizationCode: samplePosition.organizationCode,
      effectiveDate: '2024-02-15',
      initiatedBy: { id: 'user-001', name: '刘洋' },
      operationReason: '项目调整',
      createdAt: '2024-02-16T00:00:00.000Z',
    },
  ],
  fetchedAt: '2025-10-16T00:00:00.000Z',
}

const vacantPositionsResult: VacantPositionsQueryResult = {
  data: [
    {
      positionCode: 'P9000002',
      organizationCode: '2000011',
      organizationName: '北京朝阳商务区物业项目',
      jobFamilyCode: 'OPER-OPS',
      jobRoleCode: 'OPER-OPS-CLEAN',
      jobLevelCode: 'S1',
      vacantSince: '2024-05-01',
      headcountCapacity: 4,
      headcountAvailable: 2.5,
      totalAssignments: 3,
    },
  ],
  pagination: {
    total: 1,
    page: 1,
    pageSize: 25,
    hasNext: false,
    hasPrevious: false,
  },
  totalCount: 1,
  fetchedAt: '2025-10-17T00:00:00.000Z',
}

beforeEach(() => {
  mockedUseEnterprisePositions.mockReset()
  mockedUsePositionDetail.mockReset()
  mockedUseVacantPositions.mockReset()

  mockedUseEnterprisePositions.mockReturnValue({
    data: positionsQueryResult,
    isLoading: false,
    isError: false,
  })
  mockedUsePositionDetail.mockReturnValue({
    data: positionDetailResult,
    isLoading: false,
  })
  mockedUseVacantPositions.mockReturnValue({
    data: vacantPositionsResult,
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
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
  })

  it('切换列表项时展示职位详情与时间线', () => {
    render(<PositionDashboard />)

    const row = screen.getByTestId('position-row-P9000001')
    fireEvent.click(row)

    const detailCard = screen.getByTestId('position-detail-card')
    expect(detailCard).toBeInTheDocument()
    expect(detailCard).toHaveTextContent('物业保洁员')
    expect(detailCard).toHaveTextContent('岗位填充')
    expect(detailCard).toHaveTextContent('张三入职')
    expect(detailCard).toHaveTextContent('当前任职')
    expect(detailCard).toHaveTextContent('张三')
    expect(detailCard).toHaveTextContent('调动记录')
    expect(detailCard).toHaveTextContent('项目调整')
  })
})
