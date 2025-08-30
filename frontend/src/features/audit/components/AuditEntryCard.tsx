import React from 'react';
import { 
  Card, 
  Flex, 
  Box, 
  Text, 
  Heading
} from '@workday/canvas-kit-react';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { 
  plusIcon,
  editIcon,
  trashIcon,
  mediaPauseIcon,
  mediaPlayIcon
} from '@workday/canvas-system-icons-web';
import { colors, space } from '@workday/canvas-kit-react/tokens';
import type { AuditTimelineEntry, OperationType, RiskLevel } from '../../../shared/api/audit';

// 组件Props接口
interface AuditEntryCardProps {
  entry: AuditTimelineEntry;
  onExpand?: (entry: AuditTimelineEntry) => void;
  isExpanded?: boolean;
  className?: string;
}

// 操作类型配置映射
const operationConfig: Record<OperationType, {
  icon: typeof plusIcon;
  color: string;
  bgColor: string;
  label: string;
}> = {
  CREATE: {
    icon: plusIcon,
    color: colors.greenApple600,
    bgColor: colors.greenApple100,
    label: '创建'
  },
  UPDATE: {
    icon: editIcon,
    color: colors.blueberry600,
    bgColor: colors.blueberry100,
    label: '更新'
  },
  SUSPEND: {
    icon: mediaPauseIcon,
    color: colors.cantaloupe600,
    bgColor: colors.cantaloupe100,
    label: '停用'
  },
  REACTIVATE: {
    icon: mediaPlayIcon,
    color: colors.greenApple600,
    bgColor: colors.greenApple100,
    label: '重新启用'
  },
  DELETE: {
    icon: trashIcon,
    color: colors.cinnamon600,
    bgColor: colors.cinnamon100,
    label: '删除'
  }
};

// 风险级别配置映射
const riskConfig: Record<RiskLevel, {
  color: string;
  bgColor: string;
  label: string;
}> = {
  LOW: {
    color: colors.greenApple600,
    bgColor: colors.greenApple100,
    label: '低风险'
  },
  MEDIUM: {
    color: colors.cantaloupe600,
    bgColor: colors.cantaloupe100,
    label: '中风险'
  },
  HIGH: {
    color: colors.cinnamon600,
    bgColor: colors.cinnamon100,
    label: '高风险'
  }
};

/**
 * 审计记录卡片组件
 * 展示单个审计记录的详细信息，包括操作类型、时间、用户、变更摘要等
 */
export const AuditEntryCard: React.FC<AuditEntryCardProps> = ({
  entry,
  onExpand,
  isExpanded = false,
  className
}) => {
  const opConfig = operationConfig[entry.operation];
  const riskConf = riskConfig[entry.riskLevel];

  // 格式化时间戳
  const formatTimestamp = (timestamp: string): string => {
    try {
      const date = new Date(timestamp);
      return new Intl.DateTimeFormat('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        timeZone: 'Asia/Shanghai'
      }).format(date);
    } catch (_error) {
      return timestamp;
    }
  };

  // 处理卡片点击
  const handleCardClick = () => {
    if (onExpand) {
      onExpand(entry);
    }
  };

  // 处理关键变更标签点击
  const handleKeyChangeClick = (event: React.MouseEvent, change: string) => {
    event.stopPropagation();
    console.log('Key change clicked:', change);
  };

  return (
    <Card
      className={className}
      padding={space.l}
      onClick={onExpand ? handleCardClick : undefined}
      style={{
        cursor: onExpand ? 'pointer' : 'default',
        transition: 'all 0.2s ease-in-out',
        borderLeft: `4px solid ${opConfig.color}`,
        marginBottom: space.m
      }}
      onMouseEnter={(e) => {
        if (onExpand) {
          e.currentTarget.style.transform = 'translateY(-2px)';
          e.currentTarget.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';
        }
      }}
      onMouseLeave={(e) => {
        if (onExpand) {
          e.currentTarget.style.transform = 'translateY(0)';
          e.currentTarget.style.boxShadow = 'none';
        }
      }}
    >
      {/* 卡片头部 - 操作信息和时间 */}
      <Flex justifyContent="space-between" alignItems="flex-start" marginBottom={space.m}>
        <Flex alignItems="center" gap={space.m}>
          {/* 操作类型图标 */}
          <Box
            padding={space.xs}
            style={{
              backgroundColor: opConfig.bgColor,
              borderRadius: '8px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}
          >
            <SystemIcon
              icon={opConfig.icon}
              color={opConfig.color}
              size={20}
            />
          </Box>
          
          {/* 操作信息 */}
          <Box>
            <Flex alignItems="center" gap={space.s}>
              <Heading size="small" color={opConfig.color}>
                {opConfig.label}
              </Heading>
              <Text typeLevel="subtext.small" color={colors.licorice400}>
                审计记录
              </Text>
            </Flex>
            <Text typeLevel="subtext.small" color={colors.licorice600}>
              {entry.userName} • {formatTimestamp(entry.timestamp)}
            </Text>
          </Box>
        </Flex>

        {/* 风险级别标签 */}
        <Box
          padding="xs"
          style={{
            backgroundColor: riskConf.bgColor,
            color: riskConf.color,
            borderRadius: '4px',
            border: `1px solid ${riskConf.color}`,
            fontSize: '12px'
          }}
        >
          <Text typeLevel="subtext.small">{riskConf.label}</Text>
        </Box>
      </Flex>

      {/* 变更摘要 */}
      <Box marginBottom={space.m}>
        <Text typeLevel="body.medium" color={colors.licorice600}>
          {entry.changesSummary.operationSummary}
        </Text>
        
        {entry.changesSummary.totalChanges > 0 && (
          <Text typeLevel="subtext.small" color={colors.licorice400} marginTop={space.xs}>
            共 {entry.changesSummary.totalChanges} 项变更
          </Text>
        )}
      </Box>

      {/* 操作原因 */}
      {entry.operationReason && (
        <Box marginBottom={space.m}>
          <Text typeLevel="subtext.small" color={colors.licorice400}>
            原因: {entry.operationReason}
          </Text>
        </Box>
      )}

      {/* 关键变更标签 */}
      {entry.changesSummary.keyChanges && entry.changesSummary.keyChanges.length > 0 && (
        <Box>
          <Text typeLevel="subtext.small" color={colors.licorice400} marginBottom={space.xs}>
            关键变更:
          </Text>
          <Flex gap={space.xs} flexWrap="wrap">
            {entry.changesSummary.keyChanges.slice(0, 5).map((change, index) => (
              <Box
                key={index}
                onClick={(e) => handleKeyChangeClick(e, change)}
                padding="xs"
                style={{
                  backgroundColor: colors.soap200,
                  color: colors.licorice600,
                  cursor: 'pointer',
                  fontSize: '12px',
                  maxWidth: '120px',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                  borderRadius: '4px',
                  border: '1px solid ' + colors.soap300
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.backgroundColor = colors.soap300;
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.backgroundColor = colors.soap200;
                }}
              >
                <Text typeLevel="subtext.small">{change}</Text>
              </Box>
            ))}
            {entry.changesSummary.keyChanges.length > 5 && (
              <Box
                padding="xs"
                style={{
                  backgroundColor: colors.soap100,
                  color: colors.licorice400,
                  cursor: 'pointer',
                  borderRadius: '4px',
                  border: '1px solid ' + colors.soap200
                }}
              >
                <Text typeLevel="subtext.small">+{entry.changesSummary.keyChanges.length - 5} 更多</Text>
              </Box>
            )}
          </Flex>
        </Box>
      )}

      {/* 展开指示器 */}
      {onExpand && (
        <Box
          position="absolute"
          bottom={space.xs}
          right={space.s}
          style={{
            opacity: 0.6,
            transition: 'opacity 0.2s ease-in-out'
          }}
        >
          <Text typeLevel="subtext.small" color={colors.licorice400}>
            {isExpanded ? '点击收起' : '点击查看详情'}
          </Text>
        </Box>
      )}
    </Card>
  );
};