import React from 'react';
import { Table } from '@workday/canvas-kit-react/table';
import { Text } from '@workday/canvas-kit-react/text';
import { TableRow } from './TableRow';
import type { OrganizationTableProps } from './TableTypes';

const TableHeader: React.FC<{ showTemporalInfo?: boolean }> = ({ showTemporalInfo = false }) => (
  <Table.Head>
    <Table.Row>
      <Table.Header>编码</Table.Header>
      <Table.Header>名称</Table.Header>
      <Table.Header>类型</Table.Header>
      <Table.Header>状态</Table.Header>
      <Table.Header>层级</Table.Header>
      {showTemporalInfo && (
        <>
          <Table.Header>生效时间</Table.Header>
          <Table.Header>失效时间</Table.Header>
          <Table.Header>时态状态</Table.Header>
        </>
      )}
      <Table.Header>操作</Table.Header>
    </Table.Row>
  </Table.Head>
);

export const OrganizationTable: React.FC<OrganizationTableProps> = ({
  organizations,
  onTemporalManage,
  temporalMode = 'current',
  isHistorical = false,
  showTemporalInfo = false
}) => {
  return (
    <Table data-testid="temporal-organization-table">
      <TableHeader showTemporalInfo={showTemporalInfo || isHistorical} />
      <Table.Body>
        {!organizations || organizations.length === 0 ? (
          <Table.Row>
            <Table.Cell colSpan={showTemporalInfo || isHistorical ? 9 : 6}>
              <Text textAlign="center" color="hint">
                {isHistorical ? 
                  '在指定时间点没有找到组织数据' : 
                  '没有组织数据'
                }
              </Text>
            </Table.Cell>
          </Table.Row>
        ) : (
          organizations?.map((org, index) => {
            // 使用多层级唯一性保证：recordId > code+createdAt > code+index
            const uniqueKey = org.recordId || 
                             `${org.code}-${org.createdAt}` || 
                             `${org.code}-${index}`;
            
            return (
              <TableRow
                key={uniqueKey}
                organization={org}
                onTemporalManage={onTemporalManage}
                isAnyToggling={false} // 移除状态切换功能，始终为false
                temporalMode={temporalMode}
                isHistorical={isHistorical}
                showTemporalInfo={showTemporalInfo || isHistorical}
              />
            );
          })
        )}
      </Table.Body>
    </Table>
  );
};
