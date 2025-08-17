import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

export const AppShell: React.FC = () => (
  <Box as="div" height="100vh" width="100vw">
    {/* 顶部标题栏 - 占满浏览器整行 */}
    <Header />
    
    {/* 主内容区域 */}
    <Box as="div" display="flex" height="calc(100vh - 64px)">
      {/* 左侧导航 */}
      <Box 
        as="div" 
        width={240} 
        cs={{ borderRight: '1px solid #E5E5E5' }}
      >
        <Sidebar />
      </Box>
      
      {/* 主内容区 */}
      <Box as="div" flex={1} overflow="auto">
        {/* 页面内容区域 */}
        <Box as="div" padding="l">
          <Outlet />
        </Box>
      </Box>
    </Box>
  </Box>
);