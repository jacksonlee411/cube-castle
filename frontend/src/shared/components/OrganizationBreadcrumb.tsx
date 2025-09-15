import React from 'react';
import { Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { colors, space } from '@workday/canvas-kit-react/tokens';
import { toBreadcrumbItems } from '../utils/organizationPath';

export interface OrganizationBreadcrumbProps {
  codePath?: string | null;
  namePath?: string | null;
  onNavigate?: (code: string) => void;
  showCodes?: boolean;
  separator?: string; // 默认 "/"
}

export const OrganizationBreadcrumb: React.FC<OrganizationBreadcrumbProps> = ({
  codePath,
  namePath,
  onNavigate,
  showCodes = false,
  separator = '/'
}) => {
  const items = React.useMemo(() => toBreadcrumbItems(codePath, namePath), [codePath, namePath]);

  if (items.length === 0) {
    return null;
  }

  const handleClick = (code?: string) => {
    if (code && onNavigate) onNavigate(code);
  };

  return (
    <Flex alignItems="center" gap={space.xs}>
      {items.map((item, idx) => (
        <Flex key={`${item.code || item.name || idx}-${idx}`} alignItems="center" gap={space.xs}>
          <Text
            as={onNavigate && item.code ? 'button' : 'span'}
            typeLevel="subtext.medium"
            fontWeight={idx === items.length - 1 ? 'medium' : undefined}
            onClick={() => handleClick(item.code)}
            // 简单的可点击样式
            style={{
              cursor: onNavigate && item.code ? 'pointer' : 'default',
              background: 'none',
              border: 'none',
              padding: 0
            }}
          >
            {item.name || item.code || '-'}
            {showCodes && item.code && item.name ? (
              <Text as="span" typeLevel="subtext.small" color={colors.licorice500}>
                {` (${item.code})`}
              </Text>
            ) : null}
          </Text>
          {idx < items.length - 1 && (
            <Text typeLevel="subtext.medium" color="hint">
              {separator}
            </Text>
          )}
        </Flex>
      ))}
    </Flex>
  );
};

export default OrganizationBreadcrumb;
