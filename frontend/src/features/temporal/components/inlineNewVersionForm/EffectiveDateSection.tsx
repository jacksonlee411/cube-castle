import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Heading } from '@workday/canvas-kit-react/text';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { colors } from '@workday/canvas-kit-react/tokens';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';
import {
  organizationFieldHelperText,
  organizationFieldLabel,
  organizationFieldRequired,
} from '../../manifest/helpers';

export interface EffectiveDateSectionProps {
  value: string;
  error?: string;
  onChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  disabled: boolean;
}

const EffectiveDateSection: React.FC<EffectiveDateSectionProps> = ({
  value,
  error,
  onChange,
  disabled,
}) => {
  const label = organizationFieldLabel('effectiveDate', '生效日期');
  const helperText = organizationFieldHelperText('effectiveDate', error);

  return (
    <Box marginBottom="l">
      <Heading size="small" marginBottom="s" color={colors.blueberry600}>
        {label}
      </Heading>
      <Box marginLeft="m">
        <FormField isRequired={organizationFieldRequired('effectiveDate', true)} error={error ? 'error' : undefined}>
          <FormField.Label>{label}</FormField.Label>
          <FormField.Field>
            <TextInput
              type="date"
              value={value}
              onChange={onChange}
              disabled={disabled}
              data-testid={temporalEntitySelectors.form?.field?.effectiveDate}
            />
            {helperText ? <FormField.Hint>{helperText}</FormField.Hint> : null}
          </FormField.Field>
        </FormField>
      </Box>
    </Box>
  );
};

export default EffectiveDateSection;
