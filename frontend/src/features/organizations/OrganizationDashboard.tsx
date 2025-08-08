import React, { useState } from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton, TertiaryButton, DeleteButton } from '@workday/canvas-kit-react/button'
import { Table } from '@workday/canvas-kit-react/table'
import { useOrganizations, useOrganizationStats } from '../../shared/hooks/useOrganizations'
import { useCreateOrganization, useUpdateOrganization, useDeleteOrganization } from '../../shared/hooks/useOrganizationMutations'
import type { OrganizationUnit } from '../../shared/types'
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../../shared/hooks/useOrganizationMutations'

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
              <Box display="flex">
                <TertiaryButton size="small" marginRight="xs" onClick={() => onEdit(org)}>
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

export const OrganizationDashboard: React.FC = () => {
  const { data: organizationData, isLoading: orgLoading, error: orgError } = useOrganizations();
  const { data: statsData } = useOrganizationStats();
  const deleteMutation = useDeleteOrganization();
  
  const [selectedOrganization, setSelectedOrganization] = useState<OrganizationUnit | undefined>(undefined);
  
  const handleCreate = () => {
    console.log('创建新组织');
  };
  
  const handleEdit = (org: OrganizationUnit) => {
    setSelectedOrganization(org);
    console.log('编辑组织:', org);
  };
  
  const handleDelete = async (code: string) => {
    if (window.confirm('确定要删除这个组织单元吗？')) {
      await deleteMutation.mutateAsync(code);
    }
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

      {/* 统计信息显示 */}
      {statsData && (
        <Box marginBottom="l">
          <Box padding="m" border="solid" borderColor="subtle" borderRadius="m">
            <Text size="large" fontWeight="bold">统计概览</Text>
            <Box marginTop="s">
              <Text>组织单元总数: {statsData.total_count}</Text>
            </Box>
            {Object.keys(statsData.by_type).length > 0 && (
              <Box marginTop="s">
                <Text fontWeight="bold">按类型统计:</Text>
                {Object.entries(statsData.by_type).map(([key, value], index) => (
                  <Text key={`type-${key}-${index}`} marginLeft="s">{key}: {value}</Text>
                ))}
              </Box>
            )}
            {Object.keys(statsData.by_status).length > 0 && (
              <Box marginTop="s">
                <Text fontWeight="bold">按状态统计:</Text>
                {Object.entries(statsData.by_status).map(([key, value], index) => (
                  <Text key={`status-${key}-${index}`} marginLeft="s">{key}: {value}</Text>
                ))}
              </Box>
            )}
          </Box>
        </Box>
      )}

      {/* 组织单元列表 */}
      <Box border="solid" borderColor="subtle" borderRadius="m" padding="m">
        <Heading size="medium" marginBottom="m">
          组织单元列表
        </Heading>
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
      </Box>
    </Box>
  );
};