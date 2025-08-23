import React, { useState } from 'react';
import { Flex } from '@workday/canvas-kit-react/layout';
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { 
  editIcon, 
  clockPauseIcon, 
  checkCircleIcon 
} from '@workday/canvas-system-icons-web';
import { statusUtils } from '../components/StatusBadge';
import type { OrganizationStatus } from '../components/StatusBadge';

export interface Organization {
  code: string;
  name: string;
  status: OrganizationStatus;
  [key: string]: any;
}

export interface OrganizationActionsProps {
  organization: Organization;
  onUpdate?: (organization: Organization) => void;
  onSuspend?: (organization: Organization, reason: string) => Promise<void>;
  onReactivate?: (organization: Organization, reason: string) => Promise<void>;
  disabled?: boolean;
}

/**
 * 操作驱动的组织操作组件
 * 根据当前状态显示可用操作，状态由操作自动管理
 */
export const OrganizationActions: React.FC<OrganizationActionsProps> = ({
  organization,
  onUpdate,
  onSuspend,
  onReactivate,
  disabled = false
}) => {
  const [loading, setLoading] = useState<string | null>(null);
  const availableActions = statusUtils.getAvailableActions(organization.status);

  const handleAction = async (action: string) => {
    if (disabled || loading) return;

    try {
      setLoading(action);

      switch (action) {
        case 'UPDATE':
          onUpdate?.(organization);
          break;
          
        case 'SUSPEND':
          if (onSuspend) {
            const reason = prompt('请输入停用原因：');
            if (reason && reason.trim()) {
              await onSuspend(organization, reason.trim());
            }
          }
          break;
          
        case 'REACTIVATE':
          if (onReactivate) {
            const reason = prompt('请输入重新启用原因：');
            if (reason && reason.trim()) {
              await onReactivate(organization, reason.trim());
            }
          }
          break;
          
        default:
          console.warn(`未知操作: ${action}`);
      }
    } catch (error) {
      console.error(`操作失败 [${action}]:`, error);
      alert(`操作失败: ${error}`);
    } finally {
      setLoading(null);
    }
  };

  const getActionConfig = (action: string) => {
    switch (action) {
      case 'UPDATE':
        return {
          label: '编辑',
          icon: editIcon,
          variant: 'secondary' as const,
          title: '编辑组织信息'
        };
      case 'SUSPEND':
        return {
          label: '停用',
          icon: clockPauseIcon,
          variant: 'secondary' as const,
          title: '停用组织（可恢复）'
        };
      case 'REACTIVATE':
        return {
          label: '重启',
          icon: checkCircleIcon,
          variant: 'primary' as const,
          title: '重新启用组织'
        };
      default:
        return {
          label: action,
          icon: editIcon,
          variant: 'secondary' as const,
          title: action
        };
    }
  };

  if (availableActions.length === 0) {
    return null;
  }

  return (
    <Flex gap="s" alignItems="center">
      {availableActions.map(action => {
        const config = getActionConfig(action);
        const isLoading = loading === action;
        
        const ButtonComponent = config.variant === 'primary' ? PrimaryButton : SecondaryButton;
        
        return (
          <ButtonComponent
            key={action}
            size="small"
            onClick={() => handleAction(action)}
            disabled={disabled || !!loading}
            title={config.title}
          >
            <Flex alignItems="center" gap="xs">
              <SystemIcon 
                icon={config.icon} 
                size={16} 
              />
              <span>{isLoading ? '处理中...' : config.label}</span>
            </Flex>
          </ButtonComponent>
        );
      })}
    </Flex>
  );
};

/**
 * 组织操作上下文（可选的高级用法）
 */
export interface OrganizationOperationContext {
  canUpdate: boolean;
  canSuspend: boolean;
  canReactivate: boolean;
  canDelete: boolean;
}

/**
 * 获取操作权限的辅助函数
 */
export const getOperationPermissions = (
  organization: Organization,
  userRole?: string
): OrganizationOperationContext => {
  // 这里可以根据用户角色和组织状态计算权限
  const isAdmin = userRole === 'admin';
  const isManager = userRole === 'manager' || isAdmin;
  
  return {
    canUpdate: isManager,
    canSuspend: isManager && organization.status === 'ACTIVE',
    canReactivate: isManager && organization.status === 'SUSPENDED',
    canDelete: isAdmin && organization.status !== 'ACTIVE'
  };
};

export default OrganizationActions;