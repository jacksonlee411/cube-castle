// @vitest-environment jsdom
import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { beforeEach, vi, type Mock } from 'vitest'
import { PositionHeadcountDashboard } from '../components/PositionHeadcountDashboard'
import type { PositionHeadcountStats } from '@/shared/types/positions'

vi.mock('@/shared/hooks/useEnterprisePositions', () => ({
  usePositionHeadcountStats: vi.fn(),
}))

vi.mock('@workday/canvas-kit-react/text-input', () => ({
  TextInput: ({ children, ...props }: React.InputHTMLAttributes<HTMLInputElement>) => (
    <div>
      <input {...props} />
      {children}
    </div>
  ),
}))

vi.mock('@workday/canvas-kit-react/button', () => ({
  PrimaryButton: ({ children, type, ...rest }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    <button type={type ?? 'button'} {...rest}>
      {children}
    </button>
  ),
  SecondaryButton: ({ children, type, ...rest }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    <button type={type ?? 'button'} {...rest}>
      {children}
    </button>
  ),
}))

vi.mock('@workday/canvas-kit-react/checkbox', () => ({
  Checkbox: ({ children, ...props }: React.InputHTMLAttributes<HTMLInputElement>) => (
    <label>
      <input type="checkbox" {...props} />
      {children}
    </label>
  ),
}))

vi.mock('@workday/canvas-kit-react/card', () => ({
  Card: ({ children, style, ...rest }: React.HTMLAttributes<HTMLDivElement>) => (
    <div style={style as React.CSSProperties} {...rest}>
      {children}
    </div>
  ),
}))

vi.mock('@workday/canvas-kit-react/layout', () => ({
  Box: ({
    children,
    style,
    minWidth,
    borderRadius,
    backgroundColor,
    border,
    boxShadow,
    padding,
    gap,
    ...rest
  }: React.HTMLAttributes<HTMLDivElement> & {
    minWidth?: string
    borderRadius?: string
    backgroundColor?: string
    border?: string
    boxShadow?: string
    padding?: string
    gap?: string
  }) => (
    <div
      style={{ minWidth, borderRadius, backgroundColor, border, boxShadow, padding, gap, ...(style as React.CSSProperties) }}
      {...rest}
    >
      {children}
    </div>
  ),
  Flex: ({
    children,
    style,
    flexDirection,
    justifyContent,
    alignItems,
    flexWrap,
    rowGap,
    gap,
    as,
    ...rest
  }: React.HTMLAttributes<HTMLDivElement> & {
    flexDirection?: string
    justifyContent?: string
    alignItems?: string
    flexWrap?: string
    rowGap?: string
    gap?: string
    as?: string
  }) => (
    React.createElement(
      as === 'form' ? 'form' : 'div',
      {
        ...rest,
        style: {
          display: 'flex',
          flexDirection,
          justifyContent,
          alignItems,
          flexWrap,
          rowGap,
          gap,
          ...(style as React.CSSProperties),
        },
      },
      children,
    )
  ),
}))

vi.mock('@workday/canvas-kit-react/table', () => {
  const Head = ({ children, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
    <thead {...props}>{children}</thead>
  )
  const Body = ({ children, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) => (
    <tbody {...props}>{children}</tbody>
  )
  const Row = ({ children, ...props }: React.HTMLAttributes<HTMLTableRowElement>) => <tr {...props}>{children}</tr>
  const Header = ({ children, ...props }: React.ThHTMLAttributes<HTMLTableCellElement>) => (
    <th {...props}>{children}</th>
  )
  const Cell = ({ children, ...props }: React.TdHTMLAttributes<HTMLTableCellElement>) => (
    <td {...props}>{children}</td>
  )

  const TableComponent = ({ children, ...props }: React.HTMLAttributes<HTMLTableElement>) => (
    <table {...props}>{children}</table>
  )

  return {
    Table: Object.assign(TableComponent, {
      Head,
      Body,
      Row,
      Header,
      Cell,
    }),
  }
})

vi.mock('@workday/canvas-kit-react/text', () => ({
  Heading: ({ children, ...props }: React.HTMLAttributes<HTMLHeadingElement>) => <h3 {...props}>{children}</h3>,
  Text: ({ children, ...props }: React.HTMLAttributes<HTMLSpanElement>) => <span {...props}>{children}</span>,
}))

vi.mock('@workday/canvas-kit-react/tokens', () => ({
  colors: {
    blueberry500: '#1a73e8',
    cantaloupe500: '#ffb300',
    cinnamon500: '#c75a00',
    greenApple500: '#2eb872',
    licorice300: '#666',
    licorice400: '#444',
    licorice500: '#222',
    frenchVanilla100: '#fff',
    soap400: '#ddd',
  },
  space: {
    l: '24px',
    m: '16px',
    s: '12px',
    xxs: '4px',
    xxxs: '2px',
  },
}))

const { usePositionHeadcountStats } = await import('@/shared/hooks/useEnterprisePositions')
const mockedUseHeadcountStats = usePositionHeadcountStats as unknown as Mock

const sampleStats: PositionHeadcountStats = {
  organizationCode: '1000000',
  organizationName: '根组织',
  totalCapacity: 120,
  totalFilled: 90,
  totalAvailable: 30,
  fillRate: 0.75,
  byLevel: [
    { jobLevelCode: 'S1', capacity: 40, utilized: 30, available: 10 },
    { jobLevelCode: 'S2', capacity: 80, utilized: 60, available: 20 },
  ],
  byType: [
    { positionType: 'REGULAR', capacity: 100, filled: 80, available: 20 },
    { positionType: 'CONTRACT', capacity: 20, filled: 10, available: 10 },
  ],
  fetchedAt: '2025-10-17T00:00:00.000Z',
}

beforeEach(() => {
  mockedUseHeadcountStats.mockReset()
  mockedUseHeadcountStats.mockReturnValue({
    data: sampleStats,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
    isFetching: false,
  })
  if (!URL.createObjectURL) {
    URL.createObjectURL = vi.fn(() => 'blob:mock') as unknown as typeof URL.createObjectURL
  } else {
    vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:mock')
  }
  if (!URL.revokeObjectURL) {
    URL.revokeObjectURL = vi.fn() as unknown as typeof URL.revokeObjectURL
  } else {
    vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})
  }
  vi.spyOn(document.body, 'appendChild')
  vi.spyOn(document.body, 'removeChild')
})

describe('PositionHeadcountDashboard', () => {
  it('renders summary tiles and tables when data is available', () => {
    render(<PositionHeadcountDashboard organizationCode="1000000" />)

    expect(screen.getByTestId('position-headcount-dashboard')).toBeInTheDocument()
    expect(screen.getByText('总编制')).toBeInTheDocument()
    expect(screen.getByText('120')).toBeInTheDocument()
    expect(screen.getByText('占用率')).toBeInTheDocument()
    expect(screen.getByTestId('headcount-level-table')).toBeInTheDocument()
    expect(screen.getByTestId('headcount-type-table')).toBeInTheDocument()
  })

  it('submits organization code and triggers export', async () => {
    render(<PositionHeadcountDashboard />)

    const input = screen.getByTestId('headcount-org-input') as HTMLInputElement
    fireEvent.change(input, { target: { value: '1000000' } })
    fireEvent.click(screen.getByRole('button', { name: '加载统计' }))

    await waitFor(() => {
      expect(mockedUseHeadcountStats).toHaveBeenCalledWith({
        organizationCode: '1000000',
        includeSubordinates: true,
      })
    })

    const exportButton = screen.getByTestId('headcount-export')
    fireEvent.click(exportButton)

    expect(URL.createObjectURL).toHaveBeenCalled()
  })
})
