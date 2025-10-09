import { useCallback, useEffect, useMemo, useState } from 'react';
import type { TemporalEditFormData } from '../TemporalEditForm';
import {
  DEFAULT_FORM_DATA,
  deriveCodePath,
  deriveNamePath,
  getCurrentMonthFirstDay,
  normalizeInitialData,
} from './utils';
import { createFormActions } from './formActions';
import type {
  DeleteConfirmMode,
  InlineNewVersionFormMode,
  InlineNewVersionFormProps,
  InlineVersionRecord,
} from './types';
import type { InlineSubmitEvent } from './FormActions';

export interface UseInlineNewVersionFormResult {
  formData: TemporalEditFormData;
  errors: Record<string, string>;
  parentError: string;
  suggestedEffectiveDate?: string;
  isEditingHistory: boolean;
  originalHistoryData: InlineVersionRecord | null;
  deleteConfirmMode: DeleteConfirmMode;
  isDeactivating: boolean;
  deleteProcessing: boolean;
  loading: boolean;
  errorMessage: string | null;
  successMessage: string | null;
  currentMode: InlineNewVersionFormMode;
  levelDisplay?: number;
  codePathDisplay: string;
  namePathDisplay: string;
  handleInputChange: (
    field: keyof TemporalEditFormData
  ) => (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => void;
  handleParentOrganizationChange: (parentCode: string | undefined) => void;
  handleParentOrganizationError: (message?: string) => void;
  handleApplySuggestedEffectiveDate: () => void;
  handleResetParentSelection: () => void;
  handleSubmit: (event?: InlineSubmitEvent) => Promise<void> | void;
  handleEditHistoryToggle: () => void;
  handleCancelEditHistory: () => void;
  handleEditHistorySubmit: () => Promise<void>;
  handleDeactivateClick: () => void;
  handleDeleteOrganizationClick: () => void;
  handleConfirmDelete: () => Promise<void>;
  handleDeactivateCancel: () => void;
  handleStartInsertVersion: () => void;
  handleUnitTypeChange: (value: string) => void;
}

const useInlineNewVersionForm = (props: InlineNewVersionFormProps): UseInlineNewVersionFormResult => {
  const {
    onSubmit,
    mode = 'create',
    initialData,
    selectedVersion = null,
    allVersions = null,
    onEditHistory,
    onDeactivate,
    onInsertRecord: _onInsertRecord,
    activeTab: _activeTab,
    onTabChange: _onTabChange,
    hierarchyPaths = null,
    canDeleteOrganization = false,
    onDeleteOrganization,
    isDeletingOrganization = false,
  } = props;

  const [formData, setFormData] = useState<TemporalEditFormData>({ ...DEFAULT_FORM_DATA });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [parentError, setParentError] = useState<string>('');
  const [suggestedEffectiveDate, setSuggestedEffectiveDate] = useState<string | undefined>(undefined);
  const [isEditingHistory, setIsEditingHistory] = useState(false);
  const [originalHistoryData, setOriginalHistoryData] = useState<InlineVersionRecord | null>(null);
  const [deleteConfirmMode, setDeleteConfirmMode] = useState<DeleteConfirmMode>(null);
  const [isDeactivating, setIsDeactivating] = useState(false);
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const currentMode: InlineNewVersionFormMode = mode ?? 'create';

  const levelDisplay = selectedVersion?.level;

  const codePathDisplay = useMemo(
    () => deriveCodePath(selectedVersion, hierarchyPaths),
    [hierarchyPaths, selectedVersion]
  );

  const namePathDisplay = useMemo(
    () => deriveNamePath(hierarchyPaths),
    [hierarchyPaths]
  );

  useEffect(() => {
    const firstDayOfMonth = getCurrentMonthFirstDay();

    if (currentMode === 'edit' && initialData) {
      setFormData(normalizeInitialData(initialData, firstDayOfMonth));
      if (selectedVersion) {
        setOriginalHistoryData(selectedVersion);
        setIsEditingHistory(false);
      }
    } else {
      setFormData({
        ...DEFAULT_FORM_DATA,
        effectiveDate: firstDayOfMonth,
      });
      setOriginalHistoryData(null);
      setIsEditingHistory(false);
    }

    setErrors({});
    setParentError('');
    setSuggestedEffectiveDate(undefined);
    setSuccessMessage(null);
    setErrorMessage(null);
  }, [currentMode, initialData, selectedVersion]);

  useEffect(() => {
    if (!successMessage) {
      return;
    }
    const timer = setTimeout(() => setSuccessMessage(null), 3000);
    return () => clearTimeout(timer);
  }, [successMessage]);

  useEffect(() => {
    if (!errorMessage) {
      return;
    }
    const timer = setTimeout(() => setErrorMessage(null), 5000);
    return () => clearTimeout(timer);
  }, [errorMessage]);

  const actions = useMemo(
    () =>
      createFormActions({
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
        setShowDeactivateConfirm: setDeleteConfirmMode,
        setIsDeactivating,
        isDeactivating,
        setLoading,
        setErrorMessage,
        setSuccessMessage,
      }),
    [
      allVersions,
      currentMode,
      errors,
      formData,
      isDeactivating,
      isEditingHistory,
      onDeactivate,
      onEditHistory,
      onSubmit,
      originalHistoryData,
      parentError,
      selectedVersion,
      suggestedEffectiveDate,
    ]
  );

  const {
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
  } = actions;

  const handleDeleteOrganizationClick = useCallback(() => {
    if (!canDeleteOrganization || isDeactivating || isDeletingOrganization) {
      return;
    }
    setDeleteConfirmMode('organization');
    setErrorMessage(null);
    setSuccessMessage(null);
  }, [
    canDeleteOrganization,
    isDeactivating,
    isDeletingOrganization,
    setDeleteConfirmMode,
    setErrorMessage,
    setSuccessMessage,
  ]);

  const handleDeleteOrganizationConfirm = useCallback(async () => {
    if (!onDeleteOrganization) {
      setErrorMessage('当前环境暂不支持删除组织编码');
      setDeleteConfirmMode(null);
      return;
    }

    try {
      setIsDeactivating(true);
      setErrorMessage(null);
      setSuccessMessage(null);
      await onDeleteOrganization();
      setDeleteConfirmMode(null);
      setSuccessMessage('组织编码已删除');
    } catch (error) {
      const message = error instanceof Error ? error.message : '删除失败，请重试';
      setErrorMessage(message);
      setDeleteConfirmMode(null);
    } finally {
      setIsDeactivating(false);
    }
  }, [
    onDeleteOrganization,
    setDeleteConfirmMode,
    setErrorMessage,
    setIsDeactivating,
    setSuccessMessage,
  ]);

  const handleConfirmDelete = useCallback(async () => {
    if (deleteConfirmMode === 'organization') {
      await handleDeleteOrganizationConfirm();
      return;
    }
    if (deleteConfirmMode === 'record') {
      await handleDeactivateConfirm();
    }
  }, [deleteConfirmMode, handleDeactivateConfirm, handleDeleteOrganizationConfirm]);

  const deleteProcessing = isDeactivating || isDeletingOrganization;

  return {
    formData,
    errors,
    parentError,
    suggestedEffectiveDate,
    isEditingHistory,
    originalHistoryData,
    deleteConfirmMode,
    isDeactivating,
    deleteProcessing,
    loading,
    errorMessage,
    successMessage,
    currentMode,
    levelDisplay,
    codePathDisplay,
    namePathDisplay,
    handleInputChange,
    handleParentOrganizationChange,
    handleParentOrganizationError,
    handleApplySuggestedEffectiveDate,
    handleResetParentSelection,
    handleSubmit,
    handleEditHistoryToggle,
    handleCancelEditHistory,
    handleEditHistorySubmit,
    handleDeactivateClick,
    handleDeleteOrganizationClick,
    handleConfirmDelete,
    handleDeactivateCancel,
    handleStartInsertVersion,
    handleUnitTypeChange,
  };
};

export default useInlineNewVersionForm;
