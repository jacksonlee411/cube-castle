import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Heading } from '@workday/canvas-kit-react/text';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { colors } from '@workday/canvas-kit-react/tokens';

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
  return (
    <Box marginBottom="l">
      <Heading size="small" marginBottom="s" color={colors.blueberry600}>
        生效日期
      </Heading>
      <Box marginLeft="m">
        <FormField isRequired error={error ? 'error' : undefined}>
          <FormField.Label>生效日期 *</FormField.Label>
          <FormField.Field>
            <TextInput
              type="date"
              value={value}
              onChange={onChange}
              disabled={disabled}
              data-testid="form-field-effective-date"
            />
            {error ? <FormField.Hint>{error}</FormField.Hint> : null}
          </FormField.Field>
        </FormField>
      </Box>
    </Box>
  );
};

export default EffectiveDateSection;
