import React from 'react';
import { TertiaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Text } from '@workday/canvas-kit-react/text';
import { Tooltip } from '@workday/canvas-kit-react/tooltip';
import type { TableActionsProps } from './TableTypes';

export const TableActions: React.FC<TableActionsProps> = ({
  organization,
  onEdit,
  onToggleStatus,
  isToggling,
  disabled,
  isHistorical = false
}) => {
  const handleEdit = () => onEdit?.(organization);
  const handleToggleStatus = () => onToggleStatus?.(organization.code, organization.status);

  const isActive = organization.status === 'ACTIVE';
  const buttonText = isActive ? '停用' : '启用';
  const loadingText = isActive ? '停用中...' : '启用中...';

  // 在历史模式下显示禁用状态
  if (isHistorical) {
    return (
      <div style={{ display: 'flex', gap: '4px' }}>
        <Tooltip title="历史数据不支持编辑">
          <TertiaryButton 
            size="small" 
            disabled={true}
            data-testid={`edit-button-${organization.code}`}
          >
            编辑
          </TertiaryButton>
        </Tooltip>
        <Tooltip title="历史数据不支持状态变更">
          <SecondaryButton 
            size="small" 
            disabled={true}
            data-testid={`toggle-status-button-${organization.code}`}
            variant={isActive ? 'inverse' : 'primary'}
          >
            {buttonText}
          </SecondaryButton>
        </Tooltip>
        <Text typeLevel="subtext.small" color="hint">
          📖
        </Text>
      </div>
    );
  }

  // 正常模式下的操作按钮
  return (
    <div style={{ display: 'flex', gap: '4px' }}>
      <TertiaryButton 
        size="small" 
        onClick={handleEdit}
        disabled={disabled || !onEdit}
        data-testid={`edit-button-${organization.code}`}
      >
        编辑
      </TertiaryButton>
      <SecondaryButton 
        size="small" 
        onClick={handleToggleStatus}
        disabled={disabled || !onToggleStatus}
        data-testid={`toggle-status-button-${organization.code}`}
        variant={isActive ? 'inverse' : 'primary'}
      >
        {isToggling ? loadingText : buttonText}
      </SecondaryButton>
    </div>
  );
};