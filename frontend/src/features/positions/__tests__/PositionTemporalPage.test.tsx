// @vitest-environment jsdom
import { render, screen } from '@testing-library/react'
import { beforeEach, afterEach, afterAll, describe, expect, it, vi, type Mock } from 'vitest'
import type { PositionDetailResult, PositionRecord } from '@/shared/types/positions'

vi.stubEnv('VITE_POSITIONS_MOCK_MODE', 'false')

const navigateMock = vi.fn()
let params: { code?: string } = { code: 'P9000001' }

vi.mock('@/shared/hooks/useEnterprisePositions', () => ({
  usePositionDetail: vi.fn(),
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => navigateMock,
    useParams: () => params,
  }
})

vi.mock('../components/PositionDetails', () => ({
  PositionDetails: () => <div data-testid="position-details" />,
}))

vi.mock('../components/PositionForm', () => ({
  PositionForm: ({ mode }: { mode: string }) => <div data-testid={`position-form-${mode}`} />,
}))

const { usePositionDetail } = await import('@/shared/hooks/useEnterprisePositions')
const mockedUsePositionDetail = usePositionDetail as unknown as Mock

const { PositionTemporalPage } = await import('../PositionTemporalPage')

const createPositionRecord = (overrides: Partial<PositionRecord> = {}): PositionRecord => ({
  code: 'P9000001',
  recordId: 'rec-001',
  title: 'HR Manager',
  jobFamilyGroupCode: 'PROF',
  jobFamilyGroupName: 'Professional',
  jobFamilyCode: 'PROF-HR',
  jobFamilyName: 'Human Resources',
  jobRoleCode: 'PROF-HR-MGR',
  jobRoleName: 'HR Manager',
  jobLevelCode: 'P3',
  jobLevelName: 'P3',
  organizationCode: '2000001',
  organizationName: '总部人力资源部',
  positionType: 'REGULAR',
  employmentType: 'FULL_TIME',
  headcountCapacity: 1,
  headcountInUse: 1,
  availableHeadcount: 0,
  gradeLevel: 'L3',
  reportsToPositionCode: 'P8000001',
  status: 'ACTIVE',
  effectiveDate: '2024-01-01',
  endDate: undefined,
  isCurrent: true,
  isFuture: false,
  createdAt: '2024-01-01T00:00:00.000Z',
  updatedAt: '2024-01-01T00:00:00.000Z',
  ...overrides,
})

const createDetailResult = (): PositionDetailResult => ({
  position: createPositionRecord(),
  timeline: [
    {
      id: 'timeline-001',
      status: 'ACTIVE',
      title: '职位创建',
      effectiveDate: '2024-01-01',
    },
  ],
  currentAssignment: null,
  assignments: [],
  transfers: [],
  versions: [
    createPositionRecord({ updatedAt: '2024-01-01T00:00:00.000Z', recordId: 'rec-001', isCurrent: true }),
    createPositionRecord({
      recordId: 'rec-002',
      effectiveDate: '2024-06-01',
      createdAt: '2024-05-01T00:00:00.000Z',
      updatedAt: '2024-05-01T00:00:00.000Z',
      status: 'PLANNED',
      isCurrent: false,
      isFuture: true,
    }),
  ],
  fetchedAt: '2025-10-18T00:00:00.000Z',
})

describe('PositionTemporalPage', () => {
  beforeEach(() => {
    params = { code: 'P9000001' }
    mockedUsePositionDetail.mockReset()
    mockedUsePositionDetail.mockReturnValue({
      data: createDetailResult(),
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
  })

  afterEach(() => {
    navigateMock.mockReset()
  })

  afterAll(() => {
    vi.unstubAllEnvs()
  })

  it('renders GraphQL versions list when数据可用', () => {
    render(<PositionTemporalPage />)

    expect(screen.getByTestId('position-temporal-page')).toBeInTheDocument()
    expect(screen.getByTestId('position-details')).toBeInTheDocument()
    expect(screen.getByTestId('position-version-list')).toBeInTheDocument()
    expect(screen.getByText('职位版本记录')).toBeInTheDocument()
    expect(screen.getByText('当前版本')).toBeInTheDocument()
    expect(screen.getByText('计划版本')).toBeInTheDocument()
  })

  it('shows guidance when职位编码缺失', () => {
    params = {}
    mockedUsePositionDetail.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    render(<PositionTemporalPage />)

    expect(screen.getByText('未提供职位编码，请从职位列表进入详情页。')).toBeInTheDocument()
  })

  it('validates职位编码格式', () => {
    params = { code: 'INVALID' }
    mockedUsePositionDetail.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    render(<PositionTemporalPage />)

    expect(screen.getByText('职位编码格式不正确，请从职位列表页面重新进入。')).toBeInTheDocument()
  })
})
