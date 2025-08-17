/**
 * 组织组织详情专用页面
 * 路由: /organizations/{code}/temporal
 * 集成TemporalMasterDetailView组件实现完整的组织详情体验
 */
import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { SecondaryButton } from '@workday/canvas-kit-react/button';
import { TemporalMasterDetailView } from '../temporal/components/TemporalMasterDetailView';

/**
 * 组织组织详情页面组件
 * 提供特定组织的组织详情功能集成中心
 */
export const OrganizationTemporalPage: React.FC = () => {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();

  // 路由参数验证
  if (!code) {
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
            {code}
          </Text>
          <Text typeLevel="subtext.medium" color="hint">
            /
          </Text>
          <Text typeLevel="subtext.medium" fontWeight="medium">
            组织详情
          </Text>
        </Flex>
      </Box>

      {/* 主要内容区：组织详情主从视图 */}
      <TemporalMasterDetailView
        organizationCode={code}
        onBack={handleBackToList}
        readonly={false}
      />
    </Box>
  );
};

export default OrganizationTemporalPage;