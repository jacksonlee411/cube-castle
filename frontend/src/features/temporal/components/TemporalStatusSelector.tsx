import React from 'react';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TEMPORAL_STATUS_OPTIONS, temporalStatusUtils, type TemporalStatus } from '../constants/temporalStatus';

// 重新导出以保持向后兼容
export { TEMPORAL_STATUS_OPTIONS, temporalStatusUtils, type TemporalStatus };

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

