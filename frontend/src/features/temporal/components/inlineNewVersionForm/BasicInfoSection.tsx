import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import { SecondaryButton } from '@workday/canvas-kit-react/button';
import { colors } from '@workday/canvas-kit-react/tokens';
import { StatusBadge } from '../../../../shared/components/StatusBadge';
import ParentOrganizationSelector from '../ParentOrganizationSelector';
import type { TemporalEditFormData } from '../TemporalEditForm';
import { mapLifecycleStatusToOrganizationStatus } from './utils';
import UnitTypeSelector from './UnitTypeSelector';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

export interface BasicInfoSectionProps {
  formData: TemporalEditFormData;
  errors: Record<string, string>;
  disabled: boolean;
  organizationCode: string | null;
  onFieldChange: (
    field: keyof TemporalEditFormData
  ) => (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => void;
  onParentChange: (parentCode: string | undefined) => void;
  onParentError: (message?: string) => void;
  parentError: string;
  suggestedEffectiveDate?: string;
  onApplySuggestedEffectiveDate: () => void;
  onResetParentSelection: () => void;
  isSubmitting: boolean;
  onUnitTypeChange: (value: string) => void;
}

const BasicInfoSection: React.FC<BasicInfoSectionProps> = ({
  formData,
  errors,
  disabled,
  organizationCode,
  onFieldChange,
  onParentChange,
  onParentError,
  parentError,
  suggestedEffectiveDate,
  onApplySuggestedEffectiveDate,
  onResetParentSelection,
  isSubmitting,
  onUnitTypeChange,
}) => {
  return (
    <Box marginBottom="l">
      <Heading size="small" marginBottom="s" color={colors.blueberry600}>
        基本信息
      </Heading>
      <Box marginLeft="m">
        <FormField isRequired error={errors.name ? 'error' : undefined}>
          <FormField.Label>组织名称 *</FormField.Label>
          <FormField.Field>
            <TextInput
              value={formData.name}
              onChange={onFieldChange('name')}
              placeholder="请输入组织名称"
              disabled={disabled}
              data-testid={temporalEntitySelectors.form?.field?.name}
            />
            {errors.name ? <FormField.Hint>{errors.name}</FormField.Hint> : null}
          </FormField.Field>
        </FormField>

        <Box marginTop="m">
          <ParentOrganizationSelector
            currentCode={organizationCode ?? ''}
            effectiveDate={formData.effectiveDate}
            currentParentCode={formData.parentCode}
            onChange={onParentChange}
            onValidationError={onParentError}
            disabled={disabled}
          />
          {parentError ? (
            <Text typeLevel="subtext.small" color="error" marginTop="xs">
              {parentError}
            </Text>
          ) : null}
          {suggestedEffectiveDate ? (
            <Flex gap="s" marginTop="xs">
              <SecondaryButton
                type="button"
                onClick={onApplySuggestedEffectiveDate}
                disabled={isSubmitting}
              >
                调整生效日期至 {suggestedEffectiveDate}
              </SecondaryButton>
              <SecondaryButton
                type="button"
                onClick={onResetParentSelection}
                disabled={isSubmitting}
              >
                重新选择上级组织
              </SecondaryButton>
            </Flex>
          ) : null}
          <Text typeLevel="subtext.small" color="hint" marginTop="xs">
            仅允许选择在生效日期有效且状态为 ACTIVE 的组织
          </Text>
        </Box>

        <UnitTypeSelector
          value={formData.unitType}
          onChange={onUnitTypeChange}
          disabled={disabled}
          label="组织类型"
          required
        />

        <FormField>
          <FormField.Label>组织状态 *</FormField.Label>
          <FormField.Field>
            <StatusBadge status={mapLifecycleStatusToOrganizationStatus(formData.lifecycleStatus)} size="medium" />
            <Text typeLevel="subtext.small" color="hint" marginTop="xs">
              状态由系统根据操作自动管理
            </Text>
          </FormField.Field>
        </FormField>

        <FormField>
          <FormField.Label>描述信息</FormField.Label>
          <FormField.Field>
            <TextArea
              value={formData.description}
              onChange={onFieldChange('description')}
              placeholder="请输入组织描述信息"
              disabled={disabled}
              rows={3}
              data-testid={temporalEntitySelectors.form?.field?.description}
            />
          </FormField.Field>
        </FormField>
      </Box>
    </Box>
  );
};

export default BasicInfoSection;
