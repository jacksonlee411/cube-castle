/**
 * GraphQL企业级响应信封适配器
 * 适配后端即将实施的GraphQL企业级响应信封格式
 * 基于API契约v4.2.1和后端团队P0级优先任务
 */

import type { APIResponse } from '../types/api';
import { UnifiedGraphQLClient } from './unified-client';
import { authManager } from './auth';

// 企业级GraphQL响应信封接口
interface EnterpriseGraphQLResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: unknown;
  };
  message?: string;
  timestamp: string;
  requestId?: string;
}

// 传统GraphQL响应接口（当前格式）
interface StandardGraphQLResponse<T> {
  data?: T;
  errors?: Array<{
    message: string;
    locations?: Array<{ line: number; column: number }>;
    path?: Array<string | number>;
    extensions?: Record<string, unknown>;
  }>;
}

/**
 * GraphQL企业级响应适配器类
 * 处理标准GraphQL格式到企业级信封格式的转换
 */
export class GraphQLEnterpriseAdapter {
  private client: UnifiedGraphQLClient;

  constructor(client: UnifiedGraphQLClient) {
    this.client = client;
  }

  /**
   * 获取访问令牌
   */
  private async getAccessToken(): Promise<string> {
    try {
      return await authManager.getAccessToken();
    } catch (error) {
      console.warn('Failed to get access token, using empty token:', error);
      return '';
    }
  }

  /**
   * 检测响应格式类型
   */
  private isEnterpriseFormat<T>(response: unknown): response is EnterpriseGraphQLResponse<T> {
    return (
      typeof response === 'object' &&
      response !== null &&
      'success' in response &&
      'timestamp' in response
    );
  }

  /**
   * 检测标准GraphQL格式
   */
  private isStandardFormat<T>(response: unknown): response is StandardGraphQLResponse<T> {
    return (
      typeof response === 'object' &&
      response !== null &&
      ('data' in response || 'errors' in response) &&
      !('success' in response)
    );
  }

  /**
   * 将标准GraphQL响应转换为企业级信封格式
   */
  private transformToEnterpriseFormat<T>(
    standardResponse: StandardGraphQLResponse<T>
  ): EnterpriseGraphQLResponse<T> {
    const timestamp = new Date().toISOString();
    const requestId = `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

    // 如果有错误，返回错误格式
    if (standardResponse.errors && standardResponse.errors.length > 0) {
      const firstError = standardResponse.errors[0];
      return {
        success: false,
        error: {
          code: 'GRAPHQL_ERROR',
          message: firstError.message,
          details: {
            locations: firstError.locations,
            path: firstError.path,
            extensions: firstError.extensions
          }
        },
        message: `GraphQL查询失败: ${firstError.message}`,
        timestamp,
        requestId
      };
    }

    // 成功响应
    return {
      success: true,
      data: standardResponse.data,
      message: 'GraphQL查询成功',
      timestamp,
      requestId
    };
  }

  /**
   * 统一的GraphQL请求方法，自动适配响应格式
   */
  async request<T>(
    query: string, 
    variables?: Record<string, unknown>
  ): Promise<APIResponse<T>> {
    try {
      // 使用UnifiedGraphQLClient的完整功能，包括认证
      // 但需要处理其可能抛出的异常并转换为企业级格式
      try {
        // 先尝试使用客户端的原生请求方法
        const result = await this.client.request<T>(query, variables);
        
        // 如果成功，将结果包装为企业级格式
        return {
          success: true,
          data: result,
          message: 'GraphQL查询成功',
          timestamp: new Date().toISOString(),
          requestId: `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
        };
      } catch (clientError) {
        // UnifiedGraphQLClient 抛出了异常，可能是因为响应格式已经是企业级
        // 尝试直接发送请求并检查响应格式
        const rawResponse = await fetch(this.client['endpoint'] || 'http://localhost:8090/graphql', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${await this.getAccessToken()}`,
          },
          body: JSON.stringify({ query, variables })
        });

        if (!rawResponse.ok) {
          throw new Error(`HTTP Error: ${rawResponse.status} ${rawResponse.statusText}`);
        }

        const responseData = await rawResponse.json();

        // 检查响应格式并适配
        if (this.isEnterpriseFormat<T>(responseData)) {
          // 已经是企业级格式，直接返回
          return responseData;
        } else if (this.isStandardFormat<T>(responseData)) {
          // 标准GraphQL格式，需要转换
          return this.transformToEnterpriseFormat<T>(responseData);
        } else {
          // 未知格式，返回错误
          return {
            success: false,
            error: {
              code: 'UNKNOWN_RESPONSE_FORMAT',
              message: '未知的GraphQL响应格式',
              details: responseData
            },
            timestamp: new Date().toISOString(),
            requestId: `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
          };
        }
      }
    } catch (error) {
      // 网络或其他错误
      return {
        success: false,
        error: {
          code: 'NETWORK_ERROR',
          message: error instanceof Error ? error.message : '网络请求失败',
          details: error
        },
        timestamp: new Date().toISOString(),
        requestId: `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
      };
    }
  }

  /**
   * 批量GraphQL请求支持
   */
  async batchRequest<T>(
    requests: Array<{
      query: string;
      variables?: Record<string, unknown>;
      operationName?: string;
    }>
  ): Promise<APIResponse<T[]>> {
    try {
      const results = await Promise.all(
        requests.map(req => this.request<T>(req.query, req.variables))
      );

      // 检查是否有任何请求失败
      const failedRequests = results.filter(result => !result.success);
      
      if (failedRequests.length > 0) {
        return {
          success: false,
          error: {
            code: 'BATCH_REQUEST_PARTIAL_FAILURE',
            message: `批量请求中有 ${failedRequests.length} 个失败`,
            details: failedRequests
          },
          timestamp: new Date().toISOString(),
          requestId: `batch_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
        };
      }

      // 所有请求成功，提取数据
      const data = results.map(result => result.data).filter((data): data is T => data !== undefined);

      return {
        success: true,
        data,
        message: `批量请求成功完成 ${results.length} 个操作`,
        timestamp: new Date().toISOString(),
        requestId: `batch_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
      };
    } catch (error) {
      return {
        success: false,
        error: {
          code: 'BATCH_REQUEST_ERROR',
          message: error instanceof Error ? error.message : '批量请求失败',
          details: error
        },
        timestamp: new Date().toISOString(),
        requestId: `batch_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
      };
    }
  }
}

// 创建全局适配器实例
export const graphqlEnterpriseAdapter = new GraphQLEnterpriseAdapter(
  new UnifiedGraphQLClient()
);

// 企业级GraphQL客户端Hook
export const useEnterpriseGraphQL = () => {
  return {
    request: graphqlEnterpriseAdapter.request.bind(graphqlEnterpriseAdapter),
    batchRequest: graphqlEnterpriseAdapter.batchRequest.bind(graphqlEnterpriseAdapter)
  };
};