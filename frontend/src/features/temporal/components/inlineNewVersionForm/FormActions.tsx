import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';
import { colors } from '@workday/canvas-kit-react/tokens';
import type { InlineNewVersionFormMode, InlineVersionRecord } from './types';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

export type InlineSubmitEvent =
  | React.FormEvent<HTMLFormElement>
  | React.MouseEvent<HTMLButtonElement, MouseEvent>
  | React.MouseEvent<HTMLAnchorElement, MouseEvent>
  | undefined;

export interface FormActionsProps {
  currentMode: InlineNewVersionFormMode;
  isEditingHistory: boolean;
  isSubmitting: boolean;
  loading: boolean;
  selectedVersion?: InlineVersionRecord | null;
  onCancel: () => void;
  onDeactivateClick: () => void;
  onToggleEditHistory: () => void;
  onCancelEditHistory: () => void;
  onSubmitEditHistory: () => Promise<void>;
  onSubmitNewVersion: (event?: InlineSubmitEvent) => Promise<void> | void;
  originalHistoryData: InlineVersionRecord | null;
  onStartInsertVersion: () => void;
  isDeactivating: boolean;
  canDeleteOrganization?: boolean;
  onDeleteOrganizationClick?: () => void;
  isProcessingDelete?: boolean;
}

const FormActions: React.FC<FormActionsProps> = ({
  currentMode,
  isEditingHistory,
  isSubmitting,
  loading,
  selectedVersion,
  onCancel,
  onDeactivateClick,
  onDeleteOrganizationClick,
  onToggleEditHistory,
  onCancelEditHistory,
  onSubmitEditHistory,
  onSubmitNewVersion,
  originalHistoryData,
  onStartInsertVersion,
  isDeactivating,
  canDeleteOrganization = false,
  isProcessingDelete = false,
}) => {
  if (currentMode === 'edit') {
    const deleteButtonDisabled =
      isSubmitting || isDeactivating || isProcessingDelete;
    const showOrganizationDelete =
      canDeleteOrganization && !!selectedVersion && !isEditingHistory;
    const showRecordDelete =
      !showOrganizationDelete && !!selectedVersion && !isEditingHistory;

    return (
      <Box marginTop="xl" paddingTop="l" borderTop={`1px solid ${colors.soap300}`}>
        <Flex gap="s" justifyContent="space-between">
          <Box data-testid={temporalEntitySelectors.form?.actions?.deleteRecordWrapper}>
            {showOrganizationDelete ? (
              <TertiaryButton
                onClick={onDeleteOrganizationClick}
                disabled={deleteButtonDisabled || !onDeleteOrganizationClick}
                data-testid={temporalEntitySelectors.form?.actions?.deleteOrganization}
              >
                删除组织编码
              </TertiaryButton>
            ) : null}
            {showRecordDelete ? (
              <TertiaryButton
                onClick={onDeactivateClick}
                disabled={deleteButtonDisabled}
                data-testid={temporalEntitySelectors.form?.actions?.deleteRecord}
              >
                删除此记录
              </TertiaryButton>
            ) : null}
          </Box>
          <Flex gap="s">
            {!isEditingHistory ? (
              <>
                <SecondaryButton
                  onClick={onStartInsertVersion}
                  disabled={isSubmitting || loading}
                  data-testid={temporalEntitySelectors.form?.actions?.startInsertVersion}
                >
                  插入新版本
                </SecondaryButton>
                <SecondaryButton
                  onClick={onToggleEditHistory}
                  disabled={isSubmitting || loading}
                  data-testid={temporalEntitySelectors.form?.actions?.editHistoryToggle}
                >
                  修改记录
                </SecondaryButton>
                <PrimaryButton
                  onClick={onCancel}
                  disabled={isSubmitting}
                  data-testid={temporalEntitySelectors.form?.actions?.formClose}
                >
                  关闭
                </PrimaryButton>
              </>
            ) : (
              <>
                <SecondaryButton
                  onClick={onCancelEditHistory}
                  disabled={isSubmitting || loading}
                  data-testid={temporalEntitySelectors.form?.actions?.cancelEditHistory}
                >
                  取消编辑
                </SecondaryButton>
                <PrimaryButton
                  onClick={(event) =>
                    originalHistoryData ? onSubmitEditHistory() : onSubmitNewVersion(event)
                  }
                  disabled={isSubmitting || loading}
                  data-testid={temporalEntitySelectors.form?.actions?.submitEditHistory}
                >
                  {isSubmitting || loading
                    ? '提交中...'
                    : originalHistoryData
                      ? '提交修改'
                      : '插入新版本'}
                </PrimaryButton>
              </>
            )}
          </Flex>
        </Flex>
      </Box>
    );
  }

  return (
    <Box marginTop="xl" paddingTop="l" borderTop={`1px solid ${colors.soap300}`}>
      <Flex gap="s" justifyContent="flex-end">
        <SecondaryButton
          onClick={onCancel}
          disabled={isSubmitting || loading}
          data-testid={temporalEntitySelectors.form?.actions?.formCancel}
        >
          取消
        </SecondaryButton>
        <PrimaryButton
          type="submit"
          disabled={isSubmitting || loading}
          data-testid={temporalEntitySelectors.form?.actions?.formSubmit}
        >
          {isSubmitting || loading
            ? currentMode === 'create'
              ? '创建中...'
              : '保存中...'
            : currentMode === 'create'
              ? '创建组织'
              : isEditingHistory
                ? '保存修改'
                : '保存新版本'}
        </PrimaryButton>
      </Flex>
    </Box>
  );
};

export default FormActions;
