/**
 * 统一的消息管理Hook
 * 替代alert()调用，提供企业级用户体验
 * 基于Canvas Kit v13设计标准
 */
import { useState, useCallback, useEffect, useRef } from 'react';

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
  const successTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const errorTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearSuccessTimer = useCallback(() => {
    if (successTimerRef.current) {
      clearTimeout(successTimerRef.current);
      successTimerRef.current = null;
    }
  }, []);

  const clearErrorTimer = useCallback(() => {
    if (errorTimerRef.current) {
      clearTimeout(errorTimerRef.current);
      errorTimerRef.current = null;
    }
  }, []);

  const clearAllTimers = useCallback(() => {
    clearSuccessTimer();
    clearErrorTimer();
  }, [clearErrorTimer, clearSuccessTimer]);

  useEffect(() => () => {
    clearAllTimers();
  }, [clearAllTimers]);

  // 显示成功消息
  const showSuccess = useCallback((message: string) => {
    setError(null); // 清除错误消息
    setSuccessMessage(message);
    clearSuccessTimer();
    successTimerRef.current = setTimeout(() => {
      setSuccessMessage(null);
      successTimerRef.current = null;
    }, 3000);
  }, [clearSuccessTimer]);

  // 显示错误消息
  const showError = useCallback((message: string) => {
    setSuccessMessage(null); // 清除成功消息
    setError(message);
    clearErrorTimer();
    errorTimerRef.current = setTimeout(() => {
      setError(null);
      errorTimerRef.current = null;
    }, 5000);
  }, [clearErrorTimer]);

  // 手动清除所有消息
  const clearMessages = useCallback(() => {
    setSuccessMessage(null);
    setError(null);
    clearAllTimers();
  }, [clearAllTimers]);

  return {
    successMessage,
    error,
    showSuccess,
    showError,
    clearMessages
  };
};

export default useMessages;
