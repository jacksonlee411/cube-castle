/**
 * 组织详情专用页面
 * 路由: /organizations/{code}/temporal 或 /organizations/new
 * 集成TemporalMasterDetailView组件实现完整的组织详情和创建体验
 */
import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { SecondaryButton } from '@workday/canvas-kit-react/button';
import { TemporalMasterDetailView } from '../temporal/components/TemporalMasterDetailView';

/**
 * 组织详情页面组件
 * 支持查看现有组织详情或创建新组织
 */
export const OrganizationTemporalPage: React.FC = () => {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();

  // 检测创建模式
  const isCreateMode = code === 'new';
  
  // 路由参数验证 - 对于创建模式，允许code为'new'，对于组织编码，只允许7位数字
  // 支持格式: 7位数字(1000000)
  if (!code || (code !== 'new' && !code.match(/^[0-9]{7}$/))) {
    return (
      <Box padding="xl" textAlign="center">
        <Heading size="medium" marginBottom="m">
          无效的组织编码
        </Heading>
        <Text typeLevel="body.medium" color="hint" marginBottom="l">
          请从组织列表页面正确访问组织详情功能
        </Text>
        <SecondaryButton onClick={() => navigate('/organizations')}>
          返回组织列表
        </SecondaryButton>
      </Box>
    );
  }

  // 返回组织列表页面
  const handleBackToList = () => {
    navigate('/organizations');
  };

  // 创建成功后的页面跳转处理
  const handleCreateSuccess = (newOrganizationCode: string) => {
    // 跳转到新创建的组织详情页面
    navigate(`/organizations/${newOrganizationCode}/temporal`, { replace: true });
  };

  return (
    <Box>
      {/* 面包屑导航 */}
      <Box padding="m" borderBottom="solid" borderColor="soap300" marginBottom="m">
        <Flex alignItems="center" gap="s">
          <SecondaryButton
            size="small"
            onClick={handleBackToList}
          >
            ← 组织列表
          </SecondaryButton>
          <Text typeLevel="subtext.medium" color="hint">
            /
          </Text>
          <Text typeLevel="subtext.medium" fontWeight="medium">
            {isCreateMode ? '新建组织' : code}
          </Text>
          <Text typeLevel="subtext.medium" color="hint">
            /
          </Text>
          <Text typeLevel="subtext.medium" fontWeight="medium">
            {isCreateMode ? '编辑组织信息' : '组织详情'}
          </Text>
        </Flex>
      </Box>

      {/* 主要内容区：组织详情主从视图 */}
      <TemporalMasterDetailView
        organizationCode={isCreateMode ? null : code}
        onBack={handleBackToList}
        onCreateSuccess={isCreateMode ? handleCreateSuccess : undefined}
        readonly={false}
        isCreateMode={isCreateMode}
      />
    </Box>
  );
};

export default OrganizationTemporalPage;