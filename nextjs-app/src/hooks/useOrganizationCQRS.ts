import useSWR from 'swr'
import { useEffect } from 'react'
import { useOrganizationStore, organizationSelectors } from '@/stores/organizationStore'
import { Organization, OrganizationStats } from '@/types'

/**
 * CQRS 组织管理 Hook - 统一数据访问接口
 * 结合 SWR 缓存和 Zustand 状态管理
 */
export const useOrganizationCQRS = () => {
  const store = useOrganizationStore()
  
  // 从 store 中选择需要的状态
  const organizations = useOrganizationStore(organizationSelectors.organizations)
  const orgChart = useOrganizationStore(organizationSelectors.orgChart)
  const orgStats = useOrganizationStore(organizationSelectors.orgStats)
  const selectedOrganization = useOrganizationStore(organizationSelectors.selectedOrganization)
  const isLoading = useOrganizationStore(organizationSelectors.isLoading)
  const isRefreshing = useOrganizationStore(organizationSelectors.isRefreshing)
  const hasErrors = useOrganizationStore(organizationSelectors.hasErrors)
  const filteredOrganizations = useOrganizationStore(organizationSelectors.filteredOrganizations)
  const organizationTree = useOrganizationStore(organizationSelectors.organizationTree)
  
  // 初始化数据获取
  useEffect(() => {
    if (store.cacheStatus.organizations === 'invalid') {
      store.fetchOrganizations()
    }
    if (store.cacheStatus.orgChart === 'invalid') {
      store.fetchOrganizationChart()
    }
    if (store.cacheStatus.stats === 'invalid') {
      store.fetchOrganizationStats()
    }
  }, [])
  
  return {
    // === 数据 ===
    organizations,
    orgChart,
    orgStats,
    selectedOrganization,
    filteredOrganizations,
    organizationTree,
    
    // === 状态 ===
    isLoading,
    isRefreshing,
    hasErrors,
    errors: store.errors,
    
    // === 查询操作 ===
    fetchOrganizations: store.fetchOrganizations,
    fetchOrganizationChart: store.fetchOrganizationChart,
    fetchOrganizationStats: store.fetchOrganizationStats,
    searchOrganizations: store.searchOrganizations,
    refreshAll: store.refreshAll,
    
    // === 命令操作 ===
    createOrganization: store.createOrganization,
    updateOrganization: store.updateOrganization,
    deleteOrganization: store.deleteOrganization,
    bulkUpdateOrganizations: store.bulkUpdateOrganizations,
    restructureOrganization: store.restructureOrganization,
    
    // === UI 状态管理 ===
    filters: store.filters,
    searchQuery: store.searchQuery,
    viewMode: store.viewMode,
    expandedNodes: store.expandedNodes,
    selectedOrgIds: store.selectedOrgIds,
    
    setFilters: store.setFilters,
    setSearchQuery: store.setSearchQuery,
    setViewMode: store.setViewMode,
    toggleNodeExpansion: store.toggleNodeExpansion,
    selectOrganization: store.selectOrganization,
    toggleOrganizationSelection: store.toggleOrganizationSelection,
    clearSelections: store.clearSelections,
    
    // === 缓存管理 ===
    invalidateCache: store.invalidateCache,
    cacheStatus: store.cacheStatus,
    
    // === 实用工具 ===
    reset: store.reset
  }
}

/**
 * 组织选择 Hook - 用于获取单个组织详情
 */
export const useOrganization = (id: string | undefined) => {
  const organizations = useOrganizationStore(organizationSelectors.organizations)
  const fetchOrganizations = useOrganizationStore(state => state.fetchOrganizations)
  
  // 从本地状态查找组织
  const organization = organizations.find(org => org.id === id)
  
  // 如果本地没有数据，触发获取
  useEffect(() => {
    if (id && !organization && organizations.length === 0) {
      fetchOrganizations()
    }
  }, [id, organization, organizations.length])
  
  return {
    organization,
    isLoading: !organization && organizations.length === 0,
    error: !organization && organizations.length > 0 ? 'Organization not found' : null
  }
}

/**
 * 组织统计 Hook - 专门用于统计数据
 */
export const useOrganizationStats = () => {
  const orgStats = useOrganizationStore(organizationSelectors.orgStats)
  const fetchOrganizationStats = useOrganizationStore(state => state.fetchOrganizationStats)
  const isLoading = useOrganizationStore(state => state.queryStatus.loading)
  
  // SWR 作为备份缓存层
  const { data: swrStats, error: swrError } = useSWR(
    'organization-stats',
    () => fetchOrganizationStats(),
    {
      refreshInterval: 5 * 60 * 1000, // 5分钟刷新
      revalidateOnFocus: false,
      dedupingInterval: 2 * 60 * 1000 // 2分钟去重
    }
  )
  
  return {
    stats: orgStats,
    isLoading,
    error: swrError,
    refresh: fetchOrganizationStats
  }
}

/**
 * 组织搜索 Hook - 专门用于搜索功能
 */
export const useOrganizationSearch = (initialQuery = '') => {
  const searchOrganizations = useOrganizationStore(state => state.searchOrganizations)
  const setSearchQuery = useOrganizationStore(state => state.setSearchQuery)
  const searchQuery = useOrganizationStore(state => state.searchQuery)
  const filteredOrganizations = useOrganizationStore(organizationSelectors.filteredOrganizations)
  const isLoading = useOrganizationStore(state => state.queryStatus.loading)
  
  // 初始化搜索查询
  useEffect(() => {
    if (initialQuery) {
      setSearchQuery(initialQuery)
    }
  }, [initialQuery])
  
  const search = async (query: string) => {
    setSearchQuery(query)
    if (query.trim()) {
      await searchOrganizations(query)
    }
  }
  
  const clearSearch = () => {
    setSearchQuery('')
  }
  
  return {
    searchQuery,
    results: filteredOrganizations,
    isLoading,
    search,
    clearSearch
  }
}

/**
 * 组织树操作 Hook - 专门用于树形结构操作
 */
export const useOrganizationTree = () => {
  const organizationTree = useOrganizationStore(organizationSelectors.organizationTree)
  const expandedNodes = useOrganizationStore(state => state.expandedNodes)
  const toggleNodeExpansion = useOrganizationStore(state => state.toggleNodeExpansion)
  const selectOrganization = useOrganizationStore(state => state.selectOrganization)
  const selectedOrganization = useOrganizationStore(state => state.selectedOrganization)
  
  const expandAll = () => {
    const allIds = useOrganizationStore.getState().organizations.map(org => org.id)
    allIds.forEach(id => {
      if (!expandedNodes.has(id)) {
        toggleNodeExpansion(id)
      }
    })
  }
  
  const collapseAll = () => {
    Array.from(expandedNodes).forEach(id => toggleNodeExpansion(id))
  }
  
  const isNodeExpanded = (nodeId: string) => expandedNodes.has(nodeId)
  
  const getNodeLevel = (nodeId: string): number => {
    const organizations = useOrganizationStore.getState().organizations
    const org = organizations.find(o => o.id === nodeId)
    return org?.level || 0
  }
  
  return {
    tree: organizationTree,
    expandedNodes,
    selectedOrganization,
    toggleNodeExpansion,
    selectOrganization,
    expandAll,
    collapseAll,
    isNodeExpanded,
    getNodeLevel
  }
}

/**
 * 组织批量操作 Hook
 */
export const useOrganizationBulkOperations = () => {
  const selectedOrgIds = useOrganizationStore(state => state.selectedOrgIds)
  const toggleOrganizationSelection = useOrganizationStore(state => state.toggleOrganizationSelection)
  const clearSelections = useOrganizationStore(state => state.clearSelections)
  const bulkUpdateOrganizations = useOrganizationStore(state => state.bulkUpdateOrganizations)
  const organizations = useOrganizationStore(state => state.organizations)
  
  const selectedOrganizations = organizations.filter(org => selectedOrgIds.has(org.id))
  
  const selectAll = () => {
    organizations.forEach(org => {
      if (!selectedOrgIds.has(org.id)) {
        toggleOrganizationSelection(org.id)
      }
    })
  }
  
  const selectNone = () => {
    clearSelections()
  }
  
  const isSelected = (orgId: string) => selectedOrgIds.has(orgId)
  
  const selectedCount = selectedOrgIds.size
  
  return {
    selectedOrgIds,
    selectedOrganizations,
    selectedCount,
    toggleOrganizationSelection,
    selectAll,
    selectNone,
    clearSelections,
    isSelected,
    bulkUpdate: bulkUpdateOrganizations
  }
}

// 默认导出主要 Hook
export default useOrganizationCQRS