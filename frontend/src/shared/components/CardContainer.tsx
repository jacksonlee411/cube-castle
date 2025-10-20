import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { borderRadius, colors, space } from '@workday/canvas-kit-react/tokens'

export interface CardContainerProps {
  children: React.ReactNode
}

export const CardContainer: React.FC<CardContainerProps> = ({ children }) => (
  <Box
    padding={space.l}
    borderRadius={borderRadius.l}
    backgroundColor={colors.frenchVanilla100}
    border={`1px solid ${colors.soap400}`}
  >
    {children}
  </Box>
)
