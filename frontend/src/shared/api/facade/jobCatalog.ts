/**
 * Plan 257 - 领域 Facade（Job Catalog）
 * 提供职类/职种/职务/职级相关命令端 REST 封装。
 */
import { unifiedRESTClient } from '@/shared/api/unified-client';
import type { APIResponse } from '@/shared/types/api';
import type { JobCatalogStatus } from '@/generated/graphql-types';

const json = { 'Content-Type': 'application/json' } as const;
const withIfMatch = (recordId: string) => ({ ...json, 'If-Match': recordId });

export interface CreateJobFamilyGroupInput {
  code: string;
  name: string;
  status: JobCatalogStatus;
  effectiveDate: string;
  description?: string | null;
}
export interface CreateJobFamilyInput {
  code: string;
  jobFamilyGroupCode: string;
  name: string;
  status: JobCatalogStatus;
  effectiveDate: string;
  description?: string | null;
}
export interface CreateJobRoleInput {
  code: string;
  jobFamilyCode: string;
  name: string;
  status: JobCatalogStatus;
  effectiveDate: string;
  description?: string | null;
}
export interface CreateJobLevelInput {
  code: string;
  jobRoleCode: string;
  name: string;
  levelRank: number;
  status: JobCatalogStatus;
  effectiveDate: string;
  description?: string | null;
}
export interface CreateCatalogVersionInput {
  code: string;
  name: string;
  status: JobCatalogStatus;
  effectiveDate: string;
  description?: string | null;
}
export interface UpdateJobFamilyGroupInput {
  code: string;
  recordId: string;
  name: string;
  status: JobCatalogStatus;
  effectiveDate: string;
  description?: string | null;
}
export interface UpdateJobFamilyInput {
  code: string;
  recordId: string;
  jobFamilyGroupCode?: string;
  name: string;
  status: JobCatalogStatus;
  effectiveDate: string;
  description?: string | null;
}
export interface UpdateJobRoleInput {
  code: string;
  recordId: string;
  jobFamilyCode?: string;
  name: string;
  status: JobCatalogStatus;
  effectiveDate: string;
  description?: string | null;
}
export interface UpdateJobLevelInput {
  code: string;
  recordId: string;
  jobRoleCode?: string;
  name: string;
  status: JobCatalogStatus;
  effectiveDate: string;
  description?: string | null;
  levelRank?: number;
}

const ensure = <T>(resp: APIResponse<T>, msg: string): T => {
  if (!resp.success) throw new Error(resp.error?.message || msg);
  return (resp.data as T) ?? ({} as T);
};

export async function createJobFamilyGroup(input: CreateJobFamilyGroupInput): Promise<void> {
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>('/job-family-groups', {
    method: 'POST',
    headers: json,
    body: JSON.stringify(input),
  });
  ensure(resp, 'createJobFamilyGroup failed');
}
export async function updateJobFamilyGroup(input: UpdateJobFamilyGroupInput): Promise<void> {
  const { code, recordId, ...payload } = input;
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-family-groups/${encodeURIComponent(code)}`, {
    method: 'PUT',
    headers: withIfMatch(recordId),
    body: JSON.stringify(payload),
  });
  ensure(resp, 'updateJobFamilyGroup failed');
}

export async function createJobFamily(input: CreateJobFamilyInput): Promise<void> {
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>('/job-families', {
    method: 'POST',
    headers: json,
    body: JSON.stringify(input),
  });
  ensure(resp, 'createJobFamily failed');
}
export async function updateJobFamily(input: UpdateJobFamilyInput): Promise<void> {
  const { code, recordId, ...payload } = input;
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-families/${encodeURIComponent(code)}`, {
    method: 'PUT',
    headers: withIfMatch(recordId),
    body: JSON.stringify(payload),
  });
  ensure(resp, 'updateJobFamily failed');
}

export async function createJobRole(input: CreateJobRoleInput): Promise<void> {
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>('/job-roles', {
    method: 'POST',
    headers: json,
    body: JSON.stringify(input),
  });
  ensure(resp, 'createJobRole failed');
}
export async function updateJobRole(input: UpdateJobRoleInput): Promise<void> {
  const { code, recordId, ...payload } = input;
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-roles/${encodeURIComponent(code)}`, {
    method: 'PUT',
    headers: withIfMatch(recordId),
    body: JSON.stringify(payload),
  });
  ensure(resp, 'updateJobRole failed');
}

export async function createJobLevel(input: CreateJobLevelInput): Promise<void> {
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>('/job-levels', {
    method: 'POST',
    headers: json,
    body: JSON.stringify(input),
  });
  ensure(resp, 'createJobLevel failed');
}
export async function updateJobLevel(input: UpdateJobLevelInput): Promise<void> {
  const { code, recordId, ...payload } = input;
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-levels/${encodeURIComponent(code)}`, {
    method: 'PUT',
    headers: withIfMatch(recordId),
    body: JSON.stringify(payload),
  });
  ensure(resp, 'updateJobLevel failed');
}

export async function createJobFamilyGroupVersion(input: CreateCatalogVersionInput): Promise<void> {
  const { code, ...payload } = input;
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-family-groups/${encodeURIComponent(code)}/versions`, {
    method: 'POST',
    headers: json,
    body: JSON.stringify(payload),
  });
  ensure(resp, 'createJobFamilyGroupVersion failed');
}
export async function createJobFamilyVersion(input: CreateCatalogVersionInput): Promise<void> {
  const { code, ...payload } = input;
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-families/${encodeURIComponent(code)}/versions`, {
    method: 'POST',
    headers: json,
    body: JSON.stringify(payload),
  });
  ensure(resp, 'createJobFamilyVersion failed');
}
export async function createJobRoleVersion(input: CreateCatalogVersionInput): Promise<void> {
  const { code, ...payload } = input;
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-roles/${encodeURIComponent(code)}/versions`, {
    method: 'POST',
    headers: json,
    body: JSON.stringify(payload),
  });
  ensure(resp, 'createJobRoleVersion failed');
}
export async function createJobLevelVersion(input: CreateCatalogVersionInput): Promise<void> {
  const { code, ...payload } = input;
  const resp = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-levels/${encodeURIComponent(code)}/versions`, {
    method: 'POST',
    headers: json,
    body: JSON.stringify(payload),
  });
  ensure(resp, 'createJobLevelVersion failed');
}

