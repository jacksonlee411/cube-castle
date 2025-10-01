import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Heading } from '@workday/canvas-kit-react/text';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { colors } from '@workday/canvas-kit-react/tokens';
import type { InlineNewVersionFormMode, InlineVersionRecord } from './types';

export interface HierarchySectionProps {
  currentMode: InlineNewVersionFormMode;
  selectedVersion?: InlineVersionRecord | null;
  levelDisplay?: number;
  codePathDisplay: string;
  namePathDisplay: string;
}

const HierarchySection: React.FC<HierarchySectionProps> = ({
  currentMode,
  selectedVersion,
  levelDisplay,
  codePathDisplay,
  namePathDisplay,
}) => {
  if (currentMode !== 'edit' || !selectedVersion) {
    return null;
  }

  return (
    <Box marginBottom="l">
      <Heading size="small" marginBottom="s" color={colors.blueberry600}>
        层级与路径
      </Heading>
      <Box marginLeft="m">
        <FormField>
          <FormField.Label>组织层级</FormField.Label>
          <FormField.Field>
            <TextInput value={levelDisplay !== undefined ? String(levelDisplay) : '—'} disabled />
          </FormField.Field>
          <FormField.Hint>层级由后端计算，不可编辑</FormField.Hint>
        </FormField>

        <Box marginTop="m">
          <FormField>
            <FormField.Label>组织路径（编码）</FormField.Label>
            <FormField.Field>
              <TextInput value={codePathDisplay.trim() || '路径数据暂不可用'} disabled />
            </FormField.Field>
            <FormField.Hint>统一 codePath，已与顶部复制按钮联动</FormField.Hint>
          </FormField>
        </Box>

        <Box marginTop="m">
          <FormField>
            <FormField.Label>组织路径（名称）</FormField.Label>
            <FormField.Field>
              <TextInput value={namePathDisplay.trim() || '路径数据暂不可用'} disabled />
            </FormField.Field>
            <FormField.Hint>读取 GraphQL namePath，提供可读的路径描述</FormField.Hint>
          </FormField>
        </Box>
      </Box>
    </Box>
  );
};

export default HierarchySection;
