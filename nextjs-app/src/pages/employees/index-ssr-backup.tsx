import React, { useState, useEffect } from 'react';
import dynamic from 'next/dynamic';
import { useRouter } from 'next/router';
import { format } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { GetServerSideProps } from 'next';
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
  Trash2,
  Grid,
  List,
  TrendingUp,
  UserCheck,
  UserPlus,
  Building
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

// 新增的UI组件
import StatCard, { StatCardsGrid } from '@/components/ui/stat-card';
import EmployeeCard, { EmployeeCardsGrid, EmployeeCardSkeleton } from '@/components/ui/employee-card';
import SmartFilter, { FilterOption, ActiveFilter } from '@/components/ui/smart-filter';
import { PieChart, BarChart } from '@/components/ui/data-visualization';

interface Employee {
  id: string;
  employeeId: string;
  legalName: string;
  preferredName?: string;
  email: string;
  phone?: string;
  status: 'active' | 'inactive' | 'pending';
  hireDate: string;
  department?: string;
  position?: string;
  managerId?: string;
  managerName?: string;
  avatar?: string;
}

interface EmployeesPageProps {
  initialEmployees: Employee[];
  error?: string;
}

const EmployeesPage: React.FC<EmployeesPageProps> = ({ initialEmployees, error }) => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [employees, setEmployees] = useState<Employee[]>(initialEmployees || []);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [formData, setFormData] = useState<Partial<Employee>>({});
  
  // 新增状态管理
  const [viewMode, setViewMode] = useState<'table' | 'card'>('table');
  const [searchValue, setSearchValue] = useState('');
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [selectedEmployees, setSelectedEmployees] = useState<string[]>([]);

  // 初始化时显示日志
  useEffect(() => {
    console.log('✅ 页面已初始化，显示', employees.length, '个员工');
    if (error) {
      console.error('❌ 服务器端数据加载错误:', error);
    }
  }, [employees.length, error]);

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
          status: 'active',
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

  // 新增功能函数
  const filteredEmployees = employees.filter(employee => {
    // 搜索过滤
    if (searchValue) {
      const searchLower = searchValue.toLowerCase();
      if (!employee.legalName.toLowerCase().includes(searchLower) &&
          !employee.employeeId.toLowerCase().includes(searchLower) &&
          !employee.email.toLowerCase().includes(searchLower) &&
          !(employee.department?.toLowerCase().includes(searchLower)) &&
          !(employee.position?.toLowerCase().includes(searchLower))) {
        return false;
      }
    }

    // 筛选条件过滤
    for (const filter of activeFilters) {
      if (filter.key === 'department' && employee.department !== filter.value) {
        return false;
      }
      if (filter.key === 'status' && employee.status !== filter.value) {
        return false;
      }
      if (filter.key === 'position' && employee.position !== filter.value) {
        return false;
      }
    }

    return true;
  });

  // 统计数据计算
  const stats = {
    total: employees.length,
    active: employees.filter(emp => emp.status === 'active').length,
    inactive: employees.filter(emp => emp.status === 'inactive').length,
    pending: employees.filter(emp => emp.status === 'pending').length,
    departments: new Set(employees.map(emp => emp.department).filter(Boolean)).size,
  };

  // 部门分布数据
  const departmentData = Array.from(
    employees.reduce((acc, emp) => {
      if (emp.department) {
        acc.set(emp.department, (acc.get(emp.department) || 0) + 1);
      }
      return acc;
    }, new Map())
  ).map(([department, count]) => ({
    label: department,
    value: count,
    color: `hsl(${Math.random() * 360}, 70%, 60%)`
  }));

  // 筛选选项配置
  const filterOptions: FilterOption[] = [
    {
      key: 'department',
      label: '部门',
      type: 'select',
      options: Array.from(new Set(employees.map(emp => emp.department).filter(Boolean)))
        .map(dept => ({ label: dept!, value: dept! })) // 添加非空断言
    },
    {
      key: 'status',
      label: '状态',
      type: 'select',
      options: [
        { label: '在职', value: 'active' },
        { label: '离职', value: 'inactive' },
        { label: '待入职', value: 'pending' }
      ]
    },
    {
      key: 'position',
      label: '职位',
      type: 'text',
      placeholder: '输入职位关键词'
    }
  ];

  // 预设筛选方案
  const filterPresets = [
    {
      label: '全部在职员工',
      icon: <UserCheck className="w-4 h-4" />,
      filters: [{ key: 'status', label: '状态', value: 'active', displayValue: '在职' }]
    },
    {
      label: '技术部员工',
      icon: <Building className="w-4 h-4" />,
      filters: [{ key: 'department', label: '部门', value: '技术部', displayValue: '技术部' }]
    }
  ];

  const getStatusColor = (status: string): "default" | "destructive" | "secondary" => {
    const colors = {
      active: 'default' as const,
      inactive: 'destructive' as const,
      pending: 'secondary' as const
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getStatusLabel = (status: string) => {
    const labels = {
      active: '在职',
      inactive: '离职',
      pending: '待入职'
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
    <div className="p-4 sm:p-6 space-y-4 sm:space-y-6 page-enter">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
        <div>
          <h1 className="text-2xl sm:text-display-large">员工管理</h1>
          <p className="text-sm sm:text-body-large text-muted-foreground mt-2">
            基于Workday风格的现代化员工管理系统 - 完整CRUD功能与数据可视化
          </p>
        </div>
        <Button 
          size="lg"
          onClick={() => setIsModalVisible(true)}
          className="w-full sm:w-auto btn-primary-animate"
        >
          <Plus className="mr-2 h-4 w-4" />
          新增员工
        </Button>
      </div>

      {/* 统计卡片区域 */}
      <StatCardsGrid columns={4}>
        <StatCard
          title="总员工数"
          value={stats.total}
          change={8.5}
          changeLabel="较上月"
          icon={<Users className="w-8 h-8" />}
          variant="primary"
        />
        <StatCard
          title="在职员工"
          value={stats.active}
          change={2.1}
          changeLabel="较上月"
          icon={<UserCheck className="w-8 h-8" />}
          variant="success"
        />
        <StatCard
          title="待入职"
          value={stats.pending}
          change={-1.2}
          changeLabel="较上月"
          icon={<UserPlus className="w-8 h-8" />}
          variant="warning"
        />
        <StatCard
          title="部门数量"
          value={stats.departments}
          change={0}
          changeLabel="较上月"
          icon={<Building className="w-8 h-8" />}
          variant="default"
        />
      </StatCardsGrid>

      {/* 数据可视化区域 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6">
        <PieChart
          data={departmentData}
          title="部门分布"
          description="各部门员工数量分布情况"
          loading={loading}
        />
        <BarChart
          data={[
            { label: '在职', value: stats.active },
            { label: '离职', value: stats.inactive },
            { label: '待入职', value: stats.pending }
          ]}
          title="员工状态统计"
          description="不同状态员工数量对比"
          loading={loading}
        />
      </div>

      {/* 智能筛选工具栏 */}
      <SmartFilter
        filterOptions={filterOptions}
        activeFilters={activeFilters}
        onFiltersChange={setActiveFilters}
        searchValue={searchValue}
        onSearchChange={setSearchValue}
        presets={filterPresets}
        searchPlaceholder="搜索员工姓名、工号、部门或职位..."
      />

      {/* 视图切换和操作工具栏 */}
      <Card className="p-3 sm:p-4">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="flex items-center gap-3">
            <span className="text-sm font-medium">视图模式:</span>
            <div className="flex items-center gap-1 bg-muted rounded-lg p-1">
              <Button
                variant={viewMode === 'table' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('table')}
                className="flex items-center gap-2 text-xs sm:text-sm"
              >
                <List className="h-4 w-4" />
                <span className="hidden sm:inline">表格视图</span>
              </Button>
              <Button
                variant={viewMode === 'card' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('card')}
                className="flex items-center gap-2 text-xs sm:text-sm"
              >
                <Grid className="h-4 w-4" />
                <span className="hidden sm:inline">卡片视图</span>
              </Button>
            </div>
          </div>
          
          <div className="flex items-center gap-2 text-xs sm:text-sm">
            {selectedEmployees.length > 0 && (
              <Badge variant="secondary" className="bg-primary/10 text-primary">
                已选择 {selectedEmployees.length} 个员工
              </Badge>
            )}
            <span className="text-muted-foreground">
              显示 {filteredEmployees.length} / {employees.length} 个员工
            </span>
          </div>
        </div>
      </Card>

      {/* 主内容区域 */}
      {viewMode === 'table' ? (
        <Card>
          <CardContent className="p-6">
            <DataTable
              columns={columns}
              data={filteredEmployees}
              searchKey="legalName"
              searchPlaceholder="搜索员工姓名、工号、邮箱或职位..."
            />
          </CardContent>
        </Card>
      ) : (
        <EmployeeCardsGrid columns={3}>
          {loading ? (
            Array.from({ length: 6 }).map((_, index) => (
              <EmployeeCardSkeleton key={index} />
            ))
          ) : (
            filteredEmployees.map((employee) => (
              <EmployeeCard
                key={employee.id}
                employee={{
                  ...employee,
                  name: employee.legalName
                }}
                selectable={true}
                selected={selectedEmployees.includes(employee.id)}
                onSelectionChange={(selected) => {
                  if (selected) {
                    setSelectedEmployees(prev => [...prev, employee.id]);
                  } else {
                    setSelectedEmployees(prev => prev.filter(id => id !== employee.id));
                  }
                }}
                onClick={() => router.push(`/employees/${employee.id}`)}
                actions={[
                  {
                    label: '编辑信息',
                    icon: <Edit2 className="w-4 h-4" />,
                    onClick: () => handleEdit(employee)
                  },
                  {
                    label: '职位历史',
                    icon: <History className="w-4 h-4" />,
                    onClick: () => router.push(`/employees/positions/${employee.id}`)
                  },
                  {
                    label: '删除员工',
                    icon: <Trash2 className="w-4 h-4" />,
                    onClick: () => handleDelete(employee),
                    variant: 'destructive'
                  }
                ]}
              />
            ))
          )}
        </EmployeeCardsGrid>
      )}

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

// Server-side data fetching
export const getServerSideProps: GetServerSideProps<EmployeesPageProps> = async (context) => {
  try {
    console.log('🚀 服务器端：开始获取员工数据...');
    
    // 直接调用后端API (服务器端可以直接访问localhost:8080)
    const backendUrl = 'http://localhost:8080/api/v1/corehr/employees?page=1&page_size=50';
    console.log('📡 服务器端：请求后端API:', backendUrl);
    
    const response = await fetch(backendUrl);
    console.log('📨 服务器端：收到响应:', response.status, response.statusText);
    
    if (!response.ok) {
      throw new Error(`Backend responded with ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    console.log('📦 服务器端：解析数据，员工数量:', data.employees?.length || 0);
    
    // Convert API data to Employee interface
    const employees: Employee[] = data.employees.map((emp: any) => ({
      id: emp.id,
      employeeId: emp.employee_number,
      legalName: `${emp.first_name} ${emp.last_name}`,
      preferredName: emp.first_name || null,
      email: emp.email,
      phone: emp.phone_number || null,
      status: emp.status?.toLowerCase() === 'active' ? 'active' : 'inactive',
      hireDate: emp.hire_date,
      department: emp.department || '未分配部门',
      position: emp.position || '未设置职位',
      managerName: emp.manager_name || null,
    }));
    
    console.log('✅ 服务器端：成功转换员工数据:', employees.length, '个员工');
    
    return {
      props: {
        initialEmployees: employees,
      },
    };
  } catch (error: any) {
    console.error('❌ 服务器端：获取员工数据失败:', error.message);
    
    return {
      props: {
        initialEmployees: [],
        error: error.message || 'Failed to fetch employees',
      },
    };
  }
};

export default EmployeesPage;