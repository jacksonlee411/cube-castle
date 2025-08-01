'use client'

import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { X } from 'lucide-react'
import { Organization, EmployeeStatus } from '@/types'

interface EmployeeFiltersProps {
  filters: {
    search: string
    status: string
    organizationId: string
  }
  organizations: Organization[]
  onFiltersChange: (filters: Partial<EmployeeFiltersProps['filters']>) => void
  onReset: () => void
}

const statusOptions = [
  { value: '', label: '全部状态' },
  { value: EmployeeStatus.ACTIVE, label: '在职' },
  { value: EmployeeStatus.INACTIVE, label: '离职' },
  { value: EmployeeStatus.ON_LEAVE, label: '请假' },
  { value: EmployeeStatus.TERMINATED, label: '终止' },
]

export function EmployeeFilters({
  filters,
  organizations,
  onFiltersChange,
  onReset
}: EmployeeFiltersProps) {
  // 获取当前激活的筛选器数量
  const activeFiltersCount = [
    filters.status,
    filters.organizationId
  ].filter(Boolean).length

  // 获取状态显示文本
  const getStatusLabel = (status: string) => {
    const option = statusOptions.find(opt => opt.value === status)
    return option?.label ?? status
  }

  // 获取组织显示文本
  const getOrganizationLabel = (organizationId: string) => {
    const org = organizations.find(o => o.id === organizationId)
    return org?.name ?? organizationId
  }

  return (
    <div className="space-y-4">
      {/* 筛选器控件 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* 状态筛选 */}
        <div className="space-y-2">
          <label className="text-sm font-medium">员工状态</label>
          <Select 
            value={filters.status} 
            onValueChange={(value) => onFiltersChange({ status: value })}
          >
            <SelectTrigger>
              <SelectValue placeholder="选择状态" />
            </SelectTrigger>
            <SelectContent>
              {statusOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* 部门筛选 */}
        <div className="space-y-2">
          <label className="text-sm font-medium">所属部门</label>
          <Select 
            value={filters.organizationId} 
            onValueChange={(value) => onFiltersChange({ organizationId: value })}
          >
            <SelectTrigger>
              <SelectValue placeholder="选择部门" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">全部部门</SelectItem>
              {organizations.map((org) => (
                <SelectItem key={org.id} value={org.id}>
                  {org.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* 操作按钮 */}
        <div className="flex items-end space-x-2">
          <Button
            variant="outline"
            onClick={onReset}
            disabled={activeFiltersCount === 0}
            className="flex-1"
          >
            重置筛选
            {activeFiltersCount > 0 && (
              <Badge variant="secondary" className="ml-2">
                {activeFiltersCount}
              </Badge>
            )}
          </Button>
        </div>
      </div>

      {/* 激活的筛选器标签 */}
      {activeFiltersCount > 0 && (
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted-foreground">激活筛选器:</span>
          
          {filters.status && (
            <Badge variant="secondary" className="gap-1">
              状态: {getStatusLabel(filters.status)}
              <button
                onClick={() => onFiltersChange({ status: '' })}
                className="ml-1 hover:bg-background rounded-full"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
          
          {filters.organizationId && (
            <Badge variant="secondary" className="gap-1">
              部门: {getOrganizationLabel(filters.organizationId)}
              <button
                onClick={() => onFiltersChange({ organizationId: '' })}
                className="ml-1 hover:bg-background rounded-full"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
        </div>
      )}

      {/* 筛选结果提示 */}
      {activeFiltersCount > 0 && (
        <div className="text-sm text-muted-foreground">
          已应用 {activeFiltersCount} 个筛选条件
        </div>
      )}
    </div>
  )
}