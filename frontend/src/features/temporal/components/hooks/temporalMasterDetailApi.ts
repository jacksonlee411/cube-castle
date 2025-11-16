import { logger } from '@/shared/utils/logger';
import {
  unifiedGraphQLClient,
  unifiedRESTClient,
} from "@/shared/api/unified-client";
import {
  listOrganizationVersions,
  getOrganizationByCode,
  createOrganization as facadeCreateOrganization,
} from '@/shared/api/facade/organization';
import { env } from "@/shared/config/environment";
import type { OrganizationRequest } from "@/shared/types/organization";
import type { TemporalVersionPayload } from "@/shared/types/temporal";
import type { TimelineVersion } from '../TimelineComponent';
import {
  organizationTimelineAdapter,
  type OrganizationTimelineSource,
} from '@/features/temporal/entity/timelineAdapter';

export interface HierarchyPaths {
  codePath: string;
  namePath: string;
}

interface OrganizationVersion extends OrganizationTimelineSource {
  code: string;
  name: string;
  unitType: string;
  status: string;
  level: number;
  parentCode?: string;
  description?: string;
  codePath?: string | null;
  namePath?: string | null;
  effectiveDate: string;
  endDate?: string | null;
  recordId: string;
  createdAt: string;
  updatedAt: string;
}

// 版本与快照查询改由 Facade 提供

interface TimelineItemResponse extends OrganizationTimelineSource {
  recordId: string;
  code: string;
  name: string;
  unitType: string;
  status: string;
  level: number;
  effectiveDate: string;
  endDate: string | null;
  isCurrent: boolean;
  createdAt: string;
  updatedAt: string;
  parentCode?: string | null;
  description?: string | null;
  codePath?: string | null;
  namePath?: string | null;
}

interface TimelineEventData {
  code?: string;
  status?: string;
  operationType?: string;
  recordId?: string | null;
  timeline?: TimelineItemResponse[];
}

interface SuccessEnvelope<T> {
  success?: boolean;
  data?: T;
  message?: string;
}

type OrganizationEventResponse = SuccessEnvelope<TimelineEventData>;

// 创建组织的响应由 Facade 处理，此处不再定义本地类型

export interface FetchVersionsResult {
  versions: TimelineVersion[];
  fallbackMessage?: string;
}

interface GraphQLResponseError {
  response?: { status: number; statusText?: string };
  message?: string;
}

// GraphQL 查询常量：仅保留层级路径查询（其余由 Facade 承担）

const ORGANIZATION_HIERARCHY_QUERY = `
  query TemporalEntityHierarchyPaths($code: String!, $tenantId: String!) {
    organizationHierarchy(code: $code, tenantId: $tenantId) {
      codePath
      namePath
    }
  }
`;

// 注意：当前 GraphQL 仅返回 status=ACTIVE/INACTIVE 与 isCurrent。
// 这里将 lifecycleStatus 固定映射为 CURRENT/HISTORICAL，dataStatus 固定为 'NORMAL'，
// 以避免误解为后端已提供五态或软删除数据。
const mapOrganizationVersions = (organizations: OrganizationVersion[]): TimelineVersion[] =>
  organizationTimelineAdapter.toTimelineVersions(organizations);

const mapTimelineItem = (item: TimelineItemResponse): TimelineVersion =>
  organizationTimelineAdapter.toTimelineVersion(item);

export const fetchOrganizationVersions = async (
  organizationCode: string,
): Promise<FetchVersionsResult> => {
  try {
    // Plan 257: 优先通过 Facade 查询
    const list = await listOrganizationVersions(organizationCode);
    return { versions: mapOrganizationVersions(list as unknown as OrganizationVersion[]) };
  } catch (graphqlError) {
    logger.warn(
      "organizationVersions查询失败，回退到单体快照逻辑:",
      graphqlError,
    );

    try {
      // 回退使用 Facade 快照
      const snapshot = await getOrganizationByCode(organizationCode);
      const snapshotVersions = snapshot
        ? mapOrganizationVersions([snapshot as unknown as OrganizationVersion])
        : [];

      return {
        versions: snapshotVersions,
        fallbackMessage: "历史列表不可用，展示当前快照",
      };
    } catch (fallbackError) {
      const typedError = fallbackError as GraphQLResponseError;
      if (typedError?.response?.status) {
        const statusCode = typedError.response.status;
        const statusText = typedError.response.statusText || "Unknown Error";
        throw new Error(`服务器响应错误 (${statusCode}): ${statusText}`);
      }
      throw new Error(`GraphQL调用失败: ${typedError?.message || "未知错误"}`);
    }
  }
};

export const fetchHierarchyPaths = async (
  code: string,
): Promise<HierarchyPaths | null> => {
  const response = await unifiedGraphQLClient.request<{
    organizationHierarchy: HierarchyPaths | null;
  }>(ORGANIZATION_HIERARCHY_QUERY, {
    code,
    tenantId: env.defaultTenantId,
  });

  return response?.organizationHierarchy || null;
};

export const deactivateOrganizationVersion = async (
  organizationCode: string,
  version: TimelineVersion,
): Promise<TimelineVersion[] | null> => {
  const response = await unifiedRESTClient.request<OrganizationEventResponse>(
    `/organization-units/${organizationCode}/events`,
    {
      method: "POST",
      body: JSON.stringify({
        eventType: "DEACTIVATE",
        recordId: version.recordId,
        effectiveDate: version.effectiveDate,
        changeReason: "通过组织详情页面作废版本",
      }),
    },
  );

  const timeline = response.data?.timeline;
  if (!timeline || timeline.length === 0) {
    return null;
  }

  return timeline
    .map(mapTimelineItem)
    .sort(
      (a, b) =>
        new Date(b.effectiveDate).getTime() -
        new Date(a.effectiveDate).getTime(),
    );
};

export const createOrganizationUnit = async (
  payload: OrganizationRequest,
): Promise<string | null> => {
  // Plan 257: 通过 Facade 创建组织
  const unit = await facadeCreateOrganization(payload);
  return unit?.code ?? null;
};

export const createTemporalVersion = async (
  organizationCode: string,
  payload: TemporalVersionPayload,
): Promise<void> => {
  await unifiedRESTClient.request(
    `/organization-units/${organizationCode}/versions`,
    {
      method: "POST",
      body: JSON.stringify(payload),
    },
  );
};

export const updateHistoryRecord = async (
  organizationCode: string,
  recordId: string,
  payload: TemporalVersionPayload,
): Promise<void> => {
  await unifiedRESTClient.request(
    `/organization-units/${organizationCode}/history/${recordId}`,
    {
      method: "PUT",
      body: JSON.stringify(payload),
    },
  );
};
