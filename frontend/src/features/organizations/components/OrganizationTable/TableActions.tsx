import React from 'react';
import { TertiaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import type { TableActionsProps } from './TableTypes';

export const TableActions: React.FC<TableActionsProps> = ({
  organization,
  onEdit,
  onToggleStatus,
  isToggling,
  disabled
}) => {
  const handleEdit = () => onEdit(organization);
  const handleToggleStatus = () => onToggleStatus(organization.code, organization.status);

  const isActive = organization.status === 'ACTIVE';
  const buttonText = isActive ? '停用' : '启用';
  const loadingText = isActive ? '停用中...' : '启用中...';

  return (
    <div style={{ display: 'flex', gap: '4px' }}>
      <TertiaryButton 
        size="small" 
        onClick={handleEdit}
        disabled={disabled}
        data-testid={`edit-button-${organization.code}`}
      >
        编辑
      </TertiaryButton>
      <SecondaryButton 
        size="small" 
        onClick={handleToggleStatus}
        disabled={disabled}
        data-testid={`toggle-status-button-${organization.code}`}
        variant={isActive ? 'inverse' : 'primary'}
      >
        {isToggling ? loadingText : buttonText}
      </SecondaryButton>
    </div>
  );
};