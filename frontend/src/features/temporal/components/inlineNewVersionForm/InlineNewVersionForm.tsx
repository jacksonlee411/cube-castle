import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Card } from '@workday/canvas-kit-react/card';
import { useModalModel } from '@workday/canvas-kit-react/modal';
import type { InlineNewVersionFormProps } from './types';
import useInlineNewVersionForm from './useInlineNewVersionForm';
import FormHeader from './FormHeader';
import FormMessages from './FormMessages';
import EffectiveDateSection from './EffectiveDateSection';
import BasicInfoSection from './BasicInfoSection';
import HierarchySection from './HierarchySection';
import RecordInfoSection from './RecordInfoSection';
import FormActions from './FormActions';
import DeactivateConfirmModal from './DeactivateConfirmModal';

const InlineNewVersionForm: React.FC<InlineNewVersionFormProps> = (props) => {
  const {
    organizationCode,
    onCancel,
    isSubmitting = false,
    selectedVersion = null,
  } = props;

  const {
    formData,
    errors,
    parentError,
    suggestedEffectiveDate,
    isEditingHistory,
    originalHistoryData,
    showDeactivateConfirm,
    isDeactivating,
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
    handleDeactivateConfirm,
    handleDeactivateCancel,
    handleStartInsertVersion,
    handleUnitTypeChange,
  } = useInlineNewVersionForm(props);

  const deactivateModalModel = useModalModel();

  React.useEffect(() => {
    if (showDeactivateConfirm && deactivateModalModel.state.visibility !== 'visible') {
      deactivateModalModel.events.show();
    } else if (!showDeactivateConfirm && deactivateModalModel.state.visibility === 'visible') {
      deactivateModalModel.events.hide();
    }
  }, [deactivateModalModel, showDeactivateConfirm]);

  const fieldDisabled = isSubmitting || (currentMode === 'edit' && !isEditingHistory);

  return (
    <Box flex="1">
      <Card padding="l" data-testid="organization-form">
        <FormHeader
          currentMode={currentMode}
          isEditingHistory={isEditingHistory}
          organizationCode={organizationCode}
          originalHistoryData={originalHistoryData}
          selectedVersion={selectedVersion}
        />
        <FormMessages errorMessage={errorMessage} successMessage={successMessage} />

        <form onSubmit={(event) => handleSubmit(event)} data-testid="organization-form-content">
          <EffectiveDateSection
            value={formData.effectiveDate}
            error={errors.effectiveDate}
            onChange={handleInputChange('effectiveDate')}
            disabled={fieldDisabled}
          />

          <BasicInfoSection
            formData={formData}
            errors={errors}
            disabled={fieldDisabled}
            organizationCode={organizationCode}
            onFieldChange={handleInputChange}
            onParentChange={handleParentOrganizationChange}
            onParentError={handleParentOrganizationError}
            parentError={parentError}
            suggestedEffectiveDate={suggestedEffectiveDate}
            onApplySuggestedEffectiveDate={handleApplySuggestedEffectiveDate}
            onResetParentSelection={handleResetParentSelection}
            isSubmitting={isSubmitting || loading}
            onUnitTypeChange={handleUnitTypeChange}
          />

          <HierarchySection
            currentMode={currentMode}
            selectedVersion={selectedVersion}
            levelDisplay={levelDisplay}
            codePathDisplay={codePathDisplay}
            namePathDisplay={namePathDisplay}
          />

          <RecordInfoSection originalHistoryData={originalHistoryData} />

          <FormActions
            currentMode={currentMode}
            isEditingHistory={isEditingHistory}
            isSubmitting={isSubmitting}
            loading={loading}
            selectedVersion={selectedVersion}
            onCancel={onCancel}
            onDeactivateClick={handleDeactivateClick}
            onToggleEditHistory={handleEditHistoryToggle}
            onCancelEditHistory={handleCancelEditHistory}
            onSubmitEditHistory={handleEditHistorySubmit}
            onSubmitNewVersion={handleSubmit}
            originalHistoryData={originalHistoryData}
            onStartInsertVersion={handleStartInsertVersion}
            isDeactivating={isDeactivating}
          />
        </form>
      </Card>

      <DeactivateConfirmModal
        visible={showDeactivateConfirm}
        modalModel={deactivateModalModel}
        selectedVersion={selectedVersion}
        onConfirm={handleDeactivateConfirm}
        onCancel={handleDeactivateCancel}
        isDeactivating={isDeactivating}
      />
    </Box>
  );
};

export default InlineNewVersionForm;
export { InlineNewVersionForm };
