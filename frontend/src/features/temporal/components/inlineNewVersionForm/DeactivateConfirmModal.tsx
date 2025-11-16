import React from 'react';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { colors } from '@workday/canvas-kit-react/tokens';
import { exclamationCircleIcon } from '@workday/canvas-system-icons-web';
import type { DeleteConfirmMode, InlineVersionRecord } from './types';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

export interface DeactivateConfirmModalProps {
  visible: boolean;
  modalModel: ReturnType<typeof useModalModel>;
  selectedVersion?: InlineVersionRecord | null;
  mode: DeleteConfirmMode;
  organizationCode?: string | null;
  onConfirm: () => Promise<void>;
  onCancel: () => void;
  isProcessing: boolean;
}

const DeactivateConfirmModal: React.FC<DeactivateConfirmModalProps> = ({
  visible,
  modalModel,
  selectedVersion,
  mode,
  organizationCode,
  onConfirm,
  onCancel,
  isProcessing,
}) => {
  if (!visible || !selectedVersion || !mode) {
    return null;
  }

  const effectiveDate = new Date(selectedVersion.effectiveDate).toLocaleDateString('zh-CN');
  const heading =
    mode === 'organization' ? '确认删除组织编码' : '确认删除版本';
  const confirmLabel =
    mode === 'organization' ? '删除组织编码' : '确认删除';
  const warningText =
    mode === 'organization'
      ? '组织编码删除后将无法恢复，请确保已处理所有子组织'
      : '删除后记录将标记为已删除状态，此操作不可撤销';
  const targetName =
    mode === 'organization'
      ? organizationCode ?? selectedVersion.code
      : selectedVersion.name;

  return (
    <Modal model={modalModel}>
      <Modal.Overlay>
        <Modal.Card>
          <Modal.CloseIcon onClick={onCancel} />
          <Modal.Heading>{heading}</Modal.Heading>
          <Modal.Body>
            <Box padding="l">
              <Flex alignItems="flex-start" gap="m" marginBottom="l">
                <SystemIcon icon={exclamationCircleIcon} size={24} color={colors.cinnamon600} />
                <Box>
                  <Text typeLevel="body.medium" marginBottom="s">
                    {mode === 'organization' ? (
                      <>
                        确定要删除组织编码 <strong>{targetName}</strong> 吗？
                      </>
                    ) : (
                      <>
                        确定要删除生效日期为 <strong>{effectiveDate}</strong> 的版本吗？
                      </>
                    )}
                  </Text>
                  {mode === 'organization' ? (
                    <Text typeLevel="subtext.small" color="hint" marginBottom="s">
                      生效日期: {effectiveDate}
                    </Text>
                  ) : (
                    <Text typeLevel="subtext.small" color="hint" marginBottom="s">
                      版本名称: {selectedVersion.name}
                    </Text>
                  )}
                  <Text typeLevel="subtext.small" color={colors.cinnamon600}>
                    {warningText}
                  </Text>
                </Box>
              </Flex>
              <Flex gap="s" justifyContent="flex-end">
                <SecondaryButton
                  onClick={onCancel}
                  disabled={isProcessing}
                  data-testid={temporalEntitySelectors.form?.actions?.cancelEditHistory}
                >
                  取消
                </SecondaryButton>
                <PrimaryButton
                  onClick={onConfirm}
                  disabled={isProcessing}
                  data-testid={temporalEntitySelectors.form?.actions?.submitEditHistory}
                >
                  {isProcessing ? '删除中...' : confirmLabel}
                </PrimaryButton>
              </Flex>
            </Box>
          </Modal.Body>
        </Modal.Card>
      </Modal.Overlay>
    </Modal>
  );
};

export default DeactivateConfirmModal;
