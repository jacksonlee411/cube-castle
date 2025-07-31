// src/hooks/useRealtimeSync.ts
/**
 * Phase 2: 实时数据同步Hook
 * 企业级实时数据同步管理，集成WebSocket订阅与状态管理
 */

import { useEffect, useCallback, useRef } from 'react';
import { useAppActions, useRealtimeState } from '@/store';
import { apolloClient } from '@/lib/graphql-client';
import { useWebSocket } from './useWebSocket';

interface RealtimeSyncOptions {
  // 启用的订阅类型
  subscriptions?: ('employees' | 'organizations' | 'positions' | 'workflows')[];
  
  // 自动重连配置
  autoReconnect?: boolean;
  reconnectDelay?: number;
  maxReconnectAttempts?: number;
  
  // 数据同步配置
  enableOptimisticUpdates?: boolean;
  syncInterval?: number; // 定期同步间隔（毫秒）
  
  // 错误处理
  onError?: (error: Error) => void;
  onConnectionChange?: (connected: boolean) => void;
}

export const useRealtimeSync = (options: RealtimeSyncOptions = {}) => {
  const {
    subscriptions = ['employees', 'organizations', 'positions'],
    autoReconnect = true,
    reconnectDelay = 3000,
    maxReconnectAttempts = 5,
    enableOptimisticUpdates = true,
    syncInterval = 30000, // 30秒
    onError,
    onConnectionChange,
  } = options;

  const {
    setRealtimeConnection,
    setSubscription,
    updateLastUpdate,
    syncWithApollo,
    refreshApolloCache,
  } = useAppActions();

  const realtimeState = useRealtimeState();
  
  // WebSocket连接管理
  const { 
    connect, 
    disconnect, 
    subscribe, 
    unsubscribe, 
    isConnected 
  } = useWebSocket({
    onConnect: () => {
      setRealtimeConnection(true);
      onConnectionChange?.(true);
      
      // 连接成功后启用订阅
      subscriptions.forEach(type => {
        setSubscription(type, true);
      });
    },
    
    onDisconnect: () => {
      setRealtimeConnection(false);
      onConnectionChange?.(false);
      
      // 断线后禁用订阅
      subscriptions.forEach(type => {
        setSubscription(type, false);
      });
    },
    
    onError: (error) => {
      console.error('WebSocket error:', error);
      onError?.(error);
    },
  });

  // 重连尝试计数
  const reconnectAttempts = useRef(0);
  const reconnectTimer = useRef<NodeJS.Timeout>();
  const syncTimer = useRef<NodeJS.Timeout>();

  // 自动重连逻辑
  const attemptReconnect = useCallback(() => {
    if (!autoReconnect || reconnectAttempts.current >= maxReconnectAttempts) {
      return;
    }

    reconnectAttempts.current += 1;
    
    reconnectTimer.current = setTimeout(() => {
      console.log(`Attempting to reconnect (${reconnectAttempts.current}/${maxReconnectAttempts})`);
      connect();
    }, reconnectDelay * reconnectAttempts.current); // 指数退避
    
  }, [autoReconnect, maxReconnectAttempts, reconnectDelay, connect]);

  // 处理实时数据更新
  const handleDataUpdate = useCallback(async (type: string, data: any) => {
    try {
      // 更新最后更新时间
      updateLastUpdate();

      // 根据数据类型更新Apollo缓存
      switch (type) {
        case 'EMPLOYEE_UPDATED':
        case 'EMPLOYEE_CREATED':
        case 'EMPLOYEE_DELETED':
          await apolloClient.writeFragment({
            id: `Employee:${data.id}`,
            fragment: require('graphql-tag')`
              fragment UpdatedEmployee on Employee {
                id
                firstName
                lastName
                email
                status
                positionId
                organizationId
                updatedAt
              }
            `,
            data: type === 'EMPLOYEE_DELETED' ? null : data,
          });
          break;

        case 'ORGANIZATION_UPDATED':
        case 'ORGANIZATION_CREATED':
        case 'ORGANIZATION_DELETED':
          await apolloClient.writeFragment({
            id: `Organization:${data.id}`,
            fragment: require('graphql-tag')`
              fragment UpdatedOrganization on Organization {
                id
                name
                parentId
                level
                type
                managerId
                employeeCount
                updatedAt
              }
            `,
            data: type === 'ORGANIZATION_DELETED' ? null : data,
          });
          break;

        case 'POSITION_UPDATED':
        case 'POSITION_CREATED':
        case 'POSITION_DELETED':
          await apolloClient.writeFragment({
            id: `Position:${data.id}`,
            fragment: require('graphql-tag')`
              fragment UpdatedPosition on Position {
                id
                title
                organizationId
                jobLevel
                minSalary
                maxSalary
                currency
                isActive
                occupancyRate
                updatedAt
              }
            `,
            data: type === 'POSITION_DELETED' ? null : data,
          });
          break;

        default:
          console.warn('Unknown data update type:', type);
      }

      // 同步状态到Apollo
      await syncWithApollo();

    } catch (error) {
      console.error('Failed to handle data update:', error);
      onError?.(error as Error);
    }
  }, [updateLastUpdate, syncWithApollo, onError]);

  // 定期数据同步
  const performPeriodicSync = useCallback(async () => {
    if (!isConnected) return;

    try {
      // 刷新所有订阅的数据类型
      const cacheKeys = subscriptions.filter(type => 
        realtimeState.subscriptions[type]
      );
      
      if (cacheKeys.length > 0) {
        await refreshApolloCache(cacheKeys);
      }
      
    } catch (error) {
      console.error('Periodic sync failed:', error);
    }
  }, [isConnected, subscriptions, realtimeState.subscriptions, refreshApolloCache]);

  // 启动实时同步
  const startSync = useCallback(() => {
    // 连接WebSocket
    connect();

    // 设置定期同步
    if (syncInterval > 0) {
      syncTimer.current = setInterval(performPeriodicSync, syncInterval);
    }

    // 订阅实时数据更新
    subscriptions.forEach(type => {
      const subscriptionQuery = getSubscriptionQuery(type);
      if (subscriptionQuery) {
        subscribe(subscriptionQuery, (data) => {
          handleDataUpdate(type.toUpperCase() + '_UPDATED', data);
        });
      }
    });

  }, [connect, performPeriodicSync, syncInterval, subscriptions, subscribe, handleDataUpdate]);

  // 停止实时同步
  const stopSync = useCallback(() => {
    // 断开WebSocket
    disconnect();

    // 清理定时器
    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current);
    }
    if (syncTimer.current) {
      clearInterval(syncTimer.current);
    }

    // 取消所有订阅
    subscriptions.forEach(type => {
      unsubscribe(type);
      setSubscription(type, false);
    });

    // 重置重连计数
    reconnectAttempts.current = 0;

  }, [disconnect, subscriptions, unsubscribe, setSubscription]);

  // 手动触发数据同步
  const manualSync = useCallback(async (types?: string[]) => {
    const syncTypes = types || subscriptions;
    await refreshApolloCache(syncTypes);
  }, [subscriptions, refreshApolloCache]);

  // 组件挂载时启动，卸载时停止
  useEffect(() => {
    startSync();
    
    return () => {
      stopSync();
    };
  }, [startSync, stopSync]);

  // 连接状态变化时的重连逻辑
  useEffect(() => {
    if (!isConnected && realtimeState.connected) {
      // 从连接状态变为断开，尝试重连
      attemptReconnect();
    } else if (isConnected && !realtimeState.connected) {
      // 重连成功，重置计数器
      reconnectAttempts.current = 0;
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
      }
    }
  }, [isConnected, realtimeState.connected, attemptReconnect]);

  return {
    // 状态
    isConnected: realtimeState.connected,
    subscriptions: realtimeState.subscriptions,
    lastUpdate: realtimeState.lastUpdate,
    
    // 操作
    startSync,
    stopSync,
    manualSync,
    
    // 连接管理
    reconnectAttempts: reconnectAttempts.current,
    maxReconnectAttempts,
  };
};

// 获取订阅查询
function getSubscriptionQuery(type: string): string | null {
  switch (type) {
    case 'employees':
      return `
        subscription EmployeeUpdates {
          employeeUpdated {
            id
            firstName
            lastName
            email
            status
            positionId
            organizationId
            updatedAt
          }
        }
      `;
    
    case 'organizations':
      return `
        subscription OrganizationUpdates {
          organizationUpdated {
            id
            name
            parentId
            level
            type
            managerId
            employeeCount
            updatedAt
          }
        }
      `;
    
    case 'positions':
      return `
        subscription PositionUpdates {
          positionUpdated {
            id
            title
            organizationId
            jobLevel
            minSalary
            maxSalary
            currency
            isActive
            occupancyRate
            updatedAt
          }
        }
      `;
    
    case 'workflows':
      return `
        subscription WorkflowUpdates {
          workflowUpdated {
            id
            employeeId
            type
            status
            progress
            updatedAt
          }
        }
      `;
    
    default:
      return null;
  }
}

export default useRealtimeSync;