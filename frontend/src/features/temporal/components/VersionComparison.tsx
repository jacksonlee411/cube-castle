/**
 * 历史版本对比组件
 * 对比和展示组织架构的不同历史版本
 */
import React, { useState, useMemo, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Badge } from '@workday/canvas-kit-react/badge';
import { Tooltip } from '@workday/canvas-kit-react/tooltip';
import { Select } from '@workday/canvas-kit-react/select';
import { Tabs } from '@workday/canvas-kit-react/tabs';
import { 
  colors, 
  space, 
  borderRadius,
  fontSizes 
} from '@workday/canvas-kit-react/tokens';
import {
  ArrowLeftIcon,
  ArrowRightIcon,
  CompareIcon,
  HistoryIcon,
  InfoIcon
} from '@workday/canvas-kit-react/icon';
import { useOrganizationHistory } from '../../shared/hooks/useTemporalQuery';
import type { 
  TemporalOrganizationUnit,
  TemporalQueryParams
} from '../../shared/types/temporal';

export interface VersionComparisonProps {
  /** 组织代码 */
  organizationCode: string;
  /** 查询参数 */
  queryParams?: Partial<TemporalQueryParams>;
  /** 预设的版本1 */
  version1?: TemporalOrganizationUnit;
  /** 预设的版本2 */
  version2?: TemporalOrganizationUnit;
  /** 默认选中的版本索引 */
  defaultVersions?: [number, number];
  /** 是否紧凑模式 */
  compact?: boolean;
  /** 是否显示元数据 */
  showMetadata?: boolean;
  /** 版本选择回调 */
  onVersionSelect?: (leftVersion: TemporalOrganizationUnit, rightVersion: TemporalOrganizationUnit) => void;
}

/**
 * 字段差异项组件
 */
interface FieldDiffProps {
  field: string;
  fieldLabel: string;
  leftValue: unknown;
  rightValue: unknown;
  compact: boolean;
}

const FieldDiff: React.FC<FieldDiffProps> = ({
  field,
  fieldLabel,
  leftValue,
  rightValue,
  compact
}) => {
  const hasChange = leftValue !== rightValue;
  
  const formatValue = (value: unknown) => {
    if (value === null || value === undefined) {
      return <Text color={colors.licorice400} fontStyle="italic">空</Text>;
    }
    if (typeof value === 'boolean') {
      return value ? '是' : '否';
    }
    if (typeof value === 'object') {
      return JSON.stringify(value, null, 2);
    }
    return String(value);
  };

  const getChangeType = () => {
    if (leftValue === null || leftValue === undefined) {
      return 'added'; // 新增
    }
    if (rightValue === null || rightValue === undefined) {
      return 'removed'; // 删除
    }
    return 'modified'; // 修改
  };

  const changeType = hasChange ? getChangeType() : null;
  
  const getChangeStyles = (type: string | null) => {
    switch (type) {
      case 'added':
        return { 
          bgColor: colors.greenFresca100, 
          borderColor: colors.greenFresca300,
          textColor: colors.greenFresca700
        };
      case 'removed':
        return { 
          bgColor: colors.cinnamon100, 
          borderColor: colors.cinnamon300,
          textColor: colors.cinnamon700
        };
      case 'modified':
        return { 
          bgColor: colors.cantaloupe100, 
          borderColor: colors.cantaloupe300,
          textColor: colors.cantaloupe700
        };
      default:
        return { 
          bgColor: 'transparent', 
          borderColor: colors.soap300,
          textColor: colors.licorice600
        };
    }
  };

  const changeStyles = getChangeStyles(changeType);

  return (
    <Box
      padding={compact ? space.s : space.m}
      backgroundColor={changeStyles.bgColor}
      border={`1px solid ${changeStyles.borderColor}`}
      borderRadius={borderRadius.s}
      marginBottom={space.s}
    >
      <Flex alignItems="center" justifyContent="space-between" marginBottom={space.xs}>
        <Text 
          fontWeight="medium" 
          fontSize={compact ? 'small' : 'medium'}
          color={changeStyles.textColor}
        >
          {fieldLabel}
        </Text>
        {hasChange && (
          <Badge 
            color={changeStyles.textColor}
            variant="outline"
            size="small"
          >
            {changeType === 'added' ? '新增' : 
             changeType === 'removed' ? '删除' : '修改'}
          </Badge>
        )}
      </Flex>

      <Flex gap={space.m}>
        {/* 左侧版本值 */}
        <Box flex="1">
          <Text fontSize="small" color={colors.licorice500} marginBottom={space.xxxs}>
            旧版本
          </Text>
          <Box
            padding={space.s}
            backgroundColor={colors.soap100}
            borderRadius={borderRadius.s}
            fontSize={compact ? fontSizes.small : fontSizes.medium}
          >
            {formatValue(leftValue)}
          </Box>
        </Box>

        {/* 箭头指示 */}
        {hasChange && (
          <Flex alignItems="center" justifyContent="center" paddingTop={space.l}>
            <ArrowRightIcon size="small" color={changeStyles.textColor} />
          </Flex>
        )}

        {/* 右侧版本值 */}
        <Box flex="1">
          <Text fontSize="small" color={colors.licorice500} marginBottom={space.xxxs}>
            新版本
          </Text>
          <Box
            padding={space.s}
            backgroundColor={hasChange ? colors.soap100 : colors.soap50}
            borderRadius={borderRadius.s}
            fontSize={compact ? fontSizes.small : fontSizes.medium}
          >
            {formatValue(rightValue)}
          </Box>
        </Box>
      </Flex>
    </Box>
  );
};

/**
 * 版本信息卡片组件
 */
interface VersionCardProps {
  version: TemporalOrganizationUnit;
  title: string;
  color: string;
  compact: boolean;
}

const VersionCard: React.FC<VersionCardProps> = ({
  version,
  title,
  color,
  compact
}) => {
  const formatDate = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleString('zh-CN');
    } catch {
      return dateStr;
    }
  };

  return (
    <Card 
      padding={compact ? space.s : space.m}
      border={`2px solid ${color}`}
    >
      <Flex alignItems="center" gap={space.s} marginBottom={space.s}>
        <Text fontWeight="bold" color={color}>
          {title}
        </Text>
        {version.version && (
          <Badge color={color} variant="outline">
            v{version.version}
          </Badge>
        )}
      </Flex>

      <Box marginBottom={space.s}>
        <Text fontWeight="medium" fontSize={compact ? 'small' : 'medium'}>
          {version.name}
        </Text>
        <Text fontSize="small" color={colors.licorice500}>
          {version.code}
        </Text>
      </Box>

      <Flex gap={space.m} fontSize="small" color={colors.licorice500}>
        {version.effective_from && (
          <Text>
            生效: {formatDate(version.effective_from)}
          </Text>
        )}
        {version.effective_to && (
          <Text>
            失效: {formatDate(version.effective_to)}
          </Text>
        )}
      </Flex>

      {version.change_reason && !compact && (
        <Box marginTop={space.s}>
          <Text fontSize="small" color={colors.licorice600}>
            变更原因: {version.change_reason}
          </Text>
        </Box>
      )}
    </Card>
  );
};

/**
 * 历史版本对比组件
 */
export const VersionComparison: React.FC<VersionComparisonProps> = ({
  organizationCode,
  queryParams,
  defaultVersions = [0, 1],
  compact = false,
  onVersionSelect
}) => {
  const [selectedVersions, setSelectedVersions] = useState<[number, number]>(defaultVersions);
  const [activeTab, setActiveTab] = useState('diff');

  // 获取历史版本数据
  const {
    data: versions = [],
    isLoading,
    isError,
    error,
    hasHistory
  } = useOrganizationHistory(organizationCode, queryParams);

  // 当前选中的两个版本
  const [leftVersion, rightVersion] = useMemo(() => {
    if (versions.length < 2) return [null, null];
    return [
      versions[selectedVersions[0]] || null,
      versions[selectedVersions[1]] || null
    ];
  }, [versions, selectedVersions]);

  // 计算字段差异
  const fieldDiffs = useMemo(() => {
    if (!leftVersion || !rightVersion) return [];

    const fieldsToCompare = [
      { key: 'name', label: '名称' },
      { key: 'unit_type', label: '组织类型' },
      { key: 'status', label: '状态' },
      { key: 'level', label: '层级' },
      { key: 'parent_code', label: '上级组织' },
      { key: 'sort_order', label: '排序' },
      { key: 'description', label: '描述' },
      { key: 'effective_from', label: '生效时间' },
      { key: 'effective_to', label: '失效时间' },
      { key: 'change_reason', label: '变更原因' }
    ];

    return fieldsToCompare.map(field => ({
      ...field,
      leftValue: (leftVersion as any)[field.key],
      rightValue: (rightVersion as any)[field.key],
      hasChange: (leftVersion as any)[field.key] !== (rightVersion as any)[field.key]
    }));
  }, [leftVersion, rightVersion]);

  // 统计差异数量
  const diffStats = useMemo(() => {
    const totalFields = fieldDiffs.length;
    const changedFields = fieldDiffs.filter(diff => diff.hasChange).length;
    const unchangedFields = totalFields - changedFields;

    return { totalFields, changedFields, unchangedFields };
  }, [fieldDiffs]);

  // 处理版本选择
  const handleVersionChange = useCallback((position: 'left' | 'right', versionIndex: number) => {
    const newVersions: [number, number] = position === 'left' 
      ? [versionIndex, selectedVersions[1]]
      : [selectedVersions[0], versionIndex];

    setSelectedVersions(newVersions);

    if (versions[newVersions[0]] && versions[newVersions[1]]) {
      onVersionSelect?.(versions[newVersions[0]], versions[newVersions[1]]);
    }
  }, [selectedVersions, versions, onVersionSelect]);

  if (isLoading) {
    return (
      <Card padding={space.m}>
        <Text>🔄 加载版本历史数据...</Text>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card padding={space.m}>
        <Text color={colors.cinnamon600}>
          ❌ 加载版本历史失败: {error?.message || '未知错误'}
        </Text>
      </Card>
    );
  }

  if (!hasHistory || versions.length < 2) {
    return (
      <Card padding={space.m}>
        <Flex justifyContent="center" alignItems="center" flexDirection="column" gap={space.s}>
          <HistoryIcon size="large" color={colors.licorice400} />
          <Text color={colors.licorice500}>
            {versions.length === 0 ? '📭 暂无历史版本' : '📄 仅有一个版本，无法对比'}
          </Text>
        </Flex>
      </Card>
    );
  }

  return (
    <Box>
      {/* 版本选择器 */}
      <Box marginBottom={space.m}>
        <Flex alignItems="center" gap={space.m} marginBottom={space.s}>
          <Text fontSize="large" fontWeight="medium">
            <CompareIcon /> 版本对比
          </Text>
          <Badge color={colors.blueberry600} variant="outline">
            {versions.length} 个版本
          </Badge>
        </Flex>

        <Flex gap={space.m} alignItems="center">
          {/* 左侧版本选择 */}
          <Box flex="1">
            <Text fontSize="small" marginBottom={space.xs}>基准版本 (旧)</Text>
            <Select
              value={selectedVersions[0].toString()}
              onChange={(value) => handleVersionChange('left', parseInt(value))}
            >
              {versions.map((version, index) => (
                <MenuItem key={index} value={index.toString()}>
                  v{version.version || index + 1} - {version.name} 
                  {version.effective_from && ` (${new Date(version.effective_from).toLocaleDateString()})`}
                </MenuItem>
              ))}
            </Select>
          </Box>

          <ArrowRightIcon color={colors.licorice400} />

          {/* 右侧版本选择 */}
          <Box flex="1">
            <Text fontSize="small" marginBottom={space.xs}>对比版本 (新)</Text>
            <Select
              value={selectedVersions[1].toString()}
              onChange={(value) => handleVersionChange('right', parseInt(value))}
            >
              {versions.map((version, index) => (
                <MenuItem key={index} value={index.toString()}>
                  v{version.version || index + 1} - {version.name}
                  {version.effective_from && ` (${new Date(version.effective_from).toLocaleDateString()})`}
                </MenuItem>
              ))}
            </Select>
          </Box>
        </Flex>
      </Box>

      {/* 对比统计 */}
      <Box marginBottom={space.m}>
        <Flex gap={space.s} alignItems="center">
          <Badge color={colors.cantaloupe600} variant="solid">
            {diffStats.changedFields} 个差异
          </Badge>
          <Badge color={colors.greenFresca600} variant="outline">
            {diffStats.unchangedFields} 个相同
          </Badge>
          <Tooltip title="总共对比的字段数量">
            <Badge color={colors.licorice400} variant="outline">
              <InfoIcon size="small" /> {diffStats.totalFields} 字段
            </Badge>
          </Tooltip>
        </Flex>
      </Box>

      {/* 对比内容标签页 */}
      <Tabs activeKey={activeTab} onSelectionChange={setActiveTab}>
        <TabsList>
          <Tab value="diff">字段差异</Tab>
          <Tab value="cards">版本卡片</Tab>
          <Tab value="raw">原始数据</Tab>
        </TabsList>

        {/* 字段差异视图 */}
        {activeTab === 'diff' && leftVersion && rightVersion && (
          <Box marginTop={space.m}>
            {fieldDiffs.length === 0 ? (
              <Text>无可对比字段</Text>
            ) : (
              <Box>
                {/* 仅显示有差异的字段 */}
                <Box marginBottom={space.m}>
                  <Text fontSize="medium" fontWeight="medium" marginBottom={space.s}>
                    变更字段 ({diffStats.changedFields})
                  </Text>
                  {fieldDiffs.filter(diff => diff.hasChange).map(diff => (
                    <FieldDiff
                      key={diff.key}
                      field={diff.key}
                      fieldLabel={diff.label}
                      leftValue={diff.leftValue}
                      rightValue={diff.rightValue}
                      compact={compact}
                    />
                  ))}
                </Box>

                {/* 相同字段（可选显示） */}
                {!compact && (
                  <Box>
                    <Text fontSize="medium" fontWeight="medium" marginBottom={space.s}>
                      相同字段 ({diffStats.unchangedFields})
                    </Text>
                    {fieldDiffs.filter(diff => !diff.hasChange).map(diff => (
                      <FieldDiff
                        key={diff.key}
                        field={diff.key}
                        fieldLabel={diff.label}
                        leftValue={diff.leftValue}
                        rightValue={diff.rightValue}
                        compact={compact}
                      />
                    ))}
                  </Box>
                )}
              </Box>
            )}
          </Box>
        )}

        {/* 版本卡片视图 */}
        {activeTab === 'cards' && leftVersion && rightVersion && (
          <Box marginTop={space.m}>
            <Flex gap={space.m}>
              <Box flex="1">
                <VersionCard
                  version={leftVersion}
                  title="基准版本"
                  color={colors.blueberry600}
                  compact={compact}
                />
              </Box>
              <Box flex="1">
                <VersionCard
                  version={rightVersion}
                  title="对比版本"
                  color={colors.peach600}
                  compact={compact}
                />
              </Box>
            </Flex>
          </Box>
        )}

        {/* 原始数据视图 */}
        {activeTab === 'raw' && leftVersion && rightVersion && (
          <Box marginTop={space.m}>
            <Flex gap={space.m}>
              <Box flex="1">
                <Text fontSize="small" fontWeight="medium" marginBottom={space.s}>
                  基准版本 JSON
                </Text>
                <Box
                  as="pre"
                  padding={space.s}
                  backgroundColor={colors.soap100}
                  borderRadius={borderRadius.s}
                  fontSize="small"
                  overflow="auto"
                  maxHeight="400px"
                >
                  {JSON.stringify(leftVersion, null, 2)}
                </Box>
              </Box>
              <Box flex="1">
                <Text fontSize="small" fontWeight="medium" marginBottom={space.s}>
                  对比版本 JSON
                </Text>
                <Box
                  as="pre"
                  padding={space.s}
                  backgroundColor={colors.soap100}
                  borderRadius={borderRadius.s}
                  fontSize="small"
                  overflow="auto"
                  maxHeight="400px"
                >
                  {JSON.stringify(rightVersion, null, 2)}
                </Box>
              </Box>
            </Flex>
          </Box>
        )}
      </Tabs>
    </Box>
  );
};

export default VersionComparison;