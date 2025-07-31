import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { format } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { 
  ArrowLeft,
  Plus, 
  MoreHorizontal,
  User,
  Briefcase,
  Calendar,
  TrendingUp,
  Building2,
  DollarSign,
  Clock,
  CheckCircle,
  AlertCircle,
  Edit2,
  Eye,
  MapPin
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
import { Textarea } from '@/components/ui/textarea';
import { DataTable, createSortableColumn, createActionsColumn } from '@/components/ui/data-table';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { DatePicker } from '@/components/ui/date-picker';

interface Employee {
  id: string;
  employeeId: string;
  legalName: string;
  preferredName?: string;
  email: string;
  hireDate: string;
  currentPosition?: string;
  currentDepartment?: string;
  avatar?: string;
}

interface PositionHistory {
  id: string;
  employeeId: string;
  positionTitle: string;
  department: string;
  jobLevel: string;
  salary?: number;
  currency: string;
  startDate: string;
  endDate?: string;
  changeType: 'hire' | 'promotion' | 'transfer' | 'demotion' | 'termination';
  changeReason?: string;
  approvedBy?: string;
  approvedAt?: string;
  isActive: boolean;
  createdAt: string;
}

const EmployeePositionHistoryPage: React.FC = () => {
  const router = useRouter();
  const { id } = router.query;
  const [loading, setLoading] = useState(false);
  const [employee, setEmployee] = useState<Employee | null>(null);
  const [positionHistory, setPositionHistory] = useState<PositionHistory[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingHistory, setEditingHistory] = useState<PositionHistory | null>(null);
  const [formData, setFormData] = useState<Partial<PositionHistory>>({});

  // Sample data
  useEffect(() => {
    if (!id) return;
    
    setLoading(true);
    setTimeout(() => {
      const sampleEmployee: Employee = {
        id: id as string,
        employeeId: 'EMP001',
        legalName: '张三',
        preferredName: 'Zhang San',
        email: 'zhangsan@company.com',
        hireDate: '2022-01-15',
        currentPosition: '高级前端工程师',
        currentDepartment: '技术部'
      };

      const sampleHistory: PositionHistory[] = [
        {
          id: '1',
          employeeId: id as string,
          positionTitle: '高级前端工程师',
          department: '技术部',
          jobLevel: 'P6',
          salary: 25000,
          currency: 'CNY',
          startDate: '2024-06-01',
          changeType: 'promotion',
          changeReason: '技术能力出色，项目贡献突出',
          approvedBy: '李强',
          approvedAt: '2024-05-25',
          isActive: true,
          createdAt: '2024-05-20'
        },
        {
          id: '2',
          employeeId: id as string,
          positionTitle: '前端工程师',
          department: '技术部',
          jobLevel: 'P5',
          salary: 18000,
          currency: 'CNY',
          startDate: '2023-01-15',
          endDate: '2024-05-31',
          changeType: 'promotion',
          changeReason: '年度绩效优秀，技术成长显著',
          approvedBy: '李强',
          approvedAt: '2022-12-20',
          isActive: false,
          createdAt: '2022-12-15'
        },
        {
          id: '3',
          employeeId: id as string,
          positionTitle: '初级前端工程师',
          department: '技术部',
          jobLevel: 'P4',
          salary: 12000,
          currency: 'CNY',
          startDate: '2022-01-15',
          endDate: '2023-01-14',
          changeType: 'hire',
          changeReason: '校园招聘，技术基础扎实',
          approvedBy: '陈静',
          approvedAt: '2022-01-10',
          isActive: false,
          createdAt: '2022-01-10'
        }
      ];
      
      setEmployee(sampleEmployee);
      setPositionHistory(sampleHistory);
      setLoading(false);
    }, 1000);
  }, [id]);

  const handleCreateHistory = async (values: any) => {
    try {
      setLoading(true);
      
      if (editingHistory) {
        // Update existing history
        const updatedHistory: PositionHistory = {
          ...editingHistory,
          positionTitle: values.positionTitle,
          department: values.department,
          jobLevel: values.jobLevel,
          salary: Number(values.salary),
          currency: values.currency,
          startDate: values.startDate ? format(new Date(values.startDate), 'yyyy-MM-dd') : '',
          endDate: values.endDate ? format(new Date(values.endDate), 'yyyy-MM-dd') : undefined,
          changeType: values.changeType,
          changeReason: values.changeReason,
          approvedBy: values.approvedBy
        };

        setPositionHistory(prev => prev.map(hist => 
          hist.id === editingHistory.id ? updatedHistory : hist
        ));

        toast.success(`职位历史记录已更新`);
      } else {
        // Create new history
        const newHistory: PositionHistory = {
          id: Date.now().toString(),
          employeeId: id as string,
          positionTitle: values.positionTitle,
          department: values.department,
          jobLevel: values.jobLevel,
          salary: Number(values.salary),
          currency: values.currency,
          startDate: values.startDate ? format(new Date(values.startDate), 'yyyy-MM-dd') : '',
          endDate: values.endDate ? format(new Date(values.endDate), 'yyyy-MM-dd') : undefined,
          changeType: values.changeType,
          changeReason: values.changeReason,
          approvedBy: values.approvedBy,
          approvedAt: format(new Date(), 'yyyy-MM-dd'),
          isActive: !values.endDate,
          createdAt: format(new Date(), 'yyyy-MM-dd')
        };

        setPositionHistory(prev => [...prev, newHistory]);
        
        toast.success(`新职位历史记录已添加`);
      }
      
      handleModalClose();
    } catch (error) {
      toast.error('操作时发生错误，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (history: PositionHistory) => {
    setEditingHistory(history);
    setFormData({
      ...history,
      startDate: history.startDate,
      endDate: history.endDate
    });
    setIsModalVisible(true);
  };

  const handleModalClose = () => {
    setIsModalVisible(false);
    setEditingHistory(null);
    setFormData({});
  };

  const getChangeTypeColor = (changeType: string): "default" | "destructive" | "secondary" => {
    const colors = {
      hire: 'default' as const,
      promotion: 'default' as const,
      transfer: 'secondary' as const,
      demotion: 'destructive' as const,
      termination: 'destructive' as const
    };
    return colors[changeType as keyof typeof colors] || 'default';
  };

  const getChangeTypeLabel = (changeType: string) => {
    const labels = {
      hire: '入职',
      promotion: '晋升',
      transfer: '调岗',
      demotion: '降职',
      termination: '离职'
    };
    return labels[changeType as keyof typeof labels] || changeType;
  };

  const formatSalary = (salary?: number, currency: string = 'CNY') => {
    if (!salary) return '未设置';
    const formatter = new Intl.NumberFormat('zh-CN');
    const currencySymbol = currency === 'CNY' ? '¥' : '$';
    return `${currencySymbol}${formatter.format(salary)}`;
  };

  const calculateDuration = (startDate: string, endDate?: string) => {
    const start = new Date(startDate);
    const end = endDate ? new Date(endDate) : new Date();
    const diffInMs = end.getTime() - start.getTime();
    const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));
    const months = Math.floor(diffInDays / 30);
    const days = diffInDays % 30;
    
    if (months === 0) return `${days}天`;
    if (days === 0) return `${months}个月`;
    return `${months}个月${days}天`;
  };

  const ActionsCell = ({ row }: { row: PositionHistory }) => (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 p-0">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => handleEdit(row)}>
          <Edit2 className="mr-2 h-4 w-4" />
          编辑记录
        </DropdownMenuItem>
        <DropdownMenuItem>
          <Eye className="mr-2 h-4 w-4" />
          查看详情
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

  const columns: ColumnDef<PositionHistory>[] = [
    {
      accessorKey: 'positionTitle',
      header: '职位信息',
      cell: ({ row }) => {
        const history = row.original;
        return (
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-blue-500 text-white flex items-center justify-center">
              <Briefcase className="h-5 w-5" />
            </div>
            <div>
              <div className="font-medium">
                {history.positionTitle}
                {history.isActive && (
                  <Badge variant="default" className="ml-2 text-xs">
                    当前职位
                  </Badge>
                )}
              </div>
              <div className="text-sm text-gray-500 flex items-center gap-2">
                <span>{history.jobLevel}</span>
                <span>•</span>
                <span>{history.department}</span>
              </div>
            </div>
          </div>
        );
      },
    },
    {
      accessorKey: 'startDate',
      header: '任职期间',
      cell: ({ row }) => {
        const history = row.original;
        const duration = calculateDuration(history.startDate, history.endDate);
        
        return (
          <div>
            <div className="flex items-center gap-2 mb-1">
              <Calendar className="h-4 w-4 text-blue-500" />
              <span className="font-medium">
                {format(new Date(history.startDate), 'yyyy年MM月dd日', { locale: zhCN })}
              </span>
            </div>
            {history.endDate && (
              <div className="flex items-center gap-2 mb-1">
                <span className="w-4 h-4 flex items-center justify-center">→</span>
                <span>
                  {format(new Date(history.endDate), 'yyyy年MM月dd日', { locale: zhCN })}
                </span>
              </div>
            )}
            <div className="text-sm text-gray-500 flex items-center gap-1">
              <Clock className="h-3 w-3" />
              <span>{duration}</span>
            </div>
          </div>
        );
      },
    },
    {
      accessorKey: 'salary',
      header: '薪资',
      cell: ({ row }) => {
        const history = row.original;
        return (
          <div className="flex items-center gap-2">
            <DollarSign className="h-4 w-4 text-green-500" />
            <span className="font-medium">
              {formatSalary(history.salary, history.currency)}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'changeType',
      header: '变更类型',
      cell: ({ row }) => {
        const history = row.original;
        return (
          <div>
            <Badge variant={getChangeTypeColor(history.changeType)} className="mb-1">
              {getChangeTypeLabel(history.changeType)}
            </Badge>
            {history.approvedBy && (
              <div className="text-xs text-gray-500 flex items-center gap-1">
                <CheckCircle className="h-3 w-3" />
                <span>审批人: {history.approvedBy}</span>
              </div>
            )}
          </div>
        );
      },
    },
    createActionsColumn<PositionHistory>(ActionsCell),
  ];

  const renderTimeline = () => {
    const sortedHistory = [...positionHistory].sort((a, b) => 
      new Date(b.startDate).getTime() - new Date(a.startDate).getTime()
    );

    return (
      <div className="space-y-4">
        {sortedHistory.map((history, index) => (
          <div key={history.id} className="flex items-start gap-4 relative">
            {/* Timeline Line */}
            {index < sortedHistory.length - 1 && (
              <div className="absolute left-6 top-12 w-px h-16 bg-gray-200"></div>
            )}
            
            {/* Timeline Node */}
            <div className={`w-12 h-12 rounded-full flex items-center justify-center ${
              history.isActive ? 'bg-blue-500' : 'bg-gray-400'
            } text-white flex-shrink-0`}>
              <Briefcase className="h-6 w-6" />
            </div>
            
            {/* Timeline Content */}
            <div className="flex-1 pb-8">
              <Card>
                <CardContent className="p-4">
                  <div className="flex justify-between items-start mb-3">
                    <div>
                      <h3 className="font-semibold text-lg flex items-center gap-2">
                        {history.positionTitle}
                        {history.isActive && (
                          <Badge variant="default">当前职位</Badge>
                        )}
                      </h3>
                      <p className="text-gray-600">{history.department} • {history.jobLevel}</p>
                    </div>
                    <Badge variant={getChangeTypeColor(history.changeType)}>
                      {getChangeTypeLabel(history.changeType)}
                    </Badge>
                  </div>
                  
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-3">
                    <div>
                      <p className="text-sm text-gray-500 mb-1">开始时间</p>
                      <p className="font-medium">
                        {format(new Date(history.startDate), 'yyyy-MM-dd')}
                      </p>
                    </div>
                    
                    {history.endDate && (
                      <div>
                        <p className="text-sm text-gray-500 mb-1">结束时间</p>
                        <p className="font-medium">
                          {format(new Date(history.endDate), 'yyyy-MM-dd')}
                        </p>
                      </div>
                    )}
                    
                    <div>
                      <p className="text-sm text-gray-500 mb-1">任职时长</p>
                      <p className="font-medium">
                        {calculateDuration(history.startDate, history.endDate)}
                      </p>
                    </div>
                    
                    <div>
                      <p className="text-sm text-gray-500 mb-1">薪资</p>
                      <p className="font-medium">
                        {formatSalary(history.salary, history.currency)}
                      </p>
                    </div>
                  </div>
                  
                  {history.changeReason && (
                    <div>
                      <p className="text-sm text-gray-500 mb-1">变更原因</p>
                      <p className="text-sm bg-gray-50 p-2 rounded">{history.changeReason}</p>
                    </div>
                  )}
                  
                  {history.approvedBy && (
                    <div className="mt-2 flex items-center text-sm text-gray-500">
                      <CheckCircle className="h-4 w-4 mr-1" />
                      <span>
                        由 {history.approvedBy} 于 {history.approvedAt ? format(new Date(history.approvedAt), 'yyyy-MM-dd') : '未知日期'} 审批
                      </span>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          </div>
        ))}
      </div>
    );
  };

  const changeTypes = [
    { value: 'hire', label: '入职' },
    { value: 'promotion', label: '晋升' },
    { value: 'transfer', label: '调岗' },
    { value: 'demotion', label: '降职' },
    { value: 'termination', label: '离职' }
  ];

  const jobLevels = ['P3', 'P4', 'P5', 'P6', 'P7', 'P8', 'M1', 'M2', 'M3'];

  if (!employee) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-4">员工不存在</h1>
          <p className="text-gray-600 mb-4">请检查员工ID是否正确</p>
          <Button onClick={() => router.push('/employees')}>
            返回员工列表
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6 flex justify-between items-center">
        <div className="flex items-center gap-4">
          <Button variant="ghost" onClick={() => router.back()}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            返回
          </Button>
          <div>
            <h1 className="text-2xl font-bold">职位历史</h1>
            <p className="text-gray-600 mt-1">
              {employee.legalName} ({employee.employeeId}) 的职位变更历史
            </p>
          </div>
        </div>
        <Button onClick={() => setIsModalVisible(true)}>
          <Plus className="mr-2 h-4 w-4" />
          新增记录
        </Button>
      </div>

      {/* Employee Info Card */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-blue-500 text-white flex items-center justify-center text-xl font-bold">
              {employee.legalName.charAt(0)}
            </div>
            <div className="flex-1">
              <h2 className="text-xl font-bold">
                {employee.legalName}
                {employee.preferredName && (
                  <span className="text-gray-500 ml-2 text-base">
                    ({employee.preferredName})
                  </span>
                )}
              </h2>
              <div className="flex items-center gap-4 mt-2 text-gray-600">
                <div className="flex items-center gap-1">
                  <User className="h-4 w-4" />
                  <span>{employee.employeeId}</span>
                </div>
                <div className="flex items-center gap-1">
                  <Calendar className="h-4 w-4" />
                  <span>入职: {format(new Date(employee.hireDate), 'yyyy年MM月dd日', { locale: zhCN })}</span>
                </div>
                {employee.currentPosition && (
                  <div className="flex items-center gap-1">
                    <Briefcase className="h-4 w-4" />
                    <span>{employee.currentPosition}</span>
                  </div>
                )}
                {employee.currentDepartment && (
                  <div className="flex items-center gap-1">
                    <Building2 className="h-4 w-4" />
                    <span>{employee.currentDepartment}</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">总职位数</p>
                <p className="text-2xl font-bold">{positionHistory.length}</p>
              </div>
              <Briefcase className="h-8 w-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">在职时长</p>
                <p className="text-2xl font-bold text-green-600">
                  {calculateDuration(employee.hireDate)}
                </p>
              </div>
              <Clock className="h-8 w-8 text-green-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">晋升次数</p>
                <p className="text-2xl font-bold text-purple-600">
                  {positionHistory.filter(h => h.changeType === 'promotion').length}
                </p>
              </div>
              <TrendingUp className="h-8 w-8 text-purple-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">当前薪资</p>
                <p className="text-2xl font-bold text-orange-600">
                  {(() => {
                    const current = positionHistory.find(h => h.isActive);
                    return current ? formatSalary(current.salary, current.currency) : '未设置';
                  })()}
                </p>
              </div>
              <DollarSign className="h-8 w-8 text-orange-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Timeline View */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <MapPin className="h-5 w-5" />
              职业时间线
            </CardTitle>
          </CardHeader>
          <CardContent className="p-6">
            {loading ? (
              <div className="flex items-center justify-center py-8">
                <div className="text-gray-500">加载中...</div>
              </div>
            ) : positionHistory.length > 0 ? (
              renderTimeline()
            ) : (
              <Alert>
                <AlertDescription>
                  暂无职位历史记录
                </AlertDescription>
              </Alert>
            )}
          </CardContent>
        </Card>

        {/* Table View */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Briefcase className="h-5 w-5" />
              详细记录
            </CardTitle>
          </CardHeader>
          <CardContent className="p-6">
            <DataTable
              columns={columns}
              data={positionHistory}
              searchKey="positionTitle"
              searchPlaceholder="搜索职位..."
            />
          </CardContent>
        </Card>
      </div>

      {/* Create/Edit History Modal */}
      <Dialog open={isModalVisible} onOpenChange={setIsModalVisible}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {editingHistory ? '编辑职位记录' : '新增职位记录'}
            </DialogTitle>
          </DialogHeader>
          
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium">职位名称 *</label>
              <Input 
                placeholder="如: 高级软件工程师"
                value={formData.positionTitle || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, positionTitle: e.target.value }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">所属部门 *</label>
              <Input 
                placeholder="如: 技术部"
                value={formData.department || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, department: e.target.value }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">职级 *</label>
              <Select 
                value={formData.jobLevel || ''}
                onValueChange={(value) => setFormData(prev => ({ ...prev, jobLevel: value }))}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择职级" />
                </SelectTrigger>
                <SelectContent>
                  {jobLevels.map(level => (
                    <SelectItem key={level} value={level}>{level}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <label className="text-sm font-medium">薪资</label>
              <Input 
                type="number"
                placeholder="18000"
                value={formData.salary || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, salary: Number(e.target.value) }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">开始日期 *</label>
              <DatePicker 
                date={formData.startDate ? new Date(formData.startDate) : undefined}
                onDateChange={(date) => setFormData(prev => ({ 
                  ...prev, 
                  startDate: date ? format(date, 'yyyy-MM-dd') : ''
                }))}
                placeholder="选择开始日期"
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">结束日期</label>
              <DatePicker 
                date={formData.endDate ? new Date(formData.endDate) : undefined}
                onDateChange={(date) => setFormData(prev => ({ 
                  ...prev, 
                  endDate: date ? format(date, 'yyyy-MM-dd') : ''
                }))}
                placeholder="选择结束日期(可选)"
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">变更类型 *</label>
              <Select 
                value={formData.changeType || ''}
                onValueChange={(value) => setFormData(prev => ({ ...prev, changeType: value as PositionHistory['changeType'] }))}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择变更类型" />
                </SelectTrigger>
                <SelectContent>
                  {changeTypes.map(type => (
                    <SelectItem key={type.value} value={type.value}>
                      {type.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <label className="text-sm font-medium">审批人</label>
              <Input 
                placeholder="审批人姓名"
                value={formData.approvedBy || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, approvedBy: e.target.value }))}
              />
            </div>
          </div>

          <div>
            <label className="text-sm font-medium">变更原因</label>
            <Textarea 
              placeholder="描述本次职位变更的原因..."
              rows={3}
              value={formData.changeReason || ''}
              onChange={(e) => setFormData(prev => ({ ...prev, changeReason: e.target.value }))}
            />
          </div>

          <div className="flex justify-end gap-2 mt-6">
            <Button variant="outline" onClick={handleModalClose}>
              取消
            </Button>
            <Button 
              onClick={() => handleCreateHistory(formData)} 
              disabled={loading}
            >
              {editingHistory ? '更新' : '创建'}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default EmployeePositionHistoryPage;