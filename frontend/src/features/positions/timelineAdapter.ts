import type { TimelineVersion } from '@/features/temporal/components'
import type { PositionRecord, PositionTimelineEvent } from '@/shared/types/positions'

const DEFAULT_LEVEL = 0
const DEFAULT_SORT_ORDER = 1

const toLifecycleStatus = (record: PositionRecord | PositionTimelineEvent): 'CURRENT' | 'HISTORICAL' | 'PLANNED' => {
  if ('isFuture' in record && record.isFuture) {
    return 'PLANNED'
  }

  if (record.isCurrent) {
    return 'CURRENT'
  }

  return 'HISTORICAL'
}

const toBusinessStatus = (status: string): 'ACTIVE' | 'INACTIVE' => {
  if (status === 'INACTIVE' || status === 'DELETED' || status === 'SUSPENDED') {
    return 'INACTIVE'
  }

  return 'ACTIVE'
}

export const buildPositionVersionKey = (record: PositionRecord, fallbackIndex: number): string => {
  if (record.recordId && record.recordId.trim().length > 0) {
    return record.recordId
  }

  return `${record.code}-${record.effectiveDate}-${fallbackIndex}`
}

export const sortPositionVersions = (versions: PositionRecord[]): PositionRecord[] =>
  [...versions].sort((a, b) => new Date(b.effectiveDate).getTime() - new Date(a.effectiveDate).getTime())

export const createTimelineVersion = (version: PositionRecord, index: number): TimelineVersion => ({
  recordId: buildPositionVersionKey(version, index),
  code: version.code,
  name: version.title,
  unitType: 'POSITION',
  status: version.status,
  level: DEFAULT_LEVEL,
  effectiveDate: version.effectiveDate,
  endDate: version.endDate ?? undefined,
  isCurrent: version.isCurrent,
  createdAt: version.createdAt,
  updatedAt: version.updatedAt,
  parentCode: version.reportsToPositionCode ?? undefined,
  description: undefined,
  lifecycleStatus: toLifecycleStatus(version),
  businessStatus: toBusinessStatus(version.status),
  dataStatus: version.status === 'DELETED' ? 'DELETED' : 'NORMAL',
  codePath: undefined,
  namePath: undefined,
  sortOrder: DEFAULT_SORT_ORDER,
  changeReason: undefined,
  suspended_at: undefined,
  suspended_by: undefined,
  suspension_reason: undefined,
  deleted_at: undefined,
  deleted_by: undefined,
  deletion_reason: undefined,
})

export const mapPositionVersionsToTimeline = (
  versions: PositionRecord[],
  options: { sorted?: boolean } = {},
): TimelineVersion[] => {
  const source = options.sorted ? versions : sortPositionVersions(versions)

  return source.map((version, index) => createTimelineVersion(version, index))
}

export const mapTimelineEventsToTimeline = (events: PositionTimelineEvent[]): TimelineVersion[] =>
  [...events]
    .sort((a, b) => new Date(b.effectiveDate).getTime() - new Date(a.effectiveDate).getTime())
    .map((event, index) => ({
      recordId: `${event.id}-${index}`,
      code: event.id,
      name: event.title,
      unitType: 'POSITION',
      status: event.status,
      level: DEFAULT_LEVEL,
      effectiveDate: event.effectiveDate,
      endDate: event.endDate ?? undefined,
      isCurrent: Boolean(event.isCurrent),
      createdAt: event.effectiveDate,
      updatedAt: event.endDate ?? event.effectiveDate,
      parentCode: undefined,
      description: event.changeReason ?? undefined,
      lifecycleStatus: toLifecycleStatus(event),
      businessStatus: toBusinessStatus(event.status),
      dataStatus: 'NORMAL',
      codePath: undefined,
      namePath: undefined,
      sortOrder: DEFAULT_SORT_ORDER,
      changeReason: event.changeReason ?? undefined,
      suspended_at: undefined,
      suspended_by: undefined,
      suspension_reason: undefined,
      deleted_at: undefined,
      deleted_by: undefined,
      deletion_reason: undefined,
    }))
