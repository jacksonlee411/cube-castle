import { logger } from '@/shared/utils/logger'
import React from 'react'
import { Flex } from '@workday/canvas-kit-react/layout'
import { Text } from '@workday/canvas-kit-react/text'
import { SystemIcon } from '@workday/canvas-kit-react/icon'
import type { OrganizationStatus } from '@/shared/types'
import {
  TEMPORAL_ENTITY_STATUS_META,
  getOrganizationStatusMeta,
} from '@/features/temporal/entity/statusMeta'

export interface StatusBadgeProps {
  status: OrganizationStatus;
  size?: 'small' | 'medium' | 'large';
  showIcon?: boolean;
}

/**
 * 简化的状态显示组件
 * 只显示当前 API 可用的三种状态（ACTIVE/INACTIVE 以及调用方派生的 PLANNED）。
 * 若未来扩展 DELETED 等状态，需要先更新契约与 STATUS_CONFIG。
 */
export const StatusBadge: React.FC<StatusBadgeProps> = ({ status, size = 'medium', showIcon = true }) => {
  const key = status?.toString().toUpperCase()
  const registry = TEMPORAL_ENTITY_STATUS_META.organization
  if (!registry[key]) {
    logger.warn(`未知的组织状态: ${status}`)
  }
  const config = getOrganizationStatusMeta(status)

  const getStyles = () => {
    switch (size) {
      case 'small':
        return {
          padding: '2px 6px',
          fontSize: '12px',
          iconSize: 12,
          gap: 'xs' as const
        };
      case 'large':
        return {
          padding: '6px 12px',
          fontSize: '16px',
          iconSize: 18,
          gap: 's' as const
        };
      default: // medium
        return {
          padding: '4px 8px',
          fontSize: '14px',
          iconSize: 16,
          gap: 'xs' as const
        };
    }
  };

  const styles = getStyles()

  return (
    <Flex
      alignItems="center"
      gap={styles.gap}
      style={{
        padding: styles.padding,
        backgroundColor: config.background,
        border: `1px solid ${config.border}`,
        borderRadius: '4px',
        display: 'inline-flex'
      }}
    >
      {showIcon && config.icon && (
        <SystemIcon
          icon={config.icon}
          size={styles.iconSize}
          color={config.color}
        />
      )}
      <Text
        style={{
          fontSize: styles.fontSize,
          fontWeight: 'medium',
          color: config.color
        }}
      >
        {config.label}
      </Text>
    </Flex>
  );
};


export default StatusBadge;
