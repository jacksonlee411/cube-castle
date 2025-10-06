import { logger } from '@/shared/utils/logger';
import { normalizeParentCode } from '@/shared/utils/organization-helpers';
import type { TemporalEditFormData } from '../TemporalEditForm';
import type { TimelineVersion } from '../TimelineComponent';
import {
  fetchHierarchyPaths,
  fetchOrganizationVersions,
} from './temporalMasterDetailApi';
import type {
  ApiOrganizationStatus,
  TemporalMasterDetailStateUpdaters,
} from './temporalMasterDetailTypes';

export type LoadVersionsFn = (
  isRetry?: boolean,
  focusRecordId?: string,
) => Promise<void>;

type LoadVersionsDeps = {
  organizationCode: string | null;
  setters: Pick<
    TemporalMasterDetailStateUpdaters,
    | 'setIsLoading'
    | 'setLoadingError'
    | 'setRetryCount'
    | 'setVersions'
    | 'setSelectedVersion'
    | 'setCurrentETag'
    | 'setFormMode'
    | 'setFormInitialData'
    | 'setDisplayPaths'
    | 'setSuccessMessage'
  >;
};

export type VersionSelectDeps = {
  setters: Pick<
    TemporalMasterDetailStateUpdaters,
    | 'setSelectedVersion'
    | 'setFormMode'
    | 'setFormInitialData'
    | 'setActiveTab'
    | 'setDisplayPaths'
  >;
};

const mapLifecycleStatusToApiStatus = (
  lifecycleStatus: TemporalEditFormData['lifecycleStatus'],
): ApiOrganizationStatus => {
  switch (lifecycleStatus) {
    case 'CURRENT':
      return 'ACTIVE';
    case 'PLANNED':
      return 'PLANNED';
    case 'HISTORICAL':
    case 'INACTIVE':
    case 'DELETED':
    default:
      return 'INACTIVE';
  }
};

export const createLoadVersions = ({
  organizationCode,
  setters,
}: LoadVersionsDeps): LoadVersionsFn => {
  const {
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
  } = setters;

  return async (isRetry = false, focusRecordId) => {
    if (!organizationCode) {
      setIsLoading(false);
      setVersions([]);
      setSelectedVersion(null);
      setCurrentETag(null);
      return;
    }

    try {
      setIsLoading(true);
      setLoadingError(null);
      if (!isRetry) {
        setRetryCount(0);
      }
      const { versions: loadedVersions, fallbackMessage } =
        await fetchOrganizationVersions(organizationCode);

      setVersions(loadedVersions);

      if (fallbackMessage) {
        setLoadingError(fallbackMessage);
        setTimeout(() => setLoadingError(null), 3000);
      } else if (isRetry) {
        setSuccessMessage('数据加载成功！');
        setTimeout(() => setSuccessMessage(null), 3000);
      }

      const preferredVersion = focusRecordId
        ? loadedVersions.find((v) => v.recordId === focusRecordId) || null
        : null;

      const currentVersion =
        preferredVersion ||
        loadedVersions.find((v) => v.isCurrent) ||
        loadedVersions.at(-1) ||
        null;

      const nextETag =
        loadedVersions.find((v) => v.isCurrent)?.recordId ?? null;
      setCurrentETag(nextETag);

      if (currentVersion) {
        setSelectedVersion(currentVersion);
        setFormMode('edit');
        setFormInitialData({
          name: currentVersion.name,
          unitType: currentVersion.unitType,
          status: currentVersion.status,
          lifecycleStatus: currentVersion.lifecycleStatus,
          description: currentVersion.description || '',
          parentCode: normalizeParentCode.forForm(currentVersion.parentCode),
          effectiveDate: currentVersion.effectiveDate,
        });

        try {
          const hierarchy = await fetchHierarchyPaths(currentVersion.code);
          setDisplayPaths(hierarchy);
        } catch (pathError) {
          logger.warn(
            '加载组织层级路径失败（忽略，不阻塞详情展示）:',
            pathError,
          );
          setDisplayPaths(null);
        }
      } else {
        setSelectedVersion(null);
        setFormInitialData(null);
        setDisplayPaths(null);
      }
    } catch (loadingException) {
      logger.error('Error loading temporal versions:', loadingException);
      const errorMessage =
        loadingException instanceof Error
          ? loadingException.message
          : '加载版本数据时发生未知错误';
      setLoadingError(errorMessage);
      setRetryCount((prev) => prev + 1);
    } finally {
      setIsLoading(false);
    }
  };
};

export const createHandleVersionSelect = ({
  setters,
}: VersionSelectDeps) => (version: TimelineVersion) => {
  const {
    setSelectedVersion,
    setFormMode,
    setFormInitialData,
    setActiveTab,
    setDisplayPaths,
  } = setters;

  setSelectedVersion(version);
  setFormMode('edit');
  setFormInitialData({
    name: version.name,
    unitType: version.unitType,
    status: version.status,
    lifecycleStatus: version.lifecycleStatus,
    description: version.description || '',
    parentCode: normalizeParentCode.forForm(version.parentCode),
    effectiveDate: version.effectiveDate,
  });
  setActiveTab('edit-history');

  (async () => {
    try {
      const hierarchy = await fetchHierarchyPaths(version.code);
      setDisplayPaths(hierarchy);
    } catch (pathError) {
      logger.warn(
        '加载组织层级路径失败（忽略，不阻塞详情展示）:',
        pathError,
      );
      setDisplayPaths(null);
    }
  })();
};

export { mapLifecycleStatusToApiStatus };
