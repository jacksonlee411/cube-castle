import React from 'react';
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

const DashboardHeader: React.FC<{
  onCreateClick: () => void;
  isToggling: boolean;
}> = ({ onCreateClick, isToggling }) => (
  <Box marginBottom="l">
    <Heading size="large">组织架构管理</Heading>
    <Box paddingTop="m">
      <PrimaryButton 
        marginRight="s" 
        onClick={onCreateClick}
        disabled={isToggling}
      >
        新增组织单元
      </PrimaryButton>
      <SecondaryButton 
        marginRight="s"
        disabled={isToggling}
      >
        导入数据
      </SecondaryButton>
      <TertiaryButton disabled={isToggling}>
        导出报告
      </TertiaryButton>
      {isToggling && (
        <Text typeLevel="subtext.small" color="hint" marginLeft="m">
          正在更新组织状态...
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

  if (isLoading) {
    return <LoadingState />;
  }

  if (error) {
    return <ErrorState error={error} />;
  }

  const hasOrganizations = organizations && organizations.length > 0;

  return (
    <Box data-testid="organization-dashboard">
      <DashboardHeader 
        onCreateClick={handleCreate}
        isToggling={isToggling}
      />
      
      {stats && <StatsCards stats={stats} />}
      
      <OrganizationFilters 
        filters={filters}
        onFiltersChange={setFilters}
      />
      
      <Card>
        <Card.Heading>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>组织单元列表</span>
            {isFetching && (
              <Text typeLevel="subtext.small" color="hint">
                加载中...
              </Text>
            )}
          </div>
        </Card.Heading>
        <Card.Body>
          {hasOrganizations ? (
            <>
              <OrganizationTable
                organizations={organizations}
                onEdit={handleEdit}
                onToggleStatus={handleToggleStatus}
                loading={isFetching}
                togglingId={togglingId}
              />
              
              <PaginationControls
                currentPage={filters.page}
                totalCount={totalCount}
                pageSize={filters.pageSize}
                onPageChange={handlePageChange}
                disabled={isFetching || isToggling}
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

      <OrganizationForm 
        organization={selectedOrg}
        isOpen={isFormOpen}
        onClose={handleFormClose}
        onSubmit={handleFormSubmit}
      />
    </Box>
  );
};