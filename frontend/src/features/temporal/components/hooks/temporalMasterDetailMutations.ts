import { logger } from '@/shared/utils/logger';
import type { OrganizationStateMutationResult } from '@/shared/hooks/useOrganizationMutations';
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
