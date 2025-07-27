'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, Search, Filter, Download, Users } from 'lucide-react'
import { AppNav } from '@/components/business/app-nav'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { EmployeeTable } from '@/components/business/employee-table'
import { EmployeeCreateDialog } from '@/components/business/employee-create-dialog'
import { EmployeeFilters } from '@/components/business/employee-filters'
import { useEmployeeStore, useOrganizationStore } from '@/store'
import { apiClient } from '@/lib/api-client'
import { Employee, EmployeeStatus } from '@/types'
import toast from 'react-hot-toast'

export default function EmployeesPage() {
  const router = useRouter()
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [isFiltersOpen, setIsFiltersOpen] = useState(false)
  const [selectedEmployees, setSelectedEmployees] = useState<string[]>([])

  // Store 状态
  const {
    employees,
    loading,
    error,
    filters,
    pagination,
    setEmployees,
    setLoading,
    setError,
    setFilters,
    setPagination
  } = useEmployeeStore()

  const { organizations, setOrganizations } = useOrganizationStore()

  // 加载员工数据
  const loadEmployees = async () => {
    try {
      setLoading(true)
      setError(null)
      
      const response = await apiClient.employees.getEmployees({
        page: pagination.current,
        pageSize: pagination.pageSize,
        search: filters.search || undefined,
        status: filters.status || undefined,
        organizationId: filters.organizationId || undefined
      })
      
      setEmployees(response.employees)
      setPagination({
        total: response.pagination.total
      })
    } catch (err) {
      setError('加载员工数据失败')
      console.error('Failed to load employees:', err)
    } finally {
      setLoading(false)
    }
  }

  // 加载组织数据（用于筛选）
  const loadOrganizations = async () => {
    try {
      const response = await apiClient.organizations.getOrganizations()
      setOrganizations(response.organizations)
    } catch (err) {
      console.error('Failed to load organizations:', err)
    }
  }

  // 初始加载
  useEffect(() => {
    loadEmployees()
    loadOrganizations()
  }, [])

  // 监听筛选条件和分页变化
  useEffect(() => {
    loadEmployees()
  }, [filters, pagination.current, pagination.pageSize])

  // 搜索处理
  const handleSearch = (searchTerm: string) => {
    setFilters({ search: searchTerm })
    setPagination({ current: 1 }) // 重置到第一页
  }

  // 筛选处理
  const handleFiltersChange = (newFilters: Partial<typeof filters>) => {
    setFilters(newFilters)
    setPagination({ current: 1 })
  }

  // 分页处理
  const handlePageChange = (page: number, pageSize?: number) => {
    setPagination({ 
      current: page,
      ...(pageSize && { pageSize })
    })
  }

  // 员工创建成功
  const handleEmployeeCreated = (employee: Employee) => {
    loadEmployees() // 重新加载数据
    setIsCreateDialogOpen(false)
    toast.success('员工创建成功')
  }

  // 导出员工数据
  const handleExport = async () => {
    try {
      // 这里实现导出逻辑
      toast.success('导出功能开发中...')
    } catch (err) {
      toast.error('导出失败')
    }
  }

  // 批量操作
  const handleBulkAction = async (action: string) => {
    if (selectedEmployees.length === 0) {
      toast.error('请选择要操作的员工')
      return
    }

    try {
      switch (action) {
        case 'activate':
          await apiClient.employees.bulkUpdateEmployees(
            selectedEmployees, 
            { status: EmployeeStatus.ACTIVE }
          )
          loadEmployees()
          setSelectedEmployees([])
          break
        case 'deactivate':
          await apiClient.employees.bulkUpdateEmployees(
            selectedEmployees, 
            { status: EmployeeStatus.INACTIVE }
          )
          loadEmployees()
          setSelectedEmployees([])
          break
        default:
          toast.error('未知操作')
      }
    } catch (err) {
      toast.error('批量操作失败')
    }
  }

  // 获取状态统计
  const statusStats = {
    total: employees.length,
    active: employees.filter(emp => emp.status === EmployeeStatus.ACTIVE).length,
    inactive: employees.filter(emp => emp.status === EmployeeStatus.INACTIVE).length,
    onLeave: employees.filter(emp => emp.status === EmployeeStatus.ON_LEAVE).length
  }

  return (
    <div className="min-h-screen bg-background">
      <AppNav />
      <div className="flex-1 space-y-6 p-6">
      {/* 页面标题和操作 */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">员工管理</h1>
          <p className="text-muted-foreground">
            管理公司员工信息、组织关系和状态
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="outline" onClick={handleExport}>
            <Download className="mr-2 h-4 w-4" />
            导出
          </Button>
          <Button onClick={() => setIsCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            新增员工
          </Button>
        </div>
      </div>

      {/* 状态统计卡片 */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总员工数</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{statusStats.total}</div>
            <p className="text-xs text-muted-foreground">
              全部员工
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">在职员工</CardTitle>
            <Badge variant="default" className="h-6 w-6 rounded-full p-0">
              {statusStats.active}
            </Badge>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{statusStats.active}</div>
            <p className="text-xs text-muted-foreground">
              正常在职状态
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">离职员工</CardTitle>
            <Badge variant="secondary" className="h-6 w-6 rounded-full p-0">
              {statusStats.inactive}
            </Badge>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-gray-600">{statusStats.inactive}</div>
            <p className="text-xs text-muted-foreground">
              已离职状态
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">请假员工</CardTitle>
            <Badge variant="outline" className="h-6 w-6 rounded-full p-0">
              {statusStats.onLeave}
            </Badge>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-yellow-600">{statusStats.onLeave}</div>
            <p className="text-xs text-muted-foreground">
              请假状态
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 搜索和筛选 */}
      <div className="flex items-center space-x-2">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="搜索员工姓名、邮箱或员工编号..."
              value={filters.search}
              onChange={(e) => handleSearch(e.target.value)}
              className="pl-10"
            />
          </div>
        </div>
        <Button
          variant="outline"
          onClick={() => setIsFiltersOpen(!isFiltersOpen)}
        >
          <Filter className="mr-2 h-4 w-4" />
          筛选
          {(filters.status || filters.organizationId) && (
            <Badge variant="secondary" className="ml-2">
              已筛选
            </Badge>
          )}
        </Button>
      </div>

      {/* 筛选器面板 */}
      {isFiltersOpen && (
        <Card>
          <CardContent className="pt-6">
            <EmployeeFilters
              filters={filters}
              organizations={organizations}
              onFiltersChange={handleFiltersChange}
              onReset={() => {
                setFilters({ search: '', status: '', organizationId: '' })
                setIsFiltersOpen(false)
              }}
            />
          </CardContent>
        </Card>
      )}

      {/* 批量操作 */}
      {selectedEmployees.length > 0 && (
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div className="text-sm text-muted-foreground">
                已选择 {selectedEmployees.length} 名员工
              </div>
              <div className="flex items-center space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleBulkAction('activate')}
                >
                  批量激活
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleBulkAction('deactivate')}
                >
                  批量停用
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setSelectedEmployees([])}
                >
                  取消选择
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 员工表格 */}
      <Card>
        <CardContent className="p-0">
          <EmployeeTable
            employees={employees}
            loading={loading}
            error={error}
            organizations={organizations}
            pagination={{
              current: pagination.current,
              pageSize: pagination.pageSize,
              total: pagination.total
            }}
            selectedEmployees={selectedEmployees}
            onSelectionChange={setSelectedEmployees}
            onPageChange={handlePageChange}
            onEmployeeClick={(employee) => {
              router.push(`/employees/${employee.id}` as any)
            }}
            onEmployeeEdit={(employee) => {
              router.push(`/employees/${employee.id}/edit` as any)
            }}
            onEmployeeDelete={async (employee) => {
              if (confirm(`确定要删除员工 ${employee.fullName} 吗？`)) {
                try {
                  await apiClient.employees.deleteEmployee(employee.id)
                  loadEmployees()
                } catch (err) {
                  toast.error('删除员工失败')
                }
              }
            }}
          />
        </CardContent>
      </Card>

      {/* 创建员工对话框 */}
      <EmployeeCreateDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onEmployeeCreated={handleEmployeeCreated}
        organizations={organizations}
      />
      </div>
    </div>
  )
}