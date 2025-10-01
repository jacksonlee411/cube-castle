import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { colors } from '@workday/canvas-kit-react/tokens';
import type { InlineVersionRecord } from './types';

export interface RecordInfoSectionProps {
  originalHistoryData: InlineVersionRecord | null;
}

const RecordInfoSection: React.FC<RecordInfoSectionProps> = ({ originalHistoryData }) => {
  if (!originalHistoryData) {
    return null;
  }

  return (
    <Box marginBottom="l" marginTop="l">
      <Heading size="small" marginBottom="s" color={colors.licorice600}>
        记录信息
      </Heading>
      <Box
        cs={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
          gap: '12px',
        }}
      >
        <Box>
          <Text typeLevel="subtext.small" fontWeight="bold" color={colors.licorice500}>
            记录UUID:
          </Text>
          <Text
            typeLevel="subtext.small"
            marginTop="xs"
            color={colors.licorice700}
            style={{ fontFamily: 'monospace' }}
          >
            {originalHistoryData.recordId}
          </Text>
        </Box>
        <Box>
          <Text typeLevel="subtext.small" fontWeight="bold" color={colors.licorice500}>
            创建时间:
          </Text>
          <Text typeLevel="subtext.small" marginTop="xs" color={colors.licorice700}>
            {new Date(originalHistoryData.createdAt).toLocaleString('zh-CN')}
          </Text>
        </Box>
        <Box>
          <Text typeLevel="subtext.small" fontWeight="bold" color={colors.licorice500}>
            最后更新:
          </Text>
          <Text typeLevel="subtext.small" marginTop="xs" color={colors.licorice700}>
            {new Date(originalHistoryData.updatedAt).toLocaleString('zh-CN')}
          </Text>
        </Box>
      </Box>
    </Box>
  );
};

export default RecordInfoSection;
