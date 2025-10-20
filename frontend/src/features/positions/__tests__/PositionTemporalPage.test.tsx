// @vitest-environment jsdom

import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, afterAll, afterEach, describe, expect, it, vi, type Mock } from 'vitest'
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

vi.mock('@/features/audit/components/AuditHistorySection', () => ({
  AuditHistorySection: ({ recordId }: { recordId: string }) => (
    <div data-testid="audit-history-section">{recordId}</div>
  ),
}))

vi.mock('@/features/temporal/components', () => ({
  TimelineComponent: ({ versions }: { versions: unknown[] }) => (
    <div data-testid="timeline-component">{versions.length}</div>
  ),
}))

vi.mock('@workday/canvas-kit-react/layout', () => ({
  Box: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  Flex: ({ children, ...props }: any) => <div {...props}>{children}</div>,
}))

vi.mock('@workday/canvas-kit-react/text', () => ({
  Heading: ({ children, ...props }: any) => <h3 {...props}>{children}</h3>,
  Text: ({ children, ...props }: any) => <span {...props}>{children}</span>,
}))

vi.mock('@workday/canvas-kit-react/button', () => ({
  PrimaryButton: ({ children, ...props }: any) => <button {...props}>{children}</button>,
  SecondaryButton: ({ children, ...props }: any) => <button {...props}>{children}</button>,
}))

vi.mock('@workday/canvas-kit-react/card', () => ({
  Card: ({ children, ...props }: any) => <div {...props}>{children}</div>,
}))

vi.mock('@workday/canvas-kit-react/tokens', () => ({
  colors: {
    licorice400: '#333',
    licorice500: '#222',
    licorice600: '#111',
    cinnamon600: '#c00',
    cinnamon500: '#d22',
    cinnamon100: '#fee',
    soap300: '#ccc',
    soap200: '#ddd',
    soap100: '#eee',
    soap400: '#bbb',
    frenchVanilla100: '#fff',
    blueberry600: '#08c',
    blueberry400: '#09c',
    blueberry50: '#def',
    cantaloupe600: '#f80',
  },
  space: {
    xxxs: '2px',
    xxs: '4px',
    xs: '8px',
    s: '12px',
    m: '16px',
    l: '20px',
    xl: '24px',
  },
}))

vi.mock('@workday/canvas-kit-react/table', () => {
  const Table = ({ children, ...props }: any) => <table {...props}>{children}</table>
  Table.Head = ({ children, ...props }: any) => <thead {...props}>{children}</thead>
  Table.Body = ({ children, ...props }: any) => <tbody {...props}>{children}</tbody>
  Table.Row = ({ children, ...props }: any) => <tr {...props}>{children}</tr>
  Table.Header = ({ children, ...props }: any) => <th {...props}>{children}</th>
  Table.Cell = ({ children, ...props }: any) => <td {...props}>{children}</td>
  return { Table }
})

vi.mock('@workday/canvas-kit-react/switch', () => ({
  Switch: ({ checked, onChange, ...props }: any) => (
    <input
      type="checkbox"
      checked={checked}
      onChange={event => onChange?.({ target: { checked: event.target.checked } })}
      {...props}
    />
  ),
}))

vi.mock('../components/PositionForm', () => ({
  PositionForm: ({ mode }: { mode: string }) => <div data-testid={`position-form-${mode}`} />,
}))

vi.mock('../components/transfer/PositionTransferDialog', () => ({
  PositionTransferDialog: () => <div data-testid="position-transfer-dialog" />,
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
    createPositionRecord({
      updatedAt: '2024-01-01T00:00:00.000Z',
      recordId: 'rec-001',
      isCurrent: true,
    }),
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
    vi.stubEnv('VITE_POSITIONS_MOCK_MODE', 'false')
    params = { code: 'P9000001' }
    mockedUsePositionDetail.mockReset()
    mockedUsePositionDetail.mockImplementation((_code: string | undefined, _options?: unknown) => ({
      data: createDetailResult(),
      isLoading: false,
      isFetching: false,
      isError: false,
      error: undefined,
      refetch: vi.fn(),
    }))
  })

  afterEach(() => {
    navigateMock.mockReset()
    vi.stubEnv('VITE_POSITIONS_MOCK_MODE', 'false')
  })

  afterAll(() => {
    vi.unstubAllEnvs()
  })

  it('renders detail layout并可打开版本历史页签', () => {
    render(<PositionTemporalPage />)

    expect(screen.getByTestId('position-temporal-page')).toBeInTheDocument()
    expect(screen.getByTestId('position-overview-card')).toBeInTheDocument()
    expect(screen.getByText('概览')).toBeInTheDocument()
    expect(screen.getByText('版本历史')).toBeInTheDocument()

    fireEvent.click(screen.getByText('版本历史'))
    expect(screen.getByTestId('position-version-toolbar')).toBeInTheDocument()
    expect(screen.getByTestId('position-version-list')).toBeInTheDocument()
    expect(screen.getByText('职位版本记录')).toBeInTheDocument()
  })

  it('toggles includeDeleted flag when开关切换', async () => {
    render(<PositionTemporalPage />)

    expect(mockedUsePositionDetail).toHaveBeenLastCalledWith(
      'P9000001',
      expect.objectContaining({ includeDeleted: false }),
    )

    fireEvent.click(screen.getByText('版本历史'))

    const toggle = screen.getByTestId('position-version-include-deleted')
    fireEvent.click(toggle)

    await waitFor(() =>
      expect(mockedUsePositionDetail).toHaveBeenLastCalledWith(
        'P9000001',
        expect.objectContaining({ includeDeleted: true }),
      ),
    )
  })

  it('navigates to overview when版本表点击行', () => {
    render(<PositionTemporalPage />)

    fireEvent.click(screen.getByText('版本历史'))
    const versionRow = screen.getAllByTestId(/position-version-row/)[1]
    fireEvent.click(versionRow)

    expect(screen.getByTestId('position-overview-card')).toBeInTheDocument()
    expect(screen.getByText(/当前版本：/)).toBeInTheDocument()
  })

  it('shows提示 when职位编码缺失', () => {
    params = {}
    mockedUsePositionDetail.mockReturnValue({
      data: undefined,
      isLoading: false,
      isFetching: false,
      isError: false,
      error: undefined,
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
      isFetching: false,
      isError: false,
      error: undefined,
      refetch: vi.fn(),
    })

    render(<PositionTemporalPage />)

    expect(screen.getByText('职位编码格式不正确，请从职位列表页面重新进入。')).toBeInTheDocument()
  })

  it('Mock 模式下隐藏写操作', () => {
    vi.stubEnv('VITE_POSITIONS_MOCK_MODE', 'true')

    render(<PositionTemporalPage />)

    expect(screen.getByTestId('position-temporal-page')).toBeInTheDocument()
    expect(screen.getByTestId('position-mock-banner')).toBeInTheDocument()
    expect(screen.queryByTestId('position-edit-button')).not.toBeInTheDocument()
    expect(screen.queryByTestId('position-version-button')).not.toBeInTheDocument()
  })

  it('Mock 模式下创建页面仅展示指引', () => {
    vi.stubEnv('VITE_POSITIONS_MOCK_MODE', 'true')
    params = { code: 'new' }
    mockedUsePositionDetail.mockReturnValue({
      data: undefined,
      isLoading: false,
      isFetching: false,
      isError: false,
      error: undefined,
      refetch: vi.fn(),
    })

    render(<PositionTemporalPage />)

    expect(screen.getByText('⚠️ Mock 模式下无法创建职位。')).toBeInTheDocument()
  })
})
