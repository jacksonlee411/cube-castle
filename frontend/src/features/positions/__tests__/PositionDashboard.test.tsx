// @vitest-environment jsdom
import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { PositionDashboard } from '../PositionDashboard'

describe('PositionDashboard (Stage 0)', () => {
  it('renders mock summary and table rows', () => {
    render(<PositionDashboard />)

    expect(screen.getByTestId('position-dashboard')).toBeInTheDocument()
    expect(screen.getByText('岗位总数')).toBeInTheDocument()
    expect(screen.getByText('职位名称')).toBeInTheDocument()
    expect(screen.getByTestId('position-row-P1000101')).toBeInTheDocument()
  })

  it('shows detail panel for selected position', () => {
    render(<PositionDashboard />)

    const supervisorRow = screen.getByTestId('position-row-P1000102')
    fireEvent.click(supervisorRow)

    const detailCard = screen.getByTestId('position-detail-card')
    expect(detailCard).toBeInTheDocument()
    expect(detailCard).toHaveTextContent('保洁主管')
    expect(detailCard).toHaveTextContent('编制：1 / 1')
  })
})
