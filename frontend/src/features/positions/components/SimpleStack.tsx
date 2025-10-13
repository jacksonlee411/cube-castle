import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'

interface SimpleStackProps {
  children: React.ReactNode
  gap?: string | number
}

export const SimpleStack: React.FC<SimpleStackProps> = ({ children, gap = '16px' }) => (
  <Box display="flex" flexDirection="column" gap={gap}>
    {children}
  </Box>
)

export default SimpleStack
