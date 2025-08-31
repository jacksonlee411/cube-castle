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
// ✅ P2修复: 移除缺失的audit.ts类型依赖，定义本地类型
export interface AuditTimelineEntry {
  auditId: string;
  operation: string;
  timestamp: string;
  userName: string;
  operationReason?: string;
  dataChanges: {
    beforeData?: Record<string, unknown>;
    afterData?: Record<string, unknown>;
    modifiedFields: string[];
  };
}

export type OperationType = 'CREATE' | 'UPDATE' | 'SUSPEND' | 'REACTIVATE' | 'DELETE';
export type RiskLevel = 'LOW' | 'MEDIUM' | 'HIGH';

// 组件Props接口
interface AuditEntryCardProps {
  entry: AuditTimelineEntry;
  onExpand?: () => void;
  isExpanded?: boolean;
  isHighlighted?: boolean;
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
  isHighlighted = false,
  className
}) => {
  const opConfig = operationConfig[entry.operation as OperationType] || operationConfig.UPDATE;
  
  // 简单风险评估：基于修改字段数量
  const calculateRiskLevel = (modifiedFields: string[]): RiskLevel => {
    if (modifiedFields.length >= 5) return 'HIGH';
    if (modifiedFields.length >= 3) return 'MEDIUM';
    return 'LOW';
  };
  
  const riskLevel = calculateRiskLevel(entry.dataChanges.modifiedFields);
  const riskConf = riskConfig[riskLevel];

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
      onExpand();
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
        marginBottom: space.m,
        backgroundColor: isHighlighted ? colors.blueberry100 : undefined
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
          {entry.operationReason || `执行了${opConfig.label}操作`}
        </Text>
        
        {entry.dataChanges.modifiedFields.length > 0 && (
          <Text typeLevel="subtext.small" color={colors.licorice400} marginTop={space.xs}>
            共 {entry.dataChanges.modifiedFields.length} 项变更
          </Text>
        )}
      </Box>

      {/* 修改字段列表 */}
      {entry.dataChanges.modifiedFields.length > 0 && (
        <Box marginBottom={space.m}>
          <Text typeLevel="subtext.small" color={colors.licorice400} marginBottom={space.xs}>
            变更字段:
          </Text>
          <Flex gap={space.xs} flexWrap="wrap">
            {entry.dataChanges.modifiedFields.slice(0, 5).map((field, index) => (
              <Box
                key={index}
                onClick={(e) => handleKeyChangeClick(e, field)}
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
                <Text typeLevel="subtext.small">{field}</Text>
              </Box>
            ))}
            {entry.dataChanges.modifiedFields.length > 5 && (
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
                <Text typeLevel="subtext.small">+{entry.dataChanges.modifiedFields.length - 5} 更多</Text>
              </Box>
            )}
          </Flex>
        </Box>
      )}

      {/* 展开详情：显示before/after数据 */}
      {isExpanded && (entry.dataChanges.beforeData || entry.dataChanges.afterData) && (
        <Box marginTop={space.m} padding={space.m} style={{
          backgroundColor: colors.soap100,
          borderRadius: '8px',
          border: `1px solid ${colors.soap300}`
        }}>
          <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom={space.s}>
            数据变更详情:
          </Text>
          
          {entry.dataChanges.beforeData && (
            <Box marginBottom={space.s}>
              <Text typeLevel="subtext.small" color={colors.cinnamon600} marginBottom={space.xs}>
                变更前:
              </Text>
              <Box style={{
                backgroundColor: colors.frenchVanilla100,
                padding: space.xs,
                borderRadius: '4px',
                fontSize: '12px',
                maxHeight: '100px',
                overflow: 'auto'
              }}>
                <pre>{JSON.stringify(entry.dataChanges.beforeData, null, 2)}</pre>
              </Box>
            </Box>
          )}
          
          {entry.dataChanges.afterData && (
            <Box>
              <Text typeLevel="subtext.small" color={colors.greenApple600} marginBottom={space.xs}>
                变更后:
              </Text>
              <Box style={{
                backgroundColor: colors.frenchVanilla100,
                padding: space.xs,
                borderRadius: '4px',
                fontSize: '12px',
                maxHeight: '100px',
                overflow: 'auto'
              }}>
                <pre>{JSON.stringify(entry.dataChanges.afterData, null, 2)}</pre>
              </Box>
            </Box>
          )}
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