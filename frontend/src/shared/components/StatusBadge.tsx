import React from 'react';
import { Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { 
  checkCircleIcon, 
  clockPauseIcon, 
  clockIcon 
} from '@workday/canvas-system-icons-web';
import { colors } from '@workday/canvas-kit-react/tokens';

// 简化的组织状态类型（3个状态）
export type OrganizationStatus = 'ACTIVE' | 'SUSPENDED' | 'PLANNED';

// 状态配置
const STATUS_CONFIG = {
  ACTIVE: {
    label: '启用',
    color: colors.greenApple600,
    icon: checkCircleIcon,
    backgroundColor: colors.greenApple100,
    borderColor: colors.greenApple300
  },
  SUSPENDED: {
    label: '停用',
    color: colors.cantaloupe600,
    icon: clockPauseIcon,
    backgroundColor: colors.cantaloupe100,
    borderColor: colors.cantaloupe300
  },
  PLANNED: {
    label: '计划中',
    color: colors.blueberry600,
    icon: clockIcon,
    backgroundColor: colors.blueberry100,
    borderColor: colors.blueberry300
  }
} as const;

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

/**
 * 状态工具函数
 */
export const statusUtils = {
  // 获取状态标签
  getStatusLabel: (status: OrganizationStatus): string => {
    return STATUS_CONFIG[status]?.label || status;
  },

  // 获取状态颜色
  getStatusColor: (status: OrganizationStatus): string => {
    return STATUS_CONFIG[status]?.color || colors.licorice400;
  },

  // 验证状态有效性
  isValidStatus: (status: string): status is OrganizationStatus => {
    return ['ACTIVE', 'SUSPENDED', 'PLANNED'].includes(status);
  },

  // 获取可用操作
  getAvailableActions: (status: OrganizationStatus): string[] => {
    switch (status) {
      case 'ACTIVE':
        return ['UPDATE', 'SUSPEND'];
      case 'SUSPENDED':
        return ['REACTIVATE'];
      case 'PLANNED':
        return ['UPDATE'];
      default:
        return [];
    }
  }
};

export default StatusBadge;