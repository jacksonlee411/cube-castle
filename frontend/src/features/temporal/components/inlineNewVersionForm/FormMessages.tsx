import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { colors } from '@workday/canvas-kit-react/tokens';
import { checkCircleIcon, exclamationCircleIcon } from '@workday/canvas-system-icons-web';

export interface FormMessagesProps {
  errorMessage: string | null;
  successMessage: string | null;
}

const FormMessages: React.FC<FormMessagesProps> = ({ errorMessage, successMessage }) => {
  if (!errorMessage && !successMessage) {
    return null;
  }

  return (
    <Box marginBottom="l">
      {errorMessage ? (
        <Box
          padding="m"
          backgroundColor={colors.cinnamon100}
          border={`1px solid ${colors.cinnamon600}`}
          borderRadius="4px"
          marginBottom={successMessage ? 's' : undefined}
        >
          <Flex alignItems="center" gap="s">
            <SystemIcon icon={exclamationCircleIcon} color={colors.cinnamon600} size={20} />
            <Text color={colors.cinnamon600} typeLevel="body.small" fontWeight="medium">
              {errorMessage}
            </Text>
          </Flex>
        </Box>
      ) : null}

      {successMessage ? (
        <Box
          padding="m"
          backgroundColor={colors.greenApple100}
          border={`1px solid ${colors.greenApple600}`}
          borderRadius="4px"
        >
          <Flex alignItems="center" gap="s">
            <SystemIcon icon={checkCircleIcon} color={colors.greenApple600} size={20} />
            <Text color={colors.greenApple600} typeLevel="body.small" fontWeight="medium">
              {successMessage}
            </Text>
          </Flex>
        </Box>
      ) : null}
    </Box>
  );
};

export default FormMessages;
