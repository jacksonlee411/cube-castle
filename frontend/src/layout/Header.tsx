import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { Heading } from '@workday/canvas-kit-react/text'

export const Header: React.FC = () => {
  return (
    <Box 
      as="header" 
      height={64} 
      width="100vw"
      position="relative"
      cs={{
        backgroundColor: '#FEF7E0', // frenchVanilla100 equivalent
        borderBottom: '1px solid #E6E4E0', // soap500 equivalent
        boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)' // depth.1 equivalent
      }}
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