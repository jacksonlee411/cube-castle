/**
 * 统一Hook导出 - P2级Hook合并优化完成 ⭐ (2025-09-09)
 * 
 * 🏆 Hook重复代码彻底消除成果：
 * - ✅ 主要实现：useEnterpriseOrganizations (唯一组织查询Hook)
 * - ❌ 消除重复：useOrganizationList (不必要的包装器)
 * - ❌ 消除重复：useOrganization (功能重叠)
 * - 🎯 重复消除率：83% (6个Hook → 1个核心Hook)
 */

// 🎯 唯一组织Hook实现
export {
  useEnterpriseOrganizations,
  type OrganizationStats,
  type OrganizationTemporalSummary,
  type OrganizationsQueryResult,
  type NormalizedQueryParams,
  type UseEnterpriseOrganizationsResult,
  ORGANIZATIONS_QUERY_ROOT_KEY,
  organizationsQueryKey,
  organizationByCodeQueryKey,
} from './useEnterpriseOrganizations';
export { default as useEnterpriseOrganizationsDefault } from './useEnterpriseOrganizations';

// 🎯 职位管理查询 Hook
export {
  useEnterprisePositions,
  usePositionDetail,
  useVacantPositions,
  usePositionHeadcountStats,
  type PositionQueryParams,
  type VacantPositionsQueryParams,
  type PositionHeadcountStatsParams,
  type VacantPositionSortField,
  type PositionDetailOptions,
  POSITIONS_QUERY_ROOT_KEY,
  POSITION_DETAIL_QUERY_ROOT_KEY,
  VACANT_POSITIONS_QUERY_ROOT_KEY,
  POSITION_HEADCOUNT_STATS_QUERY_ROOT_KEY,
  positionsQueryKey,
  positionDetailQueryKey,
  vacantPositionsQueryKey,
  positionHeadcountStatsQueryKey,
} from './useEnterprisePositions';
export { default as useEnterprisePositionsDefault } from './useEnterprisePositions';

// 🔧 专用工具Hook
export * from './useOrganizationMutations';
export * from './useDebounce';

// ⚠️ DEPRECATED: 消除重复Hook别名
// useOrganizationList 是不必要的重复，直接使用 useEnterpriseOrganizations

/**
 * 🚀 统一Hook使用指南:
 * 
 * 主要使用：
 * import { useEnterpriseOrganizations } from '@/shared/hooks';
 * const { organizations, loading, fetchOrganizations } = useEnterpriseOrganizations();
 * 
 * 简化使用：
 * import { useOrganizationList } from '@/shared/hooks';
 * const { organizations, loading } = useOrganizationList();
 * 
 * 特定功能：
 * - useOrganizationMutations: 创建/更新/删除操作
 * 
 * ❌ 已删除的Hook：
 * - useOrganizationActions (功能已整合)
 * - useOrganizationDashboard (功能已整合)  
 * - useOrganizationFilters (功能已整合)
 */
