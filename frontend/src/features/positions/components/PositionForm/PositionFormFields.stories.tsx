import type { Meta, StoryObj } from '@storybook/react'
import React, { useState } from 'react'
import { PositionFormFields } from './FormFields'
import type { PositionFormErrors, PositionFormState, SelectOption } from './types'

const initialState: PositionFormState = {
  title: '高级软件工程师',
  jobFamilyGroupCode: 'ENG',
  jobFamilyCode: 'ENG-SWE',
  jobRoleCode: 'ENG-SWE-BE',
  jobLevelCode: 'P4',
  organizationCode: '2000001',
  positionType: 'REGULAR',
  employmentType: 'FULL_TIME',
  gradeLevel: 'L4',
  headcountCapacity: '2',
  reportsToPositionCode: 'P1000001',
  effectiveDate: '2025-01-01',
  operationReason: '年度 HC 调整',
}

const defaultOptions: SelectOption[] = [
  { value: '', label: '请选择' },
  { value: 'ENG', label: '工程 (ENG)' },
  { value: 'OPS', label: '运营 (OPS)' },
]

const meta: Meta<typeof PositionFormFields> = {
  title: 'Positions/PositionForm/Fields',
  component: PositionFormFields,
  decorators: [Story => <div style={{ maxWidth: 720 }}><Story /></div>],
}

export default meta
type Story = StoryObj<typeof PositionFormFields>

const Template: React.FC<{
  catalogAvailable: boolean
}> = ({ catalogAvailable }) => {
  const [state, setState] = useState<PositionFormState>(initialState)
  const [errors, setErrors] = useState<PositionFormErrors>({})

  const handleChange = (key: keyof PositionFormState) =>
    (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
      const value = event.target.value
      setState(prev => ({ ...prev, [key]: value }))
      setErrors(prev => {
        if (!prev[key]) return prev
        const next = { ...prev }
        delete next[key]
        return next
      })
    }

  return (
    <PositionFormFields
      state={state}
      errors={errors}
      onChange={handleChange}
      isVersion={false}
      catalogAvailable={catalogAvailable}
      catalogLoading={false}
      jobFamilyGroupOptions={catalogAvailable ? defaultOptions : [{ value: '', label: '请选择职类' }]}
      jobFamilyOptions={catalogAvailable ? defaultOptions.map(item => ({ ...item, label: item.label.replace('请选择', '请选择职种') })) : [{ value: '', label: '请选择职种' }]}
      jobRoleOptions={catalogAvailable ? defaultOptions.map(item => ({ ...item, label: item.label.replace('请选择', '请选择职务') })) : [{ value: '', label: '请选择职务' }]}
      jobLevelOptions={catalogAvailable ? defaultOptions.map(item => ({ ...item, label: item.label.replace('请选择', '请选择职级') })) : [{ value: '', label: '请选择职级' }]}
      onJobFamilyGroupChange={event => handleChange('jobFamilyGroupCode')(event)}
      onJobFamilyChange={event => handleChange('jobFamilyCode')(event)}
      onJobRoleChange={event => handleChange('jobRoleCode')(event)}
      onJobLevelChange={event => handleChange('jobLevelCode')(event)}
    />
  )
}

export const Default: Story = {
  render: () => <Template catalogAvailable />,
}

export const CatalogUnavailable: Story = {
  render: () => <Template catalogAvailable={false} />,
}
