import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Flex,
  Text,
  Heading
} from '@workday/canvas-kit-react';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';
import { resetIcon, downloadIcon } from '@workday/canvas-system-icons-web';
import { colors, space } from '@workday/canvas-kit-react/tokens';
import { useAuditHistory } from '../hooks/useAuditHistory';
import { AuditEntryCard } from './AuditEntryCard';
import { useMessages } from '../../../shared/hooks/useMessages';
import type { AuditQueryParams, AuditTimelineEntry } from '../../../shared/api/audit';

// 组件Props接口
interface AuditHistoryTimelineProps {
  organizationCode: string;
  initialParams?: AuditQueryParams;
  onEntryClick?: (entry: AuditTimelineEntry) => void;
  className?: string;
}

/**
 * 审计信息时间线组件
 * 显示审计记录时间线，支持无限滚动
 */
export const AuditHistoryTimeline: React.FC<AuditHistoryTimelineProps> = ({
  organizationCode,
  initialParams = {},
  onEntryClick,
  className
}) => {
  const [expandedEntry, setExpandedEntry] = useState<string | null>(null);
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  // 使用消息管理Hook - 替代alert()
  const { showError } = useMessages();

  // 使用审计信息Hook
  const {
    auditHistory,
    auditTimeline,
    loading,
    error,
    hasMore,
    totalRecords,
    refetch,
    loadMore,
    clearError
  } = useAuditHistory({
    code: organizationCode,
    initialParams,
    autoFetch: true
  });


  // 处理刷新按钮点击
  const handleRefreshClick = () => {
    refetch();
  };

  // 处理记录卡片点击
  const handleEntryClick = (entry: AuditTimelineEntry) => {
    if (onEntryClick) {
      onEntryClick(entry);
    }
    
    // 切换展开状态
    setExpandedEntry(prev => 
      prev === entry.auditId ? null : entry.auditId
    );
  };

  // 处理加载更多
  const handleLoadMore = useCallback(async () => {
    if (isLoadingMore || !hasMore || loading) return;
    
    setIsLoadingMore(true);
    try {
      await loadMore();
    } finally {
      setIsLoadingMore(false);
    }
  }, [isLoadingMore, hasMore, loading, loadMore]);

  // 处理下载审计报告（占位功能）
  const handleDownloadReport = () => {
    // TODO: 实现审计报告下载功能
    showError('审计报告下载功能开发中...');
  };

  // 无限滚动监听
  useEffect(() => {
    const handleScroll = () => {
      if (
        window.innerHeight + document.documentElement.scrollTop 
        >= document.documentElement.offsetHeight - 1000
      ) {
        handleLoadMore();
      }
    };

    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, [isLoadingMore, hasMore, loading, handleLoadMore]);

  // 错误处理显示
  if (error) {
    return (
      <Box className={className} padding={space.l}>
        <Flex
          justifyContent="center"
          alignItems="center"
          flexDirection="column"
          minHeight="200px"
        >
          <Text color={colors.cinnamon600} marginBottom={space.m}>
            加载审计信息失败: {error}
          </Text>
          <PrimaryButton onClick={clearError}>
            重试
          </PrimaryButton>
        </Flex>
      </Box>
    );
  }

  return (
    <Box className={className}>

      {/* 时间线头部 */}
      <Flex justifyContent="space-between" alignItems="center" marginBottom={space.l}>
        <Box>
          <Heading size="medium">
            审计信息时间线
          </Heading>
          {auditHistory && (
            <Text typeLevel="subtext.small" color={colors.licorice400}>
              组织: {auditHistory.entityName} ({organizationCode}) • 
              共 {totalRecords} 条记录 • 
              {auditHistory.totalVersions} 个版本
            </Text>
          )}
        </Box>

        <Flex gap={space.s}>
          <SecondaryButton
            size="small"
            onClick={handleRefreshClick}
            disabled={loading}
            icon={resetIcon}
          >
            刷新
          </SecondaryButton>
          <SecondaryButton
            size="small"
            onClick={handleDownloadReport}
            disabled={loading}
            icon={downloadIcon}
          >
            导出报告
          </SecondaryButton>
        </Flex>
      </Flex>

      {/* 初始加载状态 */}
      {loading && auditTimeline.length === 0 && (
        <Flex justifyContent="center" alignItems="center" minHeight="200px">
          <LoadingDots />
          <Text marginLeft={space.m} color={colors.licorice400} typeLevel="body.medium">
            正在加载审计信息...
          </Text>
        </Flex>
      )}

      {/* 时间线内容 */}
      {auditTimeline.length > 0 && (
        <Box>
          {/* 统计摘要 */}
          {auditHistory?.meta && (
            <Box
              padding={space.m}
              marginBottom={space.l}
              style={{
                backgroundColor: colors.soap100,
                borderRadius: '8px',
                border: `1px solid ${colors.soap300}`
              }}
            >
              <Text typeLevel="subtext.small" color={colors.licorice600}>
                <strong>操作统计:</strong> 
                {auditHistory.meta.operationsSummary.create > 0 && ` 创建 ${auditHistory.meta.operationsSummary.create}`}
                {auditHistory.meta.operationsSummary.update > 0 && ` 更新 ${auditHistory.meta.operationsSummary.update}`}
                {auditHistory.meta.operationsSummary.suspend > 0 && ` 停用 ${auditHistory.meta.operationsSummary.suspend}`}
                {auditHistory.meta.operationsSummary.reactivate > 0 && ` 重启 ${auditHistory.meta.operationsSummary.reactivate}`}
                {auditHistory.meta.operationsSummary.delete > 0 && ` 删除 ${auditHistory.meta.operationsSummary.delete}`}
              </Text>
              <Text typeLevel="subtext.small" color={colors.licorice400}>
                时间范围: {auditHistory.meta.dateRange.earliest} ~ {auditHistory.meta.dateRange.latest}
              </Text>
            </Box>
          )}

          {/* 审计记录列表 */}
          <Box>
            {auditTimeline.map((entry) => (
              <AuditEntryCard
                key={entry.auditId}
                entry={entry}
                onExpand={handleEntryClick}
                isExpanded={expandedEntry === entry.auditId}
              />
            ))}
          </Box>

          {/* 加载更多按钮或状态 */}
          <Flex justifyContent="center" marginTop={space.l}>
            {hasMore ? (
              <SecondaryButton
                onClick={handleLoadMore}
                disabled={isLoadingMore || loading}
              >
                {isLoadingMore ? (
                  <>
                    <LoadingDots />
                    <Text marginLeft={space.s} typeLevel="body.medium">加载中...</Text>
                  </>
                ) : (
                  '加载更多'
                )}
              </SecondaryButton>
            ) : (
              <Text typeLevel="subtext.small" color={colors.licorice400}>
                {auditTimeline.length > 0 ? '已显示所有记录' : '暂无审计记录'}
              </Text>
            )}
          </Flex>
        </Box>
      )}

      {/* 无数据状态 */}
      {!loading && auditTimeline.length === 0 && (
        <Flex
          justifyContent="center"
          alignItems="center"
          flexDirection="column"
          minHeight="200px"
        >
          <Text color={colors.licorice400} marginBottom={space.s}>
            未找到符合条件的审计记录
          </Text>
          <Text typeLevel="subtext.small" color={colors.licorice300}>
            尝试调整筛选条件或检查组织代码是否正确
          </Text>
        </Flex>
      )}
    </Box>
  );
};