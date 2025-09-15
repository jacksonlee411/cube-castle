import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GraphQLEnterpriseAdapter } from '../graphql-enterprise-adapter';
import type { UnifiedGraphQLClient } from '../unified-client';

describe('GraphQLEnterpriseAdapter (Vitest)', () => {
  let adapter: GraphQLEnterpriseAdapter;
  let mockClient: Pick<UnifiedGraphQLClient, 'request'>;

  beforeEach(() => {
    mockClient = { request: vi.fn() } as unknown as Pick<UnifiedGraphQLClient, 'request'>;
    adapter = new GraphQLEnterpriseAdapter(mockClient as UnifiedGraphQLClient);
  });

  it('标准GraphQL成功结果应包装为企业级成功信封', async () => {
    (mockClient.request as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      organizations: { data: [{ code: 'TEST001', name: 'Test Org' }] }
    });

    const result = await adapter.request('{ organizations { data { code name } } }');
    expect(result.success).toBe(true);
    expect(result.data).toEqual({ organizations: { data: [{ code: 'TEST001', name: 'Test Org' }] } });
    expect(result.message).toBe('GraphQL查询成功');
  });

  it('客户端抛错应映射为企业级错误信封', async () => {
    (mockClient.request as unknown as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Network timeout'));

    const result = await adapter.request('{ organizations { code } }');
    expect(result.success).toBe(false);
    expect(result.error).toMatchObject({ code: 'GRAPHQL_CLIENT_ERROR', message: 'Network timeout' });
  });

  it('批量请求：全部成功', async () => {
    const adapterSpy = vi.spyOn(adapter, 'request');
    adapterSpy.mockResolvedValueOnce({ success: true, data: { organization: { code: 'A' } }, message: '', timestamp: '', requestId: '' });
    adapterSpy.mockResolvedValueOnce({ success: true, data: { organization: { code: 'B' } }, message: '', timestamp: '', requestId: '' });

    const res = await adapter.batchRequest([{ query: '{a}' }, { query: '{b}' }]);
    expect(res.success).toBe(true);
    expect(res.data).toEqual([{ organization: { code: 'A' } }, { organization: { code: 'B' } }]);
  });

  it('批量请求：部分失败', async () => {
    const adapterSpy = vi.spyOn(adapter, 'request');
    adapterSpy.mockResolvedValueOnce({ success: true, data: { organization: { code: 'A' } }, message: '', timestamp: '', requestId: '' });
    adapterSpy.mockResolvedValueOnce({ success: false, error: { code: 'NOT_FOUND', message: 'x' }, timestamp: '', requestId: '' });

    const res = await adapter.batchRequest([{ query: '{a}' }, { query: '{b}' }]);
    expect(res.success).toBe(false);
    expect(res.error?.code).toBe('BATCH_REQUEST_PARTIAL_FAILURE');
  });
});
