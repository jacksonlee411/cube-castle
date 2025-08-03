// Employee CQRS Commands - Command side implementation
import { Employee } from '@/types/employee'
import { logger } from '@/lib/logger'
import { API_BASE_URL, DEFAULT_TENANT_ID, getEmployeeCommandUrl } from '@/lib/routes'

// Employee Command DTOs
export interface CreateEmployeeCommand {
  tenant_id: string
  employee_type: 'FULL_TIME' | 'PART_TIME' | 'CONTRACTOR' | 'INTERN'
  first_name: string
  last_name: string
  email: string
  phone?: string
  hire_date: string
  department?: string
  position?: string
  manager_id?: string
}

export interface UpdateEmployeeCommand {
  id: string
  tenant_id: string
  first_name?: string
  last_name?: string
  email?: string
  phone?: string
  employment_status?: 'ACTIVE' | 'TERMINATED' | 'ON_LEAVE'
  department?: string
  position?: string
  manager_id?: string
}

export interface TerminateEmployeeCommand {
  id: string
  tenant_id: string
  termination_date: string
  reason?: string
}

// Command Response
export interface CommandResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  timestamp: Date
}

// Employee API Base URL
const EMPLOYEE_API_BASE = API_BASE_URL

/**
 * Employee Command API Client - CQRS Command Side
 * Uses CQRS command handlers for state mutations
 */
export class EmployeeCommandAPI {
  private readonly baseURL: string
  private readonly tenantId: string

  constructor(tenantId?: string) {
    this.baseURL = EMPLOYEE_API_BASE
    this.tenantId = tenantId || DEFAULT_TENANT_ID
  }

  /**
   * Create new employee using CQRS command handler
   */
  async createEmployee(command: Omit<CreateEmployeeCommand, 'tenant_id'>): Promise<CommandResponse<Employee>> {
    try {
      const url = getEmployeeCommandUrl('hire')
      const payload = {
        ...command,
        tenant_id: this.tenantId,
      }

      logger.info('Employee CQRS Command: CreateEmployee', { url, payload })

      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Tenant-ID': this.tenantId,
        },
        body: JSON.stringify(payload),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`HTTP ${response.status}: ${errorText}`)
      }

      const data = await response.json()
      
      return {
        success: true,
        data: this.transformEmployeeResponse(data),
        timestamp: new Date(),
      }
    } catch (error) {
      logger.error('Employee CQRS Command: CreateEmployee failed', { error, command })
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
        timestamp: new Date(),
      }
    }
  }

  /**
   * Update employee using CQRS command handler
   */
  async updateEmployee(command: Omit<UpdateEmployeeCommand, 'tenant_id'>): Promise<CommandResponse<Employee>> {
    try {
      const url = getEmployeeCommandUrl('update')
      const payload = {
        ...command,
        tenant_id: this.tenantId,
      }

      logger.info('Employee CQRS Command: UpdateEmployee', { url, payload })

      const response = await fetch(url, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-Tenant-ID': this.tenantId,
        },
        body: JSON.stringify(payload),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`HTTP ${response.status}: ${errorText}`)
      }

      const data = await response.json()
      
      return {
        success: true,
        data: this.transformEmployeeResponse(data),
        timestamp: new Date(),
      }
    } catch (error) {
      logger.error('Employee CQRS Command: UpdateEmployee failed', { error, command })
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
        timestamp: new Date(),
      }
    }
  }

  /**
   * Terminate employee using CQRS command handler
   */
  async terminateEmployee(command: Omit<TerminateEmployeeCommand, 'tenant_id'>): Promise<CommandResponse<boolean>> {
    try {
      const url = getEmployeeCommandUrl('terminate')
      const payload = {
        ...command,
        tenant_id: this.tenantId,
      }

      logger.info('Employee CQRS Command: TerminateEmployee', { url, payload })

      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Tenant-ID': this.tenantId,
        },
        body: JSON.stringify(payload),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`HTTP ${response.status}: ${errorText}`)
      }

      return {
        success: true,
        data: true,
        timestamp: new Date(),
      }
    } catch (error) {
      logger.error('Employee CQRS Command: TerminateEmployee failed', { error, command })
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
        timestamp: new Date(),
      }
    }
  }

  /**
   * Transform backend employee response to frontend Employee interface
   */
  private transformEmployeeResponse(data: any): Employee {
    // Handle both CQRS command response and direct employee data
    const emp = data.employee || data

    const firstName = emp.first_name || emp.firstName || ''
    const lastName = emp.last_name || emp.lastName || ''
    const legalName = emp.legal_name || `${firstName} ${lastName}`.trim() || emp.legalName || 'Unknown'

    return {
      id: emp.id || emp.employee_id || '',
      employeeId: emp.employee_id || emp.employeeId || emp.id || '',
      legalName,
      preferredName: emp.preferred_name || emp.preferredName || null,
      email: emp.email || '',
      phone: emp.phone || null,
      status: this.normalizeEmployeeStatus(emp.employment_status || emp.status || 'pending'),
      hireDate: emp.hire_date || emp.hireDate || new Date().toISOString(),
      department: emp.department || undefined,
      position: emp.position || undefined,
      managerId: emp.manager_id || emp.managerId || undefined,
      managerName: emp.manager_name || emp.managerName || null,
      avatar: emp.avatar || undefined,
    }
  }

  /**
   * Normalize employee status to frontend enum values
   */
  private normalizeEmployeeStatus(status: string): 'active' | 'inactive' | 'pending' {
    const normalizedStatus = status.toLowerCase()
    
    if (normalizedStatus === 'active') return 'active'
    if (normalizedStatus === 'inactive' || normalizedStatus === 'terminated') return 'inactive'
    
    return 'pending'
  }
}

/**
 * Employee CQRS Commands - Singleton instance
 */
export const employeeCommands = new EmployeeCommandAPI()

// Export for convenience
export { employeeCommands as default }