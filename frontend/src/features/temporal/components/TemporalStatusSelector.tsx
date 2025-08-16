import React from 'react';
// import { Select, SelectOption } from '@workday/canvas-kit-react/select';
import { FormField } from '@workday/canvas-kit-react/form-field';

export type TemporalStatus = 'ACTIVE' | 'PLANNED' | 'INACTIVE';

export interface TemporalStatusOption {
  value: TemporalStatus;
  label: string;
  description: string;
}

export const TEMPORAL_STATUS_OPTIONS: TemporalStatusOption[] = [
  {
    value: 'ACTIVE',
    label: '启用',
    description: '当前生效的组织'
  },
  {
    value: 'PLANNED', 
    label: '计划',
    description: '计划在未来生效的组织'
  },
  {
    value: 'INACTIVE',
    label: '停用',
    description: '已失效或停用的组织'
  }
];

export interface TemporalStatusSelectorProps {
  label?: string;
  value?: TemporalStatus;
  onChange: (value: TemporalStatus) => void;
  error?: string;
  required?: boolean;
  disabled?: boolean;
  placeholder?: string;
  includeAll?: boolean;
  helperText?: string;
}

export const TemporalStatusSelector: React.FC<TemporalStatusSelectorProps> = ({
  label = '时态状态',
  value,
  onChange,
  error,
  required = false,
  disabled = false,
  placeholder = '请选择状态',
  includeAll = false,
  helperText,
}) => {
  const options = includeAll 
    ? [{ value: '', label: '全部状态', description: '显示所有时态状态的组织' }, ...TEMPORAL_STATUS_OPTIONS]
    : TEMPORAL_STATUS_OPTIONS;

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const selectedValue = event.target.value;
    if (selectedValue && selectedValue !== '') {
      onChange(selectedValue as TemporalStatus);
    }
  };

  return (
    <FormField
      isRequired={required}
      error={error ? "error" : undefined}
    >
      <FormField.Label>{label}</FormField.Label>
      <FormField.Field>
        <select
          value={value || ''}
          onChange={handleChange}
          disabled={disabled}
          style={{ 
            width: '100%', 
            padding: '8px', 
            borderRadius: '4px', 
            border: '1px solid #ddd', 
            fontSize: '14px' 
          }}
        >
          <option value="" disabled>{placeholder}</option>
          {options.map((option) => (
            <option 
              key={option.value} 
              value={option.value}
              title={option.description}
            >
              {option.label}
            </option>
          ))}
        </select>
        {(error || helperText) && (
          <FormField.Hint>{error || helperText}</FormField.Hint>
        )}
      </FormField.Field>
    </FormField>
  );
};

// 时态状态工具函数
export const temporalStatusUtils = {
  // 获取状态标签
  getStatusLabel: (status: TemporalStatus): string => {
    const option = TEMPORAL_STATUS_OPTIONS.find(opt => opt.value === status);
    return option?.label || status;
  },

  // 获取状态描述
  getStatusDescription: (status: TemporalStatus): string => {
    const option = TEMPORAL_STATUS_OPTIONS.find(opt => opt.value === status);
    return option?.description || '';
  },

  // 根据日期计算状态
  calculateStatus: (effectiveDate?: string, endDate?: string): TemporalStatus => {
    const today = new Date().toISOString().split('T')[0];
    
    // 如果没有生效日期，默认为启用
    if (!effectiveDate) return 'ACTIVE';
    
    // 如果生效日期在未来，为计划状态
    if (effectiveDate > today) return 'PLANNED';
    
    // 如果有结束日期且已过期，为停用状态
    if (endDate && endDate < today) return 'INACTIVE';
    
    // 其他情况为启用状态
    return 'ACTIVE';
  },

  // 判断是否为时态组织
  isTemporal: (effectiveDate?: string, endDate?: string): boolean => {
    return !!(effectiveDate || endDate);
  },

  // 获取状态颜色
  getStatusColor: (status: TemporalStatus): string => {
    switch (status) {
      case 'ACTIVE': return '#00A844'; // 绿色
      case 'PLANNED': return '#0875E1'; // 蓝色  
      case 'INACTIVE': return '#999999'; // 灰色
      default: return '#333333';
    }
  },

  // 获取状态图标
  getStatusIcon: (status: TemporalStatus): string => {
    switch (status) {
      case 'ACTIVE': return '✓';
      case 'PLANNED': return '';
      case 'INACTIVE': return '';
      default: return '•';
    }
  }
};