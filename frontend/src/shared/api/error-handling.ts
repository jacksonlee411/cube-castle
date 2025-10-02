import { logger } from '@/shared/utils/logger';
import React from 'react';
import { isValidationError, isAPIError, isNetworkError } from './type-guards';
import { authManager } from './auth';
import { ORGANIZATION_API_ERRORS, getErrorMessage, formatErrorForUser, SUCCESS_MESSAGES } from './error-messages';
import type { JsonObject } from '../types/json';

type RawError = JsonObject | Error | null | undefined;

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

// API错误类实现
export class APIErrorImpl extends Error implements APIError {
  public status: number;
  public statusText: string;
  public response?: APIErrorResponse;

  constructor(status: number, statusText: string, response?: RawError) {
    super(`API Error: ${status} ${statusText}`);
    this.name = 'APIError';
    this.status = status;
    this.statusText = statusText;
    this.response = response as APIErrorResponse;
  }
}

// OAuth认证错误类
export class OAuthError extends Error {
  public readonly code: string;
  public readonly status?: number;
  
  constructor(message: string, code: string = 'OAUTH_ERROR', status?: number) {
    super(message);
    this.name = 'OAuthError';
    this.code = code;
    this.status = status;
  }
}

// 统一错误处理器
export class ErrorHandler {
  private static logError(context: string, error: RawError): void {
    const timestamp = new Date().toISOString();
    logger.group(`[${timestamp}] Error in ${context}`);
    logger.error('Error details:', error);
    
    if (isValidationError(error)) {
      logger.error('Validation details:', error.details);
    } else if (isAPIError(error)) {
      logger.error('API Error:', error.status, error.statusText);
      if (error.response) {
        logger.error('Response:', error.response);
      }
    } else if (isNetworkError(error)) {
      logger.error('Network Error:', error.message);
    }
    
    logger.groupEnd();
  }
  
  private static createUserMessage(error: RawError): string {
    if (isValidationError(error)) {
      return '数据验证失败，请检查输入的信息格式是否正确';
    } else if (isOAuthError(error)) {
      if (error.status === 401) {
        return 'OAuth认证失败，正在重新获取访问令牌...';
      }
      return 'OAuth认证错误，请联系系统管理员';
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
  
  // API调用错误处理 - 增强OAuth认证错误处理
  static async handleAPIError(context: string, error: RawError): Promise<never> {
    this.logError(context, error);
    
    if (isValidationError(error)) {
      throw new UserFriendlyError(
        this.createUserMessage(error),
        'VALIDATION_ERROR',
        error
      );
    } else if (isOAuthError(error)) {
      // 尝试清除无效的认证状态
      if (error.status === 401) {
        logger.info('[OAuth] Clearing invalid authentication state...');
        authManager.clearAuth();
      }
      
      throw new UserFriendlyError(
        this.createUserMessage(error),
        'OAUTH_ERROR',
        error
      );
    } else if (isAPIError(error)) {
      // 检查是否是OAuth相关的401错误
      if (error.status === 401) {
        logger.info('[Auth] API returned 401, clearing auth state...');
        authManager.clearAuth();
        
        throw new UserFriendlyError(
          'OAuth认证已过期，请刷新页面重新认证',
          'AUTH_EXPIRED',
          error
        );
      }
      
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
      logger.warn('Error reporting not implemented');
    }
  }
  
  // 表单验证错误处理
  static handleFormValidationError(error: RawError): FormValidationErrors {
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
  public readonly originalError: RawError;
  public readonly timestamp: Date;
  
  constructor(message: string, code: string, originalError: RawError) {
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
export const isUserFriendlyError = (error: RawError): error is UserFriendlyError => {
  return error instanceof UserFriendlyError;
};

// OAuth错误类型守卫
export const isOAuthError = (error: RawError): error is OAuthError => {
  return error instanceof OAuthError;
};

// 异步操作错误包装器 - 更新为支持异步错误处理
export const withErrorHandling = <T extends readonly RawError[], R>(
  fn: (...args: T) => Promise<R>,
  context: string,
) => {
  return async (...args: T): Promise<R> => {
    try {
      return await fn(...args);
    } catch (error) {
      await ErrorHandler.handleAPIError(context, error);
      throw error; // 这行永远不会执行，但TypeScript需要它
    }
  };
};

// React Hook 错误处理 - 更新为支持异步
export const useErrorHandler = () => {
  const handleError = (context: string) => async (error: RawError) => {
    await ErrorHandler.handleAPIError(context, error);
  };
  
  const handleFormError = (error: RawError): FormValidationErrors => {
    return ErrorHandler.handleFormValidationError(error);
  };
  
  return { handleError, handleFormError };
};

// 通用错误恢复策略
export const withRetry = <T extends readonly RawError[], R>(
  fn: (...args: T) => Promise<R>,
  maxRetries: number = 3,
  delay: number = 1000
) => {
  return async (...args: T): Promise<R> => {
    let lastError: RawError;
    
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

// OAuth感知的重试机制 - 专门处理认证过期
export const withOAuthRetry = <T extends readonly RawError[], R>(
  fn: (...args: T) => Promise<R>,
  maxRetries: number = 1
) => {
  return async (...args: T): Promise<R> => {
    let lastError: RawError;
    
    for (let i = 0; i <= maxRetries; i++) {
      try {
        return await fn(...args);
      } catch (error) {
        lastError = error;
        
        // 只对401错误重试，并清除认证状态
        if (isAPIError(error) && error.status === 401 && i < maxRetries) {
          logger.info('[OAuth] API returned 401, clearing auth and retrying...');
          authManager.clearAuth();
          
          // 等待一短暂时间让用户界面更新
          await new Promise(resolve => setTimeout(resolve, 500));
          continue;
        }
        
        // 其他错误或已达到最大重试次数，直接抛出
        throw error;
      }
    }
    
    throw lastError;
  };
};

// 统一的API客户端包装器 - 结合错误处理和OAuth重试
export const withOAuthAwareErrorHandling = <T extends readonly RawError[], R>(
  fn: (...args: T) => Promise<R>,
  context: string,
  enableRetry: boolean = true,
) => {
  const wrappedFn = enableRetry ? withOAuthRetry(fn) : fn;

  return async (...args: T): Promise<R> => {
    try {
      return await wrappedFn(...args);
    } catch (error) {
      await ErrorHandler.handleAPIError(context, error);
      throw error; // 这行永远不会执行，但TypeScript需要它
    }
  };
};

// 统一错误处理工具 - 整合error-messages.ts和simple-validation.ts
export const UnifiedErrorHandler = {
  // 替代formatErrorForUser
  formatForUser: formatErrorForUser,
  
  // 替代getErrorMessage
  getErrorInfo: getErrorMessage,
  
  // 获取标准错误码映射
  getErrorCodes: () => ORGANIZATION_API_ERRORS,
  
  // 获取成功消息
  getSuccessMessage: (key: keyof typeof SUCCESS_MESSAGES) => SUCCESS_MESSAGES[key],
  
  // 创建标准化的API错误
  createAPIError: (status: number, statusText: string, response?: RawError, errorCode?: string) => {
    const apiError = new APIErrorImpl(status, statusText, response);
    if (errorCode) {
      const errorInfo = getErrorMessage(errorCode);
      Object.assign(apiError, {
        errorCode,
        userMessage: errorInfo.userMessage,
        recoveryAction: errorInfo.recoveryAction
      });
    }
    return apiError;
  },
  
  // 创建标准化的验证错误  
  createValidationError: (message: string, fieldErrors: Array<{field: string, message: string}> = []) => {
    return new UserFriendlyError(message, 'VALIDATION_ERROR', { fieldErrors });
  },
  
  // 快速错误类型判断
  isAPIError: (error: RawError): error is APIError => error instanceof APIErrorImpl,
  isUserFriendlyError: (error: RawError): error is UserFriendlyError => error instanceof UserFriendlyError,
  isOAuthError: (error: RawError): error is OAuthError => error instanceof OAuthError,
  
  // 统一的错误日志记录
  logError: (context: string, error: RawError, additionalInfo?: JsonObject) => {
    const timestamp = new Date().toISOString();
    logger.group(`[${timestamp}] Error in ${context}`);
    logger.error('Error details:', error);
    
    if (additionalInfo) {
      logger.error('Additional info:', additionalInfo);
    }
    
    if (isUserFriendlyError(error)) {
      logger.error('User message:', error.message);
      logger.error('Error code:', error.code);
      logger.error('Original error:', error.originalError);
    }
    
    logger.groupEnd();
  }
};

// 向后兼容导出 - 用于逐步迁移
export { formatErrorForUser, getErrorMessage, SUCCESS_MESSAGES } from './error-messages';

// 重新导出其他错误处理系统的类型 - 统一接口
export type { ApiErrorCode, SuccessMessageKey } from './error-messages';
