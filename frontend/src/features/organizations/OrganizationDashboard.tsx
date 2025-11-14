import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Box } from '@workday/canvas-kit-react/layout';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';
import { Card } from '@workday/canvas-kit-react/card';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

import { OrganizationTable } from './components/OrganizationTable';
import { OrganizationFilters } from './OrganizationFilters';
import { PaginationControls } from './PaginationControls';

import { useEnterpriseOrganizations } from '../../shared/hooks/useEnterpriseOrganizations';
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
        data-testid="create-organization-button"
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
        data-testid="import-data-button"
      >
        导入数据
      </SecondaryButton>
      <TertiaryButton
        disabled={isHistorical}
        data-testid="export-report-button"
      >
        导出报告
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
    pageSize: 50,
  });

  const pageSize = filters.pageSize || 50;

  const isFiltered = Boolean(
    filters.searchText?.trim() ||
    filters.unitType ||
    filters.status ||
    filters.level !== undefined,
  );

  const resetFilters = () =>
    setFilters({
      searchText: '',
      unitType: undefined,
      status: undefined,
      level: undefined,
      page: 1,
      pageSize: 50,
    });

  const handlePageChange = (page: number) =>
    setFilters(prev => ({ ...prev, page }));

  // 组织数据查询
  const { organizations, loading: isLoading, error } = useEnterpriseOrganizations();

  const filteredOrganizations = React.useMemo(() => {
    return organizations.filter(org => {
      if (filters.searchText?.trim()) {
        const keyword = filters.searchText.trim().toLowerCase();
        const nameMatch = org.name.toLowerCase().includes(keyword);
        const codeMatch = org.code.toLowerCase().includes(keyword);
        if (!nameMatch && !codeMatch) {
          return false;
        }
      }

      if (filters.unitType && org.unitType !== filters.unitType) {
        return false;
      }

      if (filters.status && org.status !== filters.status) {
        return false;
      }

      if (typeof filters.level === 'number' && org.level !== filters.level) {
        return false;
      }

      return true;
    });
  }, [organizations, filters.searchText, filters.unitType, filters.status, filters.level]);

  const totalCount = filteredOrganizations.length;
  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));
  const currentPage = Math.min(filters.page, totalPages);

  React.useEffect(() => {
    if (filters.page > totalPages) {
      setFilters(prev => ({ ...prev, page: totalPages }));
    }
  }, [filters.page, totalPages]);

  const paginatedOrganizations = React.useMemo(() => {
    if (totalCount === 0) {
      return [];
    }
    const startIndex = (currentPage - 1) * pageSize;
    return filteredOrganizations.slice(startIndex, startIndex + pageSize);
  }, [filteredOrganizations, currentPage, pageSize, totalCount]);

  // 组织操作(暂时简化)
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

  const hasOrganizations = totalCount > 0;

  return (
    <Box data-testid={temporalEntitySelectors.organization.dashboardWrapper}>
      <Box data-testid={temporalEntitySelectors.organization.dashboard}>
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
                  organizations={paginatedOrganizations}
                  onTemporalManage={handleTemporalManage} // 组织详情导航
                  temporalMode={temporalMode}
                  isHistorical={isHistorical}
                />
                
                <PaginationControls
                  currentPage={currentPage}
                  totalCount={totalCount}
                  pageSize={pageSize}
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
      </Box>
    </Box>
  );
};
