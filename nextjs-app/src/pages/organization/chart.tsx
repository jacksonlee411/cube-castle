import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { 
  Plus, 
  MoreHorizontal,
  Building2,
  Users,
  Crown,
  Layers,
  Edit2,
  Trash2,
  UserPlus,
  ArrowUp,
  ArrowDown,
  Expand,
  Minimize
} from 'lucide-react';
import { toast } from 'react-hot-toast';

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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Alert, AlertDescription } from '@/components/ui/alert';

interface Organization {
  id: string;
  name: string;
  type: 'company' | 'department' | 'team' | 'group';
  parentId?: string;
  level: number;
  managerId?: string;
  managerName?: string;
  employeeCount: number;
  maxCapacity?: number;
  description?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  children?: Organization[];
}

const OrganizationChartPage: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [flatOrganizations, setFlatOrganizations] = useState<Organization[]>([]);
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingOrganization, setEditingOrganization] = useState<Organization | null>(null);
  const [selectedParentId, setSelectedParentId] = useState<string | undefined>(undefined);
  const [formData, setFormData] = useState<Partial<Organization>>({});

  // Sample data
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      const sampleOrganizations: Organization[] = [
        {
          id: '1',
          name: 'Cube Castle 科技有限公司',
          type: 'company',
          level: 0,
          managerId: 'CEO001',
          managerName: '张伟',
          employeeCount: 45,
          maxCapacity: 60,
          description: '专业的企业人力资源管理平台',
          isActive: true,
          createdAt: '2020-01-01',
          updatedAt: '2024-12-01'
        },
        {
          id: '2',
          name: '技术部',
          type: 'department',
          parentId: '1',
          level: 1,
          managerId: 'TEC001',
          managerName: '李强',
          employeeCount: 18,
          maxCapacity: 25,
          description: '负责产品研发和技术创新',
          isActive: true,
          createdAt: '2020-01-01',
          updatedAt: '2024-11-15'
        },
        {
          id: '3',
          name: '产品部',
          type: 'department',
          parentId: '1',
          level: 1,
          managerId: 'PRD001',
          managerName: '王敏',
          employeeCount: 8,
          maxCapacity: 12,
          description: '负责产品规划和用户体验',
          isActive: true,
          createdAt: '2020-06-01',
          updatedAt: '2024-10-20'
        },
        {
          id: '4',
          name: '人事部',
          type: 'department',
          parentId: '1',
          level: 1,
          managerId: 'HR001',
          managerName: '陈静',
          employeeCount: 5,
          maxCapacity: 8,
          description: '负责人力资源管理和企业文化建设',
          isActive: true,
          createdAt: '2020-01-01',
          updatedAt: '2024-09-30'
        },
        {
          id: '5',
          name: '前端开发团队',
          type: 'team',
          parentId: '2',
          level: 2,
          managerId: 'FE001',
          managerName: '刘洋',
          employeeCount: 6,
          maxCapacity: 8,
          description: '负责前端产品开发和用户界面',
          isActive: true,
          createdAt: '2020-03-01',
          updatedAt: '2024-11-01'
        },
        {
          id: '6',
          name: '后端开发团队',
          type: 'team',
          parentId: '2',
          level: 2,
          managerId: 'BE001',
          managerName: '赵磊',
          employeeCount: 8,
          maxCapacity: 10,
          description: '负责后端服务开发和系统架构',
          isActive: true,
          createdAt: '2020-03-01',
          updatedAt: '2024-10-15'
        },
        {
          id: '7',
          name: 'DevOps团队',
          type: 'team',
          parentId: '2',
          level: 2,
          managerId: 'OPS001',
          managerName: '孙杰',
          employeeCount: 4,
          maxCapacity: 6,
          description: '负责基础设施和运维自动化',
          isActive: true,
          createdAt: '2021-01-01',
          updatedAt: '2024-09-20'
        },
        {
          id: '8',
          name: '产品策划组',
          type: 'group',
          parentId: '3',
          level: 2,
          managerId: 'PM001',
          managerName: '周莉',
          employeeCount: 4,
          maxCapacity: 6,
          description: '负责产品需求分析和功能规划',
          isActive: true,
          createdAt: '2020-08-01',
          updatedAt: '2024-08-10'
        },
        {
          id: '9',
          name: 'UX设计组',
          type: 'group',
          parentId: '3',
          level: 2,
          managerId: 'UX001',
          managerName: '吴芳',
          employeeCount: 4,
          maxCapacity: 6,
          description: '负责用户体验设计和交互设计',
          isActive: true,
          createdAt: '2020-10-01',
          updatedAt: '2024-07-25'
        }
      ];
      
      // Build tree structure
      const organizationMap = new Map<string, Organization>();
      const rootOrganizations: Organization[] = [];
      
      sampleOrganizations.forEach(org => {
        organizationMap.set(org.id, { ...org, children: [] });
      });
      
      sampleOrganizations.forEach(org => {
        const orgWithChildren = organizationMap.get(org.id)!;
        if (org.parentId) {
          const parent = organizationMap.get(org.parentId);
          if (parent) {
            parent.children!.push(orgWithChildren);
          }
        } else {
          rootOrganizations.push(orgWithChildren);
        }
      });
      
      setOrganizations(rootOrganizations);
      setFlatOrganizations(sampleOrganizations);
      
      // 默认展开前两层
      const defaultExpanded = new Set<string>();
      sampleOrganizations.forEach(org => {
        if (org.level <= 1) {
          defaultExpanded.add(org.id);
        }
      });
      setExpandedNodes(defaultExpanded);
      
      setLoading(false);
    }, 1000);
  }, []);

  const handleCreateOrganization = async (values: any) => {
    try {
      setLoading(true);
      
      if (editingOrganization) {
        // Update existing organization
        const updatedOrg: Organization = {
          ...editingOrganization,
          name: values.name,
          type: values.type,
          parentId: values.parentId,
          level: calculateLevel(values.parentId),
          managerId: values.managerId,
          managerName: values.managerName,
          maxCapacity: Number(values.maxCapacity),
          description: values.description,
          isActive: values.isActive,
          updatedAt: new Date().toISOString().split('T')[0]
        };

        setFlatOrganizations(prev => prev.map(org => 
          org.id === editingOrganization.id ? updatedOrg : org
        ));

        toast.success(`组织 ${values.name} 信息已更新`);
      } else {
        // Create new organization
        const newOrg: Organization = {
          id: Date.now().toString(),
          name: values.name,
          type: values.type,
          parentId: values.parentId,
          level: calculateLevel(values.parentId),
          managerId: values.managerId,
          managerName: values.managerName,
          employeeCount: 0,
          maxCapacity: Number(values.maxCapacity),
          description: values.description,
          isActive: values.isActive ?? true,
          createdAt: new Date().toISOString().split('T')[0],
          updatedAt: new Date().toISOString().split('T')[0]
        };

        setFlatOrganizations(prev => [...prev, newOrg]);
        
        toast.success(`组织 ${values.name} 已成功创建`);
      }
      
      // Rebuild tree
      rebuildTree();
      handleModalClose();
    } catch (error) {
      toast.error('操作时发生错误，请重试');
    } finally {
      setLoading(false);
    }
  };

  const calculateLevel = (parentId?: string): number => {
    if (!parentId) return 0;
    const parent = flatOrganizations.find(org => org.id === parentId);
    return parent ? parent.level + 1 : 0;
  };

  const rebuildTree = () => {
    const organizationMap = new Map<string, Organization>();
    const rootOrganizations: Organization[] = [];
    
    flatOrganizations.forEach(org => {
      organizationMap.set(org.id, { ...org, children: [] });
    });
    
    flatOrganizations.forEach(org => {
      const orgWithChildren = organizationMap.get(org.id)!;
      if (org.parentId) {
        const parent = organizationMap.get(org.parentId);
        if (parent) {
          parent.children!.push(orgWithChildren);
        }
      } else {
        rootOrganizations.push(orgWithChildren);
      }
    });
    
    setOrganizations(rootOrganizations);
  };

  const handleEdit = (organization: Organization) => {
    setEditingOrganization(organization);
    setFormData(organization);
    setIsModalVisible(true);
  };

  const handleDelete = (organization: Organization) => {
    const hasChildren = flatOrganizations.some(org => org.parentId === organization.id);
    
    if (hasChildren) {
      toast.error(`组织 ${organization.name} 下还有子部门，无法删除`);
      return;
    }
    
    if (organization.employeeCount > 0) {
      toast.error(`组织 ${organization.name} 下还有 ${organization.employeeCount} 名员工，无法删除`);
      return;
    }

    if (confirm(`确定要删除组织 ${organization.name} 吗？此操作不可撤销。`)) {
      setFlatOrganizations(prev => prev.filter(org => org.id !== organization.id));
      rebuildTree();
      toast.success(`组织 ${organization.name} 已从系统中删除`);
    }
  };

  const handleAddChild = (parentOrg: Organization) => {
    setSelectedParentId(parentOrg.id);
    setFormData({ 
      parentId: parentOrg.id,
      type: getDefaultChildType(parentOrg.type),
      isActive: true 
    });
    setIsModalVisible(true);
  };

  const getDefaultChildType = (parentType: Organization['type']): Organization['type'] => {
    switch (parentType) {
      case 'company': return 'department';
      case 'department': return 'team';
      case 'team': return 'group';
      default: return 'group';
    }
  };

  const handleModalClose = () => {
    setIsModalVisible(false);
    setEditingOrganization(null);
    setSelectedParentId(undefined);
    setFormData({});
  };

  const toggleExpanded = (nodeId: string) => {
    const newExpanded = new Set(expandedNodes);
    if (newExpanded.has(nodeId)) {
      newExpanded.delete(nodeId);
    } else {
      newExpanded.add(nodeId);
    }
    setExpandedNodes(newExpanded);
  };

  const expandAll = () => {
    const allIds = new Set(flatOrganizations.map(org => org.id));
    setExpandedNodes(allIds);
  };

  const collapseAll = () => {
    setExpandedNodes(new Set());
  };

  const getTypeColor = (type: Organization['type']) => {
    const colors = {
      company: 'bg-blue-500',
      department: 'bg-purple-500', 
      team: 'bg-green-500',
      group: 'bg-orange-500'
    };
    return colors[type] || 'bg-gray-500';
  };

  const getTypeLabel = (type: Organization['type']) => {
    const labels = {
      company: '公司',
      department: '部门',
      team: '团队',
      group: '小组'
    };
    return labels[type] || type;
  };

  const getOccupancyColor = (rate: number): string => {
    if (rate >= 0.9) return 'text-red-600';
    if (rate >= 0.7) return 'text-yellow-600';
    return 'text-green-600';
  };

  const OrgNodeActions = ({ org }: { org: Organization }) => (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0">
          <MoreHorizontal className="h-3 w-3" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => handleEdit(org)}>
          <Edit2 className="mr-2 h-3 w-3" />
          编辑组织
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => handleAddChild(org)}>
          <UserPlus className="mr-2 h-3 w-3" />
          添加子部门
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem 
          onClick={() => handleDelete(org)} 
          className="text-destructive"
          disabled={org.employeeCount > 0}
        >
          <Trash2 className="mr-2 h-3 w-3" />
          删除组织
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

  const renderOrgNode = (org: Organization, depth: number = 0) => {
    const hasChildren = org.children && org.children.length > 0;
    const isExpanded = expandedNodes.has(org.id);
    const occupancyRate = org.maxCapacity ? org.employeeCount / org.maxCapacity : 0;
    
    return (
      <div key={org.id} className="mb-2">
        {/* Organization Node */}
        <div 
          className={`relative flex items-center p-3 bg-white border rounded-lg shadow-sm hover:shadow-md transition-shadow ${
            depth > 0 ? 'ml-8' : ''
          }`}
          style={{ marginLeft: depth * 24 }}
        >
          {/* Connection Lines */}
          {depth > 0 && (
            <>
              <div className="absolute -left-6 top-1/2 w-6 h-px bg-gray-300"></div>
              <div className="absolute -left-6 -top-3 w-px h-6 bg-gray-300"></div>
            </>
          )}
          
          {/* Expand/Collapse Button */}
          {hasChildren && (
            <Button
              variant="ghost"
              size="sm"
              className="h-6 w-6 p-0 mr-2"
              onClick={() => toggleExpanded(org.id)}
            >
              {isExpanded ? (
                <ArrowDown className="h-3 w-3" />
              ) : (
                <ArrowUp className="h-3 w-3" />
              )}
            </Button>
          )}
          
          {/* Organization Icon */}
          <div className={`w-8 h-8 rounded-full ${getTypeColor(org.type)} text-white flex items-center justify-center mr-3`}>
            <Building2 className="h-4 w-4" />
          </div>
          
          {/* Organization Info */}
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              <h3 className="font-medium text-sm">{org.name}</h3>
              <Badge variant="outline" className="text-xs">
                {getTypeLabel(org.type)}
              </Badge>
              {!org.isActive && (
                <Badge variant="secondary" className="text-xs">
                  已停用
                </Badge>
              )}
            </div>
            
            <div className="flex items-center gap-4 text-xs text-gray-500">
              {org.managerName && (
                <div className="flex items-center gap-1">
                  <Crown className="h-3 w-3" />
                  <span>{org.managerName}</span>
                </div>
              )}
              
              <div className="flex items-center gap-1">
                <Users className="h-3 w-3" />
                <span>{org.employeeCount}</span>
                {org.maxCapacity && (
                  <>
                    <span>/</span>
                    <span>{org.maxCapacity}</span>
                    <span className={getOccupancyColor(occupancyRate)}>
                      ({(occupancyRate * 100).toFixed(0)}%)
                    </span>
                  </>
                )}
              </div>
              
              <div className="flex items-center gap-1">
                <Layers className="h-3 w-3" />
                <span>L{org.level}</span>
              </div>
            </div>
          </div>
          
          {/* Actions */}
          <OrgNodeActions org={org} />
        </div>
        
        {/* Children */}
        {hasChildren && isExpanded && org.children && (
          <div className="mt-2">
            {org.children.map(child => renderOrgNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  const organizationTypes = [
    { value: 'company', label: '公司' },
    { value: 'department', label: '部门' },
    { value: 'team', label: '团队' },
    { value: 'group', label: '小组' }
  ];

  const getParentOptions = () => {
    return flatOrganizations
      .filter(org => org.id !== editingOrganization?.id)
      .map(org => ({
        value: org.id,
        label: `${org.name} (${getTypeLabel(org.type)}) - L${org.level}`
      }));
  };

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6 flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold">组织架构图</h1>
          <p className="text-gray-600 mt-1">
            可视化组织结构管理 - 支持层级展示、拖拽编辑和人员配置
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={expandAll}>
            <Expand className="mr-2 h-4 w-4" />
            全部展开
          </Button>
          <Button variant="outline" size="sm" onClick={collapseAll}>
            <Minimize className="mr-2 h-4 w-4" />
            全部收起
          </Button>
          <Button onClick={() => setIsModalVisible(true)}>
            <Plus className="mr-2 h-4 w-4" />
            新增组织
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">组织总数</p>
                <p className="text-2xl font-bold">{flatOrganizations.length}</p>
              </div>
              <Building2 className="h-8 w-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">总员工数</p>
                <p className="text-2xl font-bold text-green-600">
                  {flatOrganizations.reduce((sum, org) => sum + org.employeeCount, 0)}
                </p>
              </div>
              <Users className="h-8 w-8 text-green-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">最大层级</p>
                <p className="text-2xl font-bold text-purple-600">
                  {Math.max(...flatOrganizations.map(org => org.level)) + 1}
                </p>
              </div>
              <Layers className="h-8 w-8 text-purple-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">平均占用率</p>
                <p className="text-2xl font-bold text-orange-600">
                  {flatOrganizations.length > 0 
                    ? Math.round(flatOrganizations
                        .filter(org => org.maxCapacity)
                        .reduce((sum, org) => 
                          sum + (org.maxCapacity! > 0 ? org.employeeCount / org.maxCapacity! : 0), 0
                        ) / flatOrganizations.filter(org => org.maxCapacity).length * 100) 
                    : 0}%
                </p>
              </div>
              <Crown className="h-8 w-8 text-orange-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Organization Tree */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            组织架构树
          </CardTitle>
        </CardHeader>
        <CardContent className="p-6">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <div className="text-gray-500">加载中...</div>
            </div>
          ) : organizations.length > 0 ? (
            <div className="space-y-2">
              {organizations.map(org => renderOrgNode(org))}
            </div>
          ) : (
            <Alert>
              <AlertDescription>
                暂无组织架构数据，请先创建组织。
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>

      {/* Create/Edit Organization Modal */}
      <Dialog open={isModalVisible} onOpenChange={setIsModalVisible}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {editingOrganization ? '编辑组织信息' : '新增组织'}
            </DialogTitle>
          </DialogHeader>
          
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium">组织名称 *</label>
              <Input 
                placeholder="如: 技术部"
                value={formData.name || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">组织类型 *</label>
              <Select 
                value={formData.type || ''}
                onValueChange={(value) => setFormData(prev => ({ ...prev, type: value as Organization['type'] }))}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择类型" />
                </SelectTrigger>
                <SelectContent>
                  {organizationTypes.map(type => (
                    <SelectItem key={type.value} value={type.value}>
                      {type.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <label className="text-sm font-medium">上级组织</label>
              <Select 
                value={formData.parentId || ''}
                onValueChange={(value) => setFormData(prev => ({ ...prev, parentId: value || undefined }))}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择上级组织" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">无上级组织</SelectItem>
                  {getParentOptions().map(option => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <label className="text-sm font-medium">负责人姓名</label>
              <Input 
                placeholder="负责人姓名"
                value={formData.managerName || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, managerName: e.target.value }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">最大容量</label>
              <Input 
                type="number"
                placeholder="如: 20"
                value={formData.maxCapacity || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, maxCapacity: Number(e.target.value) }))}
              />
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
                  <SelectItem value="true">正常运营</SelectItem>
                  <SelectItem value="false">已停用</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div>
            <label className="text-sm font-medium">组织描述</label>
            <Textarea 
              placeholder="描述该组织的主要职能和职责..."
              rows={3}
              value={formData.description || ''}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
            />
          </div>

          <div className="flex justify-end gap-2 mt-6">
            <Button variant="outline" onClick={handleModalClose}>
              取消
            </Button>
            <Button 
              onClick={() => handleCreateOrganization(formData)} 
              disabled={loading}
            >
              {editingOrganization ? '更新' : '创建'}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default OrganizationChartPage;