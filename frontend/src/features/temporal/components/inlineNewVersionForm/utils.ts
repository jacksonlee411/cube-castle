import type { OrganizationStatus } from '../../../../shared/components/StatusBadge';
import type { TemporalEditFormData } from '../TemporalEditForm';
import type {
  InlineHierarchyPaths,
  InlineNewVersionInitialData,
  InlineVersionRecord,
  InlineVersionSummary,
} from './types';

export const mapLifecycleStatusToOrganizationStatus = (
  lifecycleStatus: string
): OrganizationStatus => {
  switch (lifecycleStatus) {
    case 'CURRENT':
    case 'ACTIVE':
      return 'ACTIVE';
    case 'INACTIVE':
      return 'INACTIVE';
    case 'PLANNED':
      return 'PLANNED';
    default:
      return 'ACTIVE';
  }
};

export const getCurrentMonthFirstDay = (): string => {
  const now = new Date();
  const year = now.getFullYear();
  const month = now.getMonth() + 1;
  const paddedMonth = month.toString().padStart(2, '0');
  return `${year}-${paddedMonth}-01`;
};

export const computeEditDateRange = (
  selectedVersion: InlineVersionRecord | null | undefined,
  versions: InlineVersionSummary[] | null | undefined
): { minDate: string | null; maxDate: string | null } => {
  if (!selectedVersion || !versions || versions.length === 0) {
    return { minDate: null, maxDate: null };
  }

  const sorted = [...versions].sort(
    (a, b) => new Date(a.effectiveDate).getTime() - new Date(b.effectiveDate).getTime()
  );

  const currentIndex = sorted.findIndex((item) => item.recordId === selectedVersion.recordId);
  if (currentIndex === -1) {
    return { minDate: null, maxDate: null };
  }

  const previous = currentIndex > 0 ? sorted[currentIndex - 1] : null;
  const next = currentIndex < sorted.length - 1 ? sorted[currentIndex + 1] : null;

  let minDate: string | null = null;
  if (previous) {
    const candidate = new Date(previous.effectiveDate);
    candidate.setDate(candidate.getDate() + 1);
    minDate = candidate.toISOString().split('T')[0];
  }

  let maxDate: string | null = null;
  if (next) {
    const candidate = new Date(next.effectiveDate);
    candidate.setDate(candidate.getDate() - 1);
    maxDate = candidate.toISOString().split('T')[0];
  }

  return { minDate, maxDate };
};

export const formatDisplayDate = (date: string): string => {
  return new Date(date).toLocaleDateString('zh-CN');
};

export const DEFAULT_FORM_DATA: TemporalEditFormData = {
  name: '',
  unitType: 'DEPARTMENT',
  lifecycleStatus: 'PLANNED',
  description: '',
  effectiveDate: getCurrentMonthFirstDay(),
  parentCode: '',
};

export const normalizeInitialData = (
  data: InlineNewVersionInitialData | null | undefined,
  fallbackDate: string
): TemporalEditFormData => {
  if (!data || !('name' in data)) {
    return {
      ...DEFAULT_FORM_DATA,
      effectiveDate: fallbackDate,
    };
  }

  const initial = data;

  return {
    name: initial.name,
    unitType: initial.unitType,
    lifecycleStatus:
      (initial.status as TemporalEditFormData['lifecycleStatus']) ?? DEFAULT_FORM_DATA.lifecycleStatus,
    description: initial.description ?? '',
    effectiveDate: initial.effectiveDate
      ? new Date(initial.effectiveDate).toISOString().split('T')[0]
      : fallbackDate,
    parentCode: initial.parentCode ?? '',
  };
};

export const deriveCodePath = (
  selectedVersion: InlineVersionRecord | null | undefined,
  hierarchyPaths: InlineHierarchyPaths | null | undefined
): string => {
  if (selectedVersion?.path) {
    return selectedVersion.path;
  }
  if (hierarchyPaths?.codePath) {
    return hierarchyPaths.codePath;
  }
  return '';
};

export const deriveNamePath = (
  hierarchyPaths: InlineHierarchyPaths | null | undefined
): string => {
  if (hierarchyPaths?.namePath) {
    return hierarchyPaths.namePath;
  }
  return '';
};

export interface TemporalParentErrorDetail {
  code?: string;
  message?: string;
  context?: {
    suggestedDate?: string;
  };
}

type InlineNewVersionInitialData = {
  name: string;
  unitType: string;
  status: string;
  description?: string;
  parentCode?: string;
  effectiveDate?: string;
};
