import { describe, expect, it } from 'vitest'
import { validatePositionForm, parseHeadcount } from '../validation'
import type { PositionFormState } from '../types'

const createValidState = (overrides: Partial<PositionFormState> = {}): PositionFormState => ({
  title: 'HR Manager',
  jobFamilyGroupCode: 'PROF',
  jobFamilyCode: 'PROF-HR',
  jobRoleCode: 'PROF-HR-MGR',
  jobLevelCode: 'P3',
  organizationCode: '2000001',
  positionType: 'REGULAR',
  employmentType: 'FULL_TIME',
  gradeLevel: 'L3',
  headcountCapacity: '1',
  reportsToPositionCode: 'P1000001',
  effectiveDate: '2025-01-01',
  operationReason: 'Initial load',
  ...overrides,
})

describe('validatePositionForm', () => {
  it('返回缺失字段错误', () => {
    const errors = validatePositionForm(createValidState({
      title: '',
      organizationCode: '',
      operationReason: '',
    }))

    expect(errors.title).toBe('请填写职位名称')
    expect(errors.organizationCode).toBe('请填写所属组织编码')
    expect(errors.operationReason).toBe('请填写操作原因')
  })

  it('校验组织编码与汇报职位编码格式', () => {
    const errors = validatePositionForm(
      createValidState({
        organizationCode: '0123456',
        reportsToPositionCode: 'INVALID',
        headcountCapacity: '-1',
      }),
    )

    expect(errors.organizationCode).toBe('组织编码需为7位数字，且首位不能为0')
    expect(errors.reportsToPositionCode).toBe('汇报职位编码需为 P + 7 位数字')
    expect(errors.headcountCapacity).toBe('编制容量需为非负数字')
  })

  it('有效输入返回空错误集', () => {
    const errors = validatePositionForm(createValidState())

    expect(errors).toEqual({})
  })
})

describe('parseHeadcount', () => {
  it('解析有效数字', () => {
    expect(parseHeadcount('1.5')).toBeCloseTo(1.5)
  })

  it('非法输入返回 null', () => {
    expect(parseHeadcount('abc')).toBeNull()
    expect(parseHeadcount('-2')).toBeNull()
  })
})

