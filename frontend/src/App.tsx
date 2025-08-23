import { Routes, Route, Navigate } from 'react-router-dom'
import { AppShell } from './layout/AppShell'
import { OrganizationDashboard } from './features/organizations/OrganizationDashboard'
import { OrganizationTemporalPage } from './features/organizations/OrganizationTemporalPage'

function App() {
  return (
    <Routes>
      <Route path="/" element={<AppShell />}>
        {/* 默认重定向到组织架构管理页面 */}
        <Route index element={<Navigate to="/organizations" replace />} />
        
        {/* 组织管理模块 */}
        <Route path="/organizations" element={<OrganizationDashboard />} />
        
        {/* 组织相关页面 - 统一使用参数化路由 */}
        <Route path="/organizations/:code" element={<OrganizationTemporalPage />} />
        <Route path="/organizations/:code/temporal" element={<OrganizationTemporalPage />} />
        
        {/* 其他功能模块占位 */}
        <Route path="/dashboard" element={<div>仪表板 - 开发中</div>} />
      </Route>
    </Routes>
  )
}

export default App