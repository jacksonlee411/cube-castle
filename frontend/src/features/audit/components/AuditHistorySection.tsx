/**
 * 审计历史区域组件
 * 基于auditHistory GraphQL查询展示组织的完整审计记录
 */
import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { activityStreamIcon, exclamationCircleIcon } from '@workday/canvas-system-icons-web';

import { AuditEntryCard } from './AuditEntryCard';
import organizationAPI from '../../../shared/api/organizations';
import type { TemporalQueryParams } from '../../../shared/types/temporal';

export interface AuditHistorySectionProps {
  /** 组织记录ID (recordId) */
  recordId: string;
  /** 时态查询参数 */
  params?: TemporalQueryParams;
  /** 高亮显示的审计ID */
  highlightedAuditId?: string;
}

/**
 * 审计历史区域主组件
 */
export const AuditHistorySection: React.FC<AuditHistorySectionProps> = ({
  recordId,
  params,
  highlightedAuditId
}) => {
  const [expandedEntries, setExpandedEntries] = useState<Set<string>>(new Set());

  // 获取审计历史数据
  const {
    data: auditHistory,
    isLoading,
    error,
    refetch
  } = useQuery({
    queryKey: ['auditHistory', recordId, params],
    queryFn: () => organizationAPI.getAuditHistory(recordId, params),
    enabled: !!recordId,
    staleTime: 30000, // 30秒内数据视为新鲜
    gcTime: 300000,   // 5分钟垃圾回收
  });

  // 处理展开/收起
  const handleToggleExpand = (auditId: string) => {
    setExpandedEntries(prev => {
      const newSet = new Set(prev);
      if (newSet.has(auditId)) {
        newSet.delete(auditId);
      } else {
        newSet.add(auditId);
      }
      return newSet;
    });
  };

  // 数据适配器：GraphQL → UI格式
  const transformAuditData = (audit: Record<string, unknown>) => {
    return {
      auditId: audit.auditId as string,
      operation: audit.operation as string,
      timestamp: audit.timestamp as string,
      userName: (audit.userInfo as { userName: string })?.userName || '未知用户',
      operationReason: audit.operationReason as string,
      dataChanges: audit.dataChanges as {
        beforeData?: Record<string, unknown>;
        afterData?: Record<string, unknown>;
        modifiedFields: string[];
      }
    };
  };

  // 加载状态
  if (isLoading) {
    return (
      <Card padding="m">
        <Flex justifyContent="center" alignItems="center" height="200px">
          <LoadingDots />
          <Text marginLeft="m">加载审计历史中...</Text>
        </Flex>
      </Card>
    );
  }

  // 错误状态
  if (error) {
    return (
      <Card padding="m">
        <Flex alignItems="center" gap="xs" marginBottom="m">
          <SystemIcon icon={exclamationCircleIcon} size={20} color="cinnamon600" />
          <Text color="cinnamon600" typeLevel="heading.medium">
            加载审计历史失败
          </Text>
        </Flex>
        <Text marginBottom="m">
          {error instanceof Error ? error.message : '网络错误，请重试'}
        </Text>
        <button onClick={() => refetch()}>重试</button>
      </Card>
    );
  }

  // 空状态
  if (!auditHistory || auditHistory.length === 0) {
    return (
      <Card padding="m">
        <Flex alignItems="center" gap="xs" marginBottom="m">
          <SystemIcon icon={activityStreamIcon} size={16} />
          <Text as="h3" typeLevel="subtext.large" fontWeight="bold">
            审计历史
          </Text>
        </Flex>
        <Text typeLevel="body.medium" color="hint">
          暂无审计记录
        </Text>
      </Card>
    );
  }

  return (
    <Card padding="m">
      <Flex alignItems="center" gap="xs" marginBottom="m">
        <SystemIcon icon={activityStreamIcon} size={16} />
        <Text as="h3" typeLevel="subtext.large" fontWeight="bold">
          审计历史
        </Text>
        <Text typeLevel="subtext.small" color="hint">
          ({auditHistory.length} 条记录)
        </Text>
      </Flex>

      <Box>
        {auditHistory.map((audit) => {
          const transformedAudit = transformAuditData(audit);
          return (
            <Box key={transformedAudit.auditId} marginBottom="s">
              <AuditEntryCard
                entry={transformedAudit}
                isExpanded={expandedEntries.has(transformedAudit.auditId)}
                isHighlighted={transformedAudit.auditId === highlightedAuditId}
                onExpand={() => handleToggleExpand(transformedAudit.auditId)}
              />
            </Box>
          );
        })}
      </Box>
    </Card>
  );
};

export default AuditHistorySection;