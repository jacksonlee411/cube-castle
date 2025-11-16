import React, { useMemo, useState } from 'react';
import { Flex } from '@workday/canvas-kit-react/layout';
import { SecondaryButton, PrimaryButton } from '@workday/canvas-kit-react/button';
import { mediaPauseIcon, mediaPlayIcon } from '@workday/canvas-system-icons-web';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { Text } from '@workday/canvas-kit-react/text';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import { logger } from '@/shared/utils/logger';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';
import {
  useSuspendOrganization,
  useActivateOrganization,
  type OrganizationStateMutationResult,
} from '@/shared/hooks/useOrganizationMutations';

interface SuspendActivateButtonsProps {
  organizationCode: string;
  currentStatus: string | undefined;
  currentETag: string | null;
  readonly?: boolean;
  disabled?: boolean;
  onETagChange: (etag: string | null) => void;
  onSuccess: (message: string) => void;
  onError: (message: string) => void;
  onCompleted: (
    action: 'suspend' | 'activate',
    result: OrganizationStateMutationResult,
  ) => Promise<void> | void;
}

const getTodayISODate = (): string => {
  const now = new Date();
  const month = `${now.getMonth() + 1}`.padStart(2, '0');
  const day = `${now.getDate()}`.padStart(2, '0');
  return `${now.getFullYear()}-${month}-${day}`;
};

export const SuspendActivateButtons: React.FC<SuspendActivateButtonsProps> = ({
  organizationCode,
  currentStatus,
  currentETag,
  readonly = false,
  disabled = false,
  onETagChange,
  onSuccess,
  onError,
  onCompleted,
}) => {
  const {
    mutateAsync: suspendAsync,
    isPending: isSuspending,
  } = useSuspendOrganization();
  const {
    mutateAsync: activateAsync,
    isPending: isActivating,
  } = useActivateOrganization();

  const effectiveDate = useMemo(getTodayISODate, []);
  const [dialogMode, setDialogMode] = useState<'suspend' | 'activate' | null>(null);
  const [selectedDate, setSelectedDate] = useState(effectiveDate);
  const [operationReasonInput, setOperationReasonInput] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);
  const modalModel = useModalModel();

  const isInactive = currentStatus === 'INACTIVE';

  const baseDisabled = disabled || readonly || isSuspending || isActivating;

  const openDialog = (mode: 'suspend' | 'activate') => {
    setDialogMode(mode);
    setSelectedDate(getTodayISODate());
    setOperationReasonInput('');
    modalModel.events.show();
  };

  const closeDialog = () => {
    modalModel.events.hide();
    setDialogMode(null);
    setIsProcessing(false);
  };

  const handleConfirm = async () => {
    if (!dialogMode) {
      return;
    }

    setIsProcessing(true);

    try {
      let result: OrganizationStateMutationResult;
      if (dialogMode === 'suspend') {
        result = await suspendAsync({
          code: organizationCode,
          effectiveDate: selectedDate,
          currentETag: currentETag ?? undefined,
          operationReason: operationReasonInput.trim() || undefined,
        });
      } else {
        result = await activateAsync({
          code: organizationCode,
          effectiveDate: selectedDate,
          currentETag: currentETag ?? undefined,
          operationReason: operationReasonInput.trim() || undefined,
        });
      }

      onETagChange(result.etag);
      await onCompleted(dialogMode, result);
      onSuccess(dialogMode === 'suspend' ? '组织已停用' : '组织已重新启用');
      closeDialog();
    } catch (error) {
      logger.error('状态变更失败', error);
      const message =
        error instanceof Error ? error.message : '状态变更失败，请稍后再试';
      onError(message);
      setIsProcessing(false);
    }
  };

  if (!organizationCode || readonly) {
    return null;
  }

  if (currentStatus === 'DELETED') {
    return null;
  }

  return (
    <>
      <Flex gap="s" alignItems="center">
        <SecondaryButton
          data-testid={
            isInactive
              ? temporalEntitySelectors.organization.stateChange?.activateButton
              : temporalEntitySelectors.organization.stateChange?.suspendButton
          }
        onClick={() => openDialog(isInactive ? 'activate' : 'suspend')}
        disabled={baseDisabled}
        icon={isInactive ? mediaPlayIcon : mediaPauseIcon}
        iconPosition="start"
      >
        {isInactive ? '重新启用' : '暂停组织'}
      </SecondaryButton>
      </Flex>

      {dialogMode && (
        <Modal model={modalModel}>
          <Modal.Overlay>
            <Modal.Card width={480}>
              <Modal.CloseIcon aria-label="关闭" onClick={closeDialog} />
              <Modal.Heading>
                {dialogMode === 'suspend' ? '暂停组织' : '重新启用组织'}
              </Modal.Heading>
              <Modal.Body>
                <Flex flexDirection="column" gap="m">
                  <div>
                    <Text typeLevel="body.small" marginBottom="xs">
                      生效日期
                    </Text>
                    <TextInput
                      type="date"
                      value={selectedDate}
                      onChange={(e) => setSelectedDate(e.target.value)}
                      data-testid={temporalEntitySelectors.organization.stateChange?.dateInput}
                    />
                  </div>

                  <div>
                    <Text typeLevel="body.small" marginBottom="xs">
                      操作原因（可选）
                    </Text>
                    <TextArea
                      rows={3}
                      value={operationReasonInput}
                      onChange={(e) => setOperationReasonInput(e.target.value)}
                      placeholder="请输入此次停用/启用原因"
                      data-testid={temporalEntitySelectors.organization.stateChange?.reasonInput}
                    />
                  </div>

                  <Flex justifyContent="flex-end" gap="s">
                    <SecondaryButton
                      onClick={closeDialog}
                      disabled={isProcessing}
                      data-testid={temporalEntitySelectors.organization.stateChange?.cancel}
                    >
                      取消
                    </SecondaryButton>
                    <PrimaryButton
                      onClick={handleConfirm}
                      disabled={isProcessing || !selectedDate}
                      data-testid={temporalEntitySelectors.organization.stateChange?.confirm}
                    >
                      {isProcessing
                        ? '处理中...'
                        : dialogMode === 'suspend'
                          ? '确认暂停'
                          : '确认启用'}
                    </PrimaryButton>
                  </Flex>
                </Flex>
              </Modal.Body>
            </Modal.Card>
          </Modal.Overlay>
        </Modal>
      )}
    </>
  );
};

export default SuspendActivateButtons;
