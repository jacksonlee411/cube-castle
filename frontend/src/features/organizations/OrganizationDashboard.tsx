import React, { useState, useEffect } from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { Card } from '@workday/canvas-kit-react/card'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton, TertiaryButton, DeleteButton } from '@workday/canvas-kit-react/button'
import { Table } from '@workday/canvas-kit-react/table'
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal'
import { FormField } from '@workday/canvas-kit-react/form-field'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { TextArea } from '@workday/canvas-kit-react/text-area'
import { useOrganizations, useOrganizationStats } from '../../shared/hooks/useOrganizations'
import { useCreateOrganization, useUpdateOrganization, useDeleteOrganization } from '../../shared/hooks/useOrganizationMutations'
import type { OrganizationUnit } from '../../shared/types'
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../../shared/hooks/useOrganizationMutations'
import type { OrganizationQueryParams } from '../../shared/api/organizations'
import { OrganizationFilters, type FilterState } from './OrganizationFilters'
import { PaginationControls } from './PaginationControls'

// 组织单元表单组件 - 使用Canvas Kit v13最佳实践
interface OrganizationFormProps {
  organization?: OrganizationUnit;
  onClose: () => void;
  isOpen: boolean;
}

const OrganizationForm: React.FC<OrganizationFormProps> = ({ organization, onClose, isOpen }) => {
  const createMutation = useCreateOrganization();
  const updateMutation = useUpdateOrganization();
  const isEditing = !!organization;
  
  // 添加提交锁定状态
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  
  // Canvas Kit v13 Modal - 使用正确的API模式
  const model = useModalModel();

  const [formData, setFormData] = useState({
    code: organization?.code || '',
    name: organization?.name || '',
    unit_type: organization?.unit_type || 'DEPARTMENT',
    status: organization?.status || 'ACTIVE',
    description: organization?.description || '',
    parent_code: organization?.parent_code || '',
    level: organization?.level || 1,
    sort_order: organization?.sort_order || 0,
  });

  // 正确的Modal状态管理 - 使用事件API
  React.useEffect(() => {
    if (isOpen && model.state.visibility !== 'visible') {
      model.events.show();
    } else if (!isOpen && model.state.visibility === 'visible') {
      model.events.hide();
    }
  }, [isOpen, model]);

  // 重置表单数据当organization改变时
  useEffect(() => {
    setFormData({
      code: organization?.code || '',
      name: organization?.name || '',
      unit_type: organization?.unit_type || 'DEPARTMENT',
      status: organization?.status || 'ACTIVE',
      description: organization?.description || '',
      parent_code: organization?.parent_code || '',
      level: organization?.level || 1,
      sort_order: organization?.sort_order || 0,
    });
  }, [organization]);

  const handleSubmit = React.useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    console.log('[Form] handleSubmit调用 - 时间戳:', Date.now());
    
    // 强制防重复提交检查
    if (isSubmitting || createMutation.isPending || updateMutation.isPending) {
      console.log('[Form] 🚫 阻止重复提交 - 当前状态:', { 
        isSubmitting, 
        createPending: createMutation.isPending, 
        updatePending: updateMutation.isPending 
      });
      return;
    }
    
    // 设置提交锁定
    setIsSubmitting(true);
    console.log('[Form] 🔒 设置提交锁定 - 时间戳:', Date.now());
    
    try {
      if (isEditing) {
        const updateData: UpdateOrganizationInput = {
          code: formData.code,
          name: formData.name,
          status: formData.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
          description: formData.description,
          sort_order: formData.sort_order,
        };
        
        console.log('[Form] Submitting update:', updateData);
        await updateMutation.mutateAsync(updateData);
        console.log('[Form] Update successful');
      } else {
        const createData: CreateOrganizationInput = {
          code: formData.code.trim() || undefined, // 空字符串转为undefined，让后端自动生成
          name: formData.name,
          unit_type: formData.unit_type as 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM',
          status: formData.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
          level: formData.level,
          sort_order: formData.sort_order,
          description: formData.description,
          parent_code: formData.parent_code || undefined,
        };
        
        console.log('[Form] Submitting create:', createData);
        await createMutation.mutateAsync(createData);
        console.log('[Form] Create successful');
      }
      
      // 添加成功提示
      console.log(`[Form] ${isEditing ? '更新' : '创建'}成功！`);
      
      // 重置表单数据
      if (!isEditing) {
        setFormData({
          code: '',
          name: '',
          unit_type: 'DEPARTMENT',
          status: 'ACTIVE',
          description: '',
          parent_code: '',
          level: 1,
          sort_order: 0,
        });
      }
      
      // 使用Modal事件API关闭
      model.events.hide();
      onClose();
    } catch (error) {
      console.error(`[Form] ${isEditing ? '更新' : '创建'}失败:`, error);
      
      // 改进的错误信息处理
      let errorMessage = '操作失败';
      
      if (error && typeof error === 'object' && 'message' in error) {
        const apiError = error as any;
        
        // 检查是否包含具体的数据库错误信息
        if (apiError.message.includes('duplicate key value violates unique constraint')) {
          if (apiError.message.includes('uk_tenant_name')) {
            errorMessage = '组织名称已存在，请使用不同的名称';
          } else {
            errorMessage = '数据重复，请检查输入信息';
          }
        } else if (apiError.message.includes('Network connection failed')) {
          errorMessage = '网络连接失败，请检查服务器状态';
        } else {
          // 使用原始错误信息，但去掉技术细节
          errorMessage = apiError.message.split(' - ')[0] || errorMessage;
        }
      } else if (error instanceof Error) {
        errorMessage = error.message;
      }
      
      alert(errorMessage);
    } finally {
      // 无论成功失败都释放锁定
      setIsSubmitting(false);
      console.log('[Form] 🔓 释放提交锁定 - 时间戳:', Date.now());
    }
  }, [isEditing, formData, createMutation, updateMutation, isSubmitting, model, onClose]);

  // 处理Modal关闭 - 使用正确的事件API
  const handleClose = () => {
    // 重置提交状态
    setIsSubmitting(false);
    console.log('[Form] Modal关闭，重置提交状态');
    
    model.events.hide();
    onClose();
  };

  return (
    <Modal model={model}>
      <Modal.Overlay>
        <Modal.Card width={600}>
          <Modal.CloseIcon aria-label="关闭" onClick={handleClose} />
          <Modal.Heading>{isEditing ? '编辑组织单元' : '新增组织单元'}</Modal.Heading>
          <Modal.Body>
            <form onSubmit={handleSubmit}>
              <FormField marginBottom="m">
                <FormField.Label>组织编码</FormField.Label>
                <FormField.Field>
                  <FormField.Input
                    as={TextInput}
                    value={formData.code}
                    onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                    disabled={true}
                    placeholder="系统自动生成编码"
                    style={{ backgroundColor: '#f5f5f5', cursor: 'not-allowed' }}
                  />
                </FormField.Field>
                <FormField.Hint>
                  {isEditing ? "编码不可修改" : "系统将自动生成唯一编码"}
                </FormField.Hint>
              </FormField>

              <FormField marginBottom="m">
                <FormField.Label>组织名称 *</FormField.Label>
                <FormField.Field>
                  <FormField.Input
                    as={TextInput}
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="请输入组织名称"
                    required
                  />
                </FormField.Field>
              </FormField>

              {!isEditing && (
                <>
                  <FormField marginBottom="m">
                    <FormField.Label>组织类型 *</FormField.Label>
                    <FormField.Field>
                      <select
                        value={formData.unit_type}
                        onChange={(e) => setFormData({ ...formData, unit_type: e.target.value as 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM' })}
                        style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ddd' }}
                      >
                        <option value="DEPARTMENT">部门</option>
                        <option value="COST_CENTER">成本中心</option>
                        <option value="COMPANY">公司</option>
                        <option value="PROJECT_TEAM">项目团队</option>
                      </select>
                    </FormField.Field>
                  </FormField>

                  <FormField marginBottom="m">
                    <FormField.Label>上级组织编码</FormField.Label>
                    <FormField.Field>
                      <FormField.Input
                        as={TextInput}
                        value={formData.parent_code}
                        onChange={(e) => setFormData({ ...formData, parent_code: e.target.value })}
                        placeholder="请输入上级组织编码"
                      />
                    </FormField.Field>
                  </FormField>

                  <FormField marginBottom="m">
                    <FormField.Label>组织层级 *</FormField.Label>
                    <FormField.Field>
                      <FormField.Input
                        as={TextInput}
                        type="number"
                        value={formData.level}
                        onChange={(e) => setFormData({ ...formData, level: parseInt(e.target.value) || 1 })}
                        min="1"
                        required
                      />
                    </FormField.Field>
                  </FormField>
                </>
              )}

              <FormField marginBottom="m">
                <FormField.Label>状态 *</FormField.Label>
                <FormField.Field>
                  <select
                    value={formData.status}
                    onChange={(e) => setFormData({ ...formData, status: e.target.value as 'ACTIVE' | 'INACTIVE' | 'PLANNED' })}
                    style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ddd' }}
                  >
                    <option value="ACTIVE">激活</option>
                    <option value="INACTIVE">停用</option>
                    <option value="PLANNED">计划中</option>
                  </select>
                </FormField.Field>
              </FormField>

              <FormField marginBottom="m">
                <FormField.Label>排序</FormField.Label>
                <FormField.Field>
                  <FormField.Input
                    as={TextInput}
                    type="number"
                    value={formData.sort_order}
                    onChange={(e) => setFormData({ ...formData, sort_order: parseInt(e.target.value) || 0 })}
                    min="0"
                  />
                </FormField.Field>
              </FormField>

              <FormField marginBottom="l">
                <FormField.Label>描述</FormField.Label>
                <FormField.Field>
                  <FormField.Input
                    as={TextArea}
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    placeholder="请输入组织描述"
                    rows={3}
                  />
                </FormField.Field>
              </FormField>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
                <SecondaryButton type="button" onClick={handleClose}>
                  取消
                </SecondaryButton>
                <PrimaryButton 
                  type="submit" 
                  disabled={isSubmitting || createMutation.isPending || updateMutation.isPending}
                >
                  {isEditing ? '更新' : '创建'}
                </PrimaryButton>
              </div>
            </form>
          </Modal.Body>
        </Modal.Card>
      </Modal.Overlay>
    </Modal>
  );
};

// 表格组件
const OrganizationTable: React.FC<{ 
  organizations: OrganizationUnit[]; 
  onEdit: (org: OrganizationUnit) => void;
  onDelete: (code: string) => void;
  deleteMutation: any; // 传入删除mutation以获取loading状态
}> = ({ organizations, onEdit, onDelete, deleteMutation }) => {
  return (
    <Table>
      <Table.Head>
        <Table.Row>
          <Table.Header>编码</Table.Header>
          <Table.Header>名称</Table.Header>
          <Table.Header>类型</Table.Header>
          <Table.Header>状态</Table.Header>
          <Table.Header>层级</Table.Header>
          <Table.Header>操作</Table.Header>
        </Table.Row>
      </Table.Head>
      <Table.Body>
        {organizations.map((org, index) => {
          const isDeleting = deleteMutation.isPending && deleteMutation.variables === org.code;
          return (
            <Table.Row 
              key={org.code || `org-${index}`}
              style={{ 
                opacity: isDeleting ? 0.6 : 1,
                transition: 'opacity 0.3s ease'
              }}
            >
              <Table.Cell>{org.code}</Table.Cell>
              <Table.Cell>
                {org.name}
                {isDeleting && (
                  <Text typeLevel="subtext.small" color="hint" marginLeft="xs">
                    (删除中...)
                  </Text>
                )}
              </Table.Cell>
              <Table.Cell>{org.unit_type}</Table.Cell>
              <Table.Cell>
                <Text color={
                  org.status === 'ACTIVE' ? 'positive' : 
                  org.status === 'PLANNED' ? 'hint' : 
                  'default'
                }>
                  {org.status}
                </Text>
              </Table.Cell>
              <Table.Cell>{org.level}</Table.Cell>
              <Table.Cell>
                <div style={{ display: 'flex', gap: '4px' }}>
                  <TertiaryButton 
                    size="small" 
                    onClick={() => onEdit(org)}
                    disabled={deleteMutation.isPending} // 删除进行时禁用编辑
                  >
                    编辑
                  </TertiaryButton>
                  <DeleteButton 
                    size="small" 
                    onClick={() => onDelete(org.code)}
                    disabled={deleteMutation.isPending} // 防止重复点击
                  >
                    {isDeleting ? '删除中...' : '删除'}
                  </DeleteButton>
                </div>
              </Table.Cell>
            </Table.Row>
          );
        })}
      </Table.Body>
    </Table>
  );
};

// 统计卡片组件 - 使用Canvas Kit Card
const StatsCard: React.FC<{ title: string; stats: Record<string, number> }> = ({ title, stats }) => {
  return (
    <Card height="100%">
      <Card.Heading>{title}</Card.Heading>
      <Card.Body>
        <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', height: '100%' }}>
          {Object.entries(stats).map(([key, value], index) => (
            <Box key={`${title}-${key}-${index}`} paddingY="xs">
              <Text>{key}: {value}</Text>
            </Box>
          ))}
        </div>
      </Card.Body>
    </Card>
  );
};

export const OrganizationDashboard: React.FC = () => {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [selectedOrganization, setSelectedOrganization] = useState<OrganizationUnit | undefined>(undefined);
  
  // 筛选状态管理
  const [filters, setFilters] = useState<FilterState>({
    searchText: '',
    unit_type: undefined,
    status: undefined,
    level: undefined,
    page: 1,
    pageSize: 20,
  });

  // 将筛选状态转换为API查询参数
  const queryParams: OrganizationQueryParams = {
    searchText: filters.searchText || undefined,
    unit_type: filters.unit_type || undefined,
    status: filters.status || undefined,
    level: filters.level || undefined,
    page: filters.page,
    pageSize: filters.pageSize,
  };

  const { data: organizationData, isLoading: orgLoading, error: orgError, isFetching } = useOrganizations(queryParams);
  const { data: statsData } = useOrganizationStats();
  const deleteMutation = useDeleteOrganization();
  
  const handleCreate = () => {
    setSelectedOrganization(undefined);
    setIsFormOpen(true);
  };
  
  const handleEdit = (org: OrganizationUnit) => {
    setSelectedOrganization(org);
    setIsFormOpen(true);
  };
  
  const handleDelete = async (code: string) => {
    if (window.confirm('确定要删除这个组织单元吗？')) {
      try {
        await deleteMutation.mutateAsync(code);
        // 删除成功，React Query会自动invalidateQueries刷新数据
      } catch (error) {
        // 错误处理已在mutation中完成，这里可以添加用户友好的错误提示
        console.error('Delete operation failed:', error);
        // 可以添加toast通知等
      }
    }
  };
  
  const handleFormClose = () => {
    setIsFormOpen(false);
    setSelectedOrganization(undefined);
  };

  const handleFiltersChange = (newFilters: FilterState) => {
    setFilters(newFilters);
  };

  const handlePageChange = (page: number) => {
    setFilters(prev => ({ ...prev, page }));
  };

  if (orgLoading && !isFetching) {
    return (
      <Box padding="l">
        <Text>加载组织数据中...</Text>
      </Box>
    );
  }

  if (orgError) {
    return (
      <Box padding="l">
        <Text>加载失败: {orgError.message}</Text>
      </Box>
    );
  }

  return (
    <Box>
      {/* 页面标题和操作栏 */}
      <Box marginBottom="l">
        <Heading size="large">组织架构管理</Heading>
        <Box paddingTop="m">
          <PrimaryButton 
            marginRight="s" 
            onClick={handleCreate}
            disabled={deleteMutation.isPending} // 删除进行时禁用新建
          >
            新增组织单元
          </PrimaryButton>
          <SecondaryButton 
            marginRight="s"
            disabled={deleteMutation.isPending} // 删除进行时禁用导入
          >
            导入数据
          </SecondaryButton>
          <TertiaryButton disabled={deleteMutation.isPending}>导出报告</TertiaryButton>
          {deleteMutation.isPending && (
            <Text typeLevel="subtext.small" color="hint" marginLeft="m">
              正在删除组织单元...
            </Text>
          )}
        </Box>
      </Box>

      {/* 统计信息卡片 */}
      {statsData && (
        <div style={{ marginBottom: '16px', display: 'flex', alignItems: 'stretch', gap: '16px' }}>
          <Box flex={1}>
            <StatsCard 
              title="按类型统计" 
              stats={statsData.by_type} 
            />
          </Box>
          <Box flex={1}>
            <StatsCard 
              title="按状态统计" 
              stats={statsData.by_status} 
            />
          </Box>
          <Box flex={1}>
            <Card height="100%">
              <Card.Heading>总体概况</Card.Heading>
              <Card.Body>
                <div style={{ textAlign: 'center', display: 'flex', flexDirection: 'column', justifyContent: 'center', height: '100%' }}>
                  <Text fontWeight="bold" style={{ fontSize: '2rem' }}>{statsData.total_count}</Text>
                  <Text>组织单元总数</Text>
                </div>
              </Card.Body>
            </Card>
          </Box>
        </div>
      )}

      {/* 筛选面板 */}
      <OrganizationFilters 
        filters={filters}
        onFiltersChange={handleFiltersChange}
      />

      {/* 组织单元列表 */}
      <Card>
        <Card.Heading>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>组织单元列表</span>
            {isFetching && (
              <Text typeLevel="subtext.small" color="hint">
                加载中...
              </Text>
            )}
          </div>
        </Card.Heading>
        <Card.Body>
          {organizationData && organizationData.organizations && organizationData.organizations.length > 0 ? (
            <>
              <OrganizationTable 
                organizations={organizationData.organizations} 
                onEdit={handleEdit}
                onDelete={handleDelete}
                deleteMutation={deleteMutation}
              />
              
              {/* 分页控件 */}
              <PaginationControls
                currentPage={filters.page}
                totalCount={organizationData?.total_count || 0}
                pageSize={filters.pageSize}
                onPageChange={handlePageChange}
                disabled={isFetching || deleteMutation.isPending}
              />
            </>
          ) : (
            <Box padding="xl" textAlign="center">
              <Text>
                {filters.searchText || filters.unit_type || filters.status || filters.level
                  ? '没有找到符合筛选条件的组织单元'
                  : '暂无组织数据'
                }
              </Text>
              {(filters.searchText || filters.unit_type || filters.status || filters.level) && (
                <Box marginTop="s">
                  <SecondaryButton 
                    size="small"
                    onClick={() => setFilters({
                      searchText: '',
                      unit_type: undefined,
                      status: undefined,
                      level: undefined,
                      page: 1,
                      pageSize: 20,
                    })}
                  >
                    清除筛选条件
                  </SecondaryButton>
                </Box>
              )}
            </Box>
          )}
        </Card.Body>
      </Card>

      {/* 新增/编辑模态窗口 */}
      <OrganizationForm 
        organization={selectedOrganization}
        isOpen={isFormOpen}
        onClose={handleFormClose}
      />
    </Box>
  );
};