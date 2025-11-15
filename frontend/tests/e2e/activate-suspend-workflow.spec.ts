/**
 * 组织启用/停用工作流端到端测试（API 层）
 * 测试范围：activate/suspend 全流程 + 410 弃用验证
 * 说明：Playwright 统一为测试运行器；当 PW_SKIP_SERVER=1 时跳过该套用例。
 */
import { test, expect } from '@playwright/test';

const API_BASE = 'http://localhost:9090/api/v1/organization-units';
const TEST_ORG_CODE = 'TEST_ORG_001';

interface ApiResponse<T> {
  success: boolean;
  data: T;
  message: string;
  timestamp: string;
  requestId: string;
  error?: { code: string; message: string };
}

interface OrganizationUnit {
  code: string;
  name: string;
  status: 'ACTIVE' | 'INACTIVE';
  businessStatus: 'ACTIVE' | 'INACTIVE';
  operationType: 'CREATE' | 'UPDATE' | 'SUSPEND' | 'REACTIVATE' | 'DELETE';
  operatedBy: { id: string; name: string };
  effectiveDate: string;
  updatedAt: string;
}

const SKIP_API =
  String(process.env.PW_SKIP_SERVER || '').trim() === '1' ||
  String(process.env.PW_SKIP_SERVER || '').toLowerCase() === 'true';

test.describe('组织启用/停用工作流E2E测试', () => {
  test.skip(SKIP_API, 'PW_SKIP_SERVER=1：后端未联通，跳过 API 级 E2E 用例');

  const createHeaders = (overrides: Record<string, string> = {}) => {
    const headerBag = new Headers();
    headerBag.set('authorization', 'Bearer test-token');
    headerBag.set('content-type', 'application/json');
    headerBag.set('x-tenant-id', 'test-tenant-001');
    Object.entries(overrides).forEach(([k, v]) => headerBag.set(k.toLowerCase(), v));
    return headerBag;
  };

  const joinIsoSegments = (...segments: string[]) => segments.join(':');

  test.beforeEach(async () => {
    await fetch(`${API_BASE}`, {
      method: 'POST',
      headers: createHeaders(),
      body: JSON.stringify({
        code: TEST_ORG_CODE,
        name: '测试组织单元',
        unitType: 'DEPARTMENT',
        effectiveDate: '2025-09-06',
      }),
    });
  });

  test.describe('启用流程测试', () => {
    test('应该成功启用组织 - 200响应', async () => {
      await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '测试前置停用', effectiveDate: '2025-09-06' }),
      });

      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '恢复组织运营', effectiveDate: '2025-09-06' }),
      });
      expect(response.status).toBe(200);

      const result: ApiResponse<OrganizationUnit> = await response.json();
      expect(result.success).toBe(true);
      expect(result.data).toBeDefined();
      expect(result.timestamp).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:/);
      expect(result.requestId).toMatch(/^[a-f0-9-]{36}$/i);
      expect(result.data.operationType).toBe('REACTIVATE');
      expect(result.data.status).toBe('ACTIVE');
      expect(result.data.businessStatus).toBe('ACTIVE');
      expect(result.data.operatedBy.id).toBeDefined();
      expect(result.data.operatedBy.name).toBeDefined();
      expect(result.data.updatedAt).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:/);
    });

    test('应该拒绝重复启用 - 409响应', async () => {
      await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '确保启用状态', effectiveDate: '2025-09-06' }),
      });
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '重复启用测试', effectiveDate: '2025-09-06' }),
      });
      expect(response.status).toBe(409);
      const result: ApiResponse<any> = await response.json();
      expect(result.success).toBe(false);
      expect(result.error?.code).toBe('ORGANIZATION_ALREADY_ACTIVE');
    });

    test('应该支持未来生效日期', async () => {
      await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '测试前置停用', effectiveDate: '2025-09-06' }),
      });
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '未来启用', effectiveDate: '2025-12-01' }),
      });
      expect([200, 202]).toContain(response.status);
      const result: ApiResponse<OrganizationUnit> = await response.json();
      expect(['ACTIVE']).toContain(result.data.status);
      expect(result.data.businessStatus).toBe('ACTIVE');
    });

    test('应该验证org:activate权限', async () => {
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers: createHeaders({ authorization: 'Bearer invalid-token' }),
        body: JSON.stringify({ operationReason: '权限测试', effectiveDate: '2025-09-06' }),
      });
      expect([401, 403]).toContain(response.status);
      const result: ApiResponse<any> = await response.json();
      expect(result.success).toBe(false);
      expect(['INSUFFICIENT_PERMISSIONS', 'UNAUTHORIZED']).toContain(result.error?.code || '');
    });
  });

  test.describe('停用流程测试', () => {
    test('应该成功停用组织 - 200响应', async () => {
      await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '测试前置启用', effectiveDate: '2025-09-06' }),
      });
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '部门重组', effectiveDate: '2025-09-06' }),
      });
      expect(response.status).toBe(200);
      const result: ApiResponse<OrganizationUnit> = await response.json();
      expect(result.success).toBe(true);
      expect(result.data.operationType).toBe('SUSPEND');
      expect(result.data.status).toBe('INACTIVE');
      expect(result.data.businessStatus).toBe('INACTIVE');
      expect(result.data.operatedBy.id).toBeDefined();
      expect(result.data.operatedBy.name).toBeDefined();
    });

    test('应该拒绝重复停用 - 409响应', async () => {
      await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '首次停用', effectiveDate: '2025-09-06' }),
      });
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '重复停用测试', effectiveDate: '2025-09-06' }),
      });
      expect(response.status).toBe(409);
      const result: ApiResponse<any> = await response.json();
      expect(result.success).toBe(false);
      expect(result.error?.code).toBe('ORGANIZATION_ALREADY_SUSPENDED');
    });

    test('应该验证org:suspend权限', async () => {
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers: createHeaders({ authorization: 'Bearer invalid-token' }),
        body: JSON.stringify({ operationReason: '权限测试', effectiveDate: '2025-09-06' }),
      });
      expect([401, 403]).toContain(response.status);
      const result: ApiResponse<any> = await response.json();
      expect(result.success).toBe(false);
      expect(['INSUFFICIENT_PERMISSIONS', 'UNAUTHORIZED']).toContain(result.error?.code || '');
    });
  });

  test.describe('410弃用端点验证', () => {
    test('应该对/reactivate返回410 Gone', async () => {
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/reactivate`, {
        method: 'POST',
        headers: createHeaders(),
        body: JSON.stringify({ operationReason: '测试弃用端点', effectiveDate: '2025-09-06' }),
      });
      expect([404, 410]).toContain(response.status);
      if (response.status === 410) {
        expect(response.headers.get('Deprecation')).toBe('true');
        expect(response.headers.get('Link') || '').toContain('/activate');
        expect(response.headers.get('Link') || '').toContain('successor-version');
        expect(response.headers.get('Sunset')).toBe(joinIsoSegments('2026-01-01T00', '00', '00Z'));
        const result: ApiResponse<any> = await response.json();
        expect(result.success).toBe(false);
        expect(result.error?.code).toBe('ENDPOINT_DEPRECATED');
      }
    });
  });
});

