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

// 查询：组织子树与根组织列表、父组织筛选
export interface OrganizationSubtreeNode {
  code: string;
  name: string;
  unitType: string;
  status: string;
  level: number;
  parentCode?: string | null;
  codePath?: string | null;
  namePath?: string | null;
  parentChain?: string[] | null;
  childrenCount?: number | null;
  hierarchyDepth?: number | null;
  children?: OrganizationSubtreeNode[] | null;
}

export async function getOrganizationSubtree(code: string, maxDepth = 10): Promise<OrganizationSubtreeNode | null> {
  const QUERY = /* GraphQL */ `
    query GetOrganizationSubtree($code: String!, $maxDepth: Int) {
      organizationSubtree(code: $code, maxDepth: $maxDepth) {
        code
        name
        unitType
        status
        level
        parentCode
        codePath
        namePath
        parentChain
        childrenCount
        hierarchyDepth
        children {
          code
          name
          unitType
          status
          level
          parentCode
          codePath
          namePath
          parentChain
          childrenCount
          hierarchyDepth
        }
      }
    }
  `;
  const res = await unifiedGraphQLClient.request<{ organizationSubtree?: OrganizationSubtreeNode }>(QUERY, { code, maxDepth });
  return res?.organizationSubtree ?? null;
}

export async function listRootOrganizations(): Promise<OrganizationSubtreeNode[]> {
  const QUERY = /* GraphQL */ `
    query GetRootOrganizations($filter: OrganizationFilter) {
      organizations(filter: $filter) {
        data {
          code
          name
          unitType
          status
          level
          parentCode
          codePath
          namePath
          hierarchyDepth
        }
      }
    }
  `;
  const res = await unifiedGraphQLClient.request<{ organizations?: { data: OrganizationSubtreeNode[] } }>(QUERY, { filter: { parentCode: null } });
  return res.organizations?.data ?? [];
}

export interface ParentOrgItem {
  code: string;
  name: string;
  unitType: string;
  parentCode?: string;
  level: number;
  effectiveDate: string;
  endDate?: string;
  isFuture: boolean;
  childrenCount?: number;
}

export async function searchValidParentOrganizations(asOfDate: string, currentCode: string, pageSize = 500): Promise<{ data: ParentOrgItem[]; total: number }> {
  const QUERY = /* GraphQL */ `
    query GetValidParentOrganizations($asOfDate: String!, $currentCode: String!, $pageSize: Int = 500) {
      organizations(
        filter: {
          status: ACTIVE
          asOfDate: $asOfDate
          excludeCodes: [$currentCode]
          excludeDescendantsOf: $currentCode
          includeDisabledAncestors: true
        }
        pagination: { page: 1, pageSize: $pageSize, sortBy: "code", sortOrder: "asc" }
      ) {
        data {
          code
          name
          unitType
          parentCode
          level
          effectiveDate
          endDate
          isFuture
          childrenCount
        }
        pagination { total page pageSize }
      }
    }
  `;
  const res = await unifiedGraphQLClient.request<{ organizations: { data: ParentOrgItem[]; pagination: { total: number } } }>(
    QUERY,
    { asOfDate, currentCode, pageSize }
  );
  return { data: res.organizations?.data ?? [], total: res.organizations?.pagination?.total ?? 0 };
}
