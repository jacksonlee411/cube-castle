import type { StandardObject } from '@/generated/graphql-types';
import type { TimelineVersion } from '@/features/temporal/components/TimelineComponent';

type Payload = Record<string, unknown>;

const getString = (payload: Payload, key: string): string => {
  const value = payload[key];
  if (typeof value === 'string') {
    return value;
  }
  if (value === null || value === undefined) {
    return '';
  }
  return String(value);
};

const getNumber = (payload: Payload, key: string): number => {
  const value = payload[key];
  if (typeof value === 'number') {
    return value;
  }
  const parsed = Number(value);
  return Number.isNaN(parsed) ? 0 : parsed;
};

const extractParentCode = (payload: Payload): string | undefined => {
  const parent = payload.parentCode ?? payload.parent_code;
  if (!parent) return undefined;
  const value = typeof parent === 'string' ? parent : String(parent);
  return value.trim() || undefined;
};

const deriveLifecycleStatus = (version: StandardObject['version']): 'CURRENT' | 'HISTORICAL' => {
  if (version.isCurrent) {
    return 'CURRENT';
  }
  return 'HISTORICAL';
};

export const standardObjectAdapter = {
  toTimelineVersion(aggregate: StandardObject): TimelineVersion {
    const { kernel, version } = aggregate;
    const payload = (version.payload || {}) as Payload;
    const recordId = version.versionCode || `${kernel.code}-${version.effectiveDate}`;
    const labels = kernel.labels || {};

    return {
      recordId,
      code: kernel.code,
      name: getString(payload, 'name') || kernel.displayName,
      unitType: typeof labels.unitType === 'string' ? labels.unitType : 'UNKNOWN',
      status: kernel.status,
      effectiveDate: version.effectiveDate,
      endDate: version.endDate ?? null,
      changeReason: getString(payload, 'changeReason') || getString(version.auditTrail || {}, 'reason'),
      isCurrent: version.isCurrent,
      createdAt: version.createdAt || kernel.createdAt || new Date().toISOString(),
      updatedAt: version.updatedAt || kernel.updatedAt || new Date().toISOString(),
      description: getString(payload, 'description'),
      level: getNumber(payload, 'level'),
      codePath: getString(payload, 'codePath') || null,
      namePath: getString(payload, 'namePath') || null,
      parentCode: extractParentCode(payload),
      sortOrder: getNumber(payload, 'sortOrder'),
      lifecycleStatus: deriveLifecycleStatus(version),
      businessStatus: kernel.status === 'ACTIVE' ? 'ACTIVE' : 'INACTIVE',
      dataStatus: 'NORMAL',
      suspendedAt: getString(version.auditTrail || {}, 'suspendedAt') || null,
      suspendedBy: getString(version.auditTrail || {}, 'suspendedBy') || null,
      suspensionReason: getString(version.auditTrail || {}, 'suspensionReason') || null,
      deletedAt: getString(version.auditTrail || {}, 'deletedAt') || null,
      deletedBy: getString(version.auditTrail || {}, 'deletedBy') || null,
      deletionReason: getString(version.auditTrail || {}, 'deletionReason') || null,
    };
  },

  toTimelineVersions(aggregates: StandardObject[]): TimelineVersion[] {
    return aggregates.map(standardObjectAdapter.toTimelineVersion).sort((a, b) => {
      return new Date(b.effectiveDate).getTime() - new Date(a.effectiveDate).getTime();
    });
  },
};

export type StandardObjectTimeline = ReturnType<typeof standardObjectAdapter.toTimelineVersions>;
