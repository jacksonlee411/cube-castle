/**
 * 统一消息显示组件
 * 基于Canvas Kit v13设计标准
 * 替代alert()调用
 */
import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { colors } from '@workday/canvas-kit-react/tokens';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { checkCircleIcon, exclamationCircleIcon } from '@workday/canvas-system-icons-web';

export interface MessageDisplayProps {
  successMessage?: string | null;
  errorMessage?: string | null;
  onClear?: () => void;
}

export const MessageDisplay: React.FC<MessageDisplayProps> = ({
  successMessage,
  errorMessage,
  onClear
}) => {
  if (!successMessage && !errorMessage) {
    return null;
  }

  return (
    <Box marginBottom="m">
      {/* 成功消息 */}
      {successMessage && (
        <Flex
          alignItems="center"
          gap="s"
          padding="s"
          backgroundColor={colors.greenApple100}
          borderLeft={`4px solid ${colors.greenApple600}`}
          borderRadius="s"
          marginBottom="s"
        >
          <SystemIcon 
            icon={checkCircleIcon} 
            color={colors.greenApple600}
            size="medium"
          />
          <Text 
            typeLevel="body.medium"
            color={colors.greenApple600}
            fontWeight="medium"
          >
            {successMessage}
          </Text>
        </Flex>
      )}

      {/* 错误消息 */}
      {errorMessage && (
        <Flex
          alignItems="center"
          gap="s"
          padding="s"
          backgroundColor={colors.cinnamon100}
          borderLeft={`4px solid ${colors.cinnamon600}`}
          borderRadius="s"
          marginBottom="s"
        >
          <SystemIcon 
            icon={exclamationCircleIcon} 
            color={colors.cinnamon600}
            size="medium"
          />
          <Text 
            typeLevel="body.medium"
            color={colors.cinnamon600}
            fontWeight="medium"
          >
            {errorMessage}
          </Text>
        </Flex>
      )}
    </Box>
  );
};

export default MessageDisplay;