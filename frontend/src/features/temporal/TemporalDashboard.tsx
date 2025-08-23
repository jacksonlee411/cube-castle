/**
 * 组织详情集成组件
 * 整合所有时态功能到一个统一的界面中
 */
import React, { useState, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { Tabs, useTabsModel } from '@workday/canvas-kit-react/tabs';
import { SecondaryButton } from '@workday/canvas-kit-react/button';
import { space } from '@workday/canvas-kit-react/tokens';
import {
  TemporalNavbar,
  TemporalTable,
  Timeline,
  VersionComparison
} from './components';
import type { 
  OrganizationUnit,
  OrganizationQueryParams 
} from '../../shared/types/organization';

export interface TemporalDashboardProps {
  /** 初始查询参数 */
  initialQueryParams?: OrganizationQueryParams;
  /** 是否紧凑模式 */
  compact?: boolean;
  /** 是否显示高级功能 */
  showAdvancedFeatures?: boolean;
}

/**
 * 组织详情仪表板组件
 * 集成了所有组织详情功能的主界面
 */
export const TemporalDashboard: React.FC<TemporalDashboardProps> = ({
  initialQueryParams,
  compact = false,
  showAdvancedFeatures = true
}) => {
  // 状态管理
  const [activeTab] = useState('table');
  
  // Tabs模型 (Canvas Kit v13)
  const tabsModel = useTabsModel({
    initialTab: activeTab
  });
  const [selectedOrganization, setSelectedOrganization] = useState<OrganizationUnit | null>(null);
  const [showTimelineModal, setShowTimelineModal] = useState(false);
  const [showVersionModal, setShowVersionModal] = useState(false);
  const [queryParams] = useState<OrganizationQueryParams>(
    initialQueryParams || {}
  );

  // 事件处理
  const handleRowClick = useCallback((organization: OrganizationUnit) => {
    setSelectedOrganization(organization);
  }, []);

  const handleViewTimeline = useCallback((organization: OrganizationUnit) => {
    setSelectedOrganization(organization);
    setShowTimelineModal(true);
  }, []);

  const handleViewHistory = useCallback((organization: OrganizationUnit) => {
    setSelectedOrganization(organization);
    setShowVersionModal(true);
  }, []);

  const handleEdit = useCallback((organization: OrganizationUnit) => {
    console.log('编辑组织:', organization);
  }, []);

  const handleDelete = useCallback((organization: OrganizationUnit) => {
    console.log('删除组织:', organization);
  }, []);

  return (
    <Box>
      {/* 时态导航栏 */}
      <Box marginBottom={space.m}>
        <TemporalNavbar 
          compact={compact}
          showAdvancedSettings={showAdvancedFeatures}
        />
      </Box>

      {/* 主要内容区域 */}
      <Card padding={space.m}>
        <Tabs model={tabsModel}>
          <Tabs.List>
            <Tabs.Item name="table">组织列表</Tabs.Item>
            <Tabs.Item name="timeline">时间线视图</Tabs.Item>
            {showAdvancedFeatures && <Tabs.Item name="comparison">版本对比</Tabs.Item>}
          </Tabs.List>

          {/* 组织列表标签页 */}
          <Tabs.Panel>
            <Box marginTop={space.m}>
              <TemporalTable
                queryParams={queryParams}
                showTemporalIndicators={true}
                showActions={true}
                showSelection={showAdvancedFeatures}
                compact={compact}
                onRowClick={handleRowClick}
                onEdit={handleEdit}
                onDelete={handleDelete}
                onViewHistory={handleViewHistory}
                onViewTimeline={handleViewTimeline}
              />
            </Box>
          </Tabs.Panel>
          
          {/* 时间线视图标签页 */}
          <Tabs.Panel>
            <Box marginTop={space.m}>
              <Text>时间线视图功能开发中...</Text>
            </Box>
          </Tabs.Panel>
          
          {/* 版本对比标签页 */}
          {showAdvancedFeatures && (
            <Tabs.Panel>
              <Box marginTop={space.m}>
                <Text>版本对比功能开发中...</Text>
              </Box>
            </Tabs.Panel>
          )}
        </Tabs>
      </Card>

      {/* 时间线弹窗 */}
      {showTimelineModal && selectedOrganization && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1000
          }}
        >
          <Card padding={space.l} minWidth="800px" maxHeight="80vh" overflow="auto">
            <Flex justifyContent="space-between" alignItems="center" marginBottom={space.m}>
              <Text fontSize="large" fontWeight="bold">
                {selectedOrganization.name} - 时间线
              </Text>
              <SecondaryButton onClick={() => setShowTimelineModal(false)}>
                关闭
              </SecondaryButton>
            </Flex>
            <Timeline
              organizationCode={selectedOrganization.code}
              compact={false}
              showFilters={true}
              showActions={true}
            />
          </Card>
        </div>
      )}

      {/* 版本对比弹窗 */}
      {showVersionModal && selectedOrganization && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1000
          }}
        >
          <Card padding={space.l} minWidth="1000px" maxHeight="80vh" overflow="auto">
            <Flex justifyContent="space-between" alignItems="center" marginBottom={space.m}>
              <Text fontSize="large" fontWeight="bold">
                {selectedOrganization.name} - 版本历史
              </Text>
              <SecondaryButton onClick={() => setShowVersionModal(false)}>
                关闭
              </SecondaryButton>
            </Flex>
            <VersionComparison
              organizationCode={selectedOrganization.code}
              compact={false}
            />
          </Card>
        </div>
      )}
    </Box>
  );
};

export default TemporalDashboard;