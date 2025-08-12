import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Box } from '@workday/canvas-kit-react/layout';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';
import { Card } from '@workday/canvas-kit-react/card';

import { StatsCards } from './components/StatsCards';
import { OrganizationTable } from './components/OrganizationTable';
import { OrganizationForm } from './components/OrganizationForm';
import { OrganizationFilters } from './OrganizationFilters';
import { PaginationControls } from './PaginationControls';

import { useOrganizationDashboard } from './hooks/useOrganizationDashboard';
import { useOrganizationActions } from './hooks/useOrganizationActions';

// 时态管理组件导入 - 暂时禁用以修复无限循环错误
// import { TemporalNavbar } from '../temporal/components/TemporalNavbar';
// import { useTemporalMode, useTemporalQueryState } from '../../shared/hooks/useTemporalQuery';
// import type { TemporalMode } from '../../shared/types/temporal';

const DashboardHeader: React.FC<{
  onCreateClick: () => void;
  onCreatePlannedClick?: () => void;
  isToggling: boolean;
  temporalMode?: 'current' | 'historical';
  isHistorical?: boolean;
}> = ({ onCreateClick, onCreatePlannedClick, isToggling, isHistorical = false }) => (
  <Box marginBottom="l">
    <Heading size="large">
      组织架构管理
      {isHistorical && (
        <Text as="span" typeLevel="subtext.medium" color="hint" marginLeft="s">
          (历史视图)
        </Text>
      )}
    </Heading>
    <Box paddingTop="m">
      <PrimaryButton 
        marginRight="s" 
        onClick={onCreateClick}
        disabled={isToggling || isHistorical}
      >
        {isHistorical ? '新增组织单元 (历史模式禁用)' : '新增组织单元'}
      </PrimaryButton>
      
      {/* 计划组织创建按钮 */}
      {onCreatePlannedClick && !isHistorical && (
        <SecondaryButton 
          marginRight="s" 
          onClick={onCreatePlannedClick}
          disabled={isToggling}
          style={{ borderColor: '#1890ff', color: '#1890ff' }}
        >
          📅 新增计划组织
        </SecondaryButton>
      )}
      
      <SecondaryButton 
        marginRight="s"
        disabled={isToggling || isHistorical}
      >
        导入数据
      </SecondaryButton>
      <TertiaryButton disabled={isToggling || isHistorical}>
        导出报告
      </TertiaryButton>
      {isToggling && (
        <Text typeLevel="subtext.small" color="hint" marginLeft="m">
          正在更新组织状态...
        </Text>
      )}
      {isHistorical && (
        <Text typeLevel="subtext.small" color="hint" marginLeft="m">
          📖 当前查看历史数据，部分操作已禁用
        </Text>
      )}
    </Box>
  </Box>
);

const EmptyState: React.FC<{
  isFiltered: boolean;
  onClearFilters: () => void;
}> = ({ isFiltered, onClearFilters }) => (
  <Box padding="xl" textAlign="center">
    <Text>
      {isFiltered 
        ? '没有找到符合筛选条件的组织单元'
        : '暂无组织数据'
      }
    </Text>
    {isFiltered && (
      <Box marginTop="s">
        <SecondaryButton 
          size="small"
          onClick={onClearFilters}
        >
          清除筛选条件
        </SecondaryButton>
      </Box>
    )}
  </Box>
);

const LoadingState: React.FC = () => (
  <Box padding="l">
    <Text>加载组织数据中...</Text>
  </Box>
);

const ErrorState: React.FC<{ error: Error }> = ({ error }) => (
  <Box padding="l">
    <Text>加载失败: {error.message}</Text>
  </Box>
);

export const OrganizationDashboard: React.FC = () => {
  const navigate = useNavigate();

  // 传统组织数据和操作
  const {
    organizations,
    totalCount,
    stats,
    isLoading,
    isFetching,
    error,
    filters,
    isFiltered,
    setFilters,
    resetFilters,
    handlePageChange,
  } = useOrganizationDashboard();

  const {
    selectedOrg,
    isFormOpen,
    togglingId,
    isToggling,
    handleCreate,
    handleEdit,
    handleToggleStatus,
    handleFormClose,
    handleFormSubmit,
  } = useOrganizationActions();

  // 时态管理导航处理器
  const handleTemporalManage = (organizationCode: string) => {
    navigate(`/organizations/${organizationCode}/temporal`);
  };

  // 计划组织创建处理 - 重新启用
  const handleCreatePlanned = () => {
    const plannedOrgTemplate = {
      name: '',
      unit_type: 'DEPARTMENT',
      status: 'PLANNED',
      description: '',
      parent_code: '',
      level: 1,
      sort_order: 0,
      _isPlannedCreation: true
    };
    handleCreate();
  };

  // 时态管理状态和操作 - 暂时禁用以修复无限循环错误
  // const { mode: temporalMode, isHistorical, isCurrent, isPlanning } = useTemporalMode();
  // const { loading: temporalLoading, error: temporalError, context: temporalContext } = useTemporalQueryState();
  const temporalMode = 'current' as const;
  const isHistorical = false;
  const isPlanning = false;
  const temporalLoading = { organizations: false };

  // 时态模式变更处理 - 暂时禁用以修复无限循环错误
  // const handleTemporalModeChange = (newMode: TemporalMode) => {
  //   console.log(`时态模式切换到: ${newMode}`);
  // };

  if (isLoading || temporalLoading.organizations) {
    return <LoadingState />;
  }

  if (error) {
    return <ErrorState error={error} />;
  }

  const hasOrganizations = organizations && organizations.length > 0;

  return (
    <Box data-testid="organization-dashboard">
      {/* 时态导航栏 - 暂时禁用以修复无限循环错误 */}
      {/* <Box marginBottom="l">
        <TemporalNavbar
          onModeChange={handleTemporalModeChange}
          showAdvancedSettings={true}
        />
      </Box> */}

      <DashboardHeader 
        onCreateClick={handleCreate}
        onCreatePlannedClick={handleCreatePlanned}
        isToggling={isToggling}
        temporalMode={temporalMode}
        isHistorical={isHistorical}
      />
      
      {stats && <StatsCards stats={stats} />}
      
      <OrganizationFilters 
        filters={filters}
        onFiltersChange={setFilters}
      />
      
      <Card>
        <Card.Heading>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>
              组织单元列表
              {isHistorical && (
                <Text as="span" typeLevel="subtext.small" color="hint" marginLeft="s">
                  - 历史时点: {/* temporalContext?.asOfDate ? new Date(temporalContext.asOfDate).toLocaleDateString('zh-CN') : */ '历史模式'}
                </Text>
              )}
              {isPlanning && (
                <Text as="span" typeLevel="subtext.small" color="hint" marginLeft="s">
                  - 规划视图
                </Text>
              )}
            </span>
            {(isFetching || temporalLoading.organizations) && (
              <Text typeLevel="subtext.small" color="hint">
                {temporalLoading.organizations ? '加载时态数据中...' : '加载中...'}
              </Text>
            )}
          </div>
        </Card.Heading>
        <Card.Body>
          {hasOrganizations ? (
            <>
              <OrganizationTable
                organizations={organizations}
                onEdit={isHistorical ? undefined : handleEdit} // 历史模式禁用编辑
                onToggleStatus={isHistorical ? undefined : (code, status) => { handleToggleStatus(code, status); }} // 历史模式禁用状态切换
                onTemporalManage={handleTemporalManage} // 时态管理导航
                loading={isFetching || temporalLoading.organizations}
                togglingId={togglingId}
                temporalMode={temporalMode}
                isHistorical={isHistorical}
              />
              
              <PaginationControls
                currentPage={filters.page}
                totalCount={totalCount}
                pageSize={filters.pageSize}
                onPageChange={handlePageChange}
                disabled={isFetching || isToggling || temporalLoading.organizations}
              />
            </>
          ) : (
            <EmptyState 
              isFiltered={isFiltered}
              onClearFilters={resetFilters}
            />
          )}
        </Card.Body>
      </Card>

      {/* 组织表单 - 历史模式下禁用 */}
      {!isHistorical && (
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