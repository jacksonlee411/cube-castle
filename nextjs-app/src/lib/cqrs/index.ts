// CQRS API 客户端统一导出
import { organizationCommands } from './commands'
import { organizationQueries } from './queries'
import { employeeCommands } from './employee-commands'
import { employeeQueries } from './employee-queries'

export { organizationCommands, organizationQueries, employeeCommands, employeeQueries }

// 重新导出类型以便于使用
export type { CreateOrganizationRequest, UpdateOrganizationRequest, Organization } from '@/types'
export type { Employee, EmployeeFilters, EmployeeStats } from '@/types/employee'
export type { 
  CreateEmployeeCommand, 
  UpdateEmployeeCommand, 
  TerminateEmployeeCommand,
  CommandResponse 
} from './employee-commands'
export type { 
  EmployeeSearchParams, 
  EmployeeSearchResponse 
} from './employee-queries'

// CQRS 操作状态枚举
export enum CQRSOperationStatus {
  IDLE = 'idle',
  PENDING = 'pending',
  SUCCESS = 'success',
  ERROR = 'error'
}

// CQRS 操作结果类型
export interface CQRSOperationResult<T = any> {
  status: CQRSOperationStatus
  data?: T
  error?: string
  timestamp: Date
}

// 乐观更新操作类型
export interface OptimisticUpdate<T = any> {
  id: string
  operation: 'create' | 'update' | 'delete'
  data: T
  timestamp: Date
  reverted?: boolean
}

// 事件总线类型
export interface OrganizationEvent {
  type: 'ORGANIZATION_CREATED' | 'ORGANIZATION_UPDATED' | 'ORGANIZATION_DELETED' | 'ORGANIZATION_MOVED'
  payload: {
    organization_id: string
    tenant_id: string
    organization_name?: string
    parent_id?: string
    changes?: Record<string, any>
  }
  timestamp: string
  event_id: string
}

// CQRS 客户端工具类
export class CQRSClient {
  static async executeCommand<T>(
    operation: () => Promise<T>,
    optimisticUpdate?: OptimisticUpdate
  ): Promise<CQRSOperationResult<T>> {
    try {
      const result = await operation()
      return {
        status: CQRSOperationStatus.SUCCESS,
        data: result,
        timestamp: new Date()
      }
    } catch (error) {
      return {
        status: CQRSOperationStatus.ERROR,
        error: error instanceof Error ? error.message : 'Unknown error',
        timestamp: new Date()
      }
    }
  }

  static async executeQuery<T>(
    operation: () => Promise<T>
  ): Promise<CQRSOperationResult<T>> {
    try {
      const result = await operation()
      return {
        status: CQRSOperationStatus.SUCCESS,
        data: result,
        timestamp: new Date()
      }
    } catch (error) {
      return {
        status: CQRSOperationStatus.ERROR,
        error: error instanceof Error ? error.message : 'Unknown error',
        timestamp: new Date()
      }
    }
  }
}

// 默认导出所有 CQRS 工具
export default {
  commands: organizationCommands,
  queries: organizationQueries,
  employeeCommands,
  employeeQueries,
  CQRSClient,
  CQRSOperationStatus
}