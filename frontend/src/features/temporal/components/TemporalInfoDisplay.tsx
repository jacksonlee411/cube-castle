import React from 'react';
import { Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { 
  temporalStatusUtils 
} from './TemporalStatusSelector';
import type { TemporalStatus } from './TemporalStatusSelector';
import { validateTemporalDate } from '@/shared/utils/temporal-validation-adapter';

export interface TemporalInfo {
  effectiveDate?: string;
  endDate?: string;
  status?: TemporalStatus;
  isTemporal?: boolean;
  changeReason?: string;
}

export interface TemporalInfoDisplayProps {
  temporalInfo: TemporalInfo;
  variant?: 'default' | 'compact' | 'detailed';
  showChangeReason?: boolean;
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
}) => {
  const {
    effectiveDate,
    endDate,
    status,
    isTemporal,
    changeReason,
  } = temporalInfo;

  // 如果不是时态组织，显示简单状态
  if (!isTemporal && !effectiveDate && !endDate) {
    return (
      <SimpleBadge style={{ backgroundColor: '#999999' }}>
        标准组织
      </SimpleBadge>
    );
  }

  // 计算实际状态
  const actualStatus = status || temporalStatusUtils.calculateStatus(effectiveDate, endDate);
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
        {effectiveDate && (
          <Text as="span" typeLevel="subtext.large" color="licorice300">
            {validateTemporalDate.formatDateDisplay(effectiveDate)}
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
          </Flex>

          <Flex flexDirection="column" gap="xxs">
            {effectiveDate && (
              <Text as="div" typeLevel="subtext.large">
                <strong>生效日期：</strong>
                {validateTemporalDate.formatDateDisplay(effectiveDate)}
              </Text>
            )}
            {endDate && (
              <Text as="div" typeLevel="subtext.large">
                <strong>结束日期：</strong>
                {validateTemporalDate.formatDateDisplay(endDate)}
              </Text>
            )}
            {changeReason && showChangeReason && (
              <Text as="div" typeLevel="subtext.large" color="licorice300">
                <strong>变更原因：</strong>
                {changeReason}
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
        {effectiveDate && (
          <Text as="span" typeLevel="subtext.large" color="licorice500">
            生效: {validateTemporalDate.formatDateDisplay(effectiveDate)}
          </Text>
        )}
        {endDate && (
          <Text as="span" typeLevel="subtext.large" color="licorice500">
            结束: {validateTemporalDate.formatDateDisplay(endDate)}
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
