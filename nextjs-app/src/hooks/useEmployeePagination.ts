import { useState, useCallback, useEffect } from 'react'
import { employeeQueries, EmployeeSearchParams } from '@/lib/cqrs/employee-queries'
import { Employee } from '@/types/employee'

export interface PaginationState {
  currentPage: number
  pageSize: number
  totalCount: number
  totalPages: number
}

export interface EmployeePaginationHook {
  // Data
  employees: Employee[]
  pagination: PaginationState
  isLoading: boolean
  error: string | null
  
  // Actions
  loadPage: (page: number) => Promise<void>
  changePageSize: (pageSize: number) => Promise<void>
  applyFilters: (filters: Omit<EmployeeSearchParams, 'page' | 'pageSize' | 'limit' | 'offset'>) => Promise<void>
  refresh: () => Promise<void>
}

export const useEmployeePagination = (
  initialPageSize: number = 50
): EmployeePaginationHook => {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [currentFilters, setCurrentFilters] = useState<Omit<EmployeeSearchParams, 'page' | 'pageSize' | 'limit' | 'offset'>>({})
  
  const [pagination, setPagination] = useState<PaginationState>({
    currentPage: 1,
    pageSize: initialPageSize,
    totalCount: 0,
    totalPages: 0,
  })

  const fetchEmployees = useCallback(async (
    page: number, 
    pageSize: number, 
    filters: Omit<EmployeeSearchParams, 'page' | 'pageSize' | 'limit' | 'offset'> = {}
  ) => {
    setIsLoading(true)
    setError(null)
    
    try {
      const response = await employeeQueries.searchEmployeesWithPagination(
        page,
        pageSize,
        filters
      )
      
      setEmployees(response.employees)
      setPagination({
        currentPage: page,
        pageSize: pageSize,
        totalCount: response.total_count,
        totalPages: Math.ceil(response.total_count / pageSize),
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载员工数据失败')
      console.error('Failed to fetch employees:', err)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const loadPage = useCallback(async (page: number) => {
    await fetchEmployees(page, pagination.pageSize, currentFilters)
  }, [fetchEmployees, pagination.pageSize, currentFilters])

  const changePageSize = useCallback(async (pageSize: number) => {
    // When changing page size, go back to first page
    await fetchEmployees(1, pageSize, currentFilters)
  }, [fetchEmployees, currentFilters])

  const applyFilters = useCallback(async (filters: Omit<EmployeeSearchParams, 'page' | 'pageSize' | 'limit' | 'offset'>) => {
    setCurrentFilters(filters)
    // When applying filters, go back to first page
    await fetchEmployees(1, pagination.pageSize, filters)
  }, [fetchEmployees, pagination.pageSize])

  const refresh = useCallback(async () => {
    await fetchEmployees(pagination.currentPage, pagination.pageSize, currentFilters)
  }, [fetchEmployees, pagination.currentPage, pagination.pageSize, currentFilters])

  // Load initial data
  useEffect(() => {
    fetchEmployees(1, initialPageSize, {})
  }, [fetchEmployees, initialPageSize])

  return {
    employees,
    pagination,
    isLoading,
    error,
    loadPage,
    changePageSize,
    applyFilters,
    refresh,
  }
}