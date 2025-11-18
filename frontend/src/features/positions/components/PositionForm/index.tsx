import React, { useCallback, useMemo, useState } from 'react'
import { Flex } from '@workday/canvas-kit-react/layout'
import { Heading } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { Card } from '@workday/canvas-kit-react/card'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import {
  useCreatePosition,
  useUpdatePosition,
  useCreatePositionVersion,
} from '@/shared/hooks/usePositionMutations'
import { useMessages } from '@/shared/hooks/useMessages'
import { SimpleStack } from '../layout/SimpleStack'
import {
  createInitialState,
  type PositionFormErrors,
  type PositionFormProps,
  type PositionFormState,
} from './types'
import { validatePositionForm } from './validation'
import {
  buildCreatePositionPayload,
  buildUpdatePositionPayload,
  buildCreatePositionVersionPayload,
} from './payload'
import { PositionFormFields } from './FormFields'
import { usePositionCatalogOptions } from '@/shared/hooks/usePositionCatalogOptions'
import type { UpdatePositionRequest, CreatePositionVersionRequest } from '@/shared/types/positions'
import temporalEntitySelectors from '@/shared/testids/temporalEntity'

export const PositionForm: React.FC<PositionFormProps> = ({ mode, position, onCancel, onSuccess }) => {
  const createMutation = useCreatePosition()
  const updateMutation = useUpdatePosition()
  const createVersionMutation = useCreatePositionVersion()
  const { showError, showSuccess } = useMessages()

  const [formState, setFormState] = useState<PositionFormState>(() => createInitialState(mode, position))
  const [errors, setErrors] = useState<PositionFormErrors>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  const isEditing = mode === 'edit'
  const isVersion = mode === 'version'

  const headerTitle = useMemo(() => {
    if (mode === 'create') return '创建职位'
    if (mode === 'version') return '新增时态版本'
    return '编辑职位'
  }, [mode])

  const jobCatalog = usePositionCatalogOptions({
    groupCode: formState.jobFamilyGroupCode,
    familyCode: formState.jobFamilyCode,
    roleCode: formState.jobRoleCode,
    levelCode: formState.jobLevelCode,
  })

  const validateAndSetErrors = useCallback(
    (state: PositionFormState) => {
      const nextErrors = validatePositionForm(state)
      setErrors(nextErrors)
      return Object.keys(nextErrors).length === 0
    },
    [setErrors],
  )

  const handleChange =
    (key: keyof PositionFormState) =>
    (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
      const value = event.target.value
      setFormState(prev => ({
        ...prev,
        [key]: value,
      }))
      setErrors(prev => {
        if (!prev[key]) {
          return prev
        }
        const next = { ...prev }
        delete next[key]
        return next
      })
    }

  const handleJobFamilyGroupChange: React.ChangeEventHandler<HTMLSelectElement> = event => {
    const value = event.target.value
    setFormState(prev => ({
      ...prev,
      jobFamilyGroupCode: value,
      jobFamilyCode: '',
      jobRoleCode: '',
      jobLevelCode: '',
    }))
    setErrors(prev => {
      const next = { ...prev }
      delete next.jobFamilyGroupCode
      delete next.jobFamilyCode
      delete next.jobRoleCode
      delete next.jobLevelCode
      return next
    })
  }

  const handleJobFamilyChange: React.ChangeEventHandler<HTMLSelectElement> = event => {
    const value = event.target.value
    setFormState(prev => ({
      ...prev,
      jobFamilyCode: value,
      jobRoleCode: '',
      jobLevelCode: '',
    }))
    setErrors(prev => {
      const next = { ...prev }
      delete next.jobFamilyCode
      delete next.jobRoleCode
      delete next.jobLevelCode
      return next
    })
  }

  const handleJobRoleChange: React.ChangeEventHandler<HTMLSelectElement> = event => {
    const value = event.target.value
    setFormState(prev => ({
      ...prev,
      jobRoleCode: value,
      jobLevelCode: '',
    }))
    setErrors(prev => {
      const next = { ...prev }
      delete next.jobRoleCode
      delete next.jobLevelCode
      return next
    })
  }

  const handleJobLevelChange: React.ChangeEventHandler<HTMLSelectElement> = event => {
    const value = event.target.value
    setFormState(prev => ({
      ...prev,
      jobLevelCode: value,
    }))
    setErrors(prev => {
      if (!prev.jobLevelCode) {
        return prev
      }
      const next = { ...prev }
      delete next.jobLevelCode
      return next
    })
  }

  const mutationPending =
    createMutation.isPending || updateMutation.isPending || createVersionMutation.isPending || isSubmitting

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()

    if (!validateAndSetErrors(formState) || mutationPending) {
      return
    }

    if ((isEditing || isVersion) && !position) {
      showError('缺少职位数据，无法提交。')
      return
    }

    const basePayload = buildCreatePositionPayload(formState)

    setIsSubmitting(true)
    try {
      if (mode === 'create') {
        const resource = await createMutation.mutateAsync(basePayload)
        showSuccess('职位创建成功')
        onSuccess?.({ code: resource.code })
      } else if (mode === 'edit') {
        const updatePayload: UpdatePositionRequest = buildUpdatePositionPayload(formState, position!.code)
        await updateMutation.mutateAsync(updatePayload)
        showSuccess('职位更新成功')
        onSuccess?.({ code: position!.code })
      } else {
        const versionPayload: CreatePositionVersionRequest = buildCreatePositionVersionPayload(formState, position!.code)
        const resource = await createVersionMutation.mutateAsync(versionPayload)
        showSuccess('职位时态版本创建成功')
        onSuccess?.({ code: resource.code })
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : '提交失败，请稍后重试'
      showError(message)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Card
      padding={space.l}
      backgroundColor={colors.frenchVanilla100}
      data-testid={temporalEntitySelectors.position.form ? temporalEntitySelectors.position.form(mode) : `temporal-position-form-${mode}`}
    >
      <form onSubmit={handleSubmit}>
        <SimpleStack gap={space.l}>
          <Heading size="small">{headerTitle}</Heading>

          <PositionFormFields
            state={formState}
            errors={errors}
            onChange={handleChange}
            isVersion={isVersion}
            catalogAvailable={!jobCatalog.hasError}
            catalogLoading={jobCatalog.isLoading}
            jobFamilyGroupOptions={jobCatalog.groupOptions}
            jobFamilyOptions={jobCatalog.familyOptions}
            jobRoleOptions={jobCatalog.roleOptions}
            jobLevelOptions={jobCatalog.levelOptions}
            onJobFamilyGroupChange={handleJobFamilyGroupChange}
            onJobFamilyChange={handleJobFamilyChange}
            onJobRoleChange={handleJobRoleChange}
            onJobLevelChange={handleJobLevelChange}
          />

          <Flex justifyContent="flex-end" gap={space.s}>
            {onCancel && (
              <SecondaryButton type="button" onClick={onCancel} disabled={mutationPending}>
                取消
              </SecondaryButton>
            )}
            <PrimaryButton type="submit" disabled={mutationPending}>
              {mode === 'create' ? '创建职位' : mode === 'version' ? '创建版本' : '保存修改'}
            </PrimaryButton>
          </Flex>
        </SimpleStack>
      </form>
    </Card>
  )
}

export default PositionForm
