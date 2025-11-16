/**
 * Plan 257 - 前端领域 API Facade（Organization）
 * 目的：
 * - 为业务代码提供稳定、语义化的领域 API，屏蔽 unified-client 与协议细节
 * - 查询→GraphQL；命令→REST（遵循 PostgreSQL 原生 CQRS）
 * - 类型来源：docs/api/* 契约及 Plan 256 生成物（避免第二事实来源）
 */

import { unifiedGraphQLClient, unifiedRESTClient } from '@/shared/api/unified-client';
import type { OrganizationRequest, OrganizationUnit } from '@/shared/types';

// 与 useOrganizationMutations 中一致的 ETag 规范化处理
const formatIfMatchHeader = (etag?: string): string | undefined => {
  if (!etag) return undefined;
  const trimmed = etag.trim();
  if (!trimmed) return undefined;
  if (trimmed.startsWith('"') || trimmed.startsWith('W/')) {
    return trimmed;
  }
  return `"${trimmed}"`;
};

// 查询：按 code 获取组织当前快照（示例最小实现）
export async function getOrganizationByCode(code: string): Promise<OrganizationUnit | null> {
  const QUERY = /* GraphQL */ `
    query OrganizationByCode($code: String!) {
      organization(code: $code) {
        recordId
        code
        name
        unitType
        status
        level
        parentCode
        description
        codePath
        namePath
        effectiveDate
        endDate
        createdAt
        updatedAt
      }
    }
  `;
  const res = await unifiedGraphQLClient.request<{ organization: OrganizationUnit | null }>(QUERY, { code });
  return res.organization ?? null;
}

// 查询：获取组织版本时间线（示例最小实现）
export async function listOrganizationVersions(code: string): Promise<Array<OrganizationUnit>> {
  const QUERY = /* GraphQL */ `
    query OrganizationVersions($code: String!) {
      organizationVersions(code: $code) {
        recordId
        code
        name
        unitType
        status
        level
        parentCode
        description
        codePath
        namePath
        effectiveDate
        endDate
        createdAt
        updatedAt
      }
    }
  `;
  const res = await unifiedGraphQLClient.request<{ organizationVersions: OrganizationUnit[] }>(QUERY, { code });
  return res.organizationVersions ?? [];
}

// 命令：创建组织（遵循 REST 命令端口）
export async function createOrganization(payload: OrganizationRequest): Promise<OrganizationUnit> {
  const resp = await unifiedRESTClient.request<{
    success: boolean;
    data?: OrganizationUnit;
    error?: { code?: string; message?: string };
    requestId?: string;
  }>(`/organization-units`, {
    method: 'POST',
    body: JSON.stringify(payload),
    headers: { 'Content-Type': 'application/json' }
  });
  if (!resp.success || !resp.data) {
    throw new Error(resp.error?.message || 'createOrganization failed');
  }
  return resp.data;
}

// 命令：更新组织（支持可选 If-Match/ETag 并发控制）
export async function updateOrganization(
  code: string,
  payload: OrganizationRequest,
  opts?: { etag?: string }
): Promise<OrganizationUnit> {
  const resp = await unifiedRESTClient.request<{
    success: boolean;
    data?: OrganizationUnit;
    error?: { code?: string; message?: string };
    requestId?: string;
  }>(`/organization-units/${encodeURIComponent(code)}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
    headers: {
      'Content-Type': 'application/json',
      ...(formatIfMatchHeader(opts?.etag) ? { 'If-Match': formatIfMatchHeader(opts?.etag)! } : {})
    }
  });
  if (!resp.success || !resp.data) {
    throw new Error(resp.error?.message || 'updateOrganization failed');
  }
  return resp.data;
}

// 命令：激活/暂停（示例）
export async function activateOrganization(code: string): Promise<boolean> {
  const resp = await unifiedRESTClient.request<{ success: boolean }>(`/organization-units/${encodeURIComponent(code)}/activate`, {
    method: 'POST'
  });
  return !!resp.success;
}

export async function suspendOrganization(code: string): Promise<boolean> {
  const resp = await unifiedRESTClient.request<{ success: boolean }>(`/organization-units/${encodeURIComponent(code)}/suspend`, {
    method: 'POST'
  });
  return !!resp.success;
}
