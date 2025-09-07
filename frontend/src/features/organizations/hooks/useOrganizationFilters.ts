/**
 * ⚠️  DEPRECATED - Phase 1 Hook统一化
 * 
 * 此Hook将在Phase 1完成后被移除。
 * 过滤逻辑已整合到useEnterpriseOrganizations中。
 * 
 * 迁移指南:
 * import { useEnterpriseOrganizations } from '@/shared/hooks';
 */

import { useMemo } from 'react';
import type { FilterState } from '../OrganizationFilters';

// 运行时废弃警告
if (typeof console !== 'undefined' && process.env.NODE_ENV === 'development') {
  console.warn('⚠️  useOrganizationFilters is deprecated. Please migrate to useEnterpriseOrganizations');
}

export const useOrganizationFilters = (
  filters: FilterState,
  onFiltersChange: (filters: FilterState) => void
) => {
  const hasActiveFilters = useMemo(() => {
    return !!(filters.searchText || filters.unitType || filters.status || filters.level);
  }, [filters]);

  const clearFilters = () => {
    onFiltersChange({
      searchText: '',
      unitType: undefined,
      status: undefined,
      level: undefined,
      page: 1,
      pageSize: filters.pageSize,
    });
  };

  const updateFilter = (field: keyof FilterState, value: FilterState[keyof FilterState]) => {
    onFiltersChange({
      ...filters,
      [field]: value,
      page: field !== 'page' ? 1 : filters.page, // Reset to page 1 when filtering
    });
  };

  const updateFilters = (newFilters: Partial<FilterState>) => {
    onFiltersChange({
      ...filters,
      ...newFilters,
      page: 1, // Reset to page 1 when applying multiple filters
    });
  };

  return {
    hasActiveFilters,
    clearFilters,
    updateFilter,
    updateFilters,
  };
};