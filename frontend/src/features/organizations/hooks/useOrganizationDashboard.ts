import { useState, useMemo } from 'react';
import { useOrganizations, useOrganizationStats } from '../../../shared/hooks/useOrganizations';
// import { useTemporalOrganizations, useTemporalMode, useTemporalQueryState } from '../../../shared/hooks/useTemporalQuery';
import type { OrganizationQueryParams } from '../../../shared/types/organization';
import type { OrganizationUnitType, OrganizationStatus } from '../../../shared/types/api';
import type { FilterState } from '../OrganizationFilters';

const initialFilters: FilterState = {
  searchText: '',
  unit_type: undefined,
  status: undefined,
  level: undefined,
  page: 1,
  pageSize: 100, // 增加默认页面大小以确保所有组织都能显示
};

export const useOrganizationDashboard = () => {
  const [filters, setFilters] = useState<FilterState>(initialFilters);

  // 时态模式和状态 - 暂时禁用
  // const { isCurrent, isHistorical, isPlanning } = useTemporalMode();
  // const { context: temporalContext } = useTemporalQueryState();

  // Convert filters to query parameters
  const queryParams: OrganizationQueryParams = useMemo(() => ({
    searchText: filters.searchText || undefined,
    unit_type: (filters.unit_type as OrganizationUnitType) || undefined,
    status: (filters.status as OrganizationStatus) || undefined,
    level: filters.level || undefined,
    page: filters.page,
    pageSize: filters.pageSize,
  }), [filters]);

  // 根据时态模式选择使用不同的数据获取策略 - 暂时只使用传统数据
  const useTemporalData = false; // isHistorical || isPlanning;
  
  // 传统数据获取（当前模式）
  const { 
    data: organizationData, 
    isLoading: traditionalLoading, 
    error: traditionalError, 
    isFetching: traditionalFetching 
  } = useOrganizations(queryParams); // 移除不支持的enabled选项

  // 时态数据获取（历史/规划模式）- 暂时禁用
  // const {
  //   data: temporalOrganizations,
  //   isLoading: temporalLoading,
  //   error: temporalError,
  //   isFetching: temporalFetching,
  //   temporalContext: temporalOrgContext,
  //   isHistorical: temporalIsHistorical
  // } = useTemporalOrganizations(queryParams);

  // 统一数据输出 - 只使用传统数据
  const organizations = organizationData?.organizations || [];
  const totalCount = organizationData?.total_count || 0;
  const isLoading = traditionalLoading;
  const isFetching = traditionalFetching;
  const error = traditionalError;

  const { data: stats } = useOrganizationStats(); // 移除不支持的enabled选项

  const resetFilters = () => {
    setFilters(initialFilters);
  };

  const updateFilters = (newFilters: Partial<FilterState>) => {
    setFilters(prev => ({ ...prev, ...newFilters }));
  };

  const handlePageChange = (page: number) => {
    setFilters(prev => ({ ...prev, page }));
  };

  const isFiltered = useMemo(() => {
    return !!(filters.searchText || filters.unit_type || filters.status || filters.level);
  }, [filters]);

  return {
    // Data
    organizations: organizations || [],
    totalCount,
    stats: useTemporalData ? null : stats, // 时态模式下不显示统计
    
    // Loading states
    isLoading: isLoading && !isFetching,
    isFetching,
    error,
    
    // Filter states
    filters,
    isFiltered,
    
    // Filter actions
    setFilters,
    updateFilters,
    resetFilters,
    handlePageChange,

    // Temporal context - 使用默认值
    temporalMode: 'current',
    isHistorical: false,
    isPlanning: false,
    temporalContext: null
  };
};