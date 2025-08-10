/**
 * 版本对比功能测试应用
 * 验证VersionComparison组件与后端API的集成功能
 */
import React, { useState, useCallback } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Text } from '@workday/canvas-kit-react/text';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { Card } from '@workday/canvas-kit-react/card';
import { Badge } from '@workday/canvas-kit-react/badge';
import { LoadingSpinner } from '@workday/canvas-kit-react/loading-animation';

import { useOrganizationHistory } from './shared/hooks/useTemporalQuery';
import type { TemporalOrganizationUnit } from './shared/types/temporal';

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
 * 简化的版本对比组件
 */
interface SimpleVersionComparisonProps {
  organizationCode: string;
}

const SimpleVersionComparison: React.FC<SimpleVersionComparisonProps> = ({ organizationCode }) => {
  const [selectedVersions, setSelectedVersions] = useState<[number, number]>([0, 1]);
  
  // 获取历史版本数据
  const {
    data: versions = [],
    isLoading,
    isError,
    error,
    hasHistory
  } = useOrganizationHistory(organizationCode, { limit: 20 });

  // 当前选中的两个版本
  const leftVersion = versions[selectedVersions[0]];
  const rightVersion = versions[selectedVersions[1]];

  // 字段对比
  const fieldsToCompare = [
    { key: 'name', label: '名称' },
    { key: 'unitType', label: '组织类型' },
    { key: 'status', label: '状态' },
    { key: 'level', label: '层级' },
    { key: 'parentCode', label: '上级组织' },
    { key: 'sortOrder', label: '排序' },
    { key: 'description', label: '描述' },
    { key: 'effectiveFrom', label: '生效时间' },
    { key: 'effectiveTo', label: '失效时间' },
    { key: 'changeReason', label: '变更原因' }
  ];

  const formatValue = (value: any) => {
    if (value === null || value === undefined || value === '') {
      return '(空)';
    }
    if (typeof value === 'boolean') {
      return value ? '是' : '否';
    }
    if (value instanceof Date) {
      return value.toLocaleDateString('zh-CN');
    }
    if (typeof value === 'string' && value.includes('T')) {
      // 可能是ISO日期字符串
      try {
        return new Date(value).toLocaleString('zh-CN');
      } catch {
        return String(value);
      }
    }
    return String(value);
  };

  const getDifferences = () => {
    if (!leftVersion || !rightVersion) return [];
    
    return fieldsToCompare.map(field => {
      const leftVal = (leftVersion as any)[field.key];
      const rightVal = (rightVersion as any)[field.key];
      const hasChange = leftVal !== rightVal;
      
      return {
        ...field,
        leftValue: leftVal,
        rightValue: rightVal,
        hasChange
      };
    });
  };

  const differences = getDifferences();
  const changeCount = differences.filter(d => d.hasChange).length;

  if (isLoading) {
    return (
      <Card padding="m">
        <Flex alignItems="center" gap="s">
          <LoadingSpinner size="s" />
          <Text>加载历史版本数据...</Text>
        </Flex>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card padding="m">
        <Text color="cinnamon600">
          ❌ 加载版本历史失败: {error?.message || '未知错误'}
        </Text>
      </Card>
    );
  }

  if (!hasHistory || versions.length < 2) {
    return (
      <Card padding="m">
        <Text color="hint" textAlign="center">
          {versions.length === 0 ? '📭 暂无历史版本' : '📄 仅有一个版本，无法对比'}
        </Text>
      </Card>
    );
  }

  return (
    <Box>
      {/* 版本选择 */}
      <Card marginBottom="m" padding="m">
        <Text as="h3" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
          📊 版本对比设置
        </Text>
        
        <Flex gap="m" alignItems="flex-end" marginBottom="m">
          <FormField flex="1">
            <FormField.Label>基准版本 (旧)</FormField.Label>
            <FormField.Field>
              <select
                value={selectedVersions[0]}
                onChange={(e) => setSelectedVersions([parseInt(e.target.value), selectedVersions[1]])}
                style={{ 
                  width: '100%', 
                  padding: '8px', 
                  borderRadius: '4px', 
                  border: '1px solid #ddd' 
                }}
              >
                {versions.map((version, index) => (
                  <option key={index} value={index}>
                    v{version.version || index + 1} - {version.name}
                    {version.effectiveFrom && ` (${new Date(version.effectiveFrom).toLocaleDateString()})`}
                  </option>
                ))}
              </select>
            </FormField.Field>
          </FormField>

          <Text typeLevel="subtext.medium" paddingBottom="m">→</Text>

          <FormField flex="1">
            <FormField.Label>对比版本 (新)</FormField.Label>
            <FormField.Field>
              <select
                value={selectedVersions[1]}
                onChange={(e) => setSelectedVersions([selectedVersions[0], parseInt(e.target.value)])}
                style={{ 
                  width: '100%', 
                  padding: '8px', 
                  borderRadius: '4px', 
                  border: '1px solid #ddd' 
                }}
              >
                {versions.map((version, index) => (
                  <option key={index} value={index}>
                    v{version.version || index + 1} - {version.name}
                    {version.effectiveFrom && ` (${new Date(version.effectiveFrom).toLocaleDateString()})`}
                  </option>
                ))}
              </select>
            </FormField.Field>
          </FormField>
        </Flex>

        <Flex gap="s" alignItems="center">
          <Badge color={changeCount > 0 ? "cantaloupe600" : "greenFresca600"}>
            {changeCount} 个差异
          </Badge>
          <Badge color="licorice400" variant="outline">
            {differences.length - changeCount} 个相同
          </Badge>
          <Badge color="blueberry600" variant="outline">
            共 {versions.length} 个版本
          </Badge>
        </Flex>
      </Card>

      {/* 版本信息卡片 */}
      <Flex gap="m" marginBottom="m">
        <Card flex="1" padding="m" style={{ border: '2px solid #1f77b4' }}>
          <Text as="h4" typeLevel="subtext.medium" fontWeight="bold" marginBottom="s" color="blueberry600">
            基准版本 (旧)
          </Text>
          {leftVersion && (
            <Box>
              <Text typeLevel="body.medium" marginBottom="xs">
                {leftVersion.name}
              </Text>
              <Text typeLevel="subtext.small" color="hint" marginBottom="xs">
                编码: {leftVersion.code}
              </Text>
              {leftVersion.effectiveFrom && (
                <Text typeLevel="subtext.small" color="hint">
                  生效时间: {formatValue(leftVersion.effectiveFrom)}
                </Text>
              )}
            </Box>
          )}
        </Card>

        <Card flex="1" padding="m" style={{ border: '2px solid #ff7f0e' }}>
          <Text as="h4" typeLevel="subtext.medium" fontWeight="bold" marginBottom="s" color="peach600">
            对比版本 (新)
          </Text>
          {rightVersion && (
            <Box>
              <Text typeLevel="body.medium" marginBottom="xs">
                {rightVersion.name}
              </Text>
              <Text typeLevel="subtext.small" color="hint" marginBottom="xs">
                编码: {rightVersion.code}
              </Text>
              {rightVersion.effectiveFrom && (
                <Text typeLevel="subtext.small" color="hint">
                  生效时间: {formatValue(rightVersion.effectiveFrom)}
                </Text>
              )}
            </Box>
          )}
        </Card>
      </Flex>

      {/* 差异对比 */}
      <Card padding="m">
        <Text as="h3" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
          🔍 字段差异对比
        </Text>
        
        {/* 变更字段 */}
        {changeCount > 0 && (
          <Box marginBottom="m">
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s" color="cantaloupe600">
              变更字段 ({changeCount})
            </Text>
            {differences.filter(diff => diff.hasChange).map(diff => (
              <Box
                key={diff.key}
                padding="s"
                marginBottom="s"
                style={{
                  backgroundColor: '#fff3cd',
                  border: '1px solid #ffeaa7',
                  borderRadius: '4px'
                }}
              >
                <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="xs">
                  {diff.label}
                </Text>
                <Flex gap="m">
                  <Box flex="1">
                    <Text typeLevel="subtext.small" color="hint" marginBottom="xxxs">旧值</Text>
                    <Box
                      padding="xs"
                      style={{
                        backgroundColor: '#f8f9fa',
                        borderRadius: '4px',
                        border: '1px solid #dee2e6'
                      }}
                    >
                      <Text typeLevel="body.small">{formatValue(diff.leftValue)}</Text>
                    </Box>
                  </Box>
                  <Text paddingTop="m">→</Text>
                  <Box flex="1">
                    <Text typeLevel="subtext.small" color="hint" marginBottom="xxxs">新值</Text>
                    <Box
                      padding="xs"
                      style={{
                        backgroundColor: '#f8f9fa',
                        borderRadius: '4px',
                        border: '1px solid #dee2e6'
                      }}
                    >
                      <Text typeLevel="body.small">{formatValue(diff.rightValue)}</Text>
                    </Box>
                  </Box>
                </Flex>
              </Box>
            ))}
          </Box>
        )}

        {/* 相同字段 */}
        <Box>
          <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s" color="greenFresca600">
            相同字段 ({differences.length - changeCount})
          </Text>
          {differences.filter(diff => !diff.hasChange).slice(0, 3).map(diff => (
            <Box
              key={diff.key}
              padding="s"
              marginBottom="s"
              style={{
                backgroundColor: '#f0f7ff',
                border: '1px solid #b6d7ff',
                borderRadius: '4px'
              }}
            >
              <Flex justifyContent="space-between" alignItems="center">
                <Text typeLevel="subtext.medium">{diff.label}</Text>
                <Text typeLevel="body.small">{formatValue(diff.leftValue)}</Text>
              </Flex>
            </Box>
          ))}
          {differences.length - changeCount > 3 && (
            <Text typeLevel="subtext.small" color="hint">
              ... 还有 {differences.length - changeCount - 3} 个相同字段
            </Text>
          )}
        </Box>
      </Card>
    </Box>
  );
};

/**
 * 版本对比测试应用
 */
const VersionComparisonTestApp: React.FC = () => {
  const [organizationCode, setOrganizationCode] = useState('1000001');
  const [testMode, setTestMode] = useState<'simple' | 'advanced'>('simple');

  return (
    <QueryClientProvider client={queryClient}>
      <Box padding="l">
        <Text as="h1" typeLevel="heading.large" marginBottom="l">
          🔀 版本对比功能测试
        </Text>
        
        <Text typeLevel="body.medium" marginBottom="m">
          测试VersionComparison组件与后端API的数据连接功能，验证历史版本对比和差异展示。
        </Text>

        {/* 控制面板 */}
        <Card marginBottom="l" padding="m">
          <Text as="h2" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
            🎛️ 测试控制面板
          </Text>
          
          <Flex gap="m" alignItems="flex-end" marginBottom="m">
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

            <Box>
              <PrimaryButton
                onClick={() => setTestMode(testMode === 'simple' ? 'advanced' : 'simple')}
              >
                {testMode === 'simple' ? '切换到高级模式' : '切换到简单模式'}
              </PrimaryButton>
            </Box>
          </Flex>
        </Card>

        {/* 功能测试要点 */}
        <Card marginBottom="l" padding="m" style={{ backgroundColor: '#f0f7ff', border: '1px solid #d1ecf1' }}>
          <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
            📋 版本对比功能验证要点
          </Text>
          <ul style={{ marginLeft: '20px', lineHeight: '1.6' }}>
            <li>✅ 历史版本数据API调用和响应处理</li>
            <li>✅ 版本选择和动态对比功能</li>
            <li>✅ 字段差异检测和高亮显示</li>
            <li>✅ 数据格式化和用户友好显示</li>
            <li>✅ 错误处理和状态反馈</li>
            <li>✅ 响应式布局和交互体验</li>
            <li>✅ 版本信息卡片展示</li>
            <li>✅ 实时数据更新和缓存</li>
          </ul>
        </Card>

        {/* 版本对比组件 */}
        <SimpleVersionComparison organizationCode={organizationCode} />
      </Box>
    </QueryClientProvider>
  );
};

export default VersionComparisonTestApp;