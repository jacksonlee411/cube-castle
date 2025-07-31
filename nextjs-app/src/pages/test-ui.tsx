// 测试现代化UI组件库兼容性
import React from 'react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { DataTable } from '@/components/ui/data-table';
import { DatePicker } from '@/components/ui/date-picker';
import { 
  User, 
  Edit, 
  Trash2, 
  CheckCircle,
  Info
} from 'lucide-react';
import { toast } from 'sonner';

interface TestData {
  id: string;
  name: string;
  age: number;
  address: string;
  status: 'active' | 'inactive';
}

const ModernUITest: React.FC = () => {
  const [formData, setFormData] = useState({
    name: '',
    date: new Date(),
    status: ''
  });

  const testData: TestData[] = [
    {
      id: '1',
      name: '张三',
      age: 32,
      address: '北京市朝阳区',
      status: 'active'
    },
    {
      id: '2',
      name: '李四',
      age: 28,
      address: '上海市浦东新区',
      status: 'inactive'
    }
  ];

  const columns = [
    {
      accessorKey: 'name',
      header: '姓名',
      cell: ({ row }: any) => (
        <div className="flex items-center gap-2">
          <User className="h-4 w-4" />
          {row.getValue('name')}
        </div>
      )
    },
    {
      accessorKey: 'age',
      header: '年龄'
    },
    {
      accessorKey: 'address',
      header: '地址'
    },
    {
      accessorKey: 'status',
      header: '状态',
      cell: ({ row }: any) => {
        const status = row.getValue('status');
        return (
          <Badge variant={status === 'active' ? 'default' : 'secondary'}>
            {status === 'active' ? '活跃' : '非活跃'}
          </Badge>
        );
      }
    },
    {
      id: 'actions',
      header: '操作',
      cell: () => (
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm">
            <Edit className="h-4 w-4" />
            编辑
          </Button>
          <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700">
            <Trash2 className="h-4 w-4" />
            删除
          </Button>
        </div>
      )
    }
  ];

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    toast.success('表单提交成功', {
      description: `提交的数据：${JSON.stringify(formData)}`,
      icon: <CheckCircle className="h-4 w-4" />
    });
  };

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      <div className="max-w-6xl mx-auto space-y-6">
        <div>
          <h1 className="text-3xl font-bold">现代化UI组件库测试</h1>
          <p className="text-muted-foreground mt-2">
            基于 shadcn/ui + Tailwind CSS 的现代化组件系统
          </p>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            技术栈：Next.js 14.1.4 | shadcn/ui | Tailwind CSS | Radix UI | Lucide React
          </AlertDescription>
        </Alert>

        <Card>
          <CardHeader>
            <CardTitle>基础组件测试</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            
            {/* 按钮测试 */}
            <div>
              <Label className="text-base font-semibold">按钮组件：</Label>
              <div className="flex items-center gap-2 mt-2">
                <Button>
                  <User className="h-4 w-4 mr-2" />
                  主要按钮
                </Button>
                <Button variant="outline">
                  <Edit className="h-4 w-4 mr-2" />
                  默认按钮
                </Button>
                <Button variant="ghost">
                  <Trash2 className="h-4 w-4 mr-2" />
                  幽灵按钮
                </Button>
              </div>
            </div>

            {/* 表单测试 */}
            <div>
              <Label className="text-base font-semibold">表单组件：</Label>
              <form onSubmit={handleSubmit} className="flex items-end gap-4 mt-2">
                <div className="space-y-2">
                  <Label htmlFor="name">姓名</Label>
                  <Input
                    id="name"
                    placeholder="请输入姓名"
                    value={formData.name}
                    onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                    className="w-40"
                  />
                </div>
                
                <div className="space-y-2">
                  <Label>日期</Label>
                  <DatePicker
                    date={formData.date}
                    onDateChange={(date) => setFormData(prev => ({ ...prev, date: date || new Date() }))}
                  />
                </div>
                
                <div className="space-y-2">
                  <Label>状态</Label>
                  <Select value={formData.status} onValueChange={(value) => setFormData(prev => ({ ...prev, status: value }))}>
                    <SelectTrigger className="w-32">
                      <SelectValue placeholder="选择状态" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="active">活跃</SelectItem>
                      <SelectItem value="inactive">非活跃</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                
                <Button type="submit">
                  提交测试
                </Button>
              </form>
            </div>
          </CardContent>
        </Card>

        {/* 表格测试 */}
        <Card>
          <CardHeader>
            <CardTitle>数据表格测试</CardTitle>
          </CardHeader>
          <CardContent>
            <DataTable 
              columns={columns}
              data={testData}
            />
          </CardContent>
        </Card>

        <div className="text-center">
          <div className="inline-flex items-center gap-2 text-green-600 font-medium">
            <CheckCircle className="h-5 w-5" />
            如果您能看到此页面且所有组件正常显示，说明现代化UI组件库已成功集成！
          </div>
        </div>
      </div>
    </div>
  );
};

export default ModernUITest;