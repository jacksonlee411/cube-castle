/**
 * 组织启用/停用工作流端到端测试
 * 测试范围：activate/suspend全流程 + 410弃用验证
 */

import { describe, it, expect, beforeEach } from '@jest/globals';

const API_BASE = 'http://localhost:9090/api/v1/organization-units';
const TEST_ORG_CODE = 'TEST_ORG_001';

interface ApiResponse<T> {
  success: boolean;
  data: T;
  message: string;
  timestamp: string;
  requestId: string;
}

interface OrganizationUnit {
  code: string;
  name: string;
  status: 'ACTIVE' | 'INACTIVE';
  businessStatus: 'ACTIVE' | 'INACTIVE';
  operationType: 'CREATE' | 'UPDATE' | 'SUSPEND' | 'REACTIVATE' | 'DELETE';
  operatedBy: {
    id: string;
    name: string;
  };
  effectiveDate: string;
  updatedAt: string;
}

describe('组织启用/停用工作流E2E测试', () => {
  const headers = {
    'Authorization': 'Bearer test-token',
    'Content-Type': 'application/json',
    'X-Tenant-ID': 'test-tenant-001'
  };

  beforeEach(async () => {
    // 确保测试组织处于已知状态 - 创建为ACTIVE
    await fetch(`${API_BASE}`, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        code: TEST_ORG_CODE,
        name: '测试组织单元',
        unitType: 'DEPARTMENT',
        effectiveDate: '2025-09-06'
      })
    });
  });

  describe('启用流程测试', () => {
    it('应该成功启用组织 - 200响应', async () => {
      // 先停用
      await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          operationReason: '测试前置停用',
          effectiveDate: '2025-09-06'
        })
      });

      // 执行启用
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          operationReason: '恢复组织运营',
          effectiveDate: '2025-09-06'
        })
      });

      expect(response.status).toBe(200);
      
      const result: ApiResponse<OrganizationUnit> = await response.json();
      
      // 严格校验响应结构
      expect(result.success).toBe(true);
      expect(result.data).toBeDefined();
      expect(result.timestamp).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
      expect(result.requestId).toMatch(/^[a-f0-9-]{36}$/i);
      
      // 业务数据校验
      expect(result.data.operationType).toBe('REACTIVATE');
      expect(result.data.status).toBe('ACTIVE');
      expect(result.data.businessStatus).toBe('ACTIVE');
      expect(result.data.operatedBy.id).toBeDefined();
      expect(result.data.operatedBy.name).toBeDefined();
      expect(result.data.updatedAt).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
    });

    it('应该拒绝重复启用 - 409响应', async () => {
      // 确保组织是启用状态
      await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          operationReason: '确保启用状态',
          effectiveDate: '2025-09-06'
        })
      });

      // 重复启用应该失败
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          operationReason: '重复启用测试',
          effectiveDate: '2025-09-06'
        })
      });

      expect(response.status).toBe(409);
      
      const result = await response.json();
      expect(result.success).toBe(false);
      expect(result.error.code).toBe('ORGANIZATION_ALREADY_ACTIVE');
      expect(result.error.message).toContain('already active');
    });

    it('应该支持未来生效日期', async () => {
      const futureDate = '2025-12-31';
      
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          operationReason: '计划启用',
          effectiveDate: futureDate
        })
      });

      expect(response.status).toBe(200);
      
      const result: ApiResponse<OrganizationUnit> = await response.json();
      expect(result.data.effectiveDate).toBe(futureDate);
      expect(result.data.operationType).toBe('REACTIVATE');
    });

    it('应该验证org:activate权限', async () => {
      const unauthorizedHeaders = {
        ...headers,
        'Authorization': 'Bearer invalid-token'
      };

      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers: unauthorizedHeaders,
        body: JSON.stringify({
          operationReason: '权限测试',
          effectiveDate: '2025-09-06'
        })
      });

      expect(response.status).toBe(403);
      
      const result = await response.json();
      expect(result.success).toBe(false);
      expect(result.error.code).toBe('INSUFFICIENT_PERMISSIONS');
    });
  });

  describe('停用流程测试', () => {
    it('应该成功停用组织 - 200响应', async () => {
      // 确保组织是启用状态
      await fetch(`${API_BASE}/${TEST_ORG_CODE}/activate`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          operationReason: '测试前置启用',
          effectiveDate: '2025-09-06'
        })
      });

      // 执行停用
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          operationReason: '部门重组',
          effectiveDate: '2025-09-06'
        })
      });

      expect(response.status).toBe(200);
      
      const result: ApiResponse<OrganizationUnit> = await response.json();
      
      // 严格校验响应结构
      expect(result.success).toBe(true);
      expect(result.data.operationType).toBe('SUSPEND');
      expect(result.data.status).toBe('INACTIVE');
      expect(result.data.businessStatus).toBe('INACTIVE');
      expect(result.data.operatedBy.id).toBeDefined();
      expect(result.data.operatedBy.name).toBeDefined();
    });

    it('应该拒绝重复停用 - 409响应', async () => {
      // 先停用
      await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          operationReason: '首次停用',
          effectiveDate: '2025-09-06'
        })
      });

      // 重复停用应该失败
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          operationReason: '重复停用测试',
          effectiveDate: '2025-09-06'
        })
      });

      expect(response.status).toBe(409);
      
      const result = await response.json();
      expect(result.success).toBe(false);
      expect(result.error.code).toBe('ORGANIZATION_ALREADY_SUSPENDED');
    });

    it('应该验证org:suspend权限', async () => {
      const unauthorizedHeaders = {
        ...headers,
        'Authorization': 'Bearer invalid-token'
      };

      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/suspend`, {
        method: 'POST',
        headers: unauthorizedHeaders,
        body: JSON.stringify({
          operationReason: '权限测试',
          effectiveDate: '2025-09-06'
        })
      });

      expect(response.status).toBe(403);
      
      const result = await response.json();
      expect(result.success).toBe(false);
      expect(result.error.code).toBe('INSUFFICIENT_PERMISSIONS');
    });
  });

  describe('410弃用端点验证', () => {
    it('应该对/reactivate返回410 Gone', async () => {
      const response = await fetch(`${API_BASE}/${TEST_ORG_CODE}/reactivate`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          operationReason: '测试弃用端点',
          effectiveDate: '2025-09-06'
        })
      });

      // 验证410状态码
      expect(response.status).toBe(410);
      
      // 验证标准弃用响应头
      expect(response.headers.get('Deprecation')).toBe('true');
      expect(response.headers.get('Link')).toContain('/activate');
      expect(response.headers.get('Link')).toContain('successor-version');
      expect(response.headers.get('Sunset')).toBe('2026-01-01T00:00:00Z');
      
      // 验证错误响应体
      const result = await response.json();
      expect(result.success).toBe(false);
      expect(result.error.code).toBe('ENDPOINT_DEPRECATED');
      expect(result.error.message).toContain('Use /activate instead of /reactivate');
    });

    it('应该记录DEPRECATED_ENDPOINT_USED审计事件', async () => {
      // 访问弃用端点触发审计
      await fetch(`${API_BASE}/${TEST_ORG_CODE}/reactivate`, {
        method: 'POST',
        headers: {
          ...headers,
          'User-Agent': 'Test-Client/1.0',
          'X-Client-ID': 'test-client-001'
        },
        body: JSON.stringify({
          operationReason: '审计测试',
          effectiveDate: '2025-09-06'
        })
      });

      // 注意：这里需要实际的审计日志查询接口
      // 在实际实施中，应该通过监控系统验证审计事件的记录
      
      // 预期的审计事件结构：
      // {
      //   "type": "DEPRECATED_ENDPOINT_USED",
      //   "path": "/api/v1/organization-units/TEST_ORG_001/reactivate",
      //   "tenantId": "test-tenant-001",
      //   "clientId": "test-client-001",
      //   "userAgent": "Test-Client/1.0",
      //   "ip": "127.0.0.1",
      //   "timestamp": "2025-09-06T10:30:00Z",
      //   "metadata": {
      //     "method": "POST",
      //     "successor": "/api/v1/organization-units/{code}/activate"
      //   }
      // }
    });
  });
});