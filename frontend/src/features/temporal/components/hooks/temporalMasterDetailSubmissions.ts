import { logger } from '@/shared/utils/logger';
import { normalizeParentCode } from '@/shared/utils/organization-helpers';
import type { TemporalVersionPayload } from '@/shared/types/temporal';
import type { OrganizationRequest } from '@/shared/types';
import type { TemporalEditFormData } from '../TemporalEditForm';
import type { TimelineVersion } from '../TimelineComponent';
import {
  createOrganizationUnit,
  createTemporalVersion,
  updateHistoryRecord,
} from './temporalMasterDetailApi';
import { type LoadVersionsFn } from './temporalMasterDetailLoaders';
import type {
  TemporalHistoryUpdatePayload,
  TemporalMasterDetailFormSubmitArgs,
  TemporalMasterDetailNotifications,
  TemporalMasterDetailStateUpdaters,
} from './temporalMasterDetailTypes';

const allowedLifecycleStatuses: ReadonlyArray<TimelineVersion['lifecycleStatus']> = [
  'PLANNED',
  'CURRENT',
  'HISTORICAL',
];

const normalizeLifecycleStatus = (
  value?: string | null,
): TimelineVersion['lifecycleStatus'] => {
  if (value && allowedLifecycleStatuses.includes(value as TimelineVersion['lifecycleStatus'])) {
    return value as TimelineVersion['lifecycleStatus'];
  }
  return 'HISTORICAL';
};

export const createHandleFormSubmit = ({
  notifications,
  setters,
  formArgs,
  loadVersions,
}: {
  notifications: TemporalMasterDetailNotifications;
  setters: Pick<
    TemporalMasterDetailStateUpdaters,
    | 'setIsSubmitting'
    | 'setActiveTab'
    | 'setFormMode'
    | 'setFormInitialData'
  >;
  formArgs: TemporalMasterDetailFormSubmitArgs;
  loadVersions: LoadVersionsFn;
}) =>
  async (formData: TemporalEditFormData) => {
    const { notifyError, notifySuccess } = notifications;
    const { setIsSubmitting, setActiveTab, setFormMode, setFormInitialData } = setters;
    const { isCreateMode, onCreateSuccess, organizationCode } = formArgs;

    setIsSubmitting(true);
    try {
      if (isCreateMode) {
        const requestBody: OrganizationRequest = {
          name: formData.name,
          unitType: formData.unitType as OrganizationRequest['unitType'],
          description: formData.description || '',
          parentCode: normalizeParentCode.forAPI(formData.parentCode),
          effectiveDate: formData.effectiveDate,
          changeReason: formData.changeReason,
        };

        logger.info('提交创建组织请求:', requestBody);

        const newOrganizationCode = await createOrganizationUnit(requestBody);

        if (newOrganizationCode && onCreateSuccess) {
          onCreateSuccess(newOrganizationCode);
          return;
        }

        logger.error('创建成功但未返回组织编码');
        notifyError('创建成功，但未能获取新组织编码，请手动刷新页面');
      } else {
        if (!organizationCode) {
          notifyError('缺少组织编码，无法创建时态版本');
          return;
        }

        const versionPayload: TemporalVersionPayload = {
          name: formData.name,
          unitType: formData.unitType as TemporalVersionPayload['unitType'],
          parentCode: normalizeParentCode.forAPI(formData.parentCode),
          description: formData.description || null,
          effectiveDate: formData.effectiveDate,
          lifecycleStatus: formData.lifecycleStatus,
        };

        const trimmedReason = formData.changeReason?.trim();
        if (trimmedReason) {
          versionPayload.operationReason = trimmedReason;
        }

        await createTemporalVersion(organizationCode, versionPayload);

        await loadVersions();
        setActiveTab('edit-history');
        setFormMode('create');
        setFormInitialData(null);
        notifySuccess('时态版本创建成功！');
      }
    } catch (submissionError) {
      logger.error(
        isCreateMode ? '创建组织失败:' : '创建时态版本失败:',
        submissionError,
      );

      let errorMessage = isCreateMode ? '创建组织失败' : '创建时态版本失败';
      if (submissionError instanceof Error) {
        errorMessage = submissionError.message;
      } else if (submissionError && typeof submissionError === 'string') {
        errorMessage = submissionError;
      }

      notifyError(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

export const createHandleHistoryEditSubmit = ({
  organizationCode,
  notifications,
  setters,
  loadVersions,
}: {
  organizationCode: string | null;
  notifications: TemporalMasterDetailNotifications;
  setters: Pick<
    TemporalMasterDetailStateUpdaters,
    | 'setIsSubmitting'
    | 'setActiveTab'
    | 'setFormMode'
    | 'setFormInitialData'
    | 'setSelectedVersion'
  >;
  loadVersions: LoadVersionsFn;
}) =>
  async (updateData: TemporalHistoryUpdatePayload) => {
    const { notifyError, notifySuccess } = notifications;
    const {
      setIsSubmitting,
      setActiveTab,
      setFormMode,
      setFormInitialData,
      setSelectedVersion,
    } = setters;

    setIsSubmitting(true);
    try {
      const lifecycleStatusRaw = updateData.lifecycleStatus ?? 'CURRENT';
      const lifecycleStatus = normalizeLifecycleStatus(lifecycleStatusRaw);

      if (!organizationCode) {
        notifyError('缺少组织编码，无法更新历史记录');
        setIsSubmitting(false);
        return;
      }

      await updateHistoryRecord(
        organizationCode,
        updateData.recordId,
        {
          name: updateData.name,
        unitType: updateData.unitType,
        lifecycleStatus: lifecycleStatusRaw,
        description: updateData.description ?? null,
        effectiveDate: updateData.effectiveDate,
        parentCode: normalizeParentCode.forAPI(updateData.parentCode),
        changeReason: '通过组织详情页面修改历史记录',
        operationReason: updateData.operationReason,
      },
    );

      await loadVersions(false, updateData.recordId);
      setActiveTab('edit-history');
      setFormMode('edit');
      setFormInitialData({
        name: updateData.name as string,
        unitType: updateData.unitType,
        status: (updateData.status ?? 'ACTIVE') as TimelineVersion['status'],
        lifecycleStatus,
        description: (updateData.description as string) || '',
        parentCode: normalizeParentCode.forForm(
          updateData.parentCode as string,
        ),
        effectiveDate: updateData.effectiveDate as string,
      });
      setSelectedVersion((prev) => {
        if (!prev || prev.recordId !== updateData.recordId) {
          return prev;
        }
        return {
          ...prev,
          name: updateData.name as string,
          unitType: updateData.unitType as string,
          lifecycleStatus,
          description: (updateData.description as string) || undefined,
          parentCode: (updateData.parentCode as string) || undefined,
          effectiveDate: updateData.effectiveDate as string,
        };
      });
      notifySuccess('历史记录修改成功！');
    } catch (historyError) {
      logger.error('修改历史记录失败:', historyError);
      notifyError('修改失败，请检查网络连接');
    } finally {
      setIsSubmitting(false);
    }
  };
