import React from 'react';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { colors } from '@workday/canvas-kit-react/tokens';
import { exclamationCircleIcon } from '@workday/canvas-system-icons-web';
import type { InlineVersionRecord } from './types';

export interface DeactivateConfirmModalProps {
  visible: boolean;
  modalModel: ReturnType<typeof useModalModel>;
  selectedVersion?: InlineVersionRecord | null;
  onConfirm: () => Promise<void>;
  onCancel: () => void;
  isDeactivating: boolean;
}

const DeactivateConfirmModal: React.FC<DeactivateConfirmModalProps> = ({
  visible,
  modalModel,
  selectedVersion,
  onConfirm,
  onCancel,
  isDeactivating,
}) => {
  if (!visible || !selectedVersion) {
    return null;
  }

  const effectiveDate = new Date(selectedVersion.effectiveDate).toLocaleDateString('zh-CN');

  return (
    <Modal model={modalModel}>
      <Modal.Overlay>
        <Modal.Card>
          <Modal.CloseIcon onClick={onCancel} />
          <Modal.Heading>确认删除版本</Modal.Heading>
          <Modal.Body>
            <Box padding="l">
              <Flex alignItems="flex-start" gap="m" marginBottom="l">
                <SystemIcon icon={exclamationCircleIcon} size={24} color={colors.cinnamon600} />
                <Box>
                  <Text typeLevel="body.medium" marginBottom="s">
                    确定要删除生效日期为 <strong>{effectiveDate}</strong> 的版本吗？
                  </Text>
                  <Text typeLevel="subtext.small" color="hint" marginBottom="s">
                    版本名称: {selectedVersion.name}
                  </Text>
                  <Text typeLevel="subtext.small" color={colors.cinnamon600}>
                    删除后记录将标记为已删除状态，此操作不可撤销
                  </Text>
                </Box>
              </Flex>
              <Flex gap="s" justifyContent="flex-end">
                <SecondaryButton onClick={onCancel} disabled={isDeactivating}>
                  取消
                </SecondaryButton>
                <PrimaryButton onClick={onConfirm} disabled={isDeactivating}>
                  {isDeactivating ? '删除中...' : '确认删除'}
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
