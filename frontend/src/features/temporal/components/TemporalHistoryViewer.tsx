/**
 * 时态历史查看器组件 - 基于GraphQL时态查询
 * 展示组织的完整历史记录和时间点查询功能
 */
import React, { useState, useMemo, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { Select } from '@workday/canvas-kit-react/select';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';
import { 
  colors, 
  space, 
  borderRadius,
  fontSizes 
} from '@workday/canvas-kit-react/tokens';
import {
  CalendarIcon,
  SearchIcon,
  RefreshIcon,
  HistoryIcon,
  FilterIcon
} from '@workday/canvas-system-icons-web';

// 使用新的GraphQL时态查询钩子
import { 
  useOrganizationHistory,
  useOrganizationAsOfDate,
  useTemporalQueryUtils,
  useTemporalCacheManager
} from '../../../shared/hooks/useTemporalGraphQL';
import type { 
  TemporalOrganizationUnit,
  TimelineEvent
} from '../../../shared/types/temporal';

interface TemporalHistoryViewerProps {
  organizationCode: string;
  onRecordSelect?: (record: TemporalOrganizationUnit) => void;
  onTimePointQuery?: (date: string, result: TemporalOrganizationUnit | null) => void;
  showTimePointQuery?: boolean;
  showFilters?: boolean;
  maxHeight?: string;
}

export const TemporalHistoryViewer: React.FC<TemporalHistoryViewerProps> = ({
  organizationCode,
  onRecordSelect,
  onTimePointQuery,
  showTimePointQuery = true,
  showFilters = true,
  maxHeight = '600px'
}) => {
  // 时间范围过滤状态
  const [dateRange, setDateRange] = useState({
    fromDate: '2020-01-01',
    toDate: '2050-01-01'
  });
  
  // 时间点查询状态
  const [asOfDate, setAsOfDate] = useState(() => {
    return new Date().toISOString().split('T')[0];
  });

  // GraphQL查询钩子
  const {
    data: historyRecords = [],
    isLoading: isHistoryLoading,
    error: historyError,
    refetch: refetchHistory,
    hasHistory,
    historyCount,
    latestRecord,
    currentRecord,
    historicalRecords
  } = useOrganizationHistory(organizationCode, {
    fromDate: dateRange.fromDate,
    toDate: dateRange.toDate
  });

  const {
    data: asOfDateRecord,
    isLoading: isAsOfDateLoading,
    error: asOfDateError,
    refetch: refetchAsOfDate,
    hasData: hasAsOfDateData,
    isHistoricalRecord
  } = useOrganizationAsOfDate(organizationCode, asOfDate, {
    enabled: showTimePointQuery,
    onSuccess: (data) => {
      onTimePointQuery?.(asOfDate, data);
    }
  });

  // 工具钩子
  const { getCommonDatePoints, formatTemporalRecord, compareRecords } = useTemporalQueryUtils();
  const { clearAllTemporalCache, invalidateHistoryCache } = useTemporalCacheManager();

  // 常用时间点
  const commonDates = useMemo(() => getCommonDatePoints(), [getCommonDatePoints]);

  // 处理时间范围变更
  const handleDateRangeChange = useCallback((field: 'fromDate' | 'toDate', value: string) => {
    setDateRange(prev => ({ ...prev, [field]: value }));
  }, []);

  // 处理时间点查询
  const handleAsOfDateQuery = useCallback(() => {
    refetchAsOfDate();
  }, [refetchAsOfDate]);

  // 处理记录选择
  const handleRecordSelect = useCallback((record: TemporalOrganizationUnit) => {
    onRecordSelect?.(record);
  }, [onRecordSelect]);

  // 刷新所有数据
  const handleRefreshAll = useCallback(async () => {
    await Promise.all([
      refetchHistory(),
      showTimePointQuery ? refetchAsOfDate() : Promise.resolve()
    ]);
  }, [refetchHistory, refetchAsOfDate, showTimePointQuery]);

  // 清除缓存并刷新
  const handleClearCacheAndRefresh = useCallback(async () => {
    await clearAllTemporalCache();
    await handleRefreshAll();
  }, [clearAllTemporalCache, handleRefreshAll]);

  // 渲染历史记录列表
  const renderHistoryRecords = () => {
    if (isHistoryLoading) {
      return (
        <Box padding={space.m}>
          <LoadingDots />
          <Text marginTop={space.xs}>正在加载历史记录...</Text>
        </Box>
      );
    }

    if (historyError) {
      return (
        <Box padding={space.m}>
          <Text color={colors.cinnamon600}>
            加载历史记录失败: {historyError.message}
          </Text>
          <SecondaryButton onClick={() => refetchHistory()} marginTop={space.xs}>
            重试
          </SecondaryButton>
        </Box>
      );
    }

    if (!hasHistory) {
      return (
        <Box padding={space.m}>
          <Text color={colors.licorice300}>
            该组织没有历史记录
          </Text>
        </Box>
      );
    }

    return (
      <Box>
        {historyRecords.map((record, index) => {
          const formatted = formatTemporalRecord(record);
          const isFirst = index === 0;
          const isLast = index === historyRecords.length - 1;
          
          return (
            <Card
              key={`${record.code}-${record.effective_date}`}
              padding={space.s}
              marginBottom={space.xs}
              border={`1px solid ${colors.soap400}`}
              borderRadius={borderRadius.m}
              backgroundColor={record.is_current ? colors.blueberry100 : colors.frenchVanilla100}
              cursor="pointer"
              onClick={() => handleRecordSelect(record)}
            >
              <Flex justifyContent="space-between" alignItems="flex-start">
                <Box flex={1}>
                  <Flex alignItems="center" marginBottom={space.xs}>
                    <Text 
                      fontWeight="medium" 
                      fontSize={fontSizes.body.medium}
                      color={colors.licorice500}
                    >
                      {record.name}
                    </Text>
                    {record.is_current && (
                      <Box marginLeft={space.xs}>
                        <Box
                          backgroundColor={colors.greenApple400}
                          color={colors.frenchVanilla100}
                          paddingX={space.xxxs}
                          paddingY="2px"
                          borderRadius={borderRadius.s}
                          fontSize={fontSizes.caption}
                          fontWeight="medium"
                        >
                          当前
                        </Box>
                      </Box>
                    )}
                    {isFirst && !record.is_current && (
                      <Box marginLeft={space.xs}>
                        <Box
                          backgroundColor={colors.cinnamon400}
                          color={colors.frenchVanilla100}
                          paddingX={space.xxxs}
                          paddingY="2px"
                          borderRadius={borderRadius.s}
                          fontSize={fontSizes.caption}
                          fontWeight="medium"
                        >
                          最新
                        </Box>
                      </Box>
                    )}
                  </Flex>
                  
                  <Box marginBottom={space.xs}>
                    <Text fontSize={fontSizes.body.small} color={colors.licorice300}>
                      <strong>生效期间:</strong> {formatted.effectivePeriod}
                    </Text>
                    <Text fontSize={fontSizes.body.small} color={colors.licorice300}>
                      <strong>组织类型:</strong> {formatted.organizationType} | 
                      <strong> 状态:</strong> {formatted.organizationStatus}
                    </Text>
                  </Box>

                  {record.change_reason && (
                    <Text 
                      fontSize={fontSizes.body.small} 
                      color={colors.licorice400}
                      fontStyle="italic"
                    >
                      变更原因: {record.change_reason}
                    </Text>
                  )}

                  {/* 显示与前一个记录的差异 */}
                  {index > 0 && (() => {
                    const changes = compareRecords(historyRecords[index - 1], record);
                    if (changes.length > 0) {
                      return (
                        <Box marginTop={space.xs}>
                          <Text 
                            fontSize={fontSizes.body.small} 
                            color={colors.blueberry500}
                            fontWeight="medium"
                          >
                            主要变更:
                          </Text>
                          {changes.slice(0, 3).map(change => (
                            <Text 
                              key={change.field}
                              fontSize={fontSizes.caption}
                              color={colors.licorice400}
                              display="block"
                            >
                              • {change.displayName}: {change.oldValue} → {change.newValue}
                            </Text>
                          ))}
                          {changes.length > 3 && (
                            <Text 
                              fontSize={fontSizes.caption}
                              color={colors.licorice300}
                            >
                              ...还有 {changes.length - 3} 项变更
                            </Text>
                          )}
                        </Box>
                      );
                    }
                    return null;
                  })()}
                </Box>

                <Box>
                  <Text 
                    fontSize={fontSizes.caption} 
                    color={colors.licorice300}
                    textAlign="right"
                  >
                    {record.effective_date}
                  </Text>
                </Box>
              </Flex>
            </Card>
          );
        })}
      </Box>
    );
  };

  // 渲染时间点查询部分
  const renderAsOfDateQuery = () => {
    if (!showTimePointQuery) return null;

    return (
      <Card padding={space.m} marginBottom={space.m}>
        <Flex alignItems="center" marginBottom={space.s}>
          <CalendarIcon size={20} />
          <Text 
            marginLeft={space.xs} 
            fontWeight="medium"
            fontSize={fontSizes.body.medium}
          >
            时间点查询
          </Text>
        </Flex>

        <Flex alignItems="flex-end" gap={space.s} marginBottom={space.m}>
          <FormField label="查询日期">
            <TextInput
              type="date"
              value={asOfDate}
              onChange={(e) => setAsOfDate(e.target.value)}
            />
          </FormField>
          
          <PrimaryButton onClick={handleAsOfDateQuery} disabled={isAsOfDateLoading}>
            {isAsOfDateLoading ? '查询中...' : '查询'}
          </PrimaryButton>
        </Flex>

        {/* 快速日期选择 */}
        <Flex gap={space.xs} marginBottom={space.m} flexWrap="wrap">
          <Text fontSize={fontSizes.body.small} color={colors.licorice400}>
            快速选择:
          </Text>
          {Object.entries(commonDates).map(([label, date]) => (
            <SecondaryButton
              key={label}
              size="small"
              onClick={() => setAsOfDate(date)}
            >
              {label === 'today' ? '今天' :
               label === 'yesterday' ? '昨天' :
               label === 'lastWeek' ? '一周前' :
               label === 'lastMonth' ? '一月前' :
               label === 'yearStart' ? '年初' :
               label === 'lastYearEnd' ? '去年底' : label}
            </SecondaryButton>
          ))}
        </Flex>

        {/* 查询结果显示 */}
        {isAsOfDateLoading && (
          <Box padding={space.s}>
            <LoadingDots />
            <Text marginLeft={space.xs}>正在查询 {asOfDate} 的组织状态...</Text>
          </Box>
        )}

        {asOfDateError && (
          <Box padding={space.s} backgroundColor={colors.cinnamon100}>
            <Text color={colors.cinnamon600}>
              查询失败: {asOfDateError.message}
            </Text>
          </Box>
        )}

        {!isAsOfDateLoading && hasAsOfDateData && asOfDateRecord && (
          <Card padding={space.s} backgroundColor={isHistoricalRecord ? colors.peach100 : colors.blueberry100}>
            <Flex justifyContent="space-between" alignItems="flex-start">
              <Box>
                <Text fontWeight="medium" fontSize={fontSizes.body.medium}>
                  {asOfDateRecord.name}
                </Text>
                <Text fontSize={fontSizes.body.small} color={colors.licorice300}>
                  {formatTemporalRecord(asOfDateRecord).effectivePeriod}
                </Text>
                <Text fontSize={fontSizes.body.small} color={colors.licorice300}>
                  {formatTemporalRecord(asOfDateRecord).organizationType} | 
                  {formatTemporalRecord(asOfDateRecord).organizationStatus}
                </Text>
              </Box>
              <Box
                backgroundColor={isHistoricalRecord ? colors.peach400 : colors.blueberry400}
                color={colors.frenchVanilla100}
                paddingX={space.xs}
                paddingY={space.xxxs}
                borderRadius={borderRadius.s}
                fontSize={fontSizes.caption}
              >
                {isHistoricalRecord ? '历史记录' : '当前记录'}
              </Box>
            </Flex>
          </Card>
        )}

        {!isAsOfDateLoading && !hasAsOfDateData && (
          <Box padding={space.s} backgroundColor={colors.licorice100}>
            <Text color={colors.licorice400}>
              在 {asOfDate} 时间点没有找到该组织的记录
            </Text>
          </Box>
        )}
      </Card>
    );
  };

  return (
    <Box>
      {/* 标题和操作栏 */}
      <Flex justifyContent="space-between" alignItems="center" marginBottom={space.m}>
        <Flex alignItems="center">
          <HistoryIcon size={24} />
          <Text 
            marginLeft={space.s} 
            fontWeight="medium"
            fontSize={fontSizes.body.large}
          >
            时态历史记录 ({historyCount} 条)
          </Text>
        </Flex>
        
        <Flex gap={space.s}>
          <SecondaryButton onClick={handleRefreshAll}>
            <RefreshIcon size={16} />
            刷新
          </SecondaryButton>
          <SecondaryButton onClick={handleClearCacheAndRefresh}>
            清除缓存
          </SecondaryButton>
        </Flex>
      </Flex>

      {/* 时间点查询 */}
      {renderAsOfDateQuery()}

      {/* 日期范围过滤 */}
      {showFilters && (
        <Card padding={space.m} marginBottom={space.m}>
          <Flex alignItems="center" marginBottom={space.s}>
            <FilterIcon size={16} />
            <Text marginLeft={space.xs} fontWeight="medium">
              时间范围过滤
            </Text>
          </Flex>
          
          <Flex gap={space.s}>
            <FormField label="起始日期">
              <TextInput
                type="date"
                value={dateRange.fromDate}
                onChange={(e) => handleDateRangeChange('fromDate', e.target.value)}
              />
            </FormField>
            <FormField label="结束日期">
              <TextInput
                type="date"
                value={dateRange.toDate}
                onChange={(e) => handleDateRangeChange('toDate', e.target.value)}
              />
            </FormField>
          </Flex>
        </Card>
      )}

      {/* 历史记录列表 */}
      <Box 
        maxHeight={maxHeight} 
        overflowY="auto"
        border={`1px solid ${colors.soap400}`}
        borderRadius={borderRadius.m}
        padding={space.s}
        backgroundColor={colors.frenchVanilla100}
      >
        {renderHistoryRecords()}
      </Box>
    </Box>
  );
};

export default TemporalHistoryViewer;