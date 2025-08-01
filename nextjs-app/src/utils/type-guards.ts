/**
 * 类型守卫和运行时类型检查工具
 * 用于确保数据类型安全和运行时验证
 */

import { 
  Employee, 
  EmployeeApiResponse, 
  Organization, 
  EmployeeStatus,
  BaseEntity,
  PaginationInfo
} from '@/types'
import { logger } from '@/lib/logger';

/**
 * 检查对象是否为有效的BaseEntity
 */
export const isBaseEntity = (obj: unknown): obj is BaseEntity => {
  return (
    typeof obj === 'object' &&
    obj !== null &&
    typeof (obj as BaseEntity).id === 'string' &&
    typeof (obj as BaseEntity).createdAt === 'string' &&
    typeof (obj as BaseEntity).updatedAt === 'string'
  )
}

/**
 * 检查字符串是否为有效的EmployeeStatus
 */
export const isValidEmployeeStatus = (status: string): status is EmployeeStatus => {
  return Object.values(EmployeeStatus).includes(status as EmployeeStatus)
}

/**
 * 检查对象是否为有效的Employee对象
 */
export const isValidEmployee = (obj: unknown): obj is Employee => {
  if (!obj || typeof obj !== 'object') return false
  
  const employee = obj as Employee
  
  return (
    isBaseEntity(employee) &&
    typeof employee.employeeNumber === 'string' &&
    typeof employee.firstName === 'string' &&
    typeof employee.lastName === 'string' &&
    typeof employee.fullName === 'string' &&
    typeof employee.email === 'string' &&
    typeof employee.hireDate === 'string' &&
    isValidEmployeeStatus(employee.status) &&
    typeof employee.tenantId === 'string' &&
    (employee.phoneNumber === undefined || typeof employee.phoneNumber === 'string') &&
    (employee.jobTitle === undefined || typeof employee.jobTitle === 'string') &&
    (employee.department === undefined || typeof employee.department === 'string') &&
    (employee.managerId === undefined || typeof employee.managerId === 'string') &&
    (employee.organizationId === undefined || typeof employee.organizationId === 'string')
  )
}

/**
 * 检查对象是否为有效的EmployeeApiResponse对象
 */
export const isValidEmployeeApiResponse = (obj: unknown): obj is EmployeeApiResponse => {
  if (!obj || typeof obj !== 'object') return false
  
  const apiResponse = obj as EmployeeApiResponse
  
  return (
    isBaseEntity(apiResponse) &&
    typeof apiResponse.employee_number === 'string' &&
    typeof apiResponse.first_name === 'string' &&
    typeof apiResponse.last_name === 'string' &&
    typeof apiResponse.email === 'string' &&
    typeof apiResponse.hire_date === 'string' &&
    isValidEmployeeStatus(apiResponse.status) &&
    typeof apiResponse.tenant_id === 'string' &&
    (apiResponse.phone_number === undefined || typeof apiResponse.phone_number === 'string') &&
    (apiResponse.job_title === undefined || typeof apiResponse.job_title === 'string') &&
    (apiResponse.organization_id === undefined || typeof apiResponse.organization_id === 'string') &&
    (apiResponse.position_id === undefined || typeof apiResponse.position_id === 'string') &&
    (apiResponse.manager_id === undefined || typeof apiResponse.manager_id === 'string')
  )
}

/**
 * 检查对象是否为有效的Organization对象
 */
export const isValidOrganization = (obj: unknown): obj is Organization => {
  if (!obj || typeof obj !== 'object') return false
  
  const org = obj as Organization
  
  return (
    isBaseEntity(org) &&
    typeof org.name === 'string' &&
    typeof org.code === 'string' &&
    typeof org.level === 'number' &&
    typeof org.tenantId === 'string' &&
    (org.description === undefined || typeof org.description === 'string') &&
    (org.parentId === undefined || typeof org.parentId === 'string') &&
    (org.employeeCount === undefined || typeof org.employeeCount === 'number') &&
    (org.type === undefined || ['company', 'department', 'team'].includes(org.type)) &&
    (org.status === undefined || ['active', 'inactive'].includes(org.status)) &&
    (org.managerName === undefined || typeof org.managerName === 'string')
  )
}

/**
 * 检查对象是否为有效的PaginationInfo对象
 */
export const isValidPaginationInfo = (obj: unknown): obj is PaginationInfo => {
  if (!obj || typeof obj !== 'object') return false
  
  const pagination = obj as PaginationInfo
  
  return (
    typeof pagination.page === 'number' &&
    typeof pagination.pageSize === 'number' &&
    typeof pagination.total === 'number' &&
    typeof pagination.totalPages === 'number' &&
    pagination.page > 0 &&
    pagination.pageSize > 0 &&
    pagination.total >= 0 &&
    pagination.totalPages >= 0
  )
}

/**
 * 安全的数据验证函数，返回验证结果和错误信息
 */
export const validateEmployee = (obj: unknown): { 
  isValid: boolean
  employee?: Employee
  errors: string[] 
} => {
  const errors: string[] = []
  
  if (!obj || typeof obj !== 'object') {
    errors.push('Employee data must be a valid object')
    return { isValid: false, errors }
  }
  
  const employee = obj as Partial<Employee>
  
  if (!employee.id || typeof employee.id !== 'string') {
    errors.push('Employee ID is required and must be a string')
  }
  
  if (!employee.employeeNumber || typeof employee.employeeNumber !== 'string') {
    errors.push('Employee number is required and must be a string')
  }
  
  if (!employee.firstName || typeof employee.firstName !== 'string') {
    errors.push('First name is required and must be a string')
  }
  
  if (!employee.lastName || typeof employee.lastName !== 'string') {
    errors.push('Last name is required and must be a string')
  }
  
  if (!employee.email || typeof employee.email !== 'string') {
    errors.push('Email is required and must be a string')
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(employee.email)) {
    errors.push('Email format is invalid')
  }
  
  if (!employee.status || !isValidEmployeeStatus(employee.status)) {
    errors.push('Valid employee status is required')
  }
  
  if (!employee.tenantId || typeof employee.tenantId !== 'string') {
    errors.push('Tenant ID is required and must be a string')
  }
  
  if (errors.length === 0) {
    return { 
      isValid: true, 
      employee: employee as Employee,
      errors: [] 
    }
  }
  
  return { isValid: false, errors }
}

/**
 * 运行时类型断言函数
 */
export const assertEmployee = (obj: unknown, context = 'Unknown'): Employee => {
  const validation = validateEmployee(obj)
  
  if (!validation.isValid) {
    throw new TypeError(
      `${context}: Invalid employee data. Errors: ${validation.errors.join(', ')}`
    )
  }
  
  return validation.employee!
}

/**
 * 安全的类型转换函数，带有错误处理
 */
export const safeTypeConversion = <T>(
  obj: unknown,
  typeGuard: (obj: unknown) => obj is T,
  fallback: T,
  context = 'Unknown'
): T => {
  try {
    if (typeGuard(obj)) {
      return obj
    }
    
    logger.warn(`${context}: Type conversion failed, using fallback value`)
    return fallback
  } catch (error) {
    logger.error(`${context}: Error during type conversion:`, error)
    return fallback
  }
}

/**
 * 批量验证数组中的数据
 */
export const validateArray = <T>(
  array: unknown[],
  typeGuard: (obj: unknown) => obj is T,
  context = 'Unknown'
): { validItems: T[], invalidItems: unknown[], errors: string[] } => {
  const validItems: T[] = []
  const invalidItems: unknown[] = []
  const errors: string[] = []
  
  array.forEach((item, index) => {
    if (typeGuard(item)) {
      validItems.push(item)
    } else {
      invalidItems.push(item)
      errors.push(`${context}[${index}]: Invalid item type`)
    }
  })
  
  return { validItems, invalidItems, errors }
}