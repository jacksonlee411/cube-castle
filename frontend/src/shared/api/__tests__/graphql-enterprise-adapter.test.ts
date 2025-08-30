/**
 * GraphQL企业级响应适配器测试
 * 验证响应格式自动检测和转换功能
 */

import { GraphQLEnterpriseAdapter } from '../graphql-enterprise-adapter';
import { UnifiedGraphQLClient } from '../unified-client';

// Mock UnifiedGraphQLClient
jest.mock('../unified-client');

describe('GraphQLEnterpriseAdapter', () => {
  let adapter: GraphQLEnterpriseAdapter;
  let mockClient: jest.Mocked<UnifiedGraphQLClient>;

  beforeEach(() => {
    mockClient = new UnifiedGraphQLClient() as jest.Mocked<UnifiedGraphQLClient>;
    adapter = new GraphQLEnterpriseAdapter(mockClient);
  });

  describe('企业级响应格式检测', () => {
    it('应该正确识别企业级响应格式', () => {
      const enterpriseResponse = {
        success: true,
        data: { test: 'data' },
        message: 'Success',
        timestamp: '2025-08-25T12:00:00Z',
        requestId: 'req_123'
      };

      // 使用私有方法测试 - 类型安全的访问方式
      const result = (adapter as unknown as { isEnterpriseFormat: (response: unknown) => boolean }).isEnterpriseFormat(enterpriseResponse);
      expect(result).toBe(true);
    });

    it('应该正确识别标准GraphQL响应格式', () => {
      const standardResponse = {
        data: { test: 'data' }
      };

      const result = (adapter as unknown as { isStandardFormat: (response: unknown) => boolean }).isStandardFormat(standardResponse);
      expect(result).toBe(true);
    });

    it('应该正确识别带错误的标准GraphQL响应', () => {
      const errorResponse = {
        errors: [
          {
            message: 'Test error',
            locations: [{ line: 1, column: 1 }],
            path: ['test']
          }
        ]
      };

      const result = (adapter as unknown as { isStandardFormat: (response: unknown) => boolean }).isStandardFormat(errorResponse);
      expect(result).toBe(true);
    });
  });

  describe('标准格式到企业级格式转换', () => {
    it('应该正确转换成功的标准响应', () => {
      const standardResponse = {
        data: { organizations: [{ code: 'TEST001', name: 'Test Org' }] }
      };

      const result = (adapter as unknown as { transformToEnterpriseFormat: (response: unknown) => unknown }).transformToEnterpriseFormat(standardResponse);
      
      expect(result).toMatchObject({
        success: true,
        data: { organizations: [{ code: 'TEST001', name: 'Test Org' }] },
        message: 'GraphQL查询成功'
      });
      expect(result.timestamp).toBeDefined();
      expect(result.requestId).toMatch(/^req_/);
    });

    it('应该正确转换带错误的标准响应', () => {
      const errorResponse = {
        errors: [
          {
            message: 'Field "invalid" not found',
            locations: [{ line: 2, column: 5 }],
            path: ['organizations'],
            extensions: { code: 'GRAPHQL_VALIDATION_FAILED' }
          }
        ]
      };

      const result = (adapter as unknown as { transformToEnterpriseFormat: (response: unknown) => unknown }).transformToEnterpriseFormat(errorResponse);
      
      expect(result).toMatchObject({
        success: false,
        error: {
          code: 'GRAPHQL_ERROR',
          message: 'Field "invalid" not found',
          details: {
            locations: [{ line: 2, column: 5 }],
            path: ['organizations'],
            extensions: { code: 'GRAPHQL_VALIDATION_FAILED' }
          }
        },
        message: 'GraphQL查询失败: Field "invalid" not found'
      });
      expect(result.timestamp).toBeDefined();
      expect(result.requestId).toMatch(/^req_/);
    });
  });

  describe('实际GraphQL请求处理', () => {
    beforeEach(() => {
      // Mock fetch globally
      global.fetch = jest.fn();
    });

    afterEach(() => {
      jest.resetAllMocks();
    });

    it('应该处理后端已经返回企业级格式的响应', async () => {
      const enterpriseResponse = {
        success: true,
        data: { organizations: { data: [] } },
        message: 'Query successful',
        timestamp: '2025-08-25T12:00:00Z',
        requestId: 'backend_req_123'
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(enterpriseResponse)
      });

      const result = await adapter.request('{ organizations(first: 5) { code name } }');
      
      expect(result).toEqual(enterpriseResponse);
      expect(fetch).toHaveBeenCalledWith(
        undefined, // endpoint will be undefined in test
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            query: '{ organizations(first: 5) { code name } }',
            variables: undefined
          })
        }
      );
    });

    it('应该转换标准GraphQL格式到企业级格式', async () => {
      const standardResponse = {
        data: { organizations: { data: [{ code: 'TEST001' }] } }
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(standardResponse)
      });

      const result = await adapter.request('{ organizations(first: 5) { code name } }');
      
      expect(result.success).toBe(true);
      expect(result.data).toEqual({ organizations: { data: [{ code: 'TEST001' }] } });
      expect(result.message).toBe('GraphQL查询成功');
      expect(result.timestamp).toBeDefined();
      expect(result.requestId).toMatch(/^req_/);
    });

    it('应该处理HTTP错误', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error'
      });

      const result = await adapter.request('{ organizations }');
      
      expect(result).toMatchObject({
        success: false,
        error: {
          code: 'NETWORK_ERROR',
          message: 'HTTP Error: 500 Internal Server Error'
        }
      });
      expect(result.timestamp).toBeDefined();
      expect(result.requestId).toMatch(/^req_/);
    });

    it('应该处理网络异常', async () => {
      (global.fetch as jest.Mock).mockRejectedValue(new Error('Network timeout'));

      const result = await adapter.request('{ organizations }');
      
      expect(result).toMatchObject({
        success: false,
        error: {
          code: 'NETWORK_ERROR',
          message: 'Network timeout'
        }
      });
      expect(result.timestamp).toBeDefined();
      expect(result.requestId).toMatch(/^req_/);
    });
  });

  describe('批量请求处理', () => {
    beforeEach(() => {
      global.fetch = jest.fn();
    });

    afterEach(() => {
      jest.resetAllMocks();
    });

    it('应该处理批量请求成功场景', async () => {
      const responses = [
        { success: true, data: { organization: { code: 'TEST001' } }, timestamp: '2025-08-25T12:00:00Z', requestId: 'req_1' },
        { success: true, data: { organization: { code: 'TEST002' } }, timestamp: '2025-08-25T12:00:01Z', requestId: 'req_2' }
      ];

      (global.fetch as jest.Mock)
        .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(responses[0]) })
        .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(responses[1]) });

      const requests = [
        { query: '{ organization(code: "TEST001") { code } }' },
        { query: '{ organization(code: "TEST002") { code } }' }
      ];

      const result = await adapter.batchRequest(requests);
      
      expect(result.success).toBe(true);
      expect(result.data).toEqual([
        { organization: { code: 'TEST001' } },
        { organization: { code: 'TEST002' } }
      ]);
      expect(result.message).toBe('批量请求成功完成 2 个操作');
    });

    it('应该处理批量请求部分失败场景', async () => {
      const responses = [
        { success: true, data: { organization: { code: 'TEST001' } }, timestamp: '2025-08-25T12:00:00Z', requestId: 'req_1' },
        { success: false, error: { code: 'NOT_FOUND', message: 'Organization not found' }, timestamp: '2025-08-25T12:00:01Z', requestId: 'req_2' }
      ];

      (global.fetch as jest.Mock)
        .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(responses[0]) })
        .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(responses[1]) });

      const requests = [
        { query: '{ organization(code: "TEST001") { code } }' },
        { query: '{ organization(code: "NONEXISTENT") { code } }' }
      ];

      const result = await adapter.batchRequest(requests);
      
      expect(result.success).toBe(false);
      expect(result.error?.code).toBe('BATCH_REQUEST_PARTIAL_FAILURE');
      expect(result.error?.message).toBe('批量请求中有 1 个失败');
      expect(result.error?.details).toEqual([responses[1]]);
    });
  });
});