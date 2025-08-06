import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { Card } from '@workday/canvas-kit-react/card'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button'
import { Table } from '@workday/canvas-kit-react/table'
import { useOrganizations, useOrganizationStats } from '../../shared/hooks/useOrganizations'
import type { OrganizationUnit } from '../../shared/types'

const OrganizationTable: React.FC<{ organizations: OrganizationUnit[] }> = ({ organizations }) => {
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
        {organizations.map((org) => (
          <Table.Row key={org.code}>
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
              <TertiaryButton size="small">
                查看详情
              </TertiaryButton>
            </Table.Cell>
          </Table.Row>
        ))}
      </Table.Body>
    </Table>
  );
};

const StatsCard: React.FC<{ title: string; stats: Record<string, number> }> = ({ title, stats }) => {
  return (
    <Card height="100%">
      <Card.Heading>
        {title}
      </Card.Heading>
      <Card.Body>
        <Box display="flex" flexDirection="column" justifyContent="center" height="100%">
          {Object.entries(stats).map(([key, value]) => (
            <Box key={key} paddingY="xs">
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
          <PrimaryButton marginRight="s">新增组织单元</PrimaryButton>
          <SecondaryButton marginRight="s">导入数据</SecondaryButton>
          <TertiaryButton>导出报告</TertiaryButton>
        </Box>
      </Box>

      {/* 统计信息卡片 */}
      {statsData && (
        <Box marginBottom="l" display="flex" alignItems="stretch">
          <Box flex={1} marginRight="xl">
            <StatsCard 
              title="按类型统计" 
              stats={statsData.by_type} 
            />
          </Box>
          <Box flex={1} marginRight="xl">
            <StatsCard 
              title="按状态统计" 
              stats={statsData.by_status} 
            />
          </Box>
          <Box flex={1}>
            <Card height="100%">
              <Card.Heading>
                总体概况
              </Card.Heading>
              <Card.Body>
                <Box textAlign="center" display="flex" flexDirection="column" justifyContent="center" height="100%">
                  <Text size="xxLarge" fontWeight="bold">{statsData.total_count}</Text>
                  <Text>组织单元总数</Text>
                </Box>
              </Card.Body>
            </Card>
          </Box>
        </Box>
      )}

      {/* 组织单元列表 */}
      <Card>
        <Card.Heading>
          组织单元列表
        </Card.Heading>
        <Card.Body>
          <Text marginBottom="m">共 {organizationData?.total_count || 0} 个单元</Text>
          {organizationData?.organizations && organizationData.organizations.length > 0 ? (
            <OrganizationTable organizations={organizationData.organizations} />
          ) : (
            <Box padding="xl" textAlign="center">
              <Text>暂无组织数据</Text>
            </Box>
          )}
        </Card.Body>
      </Card>
    </Box>
  );
};