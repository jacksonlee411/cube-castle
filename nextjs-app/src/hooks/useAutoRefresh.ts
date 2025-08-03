import { useEffect, useCallback, useRef } from 'react';

interface AutoRefreshOptions {
  interval?: number;           // 刷新间隔（毫秒），默认30秒
  enabled?: boolean;          // 是否启用自动刷新，默认true
  enableOnFocus?: boolean;    // 窗口获得焦点时刷新，默认true
  enableOnVisible?: boolean;  // 页面可见时刷新，默认true
}

/**
 * 自动刷新Hook - 替代WebSocket的轻量级解决方案
 * 提供定时刷新、窗口焦点刷新、页面可见性刷新
 */
export const useAutoRefresh = (
  refreshFn: () => void,
  options: AutoRefreshOptions = {}
) => {
  const {
    interval = 30000,        // 30秒默认间隔
    enabled = true,
    enableOnFocus = true,
    enableOnVisible = true,
  } = options;

  const refreshFnRef = useRef(refreshFn);
  const intervalRef = useRef<NodeJS.Timeout>();

  // 更新函数引用
  useEffect(() => {
    refreshFnRef.current = refreshFn;
  }, [refreshFn]);

  // 安全的刷新函数
  const safeRefresh = useCallback(() => {
    try {
      refreshFnRef.current();
    } catch (error) {
      console.error('Auto refresh failed:', error);
    }
  }, []);

  // 定时刷新
  useEffect(() => {
    if (!enabled) return;

    intervalRef.current = setInterval(safeRefresh, interval);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [enabled, interval, safeRefresh]);

  // 窗口焦点刷新
  useEffect(() => {
    if (!enabled || !enableOnFocus) return;

    const handleFocus = () => safeRefresh();
    
    window.addEventListener('focus', handleFocus);
    return () => window.removeEventListener('focus', handleFocus);
  }, [enabled, enableOnFocus, safeRefresh]);

  // 页面可见性刷新
  useEffect(() => {
    if (!enabled || !enableOnVisible) return;

    const handleVisibilityChange = () => {
      if (!document.hidden) {
        safeRefresh();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, [enabled, enableOnVisible, safeRefresh]);

  // 手动刷新函数
  const manualRefresh = useCallback(() => {
    safeRefresh();
  }, [safeRefresh]);

  return {
    manualRefresh,
    isEnabled: enabled,
  };
};

export default useAutoRefresh;