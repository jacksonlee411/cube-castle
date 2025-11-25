import type { ManifestFieldMap } from '@/shared/manifest/types'

export type OrganizationField =
  | 'name'
  | 'unitType'
  | 'description'
  | 'effectiveDate'
  | 'endDate'
  | 'changeReason'

export const organizationFormManifest: ManifestFieldMap<OrganizationField> = {
  name: {
    name: 'name',
    label: '组织名称',
    description: '对应 OpenAPI 中的 `name` 字段，创建时必填，建议 2~100 字符。',
    placeholder: '请输入组织名称',
    required: true,
  },
  unitType: {
    name: 'unitType',
    label: '组织类型',
    description: 'OpenAPI `unitType` 字段，用于标识部门/项目等类别。',
    placeholder: '请选择组织类型',
    required: true,
  },
  description: {
    name: 'description',
    label: '组织描述',
    description: 'OpenAPI `description` 字段，可用于补充职能说明（可选）。',
    placeholder: '可选，描述组织的职能和目的',
  },
  effectiveDate: {
    name: 'effectiveDate',
    label: '生效日期',
    description: 'OpenAPI `effectiveDate` 字段，计划组织必须大于当前日期。',
    placeholder: 'YYYY-MM-DD',
    required: true,
  },
  endDate: {
    name: 'endDate',
    label: '结束日期',
    description: 'OpenAPI `endDate` 字段，可选，需晚于生效日期。',
    placeholder: 'YYYY-MM-DD',
  },
  changeReason: {
    name: 'changeReason',
    label: '变更原因',
    description: 'OpenAPI `changeReason` 字段，用于记录审批/背景信息。',
    placeholder: '请补充变更原因（至少 5 个字）',
  },
}

export const getOrganizationFieldMeta = (field: OrganizationField) => organizationFormManifest[field]
