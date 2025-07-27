'use client'

import { useState, useEffect } from 'react'
import { Plus, Search, Filter, Building, Users, MapPin } from 'lucide-react'
import { AppNav } from '@/components/business/app-nav'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { OrganizationTree } from '@/components/business/organization-tree'
import { OrganizationCreateDialog } from '@/components/business/organization-create-dialog'
import { OrganizationTable } from '@/components/business/organization-table'
import { useOrganizationStore } from '@/store'
import { apiClient } from '@/lib/api-client'

export default function OrganizationsPage() {
  const [searchTerm, setSearchTerm] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>('all')
  const [viewMode, setViewMode] = useState<'tree' | 'table'>('tree')
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  
  const {
    organizations,
    stats,
    setOrganizations,
    setStats,
    addOrganization,
    updateOrganization,
    removeOrganization
  } = useOrganizationStore()

  // 获取组织数据
  useEffect(() => {
    const fetchOrganizations = async () => {
      try {
        setIsLoading(true)
        const [orgsResponse, statsResponse] = await Promise.all([
          apiClient.organizations.getList(),
          apiClient.organizations.getStats()
        ])
        // 处理API返回的数据结构
        setOrganizations(orgsResponse.organizations || [])
        setStats(statsResponse.data || {
          total: 0,
          totalEmployees: 0,
          active: 0,
          inactive: 0
        })
      } catch (error) {
        console.error('Failed to fetch organizations:', error)
        // 设置默认数据防止错误
        setOrganizations([])
        setStats({
          total: 0,
          totalEmployees: 0,
          active: 0,
          inactive: 0
        })
      } finally {
        setIsLoading(false)
      }
    }

    fetchOrganizations()
  }, [setOrganizations, setStats])

  // 过滤组织数据
  const filteredOrganizations = organizations.filter(org => {
    if (!org || typeof org !== 'object') return false
    
    const name = org.name || ''
    const code = org.code || ''
    
    const matchesSearch = name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         code.toLowerCase().includes(searchTerm.toLowerCase())
    
    // 暂时移除type过滤，因为当前类型定义中没有type字段
    return matchesSearch
  })

  // 统计卡片数据
  const statsCards = [
    {
      title: '总部门数',
      value: stats.total || 0,
      icon: Building,
      color: 'bg-blue-500'
    },
    {
      title: '总员工数',
      value: stats.totalEmployees || 0,
      icon: Users,
      color: 'bg-green-500'
    },
    {
      title: '活跃部门',
      value: stats.active || 0,
      icon: MapPin,
      color: 'bg-purple-500'
    }
  ]

  const handleCreateOrganization = async (data: any) => {
    try {
      const response = await apiClient.organizations.createOrganization(data)
      addOrganization(response.data)
      setCreateDialogOpen(false)
    } catch (error) {
      console.error('Failed to create organization:', error)
    }
  }

  const handleUpdateOrganization = async (id: string, data: any) => {
    try {
      const response = await apiClient.organizations.updateOrganization(id, data)
      updateOrganization(id, response.data)
    } catch (error) {
      console.error('Failed to update organization:', error)
    }
  }

  const handleDeleteOrganization = async (id: string) => {
    try {
      await apiClient.organizations.deleteOrganization(id)
      removeOrganization(id)
    } catch (error) {
      console.error('Failed to delete organization:', error)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-sm text-gray-500">加载中...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <AppNav />
      <div className="space-y-6 p-6">
      {/* 页面标题 */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">组织架构</h1>
          <p className="text-muted-foreground">
            管理组织结构和部门层级关系
          </p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          新建部门
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-3">
        {statsCards.map((card, index) => (
          <Card key={index}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {card.title}
              </CardTitle>
              <div className={`${card.color} p-2 rounded-md`}>
                <card.icon className="h-4 w-4 text-white" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{card.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* 搜索和过滤器 */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="搜索部门名称或编码..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-8"
          />
        </div>
        <Select value={typeFilter} onValueChange={setTypeFilter}>
          <SelectTrigger className="w-48">
            <SelectValue placeholder="选择部门类型" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部类型</SelectItem>
            <SelectItem value="company">公司</SelectItem>
            <SelectItem value="department">部门</SelectItem>
            <SelectItem value="team">小组</SelectItem>
          </SelectContent>
        </Select>
        <div className="flex border rounded-md">
          <Button
            variant={viewMode === 'tree' ? 'default' : 'ghost'}
            size="sm"
            onClick={() => setViewMode('tree')}
            className="rounded-r-none"
          >
            树状图
          </Button>
          <Button
            variant={viewMode === 'table' ? 'default' : 'ghost'}
            size="sm"
            onClick={() => setViewMode('table')}
            className="rounded-l-none"
          >
            表格
          </Button>
        </div>
      </div>

      {/* 组织架构视图 */}
      <Card>
        <CardHeader>
          <CardTitle>组织架构图</CardTitle>
          <CardDescription>
            {viewMode === 'tree' ? '树状结构显示组织层级关系' : '表格形式管理组织信息'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {viewMode === 'tree' ? (
            <OrganizationTree
              organizations={filteredOrganizations}
              onUpdate={handleUpdateOrganization}
              onDelete={handleDeleteOrganization}
            />
          ) : (
            <OrganizationTable
              organizations={filteredOrganizations}
              onUpdate={handleUpdateOrganization}
              onDelete={handleDeleteOrganization}
            />
          )}
        </CardContent>
      </Card>

      {/* 创建对话框 */}
      <OrganizationCreateDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSubmit={handleCreateOrganization}
      />
      </div>
    </div>
  )
}