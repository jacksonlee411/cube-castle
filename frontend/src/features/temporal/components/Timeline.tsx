/**
 * 时间线可视化组件
 * 展示组织架构的时间线事件和历史变更
 */
import React, { useState, useMemo, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton, ToolbarIconButton as IconButton } from '@workday/canvas-kit-react/button';
import { Badge } from '../../../shared/components/Badge';
import { Tooltip } from '@workday/canvas-kit-react/tooltip';
import { Menu } from '@workday/canvas-kit-react/menu';
import { 
  colors, 
  space, 
  borderRadius,
  fontSizes 
} from '@workday/canvas-kit-react/tokens';
import {
  AddIcon,
  EditIcon,
  DeleteIcon,
  MoreVerticalIcon,
  FilterIcon,
  ExpandIcon,
  CollapseIcon
} from '@workday/canvas-kit-react/icon';
import { useOrganizationTimeline } from '../../shared/hooks/useTemporalQuery';
import type { 
  TimelineEvent, 
  EventType, 
  EventStatus,
  TemporalQueryParams
} from '../../shared/types/temporal';

export interface TimelineProps {
  /** 组织代码 */
  organizationCode: string;
  /** 时间线查询参数 */
  queryParams?: Partial<TemporalQueryParams>;
  /** 是否紧凑模式 */
  compact?: boolean;
  /** 最大显示事件数 */
  maxEvents?: number;
  /** 是否显示筛选器 */
  showFilters?: boolean;
  /** 是否显示操作按钮 */
  showActions?: boolean;
  /** 事件点击回调 */
  onEventClick?: (event: TimelineEvent) => void;
  /** 新增事件回调 */
  onAddEvent?: () => void;
}

/**
 * 时间线事件项组件
 */
interface TimelineEventItemProps {
  event: TimelineEvent;
  isFirst: boolean;
  isLast: boolean;
  compact: boolean;
  showActions: boolean;
  onEventClick?: (event: TimelineEvent) => void;
}

const TimelineEventItem: React.FC<TimelineEventItemProps> = ({
  event,
  isFirst,
  isLast,
  compact,
  showActions,
  onEventClick
}) => {
  const [showMenu, setShowMenu] = useState(false);

  // 获取事件类型样式
  const getEventTypeStyle = (eventType: EventType) => {
    const styles = {
      create: { color: colors.greenFresca600, bgColor: colors.greenFresca100, icon: '🏗️' },
      update: { color: colors.blueberry600, bgColor: colors.blueberry100, icon: '✏️' },
      delete: { color: colors.cinnamon600, bgColor: colors.cinnamon100, icon: '🗑️' },
      activate: { color: colors.greenFresca600, bgColor: colors.greenFresca100, icon: '✅' },
      deactivate: { color: colors.licorice400, bgColor: colors.licorice100, icon: '🚫' },
      restructure: { color: colors.peach600, bgColor: colors.peach100, icon: '🔄' },
      merge: { color: colors.plum600, bgColor: colors.plum100, icon: '🔗' },
      split: { color: colors.cantaloupe600, bgColor: colors.cantaloupe100, icon: '✂️' },
      transfer: { color: colors.blueberry600, bgColor: colors.blueberry100, icon: '📤' },
      rename: { color: colors.peach600, bgColor: colors.peach100, icon: '📝' }
    };
    return styles[eventType] || styles.update;
  };

  // 获取状态样式
  const getStatusStyle = (status: EventStatus) => {
    const styles = {
      pending: { color: colors.cantaloupe600, label: '待处理' },
      approved: { color: colors.blueberry600, label: '已批准' },
      rejected: { color: colors.cinnamon600, label: '已拒绝' },
      completed: { color: colors.greenFresca600, label: '已完成' },
      cancelled: { color: colors.licorice400, label: '已取消' }
    };
    return styles[status] || styles.pending;
  };

  const eventTypeStyle = getEventTypeStyle(event.eventType);
  const statusStyle = getStatusStyle(event.status);

  // 格式化时间
  const formatEventTime = (dateStr: string) => {
    try {
      const date = new Date(dateStr);
      return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch {
      return dateStr;
    }
  };

  return (
    <Flex alignItems="flex-start" gap={space.s}>
      {/* 时间线连接线 */}
      <Box position="relative" display="flex" flexDirection="column" alignItems="center">
        {/* 事件图标 */}
        <Box
          width="32px"
          height="32px"
          borderRadius="50%"
          backgroundColor={eventTypeStyle.bgColor}
          border={`2px solid ${eventTypeStyle.color}`}
          display="flex"
          alignItems="center"
          justifyContent="center"
          fontSize={compact ? fontSizes.small : fontSizes.medium}
        >
          {eventTypeStyle.icon}
        </Box>
        
        {/* 连接线 */}
        {!isLast && (
          <Box
            width="2px"
            height="40px"
            backgroundColor={colors.soap300}
            marginTop={space.xs}
          />
        )}
      </Box>

      {/* 事件内容 */}
      <Card
        flex="1"
        padding={compact ? space.s : space.m}
        marginBottom={space.s}
        cursor={onEventClick ? 'pointer' : 'default'}
        onClick={() => onEventClick?.(event)}
        _hover={onEventClick ? { backgroundColor: colors.soap100 } : {}}
      >
        <Flex justifyContent="space-between" alignItems="flex-start" marginBottom={space.xs}>
          <Box flex="1">
            <Flex alignItems="center" gap={space.s} marginBottom={space.xs}>
              <Text fontWeight="medium" fontSize={compact ? 'small' : 'medium'}>
                {event.title}
              </Text>
              <Badge color={statusStyle.color} variant="outline" size="small">
                {statusStyle.label}
              </Badge>
            </Flex>

            <Text fontSize="small" color={colors.licorice600} marginBottom={space.xs}>
              {formatEventTime(event.eventDate)}
              {event.effectiveDate && event.effectiveDate !== event.eventDate && (
                <> • 生效时间: {formatEventTime(event.effectiveDate)}</>
              )}
            </Text>

            {event.description && !compact && (
              <Text fontSize="small" color={colors.licorice500}>
                {event.description}
              </Text>
            )}
          </Box>

          {/* 操作菜单 */}
          {showActions && (
            <Box position="relative">
              <IconButton
                variant="plain"
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  setShowMenu(!showMenu);
                }}
              >
                <MoreVerticalIcon />
              </IconButton>

              {showMenu && (
                <Menu onClose={() => setShowMenu(false)}>
                  <Menu.Item onClick={() => console.log('查看详情', event.id)}>
                    查看详情
                  </Menu.Item>
                  <Menu.Item onClick={() => console.log('编辑事件', event.id)}>
                    编辑事件
                  </Menu.Item>
                  <Menu.Item onClick={() => console.log('删除事件', event.id)}>
                    删除事件
                  </Menu.Item>
                </Menu>
              )}
            </Box>
          )}
        </Flex>

        {/* 元数据信息 */}
        {event.metadata && !compact && (
          <Box marginTop={space.xs}>
            <Text fontSize="small" color={colors.licorice400}>
              {typeof event.metadata === 'string' 
                ? event.metadata 
                : JSON.stringify(event.metadata, null, 2)
              }
            </Text>
          </Box>
        )}

        {/* 操作者信息 */}
        {(event.triggeredBy || event.approvedBy) && !compact && (
          <Flex gap={space.m} marginTop={space.xs}>
            {event.triggeredBy && (
              <Text fontSize="small" color={colors.licorice400}>
                触发者: {event.triggeredBy}
              </Text>
            )}
            {event.approvedBy && (
              <Text fontSize="small" color={colors.licorice400}>
                批准者: {event.approvedBy}
              </Text>
            )}
          </Flex>
        )}
      </Card>
    </Flex>
  );
};

/**
 * 时间线可视化组件
 */
export const Timeline: React.FC<TimelineProps> = ({
  organizationCode,
  queryParams,
  compact = false,
  maxEvents = 50,
  showFilters = true,
  showActions = false,
  onEventClick,
  onAddEvent
}) => {
  const [eventFilter, setEventFilter] = useState<EventType[]>([]);
  const [statusFilter, setStatusFilter] = useState<EventStatus[]>([]);
  const [expanded, setExpanded] = useState(!compact);

  // 获取时间线数据
  const {
    data: events = [],
    isLoading,
    isError,
    error,
    hasEvents,
    eventCount,
    latestEvent
  } = useOrganizationTimeline(organizationCode, queryParams);

  // 筛选事件
  const filteredEvents = useMemo(() => {
    let filtered = events;

    if (eventFilter.length > 0) {
      filtered = filtered.filter(event => eventFilter.includes(event.eventType));
    }

    if (statusFilter.length > 0) {
      filtered = filtered.filter(event => statusFilter.includes(event.status));
    }

    return filtered.slice(0, maxEvents);
  }, [events, eventFilter, statusFilter, maxEvents]);

  // 获取事件类型统计
  const eventTypeStats = useMemo(() => {
    const stats: Record<EventType, number> = {} as Record<EventType, number>;
    events.forEach(event => {
      stats[event.eventType] = (stats[event.eventType] || 0) + 1;
    });
    return stats;
  }, [events]);

  // 处理筛选器变更
  const handleEventTypeFilter = useCallback((eventType: EventType) => {
    setEventFilter(prev => 
      prev.includes(eventType) 
        ? prev.filter(t => t !== eventType)
        : [...prev, eventType]
    );
  }, []);

  const handleStatusFilter = useCallback((status: EventStatus) => {
    setStatusFilter(prev => 
      prev.includes(status) 
        ? prev.filter(s => s !== status)
        : [...prev, status]
    );
  }, []);

  if (isLoading) {
    return (
      <Card padding={space.m}>
        <Text>🔄 加载时间线数据...</Text>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card padding={space.m}>
        <Text color={colors.cinnamon600}>
          ❌ 加载时间线失败: {error?.message || '未知错误'}
        </Text>
      </Card>
    );
  }

  if (!hasEvents) {
    return (
      <Card padding={space.m}>
        <Flex justifyContent="center" alignItems="center" flexDirection="column" gap={space.m}>
          <Text color={colors.licorice500}>📭 暂无时间线事件</Text>
          {onAddEvent && (
            <SecondaryButton size="small" onClick={onAddEvent}>
              <AddIcon /> 添加事件
            </SecondaryButton>
          )}
        </Flex>
      </Card>
    );
  }

  return (
    <Box>
      {/* 时间线标题和操作 */}
      <Flex justifyContent="space-between" alignItems="center" marginBottom={space.m}>
        <Flex alignItems="center" gap={space.s}>
          <Text fontSize="large" fontWeight="medium">
            时间线
          </Text>
          <Badge variant="outline" color={colors.blueberry600}>
            {eventCount} 个事件
          </Badge>
        </Flex>

        <Flex gap={space.s}>
          {/* 筛选器按钮 */}
          {showFilters && (
            <Tooltip title="筛选事件">
              <IconButton variant="plain" size="small">
                <FilterIcon />
              </IconButton>
            </Tooltip>
          )}

          {/* 展开/收起按钮 */}
          <Tooltip title={expanded ? '收起时间线' : '展开时间线'}>
            <IconButton 
              variant="plain" 
              size="small"
              onClick={() => setExpanded(!expanded)}
            >
              {expanded ? <CollapseIcon /> : <ExpandIcon />}
            </IconButton>
          </Tooltip>

          {/* 添加事件按钮 */}
          {onAddEvent && (
            <Tooltip title="添加新事件">
              <SecondaryButton size="small" onClick={onAddEvent}>
                <AddIcon />
                {!compact && '添加事件'}
              </SecondaryButton>
            </Tooltip>
          )}
        </Flex>
      </Flex>

      {/* 时间线内容 */}
      {expanded && (
        <Card padding={space.m}>
          {/* 快速筛选标签 */}
          {showFilters && (
            <Box marginBottom={space.m}>
              <Text fontSize="small" marginBottom={space.xs} color={colors.licorice600}>
                事件类型筛选:
              </Text>
              <Flex gap={space.xs} flexWrap="wrap">
                {Object.entries(eventTypeStats).map(([eventType, count]) => (
                  eventFilter.includes(eventType as EventType) ? (
                    <PrimaryButton
                      key={eventType}
                      size="small"
                      onClick={() => handleEventTypeFilter(eventType as EventType)}
                    >
                      {eventType} ({count})
                    </PrimaryButton>
                  ) : (
                    <SecondaryButton
                      key={eventType}
                      size="small"
                      onClick={() => handleEventTypeFilter(eventType as EventType)}
                    >
                      {eventType} ({count})
                    </SecondaryButton>
                  )
                ))}
              </Flex>
            </Box>
          )}

          {/* 时间线事件列表 */}
          <Box>
            {filteredEvents.length === 0 ? (
              <Text color={colors.licorice500} textAlign="center">
                📭 没有符合筛选条件的事件
              </Text>
            ) : (
              filteredEvents.map((event, index) => (
                <TimelineEventItem
                  key={event.id}
                  event={event}
                  isFirst={index === 0}
                  isLast={index === filteredEvents.length - 1}
                  compact={compact}
                  showActions={showActions}
                  onEventClick={onEventClick}
                />
              ))
            )}
          </Box>

          {/* 显示更多按钮 */}
          {events.length > filteredEvents.length && (
            <Flex justifyContent="center" marginTop={space.m}>
              <Text fontSize="small" color={colors.licorice500}>
                显示 {filteredEvents.length} / {events.length} 个事件
              </Text>
            </Flex>
          )}
        </Card>
      )}

      {/* 收起状态的简要信息 */}
      {!expanded && latestEvent && (
        <Card padding={space.s}>
          <Flex alignItems="center" gap={space.s}>
            <Text fontSize="small" color={colors.licorice600}>
              最新: {latestEvent.title}
            </Text>
            <Badge variant="outline" size="small">
              {new Date(latestEvent.eventDate).toLocaleDateString('zh-CN')}
            </Badge>
          </Flex>
        </Card>
      )}
    </Box>
  );
};

export default Timeline;