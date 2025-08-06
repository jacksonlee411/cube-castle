import { Routes, Route, Navigate } from 'react-router-dom'
import { AppShell } from './layout/AppShell'
import { OrganizationDashboard } from './features/organizations/OrganizationDashboard'

function App() {
  return (
    <Routes>
      <Route path="/" element={<AppShell />}>
        {/* 默认重定向到组织管理 */}
        <Route index element={<Navigate to="/organizations" replace />} />
        
        {/* 组织管理模块 */}
        <Route path="/organizations" element={<OrganizationDashboard />} />
        
        {/* 其他功能模块占位 */}
        <Route path="/dashboard" element={<div>仪表板 - 开发中</div>} />
        <Route path="/employees" element={<div>员工管理 - 开发中</div>} />
        <Route path="/positions" element={<div>职位管理 - 开发中</div>} />
      </Route>
    </Routes>
  )
}

export default App
