import React, { useMemo, useState } from 'react'
import { Flex } from '@workday/canvas-kit-react/layout'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal'
import { Text } from '@workday/canvas-kit-react/text'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { TextArea } from '@workday/canvas-kit-react/text-area'
import { Checkbox } from '@workday/canvas-kit-react/checkbox'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { useTransferPosition } from '@/shared/hooks/usePositionMutations'
import type { PositionRecord } from '@/shared/types/positions'

const getTodayISODate = (): string => {
  const now = new Date()
  const month = `${now.getMonth() + 1}`.padStart(2, '0')
  const day = `${now.getDate()}`.padStart(2, '0')
  return `${now.getFullYear()}-${month}-${day}`
}

interface PositionTransferDialogProps {
  position?: PositionRecord
  disabled?: boolean
}

export const PositionTransferDialog: React.FC<PositionTransferDialogProps> = ({ position, disabled = false }) => {
  const modalModel = useModalModel()
  const {
    mutateAsync: transferAsync,
    isPending,
    error,
  } = useTransferPosition()

  const today = useMemo(getTodayISODate, [])
  const [targetOrganizationCode, setTargetOrganizationCode] = useState('')
  const [effectiveDate, setEffectiveDate] = useState(today)
  const [operationReason, setOperationReason] = useState('')
  const [reassignReports, setReassignReports] = useState(true)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [formError, setFormError] = useState<string | null>(null)

  if (!position) {
    return null
  }

  const resetForm = () => {
    setTargetOrganizationCode('')
    setEffectiveDate(getTodayISODate())
    setOperationReason('')
    setReassignReports(true)
    setFormError(null)
  }

  const openDialog = () => {
    if (disabled) {
      return
    }
    resetForm()
    setFeedback(null)
    modalModel.events.show()
  }

  const closeDialog = () => {
    modalModel.events.hide()
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!targetOrganizationCode || targetOrganizationCode.length !== 7) {
      setFormError('请输入有效的目标组织编码（7位数字）')
      return
    }
    if (!effectiveDate) {
      setFormError('请选择生效日期')
      return
    }
    if (!operationReason || operationReason.trim().length < 4) {
      setFormError('请填写至少 4 个字符的操作原因')
      return
    }

    setFormError(null)
    try {
      await transferAsync({
        code: position.code,
        targetOrganizationCode,
        effectiveDate,
        operationReason: operationReason.trim(),
        reassignReports,
      })
      setFeedback('职位转移成功，相关数据将在数秒内刷新。')
      closeDialog()
    } catch (_error) {
      // 错误将在下方的错误提示区域展示
    }
  }

  return (
    <>
      <Flex gap={space.s} alignItems="center">
        <PrimaryButton onClick={openDialog} disabled={isPending || disabled} data-testid="position-transfer-open">
          发起职位转移
        </PrimaryButton>
        {feedback && (
          <Text fontSize="12px" color={colors.greenApple500}>
            {feedback}
          </Text>
        )}
        {error && (
          <Text fontSize="12px" color={colors.cinnamon500}>
            {(error as Error).message}
          </Text>
        )}
      </Flex>

      {modalModel.state.isVisible && (
        <Modal model={modalModel}>
          <Modal.Overlay>
            <Modal.Card width={520}>
              <Modal.CloseIcon aria-label="关闭" onClick={closeDialog} />
              <Modal.Heading>职位转移</Modal.Heading>
              <form onSubmit={handleSubmit}>
                <Modal.Body>
                  <Flex flexDirection="column" gap="m">
                    <div>
                      <Text fontSize="12px" color={colors.licorice300}>
                        当前职位：{position.title}（{position.code}）
                      </Text>
                      <Text fontSize="12px" color={colors.licorice300}>
                        归属组织：{position.organizationName ?? position.organizationCode}
                      </Text>
                    </div>

                    <div>
                      <Text typeLevel="body.small" marginBottom="xxs">
                        目标组织编码
                      </Text>
                      <TextInput
                        placeholder="例如：1000002"
                        value={targetOrganizationCode}
                        onChange={event => setTargetOrganizationCode(event.target.value.trim())}
                        data-testid="position-transfer-target"
                      />
                    </div>

                    <div>
                      <Text typeLevel="body.small" marginBottom="xxs">
                        生效日期
                      </Text>
                      <TextInput
                        type="date"
                        value={effectiveDate}
                        onChange={event => setEffectiveDate(event.target.value)}
                        data-testid="position-transfer-date"
                      />
                    </div>

                    <div>
                      <Text typeLevel="body.small" marginBottom="xxs">
                        操作原因
                      </Text>
                      <TextArea
                        rows={4}
                        placeholder="请输入此次职位转移的业务原因"
                        value={operationReason}
                        onChange={event => setOperationReason(event.target.value)}
                        data-testid="position-transfer-reason"
                      />
                    </div>

                    <Checkbox
                      checked={reassignReports}
                      onChange={(_, isChecked) => setReassignReports(isChecked)}
                      data-testid="position-transfer-reassign-checkbox"
                    >
                      自动重新关联下级职位的汇报关系
                    </Checkbox>

                    {(formError || error) && (
                      <Text fontSize="12px" color={colors.cinnamon500}>
                        {formError ?? (error as Error)?.message ?? '操作失败，请稍后重试'}
                      </Text>
                    )}
                  </Flex>
                </Modal.Body>
                <Modal.Footer gap="s">
                  <SecondaryButton onClick={closeDialog} disabled={isPending}>
                    取消
                  </SecondaryButton>
                  <PrimaryButton type="submit" disabled={isPending} data-testid="position-transfer-confirm">
                    {isPending ? '处理中...' : '确认转移'}
                  </PrimaryButton>
                </Modal.Footer>
              </form>
            </Modal.Card>
          </Modal.Overlay>
        </Modal>
      )}
    </>
  )
}
