// @vitest-environment jsdom
import { render, screen, fireEvent } from '@testing-library/react'
import { beforeEach, afterEach, vi, type Mock } from 'vitest'
import { PositionDashboard } from '../PositionDashboard'
import type { PositionRecord, PositionsQueryResult } from '@/shared/types/positions'
import temporalEntitySelectors from '@/shared/testids/temporalEntity'

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

vi.mock('../components/dashboard/PositionVacancyBoard', () => ({
  // Use literal to avoid import timing/hoist issues in mock factory; equals to selector value
  PositionVacancyBoard: () => <div data-testid="temporal-position-vacancy-board" />,
}))

vi.mock('../components/transfer/PositionTransferDialog', () => ({
  PositionTransferDialog: () => <div data-testid="position-transfer-dialog" />,
}))

vi.mock('../components/dashboard/PositionHeadcountDashboard', () => ({
  // Use literal to avoid import timing/hoist issues in mock factory; equals to selector value
  PositionHeadcountDashboard: () => <div data-testid="temporal-position-headcount-dashboard" />,
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
  vi.stubEnv('VITE_POSITIONS_MOCK_MODE', 'false')
  mockedUseEnterprisePositions.mockReset()
  navigateMock.mockReset()

  mockedUseEnterprisePositions.mockReturnValue({
    data: positionsQueryResult,
    isLoading: false,
    isError: false,
  })
})

afterEach(() => {
  vi.stubEnv('VITE_POSITIONS_MOCK_MODE', 'false')
})

describe('PositionDashboard（Stage 1 数据接入）', () => {
  it('渲染职位列表与统计信息', () => {
    render(<PositionDashboard />)

    expect(screen.getByTestId(temporalEntitySelectors.position.dashboard)).toBeInTheDocument()
    expect(screen.getByText('职位管理（Stage 1 数据接入）')).toBeInTheDocument()
    expect(
      screen.getByText('当前页面依赖 GraphQL 查询服务与 REST 命令服务，请确保后端接口可用。'),
    ).toBeInTheDocument()
    expect(screen.getByText('岗位总数')).toBeInTheDocument()
    expect(screen.getByTestId(temporalEntitySelectors.position.row!(samplePosition.code))).toBeInTheDocument()
    expect(screen.getAllByText('物业保洁员')[0]).toBeInTheDocument()
    expect(screen.getByTestId(temporalEntitySelectors.position.vacancyBoard!)).toBeInTheDocument()
    expect(screen.getByTestId(temporalEntitySelectors.position.headcountDashboard!)).toBeInTheDocument()
    expect(screen.getByTestId(temporalEntitySelectors.position.createButton!)).toBeInTheDocument()
  })

  it('点击职位行时跳转到详情页', () => {
    render(<PositionDashboard />)

    const row = screen.getByTestId(temporalEntitySelectors.position.row!(samplePosition.code))
    fireEvent.click(row)

    expect(navigateMock).toHaveBeenCalledWith('/positions/P9000001')
  })

  it('点击创建职位按钮跳转到新建页面', () => {
    render(<PositionDashboard />)

    fireEvent.click(screen.getByTestId(temporalEntitySelectors.position.createButton!))
    expect(navigateMock).toHaveBeenCalledWith('/positions/new')
  })

  it('接口报错时展示错误提示', () => {
    mockedUseEnterprisePositions.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    })

    render(<PositionDashboard />)

    expect(screen.getByTestId(temporalEntitySelectors.position.errorBox!)).toHaveTextContent(
      '无法加载职位数据，请刷新页面或联系系统管理员。',
    )
    expect(screen.getByTestId(temporalEntitySelectors.position.createButton!)).toBeDisabled()
  })

  it('无数据时展示空态提醒', () => {
    mockedUseEnterprisePositions.mockReturnValue({
      data: {
        positions: [],
        pagination: positionsQueryResult.pagination,
        totalCount: 0,
        timestamp: positionsQueryResult.timestamp,
      },
      isLoading: false,
      isError: false,
    })

    render(<PositionDashboard />)

    expect(
      screen.getByText('暂无职位记录，如果这是异常情况，请检查数据同步或后端服务状态。'),
    ).toBeInTheDocument()
    expect(screen.getByText('暂无职位数据')).toBeInTheDocument()
  })

  it('Mock 模式下提示只读并禁用创建', () => {
    vi.stubEnv('VITE_POSITIONS_MOCK_MODE', 'true')

    render(<PositionDashboard />)

    expect(screen.getByTestId(temporalEntitySelectors.position.mockBanner!)).toBeInTheDocument()
    expect(screen.getByTestId(temporalEntitySelectors.position.createButton!)).toBeDisabled()
  })
})
