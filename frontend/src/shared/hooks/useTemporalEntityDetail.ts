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
} from '@/shared/types/temporal-entity';
import {
  // 复用现有职位详情 Hook，降低首批改动面
  usePositionDetail,
  // 内部 queryKey/fn 目前未导出，这里仅复用 Hook
} from '@/shared/hooks/useEnterprisePositions';
import { graphqlEnterpriseAdapter } from '@/shared/api';
import { ORGANIZATION_BY_CODE_DOCUMENT as ORG_DETAIL_DOC } from '@/shared/hooks/useEnterpriseOrganizations';
// 统一错误工厂：与其他 Hook 保持一致来源，避免重复事实来源
import { createQueryError } from '@/shared/api/queryClient';
import type { OrganizationUnit } from '@/shared/types/organization';
import type {
  PositionRecord,
  PositionAssignmentRecord,
  PositionTransferRecord,
  PositionTimelineEvent,
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
    recordId: org.recordId ?? null,
    displayName: org.name,
    organizationCode: org.parentCode ?? null,
    organizationName: org.name ?? null,
    status: org.status,
    effectiveDate: org.effectiveDate ?? null,
    endDate: org.endDate ?? null,
  };
  return rec;
}

export function useTemporalEntityDetail(
  entity: TemporalEntityType,
  code: string | undefined,
  options?: TemporalDetailOptions,
): TemporalDetailResult {
  const enabled = Boolean(code) && (options?.enabled ?? true);
  const isPosition = entity === 'position';

  // 职位：始终调用 Hook，但通过 enabled 控制是否执行，避免条件调用 Hook
  const positionQuery = usePositionDetail(code ?? '', {
    enabled: isPosition && enabled,
    includeDeleted: options?.includeDeleted,
  });

  // 组织：按 GraphQL 详情查询归一化；通过 enabled 控制执行
  const queryKey = temporalEntityDetailQueryKey('organization', code ?? 'placeholder', {
    asOfDate: options?.asOfDate,
  });
  const organizationQuery = useQuery({
    queryKey,
    queryFn: fetchOrganizationDetail,
    enabled: !isPosition && enabled,
    staleTime: 60_000,
    select: (rec) => (rec ? { record: rec } : { record: null }),
  }) as UseQueryResult<{ record: TemporalEntityRecord | null }>;

  // 统一数据映射，避免 any，使用未知类型后再收窄
  type PositionHookData = {
    position?: PositionRecord | null;
    versions?: PositionRecord[];
    timeline?: PositionTimelineEvent[];
    assignments?: PositionAssignmentRecord[];
    currentAssignment?: PositionAssignmentRecord | null;
    transfers?: PositionTransferRecord[];
  } | undefined;

  const unifiedData = useMemo(() => {
    if (isPosition) {
      const d = positionQuery.data as unknown as PositionHookData;
      if (!d || !d.position) return undefined;
      const p = d.position;
      const record: TemporalEntityRecord = {
        entityType: 'position',
        code: p.code,
        recordId: p.recordId ?? null,
        displayName: p.title ?? null,
        organizationCode: p.organizationCode ?? null,
        organizationName: p.organizationName ?? null,
        status: p.status,
        effectiveDate: p.effectiveDate ?? null,
        endDate: p.endDate ?? null,
      };
      return {
        record,
        position: p,
        timeline: d.timeline,
        assignments: d.assignments,
        currentAssignment: d.currentAssignment ?? null,
        transfers: d.transfers,
        versions: d.versions,
      };
    }
    // 组织
    const org = organizationQuery.data?.record ?? null;
    return { record: org } as { record: TemporalEntityRecord | null };
  }, [isPosition, positionQuery.data, organizationQuery.data]);

  // 合并 Query 状态，优先返回当前实体对应的 Query 状态
  const base = (isPosition ? positionQuery : organizationQuery) as unknown as TemporalDetailResult;
  return { ...base, data: unifiedData } as TemporalDetailResult;
}
