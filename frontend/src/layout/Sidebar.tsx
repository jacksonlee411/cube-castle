import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { space } from '@workday/canvas-kit-react/tokens';
import {
  dashboardIcon,
  homeIcon,
  checkIcon,
  viewTeamIcon,
} from '@workday/canvas-system-icons-web';
import { NavigationItem, type NavigationItemConfig } from './NavigationItem';

const navigationConfig: NavigationItemConfig[] = [
  {
    label: '仪表板',
    path: '/dashboard',
    icon: dashboardIcon,
  },
  {
    label: '组织架构',
    path: '/organizations',
    icon: homeIcon,
    permission: 'org:read',
  },
  {
    label: '职位管理',
    path: '/positions',
    icon: viewTeamIcon,
    permission: 'position:read',
    subItems: [
      { label: '职位列表', path: '/positions', permission: 'position:read' },
      { label: '职类管理', path: '/positions/catalog/family-groups', permission: 'job-catalog:read' },
      { label: '职种管理', path: '/positions/catalog/families', permission: 'job-catalog:read' },
      { label: '职务管理', path: '/positions/catalog/roles', permission: 'job-catalog:read' },
      { label: '职级管理', path: '/positions/catalog/levels', permission: 'job-catalog:read' },
    ],
  },
  {
    label: '契约测试',
    path: '/contract-testing',
    icon: checkIcon,
  },
];

export const Sidebar: React.FC = () => (
  <Box
    as="nav"
    aria-label="主导航"
    padding="s"
    cs={{
      display: 'flex',
      flexDirection: 'column',
      gap: space.xxs,
      height: '100%',
    }}
  >
    {navigationConfig.map(item => (
      <NavigationItem key={item.path} {...item} />
    ))}
  </Box>
);
