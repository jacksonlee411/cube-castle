import type { ManifestFieldMap } from '@/shared/manifest/types'

export type PositionFormField =
  | 'title'
  | 'jobFamilyGroupCode'
  | 'jobFamilyCode'
  | 'jobRoleCode'
  | 'jobLevelCode'
  | 'organizationCode'
  | 'reportsToPositionCode'
  | 'positionType'
  | 'employmentType'
  | 'gradeLevel'
  | 'headcountCapacity'
  | 'effectiveDate'
  | 'operationReason'

export const positionFormManifest: ManifestFieldMap<PositionFormField> = {
  title: {
    name: 'title',
    label: '职位名称',
    description: 'OpenAPI `title` 字段，必填，用于展示岗位名称。',
    placeholder: '请输入职位名称',
    required: true,
  },
  jobFamilyGroupCode: {
    name: 'jobFamilyGroupCode',
    label: '职类编码',
    description: 'OpenAPI `jobFamilyGroupCode` 字段，对应职类主数据。',
    placeholder: '例如：PROF',
    required: true,
  },
  jobFamilyCode: {
    name: 'jobFamilyCode',
    label: '职种编码',
    description: 'OpenAPI `jobFamilyCode` 字段，需要先选择职类。',
    placeholder: '例如：PROF-IT',
    required: true,
  },
  jobRoleCode: {
    name: 'jobRoleCode',
    label: '职务编码',
    description: 'OpenAPI `jobRoleCode` 字段，需要先选择职种。',
    placeholder: '例如：PROF-IT-BKND',
    required: true,
  },
  jobLevelCode: {
    name: 'jobLevelCode',
    label: '职级编码',
    description: 'OpenAPI `jobLevelCode` 字段，需要先选择职务。',
    placeholder: '例如：P5',
    required: true,
  },
  organizationCode: {
    name: 'organizationCode',
    label: '所属组织编码',
    description: 'OpenAPI `organizationCode` 字段，需填写 7 位组织编码。',
    placeholder: '如：1000100',
    required: true,
  },
  reportsToPositionCode: {
    name: 'reportsToPositionCode',
    label: '汇报职位编码',
    description: 'OpenAPI `reportsToPositionCode` 字段，可选，指向上级岗位。',
    placeholder: '例如：P1000001',
  },
  positionType: {
    name: 'positionType',
    label: '职位类型',
    description: 'OpenAPI `positionType` 字段，REGULAR/TEMPORARY/CONTRACTOR。',
    required: true,
  },
  employmentType: {
    name: 'employmentType',
    label: '雇佣方式',
    description: 'OpenAPI `employmentType` 字段，FULL_TIME/PART_TIME/INTERN。',
    required: true,
  },
  gradeLevel: {
    name: 'gradeLevel',
    label: '职级等级',
    description: 'OpenAPI `gradeLevel` 字段，可选，用于补充企业内部等级。',
    placeholder: '例如：L3',
  },
  headcountCapacity: {
    name: 'headcountCapacity',
    label: '编制容量 (FTE)',
    description: 'OpenAPI `headcountCapacity` 字段，必填，可输入小数。',
    placeholder: '例如：1 或 2.5',
    required: true,
  },
  effectiveDate: {
    name: 'effectiveDate',
    label: '生效日期',
    description: 'OpenAPI `effectiveDate` 字段；版本模式下代表新版本生效时间。',
    placeholder: 'YYYY-MM-DD',
    required: true,
  },
  operationReason: {
    name: 'operationReason',
    label: '操作原因',
    description: 'OpenAPI `operationReason` 字段，记录变更背景。',
    placeholder: '请说明此次操作原因',
    required: true,
  },
}

export const getPositionFieldMeta = (field: PositionFormField) => positionFormManifest[field]

