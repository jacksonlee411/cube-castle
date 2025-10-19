import type { PositionFormErrors, PositionFormState } from './types'

const REQUIRED_FIELDS: Array<{ key: keyof PositionFormState; message: string }> = [
  { key: 'title', message: '请填写职位名称' },
  { key: 'jobFamilyGroupCode', message: '请填写职类编码' },
  { key: 'jobFamilyCode', message: '请填写职种编码' },
  { key: 'jobRoleCode', message: '请填写职务编码' },
  { key: 'jobLevelCode', message: '请填写职级编码' },
  { key: 'organizationCode', message: '请填写所属组织编码' },
  { key: 'positionType', message: '请选择职位类型' },
  { key: 'employmentType', message: '请选择雇佣方式' },
  { key: 'headcountCapacity', message: '请填写编制容量' },
  { key: 'effectiveDate', message: '请填写生效日期' },
  { key: 'operationReason', message: '请填写操作原因' },
]

const ORGANIZATION_CODE_PATTERN = /^[1-9]\d{6}$/
const POSITION_CODE_PATTERN = /^P\d{7}$/

const isNonEmpty = (value: string) => value.trim().length > 0

export const parseHeadcount = (value: string): number | null => {
  const parsed = Number.parseFloat(value)
  if (Number.isNaN(parsed) || parsed < 0) {
    return null
  }
  return parsed
}

export const validatePositionForm = (state: PositionFormState): PositionFormErrors => {
  const errors: PositionFormErrors = {}

  REQUIRED_FIELDS.forEach(({ key, message }) => {
    if (!isNonEmpty(state[key])) {
      errors[key] = message
    }
  })

  if (state.organizationCode && !ORGANIZATION_CODE_PATTERN.test(state.organizationCode.trim())) {
    errors.organizationCode = '组织编码需为7位数字，且首位不能为0'
  }

  if (state.reportsToPositionCode && !POSITION_CODE_PATTERN.test(state.reportsToPositionCode.trim())) {
    errors.reportsToPositionCode = '汇报职位编码需为 P + 7 位数字'
  }

  if (parseHeadcount(state.headcountCapacity) === null) {
    errors.headcountCapacity = '编制容量需为非负数字'
  }

  return errors
}

