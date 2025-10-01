import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import type { InlineNewVersionFormMode, InlineVersionRecord } from './types';

export interface FormHeaderProps {
  currentMode: InlineNewVersionFormMode;
  isEditingHistory: boolean;
  organizationCode: string | null;
  originalHistoryData: InlineVersionRecord | null;
  selectedVersion?: InlineVersionRecord | null;
}

const FormHeader: React.FC<FormHeaderProps> = ({
  currentMode,
  isEditingHistory,
  organizationCode,
  originalHistoryData,
  selectedVersion,
}) => {
  const title = React.useMemo(() => {
    if (currentMode === 'create') {
      return '新建组织信息';
    }
    if (currentMode !== 'edit') {
      return '版本记录管理';
    }
    if (isEditingHistory) {
      return originalHistoryData ? '修改版本记录' : '插入新版本记录';
    }
    return '查看版本记录';
  }, [currentMode, isEditingHistory, originalHistoryData]);

  const subtitle = React.useMemo(() => {
    if (currentMode === 'create') {
      return '填写新组织的基本信息，系统将自动分配组织编码';
    }
    if (currentMode !== 'edit') {
      return organizationCode ? `为组织 ${organizationCode} 管理版本记录` : '管理组织版本记录';
    }
    if (isEditingHistory) {
      if (originalHistoryData) {
        return organizationCode
          ? `修改组织 ${organizationCode} 的现有版本记录`
          : '修改现有版本记录';
      }
      return organizationCode
        ? `为组织 ${organizationCode} 插入新的版本记录`
        : '插入新的版本记录';
    }
    return organizationCode
      ? `查看组织 ${organizationCode} 的版本记录信息`
      : '查看版本记录信息';
  }, [currentMode, isEditingHistory, originalHistoryData, organizationCode]);

  return (
    <Flex justifyContent="space-between" alignItems="center" marginBottom="l">
      <Box>
        <Heading size="medium" marginBottom="s">
          {title}
        </Heading>
        <Text typeLevel="subtext.medium" color="hint">
          {subtitle}
        </Text>
      </Box>
      {selectedVersion ? <Box /> : null}
    </Flex>
  );
};

export default FormHeader;
