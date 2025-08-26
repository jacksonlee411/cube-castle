/**
 * 统一的消息管理Hook
 * 替代alert()调用，提供企业级用户体验
 * 基于Canvas Kit v13设计标准
 */
import { useState, useCallback } from 'react';

export interface MessageState {
  successMessage: string | null;
  error: string | null;
}

export interface MessageActions {
  showSuccess: (message: string) => void;
  showError: (message: string) => void;
  clearMessages: () => void;
}

export type UseMessagesReturn = MessageState & MessageActions;

/**
 * 统一消息管理Hook
 * 提供成功/错误消息管理，自动清理机制
 */
export const useMessages = (): UseMessagesReturn => {
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  // 显示成功消息
  const showSuccess = useCallback((message: string) => {
    setError(null); // 清除错误消息
    setSuccessMessage(message);
    // 3秒后自动清除成功消息
    setTimeout(() => setSuccessMessage(null), 3000);
  }, []);

  // 显示错误消息
  const showError = useCallback((message: string) => {
    setSuccessMessage(null); // 清除成功消息
    setError(message);
    // 5秒后自动清除错误消息
    setTimeout(() => setError(null), 5000);
  }, []);

  // 手动清除所有消息
  const clearMessages = useCallback(() => {
    setSuccessMessage(null);
    setError(null);
  }, []);

  return {
    successMessage,
    error,
    showSuccess,
    showError,
    clearMessages
  };
};

export default useMessages;