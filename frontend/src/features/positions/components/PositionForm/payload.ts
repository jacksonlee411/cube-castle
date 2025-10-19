import type { CreatePositionRequest } from '@/shared/types/positions'
import type { PositionFormState } from './types'
import { parseHeadcount } from './validation'

export const buildCreatePositionPayload = (state: PositionFormState): CreatePositionRequest => ({
  title: state.title.trim(),
  jobFamilyGroupCode: state.jobFamilyGroupCode.trim(),
  jobFamilyCode: state.jobFamilyCode.trim(),
  jobRoleCode: state.jobRoleCode.trim(),
  jobLevelCode: state.jobLevelCode.trim(),
  organizationCode: state.organizationCode.trim(),
  positionType: state.positionType,
  employmentType: state.employmentType,
  gradeLevel: state.gradeLevel.trim() || undefined,
  headcountCapacity: parseHeadcount(state.headcountCapacity) ?? 0,
  reportsToPositionCode: state.reportsToPositionCode.trim() || undefined,
  effectiveDate: state.effectiveDate,
  operationReason: state.operationReason.trim(),
})

