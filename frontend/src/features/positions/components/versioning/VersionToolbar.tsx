import React from 'react';
import { Card } from '@workday/canvas-kit-react/card';
import { Flex } from '@workday/canvas-kit-react/layout';
import { Switch } from '@workday/canvas-kit-react/switch';
import { SecondaryButton } from '@workday/canvas-kit-react/button';
import { colors, space } from '@workday/canvas-kit-react/tokens';
import { SimpleStack } from '../layout/SimpleStack';

interface PositionVersionToolbarProps {
  includeDeleted: boolean;
  onIncludeDeletedChange: (checked: boolean) => void;
  onExportCsv: () => void;
  isBusy?: boolean;
  hasVersions: boolean;
}

export const PositionVersionToolbar: React.FC<PositionVersionToolbarProps> = ({
  includeDeleted,
  onIncludeDeletedChange,
  onExportCsv,
  isBusy = false,
  hasVersions,
}) => {
  return (
    <Card
      padding={space.l}
      backgroundColor={colors.frenchVanilla100}
      data-testid="position-version-toolbar"
    >
      <SimpleStack gap={space.m}>
        <Flex alignItems="center" gap={space.l} flexWrap="wrap">
          <label
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: space.xs,
              color: colors.licorice500,
              fontSize: '14px',
            }}
          >
            <Switch
              checked={includeDeleted}
              onChange={event => onIncludeDeletedChange(event.target.checked)}
              data-testid="position-version-include-deleted"
            />
            包含已删除版本
          </label>
          <SecondaryButton
            onClick={onExportCsv}
            disabled={!hasVersions || isBusy}
            size="small"
            data-testid="position-version-export-button"
          >
            导出 CSV
          </SecondaryButton>
        </Flex>
      </SimpleStack>
    </Card>
  );
};

export default PositionVersionToolbar;
