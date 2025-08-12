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
import { Tabs } from '@workday/canvas-kit-react/tabs';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';

// 组织管理和时态功能导入
import { OrganizationForm } from '../organizations/components/OrganizationForm';
import { OrganizationTable } from '../organizations/components/OrganizationTable';
import { Timeline } from '../temporal/components/Timeline';
import { VersionComparison } from '../temporal/components/VersionComparison';
import { TemporalNavbar } from '../temporal/components/TemporalNavbar';

// Hooks导入
import { useTemporalOrganization, useOrganizationHistory, useOrganizationTimeline, useTemporalMode } from '../../shared/hooks/useTemporalQuery';
import { useOrganizationActions } from '../organizations/hooks/useOrganizationActions';

// Types导入
import type { OrganizationUnit } from '../../shared/types/organization';
import type { TimelineEvent, TemporalOrganizationUnit, TemporalMode } from '../../shared/types/temporal';

export interface OrganizationDetailProps {
  /** 组织编码 */
  organizationCode: string;
  /** 是否只读模式 */
  readonly?: boolean;
  /** 返回回调 */
  onBack?: () => void;
  /** 组织更新回调 */
  onOrganizationUpdated?: (organization: OrganizationUnit) => void;
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
    return <Badge color={config.color as any}>{config.label}</Badge>;
  };

  const getUnitTypeName = (unitType: string) => {
    const typeNames = {
      'COMPANY': '公司',
      'DEPARTMENT': '部门',
      'COST_CENTER': '成本中心',
      'PROJECT_TEAM': '项目团队'
    };
    return typeNames[unitType as keyof typeof typeNames] || unitType;
  };

  return (
    <Card padding="m">
      <Flex justifyContent="space-between" alignItems="flex-start" marginBottom="m">
        <Box flex="1">
          <Flex alignItems="center" gap="s" marginBottom="s">
            <Heading size="medium">{organization.name}</Heading>
            {getStatusBadge(organization.status)}
            {isHistorical && (
              <Badge color="blueberry600">历史视图</Badge>
            )}
          </Flex>
          
          <Text typeLevel="subtext.medium" color="hint" marginBottom="s">
            编码: {organization.code} • 类型: {getUnitTypeName(organization.unit_type)} • 层级: {organization.level}
          </Text>
          
          {organization.description && (
            <Text typeLevel="body.medium" marginBottom="s">
              {organization.description}
            </Text>
          )}
          
          <Box display="flex" gap="m" marginBottom="s">
            {organization.parent_code && (
              <Text typeLevel="subtext.small">
                上级组织: {organization.parent_code}
              </Text>
            )}
            <Text typeLevel="subtext.small">
              排序: {organization.sort_order}
            </Text>
          </Box>
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
  onBack,
  onOrganizationUpdated
}) => {
  const [activeTab, setActiveTab] = useState<'overview' | 'history' | 'timeline' | 'comparison'>('overview');

  // 时态模式管理
  const { mode: temporalMode, isHistorical, isCurrent } = useTemporalMode();

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
    isLoading: historyLoading,
    hasHistory,
    latestVersion
  } = useOrganizationHistory(organizationCode, { limit: 20 });

  // 时间线事件查询
  const {
    data: timelineEvents = [],
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
    togglingId,
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
      handleToggleStatus(organization.code);
    }
  }, [organization, handleToggleStatus]);

  // 时间线事件点击处理
  const handleTimelineEventClick = useCallback((event: TimelineEvent) => {
    // 实现事件详情显示或跳转
    console.log('Timeline event clicked:', event);
    alert(`查看事件详情:\n\n${event.title}\n${event.description || ''}\n\n日期: ${new Date(event.eventDate).toLocaleString('zh-CN')}`);
  }, []);

  // 版本比较处理
  const handleVersionComparison = useCallback((version1: TemporalOrganizationUnit, version2: TemporalOrganizationUnit) => {
    // 设置版本对比标签
    setActiveTab('comparison');
  }, []);

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
            <PrimaryButton onClick={refetchOrganization} marginRight="s">
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
              🔄 刷新
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
      <Tabs>
        <Tabs.List>
          <Tabs.Item 
            name="overview"
            onClick={() => setActiveTab('overview')}
            isActive={activeTab === 'overview'}
          >
            概览信息
          </Tabs.Item>
          <Tabs.Item 
            name="timeline"
            onClick={() => setActiveTab('timeline')}
            isActive={activeTab === 'timeline'}
          >
            时间线 {hasTimelineEvents && <Badge color="blueberry600">{eventCount}</Badge>}
          </Tabs.Item>
          <Tabs.Item 
            name="history"
            onClick={() => setActiveTab('history')}
            isActive={activeTab === 'history'}
          >
            历史版本 {hasHistory && <Badge color="greenFresca600">{historyVersions.length}</Badge>}
          </Tabs.Item>
          <Tabs.Item 
            name="comparison"
            onClick={() => setActiveTab('comparison')}
            isActive={activeTab === 'comparison'}
          >
            版本对比
          </Tabs.Item>
        </Tabs.List>

        <Tabs.Panel name="overview">
          <Box marginTop="l">
            <Card padding="m">
              <Text as="h3" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
                📋 组织概览信息
              </Text>
              
              <Box display="grid" gridTemplateColumns="repeat(auto-fit, minmax(250px, 1fr))" gap="m">
                <Box>
                  <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">基本信息</Text>
                  <Text typeLevel="body.small">编码: {organization.code}</Text>
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
              </Box>
            </Card>
          </Box>
        </Tabs.Panel>

        <Tabs.Panel name="timeline">
          <Box marginTop="l">
            <Timeline
              organizationCode={organizationCode}
              queryParams={{ limit: 50 }}
              compact={false}
              maxEvents={50}
              showFilters={true}
              showActions={!readonly}
              onEventClick={handleTimelineEventClick}
              onAddEvent={readonly ? undefined : () => alert('添加事件功能开发中')}
            />
          </Box>
        </Tabs.Panel>

        <Tabs.Panel name="history">
          <Box marginTop="l">
            <Card padding="m">
              <Text as="h3" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
                📚 历史版本记录
              </Text>
              
              {historyLoading ? (
                <Flex justifyContent="center" padding="l">
                  <LoadingDots />
                  <Text marginLeft="m">加载历史版本...</Text>
                </Flex>
              ) : hasHistory ? (
                <Box>
                  <Text typeLevel="body.medium" marginBottom="m">
                    共 {historyVersions.length} 个历史版本
                  </Text>
                  {/* 这里可以展示历史版本列表，或者复用OrganizationTable组件 */}
                  <Box>
                    {historyVersions.slice(0, 5).map((version, index) => (
                      <Box
                        key={version.version || index}
                        padding="s"
                        marginBottom="s"
                        style={{ 
                          backgroundColor: index === 0 ? '#f0f7ff' : '#f8f9fa',
                          borderRadius: '4px',
                          border: '1px solid #e9ecef'
                        }}
                      >
                        <Flex justifyContent="space-between" alignItems="center">
                          <Text typeLevel="body.medium">
                            版本 {version.version} - {version.name}
                          </Text>
                          <Text typeLevel="subtext.small" color="hint">
                            {version.effectiveFrom ? new Date(version.effectiveFrom).toLocaleString('zh-CN') : ''}
                          </Text>
                        </Flex>
                        {version.changeReason && (
                          <Text typeLevel="subtext.small" color="hint" marginTop="xs">
                            变更原因: {version.changeReason}
                          </Text>
                        )}
                      </Box>
                    ))}
                  </Box>
                </Box>
              ) : (
                <Text color="hint">暂无历史版本记录</Text>
              )}
            </Card>
          </Box>
        </Tabs.Panel>

        <Tabs.Panel name="comparison">
          <Box marginTop="l">
            {hasHistory && historyVersions.length >= 2 ? (
              <VersionComparison
                organizationCode={organizationCode}
                version1={historyVersions[0]}
                version2={historyVersions[1]}
                compact={false}
                showMetadata={true}
              />
            ) : (
              <Card padding="l">
                <Text textAlign="center" color="hint">
                  需要至少2个历史版本才能进行对比
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