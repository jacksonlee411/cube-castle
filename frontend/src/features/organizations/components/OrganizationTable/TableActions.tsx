import React from 'react';
import { TertiaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Text } from '@workday/canvas-kit-react/text';
import { Tooltip } from '@workday/canvas-kit-react/tooltip';
import type { TableActionsProps } from './TableTypes';

export const TableActions: React.FC<TableActionsProps> = ({
  organization,
  onToggleStatus,
  onTemporalManage,
  isToggling,
  disabled,
  isHistorical = false
}) => {
  const handleToggleStatus = () => onToggleStatus?.(organization.code, organization.status);
  const handleTemporalManage = () => onTemporalManage?.(organization.code);

  const isActive = organization.status === 'ACTIVE';
  const buttonText = isActive ? '停用' : '启用';
  const loadingText = isActive ? '停用中...' : '启用中...';

  // 在历史模式下显示禁用状态
  if (isHistorical) {
    return (
      <div style={{ display: 'flex', gap: '4px' }}>
        <Tooltip title="历史数据不支持状态变更">
          <SecondaryButton 
            size="small" 
            disabled={true}
            data-testid={`toggle-status-button-${organization.code}`}
          >
            {buttonText}
          </SecondaryButton>
        </Tooltip>
        {onTemporalManage && (
          <Tooltip title="查看历史版本的组织详情">
            <TertiaryButton 
              aria-label="组织详情"
              onClick={handleTemporalManage}
              data-testid={`temporal-manage-button-${organization.code}`}
            >
              详情
            </TertiaryButton>
          </Tooltip>
        )}
        <Text typeLevel="subtext.small" color="hint">
          📖
        </Text>
      </div>
    );
  }

  // 正常模式下的操作按钮（移除编辑按钮）
  return (
    <div style={{ display: 'flex', gap: '4px' }}>
      <SecondaryButton 
        size="small" 
        onClick={handleToggleStatus}
        disabled={disabled || !onToggleStatus}
        data-testid={`toggle-status-button-${organization.code}`}
      >
        {isToggling ? loadingText : buttonText}
      </SecondaryButton>
      {onTemporalManage && (
        <Tooltip title="组织详情">
          <TertiaryButton 
            aria-label="组织详情"
            onClick={handleTemporalManage}
            disabled={disabled}
            data-testid={`temporal-manage-button-${organization.code}`}
          >
            详情
          </TertiaryButton>
        </Tooltip>
      )}
    </div>
  );
};