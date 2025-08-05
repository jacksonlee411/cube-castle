// API错误处理和响应规范
// 标准化错误响应格式和处理机制

export interface APIError {
  code: string;
  message: string;
  details?: string;
  traceId?: string;
  timestamp: string;
}

export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: APIError;
  meta?: {
    total?: number;
    page?: number;
    limit?: number;
  };
}

// 标准错误码
export enum ErrorCodes {
  // 通用错误
  INTERNAL_SERVER_ERROR = 'INTERNAL_SERVER_ERROR',
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  UNAUTHORIZED = 'UNAUTHORIZED',
  FORBIDDEN = 'FORBIDDEN',
  
  // 员工相关错误
  EMPLOYEE_NOT_FOUND = 'EMPLOYEE_NOT_FOUND',
  EMPLOYEE_UPDATE_FAILED = 'EMPLOYEE_UPDATE_FAILED',
  EMPLOYEE_DELETE_FAILED = 'EMPLOYEE_DELETE_FAILED',
  EMPLOYEE_ALREADY_EXISTS = 'EMPLOYEE_ALREADY_EXISTS',
  
  // 部门相关错误
  DEPARTMENT_NOT_FOUND = 'DEPARTMENT_NOT_FOUND',
  INVALID_DEPARTMENT_ASSIGNMENT = 'INVALID_DEPARTMENT_ASSIGNMENT',
  
  // 职位相关错误
  POSITION_NOT_FOUND = 'POSITION_NOT_FOUND',
  INVALID_POSITION_ASSIGNMENT = 'INVALID_POSITION_ASSIGNMENT',
}

// 错误处理工具类
export class APIErrorHandler {
  static handle(error: any): APIError {
    const traceId = this.generateTraceId();
    
    if (error.response) {
      // HTTP响应错误
      const status = error.response.status;
      const data = error.response.data;
      
      switch (status) {
        case 400:
          return {
            code: ErrorCodes.VALIDATION_ERROR,
            message: data?.message || '请求参数错误',
            details: data?.details,
            traceId,
            timestamp: new Date().toISOString(),
          };
        case 401:
          return {
            code: ErrorCodes.UNAUTHORIZED,
            message: '用户未授权',
            traceId,
            timestamp: new Date().toISOString(),
          };
        case 403:
          return {
            code: ErrorCodes.FORBIDDEN,
            message: '无权限访问',
            traceId,
            timestamp: new Date().toISOString(),
          };
        case 404:
          return {
            code: data?.code || ErrorCodes.EMPLOYEE_NOT_FOUND,
            message: data?.message || '资源不存在',
            details: data?.details,
            traceId,
            timestamp: new Date().toISOString(),
          };
        case 500:
          return {
            code: ErrorCodes.INTERNAL_SERVER_ERROR,
            message: '服务器内部错误',
            details: data?.details,
            traceId,
            timestamp: new Date().toISOString(),
          };
        default:
          return {
            code: ErrorCodes.INTERNAL_SERVER_ERROR,
            message: '未知错误',
            details: error.message,
            traceId,
            timestamp: new Date().toISOString(),
          };
      }
    } else if (error.request) {
      // 网络错误
      return {
        code: ErrorCodes.INTERNAL_SERVER_ERROR,
        message: '网络连接错误',
        details: '请检查网络连接',
        traceId,
        timestamp: new Date().toISOString(),
      };
    } else {
      // 其他错误
      return {
        code: ErrorCodes.INTERNAL_SERVER_ERROR,
        message: '未知错误',
        details: error.message,
        traceId,
        timestamp: new Date().toISOString(),
      };
    }
  }
  
  private static generateTraceId(): string {
    return Math.random().toString(36).substring(2, 15) + 
           Math.random().toString(36).substring(2, 15);
  }
}