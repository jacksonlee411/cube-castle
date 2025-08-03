import { useEffect } from 'react'
import { useEmployeeStore, employeeSelectors } from '@/stores/employeeStore'
import { Employee, EmployeeStats, EmployeeFilters } from '@/types/employee'

/**
 * CQRS 员工管理 Hook - 统一数据访问接口
 * 基于 Zustand 状态管理和 CQRS 架构
 */
export const useEmployeeCQRS = () => {
  const store = useEmployeeStore()
  
  // 从 store 中选择需要的状态
  const employees = useEmployeeStore(employeeSelectors.employees)
  const employeeStats = useEmployeeStore(employeeSelectors.employeeStats)
  const selectedEmployee = useEmployeeStore(employeeSelectors.selectedEmployee)
  const isLoading = useEmployeeStore(employeeSelectors.isLoading)
  const isRefreshing = useEmployeeStore(employeeSelectors.isRefreshing)
  const hasErrors = useEmployeeStore(employeeSelectors.hasErrors)
  const filteredEmployees = useEmployeeStore(employeeSelectors.filteredEmployees)
  
  // 初始化数据获取
  useEffect(() => {
    if (store.cacheStatus.employees === 'invalid') {
      store.fetchEmployees()
    }
    if (store.cacheStatus.stats === 'invalid') {
      store.fetchEmployeeStats()
    }
  }, [store.cacheStatus.employees, store.cacheStatus.stats, store.fetchEmployees, store.fetchEmployeeStats])
  
  return {
    // === 数据 ===
    employees,
    employeeStats,
    selectedEmployee,
    filteredEmployees,
    
    // === 状态 ===
    isLoading,
    isRefreshing,
    hasErrors,
    errors: store.errors,
    
    // === 查询操作 ===
    fetchEmployees: store.fetchEmployees,
    fetchEmployeeStats: store.fetchEmployeeStats,
    searchEmployees: store.searchEmployees,
    refreshAll: store.refreshAll,
    
    // === 命令操作 ===
    createEmployee: store.createEmployee,
    updateEmployee: store.updateEmployee,
    terminateEmployee: store.terminateEmployee,
    
    // === UI 状态管理 ===
    filters: store.filters,
    searchQuery: store.searchQuery,
    viewMode: store.viewMode,
    selectedEmployeeIds: store.selectedEmployeeIds,
    
    setFilters: store.setFilters,
    setSearchQuery: store.setSearchQuery,
    setViewMode: store.setViewMode,
    selectEmployee: store.selectEmployee,
    toggleEmployeeSelection: store.toggleEmployeeSelection,
    clearSelections: store.clearSelections,
    
    // === 缓存管理 ===
    invalidateCache: store.invalidateCache,
    cacheStatus: store.cacheStatus,
    
    // === 实用工具 ===
    reset: store.reset
  }
}

/**
 * 员工选择 Hook - 用于获取单个员工详情
 */
export const useEmployee = (id: string | undefined) => {
  const employees = useEmployeeStore(employeeSelectors.employees)
  const fetchEmployees = useEmployeeStore(state => state.fetchEmployees)
  
  // 从本地状态查找员工
  const employee = employees.find(emp => emp.id === id)
  
  // 如果本地没有数据，触发获取
  useEffect(() => {
    if (id && !employee && employees.length === 0) {
      fetchEmployees()
    }
  }, [id, employee, employees.length])
  
  return {
    employee,
    isLoading: !employee && employees.length === 0,
    error: !employee && employees.length > 0 ? 'Employee not found' : null
  }
}

/**
 * 员工统计 Hook - 专门用于统计数据
 */
export const useEmployeeStats = () => {
  const employeeStats = useEmployeeStore(employeeSelectors.employeeStats)
  const fetchEmployeeStats = useEmployeeStore(state => state.fetchEmployeeStats)
  const isLoading = useEmployeeStore(state => state.queryStatus.loading)
  const errors = useEmployeeStore(state => state.errors)
  
  // 初始化时获取统计数据
  useEffect(() => {
    if (!employeeStats) {
      fetchEmployeeStats()
    }
  }, [employeeStats, fetchEmployeeStats])
  
  return {
    stats: employeeStats,
    isLoading,
    error: errors.fetchEmployeeStats || null,
    refresh: fetchEmployeeStats
  }
}

/**
 * 员工搜索 Hook - 专门用于搜索功能
 */
export const useEmployeeSearch = (initialQuery = '') => {
  const searchEmployees = useEmployeeStore(state => state.searchEmployees)
  const setSearchQuery = useEmployeeStore(state => state.setSearchQuery)
  const searchQuery = useEmployeeStore(state => state.searchQuery)
  const filteredEmployees = useEmployeeStore(employeeSelectors.filteredEmployees)
  const isLoading = useEmployeeStore(state => state.queryStatus.loading)
  
  // 初始化搜索查询
  useEffect(() => {
    if (initialQuery) {
      setSearchQuery(initialQuery)
    }
  }, [initialQuery])
  
  const search = async (query: string) => {
    setSearchQuery(query)
    if (query.trim()) {
      await searchEmployees(query)
    }
  }
  
  const clearSearch = () => {
    setSearchQuery('')
  }
  
  return {
    searchQuery,
    results: filteredEmployees,
    isLoading,
    search,
    clearSearch
  }
}

/**
 * 员工批量操作 Hook
 */
export const useEmployeeBulkOperations = () => {
  const selectedEmployeeIds = useEmployeeStore(state => state.selectedEmployeeIds)
  const toggleEmployeeSelection = useEmployeeStore(state => state.toggleEmployeeSelection)
  const clearSelections = useEmployeeStore(state => state.clearSelections)
  const employees = useEmployeeStore(state => state.employees)
  
  const selectedEmployees = employees.filter(emp => selectedEmployeeIds.has(emp.id))
  
  const selectAll = () => {
    employees.forEach(emp => {
      if (!selectedEmployeeIds.has(emp.id)) {
        toggleEmployeeSelection(emp.id)
      }
    })
  }
  
  const selectNone = () => {
    clearSelections()
  }
  
  const isSelected = (empId: string) => selectedEmployeeIds.has(empId)
  
  const selectedCount = selectedEmployeeIds.size
  
  return {
    selectedEmployeeIds,
    selectedEmployees,
    selectedCount,
    toggleEmployeeSelection,
    selectAll,
    selectNone,
    clearSelections,
    isSelected,
  }
}

/**
 * 员工过滤器 Hook - 专门用于过滤功能
 */
export const useEmployeeFilters = () => {
  const filters = useEmployeeStore(state => state.filters)
  const setFilters = useEmployeeStore(state => state.setFilters)
  const employees = useEmployeeStore(state => state.employees)
  const filteredEmployees = useEmployeeStore(employeeSelectors.filteredEmployees)
  
  // 获取所有可用的部门列表
  const availableDepartments = Array.from(
    new Set(employees.map(emp => emp.department).filter(Boolean))
  ).sort()
  
  // 获取所有可用的职位列表
  const availablePositions = Array.from(
    new Set(employees.map(emp => emp.position).filter(Boolean))
  ).sort()
  
  const clearFilters = () => {
    setFilters({})
  }
  
  const setStatusFilter = (status: 'active' | 'inactive' | 'pending' | undefined) => {
    setFilters({ status })
  }
  
  const setDepartmentFilter = (department: string | undefined) => {
    setFilters({ department })
  }
  
  const setPositionFilter = (position: string | undefined) => {
    setFilters({ position })
  }
  
  const hasActiveFilters = Object.keys(filters).length > 0
  
  return {
    filters,
    filteredEmployees,
    availableDepartments,
    availablePositions,
    hasActiveFilters,
    setFilters,
    clearFilters,
    setStatusFilter,
    setDepartmentFilter,
    setPositionFilter,
  }
}

/**
 * 员工命令操作 Hook - 专门用于命令操作
 */
export const useEmployeeCommands = () => {
  const createEmployee = useEmployeeStore(state => state.createEmployee)
  const updateEmployee = useEmployeeStore(state => state.updateEmployee)
  const terminateEmployee = useEmployeeStore(state => state.terminateEmployee)
  const commandStatus = useEmployeeStore(state => state.commandStatus)
  const errors = useEmployeeStore(state => state.errors)
  
  return {
    // 命令操作
    createEmployee,
    updateEmployee,
    terminateEmployee,
    
    // 状态
    isCreating: commandStatus.creating,
    isUpdating: commandStatus.updating,
    isTerminating: commandStatus.terminating,
    
    // 错误信息
    createError: errors.createEmployee,
    updateError: errors.updateEmployee,
    terminateError: errors.terminateEmployee,
    
    // 检查是否有任何操作在进行中
    isOperating: commandStatus.creating || commandStatus.updating || commandStatus.terminating,
  }
}

// 默认导出主要 Hook
export default useEmployeeCQRS