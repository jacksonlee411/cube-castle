/**
 * 组织详情页面 - 基础信息和审计历史
 * 展示组织的详细信息和审计历史记录
 */
import React, { useState, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';
import { Card } from '@workday/canvas-kit-react/card';
import { Badge } from '../../../shared/components/Badge';
import { Tabs, useTabsModel } from '@workday/canvas-kit-react/tabs';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { exclamationCircleIcon, activityStreamIcon } from '@workday/canvas-system-icons-web';

// 审计历史组件导入
import { AuditHistorySection } from '../../audit/components/AuditHistorySection';

// 组织管理功能导入
import { OrganizationForm } from './OrganizationForm';
// import { TemporalNavbar } from '../../temporal/components/TemporalNavbar'; // 已删除

// Hooks导入 - 移除已删除的时态钩子
// import { useTemporalOrganization, useOrganizationHistory, useOrganizationTimeline, useTemporalMode } from '../../../shared/hooks/useTemporalQuery';
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
            {getUnitTypeBadge(organization.unitType)}
            {isHistorical && (
              <Badge color="blueberry600">历史视图</Badge>
            )}
          </Flex>
          
          <Text typeLevel="subtext.medium" color="hint" marginBottom="s">
            编码: {organization.code} • 类型: {getUnitTypeName(organization.unitType)} • 层级: {organization.level}
            {organization.recordId && (
              <>
                <br />
                UUID: {organization.recordId}
              </>
            )}
          </Text>
          
          {organization.description && (
            <Text typeLevel="body.medium" marginBottom="s">
              {organization.description}
            </Text>
          )}
          
          <Flex gap="m" marginBottom="s">
            {organization.parentCode && (
              <Text typeLevel="subtext.small">
                上级组织: {organization.parentCode}
              </Text>
            )}
            <Text typeLevel="subtext.small">
              排序: {organization.sortOrder}
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
          创建时间: {organization.createdAt ? new Date(organization.createdAt).toLocaleString('zh-CN') : '未知'}
        </Text>
        {organization.updatedAt && (
          <Text typeLevel="subtext.small" color="hint">
            更新时间: {new Date(organization.updatedAt).toLocaleString('zh-CN')}
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
  organizationCode: _organizationCode,
  readonly = false,
  onBack
}) => {
  // 状态管理
  const [activeTab] = useState('overview');
  
  // Tabs模型 (Canvas Kit v13)
  const tabsModel = useTabsModel({
    initialTab: activeTab
  });

  // 临时状态管理 - 替代已删除的时态钩子
  const [temporalMode] = useState<TemporalMode>('current');
  const [organization] = useState<OrganizationUnit | null>({
    code: 'DEMO-001',
    name: '演示部门',
    unitType: 'DEPARTMENT',
    status: 'ACTIVE',
    level: 1,
    path: '/DEMO-001',
    sortOrder: 1,
    description: '用于演示审计历史功能',
    createdAt: '2025-08-31T10:00:00Z',
    updatedAt: '2025-08-31T10:30:00Z',
    recordId: '134d58aa-ce9e-4631-8002-a1e2a01872c1' // 关键：有审计记录的ID
  });
  const [orgLoading] = useState(false);
  const [orgError] = useState(false);
  const [orgErrorMessage] = useState<string>('');
  const isHistorical = temporalMode === 'historical';
  
  // 模拟refetch函数
  const refetchOrganization = useCallback(() => {
    console.log('Refetch organization - placeholder');
  }, []);
  
  const refetchTimeline = useCallback(() => {
    console.log('Refetch timeline - placeholder');
  }, []);

  // 时间线状态 - 仅保留审计历史相关
  const [timelineLoading] = useState(false);
  const [hasTimelineEvents] = useState(false);
  const [eventCount] = useState(0);

  // 组织操作钩子
  const {
    selectedOrg,
    isFormOpen,
    handleEdit,
    handleFormClose,
    handleFormSubmit,
  } = useOrganizationActions();

  // 时态模式变更处理
  // const handleTemporalModeChange = useCallback((newMode: TemporalMode) => {
  //   console.log(`时态模式切换到: ${newMode}，重新加载组织数据`);
  //   refetchOrganization();
  // }, [refetchOrganization]);

  // 编辑组织处理
  const handleEditOrganization = useCallback(() => {
    if (organization) {
      handleEdit(organization);
    }
  }, [organization, handleEdit]);

  // 切换状态处理 - 临时禁用直到实现状态管理
  const handleToggleOrganizationStatus = useCallback(() => {
    if (organization) {
      console.log('Toggle status not implemented yet');
    }
  }, [organization]);

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
          <Flex alignItems="center" gap="xs" marginBottom="m">
            <SystemIcon icon={exclamationCircleIcon} size={20} color="cinnamon600" />
            <Text color="cinnamon600" typeLevel="heading.medium">
              加载组织详情失败
            </Text>
          </Flex>
          <Text marginBottom="m">
            {orgErrorMessage || '无法加载组织信息，请检查组织编码或网络连接'}
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
        {/* <TemporalNavbar
          onModeChange={handleTemporalModeChange}
          showAdvancedSettings={true}
        /> */}
        <Text>时态导航栏组件已移除 - 正在重构中</Text>
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
          isLoading={false}
        />
      </Box>

      {/* 详情标签页 */}
      <Tabs model={tabsModel}>
        <Tabs.List>
          <Tabs.Item data-id="overview">
            概览信息
          </Tabs.Item>
          <Tabs.Item data-id="audit">
            审计历史 {hasTimelineEvents && <Badge color="blueberry600">{eventCount}</Badge>}
          </Tabs.Item>
        </Tabs.List>

        <Tabs.Panel data-id="overview">
          <Box marginTop="l">
            <Card padding="m">
              <Text as="h3" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
                详情 组织概览信息
              </Text>
              
              <Flex flexDirection="column" gap="m">
                <Box>
                  <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">基本信息</Text>
                  <Text typeLevel="body.small">编码: {organization.code}</Text>
                  {organization.recordId && (
                    <Text typeLevel="body.small">UUID: {organization.recordId}</Text>
                  )}
                  <Text typeLevel="body.small">名称: {organization.name}</Text>
                  <Text typeLevel="body.small">状态: {organization.status}</Text>
                  <Text typeLevel="body.small">类型: {organization.unitType}</Text>
                </Box>
                
                <Box>
                  <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">层级结构</Text>
                  <Text typeLevel="body.small">层级: {organization.level}</Text>
                  <Text typeLevel="body.small">上级: {organization.parentCode || '无'}</Text>
                  <Text typeLevel="body.small">排序: {organization.sortOrder}</Text>
                </Box>
                
                <Box>
                  <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">时间信息</Text>
                  <Text typeLevel="body.small">创建: {organization.createdAt ? new Date(organization.createdAt).toLocaleDateString('zh-CN') : '未知'}</Text>
                  <Text typeLevel="body.small">更新: {organization.updatedAt ? new Date(organization.updatedAt).toLocaleDateString('zh-CN') : '未知'}</Text>
                </Box>
                
                {hasTimelineEvents && (
                  <Box>
                    <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">活动统计</Text>
                    <Text typeLevel="body.small">审计事件: {eventCount} 个</Text>
                  </Box>
                )}
              </Flex>
            </Card>
          </Box>
        </Tabs.Panel>

        <Tabs.Panel data-id="audit">
          <Box marginTop="l">
            {organization?.recordId ? (
              <AuditHistorySection
                recordId={organization.recordId}
                params={{
                  limit: 50,
                  mode: temporalMode
                }}
              />
            ) : (
              <Card padding="m">
                <Flex alignItems="center" gap="xs" marginBottom="m">
                  <SystemIcon icon={activityStreamIcon} size={16} />
                  <Text as="h3" typeLevel="subtext.large" fontWeight="bold">
                    审计历史
                  </Text>
                </Flex>
                <Text typeLevel="body.medium" color="hint">
                  需要组织记录ID才能查看审计历史
                </Text>
              </Card>
            )}
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