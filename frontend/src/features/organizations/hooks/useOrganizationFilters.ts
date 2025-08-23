import { useMemo } from 'react';
import type { FilterState } from '../OrganizationFilters';

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