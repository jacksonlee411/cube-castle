import React, { useState, useCallback, useMemo, useRef, useEffect } from 'react';
import { cn } from '@/lib/utils';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
// 暂时移除Radix UI组件，使用原生实现避免循环依赖
// import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
// import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';

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

export interface SmartFilterStableProps {
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

export function SmartFilterStable({
  filterOptions,
  activeFilters,
  onFiltersChange,
  searchValue,
  onSearchChange,
  searchPlaceholder = '搜索员工姓名、工号、部门...',
  presets = [],
  showAdvanced = true,
  className
}: SmartFilterStableProps) {
  const [showAdvancedPanel, setShowAdvancedPanel] = useState(false);
  
  // 使用ref来保存稳定的状态快照，避免循环依赖
  const filtersRef = useRef<ActiveFilter[]>(activeFilters);
  const updatePendingRef = useRef(false);
  
  // 同步外部状态到ref（单向数据流）
  useEffect(() => {
    if (!updatePendingRef.current) {
      filtersRef.current = activeFilters;
    }
  }, [activeFilters]);

  // 稳定的更新函数，使用debounce避免频繁更新
  const updateFiltersStable = useCallback((newFilters: ActiveFilter[]) => {
    filtersRef.current = newFilters;
    updatePendingRef.current = true;
    
    // 使用setTimeout批处理更新，避免同步循环
    setTimeout(() => {
      onFiltersChange(newFilters);
      updatePendingRef.current = false;
    }, 0);
  }, [onFiltersChange]);

  // 添加筛选条件 - 完全基于ref的稳定实现
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

    // 基于current ref状态，避免依赖state
    const currentFilters = [...filtersRef.current];
    const updatedFilters = currentFilters.filter(f => f.key !== option.key);
    updatedFilters.push(newFilter);
    
    updateFiltersStable(updatedFilters);
  }, [updateFiltersStable]);

  // 移除筛选条件 - 基于ref的稳定实现
  const removeFilter = useCallback((key: string) => {
    const currentFilters = [...filtersRef.current];
    const updatedFilters = currentFilters.filter(f => f.key !== key);
    updateFiltersStable(updatedFilters);
  }, [updateFiltersStable]);

  // 清除所有筛选条件
  const clearAllFilters = useCallback(() => {
    updateFiltersStable([]);
    onSearchChange('');
  }, [updateFiltersStable, onSearchChange]);

  // 应用预设方案
  const applyPreset = useCallback((preset: typeof presets[0]) => {
    updateFiltersStable([...preset.filters]);
  }, [updateFiltersStable]);

  // 原生Select组件，避免Radix UI循环依赖
  const NativeSelectFilter = useCallback(({ option }: { option: FilterOption }) => {
    const activeFilter = activeFilters.find(f => f.key === option.key);
    const currentValue = activeFilter?.value ?? '';
    
    return (
      <select
        key={`native-select-${option.key}`}
        value={currentValue}
        onChange={(e) => {
          if (e.target.value) {
            addFilter(option, e.target.value);
          } else {
            removeFilter(option.key);
          }
        }}
        className="flex h-9 w-auto min-w-[120px] items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
      >
        <option value="">{option.label}</option>
        {option.options?.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    );
  }, [activeFilters, addFilter, removeFilter]);

// 原生下拉菜单组件，避免Radix UI循环依赖
const NativeDropdownMenu: React.FC<{ 
  children: React.ReactNode, 
  items: Array<{ label: string, onClick: () => void, icon?: React.ReactNode }> 
}> = ({ children, items }) => {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
    
    // 必须有返回值或返回undefined
    return undefined;
  }, [isOpen]);

  return (
    <div className="relative" ref={menuRef}>
      <div onClick={() => setIsOpen(!isOpen)}>
        {children}
      </div>
      {isOpen && (
        <div className="absolute right-0 top-full mt-1 w-48 rounded-md border bg-popover p-1 text-popover-foreground shadow-md z-50">
          {items.map((item, index) => (
            <button
              key={index}
              onClick={() => {
                item.onClick();
                setIsOpen(false);
              }}
              className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
            >
              {item.icon}
              {item.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

NativeDropdownMenu.displayName = 'NativeDropdownMenu';

  // 渲染快速筛选按钮 - 移除循环依赖
  const renderQuickFilters = useMemo(() => {
    const quickOptions = filterOptions
      .filter(option => option.type === 'select' && option.options)
      .slice(0, 3);

    return quickOptions.map((option) => (
      <NativeSelectFilter key={option.key} option={option} />
    ));
  }, [filterOptions, NativeSelectFilter]); // 使用原生组件

  // 渲染高级筛选面板 - 稳定化实现
  const renderAdvancedPanel = useCallback(() => (
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
              <select
                value={activeFilters.find(f => f.key === option.key)?.value ?? ''}
                onChange={(e) => {
                  if (e.target.value) {
                    addFilter(option, e.target.value);
                  } else {
                    removeFilter(option.key);
                  }
                }}
                className="flex h-9 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <option value="">{option.placeholder ?? '请选择'}</option>
                {option.options?.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
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
  ), [filterOptions, activeFilters, addFilter, removeFilter]);

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
              {renderQuickFilters}
            </div>
            
            <div className="flex items-center gap-2 flex-wrap">
              {/* 预设方案 - 使用原生下拉菜单 */}
              {presets.length > 0 && (
                <NativeDropdownMenu
                  items={presets.map((preset) => ({
                    label: preset.label,
                    icon: preset.icon,
                    onClick: () => applyPreset(preset)
                  }))}
                >
                  <Button variant="outline" size="sm" className="h-9 text-xs sm:text-sm">
                    📋 <span className="hidden sm:inline ml-1">预设方案</span>
                  </Button>
                </NativeDropdownMenu>
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

export default SmartFilterStable;