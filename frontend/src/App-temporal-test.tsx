/**
 * 临时测试页面 - 验证时态管理功能集成
 * 用于测试时态导航栏和组织列表的集成效果
 */
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Box } from '@workday/canvas-kit-react/layout';
import { OrganizationDashboard } from './features/organizations/OrganizationDashboard';

// 创建React Query客户端
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 5 * 60 * 1000, // 5分钟
    },
  },
});

/**
 * 时态管理测试应用
 */
export const TemporalTestApp: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <Box padding="l">
        <OrganizationDashboard />
      </Box>
    </QueryClientProvider>
  );
};

export default TemporalTestApp;