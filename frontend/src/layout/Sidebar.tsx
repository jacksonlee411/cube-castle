import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { PrimaryButton } from '@workday/canvas-kit-react/button'
import { useNavigate, useLocation } from 'react-router-dom'
import { 
  dashboardIcon,
  homeIcon,
  checkIcon,
  activityStreamIcon
} from '@workday/canvas-system-icons-web';

const navigationItems = [
  {
    label: '仪表板',
    path: '/dashboard',
    icon: dashboardIcon
  },
  {
    label: '组织架构',
    path: '/organizations',
    icon: homeIcon
  },
  {
    label: '系统监控',
    path: '/monitoring',
    icon: activityStreamIcon
  },
  {
    label: '契约测试',
    path: '/contract-testing',
    icon: checkIcon
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