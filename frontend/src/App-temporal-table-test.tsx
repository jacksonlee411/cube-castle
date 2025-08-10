/**
 * 时态表格功能测试应用
 * 验证TemporalTable组件的完整功能和时态感知能力
 */
import React, { useState, useCallback } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Text } from '@workday/canvas-kit-react/text';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { Card } from '@workday/canvas-kit-react/card';
import { Checkbox } from '@workday/canvas-kit-react/checkbox';
import { Badge } from '@workday/canvas-kit-react/badge';

import { TemporalTable } from './features/temporal/components/TemporalTable';
import { TemporalNavbar } from './features/temporal/components/TemporalNavbar';
import { useTemporalMode } from './shared/hooks/useTemporalQuery';
import type { OrganizationUnit } from './shared/types/organization';
import type { TemporalMode } from './shared/types/temporal';

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
 * 时态表格测试组件
 */
const TemporalTableTest: React.FC = () => {
  // 时态模式状态
  const { mode: temporalMode, isHistorical, isCurrent, isPlanning } = useTemporalMode();

  // 表格配置状态
  const [tableConfig, setTableConfig] = useState({
    showTemporalIndicators: true,
    showActions: true,
    showSelection: false,
    compact: false,
    pageSize: 20
  });

  // 查询参数
  const [queryParams, setQueryParams] = useState({
    searchText: '',
    unit_type: '',
    status: '',
    page: 1,
    pageSize: 20
  });

  // 选中的组织
  const [selectedOrganizations, setSelectedOrganizations] = useState<OrganizationUnit[]>([]);

  // 时态模式变更处理
  const handleTemporalModeChange = useCallback((newMode: TemporalMode) => {
    console.log(`时态模式切换到: ${newMode}`);
  }, []);

  // 表格配置更新
  const updateTableConfig = useCallback((key: keyof typeof tableConfig, value: any) => {
    setTableConfig(prev => ({ ...prev, [key]: value }));
  }, []);

  // 查询参数更新
  const updateQueryParams = useCallback((key: keyof typeof queryParams, value: any) => {
    setQueryParams(prev => ({ ...prev, [key]: value }));
  }, []);

  // 表格事件处理
  const handleRowClick = useCallback((organization: OrganizationUnit) => {
    alert(`点击组织: ${organization.name} (${organization.code})`);
  }, []);

  const handleEdit = useCallback((organization: OrganizationUnit) => {
    alert(`编辑组织: ${organization.name}`);
  }, []);

  const handleDelete = useCallback((organization: OrganizationUnit) => {
    if (confirm(`确定删除组织 "${organization.name}" 吗？`)) {
      alert(`删除组织: ${organization.name}`);
    }
  }, []);

  const handleViewHistory = useCallback((organization: OrganizationUnit) => {
    alert(`查看 ${organization.name} 的历史版本`);
  }, []);

  const handleViewTimeline = useCallback((organization: OrganizationUnit) => {
    alert(`查看 ${organization.name} 的时间线`);
  }, []);

  const handleSelectionChange = useCallback((selected: OrganizationUnit[]) => {
    setSelectedOrganizations(selected);
    console.log('选中的组织:', selected.map(org => org.name));
  }, []);

  // 清空搜索
  const handleClearSearch = useCallback(() => {
    setQueryParams({
      searchText: '',
      unit_type: '',
      status: '',
      page: 1,
      pageSize: 20
    });
  }, []);

  return (
    <Box padding="l">
      <Text as="h1" typeLevel="heading.large" marginBottom="l">
        📊 时态表格功能测试
      </Text>
      
      <Text typeLevel="body.medium" marginBottom="m">
        测试TemporalTable组件的完整功能，包括时态感知、数据展示、操作处理和用户交互。
      </Text>

      {/* 时态导航栏 */}
      <Box marginBottom="l">
        <TemporalNavbar
          onModeChange={handleTemporalModeChange}
          showAdvancedSettings={true}
        />
      </Box>

      {/* 控制面板 */}
      <Card marginBottom="l" padding="m">
        <Text as="h2" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
          🎛️ 测试控制面板
        </Text>
        
        {/* 时态状态显示 */}
        <Box marginBottom="m">
          <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
            当前时态模式
          </Text>
          <Flex gap="s" alignItems="center">
            <Badge color={isCurrent ? "greenFresca600" : "licorice400"}>
              {isCurrent ? "✅ 当前模式" : "⭕ 当前模式"}
            </Badge>
            <Badge color={isHistorical ? "blueberry600" : "licorice400"}>
              {isHistorical ? "✅ 历史模式" : "⭕ 历史模式"}
            </Badge>
            <Badge color={isPlanning ? "peach600" : "licorice400"}>
              {isPlanning ? "✅ 规划模式" : "⭕ 规划模式"}
            </Badge>
          </Flex>
        </Box>

        {/* 表格配置 */}
        <Box marginBottom="m">
          <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
            表格配置
          </Text>
          <Flex gap="m" flexWrap="wrap">
            <Checkbox
              checked={tableConfig.showTemporalIndicators}
              onChange={(e) => updateTableConfig('showTemporalIndicators', e.target.checked)}
            >
              显示时态指示器
            </Checkbox>
            <Checkbox
              checked={tableConfig.showActions}
              onChange={(e) => updateTableConfig('showActions', e.target.checked)}
            >
              显示操作按钮
            </Checkbox>
            <Checkbox
              checked={tableConfig.showSelection}
              onChange={(e) => updateTableConfig('showSelection', e.target.checked)}
            >
              显示选择列
            </Checkbox>
            <Checkbox
              checked={tableConfig.compact}
              onChange={(e) => updateTableConfig('compact', e.target.checked)}
            >
              紧凑模式
            </Checkbox>
          </Flex>
        </Box>

        {/* 查询参数 */}
        <Box marginBottom="m">
          <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
            查询筛选
          </Text>
          <Flex gap="m" flexWrap="wrap" alignItems="flex-end">
            <FormField flex="1" minWidth="200px">
              <FormField.Label>搜索关键词</FormField.Label>
              <FormField.Field>
                <FormField.Input
                  as={TextInput}
                  value={queryParams.searchText}
                  onChange={(e) => updateQueryParams('searchText', e.target.value)}
                  placeholder="输入组织名称或编码"
                />
              </FormField.Field>
            </FormField>

            <FormField flex="1" minWidth="150px">
              <FormField.Label>组织类型</FormField.Label>
              <FormField.Field>
                <select
                  value={queryParams.unit_type}
                  onChange={(e) => updateQueryParams('unit_type', e.target.value)}
                  style={{ 
                    width: '100%', 
                    padding: '8px', 
                    borderRadius: '4px', 
                    border: '1px solid #ddd' 
                  }}
                >
                  <option value="">全部类型</option>
                  <option value="COMPANY">公司</option>
                  <option value="DEPARTMENT">部门</option>
                  <option value="COST_CENTER">成本中心</option>
                  <option value="PROJECT_TEAM">项目组</option>
                </select>
              </FormField.Field>
            </FormField>

            <FormField flex="1" minWidth="120px">
              <FormField.Label>状态</FormField.Label>
              <FormField.Field>
                <select
                  value={queryParams.status}
                  onChange={(e) => updateQueryParams('status', e.target.value)}
                  style={{ 
                    width: '100%', 
                    padding: '8px', 
                    borderRadius: '4px', 
                    border: '1px solid #ddd' 
                  }}
                >
                  <option value="">全部状态</option>
                  <option value="ACTIVE">启用</option>
                  <option value="INACTIVE">停用</option>
                  <option value="PLANNED">规划</option>
                </select>
              </FormField.Field>
            </FormField>

            <Box>
              <SecondaryButton onClick={handleClearSearch}>
                🗑️ 清空筛选
              </SecondaryButton>
            </Box>
          </Flex>
        </Box>

        {/* 选择状态 */}
        {tableConfig.showSelection && selectedOrganizations.length > 0 && (
          <Box marginBottom="s">
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
              选择统计
            </Text>
            <Flex gap="s" alignItems="center">
              <Badge color="blueberry600">
                已选择 {selectedOrganizations.length} 个组织
              </Badge>
              <SecondaryButton size="small" onClick={() => setSelectedOrganizations([])}>
                清空选择
              </SecondaryButton>
            </Flex>
            <Box marginTop="s">
              <Text typeLevel="subtext.small" color="hint">
                选中组织: {selectedOrganizations.map(org => org.name).join(', ')}
              </Text>
            </Box>
          </Box>
        )}
      </Card>

      {/* 功能验证要点 */}
      <Card marginBottom="l" padding="m" style={{ backgroundColor: '#f0f7ff', border: '1px solid #d1ecf1' }}>
        <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
          📋 时态表格功能验证要点
        </Text>
        <ul style={{ marginLeft: '20px', lineHeight: '1.6' }}>
          <li>✅ 时态模式感知和数据切换</li>
          <li>✅ 时态指示器和状态标识</li>
          <li>✅ 动态列显示 (时态字段仅在历史/规划模式显示)</li>
          <li>✅ 操作按钮状态管理 (历史模式禁用编辑/删除)</li>
          <li>✅ 行选择和批量操作</li>
          <li>✅ 数据格式化和用户友好显示</li>
          <li>✅ 响应式布局和交互体验</li>
          <li>✅ 搜索筛选和查询参数传递</li>
          <li>✅ 分页和数据加载状态</li>
          <li>✅ 错误处理和用户反馈</li>
        </ul>
      </Card>

      {/* 时态表格 */}
      <TemporalTable
        queryParams={queryParams}
        showTemporalIndicators={tableConfig.showTemporalIndicators}
        showActions={tableConfig.showActions}
        showSelection={tableConfig.showSelection}
        compact={tableConfig.compact}
        pageSize={tableConfig.pageSize}
        onRowClick={handleRowClick}
        onEdit={isHistorical ? undefined : handleEdit}
        onDelete={isHistorical ? undefined : handleDelete}
        onViewHistory={handleViewHistory}
        onViewTimeline={handleViewTimeline}
        onSelectionChange={handleSelectionChange}
      />

      {/* 测试提示 */}
      <Box marginTop="l">
        <Card padding="m" style={{ backgroundColor: '#fff3cd', border: '1px solid #ffeaa7' }}>
          <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
            💡 测试提示
          </Text>
          <ul style={{ marginLeft: '20px', lineHeight: '1.6' }}>
            <li><strong>时态模式切换</strong>: 使用上方时态导航栏切换不同模式，观察表格数据和功能变化</li>
            <li><strong>历史模式</strong>: 在历史模式下，编辑和删除按钮会被禁用</li>
            <li><strong>规划模式</strong>: 会显示规划中的组织和未来生效时间</li>
            <li><strong>时态字段</strong>: 生效时间和失效时间列仅在历史/规划模式下显示</li>
            <li><strong>选择功能</strong>: 启用选择列可以进行批量操作测试</li>
            <li><strong>筛选功能</strong>: 测试搜索、类型和状态筛选的数据更新</li>
          </ul>
        </Card>
      </Box>
    </Box>
  );
};

/**
 * 时态表格测试应用
 */
export const TemporalTableTestApp: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <TemporalTableTest />
    </QueryClientProvider>
  );
};

export default TemporalTableTestApp;