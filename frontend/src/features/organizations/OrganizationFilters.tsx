import React, { useState, useCallback, useMemo } from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Card } from '@workday/canvas-kit-react/card';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { SecondaryButton } from '@workday/canvas-kit-react/button';
import { useDebounce } from '../../shared/hooks/useDebounce';

export interface FilterState {
  searchText: string;
  unit_type: string | undefined;
  status: string | undefined;
  level: number | undefined;
  page: number;
  pageSize: number;
}

interface OrganizationFiltersProps {
  filters: FilterState;
  onFiltersChange: (filters: FilterState) => void;
}

interface SelectOption {
  label: string;
  value: string;
}

const UNIT_TYPE_OPTIONS: SelectOption[] = [
  { label: '全部类型', value: '' },
  { label: '部门', value: 'DEPARTMENT' },
  { label: '成本中心', value: 'COST_CENTER' },
  { label: '公司', value: 'COMPANY' },
  { label: '项目团队', value: 'PROJECT_TEAM' },
];

const STATUS_OPTIONS: SelectOption[] = [
  { label: '全部状态', value: '' },
  { label: '激活', value: 'ACTIVE' },
  { label: '停用', value: 'INACTIVE' },
  { label: '计划中', value: 'PLANNED' },
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
}) => {
  const [localSearchText, setLocalSearchText] = useState(filters.searchText);
  
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

  const handleFilterChange = useCallback((key: keyof FilterState, value: any) => {
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
      unit_type: '',
      status: '',
      level: undefined,
      page: 1,
      pageSize: 20,
    };
    setLocalSearchText('');
    onFiltersChange(resetFilters);
  }, [onFiltersChange]);

  // 计算是否有激活的筛选条件
  const hasActiveFilters = useMemo(() => {
    return !!(
      filters.searchText ||
      filters.unit_type ||
      filters.status ||
      filters.level
    );
  }, [filters.searchText, filters.unit_type, filters.status, filters.level]);

  return (
    <Card marginBottom="m">
      <Card.Heading>筛选条件</Card.Heading>
      <Card.Body>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px', alignItems: 'flex-end' }}>
          {/* 搜索框 */}
          <Box minWidth="200px">
            <FormField>
              <FormField.Label>组织名称</FormField.Label>
              <FormField.Field>
                <FormField.Input
                  as={TextInput}
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
                  value={filters.unit_type || ''}
                  onChange={(e) => handleFilterChange('unit_type', e.target.value || undefined)}
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

        {/* 激活筛选条件提示 */}
        {hasActiveFilters && (
          <Box marginTop="s">
            <div style={{ fontSize: '12px', color: '#666' }}>
              已激活筛选条件:
              {filters.searchText && <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#e3f2fd', borderRadius: '4px' }}>名称: {filters.searchText}</span>}
              {filters.unit_type && <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#e3f2fd', borderRadius: '4px' }}>类型: {UNIT_TYPE_OPTIONS.find(o => o.value === filters.unit_type)?.label}</span>}
              {filters.status && <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#e3f2fd', borderRadius: '4px' }}>状态: {STATUS_OPTIONS.find(o => o.value === filters.status)?.label}</span>}
              {filters.level && <span style={{ marginLeft: '8px', padding: '2px 6px', backgroundColor: '#e3f2fd', borderRadius: '4px' }}>层级: {filters.level}级</span>}
            </div>
          </Box>
        )}
      </Card.Body>
    </Card>
  );
};