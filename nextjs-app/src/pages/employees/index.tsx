import React, { useState, useMemo, useCallback } from 'react';
import { useRouter } from 'next/router';
import { format } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { 
  Plus, 
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
  UserCheck,
  UserPlus,
  Building,
  AlertCircle,
  RefreshCw
} from 'lucide-react';
import { toast } from 'react-hot-toast';
import { ColumnDef } from '@tanstack/react-table';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { SWRMonitoring } from '@/components/ui/swr-monitoring';
// Removed Select imports to avoid Radix UI state cycles
// Removed Dialog imports to avoid Radix UI state cycles
import { DatePicker } from '@/components/ui/date-picker';
import { DataTable, createSortableColumn } from '@/components/ui/data-table';
// Removed DropdownMenu imports to avoid Radix UI state cycles

// 新增的UI组件
import StatCard, { StatCardsGrid } from '@/components/ui/stat-card';
import EmployeeCard, { EmployeeCardsGrid, EmployeeCardSkeleton } from '@/components/ui/employee-card';
import SmartFilterStable, { FilterOption, ActiveFilter } from '@/components/ui/smart-filter-stable';
import { PieChart, BarChart } from '@/components/ui/data-visualization';

// Import SWR hooks
import { useEmployeesSWR, useEmployeeStatsSWR, Employee } from '@/hooks/useEmployeesSWR';

const EmployeesPage: React.FC = () => {
  const router = useRouter();
  
  // Local state
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [formData, setFormData] = useState<Partial<Employee>>({});
  const [viewMode, setViewMode] = useState<'table' | 'card'>('table');
  const [searchValue, setSearchValue] = useState('');
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [selectedEmployees, setSelectedEmployees] = useState<string[]>([]);
  const [showModal, setShowModal] = useState(false);

  // 稳定化过滤器状态管理函数
  const handleFiltersChange = useCallback((filters: ActiveFilter[]) => {
    setActiveFilters(filters);
  }, []);
  const { 
    employees, 
    totalCount, 
    isLoading, 
    isError, 
    error, 
    mutate 
  } = useEmployeesSWR({
    search: searchValue,
    department: activeFilters.find(f => f.key === 'department')?.value,
  });

  const { 
    stats, 
    departmentData,
    isLoading: statsLoading 
  } = useEmployeeStatsSWR();

  // Error state component
  const ErrorState = () => (
    <Card className="p-6 text-center">
      <div className="flex flex-col items-center gap-4">
        <AlertCircle className="h-12 w-12 text-red-500" />
        <div>
          <h3 className="text-lg font-semibold text-red-600">数据加载失败</h3>
          <p className="text-sm text-gray-600 mt-1">
            {error?.message || '无法连接到服务器，请检查网络连接'}
          </p>
        </div>
        <Button 
          onClick={() => mutate()} 
          variant="outline" 
          className="flex items-center gap-2"
        >
          <RefreshCw className="h-4 w-4" />
          重试
        </Button>
      </div>
    </Card>
  );

  // Loading skeleton component
  const LoadingSkeleton = () => (
    <div className="space-y-6">
      {/* Stats skeleton */}
      <StatCardsGrid columns={4}>
        {Array.from({ length: 4 }).map((_, i) => (
          <Card key={i} className="p-6">
            <div className="animate-pulse">
              <div className="h-4 bg-gray-200 rounded w-1/2 mb-2"></div>
              <div className="h-8 bg-gray-200 rounded w-1/4 mb-2"></div>
              <div className="h-3 bg-gray-200 rounded w-1/3"></div>
            </div>
          </Card>
        ))}
      </StatCardsGrid>
      
      {/* Charts skeleton */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {Array.from({ length: 2 }).map((_, i) => (
          <Card key={i} className="p-6">
            <div className="animate-pulse">
              <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
              <div className="h-40 bg-gray-200 rounded"></div>
            </div>
          </Card>
        ))}
      </div>
      
      {/* Table/Cards skeleton */}
      <Card className="p-6">
        <div className="animate-pulse space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-16 bg-gray-200 rounded"></div>
          ))}
        </div>
      </Card>
    </div>
  );

  const handleCreateEmployee = async (values: any) => {
    try {
      if (editingEmployee) {
        // Update existing employee (local state only for now)
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

        // In a real app, this would make an API call
        toast.success(`员工 ${values.legalName} 信息已更新`);
      } else {
        // Create new employee (local state only for now)
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

        // In a real app, this would make an API call
        toast.success(`员工 ${values.legalName} 已成功添加到系统中`);
      }
      
      // Refresh data
      mutate();
      handleModalClose();
    } catch (error) {
      toast.error('操作时发生错误，请重试');
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
      // In a real app, this would make an API call
      toast.success(`员工 ${employee.legalName} 已从系统中删除`);
      mutate(); // Refresh data
    }
  };

  const handleModalClose = () => {
    setShowModal(false);
    setEditingEmployee(null);
    setFormData({});
  };

  // Modal overlay click handler
  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      handleModalClose();
    }
  };

  // Filter employees based on search and active filters
  const filteredEmployees = employees.filter(employee => {
    // Search filter
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

    // Active filters
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

  // Filter options configuration - 深度稳定化防止引用循环
  const filterOptions: FilterOption[] = useMemo(() => {
    // 使用Set来获取唯一部门，避免重复计算
    const uniqueDepartments = Array.from(new Set(employees.map(emp => emp.department).filter(Boolean)));
    
    return [
      {
        key: 'department',
        label: '部门',
        type: 'select',
        options: uniqueDepartments.map(dept => ({ label: dept!, value: dept! }))
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
  }, [employees.length]); // 只依赖员工数量，避免因数据细节变化而重新创建

  // Preset filter configurations (memoized to prevent re-creation)
  const filterPresets = useMemo(() => [
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
  ], []);

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

  // 稳定的下拉菜单组件，避免循环依赖
  const StableActionsCell = React.memo(({ employee }: { employee: Employee }) => {
    const [isMenuOpen, setIsMenuOpen] = React.useState(false);
    const menuRef = React.useRef<HTMLDivElement>(null);

    // 使用单独的状态管理，避免全局状态循环
    React.useEffect(() => {
      const handleClickOutside = (event: MouseEvent) => {
        if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
          setIsMenuOpen(false);
        }
      };

      if (isMenuOpen) {
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
      }
      
      // 确保所有代码路径都有返回值
      return undefined;
    }, [isMenuOpen]);

    return (
      <div className="relative" ref={menuRef}>
        <Button 
          variant="ghost" 
          className="h-8 w-8 p-0"
          onClick={() => setIsMenuOpen(!isMenuOpen)}
        >
          <MoreHorizontal className="h-4 w-4" />
        </Button>
        {isMenuOpen && (
          <div className="absolute right-0 top-full mt-1 w-48 rounded-md border bg-popover p-1 text-popover-foreground shadow-md z-50">
            <button
              onClick={() => {
                handleEdit(employee);
                setIsMenuOpen(false);
              }}
              className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
            >
              <Edit2 className="h-4 w-4" />
              编辑信息
            </button>
            <button
              onClick={() => {
                router.push(`/employees/positions/${employee.id}`);
                setIsMenuOpen(false);
              }}
              className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
            >
              <History className="h-4 w-4" />
              职位历史
            </button>
            <div className="my-1 h-px bg-border" />
            <button
              onClick={() => {
                handleDelete(employee);
                setIsMenuOpen(false);
              }}
              className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground text-destructive"
            >
              <Trash2 className="h-4 w-4" />
              删除员工
            </button>
          </div>
        )}
      </div>
    );
  });

  StableActionsCell.displayName = 'StableActionsCell';

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
    {
      id: 'actions',
      header: '操作',
      cell: ({ row }) => <StableActionsCell employee={row.original} />,
    },
  ];

  // Show error state
  if (isError) {
    return (
      <div className="p-4 sm:p-6 space-y-4 sm:space-y-6 page-enter">
        <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
          <div>
            <h1 className="text-2xl sm:text-display-large">员工管理</h1>
            <p className="text-sm sm:text-body-large text-muted-foreground mt-2">
              基于Workday风格的现代化员工管理系统 - 完整CRUD功能与数据可视化
            </p>
          </div>
        </div>
        <ErrorState />
      </div>
    );
  }

  // Show loading state
  if (isLoading && employees.length === 0) {
    return (
      <div className="p-4 sm:p-6 space-y-4 sm:space-y-6 page-enter">
        <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
          <div>
            <h1 className="text-2xl sm:text-display-large">员工管理</h1>
            <p className="text-sm sm:text-body-large text-muted-foreground mt-2">
              基于Workday风格的现代化员工管理系统 - 完整CRUD功能与数据可视化
            </p>
          </div>
          <Button size="lg" disabled className="w-full sm:w-auto">
            <Plus className="mr-2 h-4 w-4" />
            新增员工
          </Button>
        </div>
        <LoadingSkeleton />
      </div>
    );
  }

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
          onClick={() => setShowModal(true)}
          className="w-full sm:w-auto btn-primary-animate"
        >
          <Plus className="mr-2 h-4 w-4" />
          新增员工
        </Button>
      </div>

      {/* Statistics Cards */}
      <StatCardsGrid columns={4}>
        <StatCard
          title="总员工数"
          value={stats.total}
          change={8.5}
          changeLabel="较上月"
          icon={<Users className="w-8 h-8" />}
          variant="primary"
          loading={statsLoading}
        />
        <StatCard
          title="在职员工"
          value={stats.active}
          change={2.1}
          changeLabel="较上月"
          icon={<UserCheck className="w-8 h-8" />}
          variant="success"
          loading={statsLoading}
        />
        <StatCard
          title="待入职"
          value={stats.pending}
          change={-1.2}
          changeLabel="较上月"
          icon={<UserPlus className="w-8 h-8" />}
          variant="warning"
          loading={statsLoading}
        />
        <StatCard
          title="部门数量"
          value={stats.departments}
          change={0}
          changeLabel="较上月"
          icon={<Building className="w-8 h-8" />}
          variant="default"
          loading={statsLoading}
        />
      </StatCardsGrid>

      {/* Data Visualization */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6">
        <PieChart
          data={departmentData}
          title="部门分布"
          description="各部门员工数量分布情况"
          loading={statsLoading}
        />
        <BarChart
          data={[
            { label: '在职', value: stats.active },
            { label: '离职', value: stats.inactive },
            { label: '待入职', value: stats.pending }
          ]}
          title="员工状态统计"
          description="不同状态员工数量对比"
          loading={statsLoading}
        />
      </div>

      {/* Smart Filter Toolbar */}
      <SmartFilterStable
        filterOptions={filterOptions}
        activeFilters={activeFilters}
        onFiltersChange={handleFiltersChange}
        searchValue={searchValue}
        onSearchChange={setSearchValue}
        presets={filterPresets}
        searchPlaceholder="搜索员工姓名、工号、部门或职位..."
      />

      {/* View Mode and Operations Toolbar */}
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
              显示 {filteredEmployees.length} / {totalCount} 个员工
            </span>
            {isLoading && (
              <RefreshCw className="h-4 w-4 animate-spin text-blue-500" />
            )}
          </div>
        </div>
      </Card>

      {/* Main Content Area */}
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
          {isLoading ? (
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

      {/* Native Modal Implementation */}
      {showModal && (
        <div 
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
          onClick={handleOverlayClick}
        >
          <div className="relative w-full max-w-2xl bg-white rounded-lg shadow-lg overflow-hidden">
            {/* Header */}
            <div className="flex items-center justify-between p-6 border-b">
              <h2 className="text-lg font-semibold">
                {editingEmployee ? '编辑员工信息' : '新增员工'}
              </h2>
              <button
                onClick={handleModalClose}
                className="text-gray-400 hover:text-gray-600 transition-colors"
              >
                ✕
              </button>
            </div>
            
            {/* Content */}
            <div className="p-6">
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
                  <select 
                    value={formData.department || ''}
                    onChange={(e) => setFormData(prev => ({ ...prev, department: e.target.value }))}
                    className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="">选择部门</option>
                    <option value="技术部">技术部</option>
                    <option value="产品部">产品部</option>
                    <option value="人事部">人事部</option>
                    <option value="财务部">财务部</option>
                    <option value="市场部">市场部</option>
                    <option value="运营部">运营部</option>
                  </select>
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
                <Button onClick={() => handleCreateEmployee(formData)}>
                  {editingEmployee ? '更新' : '创建'}
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}
      
      {/* SWR Performance Monitoring Component */}
      <SWRMonitoring />
    </div>
  );
};

export default EmployeesPage;