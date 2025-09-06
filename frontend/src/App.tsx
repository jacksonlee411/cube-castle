import React, { Suspense } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { AppShell } from './layout/AppShell'
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots'
import { Box } from '@workday/canvas-kit-react/layout'
import { Text } from '@workday/canvas-kit-react/text'

// 懒加载关键页面组件以优化初始加载性能
const OrganizationDashboard = React.lazy(() => import('./features/organizations/OrganizationDashboard').then(module => ({ default: module.OrganizationDashboard })))
const OrganizationTemporalPage = React.lazy(() => import('./features/organizations/OrganizationTemporalPage').then(module => ({ default: module.OrganizationTemporalPage })))
const ContractTestingDashboard = React.lazy(() => import('./features/contract-testing/ContractTestingDashboard').then(module => ({ default: module.ContractTestingDashboard })))
const MonitoringDashboard = React.lazy(() => import('./features/monitoring/MonitoringDashboard').then(module => ({ default: module.MonitoringDashboard })))

// 优化的加载组件
const SuspenseLoader: React.FC = () => (
  <Box 
    height="400px"
    style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      flexDirection: 'column',
      gap: '16px'
    }}
  >
    <LoadingDots />
    <Text color="licorice600">加载中...</Text>
  </Box>
)

function App() {
  return (
    <Routes>
      <Route path="/" element={<AppShell />}>
        {/* 默认重定向到组织架构管理页面 */}
        <Route index element={<Navigate to="/organizations" replace />} />
        
        {/* 组织管理模块 - 使用懒加载优化性能 */}
        <Route 
          path="/organizations" 
          element={
            <Suspense fallback={<SuspenseLoader />}>
              <OrganizationDashboard />
            </Suspense>
          } 
        />
        
        {/* 组织相关页面 - 统一使用参数化路由和懒加载 */}
        <Route 
          path="/organizations/:code" 
          element={
            <Suspense fallback={<SuspenseLoader />}>
              <OrganizationTemporalPage />
            </Suspense>
          } 
        />
        <Route 
          path="/organizations/:code/temporal" 
          element={
            <Suspense fallback={<SuspenseLoader />}>
              <OrganizationTemporalPage />
            </Suspense>
          } 
        />
        
        {/* 其他功能模块占位 */}
        <Route path="/dashboard" element={<div>仪表板 - 开发中</div>} />
        
        {/* 系统监控页面 - Prometheus/Grafana监控中心 */}
        <Route 
          path="/monitoring" 
          element={
            <Suspense fallback={<SuspenseLoader />}>
              <MonitoringDashboard />
            </Suspense>
          } 
        />
        
        {/* 契约测试监控页面 - 位于组织架构之后 */}
        <Route 
          path="/contract-testing" 
          element={
            <Suspense fallback={<SuspenseLoader />}>
              <ContractTestingDashboard />
            </Suspense>
          } 
        />
      </Route>
    </Routes>
  )
}

export default App