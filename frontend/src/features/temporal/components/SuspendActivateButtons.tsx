import React, { useMemo } from 'react';
import { Flex } from '@workday/canvas-kit-react/layout';
import { SecondaryButton } from '@workday/canvas-kit-react/button';
import { colors } from '@workday/canvas-kit-react/tokens';
import { mediaPauseIcon, mediaPlayIcon } from '@workday/canvas-system-icons-web';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { logger } from '@/shared/utils/logger';
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

  const isInactive = currentStatus === 'INACTIVE' || currentStatus === 'DELETED';

  const baseDisabled = disabled || readonly || isSuspending || isActivating;

  const handleSuspend = async () => {
    try {
      const result = await suspendAsync({
        code: organizationCode,
        effectiveDate,
        currentETag: currentETag ?? undefined,
        operationReason: '自动生成停用',
      });
      onETagChange(result.etag);
      await onCompleted('suspend', result);
      onSuccess('组织已停用');
    } catch (error) {
      logger.error('暂停组织失败', error);
      const message =
        error instanceof Error ? error.message : '暂停组织失败，请稍后再试';
      onError(message);
    }
  };

  const handleActivate = async () => {
    try {
      const result = await activateAsync({
        code: organizationCode,
        effectiveDate,
        currentETag: currentETag ?? undefined,
        operationReason: '自动生成重新启用',
      });
      onETagChange(result.etag);
      await onCompleted('activate', result);
      onSuccess('组织已重新启用');
    } catch (error) {
      logger.error('重新启用组织失败', error);
      const message =
        error instanceof Error ? error.message : '重新启用组织失败，请稍后再试';
      onError(message);
    }
  };

  if (!organizationCode || readonly) {
    return null;
  }

  return (
    <Flex gap="s" alignItems="center">
      <SecondaryButton
        data-testid={isInactive ? 'activate-organization-button' : 'suspend-organization-button'}
        onClick={isInactive ? handleActivate : handleSuspend}
        disabled={baseDisabled}
        icon={
          <SystemIcon
            icon={isInactive ? mediaPlayIcon : mediaPauseIcon}
            color={isInactive ? colors.greenApple500 : colors.cantaloupe600}
            size={16}
          />
        }
      >
        {isInactive ? '重新启用' : '暂停组织'}
      </SecondaryButton>
    </Flex>
  );
};

export default SuspendActivateButtons;
