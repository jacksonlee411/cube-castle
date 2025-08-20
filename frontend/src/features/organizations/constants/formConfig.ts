export const ORGANIZATION_UNIT_TYPES = {
  DEPARTMENT: '部门',
  COST_CENTER: '成本中心', 
  COMPANY: '公司',
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
  unit_type: 'DEPARTMENT' as const,
  status: 'ACTIVE' as const,
  level: 1,
  sort_order: 0,
} as const;

export const PAGINATION_DEFAULTS = {
  page: 1,
  pageSize: 20,
} as const;