/**
 * Plan 245 – useTemporalEntityDetail 统一 Hook（骨架）
 * 目标：提供统一的时态详情数据获取入口，供职位/组织薄封装复用。
 *
 * 说明：
 * - 最小落地，优先不改动现有消费端；后续以 codemod 渐进替换。
 * - 内部利用现有的 query key 和 fetch 函数，避免重复事实来源。
 * - 保留 Suspense/QueryKey 扩展点，但暂不强推全量迁移。
 *
 * // TODO-TEMPORARY(Plan 245):
 * - 衔接 Position/Organization 的版本列表与时间线统一结构
 * - 清理 usePositionDetail/组织详情获取的重复逻辑，改为薄封装
 */

import { useMemo } from 'react';
import { useQuery, type UseQueryResult, type QueryFunctionContext } from '@tanstack/react-query';
import type {
  TemporalEntityRecord,
  TemporalEntityTimelineEntry,
  TemporalEntityType,
  TemporalEntityDetail,
} from '@/shared/types/temporal-entity';
import {
  organizationByCodeQueryKey as orgDetailKey,
} from '@/shared/hooks/useEnterpriseOrganizations';
import {
  // 复用现有职位详情 Hook，降低首批改动面
  usePositionDetail,
  // 内部 queryKey/fn 目前未导出，这里仅复用 Hook
} from '@/shared/hooks/useEnterprisePositions';
import { graphqlEnterpriseAdapter } from '@/shared/api';
import { ORGANIZATION_BY_CODE_DOCUMENT as ORG_DETAIL_DOC } from '@/shared/hooks/useEnterpriseOrganizations';
import { createQueryError } from '@/shared/api/error-handling';
import type { OrganizationUnit } from '@/shared/types/organization';
import type {
  PositionRecord,
  PositionAssignmentRecord,
  PositionTransferRecord,
} from '@/shared/types/positions';

// GraphQL 文档与类型（直接复用 useEnterpriseOrganizations 的逻辑）
// 复用统一导出的组织详情 GraphQL 文档，避免重复与命名冲突

export interface TemporalDetailOptions {
  enabled?: boolean;
  includeDeleted?: boolean; // 职位用到；组织忽略
  asOfDate?: string; // 组织详情可选
}

export type TemporalDetailResult = UseQueryResult<{
  record: TemporalEntityRecord | null;
  versions?: TemporalEntityRecord[];
  timeline?: TemporalEntityTimelineEntry[];
  // 过渡期：为职位页面保留的旧字段（薄封装形态，保证兼容）
  position?: PositionRecord;
  assignments?: PositionAssignmentRecord[];
  currentAssignment?: PositionAssignmentRecord | null;
  transfers?: PositionTransferRecord[];
}>;

export const temporalEntityDetailQueryKey = (
  entity: TemporalEntityType,
  code: string,
  opts?: TemporalDetailOptions,
) => ['temporal-entity-detail', entity, code, opts?.includeDeleted ?? false, opts?.asOfDate ?? null] as const;

async function fetchOrganizationDetail({
  queryKey,
  signal,
}: QueryFunctionContext<ReturnType<typeof temporalEntityDetailQueryKey>>): Promise<TemporalEntityRecord | null> {
  const [, , code, _includeDeleted, asOfDate] = queryKey;
  const response = await graphqlEnterpriseAdapter.request<{ organization: OrganizationUnit | null }>(
    ORG_DETAIL_DOC,
    { code, asOfDate },
    { signal },
  );
  if (!response.success) {
    throw createQueryError(response.error?.message ?? '获取组织详情失败', {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    });
  }
  const org = response.data?.organization;
  if (!org) return null;
  // 归一化为 TemporalEntityRecord
  const rec: TemporalEntityRecord = {
    entityType: 'organization',
    code: org.code,
    recordId: (org as any).recordId ?? null,
    displayName: org.name,
    organizationCode: org.parentCode ?? null,
    organizationName: org.name ?? null,
    status: org.status as any,
    effectiveDate: (org as any).effectiveDate ?? null,
    endDate: (org as any).endDate ?? null,
  };
  return rec;
}

export function useTemporalEntityDetail(
  entity: TemporalEntityType,
  code: string | undefined,
  options?: TemporalDetailOptions,
): TemporalDetailResult {
  const enabled = Boolean(code) && (options?.enabled ?? true);

  // 职位：直接复用现有 Hook，返回结构做轻量统一
  if (entity === 'position') {
    const result = usePositionDetail(code!, { enabled, includeDeleted: options?.includeDeleted });
    return useMemo(
      () => ({
        ...result,
        data: result.data
          ? {
              record: {
                entityType: 'position',
                code: (result.data as any).position?.code ?? code!,
                recordId: (result.data as any).position?.recordId ?? null,
                displayName: (result.data as any).position?.title ?? null,
                organizationCode: (result.data as any).position?.organizationCode ?? null,
                organizationName: (result.data as any).position?.organizationName ?? null,
                status: (result.data as any).position?.status ?? null,
                effectiveDate: (result.data as any).position?.effectiveDate ?? null,
                endDate: (result.data as any).position?.endDate ?? null,
              } as TemporalEntityRecord,
              // 保留旧字段，确保消费端兼容
              position: (result.data as any).position as PositionRecord,
              timeline: (result.data as any).timeline as any,
              assignments: (result.data as any).assignments as PositionAssignmentRecord[] | undefined,
              currentAssignment: (result.data as any).currentAssignment as PositionAssignmentRecord | null | undefined,
              transfers: (result.data as any).transfers as PositionTransferRecord[] | undefined,
              // 统一版本（可选）：暂不强制使用，原始 versions 继续提供
              // versionsUnified: ((result.data as any).versions as PositionRecord[] | undefined)?.map(v => ({
              //   entityType: 'position', code: v.code, recordId: v.recordId ?? null, displayName: v.title ?? null, organizationCode: v.organizationCode ?? null, organizationName: v.organizationName ?? null, status: v.status, effectiveDate: v.effectiveDate ?? null, endDate: v.endDate ?? null,
              // })),
              // 仍提供原始结构供现有页面使用
              versions: (result.data as any).versions as any,
            }
          : undefined,
      }),
      [result, code],
    ) as TemporalDetailResult;
  }

  // 组织：按 GraphQL 详情查询归一化
  const queryKey = temporalEntityDetailQueryKey('organization', code ?? 'placeholder', {
    asOfDate: options?.asOfDate,
  });

  return useQuery({
    queryKey,
    queryFn: fetchOrganizationDetail,
    enabled,
    staleTime: 60_000,
    select: (rec) => (rec ? { record: rec } : { record: null }),
  }) as TemporalDetailResult;
}
