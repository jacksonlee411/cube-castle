import React, { useState } from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { Card } from '@workday/canvas-kit-react/card'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton, TertiaryButton, DeleteButton } from '@workday/canvas-kit-react/button'
import { Table } from '@workday/canvas-kit-react/table'
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal'
import { FormField } from '@workday/canvas-kit-react/form-field'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { Select } from '@workday/canvas-kit-react/select'
import { TextArea } from '@workday/canvas-kit-react/text-area'
import { useOrganizations, useOrganizationStats } from '../../shared/hooks/useOrganizations'
import { useCreateOrganization, useUpdateOrganization, useDeleteOrganization } from '../../shared/hooks/useOrganizationMutations'
import type { OrganizationUnit } from '../../shared/types'
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../../shared/hooks/useOrganizationMutations'

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

  const model = useModalModel({
    initialFocusRef: React.useRef<HTMLInputElement>(null),
  });

  React.useEffect(() => {
    if (isOpen && !model.state.visibility === 'visible') {
      model.events.open();
    } else if (!isOpen && model.state.visibility === 'visible') {
      model.events.close();
    }
  }, [isOpen, model]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      if (isEditing) {
        const updateData: UpdateOrganizationInput = {
          code: formData.code,
          name: formData.name,
          status: formData.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
          description: formData.description,
          sort_order: formData.sort_order,
        };
        
        await updateMutation.mutateAsync(updateData);
      } else {
        const createData: CreateOrganizationInput = {
          code: formData.code,
          name: formData.name,
          unit_type: formData.unit_type as 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM',
          status: formData.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
          level: formData.level,
          sort_order: formData.sort_order,
          description: formData.description,
          parent_code: formData.parent_code || undefined,
        };
        
        await createMutation.mutateAsync(createData);
      }
      
      onClose();
    } catch (error) {
      console.error('表单提交失败:', error);
    }
  };

  const handleClose = () => {
    model.events.close();
    onClose();
  };

  return (
    <Modal model={model}>
      <Modal.Overlay>
        <Modal.Card width={600}>
          <Modal.CloseIcon aria-label="关闭" />
          <Modal.Heading>{isEditing ? '编辑组织单元' : '新增组织单元'}</Modal.Heading>
          <Modal.Body>
            <form onSubmit={handleSubmit}>
              <FormField marginBottom="m">
                <FormField.Label>组织编码 *</FormField.Label>
                <FormField.Field>
                  <FormField.Input
                    as={TextInput}
                    value={formData.code}
                    onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                    disabled={isEditing}
                    placeholder="请输入组织编码"
                    required
                  />
                </FormField.Field>
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
                      <FormField.Input
                        as={Select}
                        value={formData.unit_type}
                        onChange={(value) => setFormData({ ...formData, unit_type: value })}
                      >
                        <option value="DEPARTMENT">部门</option>
                        <option value="COST_CENTER">成本中心</option>
                        <option value="COMPANY">公司</option>
                        <option value="PROJECT_TEAM">项目团队</option>
                      </FormField.Input>
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
                  <FormField.Input
                    as={Select}
                    value={formData.status}
                    onChange={(value) => setFormData({ ...formData, status: value })}
                  >
                    <option value="ACTIVE">激活</option>
                    <option value="INACTIVE">停用</option>
                    <option value="PLANNED">计划中</option>
                  </FormField.Input>
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

              <Box display="flex" justifyContent="flex-end" gap="s">
                <Modal.CloseButton as={SecondaryButton} onClick={handleClose}>
                  取消
                </Modal.CloseButton>
                <PrimaryButton 
                  type="submit" 
                  disabled={createMutation.isPending || updateMutation.isPending}
                >
                  {isEditing ? '更新' : '创建'}
                </PrimaryButton>
              </Box>
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
}> = ({ organizations, onEdit, onDelete }) => {
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
        {organizations.map((org, index) => (
          <Table.Row key={org.code || `org-${index}`}>
            <Table.Cell>{org.code}</Table.Cell>
            <Table.Cell>{org.name}</Table.Cell>
            <Table.Cell>{org.unit_type}</Table.Cell>
            <Table.Cell>
              <Text color={org.status === 'ACTIVE' ? 'positive' : 'default'}>
                {org.status}
              </Text>
            </Table.Cell>
            <Table.Cell>{org.level}</Table.Cell>
            <Table.Cell>
              <Box display="flex" gap="xs">
                <TertiaryButton size="small" onClick={() => onEdit(org)}>
                  编辑
                </TertiaryButton>
                <DeleteButton size="small" onClick={() => onDelete(org.code)}>
                  删除
                </DeleteButton>
              </Box>
            </Table.Cell>
          </Table.Row>
        ))}
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
        <Box display="flex" flexDirection="column" justifyContent="center" height="100%">
          {Object.entries(stats).map(([key, value], index) => (
            <Box key={`${title}-${key}-${index}`} paddingY="xs">
              <Text>{key}: {value}</Text>
            </Box>
          ))}
        </Box>
      </Card.Body>
    </Card>
  );
};

export const OrganizationDashboard: React.FC = () => {
  const { data: organizationData, isLoading: orgLoading, error: orgError } = useOrganizations();
  const { data: statsData } = useOrganizationStats();
  const deleteMutation = useDeleteOrganization();
  
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [selectedOrganization, setSelectedOrganization] = useState<OrganizationUnit | undefined>(undefined);
  
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
      await deleteMutation.mutateAsync(code);
    }
  };
  
  const handleFormClose = () => {
    setIsFormOpen(false);
    setSelectedOrganization(undefined);
  };

  if (orgLoading) {
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
          <PrimaryButton marginRight="s" onClick={handleCreate}>新增组织单元</PrimaryButton>
          <SecondaryButton marginRight="s">导入数据</SecondaryButton>
          <TertiaryButton>导出报告</TertiaryButton>
        </Box>
      </Box>

      {/* 统计信息卡片 - 恢复Canvas Kit Card组件 */}
      {statsData && (
        <Box as="div" marginBottom="l" display="flex" alignItems="stretch" gap="l">
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
                <Box as="div" textAlign="center" display="flex" flexDirection="column" justifyContent="center" height="100%">
                  <Text as="div" fontWeight="bold" style={{ fontSize: '2rem' }}>{statsData.total_count}</Text>
                  <Text>组织单元总数</Text>
                </Box>
              </Card.Body>
            </Card>
          </Box>
        </Box>
      )}

      {/* 组织单元列表 - 恢复Canvas Kit Card组件 */}
      <Card>
        <Card.Heading>组织单元列表</Card.Heading>
        <Card.Body>
          <Text marginBottom="m">共 {organizationData?.total_count || 0} 个单元</Text>
          {organizationData?.organizations && organizationData.organizations.length > 0 ? (
            <OrganizationTable 
              organizations={organizationData.organizations} 
              onEdit={handleEdit}
              onDelete={handleDelete}
            />
          ) : (
            <Box padding="xl" textAlign="center">
              <Text>暂无组织数据</Text>
            </Box>
          )}
        </Card.Body>
      </Card>

      {/* 新增/编辑模态窗口 - 恢复完整的Canvas Kit Modal + FormField */}
      <OrganizationForm 
        organization={selectedOrganization}
        isOpen={isFormOpen}
        onClose={handleFormClose}
      />
    </Box>
  );
};