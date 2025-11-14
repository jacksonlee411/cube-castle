import { useCallback, useEffect, useMemo, useState } from 'react';
import type { TabType } from '../TabNavigation';
// TODO-TEMPORARY(Plan 245): 引入统一详情 Hook（组织端），逐步用作快照/名称等来源（不替换现有版本加载）
import { useTemporalEntityDetail } from '@/shared/hooks/useTemporalEntityDetail';
import {
  createLoadVersions,
  createHandleVersionSelect,
  type LoadVersionsFn,
} from './temporalMasterDetailLoaders';
import {
  createHandleDeleteOrganization,
  createHandleDeleteVersion,
  createHandleStateMutationCompleted,
} from './temporalMasterDetailMutations';
import {
  createHandleFormSubmit,
  createHandleHistoryEditSubmit,
} from './temporalMasterDetailSubmissions';
import type {
  FormInitialData,
  TemporalMasterDetailState,
  TemporalMasterDetailViewProps,
  UseTemporalMasterDetailResult,
} from './temporalMasterDetailTypes';
import { useDeleteOrganization } from '@/shared/hooks/useOrganizationMutations';

export type {
  FormInitialData,
  TemporalMasterDetailHandlers,
  TemporalMasterDetailState,
  TemporalMasterDetailViewProps,
  UseTemporalMasterDetailResult,
} from './temporalMasterDetailTypes';

export const useTemporalMasterDetail = (
  options: TemporalMasterDetailViewProps,
): UseTemporalMasterDetailResult => {
  const {
    organizationCode,
    onBack,
    onCreateSuccess,
    isCreateMode = false,
  } = options;

  const [versions, setVersions] = useState<TemporalMasterDetailState['versions']>(
    [],
  );
  const [selectedVersion, setSelectedVersion] = useState<
    TemporalMasterDetailState['selectedVersion']
  >(null);
  const [isLoading, setIsLoading] = useState(() => Boolean(organizationCode));
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(
    null as TemporalMasterDetailState['showDeleteConfirm'],
  );
  const [isDeleting, setIsDeleting] = useState(false);
  const [loadingError, setLoadingError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [currentETag, setCurrentETag] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('edit-history');
  const [formMode, setFormMode] = useState<'create' | 'edit'>(() =>
    isCreateMode ? 'create' : 'edit',
  );
  const [formInitialData, setFormInitialData] = useState<FormInitialData | null>(
    null,
  );
  const [displayPaths, setDisplayPaths] = useState<
    TemporalMasterDetailState['displayPaths']
  >(null);
  const { mutateAsync: deleteOrganizationAsync } = useDeleteOrganization();
  // 统一详情数据（组织快照）：目前仅作为名称等信息的兜底来源，避免大范围重构
  const orgDetail = useTemporalEntityDetail(
    'organization',
    organizationCode ?? undefined,
    { enabled: Boolean(organizationCode) },
  );

  const notifySuccess = useCallback((message: string) => {
    setError(null);
    setSuccessMessage(message);
    setTimeout(() => setSuccessMessage(null), 3000);
  }, []);

  const notifyError = useCallback((message: string) => {
    setSuccessMessage(null);
    setError(message);
    setTimeout(() => setError(null), 5000);
  }, []);

  const setters = useMemo(
    () => ({
      setIsLoading,
      setLoadingError,
      setRetryCount,
      setVersions,
      setSelectedVersion,
      setCurrentETag,
      setFormMode,
      setFormInitialData,
      setDisplayPaths,
      setSuccessMessage,
      setError,
      setIsDeleting,
      setShowDeleteConfirm,
      setActiveTab,
      setIsSubmitting,
    }),
    [
      setIsLoading,
      setLoadingError,
      setRetryCount,
      setVersions,
      setSelectedVersion,
      setCurrentETag,
      setFormMode,
      setFormInitialData,
      setDisplayPaths,
      setSuccessMessage,
      setError,
      setIsDeleting,
      setShowDeleteConfirm,
      setActiveTab,
      setIsSubmitting,
    ],
  );

  const notifications = useMemo(
    () => ({ notifySuccess, notifyError }),
    [notifySuccess, notifyError],
  );

  const loadVersions: LoadVersionsFn = useMemo(
    () =>
      createLoadVersions({
        organizationCode,
        setters,
      }),
    [organizationCode, setters],
  );

  const handleStateMutationCompleted = useMemo(
    () =>
      createHandleStateMutationCompleted({
        setCurrentETag,
        loadVersions,
      }),
    [loadVersions],
  );

  const handleDeleteOrganization = useMemo(
    () =>
      createHandleDeleteOrganization({
        organizationCode,
        deleteOrganization: deleteOrganizationAsync,
        setters,
        loadVersions,
        notifySuccess,
        notifyError,
        onBack,
        currentETag,
      }),
    [
      organizationCode,
      deleteOrganizationAsync,
      setters,
      loadVersions,
      notifySuccess,
      notifyError,
      onBack,
      currentETag,
    ],
  );

  const handleDeleteVersion = useMemo(
    () =>
      createHandleDeleteVersion({
        organizationCode,
        isDeleting,
        selectedVersion,
        setters,
        loadVersions,
      }),
    [organizationCode, isDeleting, selectedVersion, setters, loadVersions],
  );

  const handleVersionSelect = useMemo(
    () =>
      createHandleVersionSelect({
        setters,
      }),
    [setters],
  );

  const handleFormSubmit = useMemo(
    () =>
      createHandleFormSubmit({
        notifications,
        setters,
        formArgs: { isCreateMode, organizationCode, onCreateSuccess },
        loadVersions,
      }),
    [
      notifications,
      setters,
      isCreateMode,
      organizationCode,
      onCreateSuccess,
      loadVersions,
    ],
  );

  const handleHistoryEditSubmit = useMemo(
    () =>
      createHandleHistoryEditSubmit({
        organizationCode,
        notifications,
        setters,
        loadVersions,
      }),
    [organizationCode, notifications, setters, loadVersions],
  );

  const handleHistoryEditClose = useCallback(() => {
    if (!isSubmitting) {
      if (onBack) {
        onBack();
      } else {
        setActiveTab('edit-history');
        setFormMode('create');
        setFormInitialData(null);
        setSelectedVersion(null);
      }
    }
  }, [isSubmitting, onBack]);

  useEffect(() => {
    if (!isCreateMode && organizationCode) {
      void loadVersions();
    }
  }, [loadVersions, isCreateMode, organizationCode]);

  useEffect(() => {
    if (organizationCode) {
      return;
    }

    setIsLoading(false);
    setFormMode('create');
    setFormInitialData(null);
    setSelectedVersion(null);
    setDisplayPaths(null);
  }, [organizationCode]);

  const currentOrganizationName = useMemo(() => {
    const currentVersion = versions.find((v) => v.isCurrent);
    // 优先使用版本中的当前名称；若不可用，则回退到统一 Hook 的 record.displayName
    return currentVersion?.name || (orgDetail.data?.record?.displayName ?? '');
  }, [versions, orgDetail.data?.record?.displayName]);

  const earliestVersion = useMemo(() => {
    if (versions.length === 0) {
      return null;
    }

    const nonDeleted = versions.filter((v) => v.status !== 'DELETED');
    if (nonDeleted.length === 0) {
      return null;
    }

    return nonDeleted[nonDeleted.length - 1];
  }, [versions]);

  const isEarliestVersionSelected = useMemo(
    () =>
      Boolean(
        selectedVersion &&
          earliestVersion &&
          selectedVersion.recordId === earliestVersion.recordId,
      ),
    [selectedVersion, earliestVersion],
  );

  const currentTimelineStatus = useMemo(() => {
    const current = versions.find((v) => v.isCurrent);
    return current?.status || selectedVersion?.status || (orgDetail.data?.record?.status as string | undefined);
  }, [versions, selectedVersion, orgDetail.data?.record?.status]);

  return [
    {
      versions,
      selectedVersion,
      isLoading,
      showDeleteConfirm,
      isDeleting,
      loadingError,
      successMessage,
      error,
      retryCount,
      isSubmitting,
      currentETag,
      activeTab,
      formMode,
      formInitialData,
      displayPaths,
      currentTimelineStatus,
      currentOrganizationName,
      earliestVersion,
      isEarliestVersionSelected,
    },
    {
      setShowDeleteConfirm,
      loadVersions,
      handleStateMutationCompleted,
      handleDeleteOrganization,
      handleDeleteVersion,
      handleVersionSelect,
      handleFormSubmit,
      handleHistoryEditClose,
      handleHistoryEditSubmit,
      setActiveTab,
      setCurrentETag,
      notifySuccess,
      notifyError,
    },
  ];
};
