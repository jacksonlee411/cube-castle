import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { PrimaryButton } from '@workday/canvas-kit-react/button'
import { useNavigate, useLocation } from 'react-router-dom'
import { 
  dashboardIcon,
  activityStreamIcon,
  clockIcon,
  userIcon,
  jobInfoIcon,
  homeIcon,
  chartIcon
} from '@workday/canvas-system-icons-web';

const navigationItems = [
  {
    label: '仪表板',
    path: '/dashboard',
    icon: dashboardIcon
  },
  {
    label: 'CRUD测试',
    path: '/test',
    icon: activityStreamIcon
  },
  {
    label: '时态组件测试',
    path: '/temporal-test',
    icon: clockIcon
  },
  {
    label: '员工管理', 
    path: '/employees',
    icon: userIcon
  },
  {
    label: '职位管理',
    path: '/positions',
    icon: jobInfoIcon
  },
  {
    label: '组织架构',
    path: '/organizations',
    icon: homeIcon
  },
  {
    label: '系统监控',
    path: '/monitoring',
    icon: chartIcon
  }
];

export const Sidebar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <Box height="100%" padding="m">
      {/* 导航菜单 */}
      {navigationItems.map((item) => {
        const isActive = location.pathname.startsWith(item.path);
        
        return (
          <Box key={item.path} marginBottom="s" width="100%">
            <PrimaryButton
              variant={isActive ? undefined : "inverse"}
              onClick={() => navigate(item.path)}
              width="100%"
              icon={item.icon}
            >
              {item.label}
            </PrimaryButton>
          </Box>
        );
      })}
    </Box>
  );
};