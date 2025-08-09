import { Routes, Route, Navigate } from 'react-router-dom'
import { AppShell } from './layout/AppShell'
import { OrganizationDashboard } from './features/organizations/OrganizationDashboard'
import { MonitoringDashboard } from './features/monitoring/MonitoringDashboard'
import TestCrud from './TestCrud'

function App() {
  return (
    <Routes>
      <Route path="/" element={<AppShell />}>
        {/* 默认重定向到测试页面 */}
        <Route index element={<Navigate to="/test" replace />} />
        
        {/* CRUD测试页面 */}
        <Route path="/test" element={<TestCrud />} />
        
        {/* 组织管理模块 */}
        <Route path="/organizations" element={<OrganizationDashboard />} />
        
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
