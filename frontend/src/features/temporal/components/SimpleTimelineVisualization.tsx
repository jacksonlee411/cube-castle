/**
 * 简化的时态数据可视化组件
 * 展示组织架构时间线事件
 */
import React, { useState, useEffect, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Badge } from '../../../shared/components/Badge';
import { colors } from '@workday/canvas-kit-react/tokens';

// 简化的时间线事件类型
interface TimelineEvent {
  id: string;
  title: string;
  description: string;
  event_type: string;
  event_date: string;
  effective_date: string;
  status: string;
  metadata?: Record<string, unknown>;
  triggered_by?: string;
}

interface SimpleTimelineVisualizationProps {
  organizationCode: string;
  onRefresh?: () => void;
}

export const SimpleTimelineVisualization: React.FC<SimpleTimelineVisualizationProps> = ({
  organizationCode,
  onRefresh
}) => {
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedEvent, setExpandedEvent] = useState<string | null>(null);

  // 获取事件类型样式
  const getEventTypeStyle = (eventType: string) => {
    const styles = {
      create: { color: '#2ECC71', bgColor: '#E8F5E8', icon: '🏗️' },
      update: { color: '#3498DB', bgColor: '#E3F2FD', icon: '✏️' },
      delete: { color: '#E74C3C', bgColor: '#FFEBEE', icon: '🗑️' },
      activate: { color: '#2ECC71', bgColor: '#E8F5E8', icon: '✅' },
      deactivate: { color: '#666666', bgColor: '#F5F5F5', icon: '🚫' },
      restructure: { color: '#F39C12', bgColor: '#FFF3E0', icon: '🔄' },
      dissolve: { color: '#E74C3C', bgColor: '#FFEBEE', icon: '💥' }
    };
    return styles[eventType as keyof typeof styles] || styles.update;
  };

  // 格式化时间
  const formatDate = (dateStr: string) => {
    try {
      const date = new Date(dateStr);
      return date.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch {
      return dateStr;
    }
  };

  // 加载时间线数据
  const loadTimeline = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      
      const response = await fetch(
        `http://localhost:9091/api/v1/organization-units/${organizationCode}/timeline?limit=50`
      );
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const data = await response.json();
      setEvents(data.timeline || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载时间线失败');
    } finally {
      setIsLoading(false);
    }
  }, [organizationCode]);

  useEffect(() => {
    loadTimeline();
  }, [organizationCode, loadTimeline]);

  const handleRefresh = () => {
    loadTimeline();
    onRefresh?.();
  };

  if (isLoading) {
    return (
      <Card padding="l">
        <Flex justifyContent="center" alignItems="center" minHeight="200px">
          <Text>🔄 加载时间线数据...</Text>
        </Flex>
      </Card>
    );
  }

  if (error) {
    return (
      <Card padding="l">
        <Flex justifyContent="center" alignItems="center" flexDirection="column" minHeight="200px" gap="m">
          <Text color={colors.cinnamon600}>❌ {error}</Text>
          <PrimaryButton size="small" onClick={handleRefresh}>
            重试
          </PrimaryButton>
        </Flex>
      </Card>
    );
  }

  if (events.length === 0) {
    return (
      <Card padding="l">
        <Flex justifyContent="center" alignItems="center" flexDirection="column" minHeight="200px" gap="m">
          <Text color="#666666">📭 暂无时间线事件</Text>
          <SecondaryButton size="small" onClick={handleRefresh}>
            刷新
          </SecondaryButton>
        </Flex>
      </Card>
    );
  }

  return (
    <Box>
      {/* 标题和控制 */}
      <Flex justifyContent="space-between" alignItems="center" marginBottom="l">
        <Box>
          <Heading size="medium" marginBottom="s">
            组织时间线可视化
          </Heading>
          <Text color="#666666" fontSize="small">
            组织代码: {organizationCode} · 共 {events.length} 个事件
          </Text>
        </Box>
        
        <SecondaryButton size="small" onClick={handleRefresh}>
          🔄 刷新
        </SecondaryButton>
      </Flex>

      {/* 时间线容器 */}
      <Card padding="l">
        <Box position="relative">
          {/* 时间线主轴 */}
          <Box
            position="absolute"
            left="20px"
            top="20px"
            bottom="20px"
            width="2px"
            backgroundColor="#E0E0E0"
            zIndex={0}
          />

          {/* 事件列表 */}
          {events.map((event, index) => {
            const eventStyle = getEventTypeStyle(event.event_type);
            const isExpanded = expandedEvent === event.id;
            const isLast = index === events.length - 1;

            return (
              <Box
                key={event.id}
                position="relative"
                marginBottom={isLast ? "0" : "l"}
                zIndex={1}
              >
                {/* 事件节点 */}
                <Flex alignItems="flex-start" gap="m">
                  {/* 时间轴点 */}
                  <Box
                    cs={{
                      width: "40px",
                      height: "40px",
                      borderRadius: "50%",
                      backgroundColor: eventStyle.bgColor,
                      border: `2px solid ${eventStyle.color}`,
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      fontSize: "18px",
                      flexShrink: 0
                    }}
                  >
                    {eventStyle.icon}
                  </Box>

                  {/* 事件内容 */}
                  <Card
                    flex="1"
                    padding="m"
                    cursor="pointer"
                    onClick={() => setExpandedEvent(isExpanded ? null : event.id)}
                    style={{
                      backgroundColor: isExpanded ? '#F8F9FA' : 'white',
                      border: isExpanded ? '2px solid #2196F3' : '1px solid #E9ECEF',
                      transition: 'all 0.2s ease'
                    }}
                  >
                    {/* 事件标题 */}
                    <Flex justifyContent="space-between" alignItems="flex-start" marginBottom="s">
                      <Box flex="1">
                        <Text fontWeight="medium" marginBottom="xs">
                          {event.title}
                        </Text>
                        <Flex alignItems="center" gap="s">
                          <Badge color="primary" size="small">
                            {event.event_type}
                          </Badge>
                          <Badge variant="outline" size="small">
                            {event.status}
                          </Badge>
                        </Flex>
                      </Box>
                    </Flex>

                    {/* 事件时间 */}
                    <Box marginBottom="s">
                      <Text fontSize="small" color="#666666">
                        🕐 事件时间: {formatDate(event.event_date)}
                      </Text>
                      <Text fontSize="small" color="#666666">
                        📅 生效时间: {formatDate(event.effective_date)}
                      </Text>
                    </Box>

                    {/* 事件描述 */}
                    <Text fontSize="small" color="#666666" marginBottom="s">
                      {event.description || '无描述'}
                    </Text>

                    {/* 展开的详情 */}
                    {isExpanded && event.metadata && (
                      <Box marginTop="m" padding="m" backgroundColor="#F0F0F0" borderRadius="4px">
                        <Text fontSize="small" fontWeight="medium" marginBottom="s">
                          详细信息:
                        </Text>
                        {Object.entries(event.metadata).map(([key, value]) => (
                          <Text key={key} fontSize="small" marginBottom="xs">
                            • {key}: {String(value)}
                          </Text>
                        ))}
                        {event.triggered_by && (
                          <Text fontSize="small" marginTop="s" color="#999999">
                            触发者: {event.triggered_by}
                          </Text>
                        )}
                      </Box>
                    )}

                    {/* 点击提示 */}
                    <Text fontSize="small" color="#999999" marginTop="s">
                      {isExpanded ? '点击收起' : '点击查看详情'}
                    </Text>
                  </Card>
                </Flex>
              </Box>
            );
          })}
        </Box>
      </Card>

      {/* 统计信息 */}
      <Box marginTop="m">
        <Card padding="m">
          <Flex justifyContent="space-between" alignItems="center">
            <Text fontSize="small" color="#666666">
              时间线统计: 最新事件 {formatDate(events[0]?.event_date || '')}
            </Text>
            <Text fontSize="small" color="#666666">
              数据来源: 时态管理服务 (端口9091)
            </Text>
          </Flex>
        </Card>
      </Box>
    </Box>
  );
};

export default SimpleTimelineVisualization;