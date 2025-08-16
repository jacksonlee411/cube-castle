import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { Heading } from '@workday/canvas-kit-react/text'

export const Header: React.FC = () => {
  return (
    <Box 
      as="header" 
      height={64} 
      width="100vw"
      backgroundColor="frenchVanilla100"
      borderBottom="1px solid" 
      borderColor="soap500"
      boxShadow="depth.1"
      position="relative"
    >
      {/* 品牌标识：占满整行 */}
      <Box 
        cs={{
          height: "100%",
          width: "100%",
          display: "flex",
          alignItems: "center",
          paddingX: "l"
        }}
      >
        <Heading size="large" color="blackPepper500" fontWeight="bold" width="100%">
          🏰 Cube Castle
        </Heading>
      </Box>
    </Box>
  );
};