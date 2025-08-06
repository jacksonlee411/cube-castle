import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

export const AppShell: React.FC = () => (
  <Box height="100vh" width="100vw">
    {/* 顶部标题栏 - 占满浏览器整行 */}
    <Header />
    
    {/* 主内容区域 */}
    <Box display="flex" height="calc(100vh - 64px)">
      {/* 左侧导航 */}
      <Box width={240} borderRight="1px solid" borderColor="neutral.300">
        <Sidebar />
      </Box>
      
      {/* 主内容区 */}
      <Box flex={1} overflow="auto">
        {/* 页面内容区域 */}
        <Box padding="l">
          <Outlet />
        </Box>
      </Box>
    </Box>
  </Box>
);