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
  Minimize,
  RefreshCw,
  Search,
  Filter,
  ChevronDown,
  ChevronRight
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

// Import CQRS hooks and components
import { useOrganizationCQRS, useOrganizationTree, useOrganizationStats } from '@/hooks/useOrganizationCQRS';
import { useAutoRefresh } from '@/hooks/useAutoRefresh';
import RESTErrorBoundary from '@/components/RESTErrorBoundary';
import { Organization, CreateOrganizationRequest } from '@/types';

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
  
  // CQRS data fetching - unified state management
  const { 
    organizations,
    orgChart,
    orgStats,
    isLoading,
    isRefreshing,
    hasErrors,
    errors,
    filteredOrganizations,
    organizationTree,
    createOrganization,
    updateOrganization,
    deleteOrganization,
    refreshAll,
    searchQuery,
    setSearchQuery,
    filters,
    setFilters,
    viewMode,
    setViewMode
  } = useOrganizationCQRS();
  
  // Tree-specific operations
  const {
    expandedNodes,
    selectedOrganization,
    toggleNodeExpansion,
    selectOrganization,
    expandAll,
    collapseAll,
    isNodeExpanded
  } = useOrganizationTree();
  
  // Stats with specialized hook
  const { stats: liveStats, refresh: refreshStats } = useOrganizationStats();

  // 自动刷新功能 (替代WebSocket实时更新)
  useAutoRefresh(refreshAll, {
    interval: 30000,        // 30秒自动刷新
    enabled: true,          // 默认启用
    enableOnFocus: true,    // 窗口获得焦点时刷新
    enableOnVisible: true,  // 页面可见时刷新
  });

  // UI state management (local state only for modal and form)
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingOrganization, setEditingOrganization] = useState<Organization | null>(null);
  const [selectedParentId, setSelectedParentId] = useState<string | undefined>(undefined);
  const [formData, setFormData] = useState<Partial<CreateOrganizationRequest>>({});

  // Use the current stats data (prioritize live stats, fallback to store stats)
  // 统一数据格式映射函数
  const mapStatsFormat = (stats: any) => ({
    total: stats?.total_organizations || 0,
    active: stats?.active_organizations || 0,
    inactive: (stats?.total_organizations || 0) - (stats?.active_organizations || 0),
    totalEmployees: stats?.total_employees || 0,
    maxLevel: stats?.max_depth || 0
  });
  
  const currentStats = liveStats ? mapStatsFormat(liveStats) 
    : orgStats ? mapStatsFormat(orgStats) 
    : {
      total: 0,
      active: 0,
      inactive: 0,
      totalEmployees: 0,
      maxLevel: 0
    };

  // Use organizationTree from CQRS hook instead of chart data
  const currentOrgTree = organizationTree.length > 0 ? organizationTree : orgChart;

  // Initialize expanded nodes when data loads - now managed by CQRS store
  useEffect(() => {
    if (currentOrgTree.length > 0) {
      // Auto-expand first two levels for better UX
      const defaultExpanded = new Set<string>();
      organizations.forEach(org => {
        if (org.level <= 1) {
          defaultExpanded.add(org.id);
        }
      });
      // Only expand if not already managed by store
      defaultExpanded.forEach(id => {
        if (!isNodeExpanded(id)) {
          toggleNodeExpansion(id);
        }
      });
    }
  }, [currentOrgTree, organizations]);

  // Create/Update organization using CQRS commands with optimistic updates
  const handleCreateOrganization = async (values: CreateOrganizationRequest) => {
    try {
      if (editingOrganization) {
        // Update existing organization via CQRS command
        console.log('📝 更新组织 (CQRS):', editingOrganization.id, values);
        
        const result = await updateOrganization(editingOrganization.id, values);
        
        if (result) {
          console.log('🎉 组织更新成功 (CQRS):', result.name, '(ID:', result.id, ')');
          toast.success(`组织 ${values.name} 信息已更新`);
        } else {
          throw new Error('更新失败');
        }
      } else {
        // Create new organization via CQRS command  
        console.log('🎯 创建新组织 (CQRS):', values);
        
        const result = await createOrganization(values);
        
        if (result) {
          console.log('🎉 组织创建成功 (CQRS):', result.name, '(ID:', result.id, ')');
          toast.success(`组织 ${values.name} 已成功创建`);
        } else {
          throw new Error('创建失败');
        }
      }
      
      handleModalClose();
    } catch (error) {
      // Error handling is managed by CQRS store
      console.error('Organization operation failed:', error);
      // Toast already shown by CQRS store
    }
  };

  const calculateLevel = (parentId?: string): number => {
    if (!parentId) return 0;
    const parent = organizations.find(org => org.id === parentId);
    return parent ? parent.level + 1 : 0;
  };

  const handleEdit = (organization: Organization) => {
    setEditingOrganization(organization);
    setFormData(organization);
    setIsModalVisible(true);
  };

  const handleDelete = async (organization: Organization) => {
    const hasChildren = organizations.some((org: Organization) => org.parent_unit_id === organization.id);
    
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
        console.log('🗑️ 删除组织 (CQRS):', organization.id, organization.name);
        
        // Delete via CQRS command with optimistic update
        const success = await deleteOrganization(organization.id);
        
        if (success) {
          console.log('✅ 组织删除成功 (CQRS)');
          // Toast already shown by CQRS store
        } else {
          throw new Error('删除失败');
        }
      } catch (error) {
        console.error('删除组织失败:', error);
        // Error toast already handled by CQRS store
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

  // Use CQRS store managed expand/collapse
  const handleExpandAll = () => expandAll();
  const handleCollapseAll = () => collapseAll();

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
    const isExpanded = isNodeExpanded(org.id);
    const occupancyRate = org.profile?.maxCapacity ? (org.employee_count || 0) / org.profile.maxCapacity : 0;
    
    return (
      <div key={org.id} className="mb-2">
        {/* Organization Node - 修复：正确的层级缩进显示 */}
        <div 
          className={`relative flex items-center p-3 bg-white border rounded-lg shadow-sm hover:shadow-md transition-shadow ${
            selectedOrganization?.id === org.id ? 'ring-2 ring-blue-500' : ''
          }`}
          style={{ marginLeft: depth * 32 }} // 修复：增加缩进量使层级更明显
          data-testid={`org-node-${org.id}`}
          onClick={() => selectOrganization(org)}
        >
          {/* Connection Lines - 修复：层级连接线 */}
          {depth > 0 && (
            <>
              <div className="absolute -left-8 top-1/2 w-8 h-px bg-gray-300"></div>
              <div className="absolute -left-8 -top-3 w-px h-6 bg-gray-300"></div>
            </>
          )}
          
          {/* Expand/Collapse Button */}
          {hasChildren && (
            <Button
              variant="ghost"
              size="sm"
              className="h-6 w-6 p-0 mr-2"
              onClick={(e) => {
                e.stopPropagation();
                toggleNodeExpansion(org.id);
              }}
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
              {/* 层级显示徽章 */}
              <Badge variant="secondary" className="text-xs">
                L{org.level}
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
                <span className="text-gray-400">·</span>
                <span className="text-gray-400">深度{depth}</span>
              </div>
            </div>
          </div>
          
          {/* Actions */}
          <OrgNodeActions org={org} />
        </div>
        
        {/* Children - 修复：递归渲染时正确传递depth+1 */}
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
    return organizations
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
            CQRS 架构 - 支持实时更新、乐观处理和智能缓存
          </p>
        </div>
        <div className="flex gap-2">
          {/* Refresh Controls */}
          <Button 
            variant="outline" 
            size="sm" 
            onClick={refreshAll}
            disabled={isRefreshing}
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
            {isRefreshing ? '刷新中...' : '刷新数据'}
          </Button>
          
          {/* View Mode Controls */}
          <Button variant="outline" size="sm" onClick={handleExpandAll}>
            <Expand className="mr-2 h-4 w-4" />
            全部展开
          </Button>
          <Button variant="outline" size="sm" onClick={handleCollapseAll}>
            <Minimize className="mr-2 h-4 w-4" />
            全部收起
          </Button>
          
          {/* Create Button */}
          <Button onClick={() => setIsModalVisible(true)} disabled={isLoading}>
            <Plus className="mr-2 h-4 w-4" />
            新增组织
          </Button>
        </div>
      </div>

      {/* Search and Filters */}
      <div className="mb-6 flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <Input
            placeholder="搜索组织名称或描述..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <Select 
          value={filters.unit_type || 'all'} 
          onValueChange={(value) => setFilters({ ...filters, unit_type: value === 'all' ? undefined : value })}
        >
          <SelectTrigger className="w-48">
            <SelectValue placeholder="组织类型" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">所有类型</SelectItem>
            <SelectItem value="COMPANY">公司</SelectItem>
            <SelectItem value="DEPARTMENT">部门</SelectItem>
            <SelectItem value="PROJECT_TEAM">项目团队</SelectItem>
            <SelectItem value="COST_CENTER">成本中心</SelectItem>
          </SelectContent>
        </Select>
        <Select 
          value={filters.status || 'all'} 
          onValueChange={(value) => setFilters({ ...filters, status: value === 'all' ? undefined : value })}
        >
          <SelectTrigger className="w-32">
            <SelectValue placeholder="状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">所有状态</SelectItem>
            <SelectItem value="ACTIVE">活跃</SelectItem>
            <SelectItem value="INACTIVE">停用</SelectItem>
            <SelectItem value="PLANNED">计划中</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Stats Cards - Using CQRS data */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">组织总数</p>
                <p className="text-2xl font-bold">{currentStats.total || 0}</p>
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
                  {currentStats.totalEmployees || 0}
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
                  {(currentStats.maxLevel || 0) + 1}
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
                  {currentStats.active || 0}
                </p>
              </div>
              <Crown className="h-8 w-8 text-orange-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Organization Tree - CQRS Enhanced */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            组织架构树
            {hasErrors && (
              <Badge variant="destructive" className="ml-2">
                有错误
              </Badge>
            )}
            {isLoading && (
              <Badge variant="secondary" className="ml-2">
                加载中...
              </Badge>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent className="p-6">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="text-gray-500">正在加载组织数据...</div>
            </div>
          ) : hasErrors ? (
            <Alert>
              <AlertDescription className="text-red-600">
                数据加载失败：{Object.values(errors).filter(Boolean).join(', ')}
                <Button 
                  variant="outline" 
                  size="sm" 
                  className="ml-4"
                  onClick={refreshAll}
                >
                  重试
                </Button>
              </AlertDescription>
            </Alert>
          ) : currentOrgTree.length > 0 ? (
            <div className="space-y-2" data-testid="org-tree">
              {/* Display filtered organizations if searching, otherwise show tree */}
              {searchQuery ? (
                // Search Results
                <div>
                  <p className="text-sm text-gray-600 mb-4">
                    搜索 "{searchQuery}" 找到 {filteredOrganizations.length} 个结果
                  </p>
                  {filteredOrganizations.map((org: Organization) => (
                    <div key={org.id} className="mb-2">
                      {renderOrgNode(org, 0)}
                    </div>
                  ))}
                </div>
              ) : (
                // Organization Tree - 修复：传递正确的depth参数
                currentOrgTree.map((org: Organization) => renderOrgNode(org, 0))
              )}
            </div>
          ) : (
            <Alert>
              <AlertDescription>
                暂无组织架构数据，请先创建组织。
                <Button 
                  variant="outline" 
                  size="sm" 
                  className="ml-4"
                  onClick={() => setIsModalVisible(true)}
                >
                  创建组织
                </Button>
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
                  handleCreateOrganization(formData as CreateOrganizationRequest);
                }
              }} 
              disabled={isLoading || !formData.name || !formData.unit_type}
            >
              {isLoading ? '处理中...' : (editingOrganization ? '更新' : '创建')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default OrganizationChartPage;