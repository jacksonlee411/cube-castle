import React from 'react'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading } from '@workday/canvas-kit-react/text'
import { SystemIcon } from '@workday/canvas-kit-react/icon'
import { cubeIcon } from '@workday/canvas-system-icons-web'

export const Header: React.FC = () => {
  return (
    <Box 
      as="header" 
      height={64} 
      width="100vw"
      position="relative"
      cs={{
        backgroundColor: '#0875e1', // blueberry500 - Workday主题蓝色
        borderBottom: '1px solid #0e5eb8', // blueberry600 equivalent
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
        <Flex cs={{ alignItems: "center", gap: "l", width: "100%" }}>
          <SystemIcon icon={cubeIcon} size={72} color="frenchVanilla100" />
          <Heading size="large" color="frenchVanilla100" fontWeight="bold">
            Cube Castle
          </Heading>
        </Flex>
      </Box>
    </Box>
  );
};