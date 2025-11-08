import React, { useEffect, useRef } from 'react'
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal'
import { Flex } from '@workday/canvas-kit-react/layout'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { Text } from '@workday/canvas-kit-react/text'
import { colors, space } from '@workday/canvas-kit-react/tokens'

export interface CatalogFormProps {
  title: string
  isOpen: boolean
  onClose: () => void
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void
  isSubmitting?: boolean
  submitLabel?: string
  cancelLabel?: string
  errorMessage?: string | null
  width?: number
  cardTestId?: string
  children: React.ReactNode
}

export const CatalogForm: React.FC<CatalogFormProps> = ({
  title,
  isOpen,
  onClose,
  onSubmit,
  isSubmitting = false,
  submitLabel = '保存',
  cancelLabel = '取消',
  errorMessage,
  width = 520,
  cardTestId,
  children,
}) => {
  const modalModel = useModalModel({ initialVisibility: isOpen ? 'visible' : 'hidden' })
  const shouldNotifyCloseRef = useRef(false)

  useEffect(() => {
    if (isOpen) {
      modalModel.events.show()
    } else if (!shouldNotifyCloseRef.current) {
      modalModel.events.hide()
    }
  }, [isOpen, modalModel.events])

  useEffect(() => {
    if (modalModel.state.visibility === 'hidden' && shouldNotifyCloseRef.current) {
      shouldNotifyCloseRef.current = false
      onClose()
    }
  }, [modalModel.state.visibility, onClose])

  if (modalModel.state.visibility !== 'visible') {
    return null
  }

  const handleClose = () => {
    shouldNotifyCloseRef.current = true
    modalModel.events.hide()
  }

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    onSubmit(event)
  }

  return (
    <Modal model={modalModel}>
      <Modal.Overlay>
        <Modal.Card width={width} paddingBottom="s" data-testid={cardTestId}>
          <Modal.CloseIcon aria-label="关闭" onClick={handleClose} />
          <Modal.Heading>{title}</Modal.Heading>
          <form onSubmit={handleSubmit}>
            <Modal.Body>
              <Flex flexDirection="column" gap={space.m}>
                {children}
                {errorMessage && (
                  <Text fontSize="12px" color={colors.cinnamon500}>
                    {errorMessage}
                  </Text>
                )}
              </Flex>
            </Modal.Body>
            <Flex justifyContent="flex-end" gap={space.s} padding={space.m}>
              <SecondaryButton type="button" onClick={handleClose} disabled={isSubmitting}>
                {cancelLabel}
              </SecondaryButton>
              <PrimaryButton type="submit" disabled={isSubmitting}>
                {isSubmitting ? '处理中…' : submitLabel}
              </PrimaryButton>
            </Flex>
          </form>
        </Modal.Card>
      </Modal.Overlay>
    </Modal>
  )
}
