import type { TemporalVersionPayload } from '@/shared/types/temporal';
import type { TimelineVersion } from '../TimelineComponent';
import type { TemporalEditFormData } from '../TemporalEditForm';

export type InlineNewVersionFormMode = 'create' | 'edit';
export type DeleteConfirmMode = 'record' | 'organization' | null;

export interface InlineVersionRecord {
  recordId: string;
  createdAt: string;
  updatedAt: string;
  code: string;
  name: string;
  unitType: string;
  status: string;
  effectiveDate: string;
  description?: string;
  parentCode?: string;
  level?: number;
  codePath?: string | null;
}

export interface InlineVersionSummary {
  recordId: string;
  effectiveDate: string;
  endDate?: string | null;
  isCurrent: boolean;
}

export interface InlineHierarchyPaths {
  codePath: string;
  namePath: string;
}

export interface InlineNewVersionFormProps {
  organizationCode: string | null;
  onSubmit: (data: TemporalEditFormData) => Promise<void>;
  onCancel: () => void;
  isSubmitting?: boolean;
  mode?: InlineNewVersionFormMode;
  initialData?: InlineNewVersionInitialData | null;
  selectedVersion?: InlineVersionRecord | null;
  allVersions?: InlineVersionSummary[] | null;
  onEditHistory?: (versionData: TemporalVersionPayload & { recordId: string }) => Promise<void>;
  onDeactivate?: (version: TimelineVersion) => Promise<void>;
  onInsertRecord?: (data: TemporalEditFormData) => Promise<void>;
  activeTab?: 'edit-history' | 'new-version' | 'audit-history';
  onTabChange?: (tab: 'edit-history' | 'new-version' | 'audit-history') => void;
  hierarchyPaths?: InlineHierarchyPaths | null;
  canDeleteOrganization?: boolean;
  onDeleteOrganization?: () => Promise<void>;
  isDeletingOrganization?: boolean;
}

export interface InlineNewVersionInitialData {
  name: string;
  unitType: string;
  status: string;
  lifecycleStatus?: string;
  description?: string;
  parentCode?: string;
  effectiveDate?: string;
}
