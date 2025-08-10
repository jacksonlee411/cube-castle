import React from 'react';
import { Table } from '@workday/canvas-kit-react/table';
import { Text } from '@workday/canvas-kit-react/text';
// import { Badge } from '@workday/canvas-kit-react';
import { TableActions } from './TableActions';
import type { OrganizationTableRowProps } from './TableTypes';

// 时态状态显示组件
const TemporalStatusBadge: React.FC<{
  organization: any; // 支持时态字段的组织对象
  isHistorical: boolean;
}> = ({ organization, isHistorical }) => {
  if (!isHistorical || !organization.temporalStatus) {
    return null;
  }

  const getBadgeProps = (status: string) => {
    switch (status) {
      case 'active':
        return { color: 'positive', text: '生效中' };
      case 'planned':
        return { color: 'neutral', text: '计划中' };
      case 'expired':
        return { color: 'critical', text: '已失效' };
      default:
        return { color: 'neutral', text: '未知' };
    }
  };

  const badgeProps = getBadgeProps(organization.temporalStatus);
  return (
    <Text color={badgeProps.color as any}>
      {badgeProps.text}
    </Text>
  );
};

// 格式化日期显示
const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-';
  try {
    const date = new Date(dateStr);
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit'
    });
  } catch {
    return dateStr;
  }
};

export const TableRow: React.FC<OrganizationTableRowProps> = ({
  organization,
  onEdit,
  onToggleStatus,
  isToggling,
  isAnyToggling,
  temporalMode = 'current',
  isHistorical = false,
  showTemporalInfo = false
}) => {
  // 时态模式下的样式调整
  const getRowStyle = () => {
    const baseStyle = {
      opacity: isToggling ? 0.6 : 1,
      transition: 'opacity 0.3s ease'
    };

    if (isHistorical) {
      return {
        ...baseStyle,
        backgroundColor: 'rgba(103, 123, 148, 0.05)', // 淡蓝色背景表示历史数据
        border: '1px solid rgba(103, 123, 148, 0.1)'
      };
    }

    return baseStyle;
  };

  return (
    <Table.Row 
      style={getRowStyle()}
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
        {isHistorical && (
          <Text as="span" typeLevel="subtext.small" color="hint" marginLeft="xs">
            📖
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
      
      {/* 时态信息列 */}
      {(showTemporalInfo || isHistorical) && (
        <>
          <Table.Cell>
            <Text typeLevel="subtext.small">
              {formatDate((organization as any).effectiveFrom)}
            </Text>
          </Table.Cell>
          <Table.Cell>
            <Text typeLevel="subtext.small">
              {formatDate((organization as any).effectiveTo)}
            </Text>
          </Table.Cell>
          <Table.Cell>
            <TemporalStatusBadge 
              organization={organization} 
              isHistorical={isHistorical} 
            />
          </Table.Cell>
        </>
      )}
      
      <Table.Cell>
        <TableActions
          organization={organization}
          onEdit={onEdit}
          onToggleStatus={onToggleStatus}
          isToggling={isToggling}
          disabled={isAnyToggling}
          isHistorical={isHistorical}
        />
      </Table.Cell>
    </Table.Row>
  );
};