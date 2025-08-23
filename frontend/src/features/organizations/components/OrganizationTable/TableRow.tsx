import React from 'react';
import { Table } from '@workday/canvas-kit-react/table';
import { Text } from '@workday/canvas-kit-react/text';
import { TableActions } from './TableActions';
import type { OrganizationTableRowProps } from './TableTypes';
import { 
  TemporalInfoDisplay, 
  TemporalDateRange 
} from '../../../temporal/components/TemporalInfoDisplay';

type TemporalStatus = 'ACTIVE' | 'PLANNED' | 'INACTIVE';

// 临时的状态工具函数
const temporalStatusUtils = {
  isTemporal: (effectiveDate?: string, endDate?: string): boolean => {
    return !!(effectiveDate || endDate);
  }
};

export const TableRow: React.FC<OrganizationTableRowProps> = ({
  organization,
  onTemporalManage,
  isAnyToggling,
  isHistorical = false,
  showTemporalInfo = false
}) => {
  // 计算时态状态
  const temporalStatus = organization.status as TemporalStatus;
  const isTemporal = temporalStatusUtils.isTemporal(
    organization.effectiveDate, 
    organization.endDate
  );
  
  // 时态模式下的样式调整
  const getRowStyle = () => {
    const baseStyle = {
      transition: 'opacity 0.3s ease'
    };

    // PLANNED组织的特殊样式
    if (temporalStatus === 'PLANNED') {
      return {
        ...baseStyle,
        backgroundColor: 'rgba(8, 117, 225, 0.05)', // 淡蓝色背景表示计划组织
        borderLeft: '3px solid #0875E1'
      };
    }

    // 历史数据样式
    if (isHistorical) {
      return {
        ...baseStyle,
        backgroundColor: 'rgba(103, 123, 148, 0.05)', 
        border: '1px solid rgba(103, 123, 148, 0.1)'
      };
    }

    // INACTIVE组织的样式
    if (temporalStatus === 'INACTIVE') {
      return {
        ...baseStyle,
        backgroundColor: 'rgba(153, 153, 153, 0.05)',
        color: '#666666'
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
        {isHistorical && (
          <Text as="span" typeLevel="subtext.small" color="hint" marginLeft="xs">
            📖
          </Text>
        )}
        {/* 计划组织标识 */}
        {temporalStatus === 'PLANNED' && (
          <Text as="span" typeLevel="subtext.small" color="positive" marginLeft="xs">
            计划 计划
          </Text>
        )}
      </Table.Cell>
      
      <Table.Cell>{organization.unitType}</Table.Cell>
      
      <Table.Cell>
        <span 
          style={{
            display: 'inline-block',
            padding: '2px 6px',
            borderRadius: '12px',
            fontSize: '11px',
            fontWeight: '500',
            backgroundColor: temporalStatus === 'ACTIVE' ? '#00A844' : 
                           temporalStatus === 'PLANNED' ? '#0875E1' : '#999999',
            color: 'white'
          }}
        >
          {temporalStatus === 'ACTIVE' ? '✓ 启用' :
           temporalStatus === 'PLANNED' ? '计划 计划' :
           temporalStatus === 'INACTIVE' ? '停用 停用' : temporalStatus}
        </span>
      </Table.Cell>
      
      <Table.Cell>{organization.level}</Table.Cell>
      
      {/* 时态信息列 */}
      {(showTemporalInfo || isTemporal) && (
        <Table.Cell>
          {isTemporal ? (
            <TemporalDateRange 
              effectiveDate={organization.effectiveDate}
              endDate={organization.endDate}
              format="short"
            />
          ) : (
            <Text typeLevel="body.small" color="hint">-</Text>
          )}
        </Table.Cell>
      )}
      
      {/* 时态详细信息 */}
      {showTemporalInfo && isTemporal && (
        <Table.Cell>
          <TemporalInfoDisplay 
            temporalInfo={{
              effectiveDate: organization.effectiveDate,
              endDate: organization.endDate,
              status: temporalStatus,
              isTemporal: isTemporal,
              changeReason: organization.changeReason,
              version: organization.version
            }}
            variant="compact"
          />
        </Table.Cell>
      )}
      
      <Table.Cell>
        <TableActions
          organization={organization}
          onTemporalManage={onTemporalManage}
          disabled={isAnyToggling}
          isHistorical={isHistorical}
        />
      </Table.Cell>
    </Table.Row>
  );
};