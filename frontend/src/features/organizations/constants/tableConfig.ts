export const TABLE_COLUMNS = [
  { key: 'code', label: '编码', width: '120px' },
  { key: 'name', label: '名称', width: 'auto' },
  { key: 'unitType', label: '类型', width: '120px' },
  { key: 'status', label: '状态', width: '100px' },
  { key: 'level', label: '层级', width: '80px' },
  { key: 'actions', label: '操作', width: '140px' }
] as const;

export const STATUS_COLORS = {
  ACTIVE: 'positive',
  INACTIVE: 'default',
  PLANNED: 'hint'
} as const;

export const LOADING_STATES = {
  IDLE: 'idle',
  LOADING: 'loading',
  DELETING: 'deleting',
  ERROR: 'error'
} as const;