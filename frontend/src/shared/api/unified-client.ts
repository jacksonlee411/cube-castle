/**
 * 统一的API客户端配置
 * 消除重复的GraphQL和REST客户端实现
 * 基于CQRS架构：查询使用GraphQL，命令使用REST API
 */
import { authManager } from './auth';
import type { GraphQLResponse } from '../types';

// 🔧 CQRS架构端点配置 - 使用代理避免CORS问题
const API_ENDPOINTS = {
  GRAPHQL_QUERY: '/graphql',     // 查询服务 (PostgreSQL GraphQL) - 通过Vite代理
  REST_COMMAND: '/api/v1'        // 命令服务 (REST API) - 通过Vite代理
} as const;

/**
 * 统一的GraphQL客户端 - 专用于查询操作
 * 遵循CQRS原则：所有查询统一使用GraphQL
 */
export class UnifiedGraphQLClient {
  private endpoint: string;

  constructor(endpoint: string = API_ENDPOINTS.GRAPHQL_QUERY) {
    this.endpoint = endpoint;
  }

  async request<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
    try {
      // 获取OAuth访问令牌
      const accessToken = await authManager.getAccessToken();
      
      const response = await fetch(this.endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${accessToken}`,
        },
        body: JSON.stringify({
          query,
          variables
        }),
      });

      if (!response.ok) {
        throw new Error(`GraphQL Error: ${response.status} ${response.statusText}`);
      }

      const result = await response.json() as GraphQLResponse<T>;
      
      if (result.errors && result.errors.length > 0) {
        throw new Error(`GraphQL Error: ${result.errors[0].message}`);
      }

      if (!result.data) {
        throw new Error('GraphQL Error: No data returned');
      }

      return result.data;
    } catch (error) {
      console.error('GraphQL request failed:', { query, variables, error });
      throw error;
    }
  }
}

/**
 * 统一的REST API客户端 - 专用于命令操作
 * 遵循CQRS原则：所有命令统一使用REST API
 */
export class UnifiedRESTClient {
  private baseURL: string;
  private defaultHeaders: Record<string, string>;

  constructor(baseURL: string = API_ENDPOINTS.REST_COMMAND) {
    this.baseURL = baseURL;
    this.defaultHeaders = {
      'Content-Type': 'application/json',
      'X-Tenant-ID': '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', // 默认租户ID
    };
  }

  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    try {
      // 获取OAuth访问令牌
      const accessToken = await authManager.getAccessToken();
      
      const url = `${this.baseURL}${endpoint}`;
      
      const response = await fetch(url, {
        headers: {
          ...this.defaultHeaders,
          'Authorization': `Bearer ${accessToken}`,
          ...options.headers,
        },
        ...options,
      });

      if (!response.ok) {
        throw new Error(`REST Error: ${response.status} ${response.statusText}`);
      }

      // 检查是否有响应体（DELETE请求可能没有响应体）
      const text = await response.text();
      return text ? JSON.parse(text) : ({} as T);
    } catch (error) {
      console.error('REST request failed:', { endpoint, options, error });
      throw error;
    }
  }
}

// 🔧 单例实例 - 全局使用统一客户端
export const unifiedGraphQLClient = new UnifiedGraphQLClient();
export const unifiedRESTClient = new UnifiedRESTClient();

// 📋 客户端工厂方法 - 支持自定义配置
export const createGraphQLClient = (endpoint?: string) => new UnifiedGraphQLClient(endpoint);
export const createRESTClient = (baseURL?: string) => new UnifiedRESTClient(baseURL);

// 🔧 架构原则检查器 - 开发模式下验证正确使用
export const validateCQRSUsage = (operation: 'query' | 'command', method: string) => {
  if (process.env.NODE_ENV === 'development') {
    if (operation === 'query' && !method.includes('GraphQL')) {
      console.warn('⚠️ CQRS违反: 查询操作应该使用GraphQL客户端');
    }
    if (operation === 'command' && !method.includes('REST')) {
      console.warn('⚠️ CQRS违反: 命令操作应该使用REST客户端');
    }
  }
};

export default {
  graphql: unifiedGraphQLClient,
  rest: unifiedRESTClient,
  validateCQRSUsage
};