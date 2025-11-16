/**
 * Plan 257 - 领域 Facade（Position）
 * 提供职位相关命令端 REST 封装，屏蔽统一客户端细节与头注入。
 */
import { unifiedRESTClient } from '@/shared/api/unified-client';
import type {
  CreatePositionRequest,
  UpdatePositionRequest,
  CreatePositionVersionRequest,
  PositionResource,
} from '@/shared/types/positions';
import type { APIResponse } from '@/shared/types/api';

const json = { 'Content-Type': 'application/json' } as const;

export async function createPosition(payload: CreatePositionRequest): Promise<PositionResource> {
  const resp = await unifiedRESTClient.request<APIResponse<PositionResource>>('/positions', {
    method: 'POST',
    headers: json,
    body: JSON.stringify(payload),
  });
  if (!resp.success || !resp.data) throw new Error(resp.error?.message || 'createPosition failed');
  return resp.data;
}

export async function updatePosition(code: string, payload: Omit<UpdatePositionRequest, 'code'>): Promise<PositionResource> {
  const resp = await unifiedRESTClient.request<APIResponse<PositionResource>>(`/positions/${encodeURIComponent(code)}`, {
    method: 'PUT',
    headers: json,
    body: JSON.stringify(payload),
  });
  if (!resp.success || !resp.data) throw new Error(resp.error?.message || 'updatePosition failed');
  return resp.data;
}

export async function createPositionVersion(code: string, payload: Omit<CreatePositionVersionRequest, 'code'>): Promise<PositionResource> {
  const resp = await unifiedRESTClient.request<APIResponse<PositionResource>>(`/positions/${encodeURIComponent(code)}/versions`, {
    method: 'POST',
    headers: json,
    body: JSON.stringify(payload),
  });
  if (!resp.success || !resp.data) throw new Error(resp.error?.message || 'createPositionVersion failed');
  return resp.data;
}

export interface TransferPositionInput {
  targetOrganizationCode: string;
  effectiveDate: string;
  operationReason: string;
  reassignReports?: boolean;
}

export async function transferPosition(
  code: string,
  input: TransferPositionInput
): Promise<{ payload: unknown; requestId?: string; timestamp?: string }> {
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>(`/positions/${encodeURIComponent(code)}/transfer`, {
    method: 'POST',
    headers: json,
    body: JSON.stringify(input),
  });
  if (!resp.success) throw new Error(resp.error?.message || 'transferPosition failed');
  return { payload: resp.data, requestId: resp.requestId, timestamp: resp.timestamp };
}

