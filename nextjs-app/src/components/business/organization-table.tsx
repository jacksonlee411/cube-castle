'use client'

import { useState } from 'react'
import { MoreHorizontal, Edit, Trash2, Building, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Checkbox } from '@/components/ui/checkbox'
import { Organization } from '@/types'

interface OrganizationTableProps {
  organizations: Organization[]
  onUpdate: (id: string, data: any) => void
  onDelete: (id: string) => void
}

export const OrganizationTable = ({ organizations, onUpdate, onDelete }: OrganizationTableProps) => {
  const [selectedItems, setSelectedItems] = useState<string[]>([])
  
  const getTypeIcon = (unitType: string) => {
    switch (unitType) {
      case 'COMPANY':
        return <Building className="h-4 w-4 text-blue-500" />
      case 'DEPARTMENT':
        return <Building className="h-4 w-4 text-green-500" />
      case 'PROJECT_TEAM':
        return <Users className="h-4 w-4 text-purple-500" />
      case 'COST_CENTER':
        return <Building className="h-4 w-4 text-orange-500" />
      default:
        return <Building className="h-4 w-4" />
    }
  }
  
  const getTypeLabel = (unitType: string) => {
    switch (unitType) {
      case 'COMPANY':
        return '公司'
      case 'DEPARTMENT':
        return '部门'
      case 'PROJECT_TEAM':
        return '项目团队'
      case 'COST_CENTER':
        return '成本中心'
      default:
        return unitType
    }
  }
  
  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ACTIVE':
        return <Badge className="bg-green-100 text-green-800">活跃</Badge>
      case 'INACTIVE':
        return <Badge variant="secondary">停用</Badge>
      case 'PLANNED':
        return <Badge className="bg-blue-100 text-blue-800">计划中</Badge>
      default:
        return <Badge variant="secondary">{status}</Badge>
    }
  }
  
  const getParentName = (parentUnitId?: string) => {
    if (!parentUnitId) return '-'
    const parent = organizations.find(org => org.id === parentUnitId)
    return parent ? parent.name : '-'
  }
  
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedItems(organizations.map(org => org.id))
    } else {
      setSelectedItems([])
    }
  }
  
  const handleSelectItem = (id: string, checked: boolean) => {
    if (checked) {
      setSelectedItems([...selectedItems, id])
    } else {
      setSelectedItems(selectedItems.filter(item => item !== id))
    }
  }
  
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('zh-CN')
  }
  
  if (organizations.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <Building className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">暂无组织数据</h3>
        <p className="text-gray-500">当前筛选条件下没有找到组织</p>
      </div>
    )
  }
  
  return (
    <div className="space-y-4">
      {/* 批量操作 */}
      {selectedItems.length > 0 && (
        <div className="flex items-center gap-2 p-3 bg-blue-50 rounded-lg">
          <span className="text-sm text-blue-800">
            已选择 {selectedItems.length} 个组织
          </span>
          <Button size="sm" variant="outline">
            批量编辑
          </Button>
          <Button size="sm" variant="outline" className="text-red-600">
            批量删除
          </Button>
        </div>
      )}
      
      {/* 表格 */}
      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                <Checkbox
                  checked={selectedItems.length === organizations.length}
                  onCheckedChange={handleSelectAll}
                />
              </TableHead>
              <TableHead>组织名称</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>上级组织</TableHead>
              <TableHead>负责人</TableHead>
              <TableHead>员工数</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>创建时间</TableHead>
              <TableHead className="w-12">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {organizations.map((org) => (
              <TableRow key={org.id}>
                <TableCell>
                  <Checkbox
                    checked={selectedItems.includes(org.id)}
                    onCheckedChange={(checked) => handleSelectItem(org.id, checked as boolean)}
                  />
                </TableCell>
                <TableCell>
                  <div className="flex items-center space-x-3">
                    {getTypeIcon(org.unit_type || 'DEPARTMENT')}
                    <div>
                      <div className="font-medium">{org.name}</div>
                      <div className="text-sm text-muted-foreground">{org.id.slice(0, 8)}</div>
                    </div>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant="outline">
                    {getTypeLabel(org.unit_type || 'DEPARTMENT')}
                  </Badge>
                </TableCell>
                <TableCell>{getParentName(org.parent_unit_id)}</TableCell>
                <TableCell>{org.profile?.managerName || '-'}</TableCell>
                <TableCell>
                  <span className="font-medium">{org.employee_count || 0}</span>
                </TableCell>
                <TableCell>{getStatusBadge(org.status || 'ACTIVE')}</TableCell>
                <TableCell className="text-muted-foreground">
                  {formatDate(org.createdAt || new Date().toISOString())}
                </TableCell>
                <TableCell>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => onUpdate(org.id, org)}>
                        <Edit className="mr-2 h-4 w-4" />
                        编辑
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={() => onDelete(org.id)}
                        className="text-red-600"
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        删除
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      
      {/* 表格底部信息 */}
      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>共 {organizations.length} 个组织</span>
        <span>显示 1-{organizations.length} 项</span>
      </div>
    </div>
  )
}