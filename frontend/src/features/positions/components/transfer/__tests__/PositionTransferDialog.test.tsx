// @vitest-environment jsdom
import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { PositionTransferDialog } from '../PositionTransferDialog'
import type { PositionRecord } from '@/shared/types/positions'
import temporalEntitySelectors from '@/shared/testids/temporalEntity'

// Mock transfer hook to avoid real network/mutations
vi.mock('@/shared/hooks/usePositionMutations', () => ({
  useTransferPosition: () => ({
    mutateAsync: vi.fn(),
    isPending: false,
    isSuccess: false,
    error: null,
  }),
}))

// Simplify Canvas Modal behavior for JSDOM: always render children
vi.mock('@workday/canvas-kit-react/modal', () => {
  const Modal = ({ children }: any) => <div data-testid="modal-root">{children}</div>
  Modal.Overlay = ({ children }: any) => <div data-testid="modal-overlay">{children}</div>
  Modal.Card = ({ children, ...rest }: any) => <div {...rest}>{children}</div>
  Modal.Heading = ({ children }: any) => <h3>{children}</h3>
  Modal.CloseIcon = ({ onClick }: any) => <button onClick={onClick}>X</button>
  const useModalModel = () => ({ state: { visibility: 'visible' }, events: { show: vi.fn(), hide: vi.fn() } })
  return { Modal, useModalModel }
})

const samplePosition: PositionRecord = {
  code: 'P9000001',
  recordId: 'rec-001',
  title: '测试职位',
  jobFamilyGroupCode: 'OPER',
  jobFamilyGroupName: 'OPER',
  jobFamilyCode: 'OPER-OPS',
  jobFamilyName: 'OPER-OPS',
  jobRoleCode: 'OPER-OPS-CLEAN',
  jobRoleName: 'OPER-OPS-CLEAN',
  jobLevelCode: 'S1',
  jobLevelName: 'S1',
  organizationCode: '1000000',
  organizationName: '根组织',
  positionType: 'REGULAR',
  employmentType: 'FULL_TIME',
  headcountCapacity: 1,
  headcountInUse: 1,
  availableHeadcount: 0,
  gradeLevel: undefined,
  reportsToPositionCode: undefined,
  status: 'ACTIVE',
  effectiveDate: '2024-01-01',
  endDate: undefined,
  isCurrent: true,
  isFuture: false,
  createdAt: '2024-01-01T00:00:00.000Z',
  updatedAt: '2024-01-01T00:00:00.000Z',
}

describe('PositionTransferDialog', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_POSITIONS_MOCK_MODE', 'false')
  })

  it('renders transfer controls with centralized testids', () => {
    render(<PositionTransferDialog position={samplePosition} />)

    // Open button
    expect(screen.getByTestId(temporalEntitySelectors.position.transferOpen!)).toBeInTheDocument()

    // Modal content (simplified mock is always visible)
    expect(screen.getByTestId('modal-root')).toBeInTheDocument()

    // Form fields
    expect(screen.getByTestId(temporalEntitySelectors.position.transferTarget!)).toBeInTheDocument()
    expect(screen.getByTestId(temporalEntitySelectors.position.transferDate!)).toBeInTheDocument()
    expect(screen.getByTestId(temporalEntitySelectors.position.transferReason!)).toBeInTheDocument()
    expect(screen.getByTestId(temporalEntitySelectors.position.transferReassign!)).toBeInTheDocument()
    expect(screen.getByTestId(temporalEntitySelectors.position.transferConfirm!)).toBeInTheDocument()
  })

  it('allows input values to be changed', () => {
    render(<PositionTransferDialog position={samplePosition} />)
    const target = screen.getByTestId(temporalEntitySelectors.position.transferTarget!) as HTMLInputElement
    fireEvent.change(target, { target: { value: '1000001' } })
    expect(target.value).toBe('1000001')
  })
})

