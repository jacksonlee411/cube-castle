// 类型转换工具函数
import { Employee, EmployeeApiResponse, EmployeeConverter } from '@/types'
import { logger } from '@/lib/logger';
import { isValidEmployeeApiResponse, isValidEmployee } from './type-guards'

/**
 * 员工数据类型转换器
 * 在前端统一类型和后端API字段格式之间进行转换
 */
export const employeeConverter: EmployeeConverter = {
  /**
   * 将API响应格式转换为前端统一Employee类型
   */
  fromApi: (apiData: EmployeeApiResponse): Employee => {
    return {
      id: apiData.id,
      createdAt: apiData.createdAt,
      updatedAt: apiData.updatedAt,
      employeeNumber: apiData.employee_number,
      firstName: apiData.first_name,
      lastName: apiData.last_name,
      fullName: `${apiData.last_name}${apiData.first_name}`, // 中文姓名格式：姓+名
      email: apiData.email,
      phoneNumber: apiData.phone_number ?? undefined,
      hireDate: apiData.hire_date,
      status: apiData.status,
      jobTitle: apiData.job_title ?? undefined,
      organizationId: apiData.organization_id ?? undefined,
      tenantId: apiData.tenant_id,
    }
  },

  /**
   * 将前端Employee类型转换为API请求格式
   */
  toApi: (employee: Partial<Employee>): Partial<EmployeeApiResponse> => {
    const apiData: Partial<EmployeeApiResponse> = {}
    
    if (employee.id) apiData.id = employee.id
    if (employee.createdAt) apiData.createdAt = employee.createdAt
    if (employee.updatedAt) apiData.updatedAt = employee.updatedAt
    if (employee.employeeNumber) apiData.employee_number = employee.employeeNumber
    if (employee.firstName) apiData.first_name = employee.firstName
    if (employee.lastName) apiData.last_name = employee.lastName
    if (employee.email) apiData.email = employee.email
    if (employee.phoneNumber) apiData.phone_number = employee.phoneNumber
    if (employee.hireDate) apiData.hire_date = employee.hireDate
    if (employee.status) apiData.status = employee.status
    if (employee.jobTitle) apiData.job_title = employee.jobTitle
    if (employee.organizationId) apiData.organization_id = employee.organizationId
    if (employee.tenantId) apiData.tenant_id = employee.tenantId
    
    return apiData
  }
}

/**
 * 批量转换员工数据
 */
export const convertEmployeesFromApi = (apiEmployees: EmployeeApiResponse[]): Employee[] => {
  return apiEmployees.map(employeeConverter.fromApi)
}

/**
 * 安全的类型转换函数，包含错误处理
 */
export const safeConvertEmployeeFromApi = (apiData: unknown): Employee | null => {
  try {
    if (!isValidEmployeeApiResponse(apiData)) {
      logger.warn('Invalid employee API response data:', apiData)
      return null
    }
    return employeeConverter.fromApi(apiData)
  } catch (error) {
    logger.error('Error converting employee from API:', error)
    return null
  }
}

// 导出类型守卫函数以供其他组件使用
export { isValidEmployee, isValidEmployeeApiResponse }