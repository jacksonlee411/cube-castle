import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Box } from '@workday/canvas-kit-react/layout';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';
import { Card } from '@workday/canvas-kit-react/card';

import { OrganizationTable } from './components/OrganizationTable';
import { OrganizationForm } from './components/OrganizationForm';
import { OrganizationFilters } from './OrganizationFilters';
import { PaginationControls } from './PaginationControls';

import { useEnterpriseOrganizations } from '../../shared/hooks/useEnterpriseOrganizations';
import { copyText } from '../../shared/utils/clipboard';
import type { OrganizationUnit } from '../../shared/types/organization';
import { OrganizationBreadcrumb } from '../../shared/components/OrganizationBreadcrumb';
// import { useOrganizationMutations } from '../../shared/hooks/useOrganizationMutations'; // TODO: Implement mutations

// 组织详情组件导入 - 暂时禁用以修复无限循环错误

const DashboardHeader: React.FC<{
  onCreateClick: () => void;
  temporalMode?: 'current' | 'historical';
  isHistorical?: boolean;
}> = ({ onCreateClick, isHistorical = false }) => (
  <Box marginBottom="l">
    <Box marginBottom="s">
      <OrganizationBreadcrumb namePath="/组织列表" />
    </Box>
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
        disabled={isHistorical}
      >
        {isHistorical ? '新增组织单元 (历史模式禁用)' : '新增组织单元'}
      </PrimaryButton>
      
      {/* ❌ 已移除计划组织创建按钮 - 简化API设计 */}
      {/* {onCreatePlannedClick && !isHistorical && (
        <SecondaryButton 
          marginRight="s" 
          onClick={onCreatePlannedClick}
          style={{ borderColor: '#1890ff', color: '#1890ff' }}
        >
          计划 新增计划组织
        </SecondaryButton>
      )} */}
      
      <SecondaryButton 
        marginRight="s"
        disabled={isHistorical}
      >
        导入数据
      </SecondaryButton>
      <TertiaryButton disabled={isHistorical}>
        导出报告
      </TertiaryButton>
      <TertiaryButton
        onClick={async () => {
          const deepLink = `${window.location.origin}/organizations`;
          await copyText(deepLink);
        }}
        marginLeft="s"
      >
        复制列表链接
      </TertiaryButton>
      {isHistorical && (
        <Text typeLevel="subtext.small" color="hint" marginLeft="m">
          当前查看历史数据，部分操作已禁用
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

export const OrganizationDashboard: React.FC = () => {
  const navigate = useNavigate();

  // 简化的filter状态管理
  const [filters, setFilters] = React.useState({ 
    searchText: '',
    unitType: undefined as string | undefined,
    status: undefined as string | undefined,
    level: undefined as number | undefined,
    page: 1, 
    pageSize: 50 
  });
  const isFiltered = false;
  const resetFilters = () => setFilters({ 
    searchText: '',
    unitType: undefined,
    status: undefined,
    level: undefined,
    page: 1, 
    pageSize: 50 
  });
  const handlePageChange = (page: number) => setFilters(prev => ({ ...prev, page }));

  // 组织数据查询
  const { organizations, loading: isLoading, error } = useEnterpriseOrganizations();
  // useEnterpriseOrganizations 已返回 organizations 数组
  const totalCount = organizations.length; // 使用数组长度作为总数

  // 组织操作(暂时简化)
  const selectedOrg: OrganizationUnit | undefined = undefined;
  const isFormOpen = false;
  const handleFormClose = () => {};
  const handleFormSubmit = () => {};

  // 新建组织处理器 - 修改为页面跳转而不是打开Modal
  const handleCreateOrganization = () => {
    navigate('/organizations/new');
  };

  // 组织详情导航处理器
  const handleTemporalManage = (organizationCode: string) => {
    navigate(`/organizations/${organizationCode}/temporal`);
  };


  const temporalMode = 'current' as const;
  const isHistorical = false;
  const isPlanning = false;
  const temporalLoading = { organizations: false };
  const isFetching = isLoading; // 使用isLoading作为isFetching


  if (isLoading || temporalLoading.organizations) {
    return <LoadingState />;
  }

  // 🔧 修复: 保持界面结构完整性，不因错误而隐藏所有UI组件

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
        onCreateClick={handleCreateOrganization}
        // onCreatePlannedClick={handleCreatePlanned} // ❌ 已移除
        temporalMode={temporalMode}
        isHistorical={isHistorical}
      />
      
      
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
          {error ? (
            <Box padding="l" style={{ textAlign: 'center' }}>
              <Text color="cinnamon600" fontWeight="medium" marginBottom="m">
                ⚠️ 数据加载失败
              </Text>
              <Text color="frenchVanilla500" marginBottom="m">
                {typeof error === 'string' ? error : (error as Error)?.message || '未知错误'}
              </Text>
              <SecondaryButton 
                onClick={() => window.location.reload()}
              >
                重新加载
              </SecondaryButton>
            </Box>
          ) : hasOrganizations ? (
            <>
              <OrganizationTable
                organizations={organizations}
                onTemporalManage={handleTemporalManage} // 组织详情导航
                temporalMode={temporalMode}
                isHistorical={isHistorical}
              />
              
              <PaginationControls
                currentPage={filters.page}
                totalCount={totalCount}
                pageSize={filters.pageSize}
                onPageChange={handlePageChange}
                disabled={isFetching || temporalLoading.organizations}
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
