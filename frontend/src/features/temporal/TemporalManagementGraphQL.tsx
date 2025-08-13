/**
 * 时态管理演示页面 - 基于GraphQL时态查询
 * 集成展示organizationAsOfDate和organizationHistory功能
 */
import React, { useState, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { Select } from '@workday/canvas-kit-react/select';
import { Tabs } from '@workday/canvas-kit-react/tabs';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { 
  colors, 
  space, 
  borderRadius
} from '@workday/canvas-kit-react/tokens';

// 简化字体大小定义
const fontSizes = {
  body: {
    small: '12px',
    medium: '14px'
  },
  heading: {
    large: '24px'
  }
};
import {
  calendarIcon,
  searchIcon,
  infoIcon
} from '@workday/canvas-system-icons-web';

// 导入新的时态组件
import { TemporalHistoryViewer } from './components/TemporalHistoryViewer';
import { TimePointQuery } from './components/TimePointQuery';
import type { TemporalOrganizationUnit } from '../../shared/types/temporal';

// 路径解析工具函数
const parseOrganizationPath = (path: string, level: number) => {
  if (!path) return { 
    pathSegments: [], 
    pathDescription: '无路径信息',
    levelDescription: '未知层级'
  };

  const segments = path.split('/').filter(segment => segment !== '');
  const levelDescriptions = [
    '未知层级',      // 0级 - 不应该存在
    '根组织',        // 1级
    '一级部门',      // 2级  
    '二级部门',      // 3级
    '三级部门',      // 4级
    '四级部门',      // 5级
    '多级部门'       // 6级及以上
  ];

  const levelDescription = level <= 5 ? levelDescriptions[level] : levelDescriptions[6];
  
  // 创建路径描述
  let pathDescription = '';
  if (segments.length === 1) {
    pathDescription = `根组织 (${segments[0]})`;
  } else if (segments.length === 2) {
    pathDescription = `${segments[0]} → ${segments[1]}`;
  } else if (segments.length >= 3) {
    pathDescription = `${segments[0]} → ... → ${segments[segments.length-1]} (共${segments.length}级)`;
  }

  return {
    pathSegments: segments,
    pathDescription,
    levelDescription
  };
};

// 简化的类型定义 - 删除重复定义

const DEMO_ORGANIZATION_CODES = [
  { label: '1000056 - 完整历史记录演示 (14条记录)', value: '1000056' },
  { label: '1000099 - 新建测试组织', value: '1000099' },
  { label: '1000001 - 标准组织', value: '1000001' },
  { label: '1000002 - 部门组织', value: '1000002' },
];

export const TemporalManagementGraphQL: React.FC = () => {
  const [selectedOrganizationCode, setSelectedOrganizationCode] = useState('1000056');
  const [customCode, setCustomCode] = useState('');
  const [useCustomCode, setUseCustomCode] = useState(false);
  const [selectedRecord, setSelectedRecord] = useState<TemporalOrganizationUnit | null>(null);
  const [timePointQueryResult, setTimePointQueryResult] = useState<{
    date: string;
    result: TemporalOrganizationUnit | null;
  } | null>(null);

  // 获取当前使用的组织代码
  const currentCode = useCustomCode && customCode ? customCode : selectedOrganizationCode;

  // 处理记录选择
  const handleRecordSelect = useCallback((record: TemporalOrganizationUnit) => {
    setSelectedRecord(record);
  }, []);

  // 处理时间点查询结果
  const handleTimePointQuery = useCallback((date: string, result: TemporalOrganizationUnit | null) => {
    setTimePointQueryResult({ date, result });
  }, []);

  // 处理组织代码选择
  const handleOrganizationSelect = useCallback((code: string) => {
    setSelectedOrganizationCode(code);
    setUseCustomCode(false);
    setSelectedRecord(null);
    setTimePointQueryResult(null);
  }, []);

  // 处理自定义代码使用
  const handleUseCustomCode = useCallback(() => {
    setUseCustomCode(true);
    setSelectedRecord(null);
    setTimePointQueryResult(null);
  }, []);

  // 渲染选中记录详情
  const renderSelectedRecordDetails = () => {
    if (!selectedRecord) return null;

    // 解析路径信息
    const pathInfo = parseOrganizationPath(selectedRecord.path, selectedRecord.level);

    return (
      <Card padding={space.m} marginTop={space.m}>
        <Flex alignItems="center" marginBottom={space.s}>
          <SystemIcon icon={infoIcon} size={16} />
          <Text marginLeft={space.xs} fontWeight="medium">
            选中记录详情
          </Text>
        </Flex>

        <Box>
          <Text fontSize={fontSizes.body.medium} fontWeight="medium" marginBottom={space.xs}>
            {selectedRecord.name}
          </Text>
          
          <Flex gap={space.l} flexWrap="wrap">
            <Box>
              <Text fontSize={fontSizes.body.small} color={colors.licorice400}>组织代码</Text>
              <Text fontSize={fontSizes.body.medium}>{selectedRecord.code}</Text>
            </Box>
            
            <Box>
              <Text fontSize={fontSizes.body.small} color={colors.licorice400}>生效日期</Text>
              <Text fontSize={fontSizes.body.medium}>{selectedRecord.effective_date}</Text>
            </Box>
            
            {selectedRecord.end_date && (
              <Box>
                <Text fontSize={fontSizes.body.small} color={colors.licorice400}>结束日期</Text>
                <Text fontSize={fontSizes.body.medium}>{selectedRecord.end_date}</Text>
              </Box>
            )}
            
            <Box>
              <Text fontSize={fontSizes.body.small} color={colors.licorice400}>组织类型</Text>
              <Text fontSize={fontSizes.body.medium}>{selectedRecord.unit_type}</Text>
            </Box>
            
            <Box>
              <Text fontSize={fontSizes.body.small} color={colors.licorice400}>状态</Text>
              <Text fontSize={fontSizes.body.medium}>{selectedRecord.status}</Text>
            </Box>
            
            <Box>
              <Text fontSize={fontSizes.body.small} color={colors.licorice400}>组织层级</Text>
              <Text fontSize={fontSizes.body.medium}>
                第 {selectedRecord.level} 级 ({pathInfo.levelDescription})
              </Text>
            </Box>
            
            <Box>
              <Text fontSize={fontSizes.body.small} color={colors.licorice400}>当前有效</Text>
              <Text fontSize={fontSizes.body.medium}>
                {selectedRecord.is_current ? '是' : '否'}
              </Text>
            </Box>
          </Flex>

          {/* 级联长路径 - 单独一行显示 */}
          <Box marginTop={space.s}>
            <Text fontSize={fontSizes.body.small} color={colors.licorice400}>级联长路径</Text>
            <Text 
              fontSize={fontSizes.body.medium} 
              fontFamily="monospace"
              backgroundColor={colors.soap200}
              padding={space.xs}
              borderRadius={borderRadius.s}
              wordBreak="break-all"
              marginTop={space.xs}
            >
              {selectedRecord.path || 'N/A'}
            </Text>
            
            {/* 路径解释 */}
            <Box marginTop={space.xs}>
              <Text fontSize={fontSizes.body.small} color={colors.licorice400}>
                <strong>路径解释：</strong> {pathInfo.pathDescription}
              </Text>
              <Text fontSize={fontSizes.body.small} color={colors.licorice400}>
                说明：显示从根组织到当前组织的完整层级路径，数字为组织代码
              </Text>
            </Box>
          </Box>

          {selectedRecord.change_reason && (
            <Box marginTop={space.s}>
              <Text fontSize={fontSizes.body.small} color={colors.licorice400}>变更原因</Text>
              <Text fontSize={fontSizes.body.medium}>{selectedRecord.change_reason}</Text>
            </Box>
          )}

          {selectedRecord.description && (
            <Box marginTop={space.s}>
              <Text fontSize={fontSizes.body.small} color={colors.licorice400}>描述</Text>
              <Text fontSize={fontSizes.body.medium}>{selectedRecord.description}</Text>
            </Box>
          )}
        </Box>
      </Card>
    );
  };

  // 渲染时间点查询结果
  const renderTimePointQueryResult = () => {
    if (!timePointQueryResult) return null;

    return (
      <Card padding={space.m} marginTop={space.m}>
        <Flex alignItems="center" marginBottom={space.s}>
          <SystemIcon icon={calendarIcon} size={16} />
          <Text marginLeft={space.xs} fontWeight="medium">
            时间点查询结果 ({timePointQueryResult.date})
          </Text>
        </Flex>

        {timePointQueryResult.result ? (
          <Box>
            <Text fontSize={fontSizes.body.medium} fontWeight="medium" marginBottom={space.xs}>
              {timePointQueryResult.result.name}
            </Text>
            <Text fontSize={fontSizes.body.small} color={colors.licorice400}>
              在 {timePointQueryResult.date} 时间点，该组织的有效记录
            </Text>
            <Text fontSize={fontSizes.body.small} color={colors.licorice400}>
              生效期间: {timePointQueryResult.result.effective_date} - {timePointQueryResult.result.end_date || '至今'}
            </Text>
          </Box>
        ) : (
          <Text color={colors.licorice400}>
            在 {timePointQueryResult.date} 时间点没有找到该组织的记录
          </Text>
        )}
      </Card>
    );
  };

  return (
    <Box padding={space.l}>
      {/* 页面标题 */}
      <Box marginBottom={space.l}>
        <Text 
          fontSize={fontSizes.heading.large} 
          fontWeight="bold" 
          color={colors.licorice500}
          marginBottom={space.s}
        >
          时态管理演示 - GraphQL版本
        </Text>
        <Text fontSize={fontSizes.body.medium} color={colors.licorice400}>
          基于Neo4j Bitemporal数据模型和GraphQL时态查询API
        </Text>
      </Box>

      {/* 组织选择器 */}
      <Card padding={space.m} marginBottom={space.l}>
        <Text fontWeight="medium" marginBottom={space.s}>选择测试组织</Text>
        
        <Flex gap={space.m} alignItems="flex-end" flexWrap="wrap">
          <FormField>
            <FormField.Label>预设组织</FormField.Label>
            <FormField.Field>
              <Select
                value={selectedOrganizationCode}
                onChange={(e) => handleOrganizationSelect(e.target.value)}
                disabled={useCustomCode}
              >
                {DEMO_ORGANIZATION_CODES.map(({ label, value }) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </Select>
            </FormField.Field>
          </FormField>

          <Text color={colors.licorice300}>或</Text>

          <FormField>
            <FormField.Label>自定义组织代码</FormField.Label>
            <FormField.Field>
              <TextInput
                value={customCode}
                onChange={(e) => setCustomCode(e.target.value)}
                placeholder="输入组织代码，如 1000001"
              />
            </FormField.Field>
          </FormField>

          <PrimaryButton 
            onClick={handleUseCustomCode}
            disabled={!customCode}
          >
            使用自定义代码
          </PrimaryButton>

          {useCustomCode && (
            <SecondaryButton onClick={() => setUseCustomCode(false)}>
              返回预设选择
            </SecondaryButton>
          )}
        </Flex>

        <Box marginTop={space.s}>
          <Text fontSize={fontSizes.body.small} color={colors.licorice400}>
            当前查询组织: <strong>{currentCode}</strong>
          </Text>
        </Box>
      </Card>

      {/* 主要功能区域 */}
      <Tabs>
        <Tabs.List>
          <Tabs.Item>历史记录查看</Tabs.Item>
          <Tabs.Item>时间点查询</Tabs.Item>
        </Tabs.List>

        <Tabs.Panel>
          {/* 历史记录查看器 - 真实功能组件 */}
          <Box paddingTop={space.m}>
            <TemporalHistoryViewer
              organizationCode={currentCode}
              onRecordSelect={handleRecordSelect}
              onTimePointQuery={handleTimePointQuery}
              showTimePointQuery={false} // 时间点查询单独显示
              showFilters={true}
              maxHeight="500px"
            />
          </Box>
        </Tabs.Panel>

        <Tabs.Panel>
          {/* 时间点查询 - 真实功能组件 */}
          <Box paddingTop={space.m}>
            <TimePointQuery
              organizationCode={currentCode}
              onQueryResult={handleTimePointQuery}
              showQuickDates={true}
              compact={false}
            />
          </Box>
        </Tabs.Panel>
      </Tabs>

      {/* 底部信息区域 */}
      <Flex gap={space.l} marginTop={space.l}>
        <Box flex={1}>
          {renderSelectedRecordDetails()}
        </Box>
        
        <Box flex={1}>
          {renderTimePointQueryResult()}
        </Box>
      </Flex>

      {/* 功能说明 */}
      <Card padding={space.m} marginTop={space.l} backgroundColor={colors.soap200}>
        <Text fontWeight="medium" marginBottom={space.s}>功能说明</Text>
        
        <Box>
          <Text fontSize={fontSizes.body.small} color={colors.licorice400} display="block" marginBottom={space.xs}>
            • <strong>历史记录查看</strong>: 展示组织的完整时态历史，支持时间范围过滤
          </Text>
          <Text fontSize={fontSizes.body.small} color={colors.licorice400} display="block" marginBottom={space.xs}>
            • <strong>时间点查询</strong>: 查询特定时间点的组织状态，支持快速日期选择
          </Text>
          <Text fontSize={fontSizes.body.small} color={colors.licorice400} display="block" marginBottom={space.xs}>
            • <strong>实时GraphQL查询</strong>: 直接查询Neo4j时态数据，响应时间&lt;100ms
          </Text>
          <Text fontSize={fontSizes.body.small} color={colors.licorice400} display="block" marginBottom={space.xs}>
            • <strong>推荐测试</strong>: 使用1000056组织代码，包含14条完整的历史记录
          </Text>
          <Text fontSize={fontSizes.body.small} color={colors.licorice400} display="block" marginBottom={space.xs}>
            • <strong>组织层级</strong>: 显示组织在层级结构中的位置（1级=根组织，2级=一级子组织）
          </Text>
          <Text fontSize={fontSizes.body.small} color={colors.licorice400} display="block">
            • <strong>级联长路径</strong>: 完整层级路径，如"/1000000/1000001/1000056"表示根组织→一级部门→当前组织
          </Text>
        </Box>
      </Card>
    </Box>
  );
};

export default TemporalManagementGraphQL;