import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import { 
  Organization, 
  OrganizationStats, 
  CreateOrganizationRequest, 
  UpdateOrganizationRequest 
} from '@/types'
import { 
  organizationCommands, 
  organizationQueries, 
  CQRSOperationStatus, 
  OptimisticUpdate,
  OrganizationEvent
} from '@/lib/cqrs'
import toast from 'react-hot-toast'

// 组织过滤器类型
export interface OrganizationFilters {
  search?: string
  unit_type?: string
  status?: string
  parent_unit_id?: string
  level?: number
}

// CQRS 状态接口
interface CQRSOrganizationState {
  // === 数据状态 ===
  organizations: Organization[]
  orgChart: Organization[]
  orgStats: OrganizationStats | null
  selectedOrganization: Organization | null
  
  // === UI 状态 ===
  expandedNodes: Set<string>
  filters: OrganizationFilters
  searchQuery: string
  viewMode: 'tree' | 'grid' | 'list'
  selectedOrgIds: Set<string>
  
  // === 操作状态 ===
  commandStatus: {
    creating: boolean
    updating: boolean
    deleting: boolean
    restructuring: boolean
  }
  queryStatus: {
    loading: boolean
    refreshing: boolean
  }
  errors: Record<string, string>
  
  // === 乐观更新 ===
  optimisticUpdates: Map<string, OptimisticUpdate<Organization>>
  
  // === 缓存管理 ===
  lastUpdated: Record<string, Date>
  cacheStatus: {
    organizations: 'fresh' | 'stale' | 'invalid'
    orgChart: 'fresh' | 'stale' | 'invalid'
    stats: 'fresh' | 'stale' | 'invalid'
  }
  
  // === Actions ===
  // 查询操作
  fetchOrganizations: (refresh?: boolean) => Promise<void>
  fetchOrganizationChart: (params?: { root_unit_id?: string, max_depth?: number }) => Promise<void>
  fetchOrganizationStats: () => Promise<void>
  searchOrganizations: (query: string) => Promise<void>
  
  // 命令操作
  createOrganization: (data: CreateOrganizationRequest) => Promise<Organization | null>
  updateOrganization: (id: string, data: UpdateOrganizationRequest) => Promise<Organization | null>
  deleteOrganization: (id: string) => Promise<boolean>
  bulkUpdateOrganizations: (updates: Array<{ id: string, data: UpdateOrganizationRequest }>) => Promise<boolean>
  restructureOrganization: (moves: Array<{ unit_id: string, new_parent_id: string | null }>) => Promise<boolean>
  
  // UI 状态管理
  setFilters: (filters: OrganizationFilters) => void
  setSearchQuery: (query: string) => void
  toggleNodeExpansion: (nodeId: string) => void
  setViewMode: (mode: 'tree' | 'grid' | 'list') => void
  selectOrganization: (org: Organization | null) => void
  toggleOrganizationSelection: (orgId: string) => void
  clearSelections: () => void
  
  // 乐观更新管理
  addOptimisticUpdate: (update: OptimisticUpdate<Organization>) => void
  removeOptimisticUpdate: (id: string) => void
  revertOptimisticUpdate: (id: string) => void
  
  // 缓存管理
  invalidateCache: (cacheKeys?: string[]) => void
  refreshAll: () => Promise<void>
  
  // 实时事件处理
  handleOrganizationEvent: (event: OrganizationEvent) => void
  
  // 重置状态
  reset: () => void
}

// 初始状态
const initialState = {
  organizations: [],
  orgChart: [],
  orgStats: null,
  selectedOrganization: null,
  expandedNodes: new Set<string>(),
  filters: {},
  searchQuery: '',
  viewMode: 'tree' as const,
  selectedOrgIds: new Set<string>(),
  commandStatus: {
    creating: false,
    updating: false,
    deleting: false,
    restructuring: false
  },
  queryStatus: {
    loading: false,
    refreshing: false
  },
  errors: {},
  optimisticUpdates: new Map<string, OptimisticUpdate<Organization>>(),
  lastUpdated: {},
  cacheStatus: {
    organizations: 'invalid' as const,
    orgChart: 'invalid' as const,
    stats: 'invalid' as const
  }
}

// 创建 Zustand 存储
export const useOrganizationStore = create<CQRSOrganizationState>()(
  subscribeWithSelector(
    immer((set, get) => ({
      ...initialState,
      
      // === 查询操作 ===
      fetchOrganizations: async (refresh = false) => {
        const state = get()
        
        // 检查缓存
        if (!refresh && state.cacheStatus.organizations === 'fresh') {
          return
        }
        
        set((state) => {
          state.queryStatus.loading = true
          state.errors.fetchOrganizations = ''
        })
        
        try {
          const result = await organizationQueries.listOrganizationUnits({
            limit: 1000 // 获取所有组织
          })
          
          set((state) => {
            state.organizations = result.units
            state.lastUpdated.organizations = new Date()
            state.cacheStatus.organizations = 'fresh'
            state.queryStatus.loading = false
          })
          
          console.log('✅ 组织列表获取成功:', result.units.length)
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : '获取组织列表失败'
          
          set((state) => {
            state.errors.fetchOrganizations = errorMessage
            state.queryStatus.loading = false
            state.cacheStatus.organizations = 'invalid'
          })
          
          console.error('❌ 组织列表获取失败:', error)
          toast.error(errorMessage)
        }
      },
      
      fetchOrganizationChart: async (params = {}) => {
        set((state) => {
          state.queryStatus.loading = true
          state.errors.fetchChart = ''
        })
        
        try {
          const result = await organizationQueries.getOrganizationChart(params)
          
          set((state) => {
            state.orgChart = result.chart
            state.lastUpdated.orgChart = new Date()
            state.cacheStatus.orgChart = 'fresh'
            state.queryStatus.loading = false
          })
          
          console.log('✅ 组织架构图获取成功:', result.chart.length)
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : '获取组织架构图失败'
          
          set((state) => {
            state.errors.fetchChart = errorMessage
            state.queryStatus.loading = false
            state.cacheStatus.orgChart = 'invalid'
          })
          
          console.error('❌ 组织架构图获取失败:', error)
          toast.error(errorMessage)
        }
      },
      
      fetchOrganizationStats: async () => {
        set((state) => {
          state.queryStatus.loading = true
          state.errors.fetchStats = ''
        })
        
        try {
          const result = await organizationQueries.getOrganizationAnalytics()
          
          set((state) => {
            state.orgStats = result.summary
            state.lastUpdated.stats = new Date()
            state.cacheStatus.stats = 'fresh'
            state.queryStatus.loading = false
          })
          
          console.log('✅ 组织统计获取成功:', result.summary)
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : '获取组织统计失败'
          
          set((state) => {
            state.errors.fetchStats = errorMessage
            state.queryStatus.loading = false
            state.cacheStatus.stats = 'invalid'
          })
          
          console.error('❌ 组织统计获取失败:', error)
        }
      },
      
      searchOrganizations: async (query: string) => {
        if (!query.trim()) {
          set((state) => {
            state.searchQuery = ''
          })
          await get().fetchOrganizations()
          return
        }
        
        set((state) => {
          state.searchQuery = query
          state.queryStatus.loading = true
        })
        
        try {
          const result = await organizationQueries.searchOrganizationUnits({
            query,
            limit: 100
          })
          
          set((state) => {
            state.organizations = result.results
            state.queryStatus.loading = false
          })
          
          console.log('✅ 组织搜索成功:', result.results.length)
        } catch (error) {
          console.error('❌ 组织搜索失败:', error)
          toast.error('搜索失败，请重试')
          
          set((state) => {
            state.queryStatus.loading = false
          })
        }
      },
      
      // === 命令操作 ===
      createOrganization: async (data: CreateOrganizationRequest): Promise<Organization | null> => {
        const tempId = `temp-${Date.now()}`
        
        // 乐观更新
        const optimisticOrg: Organization = {
          id: tempId,
          tenant_id: localStorage.getItem('tenant_id') || '',
          unit_type: data.unit_type,
          name: data.name,
          description: data.description,
          parent_unit_id: data.parent_unit_id,
          status: data.status || 'ACTIVE',
          profile: data.profile || {},
          level: 0, // 临时值
          employee_count: 0,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        }
        
        set((state) => {
          state.commandStatus.creating = true
          state.errors.createOrganization = ''
          // 添加乐观更新
          state.optimisticUpdates.set(tempId, {
            id: tempId,
            operation: 'create',
            data: optimisticOrg,
            timestamp: new Date()
          })
          state.organizations.push(optimisticOrg)
        })
        
        try {
          const result = await organizationCommands.createOrganizationUnit(data)
          
          // 移除乐观更新，添加真实数据
          set((state) => {
            state.optimisticUpdates.delete(tempId)
            // 移除临时组织
            state.organizations = state.organizations.filter(org => org.id !== tempId)
            state.commandStatus.creating = false
            // 标记缓存需要刷新
            state.cacheStatus.organizations = 'stale'
            state.cacheStatus.orgChart = 'stale'
          })
          
          // 刷新数据
          await get().fetchOrganizations(true)
          await get().fetchOrganizationChart()
          
          console.log('✅ 组织创建成功:', result)
          toast.success(`组织 "${data.name}" 创建成功`)
          
          // 返回创建的组织（从刷新的数据中查找）
          const createdOrg = get().organizations.find(org => org.name === data.name)
          return createdOrg || null
          
        } catch (error) {
          // 回滚乐观更新
          set((state) => {
            state.optimisticUpdates.delete(tempId)
            state.organizations = state.organizations.filter(org => org.id !== tempId)
            state.commandStatus.creating = false
            state.errors.createOrganization = error instanceof Error ? error.message : '创建失败'
          })
          
          console.error('❌ 组织创建失败:', error)
          toast.error('组织创建失败，请重试')
          return null
        }
      },
      
      updateOrganization: async (id: string, data: UpdateOrganizationRequest): Promise<Organization | null> => {
        const existingOrg = get().organizations.find(org => org.id === id)
        if (!existingOrg) {
          toast.error('组织不存在')
          return null
        }
        
        // 乐观更新
        const optimisticOrg: Organization = {
          ...existingOrg,
          ...data,
          updatedAt: new Date().toISOString()
        }
        
        set((state) => {
          state.commandStatus.updating = true
          state.errors.updateOrganization = ''
          // 添加乐观更新
          state.optimisticUpdates.set(id, {
            id,
            operation: 'update',
            data: optimisticOrg,
            timestamp: new Date()
          })
          // 更新组织列表
          const index = state.organizations.findIndex(org => org.id === id)
          if (index !== -1) {
            state.organizations[index] = optimisticOrg
          }
        })
        
        try {
          await organizationCommands.updateOrganizationUnit(id, data)
          
          set((state) => {
            state.optimisticUpdates.delete(id)
            state.commandStatus.updating = false
            state.cacheStatus.organizations = 'stale'
            state.cacheStatus.orgChart = 'stale'
          })
          
          // 刷新数据
          await get().fetchOrganizations(true)
          await get().fetchOrganizationChart()
          
          console.log('✅ 组织更新成功')
          toast.success(`组织 "${data.name || existingOrg.name}" 更新成功`)
          
          return get().organizations.find(org => org.id === id) || null
          
        } catch (error) {
          // 回滚乐观更新
          set((state) => {
            state.optimisticUpdates.delete(id)
            state.commandStatus.updating = false
            state.errors.updateOrganization = error instanceof Error ? error.message : '更新失败'
            // 恢复原始数据
            const index = state.organizations.findIndex(org => org.id === id)
            if (index !== -1) {
              state.organizations[index] = existingOrg
            }
          })
          
          console.error('❌ 组织更新失败:', error)
          toast.error('组织更新失败，请重试')
          return null
        }
      },
      
      deleteOrganization: async (id: string): Promise<boolean> => {
        const existingOrg = get().organizations.find(org => org.id === id)
        if (!existingOrg) {
          toast.error('组织不存在')
          return false
        }
        
        set((state) => {
          state.commandStatus.deleting = true
          state.errors.deleteOrganization = ''
          // 乐观删除
          state.optimisticUpdates.set(id, {
            id,
            operation: 'delete',
            data: existingOrg,
            timestamp: new Date()
          })
          state.organizations = state.organizations.filter(org => org.id !== id)
        })
        
        try {
          await organizationCommands.deleteOrganizationUnit(id)
          
          set((state) => {
            state.optimisticUpdates.delete(id)
            state.commandStatus.deleting = false
            state.cacheStatus.organizations = 'stale'
            state.cacheStatus.orgChart = 'stale'
          })
          
          // 刷新数据
          await get().fetchOrganizationChart()
          
          console.log('✅ 组织删除成功')
          toast.success(`组织 "${existingOrg.name}" 删除成功`)
          return true
          
        } catch (error) {
          // 回滚删除
          set((state) => {
            state.optimisticUpdates.delete(id)
            state.commandStatus.deleting = false
            state.errors.deleteOrganization = error instanceof Error ? error.message : '删除失败'
            state.organizations.push(existingOrg)
          })
          
          console.error('❌ 组织删除失败:', error)
          toast.error('组织删除失败，请重试')
          return false
        }
      },
      
      // === UI 状态管理 ===
      setFilters: (filters: OrganizationFilters) => {
        set((state) => {
          state.filters = { ...state.filters, ...filters }
        })
      },
      
      setSearchQuery: (query: string) => {
        set((state) => {
          state.searchQuery = query
        })
        
        // 防抖搜索
        if (get().searchQuery !== query) {
          setTimeout(() => {
            if (get().searchQuery === query) {
              get().searchOrganizations(query)
            }
          }, 300)
        }
      },
      
      toggleNodeExpansion: (nodeId: string) => {
        set((state) => {
          if (state.expandedNodes.has(nodeId)) {
            state.expandedNodes.delete(nodeId)
          } else {
            state.expandedNodes.add(nodeId)
          }
        })
      },
      
      setViewMode: (mode: 'tree' | 'grid' | 'list') => {
        set((state) => {
          state.viewMode = mode
        })
      },
      
      selectOrganization: (org: Organization | null) => {
        set((state) => {
          state.selectedOrganization = org
        })
      },
      
      toggleOrganizationSelection: (orgId: string) => {
        set((state) => {
          if (state.selectedOrgIds.has(orgId)) {
            state.selectedOrgIds.delete(orgId)
          } else {
            state.selectedOrgIds.add(orgId)
          }
        })
      },
      
      clearSelections: () => {
        set((state) => {
          state.selectedOrgIds.clear()
          state.selectedOrganization = null
        })
      },
      
      // === 乐观更新管理 ===
      addOptimisticUpdate: (update: OptimisticUpdate<Organization>) => {
        set((state) => {
          state.optimisticUpdates.set(update.id, update)
        })
      },
      
      removeOptimisticUpdate: (id: string) => {
        set((state) => {
          state.optimisticUpdates.delete(id)
        })
      },
      
      revertOptimisticUpdate: (id: string) => {
        const update = get().optimisticUpdates.get(id)
        if (update) {
          set((state) => {
            state.optimisticUpdates.delete(id)
            // 根据操作类型回滚
            switch (update.operation) {
              case 'create':
                state.organizations = state.organizations.filter(org => org.id !== id)
                break
              case 'update':
                const index = state.organizations.findIndex(org => org.id === id)
                if (index !== -1) {
                  // 这里需要原始数据来回滚，暂时简化处理
                  state.cacheStatus.organizations = 'stale'
                }
                break
              case 'delete':
                state.organizations.push(update.data)
                break
            }
          })
        }
      },
      
      // === 缓存管理 ===
      invalidateCache: (cacheKeys?: string[]) => {
        set((state) => {
          if (!cacheKeys) {
            state.cacheStatus.organizations = 'invalid'
            state.cacheStatus.orgChart = 'invalid'
            state.cacheStatus.stats = 'invalid'
          } else {
            cacheKeys.forEach(key => {
              if (key in state.cacheStatus) {
                ;(state.cacheStatus as any)[key] = 'invalid'
              }
            })
          }
        })
      },
      
      refreshAll: async () => {
        set((state) => {
          state.queryStatus.refreshing = true
        })
        
        try {
          await Promise.all([
            get().fetchOrganizations(true),
            get().fetchOrganizationChart(),
            get().fetchOrganizationStats()
          ])
        } finally {
          set((state) => {
            state.queryStatus.refreshing = false
          })
        }
      },
      
      // === 实时事件处理 ===
      handleOrganizationEvent: (event: OrganizationEvent) => {
        console.log('📡 收到组织事件:', event)
        
        switch (event.type) {
          case 'ORGANIZATION_CREATED':
          case 'ORGANIZATION_UPDATED':
          case 'ORGANIZATION_DELETED':
            // 标记缓存过期，触发重新获取
            get().invalidateCache()
            get().fetchOrganizations(true)
            get().fetchOrganizationChart()
            break
        }
      },
      
      // === 重置状态 ===
      reset: () => {
        set(initialState)
      },
      
      // 占位实现 - 后续完善
      bulkUpdateOrganizations: async () => false,
      restructureOrganization: async () => false
    }))
  )
)

// 导出状态选择器
export const organizationSelectors = {
  // 基础数据选择器
  organizations: (state: CQRSOrganizationState) => state.organizations,
  orgChart: (state: CQRSOrganizationState) => state.orgChart,
  orgStats: (state: CQRSOrganizationState) => state.orgStats,
  selectedOrganization: (state: CQRSOrganizationState) => state.selectedOrganization,
  
  // UI 状态选择器
  expandedNodes: (state: CQRSOrganizationState) => state.expandedNodes,
  filters: (state: CQRSOrganizationState) => state.filters,
  searchQuery: (state: CQRSOrganizationState) => state.searchQuery,
  viewMode: (state: CQRSOrganizationState) => state.viewMode,
  selectedOrgIds: (state: CQRSOrganizationState) => state.selectedOrgIds,
  
  // 状态指示器
  isLoading: (state: CQRSOrganizationState) => 
    state.queryStatus.loading || Object.values(state.commandStatus).some(Boolean),
  isRefreshing: (state: CQRSOrganizationState) => state.queryStatus.refreshing,
  hasErrors: (state: CQRSOrganizationState) => Object.values(state.errors).some(Boolean),
  
  // 派生数据选择器
  filteredOrganizations: (state: CQRSOrganizationState) => {
    let filtered = state.organizations
    
    if (state.searchQuery) {
      const query = state.searchQuery.toLowerCase()
      filtered = filtered.filter(org => 
        org.name.toLowerCase().includes(query) ||
        org.description?.toLowerCase().includes(query)
      )
    }
    
    if (state.filters.unit_type) {
      filtered = filtered.filter(org => org.unit_type === state.filters.unit_type)
    }
    
    if (state.filters.status) {
      filtered = filtered.filter(org => org.status === state.filters.status)
    }
    
    if (state.filters.parent_unit_id) {
      filtered = filtered.filter(org => org.parent_unit_id === state.filters.parent_unit_id)
    }
    
    return filtered
  },
  
  rootOrganizations: (state: CQRSOrganizationState) =>
    state.organizations.filter(org => !org.parent_unit_id),
  
  organizationTree: (state: CQRSOrganizationState) => {
    // 构建树形结构的逻辑
    const buildTree = (orgs: Organization[], parentId?: string): Organization[] => {
      return orgs
        .filter(org => org.parent_unit_id === parentId)
        .map(org => ({
          ...org,
          children: buildTree(orgs, org.id)
        }))
    }
    
    return buildTree(state.organizations)
  }
}

// 默认导出
export default useOrganizationStore