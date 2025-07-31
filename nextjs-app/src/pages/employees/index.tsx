import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { format } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { 
  Plus, 
  Search, 
  MoreHorizontal,
  User,
  Mail,
  Phone,
  Calendar,
  Users,
  History,
  Edit2,
  Trash2
} from 'lucide-react';
import { toast } from 'react-hot-toast';
import { ColumnDef } from '@tanstack/react-table';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { DatePicker } from '@/components/ui/date-picker';
import { DataTable, createSortableColumn, createActionsColumn } from '@/components/ui/data-table';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface Employee {
  id: string;
  employeeId: string;
  legalName: string;
  preferredName?: string;
  email: string;
  status: string;
  hireDate: string;
  department?: string;
  position?: string;
  managerId?: string;
  managerName?: string;
  avatar?: string;
}

const EmployeesPage: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [formData, setFormData] = useState<Partial<Employee>>({});

  // Sample data
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      const sampleEmployees: Employee[] = [
        {
          id: '1',
          employeeId: 'EMP001',
          legalName: '张三',
          preferredName: 'Zhang San',
          email: 'zhangsan@company.com',
          status: 'ACTIVE',
          hireDate: '2023-01-15',
          department: '技术部',
          position: '高级软件工程师',
          managerName: '李四'
        },
        {
          id: '2',
          employeeId: 'EMP002',
          legalName: '王五',
          email: 'wangwu@company.com',
          status: 'ACTIVE',
          hireDate: '2023-03-20',
          department: '产品部',
          position: '产品经理',
          managerName: '赵六'
        },
        {
          id: '3',
          employeeId: 'EMP003',
          legalName: '刘七',
          email: 'liuqi@company.com',
          status: 'INACTIVE',
          hireDate: '2022-08-10',
          department: '技术部',
          position: '前端工程师',
          managerName: '李四'
        },
        {
          id: '4',
          employeeId: 'EMP004',
          legalName: '陈八',
          email: 'chenba@company.com',
          status: 'ACTIVE',
          hireDate: '2024-01-08',
          department: '人事部',
          position: 'HR专员',
          managerName: '周九'
        }
      ];
      
      setEmployees(sampleEmployees);
      setLoading(false);
    }, 1000);
  }, []);

  const handleCreateEmployee = async (values: any) => {
    try {
      setLoading(true);
      
      if (editingEmployee) {
        // Update existing employee
        const updatedEmployee: Employee = {
          ...editingEmployee,
          employeeId: values.employeeId,
          legalName: values.legalName,
          preferredName: values.preferredName,
          email: values.email,
          hireDate: values.hireDate ? format(new Date(values.hireDate), 'yyyy-MM-dd') : '',
          department: values.department,
          position: values.position,
          managerName: values.managerName
        };

        setEmployees(prev => prev.map(emp => 
          emp.id === editingEmployee.id ? updatedEmployee : emp
        ));

        toast.success(`员工 ${values.legalName} 信息已更新`);
      } else {
        // Create new employee
        const newEmployee: Employee = {
          id: Date.now().toString(),
          employeeId: values.employeeId,
          legalName: values.legalName,
          preferredName: values.preferredName,
          email: values.email,
          status: 'ACTIVE',
          hireDate: values.hireDate ? format(new Date(values.hireDate), 'yyyy-MM-dd') : '',
          department: values.department,
          position: values.position,
          managerName: values.managerName
        };

        setEmployees(prev => [...prev, newEmployee]);
        
        toast.success(`员工 ${values.legalName} 已成功添加到系统中`);
      }
      
      handleModalClose();
    } catch (error) {
      toast.error('操作时发生错误，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (employee: Employee) => {
    setEditingEmployee(employee);
    setFormData({
      ...employee,
      hireDate: employee.hireDate
    });
    setIsModalVisible(true);
  };

  const handleDelete = (employee: Employee) => {
    if (confirm(`确定要删除员工 ${employee.legalName} 吗？此操作不可撤销。`)) {
      setEmployees(prev => prev.filter(emp => emp.id !== employee.id));
      toast.success(`员工 ${employee.legalName} 已从系统中删除`);
    }
  };

  const handleModalClose = () => {
    setIsModalVisible(false);
    setEditingEmployee(null);
    setFormData({});
  };

  const getStatusColor = (status: string): "default" | "destructive" | "secondary" => {
    const colors = {
      ACTIVE: 'default' as const,
      INACTIVE: 'destructive' as const,
      PENDING: 'secondary' as const
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getStatusLabel = (status: string) => {
    const labels = {
      ACTIVE: '在职',
      INACTIVE: '离职',
      PENDING: '待入职'
    };
    return labels[status as keyof typeof labels] || status;
  };

  const ActionsCell = ({ row }: { row: Employee }) => (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 p-0">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => handleEdit(row)}>
          <Edit2 className="mr-2 h-4 w-4" />
          编辑信息
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => router.push(`/employees/positions/${row.id}`)}>
          <History className="mr-2 h-4 w-4" />
          职位历史
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => handleDelete(row)} className="text-destructive">
          <Trash2 className="mr-2 h-4 w-4" />
          删除员工
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

  const columns: ColumnDef<Employee>[] = [
    {
      accessorKey: 'legalName',
      header: '员工信息',
      cell: ({ row }) => {
        const employee = row.original;
        return (
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-blue-500 text-white flex items-center justify-center">
              {employee.legalName.charAt(0)}
            </div>
            <div>
              <div className="font-medium">
                {employee.legalName}
                {employee.preferredName && (
                  <span className="text-gray-500 ml-2">
                    ({employee.preferredName})
                  </span>
                )}
              </div>
              <div className="text-sm text-gray-500 flex items-center gap-2">
                <span>{employee.employeeId}</span>
                <span>•</span>
                <span>{employee.email}</span>
              </div>
            </div>
          </div>
        );
      },
    },
    {
      accessorKey: 'position',
      header: '职位信息',
      cell: ({ row }) => {
        const employee = row.original;
        return (
          <div>
            <div className="font-medium">
              {employee.position || '未设置'}
            </div>
            <div className="text-sm text-gray-500">
              {employee.department || '未设置部门'}
            </div>
          </div>
        );
      },
    },
    {
      accessorKey: 'managerName',
      header: '直属经理',
      cell: ({ row }) => {
        const managerName = row.original.managerName;
        return (
          <div className="flex items-center gap-2">
            {managerName ? (
              <>
                <Users className="h-4 w-4 text-blue-500" />
                <span>{managerName}</span>
              </>
            ) : (
              <span className="text-gray-400">无</span>
            )}
          </div>
        );
      },
    },
    {
      accessorKey: 'hireDate',
      header: '入职日期',
      cell: ({ row }) => {
        const hireDate = row.original.hireDate;
        return (
          <div className="flex items-center gap-2">
            <Calendar className="h-4 w-4 text-green-500" />
            <span>{format(new Date(hireDate), 'yyyy年MM月dd日', { locale: zhCN })}</span>
          </div>
        );
      },
    },
    {
      accessorKey: 'status',
      header: '状态',
      cell: ({ row }) => {
        const status = row.original.status;
        return (
          <Badge variant={getStatusColor(status)}>
            {getStatusLabel(status)}
          </Badge>
        );
      },
    },
    createActionsColumn<Employee>(ActionsCell),
  ];

  const departments = Array.from(new Set(employees.map(emp => emp.department).filter(Boolean)));

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6 flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold">员工管理</h1>
          <p className="text-gray-600 mt-1">
            管理公司员工信息、职位变更和组织结构 - 完整CRUD功能
          </p>
        </div>
        <Button 
          size="lg"
          onClick={() => setIsModalVisible(true)}
        >
          <Plus className="mr-2 h-4 w-4" />
          新增员工
        </Button>
      </div>

      {/* Employee Table */}
      <Card>
        <CardContent className="p-6">
          <DataTable
            columns={columns}
            data={employees}
            searchKey="legalName"
            searchPlaceholder="搜索员工姓名、工号、邮箱或职位..."
          />
        </CardContent>
      </Card>

      {/* Create/Edit Employee Modal */}
      <Dialog open={isModalVisible} onOpenChange={setIsModalVisible}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {editingEmployee ? '编辑员工信息' : '新增员工'}
            </DialogTitle>
          </DialogHeader>
          
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium">员工工号</label>
              <Input 
                placeholder="如: EMP001"
                value={formData.employeeId || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, employeeId: e.target.value }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">法定姓名</label>
              <Input 
                placeholder="员工的法定姓名"
                value={formData.legalName || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, legalName: e.target.value }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">常用姓名</label>
              <Input 
                placeholder="常用的英文姓名(可选)"
                value={formData.preferredName || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, preferredName: e.target.value }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">邮箱地址</label>
              <Input 
                type="email"
                placeholder="employee@company.com"
                value={formData.email || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">所属部门</label>
              <Select 
                value={formData.department || ''}
                onValueChange={(value) => setFormData(prev => ({ ...prev, department: value }))}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择部门" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="技术部">技术部</SelectItem>
                  <SelectItem value="产品部">产品部</SelectItem>
                  <SelectItem value="人事部">人事部</SelectItem>
                  <SelectItem value="财务部">财务部</SelectItem>
                  <SelectItem value="市场部">市场部</SelectItem>
                  <SelectItem value="运营部">运营部</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <label className="text-sm font-medium">职位</label>
              <Input 
                placeholder="如: 高级软件工程师"
                value={formData.position || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, position: e.target.value }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">入职日期</label>
              <DatePicker 
                date={formData.hireDate ? new Date(formData.hireDate) : undefined}
                onDateChange={(date) => setFormData(prev => ({ 
                  ...prev, 
                  hireDate: date ? format(date, 'yyyy-MM-dd') : ''
                }))}
                placeholder="选择入职日期"
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">直属经理</label>
              <Input 
                placeholder="直属经理姓名(可选)"
                value={formData.managerName || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, managerName: e.target.value }))}
              />
            </div>
          </div>

          <div className="flex justify-end gap-2 mt-6">
            <Button variant="outline" onClick={handleModalClose}>
              取消
            </Button>
            <Button 
              onClick={() => handleCreateEmployee(formData)} 
              disabled={loading}
            >
              {editingEmployee ? '更新' : '创建'}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default EmployeesPage;