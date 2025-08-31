import React from 'react';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { validateTemporalDate } from '../utils/temporalValidation';

// 重新导出以保持向后兼容
export { validateTemporalDate };

export interface TemporalDatePickerProps {
  label: string;
  value?: string;
  onChange: (value: string) => void;
  error?: string;
  required?: boolean;
  minDate?: string;
  maxDate?: string;
  placeholder?: string;
  disabled?: boolean;
  helperText?: string;
}

export const TemporalDatePicker: React.FC<TemporalDatePickerProps> = ({
  label,
  value = '',
  onChange,
  error,
  required = false,
  minDate,
  maxDate,
  placeholder = 'YYYY-MM-DD',
  disabled = false,
  helperText,
}) => {
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange(event.target.value);
  };

  const inputProps = {
    type: 'date',
    value,
    onChange: handleChange,
    min: minDate,
    max: maxDate,
    disabled,
    placeholder,
  };

  return (
    <FormField
      isRequired={required}
      error={error ? "error" : undefined}
    >
      <FormField.Label>{label}</FormField.Label>
      <FormField.Field>
        <FormField.Input as={TextInput} {...inputProps} />
        {(error || helperText) && (
          <FormField.Hint>{error || helperText}</FormField.Hint>
        )}
      </FormField.Field>
    </FormField>
  );
};

