import React, { useEffect, useState } from 'react'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { TextArea } from '@workday/canvas-kit-react/text-area'
import { Text } from '@workday/canvas-kit-react/text'
import { CatalogForm } from '../shared/CatalogForm'
import { jobCatalogStatusOptions } from '../types'
import { JobCatalogStatus } from '@/generated/graphql-types'
import type { CreateJobRoleInput } from '@/shared/hooks/useJobCatalogMutations'
import { colors } from '@workday/canvas-kit-react/tokens'

interface JobRoleFormProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (input: CreateJobRoleInput) => Promise<void>
  isSubmitting?: boolean
  familyCode: string
}

const initialState = (familyCode: string): CreateJobRoleInput => ({
  code: '',
  jobFamilyCode: familyCode,
  name: '',
  status: JobCatalogStatus.ACTIVE,
  effectiveDate: '',
  description: '',
})

const validateRoleCode = (value: string) => /^[A-Z]{4,6}-[A-Z0-9]{3,6}-[A-Z0-9]{3,6}$/.test(value)

const selectStyle: React.CSSProperties = {
  width: '100%',
  padding: '8px 12px',
  borderRadius: 8,
  border: `1px solid ${colors.soap500}`,
  backgroundColor: colors.frenchVanilla100,
  fontSize: '14px',
}

export const JobRoleForm: React.FC<JobRoleFormProps> = ({
  isOpen,
  onClose,
  onSubmit,
  isSubmitting = false,
  familyCode,
}) => {
  const [form, setForm] = useState<CreateJobRoleInput>(initialState(familyCode))
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (isOpen) {
      setForm(initialState(familyCode))
      setError(null)
    }
  }, [familyCode, isOpen])

  const handleChange = <K extends keyof CreateJobRoleInput>(key: K, value: CreateJobRoleInput[K]) => {
    setForm(prev => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async () => {
    if (!validateRoleCode(form.code)) {
      setError('职务编码需为“AAAA-BBBB-CCCC”格式，使用大写字母和数字')
      return
    }
    if (!form.name.trim()) {
      setError('请输入职务名称')
      return
    }
    if (!form.effectiveDate) {
      setError('请选择生效日期')
      return
    }

    setError(null)
    await onSubmit({
      ...form,
      code: form.code.trim().toUpperCase(),
      jobFamilyCode: familyCode,
      name: form.name.trim(),
      description: form.description?.trim() || undefined,
    })
  }

  return (
    <CatalogForm
      title="新增职务"
      isOpen={isOpen}
      onClose={onClose}
      onSubmit={event => {
        event.preventDefault()
        void handleSubmit()
      }}
      isSubmitting={isSubmitting}
      submitLabel="确认创建"
      errorMessage={error}
    >
      <div>
        <Text typeLevel="body.small" color="licorice400">
          归属职种
        </Text>
        <Text fontSize="16px" fontWeight={600}>
          {familyCode}
        </Text>
      </div>

      <div>
        <Text typeLevel="body.small" marginBottom="xxs">
          职务编码
        </Text>
        <TextInput
          value={form.code}
          onChange={event => handleChange('code', event.target.value.toUpperCase())}
          placeholder="例如：PROF-ITOPS-BACK"
        />
      </div>

      <div>
        <Text typeLevel="body.small" marginBottom="xxs">
          职务名称
        </Text>
        <TextInput
          value={form.name}
          onChange={event => handleChange('name', event.target.value)}
          placeholder="请输入职务名称"
        />
      </div>

      <div>
        <Text typeLevel="body.small" marginBottom="xxs">
          生效日期
        </Text>
        <TextInput
          type="date"
          value={form.effectiveDate}
          onChange={event => handleChange('effectiveDate', event.target.value)}
        />
      </div>

      <div>
        <Text typeLevel="body.small" marginBottom="xxs">
          状态
        </Text>
        <select
          value={form.status}
          onChange={event => handleChange('status', event.target.value as JobCatalogStatus)}
          style={selectStyle}
        >
          {jobCatalogStatusOptions.map(option => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      <div>
        <Text typeLevel="body.small" marginBottom="xxs">
          描述
        </Text>
        <TextArea
          rows={3}
          value={form.description ?? ''}
          onChange={event => handleChange('description', event.target.value)}
          placeholder="可选：补充说明"
        />
      </div>
    </CatalogForm>
  )
}
