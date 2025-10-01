/**
 * GraphQL企业级响应信封适配器
 * 适配后端即将实施的GraphQL企业级响应信封格式
 * 基于API契约v4.2.1和后端团队P0级优先任务
 */

import type { APIResponse } from '../types/api';
import { UnifiedGraphQLClient } from './unified-client';
import type { JsonValue } from '../types/json';
// import { authManager } from './auth'; // 暂时移除未使用的import

/*
 * 企业级GraphQL响应接口定义 - 预留用于未来后端实施
 * interface EnterpriseGraphQLResponse<T> {
 *   success: boolean; data?: T; error?: {...}; message?: string;
 *   timestamp: string; requestId?: string;
 * }
 * 
 * 传统GraphQL响应接口 - 预留用于格式转换
 * interface StandardGraphQLResponse<T> {
 *   data?: T; errors?: Array<{...}>; 
 * }
 */

/**
 * GraphQL企业级响应适配器类
 * 处理标准GraphQL格式到企业级信封格式的转换
 */
export class GraphQLEnterpriseAdapter {
  private client: UnifiedGraphQLClient;

  constructor(client: UnifiedGraphQLClient) {
    this.client = client;
  }

  // 获取访问令牌方法已移动到UnifiedGraphQLClient

  // 响应格式检测方法 - 预留用于未来企业级格式支持

  // 标准GraphQL格式检测方法 - 预留用于格式适配

  // 企业级格式转换方法 - 预留用于未来后端企业级信封实施

  /**
   * 统一的GraphQL请求方法，自动适配响应格式
   */
  async request<T>(
    query: string, 
    variables?: Record<string, JsonValue>
  ): Promise<APIResponse<T>> {
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
    } catch (error) {
      // UnifiedGraphQLClient 抛出了异常，将错误转换为企业级格式
      console.warn('GraphQL客户端异常，转换为企业级错误格式:', error);
      
      return {
        success: false,
        error: {
          code: 'GRAPHQL_CLIENT_ERROR',
          message: error instanceof Error ? error.message : 'GraphQL客户端请求失败',
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
      variables?: Record<string, JsonValue>;
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
