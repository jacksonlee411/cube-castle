import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { PrimaryButton } from '@workday/canvas-kit-react/button'
import { useNavigate, useLocation } from 'react-router-dom'

const navigationItems = [
  {
    label: '📊 仪表板',
    path: '/dashboard'
  },
  {
    label: '👤 员工管理', 
    path: '/employees'
  },
  {
    label: '💼 职位管理',
    path: '/positions'
  },
  {
    label: '🏢 组织架构',
    path: '/organizations'
  },
  {
    label: '📈 系统监控',
    path: '/monitoring'
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
            >
              {item.label}
            </PrimaryButton>
          </Box>
        );
      })}
    </Box>
  );
};