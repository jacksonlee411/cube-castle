import React, { useEffect, useState } from 'react'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { TextArea } from '@workday/canvas-kit-react/text-area'
import { Select } from '@workday/canvas-kit-react/select'
import { Text } from '@workday/canvas-kit-react/text'
import { CatalogForm } from '../shared/CatalogForm'
import { jobCatalogStatusOptions } from '../types'
import type { JobCatalogStatus } from '@/generated/graphql-types'
import type { CreateJobFamilyGroupInput } from '@/shared/hooks/useJobCatalogMutations'

interface JobFamilyGroupFormProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (input: CreateJobFamilyGroupInput) => Promise<void>
  isSubmitting?: boolean
}

const initialFormState: CreateJobFamilyGroupInput = {
  code: '',
  name: '',
  status: 'ACTIVE',
  effectiveDate: '',
  description: '',
}

const validateCode = (value: string) => /^[A-Z]{4,6}$/.test(value)

export const JobFamilyGroupForm: React.FC<JobFamilyGroupFormProps> = ({
  isOpen,
  onClose,
  onSubmit,
  isSubmitting = false,
}) => {
  const [form, setForm] = useState<CreateJobFamilyGroupInput>(initialFormState)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (isOpen) {
      setForm(initialFormState)
      setError(null)
    }
  }, [isOpen])

  const handleChange = <K extends keyof CreateJobFamilyGroupInput>(key: K, value: CreateJobFamilyGroupInput[K]) => {
    setForm(prev => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async () => {
    if (!validateCode(form.code)) {
      setError('职类编码需为 4-6 位大写字母')
      return
    }
    if (!form.name.trim()) {
      setError('请输入职类名称')
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
      name: form.name.trim(),
      description: form.description?.trim() || undefined,
    })
  }

  return (
    <CatalogForm
      title="新增职类"
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
        <Text typeLevel="body.small" marginBottom="xxs">
          职类编码
        </Text>
        <TextInput
          value={form.code}
          onChange={event => handleChange('code', event.target.value.toUpperCase())}
          placeholder="例如：PROF"
          maxLength={6}
        />
        <Text fontSize="12px" color="licorice400">
          仅允许大写字母，长度 4-6 位
        </Text>
      </div>

      <div>
        <Text typeLevel="body.small" marginBottom="xxs">
          职类名称
        </Text>
        <TextInput
          value={form.name}
          onChange={event => handleChange('name', event.target.value)}
          placeholder="例如：专业技术类"
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
        <Select
          value={form.status}
          onChange={event => handleChange('status', event.target.value as JobCatalogStatus)}
        >
          {jobCatalogStatusOptions.map(option => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </Select>
      </div>

      <div>
        <Text typeLevel="body.small" marginBottom="xxs">
          描述
        </Text>
        <TextArea
          rows={3}
          placeholder="可选：维护该职类的说明"
          value={form.description ?? ''}
          onChange={event => handleChange('description', event.target.value)}
        />
      </div>
    </CatalogForm>
  )
}
