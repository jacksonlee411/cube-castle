/**
 * 组织详情页面 - 集成时间线功能
 * 展示组织的详细信息、历史版本和时间线事件
 */
import React, { useState, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';
import { Card } from '@workday/canvas-kit-react/card';
import { Badge } from '../../../shared/components/Badge';
import { Tabs, useTabsModel } from '@workday/canvas-kit-react/tabs';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';

// 组织管理和时态功能导入
import { OrganizationForm } from './OrganizationForm';
import { TemporalNavbar } from '../../temporal/components/TemporalNavbar';

// Hooks导入
import { useTemporalOrganization, useOrganizationHistory, useOrganizationTimeline, useTemporalMode } from '../../../shared/hooks/useTemporalQuery';
import { useOrganizationActions } from '../hooks/useOrganizationActions';

// Types导入
import type { OrganizationUnit } from '../../../shared/types/organization';
import type { TemporalMode } from '../../../shared/types/temporal';

export interface OrganizationDetailProps {
  /** 组织编码 */
  organizationCode: string;
  /** 是否只读模式 */
  readonly?: boolean;
  /** 返回回调 */
  onBack?: () => void;
}

/**
 * 组织基本信息卡片
 */
interface OrganizationInfoCardProps {
  organization: OrganizationUnit;
  isHistorical: boolean;
  onEdit?: () => void;
  onToggleStatus?: () => void;
  isLoading?: boolean;
}

const OrganizationInfoCard: React.FC<OrganizationInfoCardProps> = ({
  organization,
  isHistorical,
  onEdit,
  onToggleStatus,
  isLoading = false
}) => {
  const getStatusBadge = (status: string) => {
    const statusConfig = {
      'ACTIVE': { label: '启用', color: 'greenFresca600' },
      'INACTIVE': { label: '停用', color: 'cinnamon600' },
      'PLANNED': { label: '计划中', color: 'blueberry600' }
    };
    
    const config = statusConfig[status as keyof typeof statusConfig] || { label: status, color: 'licorice400' };
    return <Badge color={config.color as 'greenFresca600' | 'cinnamon600' | 'blueberry600' | 'licorice400'}>{config.label}</Badge>;
  };

  const getUnitTypeName = (unitType: string) => {
    const typeNames = {
      'ORGANIZATION_UNIT': '组织单位',
      'DEPARTMENT': '部门',
      'PROJECT_TEAM': '项目团队'
    };
    return typeNames[unitType as keyof typeof typeNames] || unitType;
  };

  const getUnitTypeBadge = (unitType: string) => {
    const typeConfig = {
      'ORGANIZATION_UNIT': { label: '组织单位', color: 'greenFresca600' },    // 组织单位 - 绿色（重要）
      'DEPARTMENT': { label: '部门', color: 'blueberry600' },              // 部门 - 蓝色（常见）
      'PROJECT_TEAM': { label: '项目团队', color: 'cantaloupe600' }         // 项目团队 - 橙色（临时性）
    };
    
    const config = typeConfig[unitType as keyof typeof typeConfig] || { label: unitType, color: 'licorice400' };
    return <Badge color={config.color as 'greenFresca600' | 'blueberry600' | 'cantaloupe600' | 'licorice400'}>{config.label}</Badge>;
  };

  return (
    <Card padding="m">
      <Flex justifyContent="space-between" alignItems="flex-start" marginBottom="m">
        <Box flex="1">
          <Flex alignItems="center" gap="s" marginBottom="s">
            <Heading size="medium">{organization.name}</Heading>
            {getStatusBadge(organization.status)}
            {getUnitTypeBadge(organization.unit_type)}
            {isHistorical && (
              <Badge color="blueberry600">历史视图</Badge>
            )}
          </Flex>
          
          <Text typeLevel="subtext.medium" color="hint" marginBottom="s">
            编码: {organization.code} • 类型: {getUnitTypeName(organization.unit_type)} • 层级: {organization.level}
            {organization.record_id && (
              <>
                <br />
                UUID: {organization.record_id}
              </>
            )}
          </Text>
          
          {organization.description && (
            <Text typeLevel="body.medium" marginBottom="s">
              {organization.description}
            </Text>
          )}
          
          <Flex gap="m" marginBottom="s">
            {organization.parent_code && (
              <Text typeLevel="subtext.small">
                上级组织: {organization.parent_code}
              </Text>
            )}
            <Text typeLevel="subtext.small">
              排序: {organization.sort_order}
            </Text>
          </Flex>
        </Box>

        <Box>
          <Flex gap="s">
            {!isHistorical && onEdit && (
              <PrimaryButton 
                size="small" 
                onClick={onEdit}
                disabled={isLoading}
              >
                编辑
              </PrimaryButton>
            )}
            {!isHistorical && onToggleStatus && (
              <SecondaryButton 
                size="small" 
                onClick={onToggleStatus}
                disabled={isLoading}
              >
                {organization.status === 'ACTIVE' ? '停用' : '启用'}
              </SecondaryButton>
            )}
          </Flex>
        </Box>
      </Flex>
      
      <Flex gap="m" justifyContent="space-between" alignItems="center">
        <Text typeLevel="subtext.small" color="hint">
          创建时间: {organization.created_at ? new Date(organization.created_at).toLocaleString('zh-CN') : '未知'}
        </Text>
        {organization.updated_at && (
          <Text typeLevel="subtext.small" color="hint">
            更新时间: {new Date(organization.updated_at).toLocaleString('zh-CN')}
          </Text>
        )}
      </Flex>
    </Card>
  );
};

/**
 * 组织详情页面主组件
 */
export const OrganizationDetail: React.FC<OrganizationDetailProps> = ({
  organizationCode,
  readonly = false,
  onBack
}) => {
  // 状态管理
  const [activeTab] = useState('overview');
  
  // Tabs模型 (Canvas Kit v13)
  const tabsModel = useTabsModel({
    initialTab: activeTab
  });

  // 时态模式管理
  const { mode: temporalMode, isHistorical } = useTemporalMode();

  // 组织数据查询
  const {
    data: organization,
    isLoading: orgLoading,
    isError: orgError,
    error: orgErrorMessage,
    refetch: refetchOrganization
  } = useTemporalOrganization(organizationCode);

  // 历史版本查询
  const {
    data: historyVersions = [],
    hasHistory
  } = useOrganizationHistory(organizationCode, { limit: 20 });

  // 时间线事件查询
  const {
    isLoading: timelineLoading,
    hasEvents: hasTimelineEvents,
    eventCount,
    latestEvent,
    refetch: refetchTimeline
  } = useOrganizationTimeline(organizationCode, { limit: 50 });

  // 组织操作钩子
  const {
    selectedOrg,
    isFormOpen,
    isToggling,
    handleEdit,
    handleToggleStatus,
    handleFormClose,
    handleFormSubmit,
  } = useOrganizationActions();

  // 时态模式变更处理
  const handleTemporalModeChange = useCallback((newMode: TemporalMode) => {
    console.log(`时态模式切换到: ${newMode}，重新加载组织数据`);
    refetchOrganization();
  }, [refetchOrganization]);

  // 编辑组织处理
  const handleEditOrganization = useCallback(() => {
    if (organization) {
      handleEdit(organization);
    }
  }, [organization, handleEdit]);

  // 切换状态处理
  const handleToggleOrganizationStatus = useCallback(() => {
    if (organization) {
      handleToggleStatus(organization.code, organization.status === 'ACTIVE' ? 'INACTIVE' : 'ACTIVE');
    }
  }, [organization, handleToggleStatus]);

  // 刷新所有数据
  const handleRefreshAll = useCallback(() => {
    refetchOrganization();
    refetchTimeline();
  }, [refetchOrganization, refetchTimeline]);

  // 加载状态
  if (orgLoading && !organization) {
    return (
      <Box padding="l">
        <Flex justifyContent="center" alignItems="center" height="200px">
          <LoadingDots />
          <Text marginLeft="m">加载组织详情中...</Text>
        </Flex>
      </Box>
    );
  }

  // 错误状态
  if (orgError || !organization) {
    return (
      <Box padding="l">
        <Card padding="l">
          <Text color="cinnamon600" typeLevel="heading.medium" marginBottom="m">
            ❌ 加载组织详情失败
          </Text>
          <Text marginBottom="m">
            {orgErrorMessage?.message || '无法加载组织信息，请检查组织编码或网络连接'}
          </Text>
          <Box>
            <PrimaryButton onClick={() => refetchOrganization()} marginRight="s">
              重试
            </PrimaryButton>
            {onBack && (
              <SecondaryButton onClick={onBack}>
                返回
              </SecondaryButton>
            )}
          </Box>
        </Card>
      </Box>
    );
  }

  return (
    <Box padding="l" data-testid="organization-detail">
      {/* 时态导航栏 */}
      <Box marginBottom="l">
        <TemporalNavbar
          onModeChange={handleTemporalModeChange}
          showAdvancedSettings={true}
        />
      </Box>

      {/* 页面头部 */}
      <Box marginBottom="l">
        <Flex justifyContent="space-between" alignItems="flex-start">
          <Box>
            <Heading size="large" marginBottom="s">
              组织详情
              {isHistorical && (
                <Text as="span" typeLevel="subtext.medium" color="hint" marginLeft="s">
                  (历史视图)
                </Text>
              )}
            </Heading>
          </Box>
          
          <Flex gap="s">
            <SecondaryButton 
              onClick={handleRefreshAll}
              disabled={orgLoading || timelineLoading}
            >
              刷新 刷新
            </SecondaryButton>
            {onBack && (
              <TertiaryButton onClick={onBack}>
                ← 返回
              </TertiaryButton>
            )}
          </Flex>
        </Flex>
      </Box>

      {/* 组织基本信息 */}
      <Box marginBottom="l">
        <OrganizationInfoCard
          organization={organization}
          isHistorical={isHistorical}
          onEdit={readonly ? undefined : handleEditOrganization}
          onToggleStatus={readonly ? undefined : handleToggleOrganizationStatus}
          isLoading={isToggling}
        />
      </Box>

      {/* 详情标签页 */}
      <Tabs model={tabsModel}>
        <Tabs.List>
          <Tabs.Item name="overview">
            概览信息
          </Tabs.Item>
          <Tabs.Item name="timeline">
            时间线 {hasTimelineEvents && <Badge color="blueberry600">{eventCount}</Badge>}
          </Tabs.Item>
          <Tabs.Item name="history">
            历史版本 {hasHistory && <Badge color="greenFresca600">{historyVersions.length}</Badge>}
          </Tabs.Item>
          <Tabs.Item name="comparison">
            版本对比
          </Tabs.Item>
        </Tabs.List>

        <Tabs.Panel>
          <Box marginTop="l">
            <Card padding="m">
              <Text as="h3" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
                详情 组织概览信息
              </Text>
              
              <Flex flexDirection="column" gap="m">
                <Box>
                  <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">基本信息</Text>
                  <Text typeLevel="body.small">编码: {organization.code}</Text>
                  {organization.record_id && (
                    <Text typeLevel="body.small">UUID: {organization.record_id}</Text>
                  )}
                  <Text typeLevel="body.small">名称: {organization.name}</Text>
                  <Text typeLevel="body.small">状态: {organization.status}</Text>
                  <Text typeLevel="body.small">类型: {organization.unit_type}</Text>
                </Box>
                
                <Box>
                  <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">层级结构</Text>
                  <Text typeLevel="body.small">层级: {organization.level}</Text>
                  <Text typeLevel="body.small">上级: {organization.parent_code || '无'}</Text>
                  <Text typeLevel="body.small">排序: {organization.sort_order}</Text>
                </Box>
                
                <Box>
                  <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">时间信息</Text>
                  <Text typeLevel="body.small">创建: {organization.created_at ? new Date(organization.created_at).toLocaleDateString('zh-CN') : '未知'}</Text>
                  <Text typeLevel="body.small">更新: {organization.updated_at ? new Date(organization.updated_at).toLocaleDateString('zh-CN') : '未知'}</Text>
                </Box>
                
                {hasTimelineEvents && (
                  <Box>
                    <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">活动统计</Text>
                    <Text typeLevel="body.small">时间线事件: {eventCount} 个</Text>
                    <Text typeLevel="body.small">历史版本: {historyVersions.length} 个</Text>
                    {latestEvent && (
                      <Text typeLevel="body.small">最新事件: {latestEvent.title}</Text>
                    )}
                  </Box>
                )}
              </Flex>
            </Card>
          </Box>
        </Tabs.Panel>

        <Tabs.Panel>
          <Box marginTop="l">
            <Card padding="m">
              <Text as="h3" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
                📈 时间线
              </Text>
              <Text typeLevel="body.medium">
                时间线功能开发中...
              </Text>
            </Card>
          </Box>
        </Tabs.Panel>

        <Tabs.Panel>
          <Box marginTop="l">
            <Card padding="m">
              <Text as="h3" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
                📚 历史版本
              </Text>
              <Text typeLevel="body.medium">
                历史版本功能开发中...
              </Text>
            </Card>
          </Box>
        </Tabs.Panel>

        <Tabs.Panel>
          <Box marginTop="l">
            <Card padding="m">
              <Text as="h3" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
                刷新 版本对比
              </Text>
              <Text typeLevel="body.medium">
                版本对比功能开发中...
              </Text>
            </Card>
          </Box>
        </Tabs.Panel>

      </Tabs>

      {/* 编辑表单 */}
      {!readonly && !isHistorical && (
        <OrganizationForm 
          organization={selectedOrg}
          isOpen={isFormOpen}
          onClose={handleFormClose}
          onSubmit={handleFormSubmit}
          temporalMode={temporalMode}
          isHistorical={isHistorical}
          enableTemporalFeatures={true}
        />
      )}
    </Box>
  );
};

export default OrganizationDetail;