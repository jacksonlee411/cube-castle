/**
 * 员工管理API客户端
 * 连接前端Employee接口与后端Person API
 */

import { Employee, EmployeeStatus, CreateEmployeeRequest, UpdateEmployeeRequest, EmployeeListResponse } from '@/types'

// API基础配置
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'
const EMPLOYEES_ENDPOINT = '/api/v1/corehr/employees'

// 错误类定义
export class ApiError extends Error {
  constructor(
    message: string,
    public statusCode: number,
    public code?: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export class ValidationApiError extends Error {
  constructor(
    message: string,
    public validationErrors: Array<{
      field: string
      message: string
      code: string
    }>
  ) {
    super(message)
    this.name = 'ValidationApiError'
  }

  getFieldError(field: string): string | undefined {
    const error = this.validationErrors.find(e => e.field === field)
    return error?.message
  }

  getAllErrors(): Record<string, string> {
    const errors: Record<string, string> = {}
    this.validationErrors.forEach(error => {
      errors[error.field] = error.message
    })
    return errors
  }
}

// 后端Person API响应类型
interface PersonApiResponse {
  id: string
  tenant_id: string
  employee_id: string
  legal_name: string
  preferred_name?: string
  email: string
  status: string
  hire_date: string
  termination_date?: string
  created_at: string
  updated_at: string
}

interface PersonCreateRequest {
  employee_id: string
  legal_name: string
  preferred_name?: string
  email: string
  status: string
  hire_date: string
  termination_date?: string
}

// 数据转换工具函数
export class EmployeeApiAdapter {
  /**
   * 后端Person数据转换为前端Employee
   */
  static toEmployee(person: PersonApiResponse): Employee {
    // 解析legal_name为first_name和last_name
    const nameParts = person.legal_name.split(' ')
    const firstName = nameParts[0] || ''
    const lastName = nameParts.slice(1).join(' ') || ''

    return {
      id: person.id,
      employeeNumber: person.employee_id,
      firstName,
      lastName,
      fullName: person.legal_name,
      email: person.email,
      hireDate: person.hire_date,
      status: this.mapStatus(person.status),
      tenantId: person.tenant_id,
      createdAt: person.created_at,
      updatedAt: person.updated_at,
      // 可选字段
      phoneNumber: undefined, // 后端暂不支持
      jobTitle: undefined, // 需要从Position关联获取
      department: undefined, // 需要从Organization关联获取
      managerId: undefined, // 需要从关联关系获取
      organizationId: undefined, // 需要从关联关系获取
    }
  }

  /**
   * 前端CreateEmployeeRequest转换为后端请求
   */
  static toPersonCreateRequest(employee: CreateEmployeeRequest): PersonCreateRequest {
    return {
      employee_id: employee.employeeNumber,
      legal_name: `${employee.firstName} ${employee.lastName}`.trim(),
      preferred_name: employee.firstName !== employee.lastName ? employee.firstName : undefined,
      email: employee.email,
      status: 'ACTIVE',
      hire_date: employee.hireDate,
    }
  }

  /**
   * 前端UpdateEmployeeRequest转换为后端请求
   */
  static toPersonUpdateRequest(employee: UpdateEmployeeRequest): Partial<PersonCreateRequest> {
    const request: Partial<PersonCreateRequest> = {}
    
    if (employee.firstName || employee.lastName) {
      request.legal_name = `${employee.firstName || ''} ${employee.lastName || ''}`.trim()
      request.preferred_name = employee.firstName
    }
    
    if (employee.email) {
      request.email = employee.email
    }
    
    if (employee.status) {
      request.status = this.mapStatusToBackend(employee.status)
    }

    return request
  }

  /**
   * 后端状态映射到前端状态
   */
  private static mapStatus(backendStatus: string): EmployeeStatus {
    switch (backendStatus?.toUpperCase()) {
      case 'ACTIVE':
        return EmployeeStatus.ACTIVE
      case 'INACTIVE':
        return EmployeeStatus.INACTIVE
      case 'TERMINATED':
        return EmployeeStatus.TERMINATED
      default:
        return EmployeeStatus.INACTIVE
    }
  }

  /**
   * 前端状态映射到后端状态
   */
  private static mapStatusToBackend(frontendStatus: EmployeeStatus): string {
    switch (frontendStatus) {
      case EmployeeStatus.ACTIVE:
        return 'ACTIVE'
      case EmployeeStatus.INACTIVE:
        return 'INACTIVE'
      case EmployeeStatus.TERMINATED:
        return 'TERMINATED'
      case EmployeeStatus.ON_LEAVE:
        return 'ACTIVE' // 请假状态映射为激活状态
      default:
        return 'INACTIVE'
    }
  }
}

// HTTP客户端工具
class ApiClient {
  private baseUrl: string

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl
  }

  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`
    
    const defaultHeaders = {
      'Content-Type': 'application/json',
      // TODO: 添加认证头
      // 'Authorization': `Bearer ${getAuthToken()}`,
      // 'X-Tenant-ID': getTenantId(),
    }

    const config: RequestInit = {
      ...options,
      headers: {
        ...defaultHeaders,
        ...options.headers,
      },
    }

    try {
      const response = await fetch(url, config)
      
      const contentType = response.headers.get('content-type')
      let responseData: any = {}
      
      if (contentType && contentType.includes('application/json')) {
        responseData = await response.json()
      }
      
      if (!response.ok) {
        // 处理验证错误
        if (response.status === 400 && responseData.validation_errors) {
          const validationError = new ValidationApiError(
            'Validation failed',
            responseData.validation_errors
          )
          throw validationError
        }
        
        // 处理其他API错误
        const errorMessage = responseData.error?.message || `API请求失败: ${response.status} ${response.statusText}`
        const apiError = new ApiError(errorMessage, response.status, responseData.error?.code)
        throw apiError
      }

      return responseData
    } catch (error) {
      if (error instanceof ValidationApiError || error instanceof ApiError) {
        throw error
      }
      
      // API请求错误 - error handled by throwing specific error
      throw new ApiError('网络请求失败，请检查网络连接', 0, 'NETWORK_ERROR')
    }
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' })
  }

  async post<T>(endpoint: string, data: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async put<T>(endpoint: string, data: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' })
  }
}

// 员工API服务类
export class EmployeesApi {
  private client: ApiClient

  constructor() {
    this.client = new ApiClient()
  }

  /**
   * 获取员工列表
   */
  async getEmployees(params: {
    limit?: number
    offset?: number
    search?: string
  } = {}): Promise<EmployeeListResponse> {
    const queryParams = new URLSearchParams()
    
    if (params.limit) queryParams.set('limit', params.limit.toString())
    if (params.offset) queryParams.set('offset', params.offset.toString())
    if (params.search) queryParams.set('search', params.search)

    const endpoint = `${EMPLOYEES_ENDPOINT}?${queryParams}`
    const persons = await this.client.get<PersonApiResponse[]>(endpoint)
    
    // 转换数据格式
    const employees = persons.map(person => EmployeeApiAdapter.toEmployee(person))
    
    return {
      employees,
      pagination: {
        page: Math.floor((params.offset || 0) / (params.limit || 100)) + 1,
        pageSize: params.limit || 100,
        total: employees.length, // TODO: 后端应返回总数
        totalPages: Math.ceil(employees.length / (params.limit || 100))
      }
    }
  }

  /**
   * 根据ID获取员工
   */
  async getEmployee(id: string): Promise<Employee> {
    const person = await this.client.get<PersonApiResponse>(`${EMPLOYEES_ENDPOINT}/${id}`)
    return EmployeeApiAdapter.toEmployee(person)
  }

  /**
   * 创建员工
   */
  async createEmployee(employee: CreateEmployeeRequest): Promise<Employee> {
    const personRequest = EmployeeApiAdapter.toPersonCreateRequest(employee)
    const person = await this.client.post<PersonApiResponse>(EMPLOYEES_ENDPOINT, personRequest)
    return EmployeeApiAdapter.toEmployee(person)
  }

  /**
   * 更新员工
   */
  async updateEmployee(id: string, employee: UpdateEmployeeRequest): Promise<Employee> {
    const personRequest = EmployeeApiAdapter.toPersonUpdateRequest(employee)
    const person = await this.client.put<PersonApiResponse>(`${EMPLOYEES_ENDPOINT}/${id}`, personRequest)
    return EmployeeApiAdapter.toEmployee(person)
  }

  /**
   * 删除员工
   */
  async deleteEmployee(id: string): Promise<void> {
    await this.client.delete(`${EMPLOYEES_ENDPOINT}/${id}`)
  }

  /**
   * 批量操作
   */
  async batchUpdateEmployees(updates: Array<{ id: string; data: UpdateEmployeeRequest }>): Promise<Employee[]> {
    const promises = updates.map(update => 
      this.updateEmployee(update.id, update.data)
    )
    return Promise.all(promises)
  }

  /**
   * 搜索员工
   */
  async searchEmployees(query: string, filters?: {
    department?: string
    status?: EmployeeStatus
  }): Promise<Employee[]> {
    // TODO: 实现高级搜索
    const response = await this.getEmployees({ 
      search: query,
      limit: 1000 
    })
    
    let employees = response.employees

    // 前端过滤（临时方案，应该在后端实现）
    if (filters?.department) {
      employees = employees.filter(emp => emp.department === filters.department)
    }
    
    if (filters?.status) {
      employees = employees.filter(emp => emp.status === filters.status)
    }

    return employees
  }
}

// 导出单例实例
export const employeesApi = new EmployeesApi()

// 便捷的钩子函数（可选）
export async function fetchEmployees(params?: {
  limit?: number
  offset?: number
  search?: string
}) {
  return employeesApi.getEmployees(params)
}

export async function fetchEmployee(id: string) {
  return employeesApi.getEmployee(id)
}

export async function createEmployee(employee: CreateEmployeeRequest) {
  return employeesApi.createEmployee(employee)
}

export async function updateEmployee(id: string, employee: UpdateEmployeeRequest) {
  return employeesApi.updateEmployee(id, employee)
}

export async function deleteEmployee(id: string) {
  return employeesApi.deleteEmployee(id)
}