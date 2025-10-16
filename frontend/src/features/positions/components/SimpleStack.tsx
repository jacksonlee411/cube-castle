import React from 'react'
import { Flex } from '@workday/canvas-kit-react/layout'

interface SimpleStackProps {
  children: React.ReactNode
  gap?: string | number
}

export const SimpleStack: React.FC<SimpleStackProps> = ({ children, gap = '16px' }) => (
  <Flex flexDirection="column" gap={gap}>
    {children}
  </Flex>
)

export default SimpleStack
