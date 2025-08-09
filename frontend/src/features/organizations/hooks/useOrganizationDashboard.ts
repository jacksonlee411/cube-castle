import { useState, useMemo } from 'react';
import { useOrganizations, useOrganizationStats } from '../../../shared/hooks/useOrganizations';
import type { OrganizationQueryParams } from '../../../shared/api/organizations';
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

  // Convert filters to query parameters
  const queryParams: OrganizationQueryParams = useMemo(() => ({
    searchText: filters.searchText || undefined,
    unit_type: (filters.unit_type as OrganizationUnitType) || undefined,
    status: (filters.status as OrganizationStatus) || undefined,
    level: filters.level || undefined,
    page: filters.page,
    pageSize: filters.pageSize,
  }), [filters]);

  const { 
    data: organizationData, 
    isLoading, 
    error, 
    isFetching 
  } = useOrganizations(queryParams);

  const { data: stats } = useOrganizationStats();

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
    organizations: organizationData?.organizations || [],
    totalCount: organizationData?.total_count || 0,
    stats,
    
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
  };
};