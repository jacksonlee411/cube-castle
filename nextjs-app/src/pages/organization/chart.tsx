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

// Import modern SWR hooks, error boundary, and real API
import { useOrganizationsSWR, useOrganizationChartSWR, useOrganizationStatsSWR } from '@/hooks/useOrganizationsSWR';
import RESTErrorBoundary from '@/components/RESTErrorBoundary';
import { Organization, OrganizationCreateData } from '@/types';
import { organizationApi } from '@/lib/api-client';

const OrganizationChartPage: React.FC = () => {
  return (
    <RESTErrorBoundary
      resetOnPropsChange={true}
      onError={(error: Error, errorInfo: React.ErrorInfo) => {
        console.error('🛡️ Organization Chart Error:', {
          error: error.message,
          stack: error.stack,
          componentStack: errorInfo.componentStack,
          type: 'ORGANIZATION_CHART_ERROR',
          timestamp: new Date().toISOString(),
        });
      }}
    >
      <OrganizationChartContent />
    </RESTErrorBoundary>
  );
};

// Separate content component for better error boundary isolation
const OrganizationChartContent: React.FC = () => {
  const router = useRouter();
  
  // Modern SWR data fetching (replacing useEffect)
  const { 
    organizations, 
    totalCount, 
    isLoading, 
    isError, 
    error,
    mutate 
  } = useOrganizationsSWR();
  
  const { 
    chart, 
    flatChart, 
    isLoading: isChartLoading 
  } = useOrganizationChartSWR();
  
  const { 
    stats, 
    typeData, 
    isLoading: isStatsLoading 
  } = useOrganizationStatsSWR();

  // UI state management
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingOrganization, setEditingOrganization] = useState<Organization | null>(null);
  const [selectedParentId, setSelectedParentId] = useState<string | undefined>(undefined);
  const [formData, setFormData] = useState<Partial<OrganizationCreateData>>({});

  // Initialize expanded nodes when data loads
  useEffect(() => {
    if (chart.length > 0) {
      // Auto-expand first two levels for better UX
      const defaultExpanded = new Set<string>();
      flatChart.forEach(org => {
        if (org.level <= 1) {
          defaultExpanded.add(org.id);
        }
      });
      setExpandedNodes(defaultExpanded);
    }
  }, [chart, flatChart]);

  // Create/Update organization using real PostgreSQL API (aligned with backend model)
  const handleCreateOrganization = async (values: OrganizationCreateData) => {
    try {
      if (editingOrganization) {
        // Update existing organization via PostgreSQL API
        console.log('📝 更新组织:', editingOrganization.id, values);
        
        await organizationApi.updateOrganization(editingOrganization.id, values);
        
        // Revalidate SWR data to refresh UI
        mutate();
        toast.success(`组织 ${values.name} 信息已更新`);
      } else {
        // Create new organization via PostgreSQL API  
        console.log('🎯 创建新组织 (backend model):', values);
        
        const newOrg = await organizationApi.createOrganization(values);
        
        // Revalidate SWR data to refresh UI immediately
        mutate();
        
        console.log('🎉 组织创建成功:', newOrg.name, '(ID:', newOrg.id, ')');
        toast.success(`组织 ${values.name} 已成功创建`);
      }
      
      handleModalClose();
    } catch (error) {
      // If something fails, revalidate to ensure UI is consistent
      mutate();
      toast.error('操作时发生错误，请重试');
      console.error('Organization operation failed:', error);
    }
  };

  const calculateLevel = (parentId?: string): number => {
    if (!parentId) return 0;
    const parent = flatChart.find(org => org.id === parentId);
    return parent ? parent.level + 1 : 0;
  };

  const handleEdit = (organization: Organization) => {
    setEditingOrganization(organization);
    setFormData(organization);
    setIsModalVisible(true);
  };

  const handleDelete = async (organization: Organization) => {
    const hasChildren = flatChart.some((org: Organization) => org.parent_unit_id === organization.id);
    
    if (hasChildren) {
      toast.error(`组织 ${organization.name} 下还有子部门，无法删除`);
      return;
    }
    
    if ((organization.employee_count || 0) > 0) {
      toast.error(`组织 ${organization.name} 下还有 ${organization.employee_count} 名员工，无法删除`);
      return;
    }

    if (confirm(`确定要删除组织 ${organization.name} 吗？此操作不可撤销。`)) {
      try {
        console.log('🗑️ 删除组织:', organization.id, organization.name);
        
        // Delete via PostgreSQL API
        await organizationApi.deleteOrganization(organization.id);
        
        // Revalidate data
        mutate();
        toast.success(`组织 ${organization.name} 已从系统中删除`);
      } catch (error) {
        console.error('删除组织失败:', error);
        toast.error('删除操作失败，请重试');
      }
    }
  };

  const handleAddChild = (parentOrg: Organization) => {
    setSelectedParentId(parentOrg.id);
    setFormData({ 
      parent_unit_id: parentOrg.id,
      unit_type: getDefaultChildType(parentOrg.unit_type),
      status: 'ACTIVE'
    });
    setIsModalVisible(true);
  };

  const getDefaultChildType = (parentUnitType: Organization['unit_type']): Organization['unit_type'] => {
    switch (parentUnitType) {
      case 'COMPANY': return 'DEPARTMENT';
      case 'DEPARTMENT': return 'PROJECT_TEAM';
      case 'PROJECT_TEAM': return 'COST_CENTER';
      default: return 'DEPARTMENT';
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
    const allIds = new Set(flatChart.map((org: Organization) => org.id));
    setExpandedNodes(allIds);
  };

  const collapseAll = () => {
    setExpandedNodes(new Set());
  };

  const getTypeColor = (unitType: Organization['unit_type']) => {
    const colors = {
      COMPANY: 'bg-blue-500',
      DEPARTMENT: 'bg-purple-500', 
      PROJECT_TEAM: 'bg-green-500',
      COST_CENTER: 'bg-orange-500'
    };
    return colors[unitType] || 'bg-gray-500';
  };

  const getTypeLabel = (unitType: Organization['unit_type']) => {
    const labels = {
      COMPANY: '公司',
      DEPARTMENT: '部门',
      PROJECT_TEAM: '项目团队',
      COST_CENTER: '成本中心'
    };
    return labels[unitType] || unitType;
  };

  const getOccupancyColor = (rate: number): string => {
    if (rate >= 0.9) return 'text-red-600';
    if (rate >= 0.7) return 'text-yellow-600';
    return 'text-green-600';
  };

  const OrgNodeActions = ({ org }: { org: Organization }) => (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button 
          variant="ghost" 
          size="sm" 
          className="h-6 w-6 p-0"
          data-testid={`org-action-menu-${org.id}`}
        >
          <MoreHorizontal className="h-3 w-3" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" data-testid={`org-action-menu-content-${org.id}`}>
        <DropdownMenuItem 
          onClick={() => handleEdit(org)}
          data-testid={`edit-org-${org.id}`}
        >
          <Edit2 className="mr-2 h-3 w-3" />
          编辑组织
        </DropdownMenuItem>
        <DropdownMenuItem 
          onClick={() => handleAddChild(org)}
          data-testid={`add-child-${org.id}`}
        >
          <UserPlus className="mr-2 h-3 w-3" />
          添加子部门
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem 
          onClick={() => handleDelete(org)} 
          className="text-destructive"
          disabled={(org.employee_count || 0) > 0}
          data-testid={`delete-org-${org.id}`}
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
    const occupancyRate = org.profile?.maxCapacity ? (org.employee_count || 0) / org.profile.maxCapacity : 0;
    
    return (
      <div key={org.id} className="mb-2">
        {/* Organization Node */}
        <div 
          className={`relative flex items-center p-3 bg-white border rounded-lg shadow-sm hover:shadow-md transition-shadow ${
            depth > 0 ? 'ml-8' : ''
          }`}
          style={{ marginLeft: depth * 24 }}
          data-testid={`org-node-${org.id}`}
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
          <div className={`w-8 h-8 rounded-full ${getTypeColor(org.unit_type)} text-white flex items-center justify-center mr-3`}>
            <Building2 className="h-4 w-4" />
          </div>
          
          {/* Organization Info */}
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              <h3 className="font-medium text-sm">{org.name}</h3>
              <Badge variant="outline" className="text-xs">
                {getTypeLabel(org.unit_type)}
              </Badge>
              {org.status === 'INACTIVE' && (
                <Badge variant="secondary" className="text-xs">
                  已停用
                </Badge>
              )}
            </div>
            
            <div className="flex items-center gap-4 text-xs text-gray-500">
              {org.profile?.managerName && (
                <div className="flex items-center gap-1">
                  <Crown className="h-3 w-3" />
                  <span>{org.profile.managerName}</span>
                </div>
              )}
              
              <div className="flex items-center gap-1">
                <Users className="h-3 w-3" />
                <span>{org.employee_count || 0}</span>
                {org.profile?.maxCapacity && (
                  <>
                    <span>/</span>
                    <span>{org.profile.maxCapacity}</span>
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
    { value: 'COMPANY', label: '公司' },
    { value: 'DEPARTMENT', label: '部门' },
    { value: 'PROJECT_TEAM', label: '项目团队' },
    { value: 'COST_CENTER', label: '成本中心' }
  ];

  const getParentOptions = () => {
    return flatChart
      .filter(org => org.id !== editingOrganization?.id)
      .map(org => ({
        value: org.id,
        label: `${org.name} (${getTypeLabel(org.unit_type)}) - L${org.level}`
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
                <p className="text-2xl font-bold">{stats.total}</p>
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
                  {stats.totalEmployees}
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
                  {stats.maxLevel + 1}
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
                <p className="text-sm text-gray-600">活跃组织</p>
                <p className="text-2xl font-bold text-orange-600">
                  {stats.active}
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
          {isLoading || isChartLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="text-gray-500">加载中...</div>
            </div>
          ) : chart.length > 0 ? (
            <div className="space-y-2" data-testid="org-tree">
              {chart.map((org: Organization) => renderOrgNode(org))}
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
                value={formData.unit_type || ''}
                onValueChange={(value) => setFormData(prev => ({ ...prev, unit_type: value as Organization['unit_type'] }))}
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
                value={formData.parent_unit_id || 'none'}
                onValueChange={(value) => setFormData(prev => ({ ...prev, parent_unit_id: value === 'none' ? undefined : value }))}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择上级组织" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">无上级组织</SelectItem>
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
                value={formData.profile?.managerName || ''}
                onChange={(e) => setFormData(prev => ({ 
                  ...prev, 
                  profile: { ...prev.profile, managerName: e.target.value }
                }))}
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">最大容量</label>
              <Input 
                type="number"
                placeholder="如: 20"
                value={formData.profile?.maxCapacity || ''}
                onChange={(e) => setFormData(prev => ({ 
                  ...prev, 
                  profile: { ...prev.profile, maxCapacity: Number(e.target.value) }
                }))}
              />
            </div>

            <div>
              <label className="text-sm font-medium">状态</label>
              <Select 
                value={formData.status || 'ACTIVE'}
                onValueChange={(value) => setFormData(prev => ({ ...prev, status: value as Organization['status'] }))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="ACTIVE">正常运营</SelectItem>
                  <SelectItem value="INACTIVE">已停用</SelectItem>
                  <SelectItem value="PLANNED">计划中</SelectItem>
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
              onClick={() => {
                if (formData.name && formData.unit_type) {
                  handleCreateOrganization(formData as OrganizationCreateData);
                }
              }} 
              disabled={isLoading || !formData.name || !formData.unit_type}
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