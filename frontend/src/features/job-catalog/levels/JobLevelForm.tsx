import React, { useEffect, useState } from 'react'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { TextArea } from '@workday/canvas-kit-react/text-area'
import { Text } from '@workday/canvas-kit-react/text'
import { CatalogForm } from '../shared/CatalogForm'
import { jobCatalogStatusOptions } from '../types'
import { JobCatalogStatus } from '@/generated/graphql-types'
import type { CreateJobLevelInput } from '@/shared/hooks/useJobCatalogMutations'
import { colors } from '@workday/canvas-kit-react/tokens'

interface JobLevelFormProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (input: CreateJobLevelInput) => Promise<void>
  isSubmitting?: boolean
  roleCode: string
}

const initialState = (roleCode: string): CreateJobLevelInput => ({
  code: '',
  jobRoleCode: roleCode,
  name: '',
  levelRank: 1,
  status: JobCatalogStatus.ACTIVE,
  effectiveDate: '',
  description: '',
})

const validateLevelCode = (value: string) => /^[A-Z][0-9]{1,2}$/.test(value)

const selectStyle: React.CSSProperties = {
  width: '100%',
  padding: '8px 12px',
  borderRadius: 8,
  border: `1px solid ${colors.soap500}`,
  backgroundColor: colors.frenchVanilla100,
  fontSize: '14px',
}

export const JobLevelForm: React.FC<JobLevelFormProps> = ({
  isOpen,
  onClose,
  onSubmit,
  isSubmitting = false,
  roleCode,
}) => {
  const [form, setForm] = useState<CreateJobLevelInput>(initialState(roleCode))
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (isOpen) {
      setForm(initialState(roleCode))
      setError(null)
    }
  }, [isOpen, roleCode])

  const handleChange = <K extends keyof CreateJobLevelInput>(key: K, value: CreateJobLevelInput[K]) => {
    setForm(prev => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async () => {
    if (!validateLevelCode(form.code)) {
      setError('职级编码需为“L1”形式，大写字母加 1-2 位数字')
      return
    }
    if (!form.name.trim()) {
      setError('请输入职级名称')
      return
    }
    if (!form.effectiveDate) {
      setError('请选择生效日期')
      return
    }
    if (!Number.isFinite(form.levelRank) || form.levelRank < 1) {
      setError('请输入合法的等级序号（大于等于 1 的整数）')
      return
    }

    setError(null)
    await onSubmit({
      ...form,
      code: form.code.trim().toUpperCase(),
      jobRoleCode: roleCode,
      name: form.name.trim(),
      description: form.description?.trim() || undefined,
      levelRank: Math.round(form.levelRank),
    })
  }

  return (
    <CatalogForm
      title="新增职级"
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
          归属职务
        </Text>
        <Text fontSize="16px" fontWeight={600}>
          {roleCode}
        </Text>
      </div>

      <div>
        <Text typeLevel="body.small" marginBottom="xxs">
          职级编码
        </Text>
        <TextInput
          value={form.code}
          onChange={event => handleChange('code', event.target.value.toUpperCase())}
          placeholder="例如：P5"
        />
      </div>

      <div>
        <Text typeLevel="body.small" marginBottom="xxs">
          职级名称
        </Text>
        <TextInput
          value={form.name}
          onChange={event => handleChange('name', event.target.value)}
          placeholder="请输入职级名称"
        />
      </div>

      <div>
        <Text typeLevel="body.small" marginBottom="xxs">
          等级序号
        </Text>
        <TextInput
          type="number"
          min={1}
          step={1}
          value={form.levelRank}
          onChange={event => handleChange('levelRank', Number(event.target.value))}
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
