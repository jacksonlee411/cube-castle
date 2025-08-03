'use client'

import { useState } from 'react'
import { ChevronRight, ChevronDown, Building, Users, MoreHorizontal, Plus, Edit, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Card, CardContent } from '@/components/ui/card'
import { Organization } from '@/types'

interface OrganizationTreeProps {
  organizations: Organization[]
  onUpdate: (id: string, data: any) => void
  onDelete: (id: string) => void
}

interface TreeNodeProps {
  organization: Organization
  onUpdate: (id: string, data: any) => void
  onDelete: (id: string) => void
  onAddChild: (parentId: string) => void
}

const TreeNode = ({ organization, onUpdate, onDelete, onAddChild }: TreeNodeProps) => {
  const [isExpanded, setIsExpanded] = useState(true)
  
  const hasChildren = organization.children && organization.children.length > 0
  
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
  
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ACTIVE':
        return 'bg-green-500'
      case 'INACTIVE':
        return 'bg-gray-500'
      case 'PLANNED':
        return 'bg-blue-500'
      default:
        return 'bg-gray-500'
    }
  }

  return (
    <div className="ml-4">
      <Card className="mb-2 shadow-sm hover:shadow-md transition-shadow">
        <CardContent className="p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              {/* 展开/收起按钮 */}
              {hasChildren && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setIsExpanded(!isExpanded)}
                  className="h-6 w-6 p-0"
                >
                  {isExpanded ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                </Button>
              )}
              {!hasChildren && <div className="w-6" />}
              
              {/* 组织信息 */}
              <div className="flex items-center space-x-3">
                {getTypeIcon(organization.unit_type || 'DEPARTMENT')}
                <div>
                  <div className="flex items-center space-x-2">
                    <span className="font-medium">{organization.name}</span>
                    <Badge variant="outline" className="text-xs">
                      {getTypeLabel(organization.unit_type || 'DEPARTMENT')}
                    </Badge>
                    <div className={`w-2 h-2 rounded-full ${getStatusColor(organization.status || 'ACTIVE')}`} />
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {organization.id.slice(0, 8)}
                    {organization.profile?.managerName && ` · 负责人: ${organization.profile.managerName}`}
                    · {organization.employee_count || 0} 人
                  </div>
                </div>
              </div>
            </div>
            
            {/* 操作菜单 */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onAddChild(organization.id)}>
                  <Plus className="mr-2 h-4 w-4" />
                  添加子部门
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onUpdate(organization.id, organization)}>
                  <Edit className="mr-2 h-4 w-4" />
                  编辑
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => onDelete(organization.id)}
                  className="text-red-600"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardContent>
      </Card>
      
      {/* 子组织 */}
      {hasChildren && isExpanded && (
        <div className="ml-4 border-l-2 border-gray-200 pl-4">
          {organization.children!.map((child) => (
            <TreeNode
              key={child.id}
              organization={child}
              onUpdate={onUpdate}
              onDelete={onDelete}
              onAddChild={onAddChild}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export const OrganizationTree = ({ organizations, onUpdate, onDelete }: OrganizationTreeProps) => {
  const [selectedParent, setSelectedParent] = useState<string | null>(null)
  
  // 构建树结构
  const buildTree = (orgs: Organization[]): Organization[] => {
    const orgMap = new Map<string, Organization>()
    const roots: Organization[] = []
    
    // 创建组织映射
    orgs.forEach(org => {
      orgMap.set(org.id, { ...org, children: [] })
    })
    
    // 构建父子关系
    orgs.forEach(org => {
      const orgNode = orgMap.get(org.id)!
      if (org.parent_unit_id && orgMap.has(org.parent_unit_id)) {
        const parent = orgMap.get(org.parent_unit_id)!
        parent.children!.push(orgNode)
      } else {
        roots.push(orgNode)
      }
    })
    
    return roots
  }
  
  const treeData = buildTree(organizations)
  
  const handleAddChild = (parentId: string) => {
    setSelectedParent(parentId)
    // TODO: Open create dialog and set parent ID
  }
  
  if (treeData.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <Building className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">暂无组织架构</h3>
        <p className="text-gray-500 mb-4">创建您的第一个组织部门开始构建架构</p>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          创建根部门
        </Button>
      </div>
    )
  }
  
  return (
    <div className="space-y-4">
      <div className="text-sm text-muted-foreground mb-4">
        展示组织层级结构，点击箭头展开/收起子部门
      </div>
      
      <div className="space-y-2">
        {treeData.map((org) => (
          <TreeNode
            key={org.id}
            organization={org}
            onUpdate={onUpdate}
            onDelete={onDelete}
            onAddChild={handleAddChild}
          />
        ))}
      </div>
    </div>
  )
}