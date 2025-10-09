import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import FormActions from '../FormActions'
import type { InlineVersionRecord } from '../types'

const baseVersion: InlineVersionRecord = {
  recordId: 'rec-1',
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
  code: '1000001',
  name: '示例组织',
  unitType: 'DEPARTMENT',
  status: 'ACTIVE',
  effectiveDate: '2025-01-01',
}

const renderFormActions = (override: Partial<React.ComponentProps<typeof FormActions>> = {}) => {
  const props: React.ComponentProps<typeof FormActions> = {
    currentMode: 'edit',
    isEditingHistory: false,
    isSubmitting: false,
    loading: false,
    selectedVersion: baseVersion,
    onCancel: vi.fn(),
    onDeactivateClick: vi.fn(),
    onDeleteOrganizationClick: vi.fn(),
    onToggleEditHistory: vi.fn(),
    onCancelEditHistory: vi.fn(),
    onSubmitEditHistory: vi.fn(),
    onSubmitNewVersion: vi.fn(),
    originalHistoryData: null,
    onStartInsertVersion: vi.fn(),
    isDeactivating: false,
    canDeleteOrganization: false,
    isProcessingDelete: false,
    ...override,
  }

  return render(<FormActions {...props} />)
}

describe('FormActions delete buttons', () => {
  it('shows organization delete button when earliest version selected', () => {
    renderFormActions({ canDeleteOrganization: true })

    expect(screen.getByTestId('temporal-delete-organization-button')).toBeInTheDocument()
    expect(screen.queryByTestId('temporal-delete-record-button')).toBeNull()
  })

  it('shows record delete button when not earliest', () => {
    renderFormActions({ canDeleteOrganization: false })

    expect(screen.getByTestId('temporal-delete-record-button')).toBeInTheDocument()
    expect(screen.queryByTestId('temporal-delete-organization-button')).toBeNull()
  })
})
