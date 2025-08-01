import React, { useState, useCallback } from 'react';
import { cn } from '@/lib/utils';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';

export interface FilterOption {
  key: string;
  label: string;
  type: 'text' | 'select' | 'multiselect' | 'date' | 'daterange';
  options?: Array<{ label: string; value: string }>;
  placeholder?: string;
}

export interface ActiveFilter {
  key: string;
  label: string;
  value: string;
  displayValue: string;
}

export interface SmartFilterProps {
  /** 可用的筛选选项 */
  filterOptions: FilterOption[];
  /** 当前激活的筛选条件 */
  activeFilters: ActiveFilter[];
  /** 筛选条件变化回调 */
  onFiltersChange: (filters: ActiveFilter[]) => void;
  /** 搜索关键词 */
  searchValue: string;
  /** 搜索变化回调 */
  onSearchChange: (value: string) => void;
  /** 搜索提示文本 */
  searchPlaceholder?: string;
  /** 预设筛选方案 */
  presets?: Array<{
    label: string;
    filters: ActiveFilter[];
    icon?: React.ReactNode;
  }>;
  /** 是否显示高级筛选 */
  showAdvanced?: boolean;
  /** 自定义样式 */
  className?: string;
}

export function SmartFilter({
  filterOptions,
  activeFilters,
  onFiltersChange,
  searchValue,
  onSearchChange,
  searchPlaceholder = '搜索员工姓名、工号、部门...',
  presets = [],
  showAdvanced = true,
  className
}: SmartFilterProps) {
  const [showAdvancedPanel, setShowAdvancedPanel] = useState(false);

  // 添加筛选条件
  const addFilter = useCallback((option: FilterOption, value: string) => {
    if (!value || (Array.isArray(value) && value.length === 0)) return;

    const displayValue = Array.isArray(value) 
      ? value.map(v => {
          const opt = option.options?.find(o => o.value === v);
          return opt ? opt.label : String(v);
        }).join(', ')
      : option.options?.find(o => o.value === value)?.label ?? String(value);

    const newFilter: ActiveFilter = {
      key: option.key,
      label: option.label,
      value,
      displayValue
    };

    // 替换同key的筛选条件
    const updatedFilters = activeFilters.filter(f => f.key !== option.key);
    onFiltersChange([...updatedFilters, newFilter]);
  }, [activeFilters, onFiltersChange]);

  // 移除筛选条件
  const removeFilter = useCallback((key: string) => {
    onFiltersChange(activeFilters.filter(f => f.key !== key));
  }, [activeFilters, onFiltersChange]);

  // 清除所有筛选条件
  const clearAllFilters = useCallback(() => {
    onFiltersChange([]);
    onSearchChange('');
  }, [onFiltersChange, onSearchChange]);

  // 应用预设方案
  const applyPreset = useCallback((preset: typeof presets[0]) => {
    onFiltersChange(preset.filters);
  }, [onFiltersChange]);

  // 渲染快速筛选按钮
  const renderQuickFilters = () => {
    const quickOptions = filterOptions
      .filter(option => option.type === 'select' && option.options)
      .slice(0, 3);

    return quickOptions.map((option) => {
      // 找到当前选项的活跃值
      const activeFilter = activeFilters.find(f => f.key === option.key);
      const currentValue = activeFilter?.value || undefined; // 使用undefined而不是空字符串
      
      return (
        <Select 
          key={option.key} 
          value={currentValue}
          onValueChange={(value) => addFilter(option, value)}
        >
          <SelectTrigger className="w-auto min-w-[120px] h-9 text-sm">
            <SelectValue placeholder={option.label} />
          </SelectTrigger>
          <SelectContent>
            {option.options?.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      );
    });
  };

  // 渲染高级筛选面板
  const renderAdvancedPanel = () => (
    <Card className="p-4 space-y-4 border-dashed">
      <div className="flex items-center justify-between">
        <h4 className="text-display-small font-medium">高级筛选</h4>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setShowAdvancedPanel(false)}
        >
          ✕
        </Button>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filterOptions.map((option) => (
          <div key={option.key} className="space-y-2">
            <label className="text-body-small font-medium text-gray-700">
              {option.label}
            </label>
            
            {option.type === 'text' && (
              <Input
                placeholder={option.placeholder}
                onBlur={(e) => e.target.value && addFilter(option, e.target.value)}
                className="h-9"
              />
            )}
            
            {option.type === 'select' && (
              <Select 
                value={activeFilters.find(f => f.key === option.key)?.value || undefined}
                onValueChange={(value) => addFilter(option, value)}
              >
                <SelectTrigger className="h-9">
                  <SelectValue placeholder={option.placeholder || '请选择'} />
                </SelectTrigger>
                <SelectContent>
                  {option.options?.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
            
            {option.type === 'date' && (
              <Input
                type="date"
                className="h-9"
                onChange={(e) => e.target.value && addFilter(option, e.target.value)}
              />
            )}
          </div>
        ))}
      </div>
    </Card>
  );

  return (
    <div className={cn('space-y-4', className)}>
      {/* 主筛选工具栏 */}
      <Card className="p-3 sm:p-4">
        <div className="flex flex-col gap-4">
          {/* 搜索框 */}
          <div className="relative">
            <span className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground">
              🔍
            </span>
            <Input
              value={searchValue}
              onChange={(e) => onSearchChange(e.target.value)}
              placeholder={searchPlaceholder}
              className="pl-10 h-10"
            />
          </div>

          {/* 筛选控件行 */}
          <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3 sm:gap-3">
            {/* 快速筛选 */}
            <div className="flex items-center gap-2 flex-wrap">
              {renderQuickFilters()}
            </div>
            
            <div className="flex items-center gap-2 flex-wrap">
              {/* 预设方案 */}
              {presets.length > 0 && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="sm" className="h-9 text-xs sm:text-sm">
                      📋 <span className="hidden sm:inline ml-1">预设方案</span>
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-48">
                    {presets.map((preset, index) => (
                      <DropdownMenuItem
                        key={index}
                        onClick={() => applyPreset(preset)}
                        className="flex items-center gap-2"
                      >
                        {preset.icon}
                        {preset.label}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              )}

              {/* 高级筛选按钮 */}
              {showAdvanced && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowAdvancedPanel(!showAdvancedPanel)}
                  className={cn(
                    'h-9 text-xs sm:text-sm',
                    showAdvancedPanel && 'bg-primary/10 border-primary text-primary'
                  )}
                >
                  ⚙️ <span className="hidden sm:inline ml-1">高级筛选</span>
                </Button>
              )}

              {/* 清除按钮 */}
              {(activeFilters.length > 0 || searchValue) && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={clearAllFilters}
                  className="h-9 text-muted-foreground hover:text-foreground text-xs sm:text-sm"
                >
                  清除
                </Button>
              )}
            </div>
          </div>
        </div>
      </Card>

      {/* 激活的筛选条件标签 */}
      {activeFilters.length > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-body-small text-gray-500">已应用筛选:</span>
          {activeFilters.map((filter) => (
            <Badge
              key={filter.key}
              variant="secondary"
              className="flex items-center gap-1 px-3 py-1 bg-primary/10 text-primary hover:bg-primary/20"
            >
              <span className="text-xs font-medium">{filter.label}:</span>
              <span className="text-xs">{filter.displayValue}</span>
              <button
                onClick={() => removeFilter(filter.key)}
                className="ml-1 text-primary/70 hover:text-primary text-xs"
              >
                ✕
              </button>
            </Badge>
          ))}
        </div>
      )}

      {/* 高级筛选面板 */}
      {showAdvancedPanel && renderAdvancedPanel()}

      {/* 筛选结果统计 */}
      <div className="flex items-center justify-between text-body-small text-gray-500">
        <div className="flex items-center gap-2">
          {(activeFilters.length > 0 || searchValue) && (
            <>
              <span>🔍</span>
              <span>
                已应用 {activeFilters.length + (searchValue ? 1 : 0)} 个筛选条件
              </span>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export default SmartFilter;