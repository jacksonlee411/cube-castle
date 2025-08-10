/**
 * 时间线组件数据连接测试应用
 * 验证Timeline组件与后端API的集成功能
 */
import React, { useState, useCallback } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Box } from '@workday/canvas-kit-react/layout';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Text } from '@workday/canvas-kit-react/text';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { Card } from '@workday/canvas-kit-react/card';

import { Timeline } from './features/temporal/components/Timeline';
import { useOrganizationTimeline } from './shared/hooks/useTemporalQuery';
import type { TimelineEvent, EventType } from './shared/types/temporal';

// 创建React Query客户端
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 5 * 60 * 1000,
    },
  },
});

/**
 * 时间线数据连接测试组件
 */
const TimelineDataConnectionTest: React.FC = () => {
  const [organizationCode, setOrganizationCode] = useState('1000001');
  const [maxEvents, setMaxEvents] = useState(20);
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);
  const [selectedEventTypes, setSelectedEventTypes] = useState<EventType[]>([]);
  const [dateRange, setDateRange] = useState({
    start: '',
    end: ''
  });

  // 时间线数据查询
  const {
    data: timelineEvents = [],
    isLoading,
    isError,
    error,
    hasEvents,
    eventCount,
    latestEvent,
    refetch
  } = useOrganizationTimeline(organizationCode, {
    limit: maxEvents,
    eventTypes: selectedEventTypes.length > 0 ? selectedEventTypes : undefined,
    dateRange: dateRange.start && dateRange.end ? dateRange : undefined
  }, !!organizationCode);

  // 事件点击处理
  const handleEventClick = useCallback((event: TimelineEvent) => {
    alert(`事件详情:\n\nID: ${event.id}\n标题: ${event.title}\n类型: ${event.eventType}\n日期: ${event.eventDate}\n状态: ${event.status}\n${event.description ? '\n描述: ' + event.description : ''}`);
  }, []);

  // 添加事件处理
  const handleAddEvent = useCallback(() => {
    alert('添加新事件功能将在后续版本中实现');
  }, []);

  // 刷新数据
  const handleRefresh = useCallback(() => {
    refetch();
  }, [refetch]);

  // 清除筛选
  const handleClearFilters = useCallback(() => {
    setSelectedEventTypes([]);
    setDateRange({ start: '', end: '' });
    setMaxEvents(20);
  }, []);

  // 事件类型选择
  const eventTypeOptions: { value: EventType; label: string }[] = [
    { value: 'create', label: '创建' },
    { value: 'update', label: '更新' },
    { value: 'delete', label: '删除' },
    { value: 'activate', label: '激活' },
    { value: 'deactivate', label: '停用' },
    { value: 'restructure', label: '重组' },
    { value: 'merge', label: '合并' },
    { value: 'split', label: '拆分' },
    { value: 'transfer', label: '转移' },
    { value: 'rename', label: '重命名' }
  ];

  const toggleEventType = (eventType: EventType) => {
    setSelectedEventTypes(prev => 
      prev.includes(eventType)
        ? prev.filter(t => t !== eventType)
        : [...prev, eventType]
    );
  };

  return (
    <Box padding="l">
      <Text as="h1" typeLevel="heading.large" marginBottom="l">
        🔗 时间线组件数据连接测试
      </Text>
      
      <Text typeLevel="body.medium" marginBottom="m">
        测试Timeline组件与后端API的数据连接功能，验证时态查询和实时更新。
      </Text>

      {/* 控制面板 */}
      <Card marginBottom="l" padding="m">
        <Text as="h2" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
          🎛️ 测试控制面板
        </Text>
        
        <Box display="flex" gap="m" flexWrap="wrap" marginBottom="m">
          <FormField flex="1" minWidth="200px">
            <FormField.Label>组织编码</FormField.Label>
            <FormField.Field>
              <FormField.Input
                as={TextInput}
                value={organizationCode}
                onChange={(e) => setOrganizationCode(e.target.value)}
                placeholder="输入组织编码，如: 1000001"
              />
            </FormField.Field>
          </FormField>

          <FormField flex="1" minWidth="150px">
            <FormField.Label>最大事件数</FormField.Label>
            <FormField.Field>
              <FormField.Input
                as={TextInput}
                type="number"
                value={maxEvents}
                onChange={(e) => setMaxEvents(parseInt(e.target.value) || 20)}
                min="1"
                max="100"
              />
            </FormField.Field>
          </FormField>
        </Box>

        <Box marginBottom="m">
          <PrimaryButton onClick={handleRefresh} marginRight="s">
            🔄 刷新数据
          </PrimaryButton>
          <SecondaryButton onClick={() => setShowAdvancedFilters(!showAdvancedFilters)} marginRight="s">
            {showAdvancedFilters ? '🔽 隐藏高级筛选' : '▶️ 显示高级筛选'}
          </SecondaryButton>
          <SecondaryButton onClick={handleClearFilters}>
            🗑️ 清除筛选
          </SecondaryButton>
        </Box>

        {/* 高级筛选选项 */}
        {showAdvancedFilters && (
          <Box padding="m" style={{ backgroundColor: '#f8f9fa', borderRadius: '4px', border: '1px solid #e9ecef' }}>
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
              🎯 事件类型筛选
            </Text>
            <Box display="flex" gap="s" flexWrap="wrap" marginBottom="m">
              {eventTypeOptions.map(({ value, label }) => (
                <SecondaryButton
                  key={value}
                  size="small"
                  variant={selectedEventTypes.includes(value) ? "primary" : "secondary"}
                  onClick={() => toggleEventType(value)}
                >
                  {label}
                </SecondaryButton>
              ))}
            </Box>

            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
              📅 日期范围筛选
            </Text>
            <Box display="flex" gap="s" marginBottom="s">
              <FormField flex="1">
                <FormField.Label>开始日期</FormField.Label>
                <FormField.Field>
                  <FormField.Input
                    as={TextInput}
                    type="date"
                    value={dateRange.start}
                    onChange={(e) => setDateRange(prev => ({ ...prev, start: e.target.value }))}
                  />
                </FormField.Field>
              </FormField>
              <FormField flex="1">
                <FormField.Label>结束日期</FormField.Label>
                <FormField.Field>
                  <FormField.Input
                    as={TextInput}
                    type="date"
                    value={dateRange.end}
                    onChange={(e) => setDateRange(prev => ({ ...prev, end: e.target.value }))}
                  />
                </FormField.Field>
              </FormField>
            </Box>
          </Box>
        )}
      </Card>

      {/* 数据状态信息 */}
      <Card marginBottom="l" padding="m">
        <Text as="h3" typeLevel="subtext.large" fontWeight="bold" marginBottom="s">
          📊 数据状态信息
        </Text>
        <Box display="flex" gap="l" flexWrap="wrap">
          <Text typeLevel="body.small">
            🔄 加载状态: <strong>{isLoading ? '加载中' : '已完成'}</strong>
          </Text>
          <Text typeLevel="body.small">
            ✅ 数据状态: <strong>{isError ? '错误' : hasEvents ? '有数据' : '无数据'}</strong>
          </Text>
          <Text typeLevel="body.small">
            📊 事件数量: <strong>{eventCount}</strong>
          </Text>
          {latestEvent && (
            <Text typeLevel="body.small">
              🕒 最新事件: <strong>{latestEvent.title}</strong>
            </Text>
          )}
        </Box>
        {isError && error && (
          <Text typeLevel="body.small" color="cinnamon600" marginTop="s">
            ❌ 错误信息: {error.message}
          </Text>
        )}
      </Card>

      {/* 功能测试要点 */}
      <Card marginBottom="l" padding="m" style={{ backgroundColor: '#f0f7ff', border: '1px solid #d1ecf1' }}>
        <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
          📋 数据连接功能验证要点
        </Text>
        <ul style={{ marginLeft: '20px', lineHeight: '1.6' }}>
          <li>✅ 时间线数据API调用和响应处理</li>
          <li>✅ 实时数据加载状态显示</li>
          <li>✅ 事件筛选和参数传递</li>
          <li>✅ 错误处理和用户反馈</li>
          <li>✅ 数据缓存和性能优化</li>
          <li>✅ 事件交互和详情显示</li>
          <li>✅ 响应式UI和用户体验</li>
          <li>✅ 时态查询参数集成</li>
        </ul>
      </Card>

      {/* 时间线组件 */}
      <Timeline
        organizationCode={organizationCode}
        queryParams={{
          limit: maxEvents,
          eventTypes: selectedEventTypes.length > 0 ? selectedEventTypes : undefined,
          dateRange: dateRange.start && dateRange.end ? dateRange : undefined
        }}
        compact={false}
        maxEvents={maxEvents}
        showFilters={true}
        showActions={true}
        onEventClick={handleEventClick}
        onAddEvent={handleAddEvent}
      />
    </Box>
  );
};

/**
 * 时间线测试应用
 */
export const TimelineTestApp: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <TimelineDataConnectionTest />
    </QueryClientProvider>
  );
};

export default TimelineTestApp;