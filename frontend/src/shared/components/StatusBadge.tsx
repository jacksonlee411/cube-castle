import React from 'react';
import { Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { STATUS_CONFIG, type OrganizationStatus } from '../utils/statusUtils';

// 重新导出类型以保持向后兼容
export type { OrganizationStatus };

export interface StatusBadgeProps {
  status: OrganizationStatus;
  size?: 'small' | 'medium' | 'large';
  showIcon?: boolean;
}

/**
 * 简化的状态显示组件
 * 只显示3个基本状态：启用、停用、计划中
 */
export const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  size = 'medium',
  showIcon = true
}) => {
  const config = STATUS_CONFIG[status];
  
  if (!config) {
    console.warn(`未知的组织状态: ${status}`);
    return null;
  }

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

  const styles = getStyles();

  return (
    <Flex
      alignItems="center"
      gap={styles.gap}
      style={{
        padding: styles.padding,
        backgroundColor: config.backgroundColor,
        border: `1px solid ${config.borderColor}`,
        borderRadius: '4px',
        display: 'inline-flex'
      }}
    >
      {showIcon && (
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