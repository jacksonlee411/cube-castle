import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import { 
  Employee, 
  EmployeeStats, 
  EmployeeFilters
} from '@/types/employee'
import { 
  employeeCommands, 
  employeeQueries, 
  CQRSOperationStatus, 
  OptimisticUpdate,
  CreateEmployeeCommand,
  UpdateEmployeeCommand,
  TerminateEmployeeCommand
} from '@/lib/cqrs'
import { debounce, RequestDeduplicator } from '@/lib/performance-utils'
import { CQRSError, CQRSErrorFactory, ErrorSeverity, ErrorReporter } from '@/lib/cqrs-error-handling'
import { handleEmployeeQueryError, EmployeeErrorContext, createEmployeeRequestId } from '@/lib/cqrs/employee-error-utils'
import toast from 'react-hot-toast'

// Employee event types
export interface EmployeeEvent {
  type: 'EMPLOYEE_HIRED' | 'EMPLOYEE_UPDATED' | 'EMPLOYEE_TERMINATED'
  payload: {
    employee_id: string
    tenant_id: string
    employee_name?: string
    changes?: Record<string, any>
  }
  timestamp: string
  event_id: string
}

// CQRS Employee State Interface
interface CQRSEmployeeState {
  // === 数据状态 ===
  employees: Employee[]
  employeeStats: EmployeeStats | null
  selectedEmployee: Employee | null
  
  // === UI 状态 ===
  filters: EmployeeFilters
  searchQuery: string
  viewMode: 'grid' | 'list' | 'table'
  selectedEmployeeIds: Set<string>
  
  // === 操作状态 ===
  commandStatus: {
    creating: boolean
    updating: boolean
    terminating: boolean
  }
  queryStatus: {
    loading: boolean
    refreshing: boolean
  }
  errors: Record<string, string>
  
  // === 乐观更新 ===
  optimisticUpdates: Map<string, OptimisticUpdate<Employee>>
  
  // === 缓存管理 ===
  lastUpdated: Record<string, Date>
  cacheStatus: {
    employees: 'valid' | 'invalid' | 'loading'
    stats: 'valid' | 'invalid' | 'loading'
  }
  
  // === 查询操作 ===
  fetchEmployees: () => Promise<void>
  fetchEmployeeStats: () => Promise<void>
  searchEmployees: (query: string) => Promise<void>
  refreshAll: () => Promise<void>
  
  // === 命令操作 ===
  createEmployee: (command: Omit<CreateEmployeeCommand, 'tenant_id'>) => Promise<Employee | null>
  updateEmployee: (command: Omit<UpdateEmployeeCommand, 'tenant_id'>) => Promise<Employee | null>
  terminateEmployee: (command: Omit<TerminateEmployeeCommand, 'tenant_id'>) => Promise<boolean>
  
  // === UI 状态管理 ===
  setFilters: (filters: Partial<EmployeeFilters>) => void
  setSearchQuery: (query: string) => void
  setViewMode: (mode: 'grid' | 'list' | 'table') => void
  selectEmployee: (employee: Employee | null) => void
  toggleEmployeeSelection: (employeeId: string) => void
  clearSelections: () => void
  
  // === 缓存管理 ===
  invalidateCache: (key?: keyof CQRSEmployeeState['cacheStatus']) => void
  
  // === 工具方法 ===
  reset: () => void
}

// 初始状态
const initialState = {
  employees: [],
  employeeStats: null,
  selectedEmployee: null,
  filters: {},
  searchQuery: '',
  viewMode: 'table' as const,
  selectedEmployeeIds: new Set<string>(),
  commandStatus: {
    creating: false,
    updating: false,
    terminating: false,
  },
  queryStatus: {
    loading: false,
    refreshing: false,
  },
  errors: {},
  optimisticUpdates: new Map(),
  lastUpdated: {},
  cacheStatus: {
    employees: 'invalid' as const,
    stats: 'invalid' as const,
  },
}

// 全局请求去重器实例
const requestDeduplicator = new RequestDeduplicator()

// 防抖搜索函数 - 需要在store外部定义以避免重复创建
let debouncedSearch: ((query: string, fetchFn: () => Promise<void>) => void) | null = null

const createDebouncedSearch = () => {
  return debounce(async (query: string, fetchFn: () => Promise<void>) => {
    await fetchFn()
  }, 300) // 300ms防抖
}

/**
 * Employee CQRS Store - Zustand store with CQRS pattern
 */
export const useEmployeeStore = create<CQRSEmployeeState>()(
  subscribeWithSelector(
    immer((set, get) => ({
      ...initialState,

      // === 查询操作实现 ===
      
      fetchEmployees: async () => {
        const state = get()
        if (state.queryStatus.loading) return

        // 智能缓存检查：如果数据是最近5分钟内获取的，直接返回
        const lastUpdated = state.lastUpdated.employees
        if (lastUpdated && state.cacheStatus.employees === 'valid') {
          const cacheAge = Date.now() - lastUpdated.getTime()
          const CACHE_TTL = 5 * 60 * 1000 // 5分钟缓存
          if (cacheAge < CACHE_TTL) {
            console.log('📦 Using cached employees data', { cacheAge: `${Math.round(cacheAge/1000)}s` })
            return
          }
        }

        // 请求去重：生成基于搜索参数的唯一key
        const searchParams = {
          ...state.filters,
          name: state.searchQuery || undefined,
        }
        const requestKey = `employees:${JSON.stringify(searchParams)}`

        return requestDeduplicator.dedupe(requestKey, async () => {
          set((draft) => {
            draft.queryStatus.loading = true
            draft.cacheStatus.employees = 'loading'
            draft.errors.fetchEmployees = ''
          })

          try {
            const response = await employeeQueries.searchEmployees(searchParams)
            
            set((draft) => {
              draft.employees = response.employees
              draft.queryStatus.loading = false
              draft.cacheStatus.employees = 'valid'
              draft.lastUpdated.employees = new Date()
              // 清除所有错误状态
              draft.errors = {}
            })

            console.log('✅ Employees fetched successfully', response.employees.length)
          } catch (error) {
            const requestId = createEmployeeRequestId('search')
            const context: EmployeeErrorContext = {
              operation: 'search',
              tenantId: process.env.NEXT_PUBLIC_DEFAULT_TENANT_ID || '00000000-0000-0000-0000-000000000001',
              searchParams,
              requestId,
            }
            
            const { cqrsError, strategy } = handleEmployeeQueryError(error as Error, context)
            
            set((draft) => {
              draft.queryStatus.loading = false
              draft.cacheStatus.employees = 'invalid'
              draft.errors.fetchEmployees = cqrsError.userMessage
            })

            // 如果策略指示不显示toast，则使用fallback数据
            if (strategy.fallbackData && !strategy.shouldShowToast) {
              set((draft) => {
                draft.employees = strategy.fallbackData.employees || []
                draft.cacheStatus.employees = 'valid'
              })
              console.log('📦 Using fallback data for employees', strategy.fallbackData)
            }
            
            console.error('❌ Failed to fetch employees:', cqrsError.toLogFormat())
            throw cqrsError // 重新抛出错误供去重器处理
          }
        })
      },

      fetchEmployeeStats: async () => {
        const state = get()
        if (state.queryStatus.loading) return

        set((draft) => {
          draft.cacheStatus.stats = 'loading'
          draft.errors.fetchEmployeeStats = ''
        })

        try {
          const stats = await employeeQueries.getEmployeeStats()
          
          set((draft) => {
            draft.employeeStats = stats
            draft.cacheStatus.stats = 'valid'
            draft.lastUpdated.stats = new Date()
            // 清除统计数据相关错误
            delete draft.errors.fetchEmployeeStats
          })

          console.log('✅ Employee stats fetched successfully', stats)
        } catch (error) {
          const requestId = createEmployeeRequestId('stats')
          const context: EmployeeErrorContext = {
            operation: 'stats',
            tenantId: process.env.NEXT_PUBLIC_DEFAULT_TENANT_ID || '00000000-0000-0000-0000-000000000001',
            requestId,
          }
          
          const { cqrsError, strategy } = handleEmployeeQueryError(error as Error, context)
          
          set((draft) => {
            draft.cacheStatus.stats = 'invalid'
            draft.errors.fetchEmployeeStats = cqrsError.userMessage
          })

          // 对于统计数据，总是使用fallback数据以保证用户体验
          if (strategy.fallbackData) {
            set((draft) => {
              draft.employeeStats = strategy.fallbackData
              draft.cacheStatus.stats = 'valid'
              draft.lastUpdated.stats = new Date()
            })
            console.log('📦 Using fallback stats data', strategy.fallbackData)
          }

          console.error('❌ Failed to fetch employee stats:', cqrsError.toLogFormat())
        }
      },

      searchEmployees: async (query: string) => {
        // 立即更新搜索查询状态，不等待防抖
        set((draft) => {
          draft.searchQuery = query
        })

        // 初始化防抖函数（仅第一次）
        if (!debouncedSearch) {
          debouncedSearch = createDebouncedSearch()
        }

        // 使用防抖搜索，避免频繁请求
        debouncedSearch(query, async () => {
          await get().fetchEmployees()
        })
      },

      refreshAll: async () => {
        const state = get()
        
        set((draft) => {
          draft.queryStatus.refreshing = true
          draft.cacheStatus.employees = 'invalid'
          draft.cacheStatus.stats = 'invalid'
          // 清除现有错误状态，开始新的请求
          draft.errors = {}
        })

        try {
          await Promise.all([
            state.fetchEmployees(),
            state.fetchEmployeeStats(),
          ])

          // 确保刷新成功后清除所有错误状态
          set((draft) => {
            draft.errors = {}
          })

          toast.success('员工数据已刷新')
        } catch (error) {
          toast.error('刷新员工数据失败')
        } finally {
          set((draft) => {
            draft.queryStatus.refreshing = false
          })
        }
      },

      // === 命令操作实现 ===

      createEmployee: async (command) => {
        set((draft) => {
          draft.commandStatus.creating = true
          draft.errors.createEmployee = ''
        })

        try {
          const response = await employeeCommands.createEmployee(command)
          
          if (response.success && response.data) {
            const newEmployee = response.data

            set((draft) => {
              draft.employees.unshift(newEmployee)
              draft.commandStatus.creating = false
              draft.lastUpdated.employees = new Date()
            })

            toast.success(`员工 ${newEmployee.legalName} 创建成功`)
            console.log('✅ Employee created successfully', newEmployee)
            
            return newEmployee
          } else {
            throw new Error(response.error || 'Failed to create employee')
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to create employee'
          
          set((draft) => {
            draft.commandStatus.creating = false
            draft.errors.createEmployee = errorMessage
          })

          toast.error('创建员工失败: ' + errorMessage)
          console.error('❌ Failed to create employee:', error)
          return null
        }
      },

      updateEmployee: async (command) => {
        set((draft) => {
          draft.commandStatus.updating = true
          draft.errors.updateEmployee = ''
        })

        try {
          const response = await employeeCommands.updateEmployee(command)
          
          if (response.success && response.data) {
            const updatedEmployee = response.data

            set((draft) => {
              const index = draft.employees.findIndex(emp => emp.id === command.id)
              if (index >= 0) {
                draft.employees[index] = updatedEmployee
              }
              draft.commandStatus.updating = false
              draft.lastUpdated.employees = new Date()
            })

            toast.success(`员工 ${updatedEmployee.legalName} 更新成功`)
            console.log('✅ Employee updated successfully', updatedEmployee)
            
            return updatedEmployee
          } else {
            throw new Error(response.error || 'Failed to update employee')
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to update employee'
          
          set((draft) => {
            draft.commandStatus.updating = false
            draft.errors.updateEmployee = errorMessage
          })

          toast.error('更新员工失败: ' + errorMessage)
          console.error('❌ Failed to update employee:', error)
          return null
        }
      },

      terminateEmployee: async (command) => {
        set((draft) => {
          draft.commandStatus.terminating = true
          draft.errors.terminateEmployee = ''
        })

        try {
          const response = await employeeCommands.terminateEmployee(command)
          
          if (response.success) {
            set((draft) => {
              const index = draft.employees.findIndex(emp => emp.id === command.id)
              if (index >= 0) {
                draft.employees[index].status = 'inactive'
              }
              draft.commandStatus.terminating = false
              draft.lastUpdated.employees = new Date()
            })

            toast.success('员工离职成功')
            console.log('✅ Employee terminated successfully', command.id)
            
            return true
          } else {
            throw new Error(response.error || 'Failed to terminate employee')
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to terminate employee'
          
          set((draft) => {
            draft.commandStatus.terminating = false
            draft.errors.terminateEmployee = errorMessage
          })

          toast.error('员工离职失败: ' + errorMessage)
          console.error('❌ Failed to terminate employee:', error)
          return false
        }
      },

      // === UI 状态管理 ===

      setFilters: (filters) => {
        set((draft) => {
          draft.filters = { ...draft.filters, ...filters }
        })
        
        // Auto-refresh when filters change
        setTimeout(() => get().fetchEmployees(), 100)
      },

      setSearchQuery: (query) => {
        set((draft) => {
          draft.searchQuery = query
        })
      },

      setViewMode: (mode) => {
        set((draft) => {
          draft.viewMode = mode
        })
      },

      selectEmployee: (employee) => {
        set((draft) => {
          draft.selectedEmployee = employee
        })
      },

      toggleEmployeeSelection: (employeeId) => {
        set((draft) => {
          if (draft.selectedEmployeeIds.has(employeeId)) {
            draft.selectedEmployeeIds.delete(employeeId)
          } else {
            draft.selectedEmployeeIds.add(employeeId)
          }
        })
      },

      clearSelections: () => {
        set((draft) => {
          draft.selectedEmployeeIds.clear()
          draft.selectedEmployee = null
        })
      },

      // === 缓存管理 ===

      invalidateCache: (key) => {
        set((draft) => {
          if (key) {
            draft.cacheStatus[key] = 'invalid'
          } else {
            draft.cacheStatus.employees = 'invalid'
            draft.cacheStatus.stats = 'invalid'
          }
        })
      },

      // === 工具方法 ===

      reset: () => {
        set(() => ({
          ...initialState,
          selectedEmployeeIds: new Set(),
          optimisticUpdates: new Map(),
        }))
      },
    }))
  )
)

// 选择器函数
export const employeeSelectors = {
  employees: (state: CQRSEmployeeState) => state.employees,
  employeeStats: (state: CQRSEmployeeState) => state.employeeStats,
  selectedEmployee: (state: CQRSEmployeeState) => state.selectedEmployee,
  isLoading: (state: CQRSEmployeeState) => state.queryStatus.loading,
  isRefreshing: (state: CQRSEmployeeState) => state.queryStatus.refreshing,
  hasErrors: (state: CQRSEmployeeState) => {
    // 检查是否有实际的错误（排除空字符串）
    const actualErrors = Object.values(state.errors).filter(error => error && error.trim() !== '')
    return actualErrors.length > 0
  },
  filteredEmployees: (state: CQRSEmployeeState) => {
    let result = state.employees

    // Apply filters
    if (state.filters.status) {
      result = result.filter(emp => emp.status === state.filters.status)
    }
    if (state.filters.department) {
      result = result.filter(emp => emp.department === state.filters.department)
    }
    if (state.filters.position) {
      result = result.filter(emp => emp.position === state.filters.position)
    }

    // Apply search query
    if (state.searchQuery) {
      const query = state.searchQuery.toLowerCase()
      result = result.filter(emp =>
        emp.legalName.toLowerCase().includes(query) ||
        emp.email.toLowerCase().includes(query) ||
        (emp.department && emp.department.toLowerCase().includes(query)) ||
        (emp.position && emp.position.toLowerCase().includes(query))
      )
    }

    return result
  },
}