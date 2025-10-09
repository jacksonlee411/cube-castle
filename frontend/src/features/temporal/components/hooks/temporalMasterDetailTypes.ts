import type { Dispatch, SetStateAction } from 'react';
import type { OrganizationStateMutationResult } from '@/shared/hooks/useOrganizationMutations';
import type { OrganizationRequest } from '@/shared/types/organization';
import type { TemporalVersionPayload } from '@/shared/types/temporal';
import type { TemporalEditFormData } from '../TemporalEditForm';
import type { TabType } from '../TabNavigation';
import type { TimelineVersion } from '../TimelineComponent';
import type { HierarchyPaths } from './temporalMasterDetailApi';

export interface TemporalMasterDetailViewProps {
  organizationCode: string | null;
  readonly?: boolean;
  onBack?: () => void;
  onCreateSuccess?: (newOrganizationCode: string) => void;
  isCreateMode?: boolean;
}

export interface FormInitialData {
  name: string;
  unitType: string;
  lifecycleStatus?: TimelineVersion['lifecycleStatus'];
  description?: string;
  parentCode?: string;
  effectiveDate?: string;
}

export interface TemporalMasterDetailState {
  versions: TimelineVersion[];
  selectedVersion: TimelineVersion | null;
  isLoading: boolean;
  showDeleteConfirm: TimelineVersion | null;
  isDeleting: boolean;
  loadingError: string | null;
  successMessage: string | null;
  error: string | null;
  retryCount: number;
  isSubmitting: boolean;
  currentETag: string | null;
  activeTab: TabType;
  formMode: 'create' | 'edit';
  formInitialData: FormInitialData | null;
  displayPaths: HierarchyPaths | null;
  currentTimelineStatus: string | undefined;
  currentOrganizationName: string;
  earliestVersion: TimelineVersion | null;
  isEarliestVersionSelected: boolean;
}

export interface TemporalMasterDetailHandlers {
  setShowDeleteConfirm: (version: TimelineVersion | null) => void;
  loadVersions: (isRetry?: boolean, focusRecordId?: string) => Promise<void>;
  handleStateMutationCompleted: (
    action: 'suspend' | 'activate',
    result: OrganizationStateMutationResult,
  ) => Promise<void>;
  handleDeleteVersion: (version: TimelineVersion) => Promise<void>;
  handleDeleteOrganization: (version: TimelineVersion) => Promise<void>;
  handleVersionSelect: (version: TimelineVersion) => void;
  handleFormSubmit: (formData: TemporalEditFormData) => Promise<void>;
  handleHistoryEditClose: () => void;
  handleHistoryEditSubmit: (
    versionData: TemporalVersionPayload & { recordId: string },
  ) => Promise<void>;
  setActiveTab: (tab: TabType) => void;
  setCurrentETag: (etag: string | null) => void;
  notifySuccess: (message: string) => void;
  notifyError: (message: string) => void;
}

export type UseTemporalMasterDetailResult = [
  TemporalMasterDetailState,
  TemporalMasterDetailHandlers,
];

export type TemporalMasterDetailNotifications = {
  notifySuccess: (message: string) => void;
  notifyError: (message: string) => void;
};

export type TemporalMasterDetailFormSubmitArgs = {
  isCreateMode: boolean;
  organizationCode: string | null;
  onCreateSuccess?: (code: string) => void;
};

export type TemporalMasterDetailStateUpdaters = {
  setIsLoading: (value: boolean) => void;
  setLoadingError: (value: string | null) => void;
  setRetryCount: Dispatch<SetStateAction<number>>;
  setVersions: Dispatch<SetStateAction<TimelineVersion[]>>;
  setSelectedVersion: Dispatch<SetStateAction<TimelineVersion | null>>;
  setCurrentETag: (etag: string | null) => void;
  setFormMode: Dispatch<SetStateAction<'create' | 'edit'>>;
  setFormInitialData: Dispatch<SetStateAction<FormInitialData | null>>;
  setDisplayPaths: Dispatch<SetStateAction<HierarchyPaths | null>>;
  setSuccessMessage: (value: string | null) => void;
  setError: (value: string | null) => void;
  setIsDeleting: (value: boolean) => void;
  setShowDeleteConfirm: (version: TimelineVersion | null) => void;
  setActiveTab: Dispatch<SetStateAction<TabType>>;
  setIsSubmitting: (value: boolean) => void;
};

export type TemporalMasterDetailDeps = TemporalMasterDetailViewProps & {
  notifications: TemporalMasterDetailNotifications;
  setters: TemporalMasterDetailStateUpdaters;
  loadVersions: (isRetry?: boolean, focusRecordId?: string) => Promise<void>;
};

export type OrganizationRequestPayload = OrganizationRequest;

export type TemporalHistoryUpdatePayload = TemporalVersionPayload & {
  recordId: string;
};
