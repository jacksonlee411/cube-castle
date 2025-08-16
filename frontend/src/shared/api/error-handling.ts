import React from 'react';
import { isValidationError, isAPIError, isNetworkError } from './type-guards';

// API错误接口
interface APIErrorResponse {
  status?: number;
  statusText?: string;
}

export interface APIError extends Error {
  status: number;
  statusText: string;
  response?: APIErrorResponse;
}

// 统一错误处理器
export class ErrorHandler {
  private static logError(context: string, error: unknown): void {
    const timestamp = new Date().toISOString();
    console.group(`[${timestamp}] Error in ${context}`);
    console.error('Error details:', error);
    
    if (isValidationError(error)) {
      console.error('Validation details:', error.details);
    } else if (isAPIError(error)) {
      console.error('API Error:', error.status, error.statusText);
      if (error.response) {
        console.error('Response:', error.response);
      }
    } else if (isNetworkError(error)) {
      console.error('Network Error:', error.message);
    }
    
    console.groupEnd();
  }
  
  private static createUserMessage(error: unknown): string {
    if (isValidationError(error)) {
      return '数据验证失败，请检查输入的信息格式是否正确';
    } else if (isAPIError(error)) {
      if (error.status >= 500) {
        return '服务器内部错误，请稍后重试';
      } else if (error.status === 404) {
        return '请求的资源不存在';
      } else if (error.status === 403) {
        return '没有权限执行此操作';
      } else if (error.status === 401) {
        return '身份验证失败，请重新登录';
      } else if (error.status >= 400) {
        return '请求参数有误，请检查输入';
      }
    } else if (isNetworkError(error)) {
      return '网络连接失败，请检查网络连接';
    }
    
    return '发生未知错误，请稍后重试';
  }
  
  // API调用错误处理
  static handleAPIError(context: string, error: unknown): never {
    this.logError(context, error);
    
    if (isValidationError(error)) {
      throw new UserFriendlyError(
        this.createUserMessage(error),
        'VALIDATION_ERROR',
        error
      );
    } else if (isAPIError(error)) {
      throw new UserFriendlyError(
        this.createUserMessage(error),
        'API_ERROR',
        error
      );
    } else if (isNetworkError(error)) {
      throw new UserFriendlyError(
        this.createUserMessage(error),
        'NETWORK_ERROR',
        error
      );
    } else {
      throw new UserFriendlyError(
        this.createUserMessage(error),
        'UNKNOWN_ERROR',
        error
      );
    }
  }
  
  // 组件错误边界处理
  static handleComponentError(context: string, error: Error, _errorInfo?: React.ErrorInfo): void {
    this.logError(context, error);
    
    // 发送错误报告到监控服务（如果需要）
    if (process.env.NODE_ENV === 'production') {
      // 这里可以集成错误监控服务，如 Sentry
      console.warn('Error reporting not implemented');
    }
  }
  
  // 表单验证错误处理
  static handleFormValidationError(error: unknown): FormValidationErrors {
    if (!isValidationError(error)) {
      return { _form: ['表单验证失败'] };
    }
    
    const formErrors: FormValidationErrors = {};
    
    error.details.forEach(detail => {
      const field = detail.path.join('.');
      if (!formErrors[field]) {
        formErrors[field] = [];
      }
      formErrors[field].push(detail.message);
    });
    
    return formErrors;
  }
}

// 用户友好的错误类
export class UserFriendlyError extends Error {
  public readonly code: string;
  public readonly originalError: unknown;
  public readonly timestamp: Date;
  
  constructor(message: string, code: string, originalError: unknown) {
    super(message);
    this.name = 'UserFriendlyError';
    this.code = code;
    this.originalError = originalError;
    this.timestamp = new Date();
  }
}

// 表单验证错误类型
export interface FormValidationErrors {
  [field: string]: string[];
}

// 错误类型守卫
export const isUserFriendlyError = (error: unknown): error is UserFriendlyError => {
  return error instanceof UserFriendlyError;
};

// 异步操作错误包装器
export const withErrorHandling = <T extends unknown[], R>(
  fn: (...args: T) => Promise<R>,
  context: string
) => {
  return async (...args: T): Promise<R> => {
    try {
      return await fn(...args);
    } catch (error) {
      ErrorHandler.handleAPIError(context, error);
    }
  };
};

// React Hook 错误处理
export const useErrorHandler = () => {
  const handleError = (context: string) => (error: unknown) => {
    ErrorHandler.handleAPIError(context, error);
  };
  
  const handleFormError = (error: unknown): FormValidationErrors => {
    return ErrorHandler.handleFormValidationError(error);
  };
  
  return { handleError, handleFormError };
};

// 通用错误恢复策略
export const withRetry = <T extends unknown[], R>(
  fn: (...args: T) => Promise<R>,
  maxRetries: number = 3,
  delay: number = 1000
) => {
  return async (...args: T): Promise<R> => {
    let lastError: unknown;
    
    for (let i = 0; i <= maxRetries; i++) {
      try {
        return await fn(...args);
      } catch (error) {
        lastError = error;
        
        // 不重试验证错误和4xx错误
        if (isValidationError(error) || (isAPIError(error) && error.status < 500)) {
          throw error;
        }
        
        if (i < maxRetries) {
          await new Promise(resolve => setTimeout(resolve, delay * Math.pow(2, i)));
        }
      }
    }
    
    throw lastError;
  };
};