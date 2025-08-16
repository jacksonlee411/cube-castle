import React from 'react';
import { Flex } from '@workday/canvas-kit-react/layout';
// import { Badge } from '../../../shared/components/Badge';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { 
  temporalStatusUtils 
} from './TemporalStatusSelector';
import type { TemporalStatus } from './TemporalStatusSelector';
import { validateTemporalDate } from './TemporalDatePicker';

export interface TemporalInfo {
  effective_date?: string;
  end_date?: string;
  status?: TemporalStatus;
  is_temporal?: boolean;
  change_reason?: string;
  version?: number;
}

export interface TemporalInfoDisplayProps {
  temporalInfo: TemporalInfo;
  variant?: 'default' | 'compact' | 'detailed';
  showChangeReason?: boolean;
  showVersion?: boolean;
}

// 简单的Badge实现
const SimpleBadge: React.FC<{ 
  children: React.ReactNode; 
  variant?: string;
  color?: string;
  style?: React.CSSProperties;
}> = ({ children, style = {}, color = '#666666' }) => (
  <span 
    style={{
      display: 'inline-block',
      padding: '2px 8px',
      borderRadius: '12px',
      fontSize: '12px',
      fontWeight: '500',
      backgroundColor: color,
      color: 'white',
      ...style
    }}
  >
    {children}
  </span>
);

export const TemporalInfoDisplay: React.FC<TemporalInfoDisplayProps> = ({
  temporalInfo,
  variant = 'default',
  showChangeReason = false,
  showVersion = false,
}) => {
  const {
    effective_date,
    end_date,
    status,
    is_temporal,
    change_reason,
    version,
  } = temporalInfo;

  // 如果不是时态组织，显示简单状态
  if (!is_temporal && !effective_date && !end_date) {
    return (
      <SimpleBadge style={{ backgroundColor: '#999999' }}>
        标准组织
      </SimpleBadge>
    );
  }

  // 计算实际状态
  const actualStatus = status || temporalStatusUtils.calculateStatus(effective_date, end_date);
  const statusColor = temporalStatusUtils.getStatusColor(actualStatus);
  const statusIcon = temporalStatusUtils.getStatusIcon(actualStatus);
  const statusLabel = temporalStatusUtils.getStatusLabel(actualStatus);

  // 紧凑模式
  if (variant === 'compact') {
    return (
      <Flex alignItems="center" gap="xs">
        <SimpleBadge style={{ backgroundColor: statusColor }}>
          {statusIcon} {statusLabel}
        </SimpleBadge>
        {effective_date && (
          <Text as="span" typeLevel="subtext.large" color="licorice300">
            {validateTemporalDate.formatDateDisplay(effective_date)}
          </Text>
        )}
      </Flex>
    );
  }

  // 详细模式
  if (variant === 'detailed') {
    return (
      <Card padding="s">
        <Flex flexDirection="column" gap="xs">
          <Flex alignItems="center" gap="xs">
            <SimpleBadge style={{ backgroundColor: statusColor }}>
              {statusIcon} {statusLabel}
            </SimpleBadge>
            {version && showVersion && (
              <Text as="span" typeLevel="subtext.large" color="licorice300">
                版本 {version}
              </Text>
            )}
          </Flex>

          <Flex flexDirection="column" gap="xxs">
            {effective_date && (
              <Text as="div" typeLevel="subtext.large">
                <strong>生效日期：</strong>
                {validateTemporalDate.formatDateDisplay(effective_date)}
              </Text>
            )}
            {end_date && (
              <Text as="div" typeLevel="subtext.large">
                <strong>结束日期：</strong>
                {validateTemporalDate.formatDateDisplay(end_date)}
              </Text>
            )}
            {change_reason && showChangeReason && (
              <Text as="div" typeLevel="subtext.large" color="licorice300">
                <strong>变更原因：</strong>
                {change_reason}
              </Text>
            )}
          </Flex>
        </Flex>
      </Card>
    );
  }

  // 默认模式
  return (
    <Flex alignItems="center" gap="s">
      <SimpleBadge 
        style={{ 
          backgroundColor: `${statusColor}15`,
          color: statusColor,
          border: `1px solid ${statusColor}`
        }}
      >
        {statusIcon} {statusLabel}
      </SimpleBadge>
      
      <Flex flexDirection="column" gap="xxs">
        {effective_date && (
          <Text as="span" typeLevel="subtext.large" color="licorice500">
            生效: {validateTemporalDate.formatDateDisplay(effective_date)}
          </Text>
        )}
        {end_date && (
          <Text as="span" typeLevel="subtext.large" color="licorice500">
            结束: {validateTemporalDate.formatDateDisplay(end_date)}
          </Text>
        )}
      </Flex>
    </Flex>
  );
};

// 时态状态徽章组件
export interface TemporalStatusBadgeProps {
  status: TemporalStatus;
  size?: 'small' | 'medium' | 'large';
  showIcon?: boolean;
}

export const TemporalStatusBadge: React.FC<TemporalStatusBadgeProps> = ({
  status,
  size = 'medium',
  showIcon = true,
}) => {
  const statusColor = temporalStatusUtils.getStatusColor(status);
  const statusIcon = temporalStatusUtils.getStatusIcon(status);
  const statusLabel = temporalStatusUtils.getStatusLabel(status);

  const padding = size === 'small' ? '2px 6px' : size === 'large' ? '6px 12px' : '4px 8px';
  const fontSize = size === 'small' ? '11px' : size === 'large' ? '14px' : '12px';

  return (
    <SimpleBadge 
      style={{
        backgroundColor: statusColor,
        padding,
        fontSize
      }}
    >
      {showIcon && statusIcon} {statusLabel}
    </SimpleBadge>
  );
};

// 时态日期范围显示组件
export interface TemporalDateRangeProps {
  effectiveDate?: string;
  endDate?: string;
  format?: 'short' | 'long';
  separator?: string;
}

export const TemporalDateRange: React.FC<TemporalDateRangeProps> = ({
  effectiveDate,
  endDate,
  format = 'short',
  separator = ' - ',
}) => {
  const formatDate = format === 'long' 
    ? validateTemporalDate.formatDateDisplay
    : (date: string) => date;

  if (!effectiveDate && !endDate) {
    return <Text as="span" typeLevel="subtext.large" color="licorice300">无时态限制</Text>;
  }

  if (effectiveDate && !endDate) {
    return (
      <Text as="span" typeLevel="subtext.large" color="licorice500">
        {formatDate(effectiveDate)} 起生效
      </Text>
    );
  }

  if (!effectiveDate && endDate) {
    return (
      <Text as="span" typeLevel="subtext.large" color="licorice500">
        至 {formatDate(endDate)} 结束
      </Text>
    );
  }

  return (
    <Text as="span" typeLevel="subtext.large" color="licorice500">
      {formatDate(effectiveDate!)}{separator}{formatDate(endDate!)}
    </Text>
  );
};