'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Employee, Organization, CreateEmployeeRequest } from '@/types'
import { apiClient } from '@/lib/api-client'
import toast from 'react-hot-toast'

// 表单验证架构
const employeeSchema = z.object({
  employeeNumber: z
    .string()
    .min(1, '员工编号不能为空')
    .max(50, '员工编号不能超过50个字符'),
  firstName: z
    .string()
    .min(1, '名字不能为空')
    .max(50, '名字不能超过50个字符'),
  lastName: z
    .string()
    .min(1, '姓氏不能为空')
    .max(50, '姓氏不能超过50个字符'),
  email: z
    .string()
    .min(1, '邮箱不能为空')
    .email('请输入有效的邮箱地址'),
  phoneNumber: z
    .string()
    .optional()
    .refine(
      (val) => !val || /^1[3-9]\d{9}$/.test(val),
      '请输入有效的手机号码'
    ),
  hireDate: z
    .string()
    .min(1, '入职日期不能为空'),
  jobTitle: z
    .string()
    .max(100, '职位不能超过100个字符')
    .optional(),
  organizationId: z
    .string()
    .optional(),
  managerId: z
    .string()
    .optional(),
})

type EmployeeFormData = z.infer<typeof employeeSchema>

interface EmployeeCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onEmployeeCreated: (employee: Employee) => void
  organizations: Organization[]
}

export function EmployeeCreateDialog({
  open,
  onOpenChange,
  onEmployeeCreated,
  organizations
}: EmployeeCreateDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  
  const form = useForm<EmployeeFormData>({
    resolver: zodResolver(employeeSchema),
    defaultValues: {
      employeeNumber: '',
      firstName: '',
      lastName: '',
      email: '',
      phoneNumber: '',
      hireDate: new Date().toISOString().split('T')[0], // 今天的日期
      jobTitle: '',
      organizationId: '',
      managerId: '',
    },
  })

  // 提交表单
  const onSubmit = async (data: EmployeeFormData) => {
    try {
      setIsSubmitting(true)
      
      // 构建请求数据
      const requestData: CreateEmployeeRequest = {
        employeeNumber: data.employeeNumber,
        firstName: data.firstName,
        lastName: data.lastName,
        email: data.email,
        hireDate: data.hireDate,
        ...(data.phoneNumber && { phoneNumber: data.phoneNumber }),
        ...(data.jobTitle && { jobTitle: data.jobTitle }),
        ...(data.organizationId && { organizationId: data.organizationId }),
        ...(data.managerId && { managerId: data.managerId }),
      }
      
      // 调用 API 创建员工
      const newEmployee = await apiClient.employees.createEmployee(requestData)
      
      // 通知父组件
      onEmployeeCreated(newEmployee)
      
      // 重置表单
      form.reset()
      
      // 关闭对话框
      onOpenChange(false)
      
    } catch (error: any) {
      // Failed to create employee - error handled by UI feedback
      
      // 显示具体的错误信息
      if (error.response?.data?.message) {
        toast.error(error.response.data.message)
      } else if (error.message) {
        toast.error(error.message)
      } else {
        toast.error('创建员工失败，请重试')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  // 生成员工编号建议
  const generateEmployeeNumber = () => {
    const now = new Date()
    const year = now.getFullYear().toString().slice(-2)
    const month = (now.getMonth() + 1).toString().padStart(2, '0')
    const random = Math.floor(Math.random() * 1000).toString().padStart(3, '0')
    const suggested = `EMP${year}${month}${random}`
    
    form.setValue('employeeNumber', suggested)
  }

  // 处理对话框关闭
  const handleClose = () => {
    if (!isSubmitting) {
      form.reset()
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>新增员工</DialogTitle>
          <DialogDescription>
            填写员工基本信息，创建新的员工档案。
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
              {/* 员工编号 */}
              <FormField
                control={form.control}
                name="employeeNumber"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>员工编号 *</FormLabel>
                    <div className="flex space-x-2">
                      <FormControl>
                        <Input placeholder="EMP240001" {...field} />
                      </FormControl>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={generateEmployeeNumber}
                      >
                        生成
                      </Button>
                    </div>
                    <FormDescription>
                      员工的唯一标识编号
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* 邮箱 */}
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>邮箱地址 *</FormLabel>
                    <FormControl>
                      <Input 
                        type="email" 
                        placeholder="zhangsan@company.com" 
                        {...field} 
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              {/* 姓氏 */}
              <FormField
                control={form.control}
                name="lastName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>姓氏 *</FormLabel>
                    <FormControl>
                      <Input placeholder="张" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* 名字 */}
              <FormField
                control={form.control}
                name="firstName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>名字 *</FormLabel>
                    <FormControl>
                      <Input placeholder="三" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              {/* 手机号码 */}
              <FormField
                control={form.control}
                name="phoneNumber"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>手机号码</FormLabel>
                    <FormControl>
                      <Input placeholder="13800138000" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* 入职日期 */}
              <FormField
                control={form.control}
                name="hireDate"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>入职日期 *</FormLabel>
                    <FormControl>
                      <Input type="date" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              {/* 职位 */}
              <FormField
                control={form.control}
                name="jobTitle"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>职位</FormLabel>
                    <FormControl>
                      <Input placeholder="软件开发工程师" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* 所属部门 */}
              <FormField
                control={form.control}
                name="organizationId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>所属部门</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="选择部门" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="">无部门</SelectItem>
                        {organizations.map((org) => (
                          <SelectItem key={org.id} value={org.id}>
                            {org.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={handleClose}
                disabled={isSubmitting}
              >
                取消
              </Button>
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? '创建中...' : '创建员工'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}