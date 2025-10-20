import type { PositionRecord } from '@/shared/types/positions'
import type { PositionCatalogOption } from '@/shared/hooks/usePositionCatalogOptions'

export type PositionFormMode = 'create' | 'edit' | 'version'

export interface PositionFormProps {
  mode: PositionFormMode
  position?: PositionRecord
  onCancel?: () => void
  onSuccess?: (payload: { code: string }) => void
}

export interface PositionFormState {
  title: string
  jobFamilyGroupCode: string
  jobFamilyCode: string
  jobRoleCode: string
  jobLevelCode: string
  organizationCode: string
  positionType: string
  employmentType: string
  gradeLevel: string
  headcountCapacity: string
  reportsToPositionCode: string
  effectiveDate: string
  operationReason: string
}

export type PositionFormErrors = Partial<Record<keyof PositionFormState, string>>

export type SelectOption = PositionCatalogOption

export const createInitialState = (
  mode: PositionFormMode,
  position?: PositionRecord,
): PositionFormState => ({
  title: position?.title ?? '',
  jobFamilyGroupCode: position?.jobFamilyGroupCode ?? '',
  jobFamilyCode: position?.jobFamilyCode ?? '',
  jobRoleCode: position?.jobRoleCode ?? '',
  jobLevelCode: position?.jobLevelCode ?? '',
  organizationCode: position?.organizationCode ?? '',
  positionType: position?.positionType ?? 'REGULAR',
  employmentType: position?.employmentType ?? 'FULL_TIME',
  gradeLevel: position?.gradeLevel ?? '',
  headcountCapacity: position ? String(position.headcountCapacity) : '',
  reportsToPositionCode: position?.reportsToPositionCode ?? '',
  effectiveDate: mode === 'version' ? '' : position?.effectiveDate ?? '',
  operationReason: '',
})
