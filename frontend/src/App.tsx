import React, { Suspense } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { AppShell } from './layout/AppShell'
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots'
import { Box } from '@workday/canvas-kit-react/layout'
import { Text } from '@workday/canvas-kit-react/text'
import { RequireAuth } from './shared/auth/RequireAuth'
import LoginPage from './pages/Login'

// 懒加载关键页面组件以优化初始加载性能
const OrganizationDashboard = React.lazy(() => import('./features/organizations/OrganizationDashboard').then(module => ({ default: module.OrganizationDashboard })))
const OrganizationTemporalPage = React.lazy(() => import('./features/organizations/OrganizationTemporalPage').then(module => ({ default: module.OrganizationTemporalPage })))
const PositionDashboard = React.lazy(() => import('./features/positions/PositionDashboard').then(module => ({ default: module.PositionDashboard })))
const PositionTemporalPage = React.lazy(() => import('./features/positions/PositionTemporalPage').then(module => ({ default: module.PositionTemporalPage })))
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
  const isPositionMockMode = import.meta.env.VITE_POSITIONS_MOCK_MODE !== 'false'
  const renderWithAuth = (element: React.ReactNode) =>
    <RequireAuth>
      <Suspense fallback={<SuspenseLoader />}>
        {element}
      </Suspense>
    </RequireAuth>

  const renderWithoutAuth = (element: React.ReactNode) =>
    <Suspense fallback={<SuspenseLoader />}>
      {element}
    </Suspense>

  const renderPositions = (component: React.ReactNode) =>
    isPositionMockMode ? renderWithoutAuth(component) : renderWithAuth(component)

  const renderOrganizations = (component: React.ReactNode) =>
    renderWithAuth(component)

  return (
    <Routes>
      <Route path="/" element={<AppShell />}>
        {/* 默认重定向到组织架构管理页面 */}
        <Route index element={<Navigate to="/organizations" replace />} />
        
        {/* 组织管理模块 - 使用懒加载优化性能 */}
        <Route 
          path="/organizations" 
          element={renderOrganizations(<OrganizationDashboard />)} 
        />
        
        {/* 组织相关页面 - 统一使用参数化路由和懒加载 */}
        <Route 
          path="/organizations/:code" 
          element={renderOrganizations(<OrganizationTemporalPage />)} 
        />
        <Route 
          path="/organizations/:code/temporal" 
          element={renderOrganizations(<OrganizationTemporalPage />)} 
        />
        
        {/* 职位管理 - Stage 0 页面框架 */}
        <Route 
          path="/positions" 
          element={renderPositions(<PositionDashboard />)} 
        />
        <Route
          path="/positions/:code"
          element={renderPositions(<PositionTemporalPage />)}
        />
        <Route
          path="/positions/:code/temporal"
          element={renderPositions(<PositionTemporalPage />)}
        />
        
        {/* 系统监控总览 */}
        <Route 
          path="/dashboard" 
          element={renderWithAuth(<MonitoringDashboard />)} 
        />
        
        {/* 契约测试监控页面 - 位于组织架构之后 */}
        <Route 
          path="/contract-testing" 
          element={renderWithoutAuth(<ContractTestingDashboard />)} 
        />
      </Route>
      {/* 登录页（开发态） */}
      <Route path="/login" element={<LoginPage />} />
    </Routes>
  )
}

export default App
