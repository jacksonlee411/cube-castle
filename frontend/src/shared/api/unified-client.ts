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
    const doRequest = async (): Promise<Response> => {
      // 🔧 开发和生产环境都需要JWT认证
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };
      
      // 所有环境都需要JWT认证
      const accessToken = await authManager.getAccessToken();
      if (accessToken) {
        headers['Authorization'] = `Bearer ${accessToken}`;
      }
      
      return fetch(this.endpoint, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          query,
          variables
        }),
      });
    };

    let retried = false;
    try {
      // 获取OAuth访问令牌
      let response = await doRequest();

      if (!response.ok) {
        // JWT token过期或无效时，清除认证状态并提供友好错误信息
        if (response.status === 401) {
          console.warn('[GraphQL Client] 401 未认证，尝试刷新令牌并重试一次');
          authManager.clearAuth();
          if (!retried) {
            retried = true;
            response = await doRequest();
            if (!response.ok) {
              throw new Error('认证已过期，请刷新页面重新登录');
            }
          } else {
            throw new Error('认证已过期，请刷新页面重新登录');
          }
        }
        
        // 服务器内部错误时提供更友好的错误信息
        if (response.status === 500) {
          console.error('[GraphQL Client] 服务器内部错误:', { query, variables, status: response.status });
          throw new Error('服务器内部错误，请稍后重试或联系管理员');
        }
        
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
    const url = `${this.baseURL}${endpoint}`;
    const doRequest = async (): Promise<Response> => {
      const accessToken = await authManager.getAccessToken();
      return fetch(url, {
        headers: {
          ...this.defaultHeaders,
          'Authorization': `Bearer ${accessToken}`,
          ...options.headers,
        },
        ...options,
      });
    };

    let retried = false;
    try {
      let response = await doRequest();

      // 读取文本与内容类型，按需解析JSON，避免非JSON错误体导致误导的解析错误
      const contentType = response.headers.get('content-type') || '';
      const text = await response.text();
      let result: Record<string, unknown> = {};
      if (text) {
        const looksLikeJson = contentType.includes('application/json') || /^(\s*[[{])/.test(text);
        if (looksLikeJson) {
          try {
            result = JSON.parse(text);
          } catch (parseError) {
            // 对于非OK状态，优先返回HTTP错误而不是解析错误，便于前端精确分流
            if (!response.ok) {
              console.error('[REST Client] 非JSON错误体，返回HTTP错误:', { endpoint, status: response.status, statusText: response.statusText });
              throw new Error(`REST Error: ${response.status} ${response.statusText}`);
            }
            console.error('[REST Client] JSON解析失败:', { endpoint, text, parseError });
            throw new Error(`响应解析失败: ${text.substring(0, 100)}${text.length > 100 ? '...' : ''}`);
          }
        }
      }

      if (!response.ok) {
        // JWT token过期或无效时，清除认证状态并提供友好错误信息
        if (response.status === 401) {
          console.warn('[REST Client] 401 未认证，尝试刷新令牌并重试一次');
          authManager.clearAuth();
          if (!retried) {
            retried = true;
            response = await doRequest();
          } else {
            throw new Error('认证已过期，请刷新页面重新登录');
          }
          // 重新读取响应体
          const contentTypeRetry = response.headers.get('content-type') || '';
          const textRetry = await response.text();
          let resultRetry: Record<string, unknown> = {};
          if (textRetry) {
            const looksLikeJsonRetry = contentTypeRetry.includes('application/json') || /^(\s*[[{])/.test(textRetry);
            if (looksLikeJsonRetry) {
              try { 
                resultRetry = JSON.parse(textRetry); 
              } catch (error) {
                console.warn('[REST Client] Failed to parse retry response as JSON:', error);
              }
            }
          }
          if (!response.ok) {
            throw new Error('认证已过期，请刷新页面重新登录');
          }
          return (resultRetry || {}) as T;
        }
        
        // 服务器内部错误时提供更友好的错误信息
        if (response.status === 500) {
          console.error('[REST Client] 服务器内部错误:', { endpoint, status: response.status, result });
          throw new Error('服务器内部错误，请稍后重试或联系管理员');
        }
        
        // 尝试解析服务器返回的错误信息
        if (result && typeof result === 'object' && 'error' in result) {
          const errorInfo = result.error as { message?: string };
          if (errorInfo && errorInfo.message) {
            throw new Error(errorInfo.message);
          }
        }
        
        // 如果没有具体错误信息，使用HTTP状态信息
        throw new Error(`REST Error: ${response.status} ${response.statusText}`);
      }

      // OK 情况下：有JSON返回即返回；无体则返回空对象
      return (result || {}) as T;
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
