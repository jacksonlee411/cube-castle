import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Text } from '@workday/canvas-kit-react/text';

interface ControlPanelProps {
  lastUpdated?: string;
  loading?: boolean;
  onRefresh?: () => void;
}

export const ControlPanel: React.FC<ControlPanelProps> = ({ 
  lastUpdated, 
  loading = false, 
  onRefresh 
}) => {
  return (
    <Box 
      backgroundColor="neutral.100" 
      padding="m" 
      borderRadius="s"
      marginBottom="l"
    >
      <Flex 
        alignItems="center" 
        justifyContent="space-between"
        flexDirection={{ default: 'column', medium: 'row' }}
        gap="m"
      >
        <Box textAlign={{ default: 'center', medium: 'left' }}>
          <Text variant="subtext" fontWeight="bold" marginBottom="xs">
            🔄 实时监控面板
          </Text>
          {lastUpdated && (
            <Text variant="hint" fontSize={12}>
              最后更新: {lastUpdated}
            </Text>
          )}
        </Box>
        
        <Flex gap="m" alignItems="center" flexWrap="wrap">
          <Text variant="hint" fontSize={12}>
            自动刷新: 30秒
          </Text>
          <PrimaryButton
            size="small"
            onClick={onRefresh}
            disabled={loading}
          >
            {loading ? '刷新中...' : '🔄 手动刷新'}
          </PrimaryButton>
          <SecondaryButton
            size="small"
            onClick={() => {
              // 打开Playwright测试报告
              const reportPath = '/playwright-report/index.html';
              window.open(reportPath, '_blank');
            }}
          >
            📊 测试报告
          </SecondaryButton>
        </Flex>
      </Flex>
    </Box>
  );
};