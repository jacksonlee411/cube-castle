import React from 'react';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { colors } from '@workday/canvas-kit-react/tokens';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

export interface UnitTypeSelectorProps {
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  label?: string;
  required?: boolean;
}

const unitTypeOptions = [
  {
    label: '组织单位',
    value: 'ORGANIZATION_UNIT',
    description: '企业的重要组织单位，负责特定职能和管理',
    color: colors.greenApple600,
  },
  {
    label: '部门',
    value: 'DEPARTMENT',
    description: '企业内部的功能性组织单位，执行特定业务职能',
    color: colors.blueberry600,
  },
  {
    label: '项目团队',
    value: 'PROJECT_TEAM',
    description: '临时性组织单位，专注于特定项目或任务的执行',
    color: colors.plum600,
  },
];

const UnitTypeSelector: React.FC<UnitTypeSelectorProps> = ({
  value,
  onChange,
  disabled = false,
  label = '组织类型',
  required = false,
}) => {
  const selectedOption = React.useMemo(
    () => unitTypeOptions.find((option) => option.value === value),
    [value]
  );

  return (
    <FormField isRequired={required}>
      <FormField.Label>{label} *</FormField.Label>
      <FormField.Field>
        <select
          value={value}
          onChange={(event) => onChange(event.target.value)}
          disabled={disabled}
          style={{
            width: '100%',
            padding: '8px',
            border: '1px solid #ddd',
            borderRadius: '4px',
            fontSize: '14px',
            color: selectedOption?.color ?? colors.licorice500,
          }}
          data-testid={temporalEntitySelectors.form?.field?.unitType}
        >
          {unitTypeOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </FormField.Field>
      {selectedOption && (
        <FormField.Hint>{selectedOption.description}</FormField.Hint>
      )}
    </FormField>
  );
};

export default UnitTypeSelector;
