/**
 * 时间点查询组件 - 基于GraphQL时态查询
 * 提供直观的时间点查询界面和结果展示
 */
import React, { useState, useMemo, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';
import { 
  colors, 
  space, 
  borderRadius
} from '@workday/canvas-kit-react/tokens';
import {
  calendarIcon,
  searchIcon,
  infoIcon
} from '@workday/canvas-system-icons-web';
import { SystemIcon } from '@workday/canvas-kit-react/icon';

// 使用新的GraphQL时态查询钩子
import { 
  useOrganizationAsOfDate,
  useTemporalQueryUtils
} from '../../../shared/hooks/useTemporalGraphQL';
import type { 
  TemporalOrganizationUnit
} from '../../../shared/types/temporal';

// 简化字体大小定义
const fontSizes = {
  body: {
    small: '12px',
    medium: '14px'
  },
  heading: {
    medium: '16px'
  }
};

interface TimePointQueryProps {
  organizationCode: string;
  onQueryResult?: (date: string, result: TemporalOrganizationUnit | null) => void;
  initialDate?: string;
  showQuickDates?: boolean;
  compact?: boolean;
}

export const TimePointQuery: React.FC<TimePointQueryProps> = ({
  organizationCode,
  onQueryResult,
  initialDate,
  showQuickDates = true,
  compact = false
}) => {
  const [queryDate, setQueryDate] = useState(() => {
    return initialDate || new Date().toISOString().split('T')[0];
  });

  // GraphQL查询钩子
  const {
    data: asOfDateRecord,
    isLoading: isQuerying,
    error: queryError,
    refetch: executeQuery,
    hasData,
    isEmpty,
    isHistoricalRecord
  } = useOrganizationAsOfDate(organizationCode, queryDate, {
    enabled: false // 手动触发查询
  });

  // 工具钩子
  const { getCommonDatePoints, formatTemporalRecord } = useTemporalQueryUtils();

  // 常用时间点
  const commonDates = useMemo(() => getCommonDatePoints(), [getCommonDatePoints]);

  // 处理查询执行
  const handleExecuteQuery = useCallback(() => {
    executeQuery();
  }, [executeQuery]);

  // 处理日期变更
  const handleDateChange = useCallback((date: string) => {
    setQueryDate(date);
  }, []);

  // 处理快速日期选择
  const handleQuickDateSelect = useCallback((date: string) => {
    setQueryDate(date);
    // 自动执行查询
    setTimeout(() => {
      executeQuery();
    }, 100);
  }, [executeQuery]);

  // 渲染查询结果
  const renderQueryResult = () => {
    if (isQuerying) {
      return (
        <Card padding={space.m} marginTop={space.s}>
          <Flex alignItems="center">
            <LoadingDots />
            <Text marginLeft={space.s}>正在查询 {queryDate} 的组织状态...</Text>
          </Flex>
        </Card>
      );
    }

    if (queryError) {
      return (
        <Card 
          padding={space.m} 
          marginTop={space.s}
          backgroundColor={colors.cinnamon100}
          border={`1px solid ${colors.cinnamon300}`}
        >
          <Flex alignItems="center" marginBottom={space.s}>
            <SystemIcon icon={infoIcon} size={16} color={colors.cinnamon600} />
            <Text marginLeft={space.xs} color={colors.cinnamon600} fontWeight="medium">
              查询失败
            </Text>
          </Flex>
          <Text fontSize={fontSizes.body.small} color={colors.cinnamon600}>
            {queryError.message}
          </Text>
          <SecondaryButton onClick={handleExecuteQuery} marginTop={space.s} size="small">
            重试查询
          </SecondaryButton>
        </Card>
      );
    }

    if (isEmpty) {
      return (
        <Card 
          padding={space.m} 
          marginTop={space.s}
          backgroundColor={colors.licorice100}
          border={`1px solid ${colors.licorice300}`}
        >
          <Flex alignItems="center" marginBottom={space.s}>
            <SystemIcon icon={infoIcon} size={16} color={colors.licorice400} />
            <Text marginLeft={space.xs} color={colors.licorice400} fontWeight="medium">
              无查询结果
            </Text>
          </Flex>
          <Text fontSize={fontSizes.body.small} color={colors.licorice400}>
            在 <strong>{queryDate}</strong> 时间点没有找到组织 <strong>{organizationCode}</strong> 的记录
          </Text>
        </Card>
      );
    }

    if (hasData && asOfDateRecord) {
      const formatted = formatTemporalRecord(asOfDateRecord);
      
      return (
        <Card 
          padding={space.m} 
          marginTop={space.s}
          backgroundColor={isHistoricalRecord ? colors.peach100 : colors.blueberry100}
          border={`1px solid ${isHistoricalRecord ? colors.peach300 : colors.blueberry300}`}
        >
          <Flex justifyContent="space-between" alignItems="flex-start" marginBottom={space.s}>
            <Flex alignItems="center">
              <SystemIcon 
                icon={isHistoricalRecord ? infoIcon : infoIcon} 
                size={16} 
                color={isHistoricalRecord ? colors.peach600 : colors.blueberry600} 
              />
              <Text 
                marginLeft={space.xs} 
                fontWeight="medium"
                color={isHistoricalRecord ? colors.peach600 : colors.blueberry600}
              >
                查询结果 ({queryDate})
              </Text>
            </Flex>
            <Box
              backgroundColor={isHistoricalRecord ? colors.peach500 : colors.blueberry500}
              color={colors.frenchVanilla100}
              paddingX={space.xs}
              paddingY={space.xxxs}
              borderRadius={borderRadius.s}
              fontSize={fontSizes.body.small}
              fontWeight="medium"
            >
              {isHistoricalRecord ? '历史记录' : '当前记录'}
            </Box>
          </Flex>

          <Box>
            <Text 
              fontSize={fontSizes.heading.medium} 
              fontWeight="medium" 
              marginBottom={space.xs}
              color={colors.licorice500}
            >
              {asOfDateRecord.name}
            </Text>
            
            <Flex gap={space.l} flexWrap="wrap" marginBottom={space.s}>
              <Box>
                <Text fontSize={fontSizes.body.small} color={colors.licorice400}>组织代码</Text>
                <Text fontSize={fontSizes.body.medium}>{asOfDateRecord.code}</Text>
              </Box>
              
              <Box>
                <Text fontSize={fontSizes.body.small} color={colors.licorice400}>生效期间</Text>
                <Text fontSize={fontSizes.body.medium}>{formatted.effectivePeriod}</Text>
              </Box>
              
              <Box>
                <Text fontSize={fontSizes.body.small} color={colors.licorice400}>组织类型</Text>
                <Text fontSize={fontSizes.body.medium}>{formatted.organizationType}</Text>
              </Box>
              
              <Box>
                <Text fontSize={fontSizes.body.small} color={colors.licorice400}>状态</Text>
                <Text fontSize={fontSizes.body.medium}>{formatted.organizationStatus}</Text>
              </Box>
            </Flex>

            {asOfDateRecord.changeReason && (
              <Box marginTop={space.s}>
                <Text fontSize={fontSizes.body.small} color={colors.licorice400}>变更原因</Text>
                <Text fontSize={fontSizes.body.medium} fontStyle="italic">
                  {asOfDateRecord.changeReason}
                </Text>
              </Box>
            )}
            
            {asOfDateRecord.description && (
              <Box marginTop={space.s}>
                <Text fontSize={fontSizes.body.small} color={colors.licorice400}>描述</Text>
                <Text fontSize={fontSizes.body.medium}>
                  {asOfDateRecord.description}
                </Text>
              </Box>
            )}
          </Box>
        </Card>
      );
    }

    return null;
  };

  // 渲染快速日期选择器
  const renderQuickDatePicker = () => {
    if (!showQuickDates) return null;

    const quickDateLabels = {
      today: '今天',
      yesterday: '昨天',
      lastWeek: '一周前',
      lastMonth: '一月前',
      yearStart: '年初',
      lastYearEnd: '去年底'
    };

    return (
      <Box marginTop={space.s}>
        <Text fontSize={fontSizes.body.small} color={colors.licorice400} marginBottom={space.xs}>
          快速选择:
        </Text>
        <Flex gap={space.xs} flexWrap="wrap">
          {Object.entries(commonDates).map(([key, date]) => (
            <SecondaryButton
              key={key}
              size="small"
              onClick={() => handleQuickDateSelect(date)}
            >
              {quickDateLabels[key as keyof typeof quickDateLabels] || key}
            </SecondaryButton>
          ))}
        </Flex>
      </Box>
    );
  };

  if (compact) {
    return (
      <Box>
        <Flex alignItems="flex-end" gap={space.s} marginBottom={space.s}>
          <FormField>
            <FormField.Label>查询日期</FormField.Label>
            <FormField.Field>
              <FormField.Input
                as={TextInput}
                type="date"
                value={queryDate}
                onChange={(e) => handleDateChange(e.target.value)}
              />
            </FormField.Field>
          </FormField>
          
          <PrimaryButton onClick={handleExecuteQuery} disabled={isQuerying}>
            {isQuerying ? '查询中...' : '查询'}
          </PrimaryButton>
        </Flex>
        
        {renderQuickDatePicker()}
        {renderQueryResult()}
      </Box>
    );
  }

  return (
    <Card padding={space.m}>
      <Flex alignItems="center" marginBottom={space.m}>
        <SystemIcon icon={calendarIcon} size={20} />
        <Text 
          marginLeft={space.s} 
          fontWeight="medium"
          fontSize={fontSizes.heading.medium}
        >
          时间点查询
        </Text>
      </Flex>

      <Flex alignItems="flex-end" gap={space.s} marginBottom={space.s}>
        <FormField>
          <FormField.Label>查询日期</FormField.Label>
          <FormField.Field>
            <FormField.Input
              as={TextInput}
              type="date"
              value={queryDate}
              onChange={(e) => handleDateChange(e.target.value)}
            />
          </FormField.Field>
        </FormField>
        
        <FormField>
          <FormField.Label>组织代码</FormField.Label>
          <FormField.Field>
            <FormField.Input
              as={TextInput}
              value={organizationCode}
              disabled
              placeholder="组织代码"
            />
          </FormField.Field>
        </FormField>
        
        <PrimaryButton onClick={handleExecuteQuery} disabled={isQuerying}>
          <SystemIcon icon={searchIcon} size={16} />
          {isQuerying ? '查询中...' : '查询'}
        </PrimaryButton>
      </Flex>

      {renderQuickDatePicker()}
      {renderQueryResult()}

      {/* 查询说明 */}
      <Box marginTop={space.l} padding={space.s} backgroundColor={colors.soap100}>
        <Text fontSize={fontSizes.body.small} color={colors.licorice400} display="block" marginBottom={space.xs}>
          <strong>使用说明:</strong>
        </Text>
        <Text fontSize={fontSizes.body.small} color={colors.licorice400} display="block" marginBottom={space.xs}>
          • 选择查询日期，系统将返回该时间点有效的组织记录
        </Text>
        <Text fontSize={fontSizes.body.small} color={colors.licorice400} display="block" marginBottom={space.xs}>
          • 蓝色背景表示当前有效记录，橙色背景表示历史记录
        </Text>
        <Text fontSize={fontSizes.body.small} color={colors.licorice400} display="block">
          • 可使用快速选择按钮快速设置常用日期
        </Text>
      </Box>
    </Card>
  );
};

export default TimePointQuery;