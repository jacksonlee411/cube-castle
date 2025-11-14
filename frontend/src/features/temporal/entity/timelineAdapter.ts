import type { TimelineVersion } from '@/features/temporal/components'
import type { TemporalEntityKind } from '../pages/TemporalEntityPage'
import type { PositionRecord, PositionTimelineEvent } from '@/shared/types/positions'

const DEFAULT_LEVEL = 0
const DEFAULT_SORT_ORDER = 1

type TimelineVersionWithoutId = Omit<TimelineVersion, 'recordId'>

export interface TemporalTimelineRecord extends Partial<TimelineVersionWithoutId> {
  recordId?: string | null
  code: string
  name: string
  status: TimelineVersion['status']
  effectiveDate: string
}

export interface TemporalEntityTimelineAdapterConfig<TSource> {
  entity: TemporalEntityKind
  mapRecord: (source: TSource) => TemporalTimelineRecord
  sort?: (a: TSource, b: TSource) => number
}

export interface TemporalEntityTimelineAdapter<TSource> {
  toTimelineVersion: (source: TSource, index?: number) => TimelineVersion
  toTimelineVersions: (sources: TSource[], options?: { sorted?: boolean }) => TimelineVersion[]
  sortSources: (sources: TSource[]) => TSource[]
}

const ensureRecordId = (record: TemporalTimelineRecord, index: number): string => {
  if (record.recordId && record.recordId.trim().length > 0) {
    return record.recordId
  }
  return `${record.code}-${record.effectiveDate}-${index}`
}

const defaultLifecycleStatus = (record: TemporalTimelineRecord): TimelineVersion['lifecycleStatus'] => {
  if ('lifecycleStatus' in record && record.lifecycleStatus) {
    return record.lifecycleStatus
  }
  if ('isCurrent' in record && record.isCurrent) {
    return 'CURRENT'
  }
  return 'HISTORICAL'
}

const defaultBusinessStatus = (record: TemporalTimelineRecord): TimelineVersion['businessStatus'] => {
  if ('businessStatus' in record && record.businessStatus) {
    return record.businessStatus
  }
  if (record.status === 'INACTIVE' || record.status === 'DELETED' || record.status === 'SUSPENDED') {
    return 'INACTIVE'
  }
  return 'ACTIVE'
}

const defaultDataStatus = (record: TemporalTimelineRecord): TimelineVersion['dataStatus'] => {
  if ('dataStatus' in record && record.dataStatus) {
    return record.dataStatus
  }
  if (record.status === 'DELETED') {
    return 'DELETED'
  }
  return 'NORMAL'
}

const defaultUnitType = (entity: TemporalEntityKind): TimelineVersion['unitType'] =>
  entity === 'position' ? 'POSITION' : 'ORGANIZATION'

export const createTemporalTimelineAdapter = <TSource>(
  config: TemporalEntityTimelineAdapterConfig<TSource>,
): TemporalEntityTimelineAdapter<TSource> => {
  const mapBaseRecord = (source: TSource, index: number): TimelineVersion => {
    const record = config.mapRecord(source)

    return {
      recordId: ensureRecordId(record, index),
      code: record.code,
      name: record.name,
      unitType: record.unitType ?? defaultUnitType(config.entity),
      status: record.status,
      level: record.level ?? DEFAULT_LEVEL,
      effectiveDate: record.effectiveDate,
      endDate: record.endDate ?? undefined,
      isCurrent: record.isCurrent ?? false,
      createdAt: record.createdAt ?? record.effectiveDate,
      updatedAt: record.updatedAt ?? record.endDate ?? record.effectiveDate,
      parentCode: record.parentCode ?? undefined,
      description: record.description ?? undefined,
      lifecycleStatus: defaultLifecycleStatus(record),
      businessStatus: defaultBusinessStatus(record),
      dataStatus: defaultDataStatus(record),
      codePath: record.codePath ?? undefined,
      namePath: record.namePath ?? undefined,
      sortOrder: record.sortOrder ?? DEFAULT_SORT_ORDER,
      changeReason: record.changeReason ?? undefined,
      suspendedAt: record.suspendedAt ?? undefined,
      suspendedBy: record.suspendedBy ?? undefined,
      suspensionReason: record.suspensionReason ?? undefined,
      deletedAt: record.deletedAt ?? undefined,
      deletedBy: record.deletedBy ?? undefined,
      deletionReason: record.deletionReason ?? undefined,
    }
  }

  const toTimelineVersion = (source: TSource, index = 0): TimelineVersion => mapBaseRecord(source, index)

  const sortSources = (sources: TSource[]): TSource[] => {
    if (!config.sort) {
      return [...sources]
    }
    return [...sources].sort(config.sort)
  }

  const toTimelineVersions = (sources: TSource[], options?: { sorted?: boolean }): TimelineVersion[] => {
    const base = options?.sorted ? [...sources] : sortSources(sources)
    return base.map((source, index) => toTimelineVersion(source, index))
  }

  return {
    toTimelineVersion,
    toTimelineVersions,
    sortSources,
  }
}

const compareByEffectiveDateDesc = <T extends { effectiveDate: string }>(a: T, b: T): number =>
  new Date(b.effectiveDate).getTime() - new Date(a.effectiveDate).getTime()

const toLifecycleStatus = (record: { isFuture?: boolean; isCurrent?: boolean }): TimelineVersion['lifecycleStatus'] => {
  if (record.isFuture) {
    return 'PLANNED'
  }
  if (record.isCurrent) {
    return 'CURRENT'
  }
  return 'HISTORICAL'
}

const toBusinessStatus = (status: string): TimelineVersion['businessStatus'] => {
  if (status === 'INACTIVE' || status === 'DELETED' || status === 'SUSPENDED') {
    return 'INACTIVE'
  }
  return 'ACTIVE'
}

export const positionTimelineAdapter = createTemporalTimelineAdapter<PositionRecord>({
  entity: 'position',
  sort: compareByEffectiveDateDesc,
  mapRecord: version => ({
    recordId: version.recordId,
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
    lifecycleStatus: toLifecycleStatus(version),
    businessStatus: toBusinessStatus(version.status),
    dataStatus: version.status === 'DELETED' ? 'DELETED' : 'NORMAL',
  }),
})

export const positionTimelineEventAdapter = createTemporalTimelineAdapter<PositionTimelineEvent>({
  entity: 'position',
  sort: compareByEffectiveDateDesc,
  mapRecord: event => ({
    recordId: event.id,
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
    description: event.changeReason ?? undefined,
    lifecycleStatus: toLifecycleStatus({
      isFuture: false,
      isCurrent: event.isCurrent,
    }),
    businessStatus: toBusinessStatus(event.status),
    dataStatus: 'NORMAL',
    changeReason: event.changeReason ?? undefined,
  }),
})

export interface OrganizationTimelineSource {
  recordId: string
  code: string
  name: string
  unitType?: string | null
  status: string
  level?: number | null
  effectiveDate: string
  endDate?: string | null
  createdAt?: string
  updatedAt?: string
  parentCode?: string | null
  description?: string | null
  codePath?: string | null
  namePath?: string | null
  isCurrent?: boolean | null
}

export const organizationTimelineAdapter = createTemporalTimelineAdapter<OrganizationTimelineSource>({
  entity: 'organization',
  sort: compareByEffectiveDateDesc,
  mapRecord: version => {
    const isCurrent = version.isCurrent ?? version.endDate === null
    return {
      recordId: version.recordId,
      code: version.code,
      name: version.name,
      unitType: version.unitType ?? 'ORGANIZATION',
      status: version.status,
      level: version.level ?? DEFAULT_LEVEL,
      effectiveDate: version.effectiveDate,
      endDate: version.endDate ?? undefined,
      isCurrent,
      createdAt: version.createdAt ?? version.effectiveDate,
      updatedAt: version.updatedAt ?? version.endDate ?? version.effectiveDate,
      parentCode: version.parentCode ?? undefined,
      description: version.description ?? undefined,
      codePath: version.codePath ?? undefined,
      namePath: version.namePath ?? undefined,
      lifecycleStatus: isCurrent ? 'CURRENT' : 'HISTORICAL',
      businessStatus: version.status === 'ACTIVE' ? 'ACTIVE' : 'INACTIVE',
      dataStatus: 'NORMAL',
      changeReason: '',
    }
  },
})
