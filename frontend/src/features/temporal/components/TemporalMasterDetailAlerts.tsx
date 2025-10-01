import React from "react";
import { Box, Flex } from "@workday/canvas-kit-react/layout";
import { Text } from "@workday/canvas-kit-react/text";
import { SecondaryButton } from "@workday/canvas-kit-react/button";
import { SystemIcon } from "@workday/canvas-kit-react/icon";
import {
  checkCircleIcon,
  exclamationCircleIcon,
} from "@workday/canvas-system-icons-web";
import { colors, borderRadius } from "@workday/canvas-kit-react/tokens";

interface TemporalMasterDetailAlertsProps {
  loadingError: string | null;
  error: string | null;
  successMessage: string | null;
  retryCount: number;
  isLoading: boolean;
  onRetry: () => void;
}

export const TemporalMasterDetailAlerts: React.FC<TemporalMasterDetailAlertsProps> = ({
  loadingError,
  error,
  successMessage,
  retryCount,
  isLoading,
  onRetry,
}) => {
  if (!loadingError && !error && !successMessage) {
    return null;
  }

  return (
    <Box marginBottom="l">
      {(loadingError || error) && (
        <Box
          padding="m"
          backgroundColor={colors.cinnamon100}
          border={`1px solid ${colors.cinnamon600}`}
          borderRadius={borderRadius.m}
          marginBottom="s"
        >
          <Flex alignItems="center" gap="s">
            <SystemIcon
              icon={exclamationCircleIcon}
              color={colors.cinnamon600}
              size="small"
            />
            <Box flex="1">
              <Text
                color={colors.cinnamon600}
                typeLevel="body.small"
                fontWeight="medium"
              >
                {loadingError ? "加载失败" : "操作失败"}
              </Text>
              <Text color={colors.cinnamon600} typeLevel="subtext.small">
                {loadingError || error}
              </Text>
            </Box>
            {loadingError && retryCount < 3 && (
              <SecondaryButton
                size="small"
                onClick={onRetry}
                disabled={isLoading}
              >
                重试 ({retryCount}/3)
              </SecondaryButton>
            )}
          </Flex>
        </Box>
      )}

      {successMessage && (
        <Box
          padding="m"
          backgroundColor={colors.greenApple100}
          border={`1px solid ${colors.greenApple600}`}
          borderRadius={borderRadius.m}
        >
          <Flex alignItems="center" gap="s">
            <SystemIcon
              icon={checkCircleIcon}
              color={colors.greenApple600}
              size="small"
            />
            <Text
              color={colors.greenApple600}
              typeLevel="body.small"
              fontWeight="medium"
            >
              {successMessage}
            </Text>
          </Flex>
        </Box>
      )}
    </Box>
  );
};

export default TemporalMasterDetailAlerts;
