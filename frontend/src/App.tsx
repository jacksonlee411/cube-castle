import React, { Suspense } from 'react'
import { Routes, Route, Navigate, Outlet } from 'react-router-dom'
import { AppShell } from './layout/AppShell'
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots'
import { Box } from '@workday/canvas-kit-react/layout'
import { Text } from '@workday/canvas-kit-react/text'
import { RequireAuth } from './shared/auth/RequireAuth'
import LoginPage from './pages/Login'

// 懒加载关键页面组件以优化初始加载性能
const OrganizationDashboard = React.lazy(() => import('./features/organizations/OrganizationDashboard').then(module => ({ default: module.OrganizationDashboard })))
const PositionDashboard = React.lazy(() => import('./features/positions/PositionDashboard').then(module => ({ default: module.PositionDashboard })))
const OrganizationTemporalEntityRoute = React.lazy(() =>
  import('./features/temporal/pages/entityRoutes').then(module => ({
    default: module.OrganizationTemporalEntityRoute,
  })),
)
const PositionTemporalEntityRoute = React.lazy(() =>
  import('./features/temporal/pages/entityRoutes').then(module => ({
    default: module.PositionTemporalEntityRoute,
  })),
)
const ContractTestingDashboard = React.lazy(() => import('./features/contract-testing/ContractTestingDashboard').then(module => ({ default: module.ContractTestingDashboard })))
const MonitoringDashboard = React.lazy(() => import('./features/monitoring/MonitoringDashboard').then(module => ({ default: module.MonitoringDashboard })))
const JobFamilyGroupList = React.lazy(() => import('./features/job-catalog/family-groups/JobFamilyGroupList').then(module => ({ default: module.JobFamilyGroupList })))
const JobFamilyGroupDetail = React.lazy(() => import('./features/job-catalog/family-groups/JobFamilyGroupDetail').then(module => ({ default: module.JobFamilyGroupDetail })))
const JobFamilyList = React.lazy(() => import('./features/job-catalog/families/JobFamilyList').then(module => ({ default: module.JobFamilyList })))
const JobFamilyDetail = React.lazy(() => import('./features/job-catalog/families/JobFamilyDetail').then(module => ({ default: module.JobFamilyDetail })))
const JobRoleList = React.lazy(() => import('./features/job-catalog/roles/JobRoleList').then(module => ({ default: module.JobRoleList })))
const JobRoleDetail = React.lazy(() => import('./features/job-catalog/roles/JobRoleDetail').then(module => ({ default: module.JobRoleDetail })))
const JobLevelList = React.lazy(() => import('./features/job-catalog/levels/JobLevelList').then(module => ({ default: module.JobLevelList })))
const JobLevelDetail = React.lazy(() => import('./features/job-catalog/levels/JobLevelDetail').then(module => ({ default: module.JobLevelDetail })))

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
          element={renderOrganizations(<OrganizationTemporalEntityRoute />)} 
        />
        <Route 
          path="/organizations/:code/temporal" 
          element={renderOrganizations(<OrganizationTemporalEntityRoute />)} 
        />
        
        {/* 职位管理 - 二级导航结构 */}
        <Route path="/positions" element={renderPositions(<Outlet />)}>
          <Route index element={<PositionDashboard />} />
          <Route path=":code" element={<PositionTemporalEntityRoute />} />
          <Route path="catalog">
            <Route path="family-groups" element={<JobFamilyGroupList />} />
            <Route path="family-groups/:code" element={<JobFamilyGroupDetail />} />
            <Route path="families" element={<JobFamilyList />} />
            <Route path="families/:code" element={<JobFamilyDetail />} />
            <Route path="roles" element={<JobRoleList />} />
            <Route path="roles/:code" element={<JobRoleDetail />} />
            <Route path="levels" element={<JobLevelList />} />
            <Route path="levels/:code" element={<JobLevelDetail />} />
          </Route>
        </Route>
        
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
