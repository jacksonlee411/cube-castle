import React from 'react';
import { Table } from '@workday/canvas-kit-react/table';
import { Text } from '@workday/canvas-kit-react/text';
import { TableActions } from './TableActions';
import type { OrganizationTableRowProps } from './TableTypes';

export const TableRow: React.FC<OrganizationTableRowProps> = ({
  organization,
  onEdit,
  onToggleStatus,
  isToggling,
  isAnyToggling
}) => {
  return (
    <Table.Row 
      style={{ 
        opacity: isToggling ? 0.6 : 1,
        transition: 'opacity 0.3s ease'
      }}
      data-testid={`table-row-${organization.code}`}
    >
      <Table.Cell>{organization.code}</Table.Cell>
      <Table.Cell>
        {organization.name}
        {isToggling && (
          <Text typeLevel="subtext.small" color="hint" marginLeft="xs">
            (状态更新中...)
          </Text>
        )}
      </Table.Cell>
      <Table.Cell>{organization.unit_type}</Table.Cell>
      <Table.Cell>
        <Text color={
          organization.status === 'ACTIVE' ? 'positive' : 
          organization.status === 'PLANNED' ? 'hint' : 
          'default'
        }>
          {organization.status === 'ACTIVE' ? '启用' : 
           organization.status === 'INACTIVE' ? '停用' : 
           organization.status}
        </Text>
      </Table.Cell>
      <Table.Cell>{organization.level}</Table.Cell>
      <Table.Cell>
        <TableActions
          organization={organization}
          onEdit={onEdit}
          onToggleStatus={onToggleStatus}
          isToggling={isToggling}
          disabled={isAnyToggling}
        />
      </Table.Cell>
    </Table.Row>
  );
};