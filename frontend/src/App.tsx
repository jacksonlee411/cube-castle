import { Routes, Route, Navigate } from 'react-router-dom'
import { AppShell } from './layout/AppShell'
import { OrganizationDashboard } from './features/organizations/OrganizationDashboard'
import { OrganizationTemporalPage } from './features/organizations/OrganizationTemporalPage'
import { MonitoringDashboard } from './features/monitoring/MonitoringDashboard'
import { TemporalVerificationPage } from './features/temporal/TemporalVerificationPage'
import { TemporalManagementDemo } from './features/temporal/TemporalManagementDemo'
import TemporalManagementGraphQL from './features/temporal/TemporalManagementGraphQL'
import TestCrud from './TestCrud'
import MinimalTemporalTest from './MinimalTemporalTest'

function App() {
  return (
    <Routes>
      <Route path="/" element={<AppShell />}>
        {/* 默认重定向到测试页面 */}
        <Route index element={<Navigate to="/test" replace />} />
        
        {/* CRUD测试页面 */}
        <Route path="/test" element={<TestCrud />} />
        
        {/* 时态管理组件测试页面 */}
        <Route path="/temporal-test" element={<MinimalTemporalTest />} />
        
        {/* 时态管理演示页面 */}
        <Route path="/temporal-demo" element={<TemporalManagementDemo />} />
        
        {/* 时态管理GraphQL演示页面 - P3开发成果 */}
        <Route path="/temporal-graphql" element={<TemporalManagementGraphQL />} />
        
        {/* 时态管理验证页面 */}
        <Route path="/temporal-verify" element={<TemporalVerificationPage />} />
        
        {/* 组织管理模块 */}
        <Route path="/organizations" element={<OrganizationDashboard />} />
        
        {/* 组织时态管理页面 */}
        <Route path="/organizations/:code/temporal" element={<OrganizationTemporalPage />} />
        
        {/* 系统监控模块 */}
        <Route path="/monitoring" element={<MonitoringDashboard />} />
        
        {/* 其他功能模块占位 */}
        <Route path="/dashboard" element={<div>仪表板 - 开发中</div>} />
        <Route path="/employees" element={<div>员工管理 - 开发中</div>} />
        <Route path="/positions" element={<div>职位管理 - 开发中</div>} />
      </Route>
    </Routes>
  )
}

export default App