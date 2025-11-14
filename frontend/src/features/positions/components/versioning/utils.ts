import { getPositionStatusMeta } from '@/features/temporal/entity/statusMeta';
import type { PositionRecord } from '@/shared/types/positions';

export const getVersionKey = (version: PositionRecord): string =>
  version.recordId ?? `${version.code}-${version.updatedAt}`;

const POSITION_TYPE_LABELS: Record<string, string> = {
  REGULAR: '正式职位',
  TEMPORARY: '临时职位',
  CONTRACTOR: '合同工',
};

const EMPLOYMENT_TYPE_LABELS: Record<string, string> = {
  FULL_TIME: '全职',
  PART_TIME: '兼职',
  INTERN: '实习',
};

const formatValue = (record: PositionRecord, key: keyof PositionRecord): string => {
  const raw = record[key] as unknown;

  if (raw === null || raw === undefined) {
    return '—';
  }

  if (key === 'status' && typeof raw === 'string') {
    return getPositionStatusMeta(raw).label;
  }

  if (key === 'positionType' && typeof raw === 'string') {
    return POSITION_TYPE_LABELS[raw] ? `${POSITION_TYPE_LABELS[raw]} (${raw})` : raw;
  }

  if (key === 'employmentType' && typeof raw === 'string') {
    return EMPLOYMENT_TYPE_LABELS[raw] ? `${EMPLOYMENT_TYPE_LABELS[raw]} (${raw})` : raw;
  }

  if (key === 'isCurrent' || key === 'isFuture') {
    return raw ? '是' : '否';
  }

  if (typeof raw === 'number') {
    return Number.isFinite(raw) ? String(raw) : '—';
  }

  if (typeof raw === 'boolean') {
    return raw ? '是' : '否';
  }

  return String(raw);
};

const CSV_FIELD_DEFINITIONS: Array<{ key: keyof PositionRecord; label: string }> = [
  { key: 'title', label: '职位名称' },
  { key: 'status', label: '职位状态' },
  { key: 'organizationCode', label: '组织编码' },
  { key: 'organizationName', label: '组织名称' },
  { key: 'jobFamilyGroupCode', label: '职类编码' },
  { key: 'jobFamilyGroupName', label: '职类名称' },
  { key: 'jobFamilyCode', label: '职种编码' },
  { key: 'jobFamilyName', label: '职种名称' },
  { key: 'jobRoleCode', label: '职务编码' },
  { key: 'jobRoleName', label: '职务名称' },
  { key: 'jobLevelCode', label: '职级编码' },
  { key: 'jobLevelName', label: '职级名称' },
  { key: 'positionType', label: '职位类型' },
  { key: 'employmentType', label: '雇佣方式' },
  { key: 'gradeLevel', label: '职级等级' },
  { key: 'headcountCapacity', label: '编制容量' },
  { key: 'headcountInUse', label: '编制占用' },
  { key: 'availableHeadcount', label: '可用编制' },
  { key: 'reportsToPositionCode', label: '汇报职位' },
  { key: 'effectiveDate', label: '生效日期' },
  { key: 'endDate', label: '结束日期' },
  { key: 'isCurrent', label: '当前版本' },
  { key: 'isFuture', label: '计划版本' },
  { key: 'createdAt', label: '创建时间' },
  { key: 'updatedAt', label: '更新时间' },
];

const CSV_HEADER = CSV_FIELD_DEFINITIONS.map(item => item.label);

const escapeCsvValue = (value: string): string => `"${value.replace(/"/g, '""')}"`;

export const buildVersionsCsv = (versions: PositionRecord[]): string => {
  const rows = versions.map(version =>
    CSV_FIELD_DEFINITIONS.map(definition => formatValue(version, definition.key)),
  );

  return [
    CSV_HEADER.map(escapeCsvValue).join(','),
    ...rows.map(row => row.map(escapeCsvValue).join(',')),
  ].join('\n');
};
