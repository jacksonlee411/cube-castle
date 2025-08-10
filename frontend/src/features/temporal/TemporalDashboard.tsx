/**
 * 时态管理集成组件
 * 整合所有时态功能到一个统一的界面中
 */
import React, { useState, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { Tabs, TabsList, Tab, TabPanel } from '@workday/canvas-kit-react/tabs';
import { PrimaryButton } from '@workday/canvas-kit-react/button';
import { Modal } from '@workday/canvas-kit-react/modal';
import { colors, space } from '@workday/canvas-kit-react/tokens';
import {
  TemporalNavbar,
  TemporalTable,
  Timeline,
  VersionComparison,
  DateTimePicker
} from './components';
import type { 
  OrganizationUnit,
  OrganizationQueryParams 
} from '../../shared/types/organization';
import type { 
  TemporalOrganizationUnit,
  TimelineEvent 
} from '../../shared/types/temporal';

export interface TemporalDashboardProps {
  /** 初始查询参数 */
  initialQueryParams?: OrganizationQueryParams;
  /** 是否紧凑模式 */
  compact?: boolean;
  /** 是否显示高级功能 */
  showAdvancedFeatures?: boolean;
}

/**
 * 时态管理仪表板组件
 * 集成了所有时态管理功能的主界面
 */
export const TemporalDashboard: React.FC<TemporalDashboardProps> = ({
  initialQueryParams,
  compact = false,
  showAdvancedFeatures = true
}) => {
  // 状态管理
  const [activeTab, setActiveTab] = useState('table');
  const [selectedOrganization, setSelectedOrganization] = useState<OrganizationUnit | null>(null);
  const [showTimelineModal, setShowTimelineModal] = useState(false);
  const [showVersionModal, setShowVersionModal] = useState(false);
  const [queryParams, setQueryParams] = useState<OrganizationQueryParams>(
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
    // TODO: 实现编辑功能
    console.log('编辑组织:', organization);
  }, []);

  const handleDelete = useCallback((organization: OrganizationUnit) => {
    // TODO: 实现删除功能
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
        <Tabs activeKey={activeTab} onSelectionChange={setActiveTab}>
          <TabsList>
            <Tab value="table">组织列表</Tab>
            <Tab value="timeline">时间线视图</Tab>
            {showAdvancedFeatures && <Tab value="comparison">版本对比</Tab>}
          </TabsList>

          {/* 组织列表标签页 */}
          <TabPanel value="table">
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
          </TabPanel>

          {/* 时间线视图标签页 */}
          <TabPanel value="timeline">
            <Box marginTop={space.m}>
              {selectedOrganization ? (
                <Timeline
                  organizationCode={selectedOrganization.code}
                  compact={compact}
                  showFilters={true}
                  showActions={showAdvancedFeatures}
                />
              ) : (
                <Box 
                  padding={space.l}
                  textAlign="center"
                  backgroundColor={colors.soap100}
                  borderRadius="8px"
                >
                  <Text color={colors.licorice500}>
                    请从组织列表中选择一个组织来查看其时间线
                  </Text>
                </Box>
              )}
            </Box>
          </TabPanel>

          {/* 版本对比标签页 */}
          {showAdvancedFeatures && (
            <TabPanel value="comparison">
              <Box marginTop={space.m}>
                {selectedOrganization ? (
                  <VersionComparison
                    organizationCode={selectedOrganization.code}
                    compact={compact}
                  />
                ) : (
                  <Box 
                    padding={space.l}
                    textAlign="center"
                    backgroundColor={colors.soap100}
                    borderRadius="8px"
                  >
                    <Text color={colors.licorice500}>
                      请从组织列表中选择一个组织来查看版本对比
                    </Text>
                  </Box>
                )}
              </Box>
            </TabPanel>
          )}
        </Tabs>
      </Card>

      {/* 时间线弹窗 */}
      {showTimelineModal && selectedOrganization && (
        <Modal onClose={() => setShowTimelineModal(false)}>
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
        </Modal>
      )}

      {/* 版本对比弹窗 */}
      {showVersionModal && selectedOrganization && (
        <Modal onClose={() => setShowVersionModal(false)}>
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
        </Modal>
      )}
    </Box>
  );
};

export default TemporalDashboard;