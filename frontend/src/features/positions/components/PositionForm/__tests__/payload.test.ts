import { describe, expect, it } from 'vitest'
import { buildCreatePositionPayload } from '../payload'
import type { PositionFormState } from '../types'

const baseState: PositionFormState = {
  title: 'HR Manager',
  jobFamilyGroupCode: 'PROF',
  jobFamilyCode: 'PROF-HR',
  jobRoleCode: 'PROF-HR-MGR',
  jobLevelCode: 'P3',
  organizationCode: '2000001',
  positionType: 'REGULAR',
  employmentType: 'FULL_TIME',
  gradeLevel: 'L3',
  headcountCapacity: '1.5',
  reportsToPositionCode: 'P1000001',
  effectiveDate: '2025-01-01',
  operationReason: 'Initial load',
}

describe('buildCreatePositionPayload', () => {
  it('去除多余空格并解析数值', () => {
    const payload = buildCreatePositionPayload({
      ...baseState,
      title: '  HR Manager  ',
      headcountCapacity: ' 2 ',
      gradeLevel: ' ',
      reportsToPositionCode: '  ',
    })

    expect(payload.title).toBe('HR Manager')
    expect(payload.headcountCapacity).toBe(2)
    expect(payload.gradeLevel).toBeUndefined()
    expect(payload.reportsToPositionCode).toBeUndefined()
  })

  it('保留必填字段并原样传递有效值', () => {
    const payload = buildCreatePositionPayload(baseState)

    expect(payload.jobFamilyGroupCode).toBe('PROF')
    expect(payload.effectiveDate).toBe('2025-01-01')
    expect(payload.operationReason).toBe('Initial load')
    expect(payload.headcountCapacity).toBeCloseTo(1.5)
  })
})

