/**
 * 时间点查询组件 - 快速查询特定时间点的组织状态
 */
import React, { useState, useCallback, useMemo } from 'react';
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
  ClockIcon,
  InfoIcon
} from '@workday/canvas-system-icons-web';

import { 
  useOrganizationAsOfDate,
  useTemporalQueryUtils
} from '../../../shared/hooks/useTemporalGraphQL';
import type { TemporalOrganizationUnit } from '../../../shared/types/temporal';

interface TimePointQueryProps {
  organizationCode: string;
  initialDate?: string;
  onResult?: (date: string, result: TemporalOrganizationUnit | null) => void;
  onError?: (error: Error) => void;
  showQuickDates?: boolean;
  disabled?: boolean;
}

export const TimePointQuery: React.FC<TimePointQueryProps> = ({
  organizationCode,
  initialDate,
  onResult,
  onError,
  showQuickDates = true,
  disabled = false
}) => {
  const [queryDate, setQueryDate] = useState(() => {
    return initialDate || new Date().toISOString().split('T')[0];
  });

  const [hasExecutedQuery, setHasExecutedQuery] = useState(false);

  // 使用GraphQL时间点查询钩子
  const {
    data: result,
    isLoading,
    error,
    refetch,
    hasData,
    isEmpty,
    isHistoricalRecord
  } = useOrganizationAsOfDate(organizationCode, queryDate, {
    enabled: false, // 手动触发查询
    onSuccess: (data) => {
      onResult?.(queryDate, data);
      setHasExecutedQuery(true);
    },
    onError: (error) => {
      onError?.(error);
      setHasExecutedQuery(true);
    }
  });

  const { getCommonDatePoints, formatTemporalRecord } = useTemporalQueryUtils();

  // 常用时间点
  const commonDates = useMemo(() => {
    const dates = getCommonDatePoints();
    return [
      { label: '今天', value: dates.today },
      { label: '昨天', value: dates.yesterday },
      { label: '一周前', value: dates.lastWeek },
      { label: '一个月前', value: dates.lastMonth },
      { label: '年初', value: dates.yearStart },
      { label: '去年底', value: dates.lastYearEnd },
    ];
  }, [getCommonDatePoints]);

  // 执行查询
  const handleQuery = useCallback(() => {
    if (disabled || !queryDate) return;
    refetch();
  }, [disabled, queryDate, refetch]);

  // 快速日期选择
  const handleQuickDateSelect = useCallback((date: string) => {
    setQueryDate(date);
    setHasExecutedQuery(false);
  }, []);

  // 渲染查询结果
  const renderResult = () => {
    if (!hasExecutedQuery) return null;

    if (isLoading) {
      return (
        <Box padding={space.s}>
          <Flex alignItems="center">
            <LoadingDots />
            <Text marginLeft={space.s}>正在查询 {queryDate} 的组织状态...</Text>
          </Flex>
        </Box>
      );
    }

    if (error) {
      return (
        <Card 
          padding={space.s} 
          backgroundColor={colors.cinnamon100}
          border={`1px solid ${colors.cinnamon300}`}
        >
          <Flex alignItems="center">
            <InfoIcon size={16} color={colors.cinnamon600} />
            <Text marginLeft={space.xs} color={colors.cinnamon600}>
              查询失败: {error.message}
            </Text>
          </Flex>
        </Card>
      );
    }

    if (isEmpty) {
      return (
        <Card 
          padding={space.s} 
          backgroundColor={colors.licorice100}
          border={`1px solid ${colors.licorice300}`}
        >
          <Flex alignItems="center">
            <InfoIcon size={16} color={colors.licorice400} />
            <Text marginLeft={space.xs} color={colors.licorice400}>
              在 {queryDate} 时间点没有找到该组织的记录
            </Text>
          </Flex>
        </Card>
      );
    }

    if (hasData && result) {
      const formatted = formatTemporalRecord(result);
      
      return (
        <Card 
          padding={space.s}
          backgroundColor={isHistoricalRecord ? colors.peach100 : colors.blueberry100}
          border={`1px solid ${isHistoricalRecord ? colors.peach300 : colors.blueberry300}`}
        >
          <Flex justifyContent="space-between" alignItems="flex-start" marginBottom={space.s}>
            <Text 
              fontWeight="medium" 
              fontSize={fontSizes.body.medium}
              color={colors.licorice500}
            >
              {result.name}
            </Text>
            <Box
              backgroundColor={isHistoricalRecord ? colors.peach400 : colors.blueberry400}
              color={colors.frenchVanilla100}
              paddingX={space.xs}
              paddingY={space.xxxs}
              borderRadius={borderRadius.s}
              fontSize={fontSizes.caption}
              fontWeight="medium"
            >
              {isHistoricalRecord ? '历史记录' : '当前记录'}
            </Box>
          </Flex>

          <Box>
            <Text fontSize={fontSizes.body.small} color={colors.licorice300}>
              <strong>生效期间:</strong> {formatted.effectivePeriod}
            </Text>
            <Text fontSize={fontSizes.body.small} color={colors.licorice300}>
              <strong>组织类型:</strong> {formatted.organizationType}
            </Text>
            <Text fontSize={fontSizes.body.small} color={colors.licorice300}>
              <strong>状态:</strong> {formatted.organizationStatus}
            </Text>
            {result.change_reason && (
              <Text fontSize={fontSizes.body.small} color={colors.licorice300}>
                <strong>变更原因:</strong> {result.change_reason}
              </Text>
            )}
          </Box>

          {/* 详细信息 */}
          <Box marginTop={space.s}>
            <Text fontSize={fontSizes.caption} color={colors.licorice300}>
              组织代码: {result.code} | 级别: {result.level}
              {result.parent_code && ` | 上级: ${result.parent_code}`}
            </Text>
          </Box>
        </Card>
      );
    }

    return null;
  };

  return (
    <Box>
      {/* 查询表单 */}
      <Card padding={space.m} marginBottom={space.m}>
        <Flex alignItems="center" marginBottom={space.s}>
          <ClockIcon size={20} />
          <Text 
            marginLeft={space.xs} 
            fontWeight="medium"
            fontSize={fontSizes.body.medium}
          >
            时间点查询
          </Text>
        </Flex>

        <Flex alignItems="flex-end" gap={space.s} marginBottom={space.m}>
          <FormField 
            label="查询日期"
            isRequired
            hint="选择要查询的具体日期"
          >
            <TextInput
              type="date"
              value={queryDate}
              onChange={(e) => setQueryDate(e.target.value)}
              disabled={disabled}
            />
          </FormField>
          
          <PrimaryButton 
            onClick={handleQuery} 
            disabled={disabled || isLoading || !queryDate}
          >
            <SearchIcon size={16} />
            {isLoading ? '查询中...' : '查询'}
          </PrimaryButton>
        </Flex>

        {/* 快速日期选择 */}
        {showQuickDates && (
          <Box>
            <Text 
              fontSize={fontSizes.body.small} 
              color={colors.licorice400}
              marginBottom={space.xs}
            >
              快速选择:
            </Text>
            <Flex gap={space.xs} flexWrap="wrap">
              {commonDates.map(({ label, value }) => (
                <SecondaryButton
                  key={label}
                  size="small"
                  variant={queryDate === value ? 'primary' : 'secondary'}
                  onClick={() => handleQuickDateSelect(value)}
                  disabled={disabled}
                >
                  {label}
                </SecondaryButton>
              ))}
            </Flex>
          </Box>
        )}
      </Card>

      {/* 查询结果 */}
      {renderResult()}
    </Box>
  );
};

export default TimePointQuery;