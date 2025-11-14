import { colors } from '@workday/canvas-kit-react/tokens'
import type { SystemIconProps } from '@workday/canvas-kit-react/icon'
import { checkCircleIcon, clockIcon, clockPauseIcon } from '@workday/canvas-system-icons-web'
import type { TemporalEntityKind } from '../pages/TemporalEntityPage'
import type { PositionStatus } from '@/shared/types/positions'
import type { OrganizationStatus } from '@/shared/types/contract_gen'

export interface TemporalEntityStatusMeta {
  label: string
  color: string
  background: string
  border: string
  icon?: SystemIconProps['icon']
  description?: string
}

type StatusMetaRecord = Record<string, TemporalEntityStatusMeta>

const POSITION_STATUS_META: StatusMetaRecord = {
  ACTIVE: {
    label: '在编',
    color: colors.greenApple600,
    background: colors.greenApple100,
    border: colors.greenApple300,
  },
  FILLED: {
    label: '已填充',
    color: colors.blueberry600,
    background: colors.blueberry100,
    border: colors.blueberry300,
  },
  VACANT: {
    label: '空缺',
    color: colors.cantaloupe600,
    background: colors.cantaloupe100,
    border: colors.cantaloupe300,
  },
  PLANNED: {
    label: '规划中',
    color: colors.soap600,
    background: colors.soap100,
    border: colors.soap300,
  },
  INACTIVE: {
    label: '已结束',
    color: colors.licorice400,
    background: colors.soap100,
    border: colors.soap300,
  },
  SUSPENDED: {
    label: '挂起',
    color: colors.cinnamon600,
    background: colors.cinnamon100,
    border: colors.cinnamon300,
  },
  DELETED: {
    label: '已删除',
    color: colors.soap600,
    background: colors.soap100,
    border: colors.soap400,
  },
}

const ORGANIZATION_STATUS_META: StatusMetaRecord = {
  ACTIVE: {
    label: '启用',
    color: colors.greenApple600,
    icon: checkCircleIcon,
    background: colors.greenApple100,
    border: colors.greenApple300,
    description: '正常运行状态',
  },
  INACTIVE: {
    label: '已结束',
    color: colors.licorice400,
    icon: clockPauseIcon,
    background: colors.soap100,
    border: colors.soap300,
    description: '已结束的历史版本',
  },
  PLANNED: {
    label: '计划中',
    color: colors.blueberry600,
    icon: clockIcon,
    background: colors.blueberry100,
    border: colors.blueberry300,
    description: '计划启用状态',
  },
  DELETED: {
    label: '已删除',
    color: colors.soap600,
    icon: clockPauseIcon,
    background: colors.soap100,
    border: colors.soap300,
    description: '软删除记录，仅用于审计展示',
  },
}

export const TEMPORAL_ENTITY_STATUS_META: Record<TemporalEntityKind, StatusMetaRecord> = {
  position: POSITION_STATUS_META,
  organization: ORGANIZATION_STATUS_META,
}

const buildFallbackMeta = (status: string): TemporalEntityStatusMeta => ({
  label: status || '未知',
  color: colors.licorice500,
  background: colors.soap100,
  border: colors.soap400,
})

export const getTemporalEntityStatusMeta = (
  entity: TemporalEntityKind,
  status: string,
): TemporalEntityStatusMeta => {
  const key = status?.toString().toUpperCase() ?? ''
  const registry = TEMPORAL_ENTITY_STATUS_META[entity]
  return registry[key] ?? buildFallbackMeta(key)
}

export const getPositionStatusMeta = (status: PositionStatus | string): TemporalEntityStatusMeta =>
  getTemporalEntityStatusMeta('position', status)

export const getOrganizationStatusMeta = (status: OrganizationStatus | string): TemporalEntityStatusMeta =>
  getTemporalEntityStatusMeta('organization', status)
