import type { BaseSyntheticEvent, ChangeEvent, Dispatch, SetStateAction } from 'react';
import type { TemporalVersionPayload } from '@/shared/types/temporal';
import type { TimelineVersion } from '../TimelineComponent';
import type { TemporalEditFormData } from '../TemporalEditForm';
import { normalizeParentCode } from '@/shared/utils/organization-helpers';
import {
  DEFAULT_FORM_DATA,
  computeEditDateRange,
  formatDisplayDate,
  getCurrentMonthFirstDay,
} from './utils';
import type { TemporalParentErrorDetail } from './utils';
import type {
  DeleteConfirmMode,
  InlineNewVersionFormMode,
  InlineVersionRecord,
  InlineVersionSummary,
} from './types';

interface CreateFormActionsArgs {
  currentMode: InlineNewVersionFormMode;
  formData: TemporalEditFormData;
  setFormData: Dispatch<SetStateAction<TemporalEditFormData>>;
  errors: Record<string, string>;
  setErrors: Dispatch<SetStateAction<Record<string, string>>>;
  parentError: string;
  setParentError: Dispatch<SetStateAction<string>>;
  suggestedEffectiveDate?: string;
  setSuggestedEffectiveDate: Dispatch<SetStateAction<string | undefined>>;
  isEditingHistory: boolean;
  setIsEditingHistory: Dispatch<SetStateAction<boolean>>;
  originalHistoryData: InlineVersionRecord | null;
  setOriginalHistoryData: Dispatch<SetStateAction<InlineVersionRecord | null>>;
  selectedVersion: InlineVersionRecord | null;
  allVersions: InlineVersionSummary[] | null;
  onEditHistory?: (versionData: TemporalVersionPayload & { recordId: string }) => Promise<void>;
  onDeactivate?: (version: TimelineVersion) => Promise<void>;
  onSubmit: (data: TemporalEditFormData) => Promise<void>;
  setShowDeactivateConfirm: Dispatch<SetStateAction<DeleteConfirmMode>>;
  setIsDeactivating: Dispatch<SetStateAction<boolean>>;
  isDeactivating: boolean;
  setLoading: Dispatch<SetStateAction<boolean>>;
  setErrorMessage: Dispatch<SetStateAction<string | null>>;
  setSuccessMessage: Dispatch<SetStateAction<string | null>>;
}

const buildValidateForm = (
  currentMode: InlineNewVersionFormMode,
  formData: TemporalEditFormData,
  setErrors: React.Dispatch<React.SetStateAction<Record<string, string>>>,
  selectedVersion: InlineVersionRecord | null,
  allVersions: InlineVersionSummary[] | null
) => {
  return () => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = '组织名称是必填项';
    }

    if (!formData.effectiveDate) {
      newErrors.effectiveDate = '生效日期是必填项';
    } else if (currentMode === 'edit') {
      const { minDate, maxDate } = computeEditDateRange(selectedVersion, allVersions);
      const effectiveDate = new Date(formData.effectiveDate);

      if (minDate) {
        const minDateTime = new Date(minDate);
        if (effectiveDate < minDateTime) {
          newErrors.effectiveDate = `生效日期不能早于 ${formatDisplayDate(minDate)}（前一版本生效日期之后）`;
        }
      }

      if (!newErrors.effectiveDate && maxDate) {
        const maxDateTime = new Date(maxDate);
        if (effectiveDate > maxDateTime) {
          newErrors.effectiveDate = `生效日期不能晚于 ${formatDisplayDate(maxDate)}（下一版本生效日期之前）`;
        }
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };
};

interface TemporalParentUnavailableError {
  code?: string;
  message?: string;
  details?: TemporalParentErrorDetail[];
}

const isTemporalParentUnavailableError = (
  value: unknown,
): value is TemporalParentUnavailableError =>
  typeof value === 'object' && value !== null && 'code' in value;

const buildParentTemporalErrorHandler = (
  setParentError: React.Dispatch<React.SetStateAction<string>>,
  setSuggestedEffectiveDate: React.Dispatch<React.SetStateAction<string | undefined>>
) => {
  return (error: unknown): boolean => {
    if (!isTemporalParentUnavailableError(error) || error.code !== 'TEMPORAL_PARENT_UNAVAILABLE') {
      return false;
    }

    let message =
      typeof error.message === 'string'
        ? error.message
        : '上级组织在指定日期不可用';
    let suggested: string | undefined;

    const details = Array.isArray(error.details) ? error.details : [];
    const detail = details.find((item) => item?.code === 'TEMPORAL_PARENT_UNAVAILABLE');
    if (detail?.message && typeof detail.message === 'string') {
      message = detail.message;
    }
    const candidate = detail?.context?.suggestedDate;
    if (typeof candidate === 'string' && candidate.trim().length > 0) {
      suggested = candidate;
    }

    setParentError(message);
    setSuggestedEffectiveDate(suggested);
    return true;
  };
};

// 调用方派生的 TimelineVersion；命令服务仍以 ACTIVE/INACTIVE 持久化状态，
// 因此 lifecycleStatus/dataStatus 在此仅用于前端展示退化，不代表后端已支持五态。
const toTimelineVersion = (record: InlineVersionRecord): TimelineVersion => ({
  recordId: record.recordId,
  code: record.code,
  name: record.name,
  unitType: record.unitType,
  status: record.status,
  effectiveDate: record.effectiveDate,
  endDate: null,
  changeReason: '',
  isCurrent: true,
  createdAt: record.createdAt,
  updatedAt: record.updatedAt,
  description: record.description,
  level: record.level ?? 0,
  codePath: record.codePath ?? null,
  parentCode: record.parentCode,
  sortOrder: 0,
  lifecycleStatus: record.status === 'ACTIVE' ? 'CURRENT' : 'HISTORICAL',
  businessStatus: record.status === 'ACTIVE' ? 'ACTIVE' : 'INACTIVE',
  dataStatus: 'NORMAL',
});

export const createFormActions = (
  args: CreateFormActionsArgs
) => {
  const {
    currentMode,
    formData,
    setFormData,
    errors,
    setErrors,
    parentError,
    setParentError,
    suggestedEffectiveDate,
    setSuggestedEffectiveDate,
    isEditingHistory,
    setIsEditingHistory,
    originalHistoryData,
    setOriginalHistoryData,
    selectedVersion,
    allVersions,
    onEditHistory,
    onDeactivate,
    onSubmit,
    setShowDeactivateConfirm,
    setIsDeactivating,
    isDeactivating,
    setLoading,
    setErrorMessage,
    setSuccessMessage,
  } = args;

  const validateForm = buildValidateForm(
    currentMode,
    formData,
    setErrors,
    selectedVersion,
    allVersions
  );

  const handleParentTemporalError = buildParentTemporalErrorHandler(
    setParentError,
    setSuggestedEffectiveDate
  );

  const handleInputChange = (
      field: keyof TemporalEditFormData
    ) =>
      (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
        const { value } = event.target;
        setFormData((prev) => ({ ...prev, [field]: value }));
        if (errors[field]) {
          setErrors((prev) => ({ ...prev, [field]: '' }));
        }
      };

  const handleUnitTypeChange = (value: string) => {
    setFormData((prev) => ({ ...prev, unitType: value }));
    if (errors.unitType) {
      setErrors((prev) => ({ ...prev, unitType: '' }));
    }
  };

  const handleParentOrganizationChange = (parentCode: string | undefined) => {
    setFormData((prev) => ({ ...prev, parentCode: parentCode ?? '' }));
    setParentError('');
    setSuggestedEffectiveDate(undefined);
  };

  const handleParentOrganizationError = (message?: string) => {
    setParentError(message ?? '');
    if (!message) {
      setSuggestedEffectiveDate(undefined);
    }
  };

  const handleApplySuggestedEffectiveDate = () => {
    if (!suggestedEffectiveDate) {
      return;
    }
    setFormData((prev) => ({ ...prev, effectiveDate: suggestedEffectiveDate }));
    setSuggestedEffectiveDate(undefined);
    setParentError('');
  };

  const handleResetParentSelection = () => {
    setFormData((prev) => ({ ...prev, parentCode: '' }));
    setParentError('');
    setSuggestedEffectiveDate(undefined);
  };

  const handleStartInsertVersion = () => {
    const firstDayOfMonth = getCurrentMonthFirstDay();
    setFormData({
      name: selectedVersion?.name ?? '',
      unitType: selectedVersion?.unitType ?? 'ORGANIZATION_UNIT',
      lifecycleStatus: 'CURRENT',
      description: selectedVersion?.description ?? '',
      parentCode: selectedVersion?.parentCode ?? '',
      effectiveDate: firstDayOfMonth,
    });
    setErrors({});
    setParentError('');
    setSuggestedEffectiveDate(undefined);
    setOriginalHistoryData(null);
    setIsEditingHistory(true);
    setSuccessMessage(null);
    setErrorMessage(null);
  };

  const handleEditHistoryToggle = () => {
    if (!isEditingHistory && selectedVersion) {
      setOriginalHistoryData(selectedVersion);
      setFormData({
        name: selectedVersion.name,
        unitType: selectedVersion.unitType,
        lifecycleStatus:
          (selectedVersion.status as TemporalEditFormData['lifecycleStatus']) ?? 'CURRENT',
        description: selectedVersion.description ?? '',
        effectiveDate: new Date(selectedVersion.effectiveDate).toISOString().split('T')[0],
        parentCode: selectedVersion.parentCode ?? '',
      });
      setParentError('');
      setSuggestedEffectiveDate(undefined);
      setErrors({});
    }
    setIsEditingHistory((prev) => !prev);
  };

  const handleCancelEditHistory = () => {
    if (originalHistoryData) {
      setFormData({
        name: originalHistoryData.name,
        unitType: originalHistoryData.unitType,
        lifecycleStatus:
          (originalHistoryData.status as TemporalEditFormData['lifecycleStatus']) ?? 'PLANNED',
        description: originalHistoryData.description ?? '',
        effectiveDate: new Date(originalHistoryData.effectiveDate).toISOString().split('T')[0],
        parentCode: originalHistoryData.parentCode ?? '',
      });
    } else {
      setFormData({ ...DEFAULT_FORM_DATA, effectiveDate: getCurrentMonthFirstDay() });
    }
    setIsEditingHistory(false);
    setErrors({});
    setParentError('');
    setSuggestedEffectiveDate(undefined);
    setSuccessMessage(null);
    setErrorMessage(null);
  };

  const handleEditHistorySubmit = async () => {
    if (parentError) {
      setErrorMessage('请先修正上级组织选择');
      return;
    }

    if (!validateForm() || !onEditHistory || !originalHistoryData) {
      setErrorMessage('表单验证失败或缺少必要数据，请重试');
      return;
    }

    setErrorMessage(null);
    setSuccessMessage(null);
    setLoading(true);

    try {
      const updateData: TemporalVersionPayload & { recordId: string } = {
        recordId: originalHistoryData.recordId,
        name: formData.name,
        unitType: formData.unitType as TemporalVersionPayload['unitType'],
        lifecycleStatus: formData.lifecycleStatus,
        description: formData.description ?? null,
        effectiveDate: formData.effectiveDate,
        parentCode: normalizeParentCode.forAPI(formData.parentCode),
        operationReason: formData.changeReason,
        changeReason: formData.changeReason,
      };

      await onEditHistory(updateData);
      setIsEditingHistory(false);
      setSuccessMessage('历史记录修改成功！');
    } catch (error) {
      const message = error instanceof Error ? error.message : '修改失败，请重试';
      setErrorMessage(`修改历史记录失败: ${message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleDeactivateClick = () => {
    setShowDeactivateConfirm('record');
    setErrorMessage(null);
    setSuccessMessage(null);
  };

  const handleDeactivateConfirm = async () => {
    if (!onDeactivate || !selectedVersion || isDeactivating) {
      return;
    }

    try {
      setIsDeactivating(true);
      setErrorMessage(null);
      setSuccessMessage(null);
      await onDeactivate(toTimelineVersion(selectedVersion));
      setShowDeactivateConfirm(null);
      setSuccessMessage('版本已成功作废');
    } catch (error) {
      const message = error instanceof Error ? error.message : '删除失败，请重试';
      setErrorMessage(message);
      setShowDeactivateConfirm(null);
    } finally {
      setIsDeactivating(false);
    }
  };

  const handleDeactivateCancel = () => {
    setShowDeactivateConfirm(null);
  };

  const handleSubmit = async (event?: BaseSyntheticEvent) => {
    event?.preventDefault?.();

    setErrorMessage(null);
    setSuccessMessage(null);

    if (parentError) {
      setErrorMessage('请先修正上级组织选择');
      return;
    }

    if (!validateForm()) {
      setErrorMessage('请检查表单中的错误项并重新提交');
      return;
    }

    setLoading(true);

    try {
      await onSubmit(formData);
      setSuccessMessage(
        currentMode === 'create' ? '组织创建成功！' : '版本记录保存成功！'
      );
    } catch (error) {
      if (!handleParentTemporalError(error)) {
        setSuggestedEffectiveDate(undefined);
        const message = error instanceof Error ? error.message : '操作失败，请重试';
        setErrorMessage(
          `${currentMode === 'create' ? '创建组织失败' : '保存记录失败'}: ${message}`
        );
      }
    } finally {
      setLoading(false);
    }
  };

  return {
    handleInputChange,
    handleUnitTypeChange,
    handleParentOrganizationChange,
    handleParentOrganizationError,
    handleApplySuggestedEffectiveDate,
    handleResetParentSelection,
    handleStartInsertVersion,
    handleEditHistoryToggle,
    handleCancelEditHistory,
    handleEditHistorySubmit,
    handleDeactivateClick,
    handleDeactivateConfirm,
    handleDeactivateCancel,
    handleSubmit,
  };
};

export type InlineFormActions = ReturnType<typeof createFormActions>;
