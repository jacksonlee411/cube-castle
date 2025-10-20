import React, { useEffect, useState } from 'react'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { TextArea } from '@workday/canvas-kit-react/text-area'
import { JobCatalogStatus } from '@/generated/graphql-types'
import { CatalogForm } from './CatalogForm'
import { jobCatalogStatusOptions } from '../types'
import { colors } from '@workday/canvas-kit-react/tokens'

export interface CatalogVersionFormValues {
  name: string
  status: JobCatalogStatus
  effectiveDate: string
  description?: string | null
}

interface CatalogVersionFormProps {
  title: string
  isOpen: boolean
  onClose: () => void
  onSubmit: (values: CatalogVersionFormValues) => Promise<void>
  isSubmitting?: boolean
  initialName?: string
  initialDescription?: string | null
  initialStatus?: JobCatalogStatus
  initialEffectiveDate?: string
  submitLabel?: string
}

const initialState: CatalogVersionFormValues = {
  name: '',
  status: JobCatalogStatus.ACTIVE,
  effectiveDate: '',
  description: '',
}

const selectStyle: React.CSSProperties = {
  width: '100%',
  padding: '8px 12px',
  borderRadius: 8,
  border: `1px solid ${colors.soap500}`,
  backgroundColor: colors.frenchVanilla100,
  fontSize: '14px',
}

export const CatalogVersionForm: React.FC<CatalogVersionFormProps> = ({
  title,
  isOpen,
  onClose,
  onSubmit,
  isSubmitting = false,
  initialName,
  initialDescription,
  initialStatus,
  initialEffectiveDate,
  submitLabel,
}) => {
  const [form, setForm] = useState<CatalogVersionFormValues>(initialState)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (isOpen) {
      setForm({
        name: initialName ?? '',
        status: initialStatus ?? JobCatalogStatus.ACTIVE,
        effectiveDate: initialEffectiveDate ?? '',
        description: initialDescription ?? '',
      })
      setError(null)
    }
  }, [initialDescription, initialEffectiveDate, initialName, initialStatus, isOpen])

  const handleChange = <K extends keyof CatalogVersionFormValues>(key: K, value: CatalogVersionFormValues[K]) => {
    setForm(prev => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async () => {
    if (!form.name.trim()) {
      setError('请输入名称')
      return
    }
    if (!form.effectiveDate) {
      setError('请选择生效日期')
      return
    }
    setError(null)
    await onSubmit({
      ...form,
      name: form.name.trim(),
      description: form.description?.trim() || undefined,
    })
  }

  return (
    <CatalogForm
      title={title}
      isOpen={isOpen}
      onClose={onClose}
      onSubmit={event => {
        event.preventDefault()
        void handleSubmit()
      }}
      isSubmitting={isSubmitting}
      submitLabel={submitLabel ?? '提交'}
      errorMessage={error}
    >
      <div>
        <TextInput
          placeholder="版本名称"
          value={form.name}
          onChange={event => handleChange('name', event.target.value)}
        />
      </div>
      <div>
        <TextInput
          type="date"
          value={form.effectiveDate}
          onChange={event => handleChange('effectiveDate', event.target.value)}
        />
      </div>
      <div>
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
        <TextArea
          rows={3}
          placeholder="可选：版本描述"
          value={form.description ?? ''}
          onChange={event => handleChange('description', event.target.value)}
        />
      </div>
    </CatalogForm>
  )
}
