/**
 * CQRS 错误处理系统
 * 提供统一的错误分类、重试机制和用户友好的错误信息
 */

// 错误类型枚举
export enum CQRSErrorType {
  // 网络错误
  NETWORK_ERROR = 'NETWORK_ERROR',
  TIMEOUT_ERROR = 'TIMEOUT_ERROR',
  CONNECTION_ERROR = 'CONNECTION_ERROR',
  
  // HTTP错误
  BAD_REQUEST = 'BAD_REQUEST',
  UNAUTHORIZED = 'UNAUTHORIZED',
  FORBIDDEN = 'FORBIDDEN',
  NOT_FOUND = 'NOT_FOUND',
  CONFLICT = 'CONFLICT',
  INTERNAL_SERVER_ERROR = 'INTERNAL_SERVER_ERROR',
  BAD_GATEWAY = 'BAD_GATEWAY',
  SERVICE_UNAVAILABLE = 'SERVICE_UNAVAILABLE',
  
  // 业务逻辑错误
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  BUSINESS_RULE_ERROR = 'BUSINESS_RULE_ERROR',
  TENANT_ACCESS_ERROR = 'TENANT_ACCESS_ERROR',
  
  // 系统错误
  PARSING_ERROR = 'PARSING_ERROR',
  UNKNOWN_ERROR = 'UNKNOWN_ERROR',
}

// 错误严重程度
export enum ErrorSeverity {
  LOW = 'low',       // 可忽略，不影响用户操作
  MEDIUM = 'medium', // 需要注意，可能影响部分功能
  HIGH = 'high',     // 严重，影响核心功能
  CRITICAL = 'critical', // 致命，系统不可用
}

// 增强的错误类
export class CQRSError extends Error {
  public readonly type: CQRSErrorType
  public readonly severity: ErrorSeverity
  public readonly userMessage: string
  public readonly technicalMessage: string
  public readonly retryable: boolean
  public readonly context?: Record<string, any>
  public readonly timestamp: Date
  public readonly requestId?: string

  constructor(
    type: CQRSErrorType,
    technicalMessage: string,
    userMessage?: string,
    severity: ErrorSeverity = ErrorSeverity.MEDIUM,
    retryable: boolean = false,
    context?: Record<string, any>,
    requestId?: string
  ) {
    super(technicalMessage)
    this.name = 'CQRSError'
    this.type = type
    this.severity = severity
    this.userMessage = userMessage || this.getDefaultUserMessage(type)
    this.technicalMessage = technicalMessage
    this.retryable = retryable
    this.context = context
    this.timestamp = new Date()
    this.requestId = requestId
  }

  private getDefaultUserMessage(type: CQRSErrorType): string {
    const messages: Record<CQRSErrorType, string> = {
      [CQRSErrorType.NETWORK_ERROR]: '网络连接失败，请检查网络设置',
      [CQRSErrorType.TIMEOUT_ERROR]: '请求超时，请稍后重试',
      [CQRSErrorType.CONNECTION_ERROR]: '无法连接到服务器，请稍后重试',
      
      [CQRSErrorType.BAD_REQUEST]: '请求参数有误，请检查输入信息',
      [CQRSErrorType.UNAUTHORIZED]: '未授权访问，请重新登录',
      [CQRSErrorType.FORBIDDEN]: '权限不足，无法执行此操作',
      [CQRSErrorType.NOT_FOUND]: '请求的资源不存在',
      [CQRSErrorType.CONFLICT]: '数据冲突，请刷新后重试',
      [CQRSErrorType.INTERNAL_SERVER_ERROR]: '服务器内部错误，请联系管理员',
      [CQRSErrorType.BAD_GATEWAY]: '服务暂时不可用，请稍后重试',
      [CQRSErrorType.SERVICE_UNAVAILABLE]: '服务维护中，请稍后重试',
      
      [CQRSErrorType.VALIDATION_ERROR]: '输入数据验证失败，请检查表单',
      [CQRSErrorType.BUSINESS_RULE_ERROR]: '操作违反业务规则',
      [CQRSErrorType.TENANT_ACCESS_ERROR]: '租户访问权限错误',
      
      [CQRSErrorType.PARSING_ERROR]: '数据解析错误，请联系管理员',
      [CQRSErrorType.UNKNOWN_ERROR]: '未知错误，请稍后重试',
    }
    
    return messages[type] || '系统错误，请稍后重试'
  }

  // 判断是否需要立即重试
  public shouldRetryImmediately(): boolean {
    return this.retryable && [
      CQRSErrorType.TIMEOUT_ERROR,
      CQRSErrorType.CONNECTION_ERROR,
      CQRSErrorType.BAD_GATEWAY,
      CQRSErrorType.SERVICE_UNAVAILABLE,
    ].includes(this.type)
  }

  // 获取建议的重试延迟（毫秒）
  public getRetryDelay(attempt: number): number {
    if (!this.retryable) return 0
    
    // 指数退避：基础延迟 * (2 ^ attempt) + 随机抖动
    const baseDelay = 1000 // 1秒
    const exponentialDelay = baseDelay * Math.pow(2, attempt - 1)
    const jitter = Math.random() * 1000 // 最多1秒的随机抖动
    
    return Math.min(exponentialDelay + jitter, 30000) // 最大30秒
  }

  // 序列化为日志格式
  public toLogFormat(): Record<string, any> {
    return {
      name: this.name,
      type: this.type,
      severity: this.severity,
      message: this.technicalMessage,
      userMessage: this.userMessage,
      retryable: this.retryable,
      context: this.context,
      timestamp: this.timestamp.toISOString(),
      requestId: this.requestId,
      stack: this.stack,
    }
  }
}

// 错误工厂函数
export class CQRSErrorFactory {
  static fromHttpResponse(
    response: Response, 
    context?: Record<string, any>,
    requestId?: string
  ): CQRSError {
    const status = response.status
    
    switch (true) {
      case status === 400:
        return new CQRSError(
          CQRSErrorType.BAD_REQUEST,
          `HTTP 400: ${response.statusText}`,
          undefined,
          ErrorSeverity.MEDIUM,
          false,
          { ...context, status, statusText: response.statusText },
          requestId
        )
      
      case status === 401:
        return new CQRSError(
          CQRSErrorType.UNAUTHORIZED,
          `HTTP 401: ${response.statusText}`,
          undefined,
          ErrorSeverity.HIGH,
          false,
          { ...context, status, statusText: response.statusText },
          requestId
        )
      
      case status === 403:
        return new CQRSError(
          CQRSErrorType.FORBIDDEN,
          `HTTP 403: ${response.statusText}`,
          undefined,
          ErrorSeverity.HIGH,
          false,
          { ...context, status, statusText: response.statusText },
          requestId
        )
      
      case status === 404:
        return new CQRSError(
          CQRSErrorType.NOT_FOUND,
          `HTTP 404: ${response.statusText}`,
          undefined,
          ErrorSeverity.LOW,
          false,
          { ...context, status, statusText: response.statusText },
          requestId
        )
      
      case status === 409:
        return new CQRSError(
          CQRSErrorType.CONFLICT,
          `HTTP 409: ${response.statusText}`,
          undefined,
          ErrorSeverity.MEDIUM,
          true,
          { ...context, status, statusText: response.statusText },
          requestId
        )
      
      case status >= 500 && status < 600:
        return new CQRSError(
          status === 502 ? CQRSErrorType.BAD_GATEWAY : 
          status === 503 ? CQRSErrorType.SERVICE_UNAVAILABLE : 
          CQRSErrorType.INTERNAL_SERVER_ERROR,
          `HTTP ${status}: ${response.statusText}`,
          undefined,
          ErrorSeverity.HIGH,
          true,
          { ...context, status, statusText: response.statusText },
          requestId
        )
      
      default:
        return new CQRSError(
          CQRSErrorType.UNKNOWN_ERROR,
          `HTTP ${status}: ${response.statusText}`,
          undefined,
          ErrorSeverity.MEDIUM,
          false,
          { ...context, status, statusText: response.statusText },
          requestId
        )
    }
  }

  static fromNetworkError(
    error: Error,
    context?: Record<string, any>,
    requestId?: string
  ): CQRSError {
    // 检查是否是网络相关错误
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
      return new CQRSError(
        CQRSErrorType.NETWORK_ERROR,
        `Network error: ${error.message}`,
        undefined,
        ErrorSeverity.HIGH,
        true,
        { ...context, originalError: error.message },
        requestId
      )
    }
    
    // 检查是否是超时错误
    if (error.name === 'AbortError' || error.message.includes('timeout')) {
      return new CQRSError(
        CQRSErrorType.TIMEOUT_ERROR,
        `Timeout error: ${error.message}`,
        undefined,
        ErrorSeverity.MEDIUM,
        true,
        { ...context, originalError: error.message },
        requestId
      )
    }
    
    return new CQRSError(
      CQRSErrorType.UNKNOWN_ERROR,
      error.message,
      undefined,
      ErrorSeverity.MEDIUM,
      false,
      { ...context, originalError: error.message },
      requestId
    )
  }

  static fromValidationError(
    message: string,
    validationErrors?: Array<{ field: string; message: string }>,
    context?: Record<string, any>,
    requestId?: string
  ): CQRSError {
    return new CQRSError(
      CQRSErrorType.VALIDATION_ERROR,
      message,
      '请检查输入信息并重试',
      ErrorSeverity.LOW,
      false,
      { ...context, validationErrors },
      requestId
    )
  }
}

// 重试配置
export interface RetryConfig {
  maxAttempts: number
  baseDelay: number
  maxDelay: number
  backoffMultiplier: number
  enableJitter: boolean
  retryableErrors: CQRSErrorType[]
}

// 默认重试配置
export const defaultRetryConfig: RetryConfig = {
  maxAttempts: 3,
  baseDelay: 1000,
  maxDelay: 30000,
  backoffMultiplier: 2,
  enableJitter: true,
  retryableErrors: [
    CQRSErrorType.NETWORK_ERROR,
    CQRSErrorType.TIMEOUT_ERROR,
    CQRSErrorType.CONNECTION_ERROR,
    CQRSErrorType.BAD_GATEWAY,
    CQRSErrorType.SERVICE_UNAVAILABLE,
    CQRSErrorType.CONFLICT,
  ],
}

// 智能重试机制
export class RetryManager {
  private config: RetryConfig

  constructor(config: Partial<RetryConfig> = {}) {
    this.config = { ...defaultRetryConfig, ...config }
  }

  // 执行带重试的异步操作
  async executeWithRetry<T>(
    operation: () => Promise<T>,
    context?: Record<string, any>
  ): Promise<T> {
    let lastError: CQRSError | null = null
    
    for (let attempt = 1; attempt <= this.config.maxAttempts; attempt++) {
      try {
        return await operation()
      } catch (error) {
        // 转换为CQRSError
        const cqrsError = error instanceof CQRSError 
          ? error 
          : error instanceof Error
            ? CQRSErrorFactory.fromNetworkError(error, context)
            : new CQRSError(CQRSErrorType.UNKNOWN_ERROR, String(error))
        
        lastError = cqrsError
        
        // 检查是否应该重试
        if (!this.shouldRetry(cqrsError, attempt)) {
          throw cqrsError
        }
        
        // 计算延迟时间
        const delay = this.calculateDelay(attempt)
        
        console.warn(`🔄 Retry attempt ${attempt}/${this.config.maxAttempts} for ${cqrsError.type} after ${delay}ms`, {
          error: cqrsError.toLogFormat(),
          attempt,
          delay,
        })
        
        // 等待延迟
        await this.sleep(delay)
      }
    }
    
    // 所有重试都失败了
    throw lastError!
  }

  private shouldRetry(error: CQRSError, attempt: number): boolean {
    // 检查是否超过最大重试次数
    if (attempt >= this.config.maxAttempts) {
      return false
    }
    
    // 检查错误类型是否可重试
    return this.config.retryableErrors.includes(error.type)
  }

  private calculateDelay(attempt: number): number {
    const exponentialDelay = this.config.baseDelay * Math.pow(this.config.backoffMultiplier, attempt - 1)
    
    let delay = Math.min(exponentialDelay, this.config.maxDelay)
    
    // 添加随机抖动
    if (this.config.enableJitter) {
      const jitter = Math.random() * 0.1 * delay // 10%的随机抖动
      delay += jitter
    }
    
    return Math.round(delay)
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }
}

// 错误监控和报告
export class ErrorReporter {
  private static instance: ErrorReporter
  private errorQueue: CQRSError[] = []
  private maxQueueSize = 100

  static getInstance(): ErrorReporter {
    if (!ErrorReporter.instance) {
      ErrorReporter.instance = new ErrorReporter()
    }
    return ErrorReporter.instance
  }

  // 报告错误
  report(error: CQRSError): void {
    // 添加到错误队列
    this.errorQueue.push(error)
    
    // 保持队列大小
    if (this.errorQueue.length > this.maxQueueSize) {
      this.errorQueue.shift()
    }
    
    // 根据严重程度决定处理方式
    switch (error.severity) {
      case ErrorSeverity.CRITICAL:
        this.handleCriticalError(error)
        break
      case ErrorSeverity.HIGH:
        this.handleHighSeverityError(error)
        break
      case ErrorSeverity.MEDIUM:
        this.handleMediumSeverityError(error)
        break
      case ErrorSeverity.LOW:
        this.handleLowSeverityError(error)
        break
    }
  }

  private handleCriticalError(error: CQRSError): void {
    console.error('🚨 CRITICAL ERROR:', error.toLogFormat())
    // 在实际应用中，这里应该发送到错误监控服务
    // 例如: Sentry.captureException(error)
  }

  private handleHighSeverityError(error: CQRSError): void {
    console.error('❌ HIGH SEVERITY ERROR:', error.toLogFormat())
    // 发送到错误监控服务
  }

  private handleMediumSeverityError(error: CQRSError): void {
    console.warn('⚠️ MEDIUM SEVERITY ERROR:', error.toLogFormat())
    // 发送到错误监控服务
  }

  private handleLowSeverityError(error: CQRSError): void {
    console.info('ℹ️ LOW SEVERITY ERROR:', error.toLogFormat())
    // 可选择性发送到错误监控服务
  }

  // 获取错误统计
  getErrorStats(): {
    total: number
    bySeverity: Record<ErrorSeverity, number>
    byType: Record<CQRSErrorType, number>
    recent: CQRSError[]
  } {
    const bySeverity = this.errorQueue.reduce((acc, error) => {
      acc[error.severity] = (acc[error.severity] || 0) + 1
      return acc
    }, {} as Record<ErrorSeverity, number>)

    const byType = this.errorQueue.reduce((acc, error) => {
      acc[error.type] = (acc[error.type] || 0) + 1
      return acc
    }, {} as Record<CQRSErrorType, number>)

    return {
      total: this.errorQueue.length,
      bySeverity,
      byType,
      recent: this.errorQueue.slice(-10), // 最近10个错误
    }
  }
}