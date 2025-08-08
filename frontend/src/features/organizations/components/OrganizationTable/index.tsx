import React from 'react';
import { Table } from '@workday/canvas-kit-react/table';
import { TableRow } from './TableRow';
import type { OrganizationTableProps } from './TableTypes';

const TableHeader: React.FC = () => (
  <Table.Head>
    <Table.Row>
      <Table.Header>编码</Table.Header>
      <Table.Header>名称</Table.Header>
      <Table.Header>类型</Table.Header>
      <Table.Header>状态</Table.Header>
      <Table.Header>层级</Table.Header>
      <Table.Header>操作</Table.Header>
    </Table.Row>
  </Table.Head>
);

export const OrganizationTable: React.FC<OrganizationTableProps> = ({
  organizations,
  onEdit,
  onDelete,
  deletingId
}) => {
  return (
    <Table data-testid="organization-table">
      <TableHeader />
      <Table.Body>
        {organizations.map((org, index) => {
          const isDeleting = deletingId === org.code;
          return (
            <TableRow
              key={org.code || `org-${index}`}
              organization={org}
              onEdit={onEdit}
              onDelete={onDelete}
              isDeleting={isDeleting}
              isAnyDeleting={!!deletingId}
            />
          );
        })}
      </Table.Body>
    </Table>
  );
};