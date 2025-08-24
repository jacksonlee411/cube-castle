export const ORGANIZATION_UNIT_TYPES = {
  DEPARTMENT: '部门',
  ORGANIZATION_UNIT: '组织单位',
  PROJECT_TEAM: '项目团队'
} as const;

export const ORGANIZATION_STATUSES = {
  ACTIVE: '启用',
  SUSPENDED: '停用',
  PLANNED: '计划中'
} as const;

export const ORGANIZATION_LEVELS = {
  MIN: 1,
  MAX: 10
} as const;

export const FORM_DEFAULTS = {
  unitType: 'DEPARTMENT' as const,
  status: 'ACTIVE' as const,
  level: 1,
  sortOrder: 0,
} as const;

export const PAGINATION_DEFAULTS = {
  page: 1,
  pageSize: 20,
} as const;