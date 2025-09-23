/**
 * 标准化API错误消息 (ADR-008合规)
 * 统一前端错误处理和用户提示
 */

export interface ApiErrorCode {
  code: string;
  userMessage: string;
  technicalMessage: string;
  recoveryAction?: string;
}

// 启用/停用操作标准错误码
export const ORGANIZATION_API_ERRORS: Record<string, ApiErrorCode> = {
  // 启用相关错误
  ORGANIZATION_ALREADY_ACTIVE: {
    code: 'ORGANIZATION_ALREADY_ACTIVE',
    userMessage: '该组织已处于启用状态',
    technicalMessage: 'Organization is already in active status',
    recoveryAction: '刷新页面获取最新状态'
  },

  ORGANIZATION_NOT_FOUND: {
    code: 'ORGANIZATION_NOT_FOUND', 
    userMessage: '未找到指定的组织',
    technicalMessage: 'Organization with the specified code does not exist',
    recoveryAction: '检查组织代码是否正确'
  },

  // 停用相关错误
  ORGANIZATION_ALREADY_SUSPENDED: {
    code: 'ORGANIZATION_ALREADY_SUSPENDED',
    userMessage: '该组织已处于停用状态',
    technicalMessage: 'Organization is already in suspended status',
    recoveryAction: '刷新页面获取最新状态'
  },

  ORGANIZATION_HAS_ACTIVE_CHILDREN: {
    code: 'ORGANIZATION_HAS_ACTIVE_CHILDREN',
    userMessage: '该组织下有子组织仍处于启用状态，无法停用',
    technicalMessage: 'Cannot suspend organization with active child organizations',
    recoveryAction: '请先停用所有子组织'
  },

  // 权限相关错误
  INSUFFICIENT_PERMISSIONS: {
    code: 'INSUFFICIENT_PERMISSIONS',
    userMessage: '权限不足，无法执行此操作',
    technicalMessage: 'User lacks required permissions for this operation',
    recoveryAction: '联系管理员获取相应权限'
  },

  INVALID_PERMISSION_SCOPE: {
    code: 'INVALID_PERMISSION_SCOPE',
    userMessage: '权限范围无效',
    technicalMessage: 'Required permission scope is missing or invalid',
    recoveryAction: '检查token是否包含org:activate或org:suspend权限'
  },

  // 验证相关错误
  VALIDATION_ERROR: {
    code: 'VALIDATION_ERROR',
    userMessage: '请求数据验证失败',
    technicalMessage: 'Request data validation failed',
    recoveryAction: '检查请求参数格式和必填字段'
  },

  INVALID_EFFECTIVE_DATE: {
    code: 'INVALID_EFFECTIVE_DATE',
    userMessage: '生效日期格式不正确',
    technicalMessage: 'Effective date must be in YYYY-MM-DD format',
    recoveryAction: '使用YYYY-MM-DD格式的日期'
  },

  OPERATION_REASON_REQUIRED: {
    code: 'OPERATION_REASON_REQUIRED',
    userMessage: '操作原因已改为可选项',
    technicalMessage: 'operationReason field is optional',
    recoveryAction: '如需填写，请输入5-500个字符；否则可以留空'
  },

  // 弃用端点错误  
  ENDPOINT_DEPRECATED: {
    code: 'ENDPOINT_DEPRECATED',
    userMessage: '您使用的功能已更新，系统将自动重定向',
    technicalMessage: 'API endpoint has been deprecated, use successor version',
    recoveryAction: '系统会自动使用新的接口，无需手动操作'
  },

  // 系统相关错误
  INTERNAL_SERVER_ERROR: {
    code: 'INTERNAL_SERVER_ERROR',
    userMessage: '系统内部错误，请稍后重试',
    technicalMessage: 'Internal server error occurred',
    recoveryAction: '稍后重试，如问题持续请联系技术支持'
  },

  AUDIT_WRITE_FAILURE: {
    code: 'AUDIT_WRITE_FAILURE',
    userMessage: '操作记录保存失败',
    technicalMessage: 'Failed to write audit log',
    recoveryAction: '联系技术支持，确保操作已正确执行'
  },

  NETWORK_ERROR: {
    code: 'NETWORK_ERROR',
    userMessage: '网络连接异常，请检查网络后重试',
    technicalMessage: 'Network connection failed',
    recoveryAction: '检查网络连接，刷新页面重试'
  }
};

/**
 * 获取用户友好的错误消息
 * @param errorCode API错误码
 * @returns 格式化的错误信息对象
 */
export function getErrorMessage(errorCode: string): ApiErrorCode {
  return ORGANIZATION_API_ERRORS[errorCode] || {
    code: 'UNKNOWN_ERROR',
    userMessage: '发生未知错误，请稍后重试',
    technicalMessage: `Unknown error code: ${errorCode}`,
    recoveryAction: '刷新页面重试，如问题持续请联系技术支持'
  };
}

/**
 * 标准化错误处理工具函数
 * @param error 捕获的错误对象
 * @returns 标准化的用户提示消息
 */
export function formatErrorForUser(error: unknown): string {
  if (error && typeof error === 'object') {
    const errorObj = error as Record<string, unknown>;
    
    // API响应错误
    if (errorObj.error && errorObj.error.code) {
      const errorInfo = getErrorMessage(errorObj.error.code);
      return errorInfo.userMessage;
    }
    
    // 简单错误消息
    if (errorObj.message) {
      return errorObj.message;
    }
  }
  
  // 兜底错误消息
  return '操作失败，请稍后重试';
}

/**
 * 操作成功消息标准化
 */
export const SUCCESS_MESSAGES = {
  ACTIVATE_SUCCESS: '组织启用成功',
  SUSPEND_SUCCESS: '组织停用成功',
  CREATE_SUCCESS: '组织创建成功',
  UPDATE_SUCCESS: '组织信息更新成功',
  DELETE_SUCCESS: '组织删除成功'
} as const;

export type SuccessMessageKey = keyof typeof SUCCESS_MESSAGES;
