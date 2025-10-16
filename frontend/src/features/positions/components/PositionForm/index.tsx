import React, { useCallback, useMemo, useState } from 'react'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading } from '@workday/canvas-kit-react/text'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { NativeSelect } from '@workday/canvas-kit-react/select'
import { TextArea } from '@workday/canvas-kit-react/text-area'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { Card } from '@workday/canvas-kit-react/card'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import type { PositionRecord } from '@/shared/types/positions'
import {
  useCreatePosition,
  useUpdatePosition,
  useCreatePositionVersion,
} from '@/shared/hooks/usePositionMutations'
import { useMessages } from '@/shared/hooks/useMessages'
import type {
  CreatePositionRequest,
  UpdatePositionRequest,
  CreatePositionVersionRequest,
} from '@/shared/types/positions'
import { SimpleStack } from '../SimpleStack'

const POSITION_TYPES = [
  { label: '正式职位 (REGULAR)', value: 'REGULAR' },
  { label: '临时职位 (TEMPORARY)', value: 'TEMPORARY' },
  { label: '合同工 (CONTRACTOR)', value: 'CONTRACTOR' },
]

const EMPLOYMENT_TYPES = [
  { label: '全职 (FULL_TIME)', value: 'FULL_TIME' },
  { label: '兼职 (PART_TIME)', value: 'PART_TIME' },
  { label: '实习 (INTERN)', value: 'INTERN' },
]

type PositionFormMode = 'create' | 'edit' | 'version'

export interface PositionFormProps {
  mode: PositionFormMode
  position?: PositionRecord
  onCancel?: () => void
  onSuccess?: (payload: { code: string }) => void
}

interface PositionFormState {
  title: string
  jobFamilyGroupCode: string
  jobFamilyCode: string
  jobRoleCode: string
  jobLevelCode: string
  organizationCode: string
  positionType: string
  employmentType: string
  gradeLevel: string
  headcountCapacity: string
  reportsToPositionCode: string
  effectiveDate: string
  operationReason: string
}

const createInitialState = (mode: PositionFormMode, position?: PositionRecord): PositionFormState => ({
  title: position?.title ?? '',
  jobFamilyGroupCode: position?.jobFamilyGroupCode ?? '',
  jobFamilyCode: position?.jobFamilyCode ?? '',
  jobRoleCode: position?.jobRoleCode ?? '',
  jobLevelCode: position?.jobLevelCode ?? '',
  organizationCode: position?.organizationCode ?? '',
  positionType: position?.positionType ?? 'REGULAR',
  employmentType: position?.employmentType ?? 'FULL_TIME',
  gradeLevel: position?.gradeLevel ?? '',
  headcountCapacity: position ? String(position.headcountCapacity) : '',
  reportsToPositionCode: position?.reportsToPositionCode ?? '',
  effectiveDate: mode === 'version' ? '' : position?.effectiveDate ?? '',
  operationReason: '',
})

const isRequiredFilled = (value: string) => value.trim().length > 0

const parseHeadcount = (value: string) => {
  const parsed = Number.parseFloat(value)
  if (Number.isNaN(parsed) || parsed < 0) {
    return null
  }
  return parsed
}

export const PositionForm: React.FC<PositionFormProps> = ({ mode, position, onCancel, onSuccess }) => {
  const createMutation = useCreatePosition()
  const updateMutation = useUpdatePosition()
  const createVersionMutation = useCreatePositionVersion()
  const { showError, showSuccess } = useMessages()

  const [formState, setFormState] = useState<PositionFormState>(() => createInitialState(mode, position))
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  const isEditing = mode === 'edit'
  const isVersion = mode === 'version'

  const headerTitle = useMemo(() => {
    if (mode === 'create') return '创建职位'
    if (mode === 'version') return '新增时态版本'
    return '编辑职位'
  }, [mode])

  const validate = useCallback(
    (state: PositionFormState): boolean => {
      const nextErrors: Record<string, string> = {}

      const requiredFields: Array<{ key: keyof PositionFormState; message: string }> = [
        { key: 'title', message: '请填写职位名称' },
        { key: 'jobFamilyGroupCode', message: '请填写职类编码' },
        { key: 'jobFamilyCode', message: '请填写职种编码' },
        { key: 'jobRoleCode', message: '请填写职务编码' },
        { key: 'jobLevelCode', message: '请填写职级编码' },
        { key: 'organizationCode', message: '请填写所属组织编码' },
        { key: 'positionType', message: '请选择职位类型' },
        { key: 'employmentType', message: '请选择雇佣方式' },
        { key: 'headcountCapacity', message: '请填写编制容量' },
        { key: 'effectiveDate', message: '请填写生效日期' },
        { key: 'operationReason', message: '请填写操作原因' },
      ]

      requiredFields.forEach(({ key, message }) => {
        if (!isRequiredFilled(state[key])) {
          nextErrors[key] = message
        }
      })

      if (state.organizationCode && !/^[1-9]\d{6}$/.test(state.organizationCode.trim())) {
        nextErrors.organizationCode = '组织编码需为7位数字，且首位不能为0'
      }

      if (state.reportsToPositionCode && !/^P\d{7}$/.test(state.reportsToPositionCode.trim())) {
        nextErrors.reportsToPositionCode = '汇报职位编码需为 P + 7 位数字'
      }

      const headcount = parseHeadcount(state.headcountCapacity)
      if (headcount === null) {
        nextErrors.headcountCapacity = '编制容量需为非负数字'
      }

      setErrors(nextErrors)
      return Object.keys(nextErrors).length === 0
    },
    [],
  )

  const mutationPending =
    createMutation.isPending || updateMutation.isPending || createVersionMutation.isPending || isSubmitting

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

  const buildBasePayload = (state: PositionFormState): CreatePositionRequest => {
    const payload: CreatePositionRequest = {
      title: state.title.trim(),
      jobFamilyGroupCode: state.jobFamilyGroupCode.trim(),
      jobFamilyCode: state.jobFamilyCode.trim(),
      jobRoleCode: state.jobRoleCode.trim(),
      jobLevelCode: state.jobLevelCode.trim(),
      organizationCode: state.organizationCode.trim(),
      positionType: state.positionType,
      employmentType: state.employmentType,
      gradeLevel: state.gradeLevel.trim() || undefined,
      headcountCapacity: parseHeadcount(state.headcountCapacity) ?? 0,
      reportsToPositionCode: state.reportsToPositionCode.trim() || undefined,
      effectiveDate: state.effectiveDate,
      operationReason: state.operationReason.trim(),
    }

    return payload
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()

    if (!validate(formState) || mutationPending) {
      return
    }

    if ((isEditing || isVersion) && !position) {
      showError('缺少职位数据，无法提交。')
      return
    }

    const basePayload = buildBasePayload(formState)

    setIsSubmitting(true)
    try {
      if (mode === 'create') {
        const resource = await createMutation.mutateAsync(basePayload)
        showSuccess('职位创建成功')
        onSuccess?.({ code: resource.code })
      } else if (mode === 'edit') {
        const updatePayload: UpdatePositionRequest = {
          ...basePayload,
          code: position!.code,
        }
        await updateMutation.mutateAsync(updatePayload)
        showSuccess('职位更新成功')
        onSuccess?.({ code: position!.code })
      } else {
        const versionPayload: CreatePositionVersionRequest = {
          ...basePayload,
          code: position!.code,
        }
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
    <Card padding={space.l} backgroundColor={colors.frenchVanilla100} data-testid={`position-form-${mode}`}>
      <form onSubmit={handleSubmit}>
        <SimpleStack gap={space.l}>
          <Heading size="small">{headerTitle}</Heading>

          <SimpleStack gap={space.xs}>
            <TextInput
              label="职位名称"
              value={formState.title}
              onChange={handleChange('title')}
              placeholder="请输入职位名称"
              isRequired
              error={Boolean(errors.title)}
              helperText={errors.title}
            />

            <Flex gap={space.m} flexDirection={{ base: 'column', md: 'row' }}>
              <TextInput
                label="职类编码"
                value={formState.jobFamilyGroupCode}
                onChange={handleChange('jobFamilyGroupCode')}
                placeholder="例如：PROF"
                isRequired
                error={Boolean(errors.jobFamilyGroupCode)}
                helperText={errors.jobFamilyGroupCode}
              />
              <TextInput
                label="职种编码"
                value={formState.jobFamilyCode}
                onChange={handleChange('jobFamilyCode')}
                placeholder="例如：PROF-IT"
                isRequired
                error={Boolean(errors.jobFamilyCode)}
                helperText={errors.jobFamilyCode}
              />
            </Flex>

            <Flex gap={space.m} flexDirection={{ base: 'column', md: 'row' }}>
              <TextInput
                label="职务编码"
                value={formState.jobRoleCode}
                onChange={handleChange('jobRoleCode')}
                placeholder="例如：PROF-IT-BKND"
                isRequired
                error={Boolean(errors.jobRoleCode)}
                helperText={errors.jobRoleCode}
              />
              <TextInput
                label="职级编码"
                value={formState.jobLevelCode}
                onChange={handleChange('jobLevelCode')}
                placeholder="例如：P5"
                isRequired
                error={Boolean(errors.jobLevelCode)}
                helperText={errors.jobLevelCode}
              />
            </Flex>

            <Flex gap={space.m} flexDirection={{ base: 'column', md: 'row' }}>
              <TextInput
                label="所属组织编码"
                value={formState.organizationCode}
                onChange={handleChange('organizationCode')}
                placeholder="7 位数字"
                isRequired
                error={Boolean(errors.organizationCode)}
                helperText={errors.organizationCode}
              />
              <TextInput
                label="汇报职位编码（可选）"
                value={formState.reportsToPositionCode}
                onChange={handleChange('reportsToPositionCode')}
                placeholder="例如：P1000001"
                error={Boolean(errors.reportsToPositionCode)}
                helperText={errors.reportsToPositionCode}
              />
            </Flex>

            <Flex gap={space.m} flexDirection={{ base: 'column', md: 'row' }}>
              <Box flex={1}>
                <NativeSelect
                  label="职位类型"
                  value={formState.positionType}
                  onChange={handleChange('positionType')}
                  isRequired
                >
                  {POSITION_TYPES.map(option => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </NativeSelect>
              </Box>
              <Box flex={1}>
                <NativeSelect
                  label="雇佣方式"
                  value={formState.employmentType}
                  onChange={handleChange('employmentType')}
                  isRequired
                >
                  {EMPLOYMENT_TYPES.map(option => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </NativeSelect>
              </Box>
              <TextInput
                label="职级等级（可选）"
                value={formState.gradeLevel}
                onChange={handleChange('gradeLevel')}
                placeholder="例如：L3"
              />
            </Flex>

            <Flex gap={space.m} flexDirection={{ base: 'column', md: 'row' }}>
              <TextInput
                type="number"
                label="编制容量 (FTE)"
                value={formState.headcountCapacity}
                onChange={handleChange('headcountCapacity')}
                placeholder="例如：1 或 2.5"
                isRequired
                error={Boolean(errors.headcountCapacity)}
                helperText={errors.headcountCapacity}
              />
              <TextInput
                type="date"
                label={isVersion ? '版本生效日期' : '生效日期'}
                value={formState.effectiveDate}
                onChange={handleChange('effectiveDate')}
                isRequired
                error={Boolean(errors.effectiveDate)}
                helperText={errors.effectiveDate}
              />
            </Flex>

            <TextArea
              label="操作原因"
              value={formState.operationReason}
              onChange={handleChange('operationReason')}
              placeholder={isVersion ? '请说明创建新版本的原因' : '请说明此次操作的原因'}
              isRequired
              error={Boolean(errors.operationReason)}
              helperText={errors.operationReason}
              rows={3}
            />
          </SimpleStack>

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
