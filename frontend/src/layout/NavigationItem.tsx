import React, { useEffect, useMemo } from 'react';
import styled from '@emotion/styled';
import { useLocation, useNavigate } from 'react-router-dom';
import type { CanvasSystemIcon } from '@workday/design-assets-types';
import { Expandable, useExpandableModel } from '@workday/canvas-kit-react/expandable';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { TertiaryButton } from '@workday/canvas-kit-react/button';
import { Flex } from '@workday/canvas-kit-react/layout';
import { borderRadius, colors, space } from '@workday/canvas-kit-react/tokens';
import { useAuth } from '@/shared/auth/hooks';

type SubNavigationItem = {
  label: string;
  path: string;
  permission?: string;
};

export type NavigationItemConfig = {
  label: string;
  path: string;
  icon: CanvasSystemIcon;
  permission?: string;
  subItems?: SubNavigationItem[];
};

const BaseNavigationButton = styled(TertiaryButton, {
  shouldForwardProp: prop => prop !== 'active',
})<{active: boolean}>(
  {
    display: 'flex',
    alignItems: 'center',
    width: '100%',
    border: 0,
    cursor: 'pointer',
    textAlign: 'left',
    gap: space.xs,
    borderRadius: borderRadius.l,
    padding: `${space.xs} ${space.s}`,
    font: 'inherit',
  },
  ({active}) => ({
    background: active ? colors.soap200 : 'transparent',
    color: active ? colors.blueberry400 : colors.licorice500,
  })
);

const StandaloneNavigationButton = styled(BaseNavigationButton)(
  {
    justifyContent: 'flex-start',
  },
  ({active}: {active: boolean}) => ({
    '&:hover': {
      background: active ? colors.soap200 : colors.soap100,
    },
    '&:focus-visible': {
      outline: `2px solid ${colors.blueberry400}`,
      outlineOffset: '2px',
    },
  })
);

const ExpandableTrigger = styled(Expandable.Target, {
  shouldForwardProp: prop => prop !== 'active',
})<{active: boolean}>(
  {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    width: '100%',
    gap: space.xs,
    borderRadius: borderRadius.l,
    padding: `${space.xs} ${space.s}`,
    cursor: 'pointer',
  },
  ({active}) => ({
    background: active ? colors.soap200 : 'transparent',
    color: active ? colors.blueberry400 : colors.licorice500,
  })
);

const SubNavigationButton = styled(BaseNavigationButton)(
  {
    padding: `${space.xxs} ${space.s}`,
    paddingLeft: `calc(${space.s} + 20px + ${space.xs})`,
    gap: space.xxs,
  },
  ({active}: {active: boolean}) => ({
    '&:hover': {
      background: active ? colors.soap200 : colors.soap100,
    },
    '&:focus-visible': {
      outline: `2px solid ${colors.blueberry400}`,
      outlineOffset: '2px',
    },
  })
);

const SubNavigationList = styled('ul')({
  listStyle: 'none',
  margin: space.zero,
  paddingLeft: space.zero,
  display: 'flex',
  flexDirection: 'column',
  gap: space.xxs,
});

const StyledListItem = styled('li')({
  listStyle: 'none',
});

const StyledExpandable = styled(Expandable)({
  width: '100%',
  padding: space.zero,
});

const isPathActive = (current: string, target: string) =>
  current === target || current.startsWith(`${target}/`);

type NavigationGroupProps = {
  label: string;
  path: string;
  icon: CanvasSystemIcon;
  items: SubNavigationItem[];
  currentPath: string;
  onNavigate: (path: string) => void;
};

const NavigationGroup: React.FC<NavigationGroupProps> = ({
  label,
  path,
  icon,
  items,
  currentPath,
  onNavigate,
}) => {
  const sectionActive = isPathActive(currentPath, path) ||
    items.some(item => isPathActive(currentPath, item.path));

  const model = useExpandableModel({
    initialVisibility: sectionActive ? 'visible' : 'hidden',
  });

  const visibility = model.state.visibility;

  useEffect(() => {
    if (sectionActive && visibility !== 'visible') {
      model.events.show();
    }
  }, [sectionActive, visibility, model.events]);

  return (
    <StyledExpandable model={model}>
      <ExpandableTrigger active={sectionActive} headingLevel="h3">
        <Flex cs={{ alignItems: 'center', gap: space.xs, flex: 1, minWidth: 0 }}>
          <SystemIcon icon={icon} size={20} />
          <Expandable.Title>{label}</Expandable.Title>
        </Flex>
        <Expandable.Icon iconPosition="end" />
      </ExpandableTrigger>
      <Expandable.Content>
        <SubNavigationList>
          {items.map(subItem => {
            const active = isPathActive(currentPath, subItem.path);
            return (
              <StyledListItem key={subItem.path}>
                <SubNavigationButton
                  type="button"
                  active={active}
                  aria-current={active ? 'page' : undefined}
                  onClick={() => onNavigate(subItem.path)}
                >
                  {subItem.label}
                </SubNavigationButton>
              </StyledListItem>
            );
          })}
        </SubNavigationList>
      </Expandable.Content>
    </StyledExpandable>
  );
};

export const NavigationItem: React.FC<NavigationItemConfig> = ({
  label,
  path,
  icon,
  permission,
  subItems,
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { hasPermission } = useAuth();

  const canRender = !permission || hasPermission(permission);
  const availableSubItems = useMemo(
    () =>
      (subItems ?? []).filter(item => !item.permission || hasPermission(item.permission)),
    [subItems, hasPermission]
  );

  const hasExplicitSubItems = Boolean(subItems?.length);

  if (!canRender) {
    return null;
  }

  if (hasExplicitSubItems && availableSubItems.length === 0) {
    return null;
  }

  const handleNavigate = (targetPath: string) => {
    if (location.pathname !== targetPath) {
      navigate(targetPath);
    }
  };

  if (availableSubItems.length === 0) {
    const active = isPathActive(location.pathname, path);
    return (
      <StandaloneNavigationButton
        type="button"
        active={active}
        aria-current={active ? 'page' : undefined}
        onClick={() => handleNavigate(path)}
      >
        <SystemIcon icon={icon} size={20} />
        {label}
      </StandaloneNavigationButton>
    );
  }

  return (
    <NavigationGroup
      label={label}
      path={path}
      icon={icon}
      items={availableSubItems}
      currentPath={location.pathname}
      onNavigate={handleNavigate}
    />
  );
};
