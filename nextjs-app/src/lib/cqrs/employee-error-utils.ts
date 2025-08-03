/**
 * 员工CQRS专用错误处理工具
 * 提供统一的错误处理、恢复机制和用户友好的错误信息
 */

import { 
  CQRSError, 
  CQRSErrorFactory, 
  CQRSErrorType, 
  ErrorSeverity, 
  RetryManager, 
  ErrorReporter,
  defaultRetryConfig 
} from '@/lib/cqrs-error-handling'
import { logger } from '@/lib/logger'
import toast from 'react-hot-toast'

// Employee specific error types
export interface EmployeeErrorContext {
  operation: 'search' | 'get' | 'create' | 'update' | 'terminate' | 'stats'
  employeeId?: string
  tenantId?: string
  searchParams?: Record<string, any>
  requestId?: string
}

// Error recovery strategies
export interface ErrorRecoveryStrategy {
  shouldShowToast: boolean
  fallbackData?: any
  retryConfig?: Partial<typeof defaultRetryConfig>
  userActionRequired?: 'login' | 'refresh' | 'contact_support' | 'none'
}

/**
 * Employee CQRS Error Handler
 * 为员工模块提供专门的错误处理和恢复策略
 */
export class EmployeeErrorHandler {
  private static instance: EmployeeErrorHandler
  private retryManager: RetryManager
  private errorReporter: ErrorReporter

  static getInstance(): EmployeeErrorHandler {
    if (!EmployeeErrorHandler.instance) {
      EmployeeErrorHandler.instance = new EmployeeErrorHandler()
    }
    return EmployeeErrorHandler.instance
  }

  constructor() {
    // 为员工操作定制的重试配置
    this.retryManager = new RetryManager({
      ...defaultRetryConfig,
      maxAttempts: 3,
      baseDelay: 1000,
      maxDelay: 10000, // 员工操作的最大延迟较短
    })
    this.errorReporter = ErrorReporter.getInstance()
  }

  /**
   * 处理员工查询错误
   */
  handleQueryError(
    error: Error | CQRSError, 
    context: EmployeeErrorContext
  ): { cqrsError: CQRSError; strategy: ErrorRecoveryStrategy } {
    const cqrsError = this.convertToCQRSError(error, context)
    const strategy = this.getRecoveryStrategy(cqrsError, context)
    
    // 记录错误
    this.errorReporter.report(cqrsError)
    
    // 根据策略显示用户提示
    if (strategy.shouldShowToast) {
      this.showUserNotification(cqrsError, context.operation)
    }
    
    logger.error(`Employee ${context.operation} error`, {
      error: cqrsError.toLogFormat(),
      context,
      strategy
    })
    
    return { cqrsError, strategy }
  }

  /**
   * 处理员工命令错误
   */
  handleCommandError(
    error: Error | CQRSError,
    context: EmployeeErrorContext
  ): { cqrsError: CQRSError; strategy: ErrorRecoveryStrategy } {
    const cqrsError = this.convertToCQRSError(error, context)
    const strategy = this.getCommandRecoveryStrategy(cqrsError, context)
    
    // 命令错误总是需要报告和用户通知
    this.errorReporter.report(cqrsError)
    this.showUserNotification(cqrsError, context.operation)
    
    logger.error(`Employee ${context.operation} command error`, {
      error: cqrsError.toLogFormat(),
      context,
      strategy
    })
    
    return { cqrsError, strategy }
  }

  /**
   * 执行带重试的员工查询操作
   */
  async executeQueryWithRetry<T>(
    operation: () => Promise<T>,
    context: EmployeeErrorContext
  ): Promise<T> {
    try {
      return await this.retryManager.executeWithRetry(operation, context)
    } catch (error) {
      const { cqrsError, strategy } = this.handleQueryError(error as Error, context)
      
      // 如果有fallback数据，返回fallback而不是抛出错误
      if (strategy.fallbackData !== undefined) {
        logger.info(`Using fallback data for employee ${context.operation}`, {
          context,
          fallbackData: strategy.fallbackData
        })
        return strategy.fallbackData as T
      }
      
      throw cqrsError
    }
  }

  /**
   * 执行带重试的员工命令操作
   */
  async executeCommandWithRetry<T>(
    operation: () => Promise<T>,
    context: EmployeeErrorContext
  ): Promise<T> {
    // 命令操作使用更保守的重试策略
    const conservativeRetryManager = new RetryManager({
      ...defaultRetryConfig,
      maxAttempts: 2, // 命令操作最多重试2次
      retryableErrors: [
        CQRSErrorType.NETWORK_ERROR,
        CQRSErrorType.TIMEOUT_ERROR,
        CQRSErrorType.CONNECTION_ERROR,
        CQRSErrorType.BAD_GATEWAY,
        CQRSErrorType.SERVICE_UNAVAILABLE,
        // 不重试CONFLICT错误，避免重复创建
      ],
    })

    try {
      return await conservativeRetryManager.executeWithRetry(operation, context)
    } catch (error) {
      const { cqrsError } = this.handleCommandError(error as Error, context)
      throw cqrsError
    }
  }

  /**
   * 转换为CQRSError格式
   */
  private convertToCQRSError(error: Error | CQRSError, context: EmployeeErrorContext): CQRSError {
    if (error instanceof CQRSError) {
      return error
    }

    // 检查是否是网络错误
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
      return CQRSErrorFactory.fromNetworkError(error, context, context.requestId)
    }

    // 检查是否是超时错误
    if (error.name === 'AbortError' || error.message.includes('timeout')) {
      return CQRSErrorFactory.fromNetworkError(error, context, context.requestId)
    }

    // 其他错误
    return CQRSErrorFactory.fromNetworkError(error, context, context.requestId)
  }

  /**
   * 获取查询错误的恢复策略
   */
  private getRecoveryStrategy(cqrsError: CQRSError, context: EmployeeErrorContext): ErrorRecoveryStrategy {
    const baseStrategy: ErrorRecoveryStrategy = {
      shouldShowToast: true,
      userActionRequired: 'none',
    }

    switch (cqrsError.type) {
      case CQRSErrorType.UNAUTHORIZED:
        return {
          ...baseStrategy,
          userActionRequired: 'login',
          shouldShowToast: true,
        }

      case CQRSErrorType.FORBIDDEN:
        return {
          ...baseStrategy,
          shouldShowToast: true,
          userActionRequired: 'contact_support',
        }

      case CQRSErrorType.NOT_FOUND:
        // 员工不存在是正常情况，不需要显示错误
        return {
          ...baseStrategy,
          shouldShowToast: false,
          fallbackData: context.operation === 'stats' ? {
            total: 0,
            active: 0,
            inactive: 0,
            pending: 0,
            departments: 0,
          } : null,
        }

      case CQRSErrorType.NETWORK_ERROR:
      case CQRSErrorType.CONNECTION_ERROR:
        return {
          ...baseStrategy,
          shouldShowToast: cqrsError.severity === ErrorSeverity.HIGH,
          userActionRequired: 'refresh',
          fallbackData: this.getFallbackData(context),
        }

      case CQRSErrorType.SERVICE_UNAVAILABLE:
      case CQRSErrorType.BAD_GATEWAY:
        return {
          ...baseStrategy,
          shouldShowToast: false, // 服务不可用时使用fallback数据，不显示错误
          fallbackData: this.getFallbackData(context),
        }

      case CQRSErrorType.TIMEOUT_ERROR:
        return {
          ...baseStrategy,
          shouldShowToast: cqrsError.severity !== ErrorSeverity.LOW,
          userActionRequired: 'refresh',
        }

      default:
        return {
          ...baseStrategy,
          shouldShowToast: cqrsError.severity !== ErrorSeverity.LOW,
          userActionRequired: cqrsError.severity === ErrorSeverity.CRITICAL ? 'contact_support' : 'refresh',
        }
    }
  }

  /**
   * 获取命令错误的恢复策略
   */
  private getCommandRecoveryStrategy(cqrsError: CQRSError, context: EmployeeErrorContext): ErrorRecoveryStrategy {
    const baseStrategy: ErrorRecoveryStrategy = {
      shouldShowToast: true,
      userActionRequired: 'none',
    }

    switch (cqrsError.type) {
      case CQRSErrorType.VALIDATION_ERROR:
        return {
          ...baseStrategy,
          userActionRequired: 'none', // 用户需要修正输入
        }

      case CQRSErrorType.CONFLICT:
        return {
          ...baseStrategy,
          userActionRequired: 'refresh', // 可能是数据冲突，建议刷新
        }

      case CQRSErrorType.UNAUTHORIZED:
        return {
          ...baseStrategy,
          userActionRequired: 'login',
        }

      case CQRSErrorType.FORBIDDEN:
        return {
          ...baseStrategy,
          userActionRequired: 'contact_support',
        }

      default:
        return {
          ...baseStrategy,
          userActionRequired: cqrsError.severity === ErrorSeverity.CRITICAL ? 'contact_support' : 'refresh',
        }
    }
  }

  /**
   * 获取fallback数据
   */
  private getFallbackData(context: EmployeeErrorContext): any {
    switch (context.operation) {
      case 'search':
        return {
          employees: [],
          total_count: 0,
          limit: 20,
          offset: 0,
        }

      case 'stats':
        return {
          total: 0,
          active: 0,
          inactive: 0,
          pending: 0,
          departments: 0,
        }

      case 'get':
        return null

      default:
        return undefined
    }
  }

  /**
   * 显示用户通知
   */
  private showUserNotification(cqrsError: CQRSError, operation: string): void {
    const operationNames = {
      search: '搜索员工',
      get: '获取员工信息',
      create: '创建员工',
      update: '更新员工',
      terminate: '员工离职',
      stats: '获取统计信息',
    }

    const operationName = operationNames[operation as keyof typeof operationNames] || operation

    switch (cqrsError.severity) {
      case ErrorSeverity.CRITICAL:
        toast.error(`${operationName}失败: ${cqrsError.userMessage}`, {
          duration: 8000,
          icon: '🚨',
        })
        break

      case ErrorSeverity.HIGH:
        toast.error(`${operationName}失败: ${cqrsError.userMessage}`, {
          duration: 6000,
        })
        break

      case ErrorSeverity.MEDIUM:
        toast.error(`${operationName}失败: ${cqrsError.userMessage}`, {
          duration: 4000,
        })
        break

      case ErrorSeverity.LOW:
        // 低严重程度的错误不显示toast，避免干扰用户
        break
    }
  }

  /**
   * 获取错误统计信息
   */
  getErrorStats() {
    return this.errorReporter.getErrorStats()
  }
}

/**
 * 便捷函数：处理员工查询错误
 */
export const handleEmployeeQueryError = (
  error: Error | CQRSError,
  context: EmployeeErrorContext
) => {
  return EmployeeErrorHandler.getInstance().handleQueryError(error, context)
}

/**
 * 便捷函数：处理员工命令错误
 */
export const handleEmployeeCommandError = (
  error: Error | CQRSError,
  context: EmployeeErrorContext
) => {
  return EmployeeErrorHandler.getInstance().handleCommandError(error, context)
}

/**
 * 便捷函数：执行带重试的员工查询
 */
export const executeEmployeeQueryWithRetry = <T>(
  operation: () => Promise<T>,
  context: EmployeeErrorContext
): Promise<T> => {
  return EmployeeErrorHandler.getInstance().executeQueryWithRetry(operation, context)
}

/**
 * 便捷函数：执行带重试的员工命令
 */
export const executeEmployeeCommandWithRetry = <T>(
  operation: () => Promise<T>,
  context: EmployeeErrorContext
): Promise<T> => {
  return EmployeeErrorHandler.getInstance().executeCommandWithRetry(operation, context)
}

/**
 * 创建员工操作的请求ID
 */
export const createEmployeeRequestId = (operation: string, employeeId?: string): string => {
  const timestamp = Date.now()
  const random = Math.random().toString(36).substr(2, 9)
  const suffix = employeeId ? `-${employeeId}` : ''
  return `employee-${operation}-${timestamp}-${random}${suffix}`
}

// 导出单例实例
export const employeeErrorHandler = EmployeeErrorHandler.getInstance()