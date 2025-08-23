import React, { useState, useCallback, useMemo } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Card } from '@workday/canvas-kit-react/card';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { SecondaryButton } from '@workday/canvas-kit-react/button';
import { Switch } from '@workday/canvas-kit-react/switch';
import { useDebounce } from '../../shared/hooks/useDebounce';
import { TemporalStatusSelector } from '../temporal/components/TemporalStatusSelector';
import type { TemporalStatus } from '../temporal/components/TemporalStatusSelector';
import { TemporalDatePicker, validateTemporalDate } from '../temporal/components/TemporalDatePicker';

export interface FilterState {
  searchText: string;
  unitType: string | undefined;
  status: string | undefined;
  level: number | undefined;
  page: number;
  pageSize: number;
  // 时态筛选字段
  temporalMode?: 'current' | 'historical' | 'all';
  showOnlyTemporal?: boolean;
  temporalStatus?: TemporalStatus;
  effectiveDateFrom?: string;
  effectiveDateTo?: string;
  pointInTime?: string;
}

interface OrganizationFiltersProps {
  filters: FilterState;
  onFiltersChange: (filters: FilterState) => void;
  showTemporalFilters?: boolean;
}

interface SelectOption {
  label: string;
  value: string;
}

const UNIT_TYPE_OPTIONS: SelectOption[] = [
  { label: '全部类型', value: '' },
  { label: '部门', value: 'DEPARTMENT' },
  { label: '组织单位', value: 'ORGANIZATION_UNIT' },
  { label: '项目团队', value: 'PROJECT_TEAM' },
];

const STATUS_OPTIONS: SelectOption[] = [
  { label: '全部状态', value: '' },
  { label: '激活', value: 'ACTIVE' },
  { label: '停用', value: 'INACTIVE' },
  { label: '计划中', value: 'PLANNED' },
];

const TEMPORAL_MODE_OPTIONS: SelectOption[] = [
  { label: '当前组织', value: 'current' },
  { label: '历史数据', value: 'historical' },
  { label: '全部数据', value: 'all' },
];

const LEVEL_OPTIONS: SelectOption[] = [
  { label: '全部层级', value: '' },
  { label: '1级', value: '1' },
  { label: '2级', value: '2' },
  { label: '3级', value: '3' },
  { label: '4级', value: '4' },
  { label: '5级', value: '5' },
];

const PAGE_SIZE_OPTIONS: SelectOption[] = [
  { label: '10条/页', value: '10' },
  { label: '20条/页', value: '20' },
  { label: '50条/页', value: '50' },
  { label: '100条/页', value: '100' },
];

export const OrganizationFilters: React.FC<OrganizationFiltersProps> = ({
  filters,
  onFiltersChange,
  showTemporalFilters = false,
}) => {
  const [localSearchText, setLocalSearchText] = useState(filters.searchText);
  const [showAdvancedTemporal, setShowAdvancedTemporal] = useState(false);
  
  // 使用防抖处理搜索文本
  const debouncedSearchText = useDebounce(localSearchText, 300);
  
  // 当防抖后的搜索文本变化时，更新过滤器
  React.useEffect(() => {
    if (debouncedSearchText !== filters.searchText) {
      onFiltersChange({
        ...filters,
        searchText: debouncedSearchText,
        page: 1, // 搜索时重置到第一页
      });
    }
  }, [debouncedSearchText, filters, onFiltersChange]);

  const handleFilterChange = useCallback((key: keyof FilterState, value: FilterState[keyof FilterState]) => {
    const newFilters = {
      ...filters,
      [key]: value,
      page: key === 'pageSize' ? 1 : filters.page, // 改变页面大小时重置到第一页
    };
    
    if (key === 'pageSize') {
      // 确保页面大小改变时重置页码
      newFilters.page = 1;
    }
    
    onFiltersChange(newFilters);
  }, [filters, onFiltersChange]);

  const handleReset = useCallback(() => {
    const resetFilters: FilterState = {
      searchText: '',
      unitType: '',
      status: '',
      level: undefined,
      page: 1,
      pageSize: 20,
      // 重置时态筛选
      temporalMode: 'current',
      showOnlyTemporal: false,
      temporalStatus: undefined,
      effectiveDateFrom: undefined,
      effectiveDateTo: undefined,
      pointInTime: undefined,
    };
    setLocalSearchText('');
    setShowAdvancedTemporal(false);
    onFiltersChange(resetFilters);
  }, [onFiltersChange]);

  // 计算是否有激活的筛选条件
  const hasActiveFilters = useMemo(() => {
    return !!(
      filters.searchText ||
      filters.unitType ||
      filters.status ||
      filters.level ||
      (showTemporalFilters && (
        filters.showOnlyTemporal ||
        filters.temporalStatus ||
        filters.effectiveDateFrom ||
        filters.effectiveDateTo ||
        filters.pointInTime ||
        (filters.temporalMode && filters.temporalMode !== 'current')
      ))
    );
  }, [
    filters.searchText, filters.unitType, filters.status, filters.level,
    filters.showOnlyTemporal, filters.temporalStatus, filters.effectiveDateFrom,
    filters.effectiveDateTo, filters.pointInTime, filters.temporalMode,
    showTemporalFilters
  ]);

  return (
    <Card marginBottom="m">
      <Card.Heading>筛选条件</Card.Heading>
      <Card.Body>
        <Flex flexDirection="column" gap="m">
          {/* 基本筛选条件 */}
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px', alignItems: 'flex-end' }}>
            {/* 搜索框 */}
            <Box minWidth="200px">
              <FormField>
                <FormField.Label>组织名称</FormField.Label>
                <FormField.Field>
                  <TextInput
                    placeholder="搜索组织名称..."
                    value={localSearchText}
                    onChange={(e) => setLocalSearchText(e.target.value)}
                  />
                </FormField.Field>
              </FormField>
            </Box>

            {/* 类型筛选 */}
            <Box minWidth="150px">
              <FormField>
                <FormField.Label>组织类型</FormField.Label>
                <FormField.Field>
                  <select
                    value={filters.unitType || ''}
                    onChange={(e) => handleFilterChange('unitType', e.target.value || undefined)}
                    style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ddd', fontSize: '14px' }}
                  >
                    {UNIT_TYPE_OPTIONS.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </FormField.Field>
              </FormField>
            </Box>

            {/* 状态筛选 */}
            <Box minWidth="150px">
              <FormField>
                <FormField.Label>状态</FormField.Label>
                <FormField.Field>
                  <select
                    value={filters.status || ''}
                    onChange={(e) => handleFilterChange('status', e.target.value || undefined)}
                    style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ddd', fontSize: '14px' }}
                  >
                    {STATUS_OPTIONS.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </FormField.Field>
              </FormField>
            </Box>

            {/* 层级筛选 */}
            <Box minWidth="130px">
              <FormField>
                <FormField.Label>层级</FormField.Label>
                <FormField.Field>
                  <select
                    value={filters.level ? filters.level.toString() : ''}
                    onChange={(e) => handleFilterChange('level', e.target.value ? parseInt(e.target.value) : undefined)}
                    style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ddd', fontSize: '14px' }}
                  >
                    {LEVEL_OPTIONS.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </FormField.Field>
              </FormField>
            </Box>

            {/* 每页显示数量 */}
            <Box minWidth="130px">
              <FormField>
                <FormField.Label>显示数量</FormField.Label>
                <FormField.Field>
                  <select
                    value={filters.pageSize.toString()}
                    onChange={(e) => handleFilterChange('pageSize', parseInt(e.target.value))}
                    style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ddd', fontSize: '14px' }}
                  >
                    {PAGE_SIZE_OPTIONS.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </FormField.Field>
              </FormField>
            </Box>

            {/* 重置按钮 */}
            <Box>
              <SecondaryButton 
                onClick={handleReset}
                disabled={!hasActiveFilters}
              >
                重置筛选
              </SecondaryButton>
            </Box>
          </div>

          {/* 时态筛选区域 */}
          {showTemporalFilters && (
            <Card padding="s">
              <Card.Heading>时态筛选</Card.Heading>
              <Card.Body>
                <Flex flexDirection="column" gap="m">
                  {/* 时态模式选择 */}
                  <Flex gap="m" alignItems="flex-end">
                    <Box minWidth="150px">
                      <FormField>
                        <FormField.Label>时态模式</FormField.Label>
                        <FormField.Field>
                          <select
                            value={filters.temporalMode || 'current'}
                            onChange={(e) => handleFilterChange('temporalMode', e.target.value)}
                            style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ddd', fontSize: '14px' }}
                          >
                            {TEMPORAL_MODE_OPTIONS.map(option => (
                              <option key={option.value} value={option.value}>
                                {option.label}
                              </option>
                            ))}
                          </select>
                        </FormField.Field>
                      </FormField>
                    </Box>

                    {/* 仅显示时态组织开关 */}
                    <FormField>
                      <FormField.Label>仅显示时态组织</FormField.Label>
                      <FormField.Field>
                      <Switch 
                        checked={filters.showOnlyTemporal || false}
                        onChange={(e) => handleFilterChange('showOnlyTemporal', e.target.checked)}
                      />
                      </FormField.Field>
                    </FormField>

                    {/* 高级时态筛选开关 */}
                    <SecondaryButton
                      onClick={() => setShowAdvancedTemporal(!showAdvancedTemporal)}
                    >
                      {showAdvancedTemporal ? '收起' : '展开'}高级筛选
                    </SecondaryButton>
                  </Flex>

                  {/* 时态状态筛选 */}
                  <TemporalStatusSelector
                    label="时态状态"
                    value={filters.temporalStatus}
                    onChange={(value) => handleFilterChange('temporalStatus', value)}
                    includeAll={true}
                    placeholder="选择时态状态"
                  />

                  {/* 高级时态筛选 */}
                  {showAdvancedTemporal && (
                    <Flex flexDirection="column" gap="m">
                      {/* 生效日期范围 */}
                      <Flex gap="m">
                        <TemporalDatePicker
                          label="生效日期从"
                          value={filters.effectiveDateFrom || ''}
                          onChange={(value) => handleFilterChange('effectiveDateFrom', value || undefined)}
                          maxDate={filters.effectiveDateTo}
                          helperText="筛选在此日期之后生效的组织"
                        />
                        <TemporalDatePicker
                          label="生效日期到"
                          value={filters.effectiveDateTo || ''}
                          onChange={(value) => handleFilterChange('effectiveDateTo', value || undefined)}
                          minDate={filters.effectiveDateFrom}
                          helperText="筛选在此日期之前生效的组织"
                        />
                      </Flex>

                      {/* 历史时点查询 */}
                      <TemporalDatePicker
                        label="历史时点查询"
                        value={filters.pointInTime || ''}
                        onChange={(value) => handleFilterChange('pointInTime', value || undefined)}
                        maxDate={validateTemporalDate.getTodayString()}
                        helperText="查看在指定时间点有效的组织"
                      />
                    </Flex>
                  )}
                </Flex>
              </Card.Body>
            </Card>
          )}

          {/* 激活筛选条件提示 */}
          {hasActiveFilters && (
            <Box>
              <div style={{ fontSize: '12px', color: '#666' }}>
                已激活筛选条件:
                {filters.searchText && <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#e3f2fd', borderRadius: '4px' }}>名称: {filters.searchText}</span>}
                {filters.unitType && <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#e3f2fd', borderRadius: '4px' }}>类型: {UNIT_TYPE_OPTIONS.find(o => o.value === filters.unitType)?.label}</span>}
                {filters.status && <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#e3f2fd', borderRadius: '4px' }}>状态: {STATUS_OPTIONS.find(o => o.value === filters.status)?.label}</span>}
                {filters.level && <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#e3f2fd', borderRadius: '4px' }}>层级: {filters.level}级</span>}
                {showTemporalFilters && filters.temporalMode && filters.temporalMode !== 'current' && (
                  <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#fff3e0', borderRadius: '4px' }}>
                    时态: {TEMPORAL_MODE_OPTIONS.find(o => o.value === filters.temporalMode)?.label}
                  </span>
                )}
                {filters.showOnlyTemporal && (
                  <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#fff3e0', borderRadius: '4px' }}>
                    仅时态组织
                  </span>
                )}
                {filters.temporalStatus && (
                  <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#fff3e0', borderRadius: '4px' }}>
                    时态状态: {filters.temporalStatus}
                  </span>
                )}
                {(filters.effectiveDateFrom || filters.effectiveDateTo) && (
                  <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#fff3e0', borderRadius: '4px' }}>
                    生效期间: {filters.effectiveDateFrom || '开始'} - {filters.effectiveDateTo || '结束'}
                  </span>
                )}
                {filters.pointInTime && (
                  <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#fff3e0', borderRadius: '4px' }}>
                    时点: {validateTemporalDate.formatDateDisplay(filters.pointInTime)}
                  </span>
                )}
              </div>
            </Box>
          )}
        </Flex>
      </Card.Body>
    </Card>
  );
};