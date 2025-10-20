import React from 'react'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { TextArea } from '@workday/canvas-kit-react/text-area'
import { FormField } from '@workday/canvas-kit-react/form-field'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { Text } from '@workday/canvas-kit-react/text'
import { SimpleStack } from '../layout/SimpleStack'
import type { PositionFormErrors, PositionFormState, SelectOption } from './types'

const POSITION_TYPES: SelectOption[] = [
  { label: '正式职位 (REGULAR)', value: 'REGULAR' },
  { label: '临时职位 (TEMPORARY)', value: 'TEMPORARY' },
  { label: '合同工 (CONTRACTOR)', value: 'CONTRACTOR' },
]

const EMPLOYMENT_TYPES: SelectOption[] = [
  { label: '全职 (FULL_TIME)', value: 'FULL_TIME' },
  { label: '兼职 (PART_TIME)', value: 'PART_TIME' },
  { label: '实习 (INTERN)', value: 'INTERN' },
]

interface SelectFieldProps {
  label: string
  value: string
  onChange: React.ChangeEventHandler<HTMLSelectElement>
  options: SelectOption[]
  error?: string
  helperText?: string
  isRequired?: boolean
  disabled?: boolean
}

const SELECT_BASE_STYLE: React.CSSProperties = {
  width: '100%',
  padding: '8px 12px',
  borderRadius: 8,
  border: `1px solid ${colors.soap500}`,
  backgroundColor: colors.frenchVanilla100,
  fontSize: '14px',
  lineHeight: '20px',
  appearance: 'none',
}

const SELECT_ERROR_STYLE: React.CSSProperties = {
  borderColor: colors.cinnamon500,
}

const toErrorType = (message?: string): 'error' | undefined => (message ? 'error' : undefined)

const renderFieldHint = (error?: string, helperText?: string) => {
  const hint = error ?? helperText
  return hint ? <FormField.Hint>{hint}</FormField.Hint> : null
}

interface TextInputFieldProps {
  label: string
  value: string
  onChange: React.ChangeEventHandler<HTMLInputElement>
  placeholder?: string
  type?: React.HTMLInputTypeAttribute
  error?: string
  helperText?: string
  isRequired?: boolean
  disabled?: boolean
  min?: number | string
  step?: number | string
}

const TextInputField: React.FC<TextInputFieldProps> = ({
  label,
  value,
  onChange,
  placeholder,
  type,
  error,
  helperText,
  isRequired,
  disabled,
  min,
  step,
}) => (
  <FormField isRequired={isRequired} error={toErrorType(error)}>
    <FormField.Label>{label}</FormField.Label>
    <FormField.Field>
      <FormField.Input
        as={TextInput}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        type={type}
        disabled={disabled}
        min={min}
        step={step}
      />
      {renderFieldHint(error, helperText)}
    </FormField.Field>
  </FormField>
)

interface TextAreaFieldProps {
  label: string
  value: string
  onChange: React.ChangeEventHandler<HTMLTextAreaElement>
  placeholder?: string
  error?: string
  helperText?: string
  isRequired?: boolean
  disabled?: boolean
  rows?: number
}

const TextAreaField: React.FC<TextAreaFieldProps> = ({
  label,
  value,
  onChange,
  placeholder,
  error,
  helperText,
  isRequired,
  disabled,
  rows = 3,
}) => (
  <FormField isRequired={isRequired} error={toErrorType(error)}>
    <FormField.Label>{label}</FormField.Label>
    <FormField.Field>
      <FormField.Input
        as={TextArea}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        disabled={disabled}
        rows={rows}
      />
      {renderFieldHint(error, helperText)}
    </FormField.Field>
  </FormField>
)

const responsiveRowStyle = {
  flexDirection: 'column',
  '@media (min-width: 768px)': {
    flexDirection: 'row',
  },
} as const

const SelectField: React.FC<SelectFieldProps> = ({
  label,
  value,
  onChange,
  options,
  error,
  helperText,
  isRequired,
  disabled,
}) => (
  <FormField isRequired={isRequired} error={toErrorType(error)}>
    <FormField.Label>{label}</FormField.Label>
    <FormField.Field>
      <select
        value={value}
        onChange={onChange}
        style={{ ...SELECT_BASE_STYLE, ...(error ? SELECT_ERROR_STYLE : {}) }}
        disabled={disabled}
      >
        {options.map(option => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      {renderFieldHint(error, helperText)}
    </FormField.Field>
  </FormField>
)

interface PositionFormFieldsProps {
  state: PositionFormState
  errors: PositionFormErrors
  onChange: (
    key: keyof PositionFormState,
  ) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void
  isVersion: boolean
  catalogAvailable: boolean
  catalogLoading: boolean
  jobFamilyGroupOptions: SelectOption[]
  jobFamilyOptions: SelectOption[]
  jobRoleOptions: SelectOption[]
  jobLevelOptions: SelectOption[]
  onJobFamilyGroupChange: React.ChangeEventHandler<HTMLSelectElement>
  onJobFamilyChange: React.ChangeEventHandler<HTMLSelectElement>
  onJobRoleChange: React.ChangeEventHandler<HTMLSelectElement>
  onJobLevelChange: React.ChangeEventHandler<HTMLSelectElement>
}

const renderTextInputs = (
  state: PositionFormState,
  errors: PositionFormErrors,
  onChange: PositionFormFieldsProps['onChange'],
) => (
  <>
    <TextInputField
      label="职位名称"
      value={state.title}
      onChange={onChange('title')}
      placeholder="请输入职位名称"
      isRequired
      error={errors.title}
    />

    <Flex gap={space.m} cs={responsiveRowStyle}>
      <TextInputField
        label="职类编码"
        value={state.jobFamilyGroupCode}
        onChange={onChange('jobFamilyGroupCode')}
        placeholder="例如：PROF"
        isRequired
        error={errors.jobFamilyGroupCode}
      />
      <TextInputField
        label="职种编码"
        value={state.jobFamilyCode}
        onChange={onChange('jobFamilyCode')}
        placeholder="例如：PROF-IT"
        isRequired
        error={errors.jobFamilyCode}
      />
    </Flex>

    <Flex gap={space.m} cs={responsiveRowStyle}>
      <TextInputField
        label="职务编码"
        value={state.jobRoleCode}
        onChange={onChange('jobRoleCode')}
        placeholder="例如：PROF-IT-BKND"
        isRequired
        error={errors.jobRoleCode}
      />
      <TextInputField
        label="职级编码"
        value={state.jobLevelCode}
        onChange={onChange('jobLevelCode')}
        placeholder="例如：P5"
        isRequired
        error={errors.jobLevelCode}
      />
    </Flex>
  </>
)

export const PositionFormFields: React.FC<PositionFormFieldsProps> = ({
  state,
  errors,
  onChange,
  isVersion,
  catalogAvailable,
  catalogLoading,
  jobFamilyGroupOptions,
  jobFamilyOptions,
  jobRoleOptions,
  jobLevelOptions,
  onJobFamilyGroupChange,
  onJobFamilyChange,
  onJobRoleChange,
  onJobLevelChange,
}) => (
  <SimpleStack gap={space.xs}>
    {!catalogAvailable && (
      <Text typeLevel="subtext.small" color={colors.cinnamon500}>
        未能加载岗位字典数据，可手动填写编码；请稍后刷新页面。
      </Text>
    )}

    {catalogAvailable ? (
      <>
        <TextInputField
          label="职位名称"
          value={state.title}
          onChange={onChange('title')}
          placeholder="请输入职位名称"
          isRequired
          error={errors.title}
        />

        <Flex gap={space.m} cs={responsiveRowStyle}>
          <SelectField
            label="职类"
            value={state.jobFamilyGroupCode}
            onChange={onJobFamilyGroupChange}
            options={jobFamilyGroupOptions}
            error={errors.jobFamilyGroupCode}
            isRequired
            disabled={catalogLoading}
          />
          <SelectField
            label="职种"
            value={state.jobFamilyCode}
            onChange={onJobFamilyChange}
            options={jobFamilyOptions}
            error={errors.jobFamilyCode}
            isRequired
            disabled={catalogLoading || !state.jobFamilyGroupCode}
            helperText={!state.jobFamilyGroupCode ? '请先选择职类' : undefined}
          />
        </Flex>

        <Flex gap={space.m} cs={responsiveRowStyle}>
          <SelectField
            label="职务"
            value={state.jobRoleCode}
            onChange={onJobRoleChange}
            options={jobRoleOptions}
            error={errors.jobRoleCode}
            isRequired
            disabled={catalogLoading || !state.jobFamilyCode}
            helperText={!state.jobFamilyCode ? '请先选择职种' : undefined}
          />
          <SelectField
            label="职级"
            value={state.jobLevelCode}
            onChange={onJobLevelChange}
            options={jobLevelOptions}
            error={errors.jobLevelCode}
            isRequired
            disabled={catalogLoading || !state.jobRoleCode}
            helperText={!state.jobRoleCode ? '请先选择职务' : undefined}
          />
        </Flex>
      </>
    ) : (
      renderTextInputs(state, errors, onChange)
    )}

    <Flex gap={space.m} cs={responsiveRowStyle}>
      <TextInputField
        label="所属组织编码"
        value={state.organizationCode}
        onChange={onChange('organizationCode')}
        placeholder="7 位数字"
        isRequired
        error={errors.organizationCode}
      />
      <TextInputField
        label="汇报职位编码（可选）"
        value={state.reportsToPositionCode}
        onChange={onChange('reportsToPositionCode')}
        placeholder="例如：P1000001"
        error={errors.reportsToPositionCode}
      />
    </Flex>

    <Flex gap={space.m} cs={responsiveRowStyle}>
      <Box flex={1}>
        <SelectField
          label="职位类型"
          value={state.positionType}
          onChange={onChange('positionType') as React.ChangeEventHandler<HTMLSelectElement>}
          options={POSITION_TYPES}
          error={errors.positionType}
          isRequired
        />
      </Box>
      <Box flex={1}>
        <SelectField
          label="雇佣方式"
          value={state.employmentType}
          onChange={onChange('employmentType') as React.ChangeEventHandler<HTMLSelectElement>}
          options={EMPLOYMENT_TYPES}
          error={errors.employmentType}
          isRequired
        />
      </Box>
      <TextInputField
        label="职级等级（可选）"
        value={state.gradeLevel}
        onChange={onChange('gradeLevel')}
        placeholder="例如：L3"
        error={errors.gradeLevel}
      />
    </Flex>

    <Flex gap={space.m} cs={responsiveRowStyle}>
      <TextInputField
        type="number"
        label="编制容量 (FTE)"
        value={state.headcountCapacity}
        onChange={onChange('headcountCapacity')}
        placeholder="例如：1 或 2.5"
        isRequired
        error={errors.headcountCapacity}
      />
      <TextInputField
        type="date"
        label={isVersion ? '版本生效日期' : '生效日期'}
        value={state.effectiveDate}
        onChange={onChange('effectiveDate')}
        isRequired
        error={errors.effectiveDate}
      />
    </Flex>

    <TextAreaField
      label="操作原因"
      value={state.operationReason}
      onChange={onChange('operationReason') as React.ChangeEventHandler<HTMLTextAreaElement>}
      placeholder={isVersion ? '请说明创建新版本的原因' : '请说明此次操作的原因'}
      isRequired
      error={errors.operationReason}
      rows={3}
    />
  </SimpleStack>
)
