import React from 'react';
import { TertiaryButton, DeleteButton } from '@workday/canvas-kit-react/button';
import type { TableActionsProps } from './TableTypes';

export const TableActions: React.FC<TableActionsProps> = ({
  organization,
  onEdit,
  onDelete,
  isDeleting,
  disabled
}) => {
  const handleEdit = () => onEdit(organization);
  const handleDelete = () => onDelete(organization.code);

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
      <DeleteButton 
        size="small" 
        onClick={handleDelete}
        disabled={disabled}
        data-testid={`delete-button-${organization.code}`}
      >
        {isDeleting ? '删除中...' : '删除'}
      </DeleteButton>
    </div>
  );
};