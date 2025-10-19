import React from 'react'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { TextArea } from '@workday/canvas-kit-react/text-area'
import { FormField } from '@workday/canvas-kit-react/form-field'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { SimpleStack } from '../SimpleStack'
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
  isRequired?: boolean
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

const SelectField: React.FC<SelectFieldProps> = ({ label, value, onChange, options, error, isRequired }) => (
  <FormField isRequired={isRequired} error={error}>
    <FormField.Label>{label}</FormField.Label>
    <FormField.Field>
      <select
        value={value}
        onChange={onChange}
        style={{ ...SELECT_BASE_STYLE, ...(error ? SELECT_ERROR_STYLE : {}) }}
      >
        {options.map(option => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      {error ? <FormField.Error>{error}</FormField.Error> : null}
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
}

export const PositionFormFields: React.FC<PositionFormFieldsProps> = ({ state, errors, onChange, isVersion }) => (
  <SimpleStack gap={space.xs}>
    <TextInput
      label="职位名称"
      value={state.title}
      onChange={onChange('title')}
      placeholder="请输入职位名称"
      isRequired
      error={Boolean(errors.title)}
      helperText={errors.title}
    />

    <Flex gap={space.m} flexDirection={{ base: 'column', md: 'row' }}>
      <TextInput
        label="职类编码"
        value={state.jobFamilyGroupCode}
        onChange={onChange('jobFamilyGroupCode')}
        placeholder="例如：PROF"
        isRequired
        error={Boolean(errors.jobFamilyGroupCode)}
        helperText={errors.jobFamilyGroupCode}
      />
      <TextInput
        label="职种编码"
        value={state.jobFamilyCode}
        onChange={onChange('jobFamilyCode')}
        placeholder="例如：PROF-IT"
        isRequired
        error={Boolean(errors.jobFamilyCode)}
        helperText={errors.jobFamilyCode}
      />
    </Flex>

    <Flex gap={space.m} flexDirection={{ base: 'column', md: 'row' }}>
      <TextInput
        label="职务编码"
        value={state.jobRoleCode}
        onChange={onChange('jobRoleCode')}
        placeholder="例如：PROF-IT-BKND"
        isRequired
        error={Boolean(errors.jobRoleCode)}
        helperText={errors.jobRoleCode}
      />
      <TextInput
        label="职级编码"
        value={state.jobLevelCode}
        onChange={onChange('jobLevelCode')}
        placeholder="例如：P5"
        isRequired
        error={Boolean(errors.jobLevelCode)}
        helperText={errors.jobLevelCode}
      />
    </Flex>

    <Flex gap={space.m} flexDirection={{ base: 'column', md: 'row' }}>
      <TextInput
        label="所属组织编码"
        value={state.organizationCode}
        onChange={onChange('organizationCode')}
        placeholder="7 位数字"
        isRequired
        error={Boolean(errors.organizationCode)}
        helperText={errors.organizationCode}
      />
      <TextInput
        label="汇报职位编码（可选）"
        value={state.reportsToPositionCode}
        onChange={onChange('reportsToPositionCode')}
        placeholder="例如：P1000001"
        error={Boolean(errors.reportsToPositionCode)}
        helperText={errors.reportsToPositionCode}
      />
    </Flex>

    <Flex gap={space.m} flexDirection={{ base: 'column', md: 'row' }}>
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
      <TextInput
        label="职级等级（可选）"
        value={state.gradeLevel}
        onChange={onChange('gradeLevel')}
        placeholder="例如：L3"
      />
    </Flex>

    <Flex gap={space.m} flexDirection={{ base: 'column', md: 'row' }}>
      <TextInput
        type="number"
        label="编制容量 (FTE)"
        value={state.headcountCapacity}
        onChange={onChange('headcountCapacity')}
        placeholder="例如：1 或 2.5"
        isRequired
        error={Boolean(errors.headcountCapacity)}
        helperText={errors.headcountCapacity}
      />
      <TextInput
        type="date"
        label={isVersion ? '版本生效日期' : '生效日期'}
        value={state.effectiveDate}
        onChange={onChange('effectiveDate')}
        isRequired
        error={Boolean(errors.effectiveDate)}
        helperText={errors.effectiveDate}
      />
    </Flex>

    <TextArea
      label="操作原因"
      value={state.operationReason}
      onChange={onChange('operationReason')}
      placeholder={isVersion ? '请说明创建新版本的原因' : '请说明此次操作的原因'}
      isRequired
      error={Boolean(errors.operationReason)}
      helperText={errors.operationReason}
      rows={3}
    />
  </SimpleStack>
)

