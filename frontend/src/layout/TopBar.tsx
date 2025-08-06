import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { TertiaryButton } from '@workday/canvas-kit-react/button'
import { Heading } from '@workday/canvas-kit-react/text'

export const TopBar: React.FC = () => {
  return (
    <Box height="100%" padding="m">
      <Box height="100%">
        {/* 左侧：页面标题 */}
        <Box marginBottom="s">
          <Heading size="medium">
            组织管理
          </Heading>
        </Box>
        
        {/* 右侧：用户信息和操作 */}
        <Box>
          <TertiaryButton size="small" marginRight="s">
            设置
          </TertiaryButton>
          <TertiaryButton size="small" marginRight="s">
            通知
          </TertiaryButton>
          <TertiaryButton size="small" aria-label="用户头像">
            用户
          </TertiaryButton>
        </Box>
      </Box>
    </Box>
  );
};