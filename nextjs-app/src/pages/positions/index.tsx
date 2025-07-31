import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { format } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { 
  Plus, 
  Search, 
  MoreHorizontal,
  Briefcase,
  Users,
  DollarSign,
  TrendingUp,
  Building2,
  Star,
  Edit2,
  Trash2,
  Eye
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

interface Position {
  id: string;
  title: string;
  department: string;
  jobLevel: string;
  employeeCount: number;
  maxCapacity: number;
  minSalary: number;
  maxSalary: number;
  currency: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  description?: string;
  requirements?: string;
  benefits?: string;
}

const PositionsPage: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [positions, setPositions] = useState<Position[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingPosition, setEditingPosition] = useState<Position | null>(null);
  const [formData, setFormData] = useState<Partial<Position>>({});

  // Sample data
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      const samplePositions: Position[] = [
        {
          id: '1',
          title: '高级软件工程师',
          department: '技术部',
          jobLevel: 'P6',
          employeeCount: 3,
          maxCapacity: 5,
          minSalary: 18000,
          maxSalary: 30000,
          currency: 'CNY',
          isActive: true,
          createdAt: '2023-01-15',
          updatedAt: '2024-12-01',
          description: '负责核心业务系统的开发和维护',
          requirements: '3年以上React/Node.js开发经验',
          benefits: '五险一金，年终奖，股权激励'
        },
        {
          id: '2', 
          title: '产品经理',
          department: '产品部',
          jobLevel: 'P5',
          employeeCount: 2,
          maxCapacity: 3,
          minSalary: 15000,
          maxSalary: 25000,
          currency: 'CNY',
          isActive: true,
          createdAt: '2023-03-20',
          updatedAt: '2024-11-15',
          description: '负责产品规划和需求分析',
          requirements: '2年以上产品管理经验，有B端产品经验优先',
          benefits: '弹性工作制，培训机会，健康体检'
        },
        {
          id: '3',
          title: '前端工程师',
          department: '技术部', 
          jobLevel: 'P4',
          employeeCount: 1,
          maxCapacity: 4,
          minSalary: 12000,
          maxSalary: 20000,
          currency: 'CNY',
          isActive: true,
          createdAt: '2022-08-10',
          updatedAt: '2024-10-30',
          description: '负责前端页面开发和用户体验优化',
          requirements: 'Vue/React框架熟练，有移动端开发经验',
          benefits: '技术津贴，学习基金，团建活动'
        },
        {
          id: '4',
          title: 'UI设计师',
          department: '设计部',
          jobLevel: 'P4',
          employeeCount: 0,
          maxCapacity: 2,
          minSalary: 10000,
          maxSalary: 18000,
          currency: 'CNY',
          isActive: false,
          createdAt: '2024-01-08',
          updatedAt: '2024-08-20',
          description: '负责产品界面设计和用户体验设计',
          requirements: 'Figma/Sketch熟练使用，有B端产品设计经验',
          benefits: '创意奖金，设计工具报销，作品展示机会'
        }
      ];
      
      setPositions(samplePositions);
      setLoading(false);
    }, 1000);
  }, []);

  const handleCreatePosition = async (values: any) => {
    try {
      setLoading(true);
      
      if (editingPosition) {
        // Update existing position
        const updatedPosition: Position = {
          ...editingPosition,
          title: values.title,
          department: values.department,
          jobLevel: values.jobLevel,
          maxCapacity: Number(values.maxCapacity),
          minSalary: Number(values.minSalary),
          maxSalary: Number(values.maxSalary),
          currency: values.currency,
          isActive: values.isActive,
          description: values.description,
          requirements: values.requirements,
          benefits: values.benefits,
          updatedAt: format(new Date(), 'yyyy-MM-dd')
        };

        setPositions(prev => prev.map(pos => 
          pos.id === editingPosition.id ? updatedPosition : pos
        ));

        toast.success(`职位 ${values.title} 信息已更新`);
      } else {
        // Create new position
        const newPosition: Position = {
          id: Date.now().toString(),
          title: values.title,
          department: values.department,
          jobLevel: values.jobLevel,
          employeeCount: 0,
          maxCapacity: Number(values.maxCapacity),
          minSalary: Number(values.minSalary),
          maxSalary: Number(values.maxSalary),
          currency: values.currency,
          isActive: values.isActive ?? true,
          createdAt: format(new Date(), 'yyyy-MM-dd'),
          updatedAt: format(new Date(), 'yyyy-MM-dd'),
          description: values.description,
          requirements: values.requirements,
          benefits: values.benefits
        };

        setPositions(prev => [...prev, newPosition]);
        
        toast.success(`职位 ${values.title} 已成功创建`);
      }
      
      handleModalClose();
    } catch (error) {
      toast.error('操作时发生错误，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (position: Position) => {
    setEditingPosition(position);
    setFormData(position);
    setIsModalVisible(true);
  };

  const handleDelete = (position: Position) => {
    if (position.employeeCount > 0) {
      toast.error(`职位 ${position.title} 下还有 ${position.employeeCount} 名员工，无法删除`);
      return;
    }

    if (confirm(`确定要删除职位 ${position.title} 吗？此操作不可撤销。`)) {
      setPositions(prev => prev.filter(pos => pos.id !== position.id));
      toast.success(`职位 ${position.title} 已从系统中删除`);
    }
  };

  const handleToggleStatus = (position: Position) => {
    const newStatus = !position.isActive;
    const statusText = newStatus ? '激活' : '停用';
    
    if (confirm(`确定要${statusText}职位 ${position.title} 吗？`)) {
      setPositions(prev => prev.map(pos => 
        pos.id === position.id 
          ? { ...pos, isActive: newStatus, updatedAt: format(new Date(), 'yyyy-MM-dd') }
          : pos
      ));
      toast.success(`职位 ${position.title} 已${statusText}`);
    }
  };

  const handleModalClose = () => {
    setIsModalVisible(false);
    setEditingPosition(null);
    setFormData({});
  };

  const getStatusColor = (isActive: boolean): "default" | "destructive" | "secondary" => {
    return isActive ? 'default' : 'secondary';
  };

  const getOccupancyColor = (rate: number): string => {
    if (rate >= 0.9) return 'text-red-600';
    if (rate >= 0.7) return 'text-yellow-600';
    return 'text-green-600';
  };

  const formatSalary = (min: number, max: number, currency: string) => {
    const formatter = new Intl.NumberFormat('zh-CN');
    const currencySymbol = currency === 'CNY' ? '¥' : '$';
    return `${currencySymbol}${formatter.format(min)} - ${currencySymbol}${formatter.format(max)}`;
  };

  const ActionsCell = ({ row }: { row: Position }) => (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 p-0">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => handleEdit(row)}>
          <Edit2 className="mr-2 h-4 w-4" />
          编辑职位
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => handleToggleStatus(row)}>
          <Star className="mr-2 h-4 w-4" />
          {row.isActive ? '停用职位' : '激活职位'}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem 
          onClick={() => handleDelete(row)} 
          className="text-destructive"
          disabled={row.employeeCount > 0}
        >
          <Trash2 className="mr-2 h-4 w-4" />
          删除职位
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

  const columns: ColumnDef<Position>[] = [
    {
      accessorKey: 'title',
      header: '职位信息',
      cell: ({ row }) => {
        const position = row.original;
        const occupancyRate = position.maxCapacity > 0 ? position.employeeCount / position.maxCapacity : 0;
        
        return (
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-purple-500 text-white flex items-center justify-center">
              <Briefcase className="h-5 w-5" />
            </div>
            <div>
              <div className="font-medium flex items-center gap-2">
                {position.title}
                <Badge variant={getStatusColor(position.isActive)}>
                  {position.isActive ? '招聘中' : '已停用'}
                </Badge>
              </div>
              <div className="text-sm text-gray-500 flex items-center gap-2">
                <span>{position.jobLevel}</span>
                <span>•</span>
                <span>{position.department}</span>
              </div>
            </div>
          </div>
        );
      },
    },
    {
      accessorKey: 'employeeCount',
      header: '人员配置',
      cell: ({ row }) => {
        const position = row.original;
        const occupancyRate = position.maxCapacity > 0 ? position.employeeCount / position.maxCapacity : 0;
        
        return (
          <div>
            <div className="flex items-center gap-2">
              <Users className="h-4 w-4 text-blue-500" />
              <span className="font-medium">
                {position.employeeCount} / {position.maxCapacity}
              </span>
            </div>
            <div className="text-sm">
              <span className={getOccupancyColor(occupancyRate)}>
                占用率 {(occupancyRate * 100).toFixed(0)}%
              </span>
            </div>
          </div>
        );
      },
    },
    {
      accessorKey: 'salary',
      header: '薪资范围',
      cell: ({ row }) => {
        const position = row.original;
        return (
          <div className="flex items-center gap-2">
            <DollarSign className="h-4 w-4 text-green-500" />
            <span className="font-medium">
              {formatSalary(position.minSalary, position.maxSalary, position.currency)}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'updatedAt',
      header: '最后更新',
      cell: ({ row }) => {
        const updatedAt = row.original.updatedAt;
        return (
          <div className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4 text-gray-400" />
            <span>{format(new Date(updatedAt), 'yyyy年MM月dd日', { locale: zhCN })}</span>
          </div>
        );
      },
    },
    createActionsColumn<Position>(ActionsCell),
  ];

  const departments = Array.from(new Set(positions.map(pos => pos.department).filter(Boolean)));
  const jobLevels = ['P3', 'P4', 'P5', 'P6', 'P7', 'P8', 'M1', 'M2', 'M3'];

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6 flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold">职位管理</h1>
          <p className="text-gray-600 mt-1">
            管理公司职位信息、人员配置和薪资体系 - 完整职位生命周期管理
          </p>
        </div>
        <Button 
          size="lg"
          onClick={() => setIsModalVisible(true)}
        >
          <Plus className="mr-2 h-4 w-4" />
          新增职位
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">总职位数</p>
                <p className="text-2xl font-bold">{positions.length}</p>
              </div>
              <Briefcase className="h-8 w-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">招聘中</p>
                <p className="text-2xl font-bold text-green-600">
                  {positions.filter(p => p.isActive).length}
                </p>
              </div>
              <Star className="h-8 w-8 text-green-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">在职人员</p>
                <p className="text-2xl font-bold">
                  {positions.reduce((sum, p) => sum + p.employeeCount, 0)}
                </p>
              </div>
              <Users className="h-8 w-8 text-purple-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">平均占用率</p>
                <p className="text-2xl font-bold text-orange-600">
                  {positions.length > 0 
                    ? Math.round(positions.reduce((sum, p) => 
                        sum + (p.maxCapacity > 0 ? p.employeeCount / p.maxCapacity : 0), 0
                      ) / positions.length * 100) 
                    : 0}%
                </p>
              </div>
              <TrendingUp className="h-8 w-8 text-orange-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Position Table */}
      <Card>
        <CardContent className="p-6">
          <DataTable
            columns={columns}
            data={positions}
            searchKey="title"
            searchPlaceholder="搜索职位名称、部门或级别..."
          />
        </CardContent>
      </Card>

      {/* Create/Edit Position Modal */}
      <Dialog open={isModalVisible} onOpenChange={setIsModalVisible}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {editingPosition ? '编辑职位信息' : '新增职位'}
            </DialogTitle>
          </DialogHeader>
          
          <div className="grid grid-cols-2 gap-6">
            {/* Left Column - Basic Info */}
            <div className="space-y-4">
              <h3 className="text-lg font-semibold">基本信息</h3>
              
              <div>
                <label className="text-sm font-medium">职位名称 *</label>
                <Input 
                  placeholder="如: 高级软件工程师"
                  value={formData.title || ''}
                  onChange={(e) => setFormData(prev => ({ ...prev, title: e.target.value }))}
                />
              </div>
              
              <div>
                <label className="text-sm font-medium">所属部门 *</label>
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
                    <SelectItem value="设计部">设计部</SelectItem>
                    <SelectItem value="人事部">人事部</SelectItem>
                    <SelectItem value="财务部">财务部</SelectItem>
                    <SelectItem value="市场部">市场部</SelectItem>
                    <SelectItem value="运营部">运营部</SelectItem>
                  </SelectContent>
                </Select>
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
                <label className="text-sm font-medium">最大容量 *</label>
                <Input 
                  type="number"
                  placeholder="如: 5"
                  value={formData.maxCapacity || ''}
                  onChange={(e) => setFormData(prev => ({ ...prev, maxCapacity: Number(e.target.value) }))}
                />
              </div>

              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="text-sm font-medium">最低薪资 *</label>
                  <Input 
                    type="number"
                    placeholder="12000"
                    value={formData.minSalary || ''}
                    onChange={(e) => setFormData(prev => ({ ...prev, minSalary: Number(e.target.value) }))}
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">最高薪资 *</label>
                  <Input 
                    type="number"
                    placeholder="20000"
                    value={formData.maxSalary || ''}
                    onChange={(e) => setFormData(prev => ({ ...prev, maxSalary: Number(e.target.value) }))}
                  />
                </div>
              </div>

              <div>
                <label className="text-sm font-medium">货币单位</label>
                <Select 
                  value={formData.currency || 'CNY'}
                  onValueChange={(value) => setFormData(prev => ({ ...prev, currency: value }))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="CNY">人民币 (CNY)</SelectItem>
                    <SelectItem value="USD">美元 (USD)</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div>
                <label className="text-sm font-medium">状态</label>
                <Select 
                  value={formData.isActive !== undefined ? String(formData.isActive) : 'true'}
                  onValueChange={(value) => setFormData(prev => ({ ...prev, isActive: value === 'true' }))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="true">招聘中</SelectItem>
                    <SelectItem value="false">已停用</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {/* Right Column - Detailed Info */}
            <div className="space-y-4">
              <h3 className="text-lg font-semibold">详细信息</h3>
              
              <div>
                <label className="text-sm font-medium">职位描述</label>
                <Textarea 
                  placeholder="描述该职位的主要职责和工作内容..."
                  rows={4}
                  value={formData.description || ''}
                  onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                />
              </div>

              <div>
                <label className="text-sm font-medium">任职要求</label>
                <Textarea 
                  placeholder="列出该职位所需的技能、经验和资质要求..."
                  rows={4}
                  value={formData.requirements || ''}
                  onChange={(e) => setFormData(prev => ({ ...prev, requirements: e.target.value }))}
                />
              </div>

              <div>
                <label className="text-sm font-medium">福利待遇</label>
                <Textarea 
                  placeholder="描述该职位的福利、奖金和其他待遇..."
                  rows={4}
                  value={formData.benefits || ''}
                  onChange={(e) => setFormData(prev => ({ ...prev, benefits: e.target.value }))}
                />
              </div>
            </div>
          </div>

          <div className="flex justify-end gap-2 mt-6">
            <Button variant="outline" onClick={handleModalClose}>
              取消
            </Button>
            <Button 
              onClick={() => handleCreatePosition(formData)} 
              disabled={loading}
            >
              {editingPosition ? '更新' : '创建'}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default PositionsPage;