import type { PositionRecord } from '@/shared/types/positions'

export const POSITION_VERSION_FIELDS = [
  { key: 'status', label: '状态' },
  { key: 'effectiveDate', label: '生效日期' },
  { key: 'endDate', label: '结束日期' },
  { key: 'positionType', label: '职位类型' },
  { key: 'employmentType', label: '雇佣方式' },
  { key: 'title', label: '职位名称' },
  { key: 'organizationCode', label: '所属组织编码' },
  { key: 'organizationName', label: '所属组织名称' },
  { key: 'jobFamilyGroupCode', label: '职类编码' },
  { key: 'jobFamilyGroupName', label: '职类名称' },
  { key: 'jobFamilyCode', label: '职种编码' },
  { key: 'jobFamilyName', label: '职种名称' },
  { key: 'jobRoleCode', label: '职务编码' },
  { key: 'jobRoleName', label: '职务名称' },
  { key: 'jobLevelCode', label: '职级编码' },
  { key: 'jobLevelName', label: '职级名称' },
  { key: 'gradeLevel', label: '职级等级' },
  { key: 'reportsToPositionCode', label: '汇报职位编码' },
  { key: 'headcountCapacity', label: '编制容量' },
  { key: 'headcountInUse', label: '已占用编制' },
  { key: 'availableHeadcount', label: '可用编制' },
  { key: 'isCurrent', label: '当前版本' },
  { key: 'isFuture', label: '计划版本' },
  { key: 'createdAt', label: '创建时间' },
  { key: 'updatedAt', label: '更新时间' },
] as const satisfies Array<{ key: keyof PositionRecord; label: string }>

export type PositionVersionFieldKey = (typeof POSITION_VERSION_FIELDS)[number]['key']
