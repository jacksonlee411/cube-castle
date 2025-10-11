import { logger } from '@/shared/utils/logger';
import type {
  DeleteOrganizationMutationResult,
  DeleteOrganizationVariables,
  OrganizationStateMutationResult,
} from '@/shared/hooks/useOrganizationMutations';
import type { TimelineVersion } from '../TimelineComponent';
import { deactivateOrganizationVersion } from './temporalMasterDetailApi';
import type {
  TemporalMasterDetailStateUpdaters,
} from './temporalMasterDetailTypes';
import type { LoadVersionsFn } from './temporalMasterDetailLoaders';

export const createHandleStateMutationCompleted = ({
  setCurrentETag,
  loadVersions,
}: {
  setCurrentETag: (etag: string | null) => void;
  loadVersions: LoadVersionsFn;
}) => async (
  _action: 'suspend' | 'activate',
  result: OrganizationStateMutationResult,
) => {
  if (result?.etag) {
    setCurrentETag(result.etag);
  }
  try {
    await loadVersions();
  } catch (mutationRefreshError) {
    logger.warn('状态变更后刷新失败:', mutationRefreshError);
  }
};

export const createHandleDeleteVersion = ({
  organizationCode,
  isDeleting,
  selectedVersion,
  setters,
  loadVersions,
}: {
  organizationCode: string | null;
  isDeleting: boolean;
  selectedVersion: TimelineVersion | null;
  setters: Pick<
    TemporalMasterDetailStateUpdaters,
    | 'setIsDeleting'
    | 'setVersions'
    | 'setSelectedVersion'
    | 'setCurrentETag'
    | 'setShowDeleteConfirm'
  >;
  loadVersions: LoadVersionsFn;
}) => async (version: TimelineVersion) => {
  const { setIsDeleting, setVersions, setSelectedVersion, setCurrentETag, setShowDeleteConfirm } = setters;

  if (!version || isDeleting) {
    return;
  }

  if (!organizationCode) {
    logger.warn('缺少组织编码，无法作废版本');
    return;
  }

  try {
    setIsDeleting(true);
    const timeline = await deactivateOrganizationVersion(
      organizationCode,
      version,
    );

    if (timeline) {
      setVersions(timeline);
      const current = timeline.find((v) => v.isCurrent) || timeline[0] || null;
      setSelectedVersion(current);
      setCurrentETag(current?.recordId ?? null);
    } else {
      try {
        await loadVersions();
      } catch (refreshError) {
        logger.warn('数据刷新失败，但删除操作已成功:', refreshError);
      }
    }

    setShowDeleteConfirm(null);
    if (
      !timeline &&
      selectedVersion?.effectiveDate === version.effectiveDate
    ) {
      setSelectedVersion(null);
    }
  } catch (error) {
    logger.error('Error deactivating version:', error);
    throw error;
  } finally {
    setIsDeleting(false);
  }
};

export const createHandleDeleteOrganization = ({
  organizationCode,
  deleteOrganization,
  setters,
  loadVersions,
  notifySuccess,
  notifyError,
  onBack,
  currentETag,
}: {
  organizationCode: string | null;
  deleteOrganization: (
    variables: DeleteOrganizationVariables,
  ) => Promise<DeleteOrganizationMutationResult>;
  setters: Pick<
    TemporalMasterDetailStateUpdaters,
    | 'setIsDeleting'
    | 'setVersions'
    | 'setSelectedVersion'
    | 'setCurrentETag'
    | 'setShowDeleteConfirm'
  >;
  loadVersions: LoadVersionsFn;
  notifySuccess: (message: string) => void;
  notifyError: (message: string) => void;
  onBack?: () => void;
  currentETag: string | null;
}) => async (version: TimelineVersion) => {
  const {
    setIsDeleting,
    setVersions,
    setSelectedVersion,
    setCurrentETag,
    setShowDeleteConfirm,
  } = setters;

  if (!organizationCode) {
    notifyError('无法删除组织：缺少组织编码');
    return;
  }

  try {
    setIsDeleting(true);

    await deleteOrganization({
      code: organizationCode,
      effectiveDate: version.effectiveDate,
      currentETag: currentETag ?? undefined,
    });

    setCurrentETag(null);
    setVersions([]);
    setSelectedVersion(null);
    setShowDeleteConfirm(null);

    notifySuccess('组织编码已删除');

    try {
      await loadVersions();
    } catch (refreshError) {
      logger.warn('组织删除后刷新版本失败（可忽略）:', refreshError);
    }

    if (onBack) {
      onBack();
    }
  } catch (error) {
    const message =
      error instanceof Error ? error.message : '删除组织失败，请稍后再试';
    notifyError(message);
    throw error;
  } finally {
    setIsDeleting(false);
  }
};
