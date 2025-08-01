'use client'

import { useState } from 'react'
import { format } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import {
  MoreHorizontal,
  Edit3,
  Trash2,
  Eye,
  Mail,
  Phone,
  Building2,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
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
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Employee, Organization, EmployeeStatus, EmployeeApiResponse } from '@/types'
import { employeeConverter, isValidEmployeeApiResponse } from '@/utils/type-converters'

interface EmployeeTableProps {
  employees: (Employee | EmployeeApiResponse)[] // 支持两种格式
  loading?: boolean
  error?: string | null
  organizations: Organization[]
  pagination: {
    current: number
    pageSize: number
    total: number
  }
  selectedEmployees: string[]
  onSelectionChange: (selectedIds: string[]) => void
  onPageChange: (page: number, pageSize?: number) => void
  onEmployeeClick: (employee: Employee) => void
  onEmployeeEdit: (employee: Employee) => void
  onEmployeeDelete: (employee: Employee) => void
}

export function EmployeeTable({
  employees,
  loading = false,
  error,
  organizations,
  pagination,
  selectedEmployees,
  onSelectionChange,
  onPageChange,
  onEmployeeClick,
  onEmployeeEdit,
  onEmployeeDelete
}: EmployeeTableProps) {
  // 标准化员工数据格式
  const normalizedEmployees: Employee[] = employees.map(emp => {
    if (isValidEmployeeApiResponse(emp)) {
      return employeeConverter.fromApi(emp)
    }
    return emp as Employee
  })

  // 处理全选
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      onSelectionChange(normalizedEmployees.map(emp => emp.id))
    } else {
      onSelectionChange([])
    }
  }

  // 处理单选
  const handleSelectEmployee = (employeeId: string, checked: boolean) => {
    if (checked) {
      onSelectionChange([...selectedEmployees, employeeId])
    } else {
      onSelectionChange(selectedEmployees.filter(id => id !== employeeId))
    }
  }

  // 获取状态显示
  const getStatusBadge = (status: EmployeeStatus) => {
    switch (status) {
      case EmployeeStatus.ACTIVE:
        return <Badge variant="default">在职</Badge>
      case EmployeeStatus.INACTIVE:
        return <Badge variant="secondary">离职</Badge>
      case EmployeeStatus.ON_LEAVE:
        return <Badge variant="outline">请假</Badge>
      case EmployeeStatus.TERMINATED:
        return <Badge variant="destructive">终止</Badge>
      default:
        return <Badge variant="secondary">未知</Badge>
    }
  }

  // 获取组织名称
  const getOrganizationName = (organizationId?: string) => {
    if (!organizationId) return '-'
    const org = organizations.find(o => o.id === organizationId)
    return org?.name ?? organizationId
  }

  // 分页计算
  const totalPages = Math.ceil(pagination.total / pagination.pageSize)
  const startItem = (pagination.current - 1) * pagination.pageSize + 1
  const endItem = Math.min(pagination.current * pagination.pageSize, pagination.total)

  // 生成分页按钮
  const getPageNumbers = () => {
    const pages: (number | string)[] = []
    const { current } = pagination
    
    if (totalPages <= 7) {
      // 显示所有页面
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i)
      }
    } else {
      // 显示省略号的分页逻辑
      if (current <= 4) {
        pages.push(1, 2, 3, 4, 5, '...', totalPages)
      } else if (current >= totalPages - 3) {
        pages.push(1, '...', totalPages - 4, totalPages - 3, totalPages - 2, totalPages - 1, totalPages)
      } else {
        pages.push(1, '...', current - 1, current, current + 1, '...', totalPages)
      }
    }
    
    return pages
  }

  if (loading) {
    return (
      <div className="space-y-4 p-6">
        {/* 加载骨架屏 */}
        {[...Array(5)].map((_, i) => (
          <div key={i} className="flex space-x-4">
            <div className="h-4 w-4 bg-gray-200 rounded animate-pulse" />
            <div className="h-4 flex-1 bg-gray-200 rounded animate-pulse" />
            <div className="h-4 w-24 bg-gray-200 rounded animate-pulse" />
            <div className="h-4 w-20 bg-gray-200 rounded animate-pulse" />
            <div className="h-4 w-16 bg-gray-200 rounded animate-pulse" />
          </div>
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <p className="text-red-500 mb-2">加载失败</p>
          <p className="text-sm text-muted-foreground">{error}</p>
        </div>
      </div>
    )
  }

  if (normalizedEmployees.length === 0) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <Building2 className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-lg font-medium mb-2">暂无员工数据</p>
          <p className="text-sm text-muted-foreground">
            点击"新增员工"按钮开始添加员工信息
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* 表格 */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                <Checkbox
                  checked={
                    normalizedEmployees.length > 0 && 
                    selectedEmployees.length === normalizedEmployees.length
                  }
                  onCheckedChange={handleSelectAll}
                />
              </TableHead>
              <TableHead>员工信息</TableHead>
              <TableHead>联系方式</TableHead>
              <TableHead>部门</TableHead>
              <TableHead>职位</TableHead>
              <TableHead>入职日期</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="w-12">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {normalizedEmployees.map((employee) => (
              <TableRow 
                key={employee.id}
                className="cursor-pointer hover:bg-muted/50"
                onClick={() => onEmployeeClick(employee)}
              >
                <TableCell onClick={(e) => e.stopPropagation()}>
                  <Checkbox
                    checked={selectedEmployees.includes(employee.id)}
                    onCheckedChange={(checked) => 
                      handleSelectEmployee(employee.id, checked as boolean)
                    }
                  />
                </TableCell>
                
                <TableCell>
                  <div className="space-y-1">
                    <div className="font-medium">{employee.fullName}</div>
                    <div className="text-sm text-muted-foreground">
                      #{employee.employeeNumber}
                    </div>
                  </div>
                </TableCell>
                
                <TableCell>
                  <div className="space-y-1">
                    <div className="flex items-center text-sm">
                      <Mail className="mr-1 h-3 w-3" />
                      {employee.email}
                    </div>
                    {employee.phoneNumber && (
                      <div className="flex items-center text-sm text-muted-foreground">
                        <Phone className="mr-1 h-3 w-3" />
                        {employee.phoneNumber}
                      </div>
                    )}
                  </div>
                </TableCell>
                
                <TableCell>
                  <div className="text-sm">
                    {getOrganizationName(employee.organizationId)}
                  </div>
                </TableCell>
                
                <TableCell>
                  <div className="text-sm">
                    {employee.jobTitle ?? '-'}
                  </div>
                </TableCell>
                
                <TableCell>
                  <div className="text-sm">
                    {employee.hireDate ? format(new Date(employee.hireDate), 'yyyy年MM月dd日', { 
                      locale: zhCN 
                    }) : '-'}
                  </div>
                </TableCell>
                
                <TableCell>
                  {getStatusBadge(employee.status)}
                </TableCell>
                
                <TableCell onClick={(e) => e.stopPropagation()}>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" className="h-8 w-8 p-0">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuLabel>操作</DropdownMenuLabel>
                      <DropdownMenuItem
                        onClick={() => onEmployeeClick(employee)}
                      >
                        <Eye className="mr-2 h-4 w-4" />
                        查看详情
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={() => onEmployeeEdit(employee)}
                      >
                        <Edit3 className="mr-2 h-4 w-4" />
                        编辑信息
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        onClick={() => onEmployeeDelete(employee)}
                        className="text-red-600"
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        删除员工
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {/* 分页控件 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between px-2">
          <div className="text-sm text-muted-foreground">
            显示 {startItem} - {endItem} 条，共 {pagination.total} 条记录
          </div>
          
          <div className="flex items-center space-x-2">
            {/* 跳转到第一页 */}
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(1)}
              disabled={pagination.current === 1}
            >
              <ChevronsLeft className="h-4 w-4" />
            </Button>
            
            {/* 上一页 */}
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(pagination.current - 1)}
              disabled={pagination.current === 1}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            
            {/* 页码按钮 */}
            {getPageNumbers().map((page, index) => (
              <Button
                key={index}
                variant={page === pagination.current ? "default" : "outline"}
                size="sm"
                onClick={() => typeof page === 'number' && onPageChange(page)}
                disabled={typeof page === 'string'}
                className={typeof page === 'string' ? 'cursor-default' : ''}
              >
                {page}
              </Button>
            ))}
            
            {/* 下一页 */}
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(pagination.current + 1)}
              disabled={pagination.current === totalPages}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
            
            {/* 跳转到最后一页 */}
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(totalPages)}
              disabled={pagination.current === totalPages}
            >
              <ChevronsRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}