'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Building, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

// 验证Schema
const formSchema = z.object({
  name: z.string().min(1, '组织名称不能为空').max(100, '组织名称不能超过100个字符'),
  code: z.string().min(1, '组织编码不能为空').max(50, '组织编码不能超过50个字符'),
  type: z.enum(['company', 'department', 'team'], {
    required_error: '请选择组织类型',
  }),
  parentId: z.string().optional(),
  managerName: z.string().max(50, '负责人姓名不能超过50个字符').optional(),
  description: z.string().max(500, '描述不能超过500个字符').optional(),
  status: z.enum(['active', 'inactive']).default('active'),
})

type FormData = z.infer<typeof formSchema>

interface OrganizationCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: FormData) => Promise<void>
  parentOptions?: Array<{ id: string; name: string; level: number }>
}

export const OrganizationCreateDialog = ({
  open,
  onOpenChange,
  onSubmit,
  parentOptions = []
}: OrganizationCreateDialogProps) => {
  const [isLoading, setIsLoading] = useState(false)
  
  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      code: '',
      type: 'department',
      parentId: '',
      managerName: '',
      description: '',
      status: 'active',
    },
  })
  
  const watchType = form.watch('type')
  
  // 生成组织编码
  const generateCode = () => {
    const name = form.getValues('name')
    const type = form.getValues('type')
    if (!name) return
    
    const prefix = type === 'company' ? 'COM' : type === 'department' ? 'DEPT' : 'TEAM'
    const timestamp = Date.now().toString().slice(-6)
    const pinyin = name.toLowerCase().replace(/\s+/g, '')
    const code = `${prefix}_${pinyin}_${timestamp}`.toUpperCase()
    
    form.setValue('code', code)
  }
  
  const handleSubmit = async (data: FormData) => {
    try {
      setIsLoading(true)
      await onSubmit(data)
      form.reset()
    } catch (error) {
      console.error('Failed to create organization:', error)
    } finally {
      setIsLoading(false)
    }
  }
  
  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'company':
        return '公司'
      case 'department':
        return '部门'
      case 'team':
        return '小组'
      default:
        return type
    }
  }
  
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Building className="h-5 w-5" />
            创建新组织
          </DialogTitle>
          <DialogDescription>
            填写组织信息以创建新的部门或团队
          </DialogDescription>
        </DialogHeader>
        
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            {/* 组织名称 */}
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>组织名称 *</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="请输入组织名称"
                      {...field}
                      onBlur={() => {
                        field.onBlur()
                        generateCode()
                      }}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            
            {/* 组织编码 */}
            <FormField
              control={form.control}
              name="code"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>组织编码 *</FormLabel>
                  <FormControl>
                    <div className="flex gap-2">
                      <Input placeholder="请输入组织编码" {...field} />
                      <Button
                        type="button"
                        variant="outline"
                        onClick={generateCode}
                        disabled={!form.getValues('name')}
                      >
                        生成
                      </Button>
                    </div>
                  </FormControl>
                  <FormDescription>
                    唯一标识，建议包含类型前缀
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            
            {/* 组织类型 */}
            <FormField
              control={form.control}
              name="type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>组织类型 *</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="选择组织类型" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="company">公司</SelectItem>
                      <SelectItem value="department">部门</SelectItem>
                      <SelectItem value="team">小组</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            
            {/* 上级组织 */}
            {watchType !== 'company' && (
              <FormField
                control={form.control}
                name="parentId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>上级组织</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="选择上级组织" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="">无上级组织</SelectItem>
                        {parentOptions.map((option) => (
                          <SelectItem key={option.id} value={option.id}>
                            {'  '.repeat(option.level)}{option.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            
            {/* 负责人 */}
            <FormField
              control={form.control}
              name="managerName"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>负责人</FormLabel>
                  <FormControl>
                    <Input placeholder="请输入负责人姓名" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            
            {/* 状态 */}
            <FormField
              control={form.control}
              name="status"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>状态</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="active">活跃</SelectItem>
                      <SelectItem value="inactive">停用</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            
            {/* 描述 */}
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>描述</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="请输入组织描述"
                      className="min-h-[80px]"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={isLoading}
              >
                取消
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                创建
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}