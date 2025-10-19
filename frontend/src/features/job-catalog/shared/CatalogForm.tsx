import React, { useEffect } from 'react'
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
  children,
}) => {
  const modalModel = useModalModel()

  useEffect(() => {
    if (isOpen) {
      modalModel.events.show()
    } else {
      modalModel.events.hide()
    }
  }, [isOpen, modalModel.events])

  if (modalModel.state.visibility !== 'visible') {
    return null
  }

  const handleClose = () => {
    modalModel.events.hide()
    onClose()
  }

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    onSubmit(event)
  }

  return (
    <Modal model={modalModel} onClose={handleClose} closeOnEscape closeOnOverlayClick>
      <Modal.Overlay>
        <Modal.Card width={width} paddingBottom="s">
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
