import React from 'react';
import { TertiaryButton } from '@workday/canvas-kit-react/button';
import { Text } from '@workday/canvas-kit-react/text';
import { Tooltip } from '@workday/canvas-kit-react/tooltip';
import type { TableActionsProps } from './TableTypes';

export const TableActions: React.FC<TableActionsProps> = ({
  organization,
  onTemporalManage,
  disabled,
  isHistorical = false
}) => {
  const handleTemporalManage = () => onTemporalManage?.(organization.code);

  // 统一显示详情按钮，无论历史模式还是当前模式
  return (
    <div style={{ display: 'flex', gap: '4px', alignItems: 'center' }}>
      {onTemporalManage && (
        <Tooltip title={isHistorical ? "查看历史版本的组织详情" : "管理组织详情和状态"}>
          <TertiaryButton 
            aria-label="组织详情"
            onClick={handleTemporalManage}
            disabled={disabled}
            data-testid={`temporal-manage-button-${organization.code}`}
          >
            详情管理
          </TertiaryButton>
        </Tooltip>
      )}
      {isHistorical && (
        <Text typeLevel="subtext.small" color="hint">
          📖
        </Text>
      )}
    </div>
  );
};