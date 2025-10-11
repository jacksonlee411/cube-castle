import { logger } from '@/shared/utils/logger';
import {
  unifiedGraphQLClient,
  unifiedRESTClient,
} from "@/shared/api/unified-client";
import { env } from "@/shared/config/environment";
import type { OrganizationRequest } from "@/shared/types/organization";
import type { TemporalVersionPayload } from "@/shared/types/temporal";
import type { TimelineVersion } from "../TimelineComponent";

export interface HierarchyPaths {
  codePath: string;
  namePath: string;
}

interface OrganizationVersion {
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

interface OrganizationVersionsResponse {
  organizationVersions: OrganizationVersion[];
}

interface OrganizationSnapshotResponse {
  organization: OrganizationVersion | null;
}

interface TimelineItemResponse {
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

interface OrganizationCreationData {
  code?: string;
  organization?: { code?: string };
}

type CreateOrganizationResponse = SuccessEnvelope<OrganizationCreationData> & OrganizationCreationData;

export interface FetchVersionsResult {
  versions: TimelineVersion[];
  fallbackMessage?: string;
}

interface GraphQLResponseError {
  response?: { status: number; statusText?: string };
  message?: string;
}

const ORGANIZATION_VERSIONS_QUERY = `
  query OrganizationVersions($code: String!) {
    organizationVersions(code: $code) {
      recordId
      code
      name
      unitType
      status
      level
      codePath
      namePath
      effectiveDate
      endDate
      createdAt
      updatedAt
      parentCode
      description
    }
  }
`;

const ORGANIZATION_SNAPSHOT_QUERY = `
  query GetOrganization($code: String!) {
    organization(code: $code) {
      code
      name
      unitType
      status
      level
      codePath
      namePath
      effectiveDate
      endDate
      createdAt
      updatedAt
      recordId
      parentCode
      description
      hierarchyDepth
    }
  }
`;

const ORGANIZATION_HIERARCHY_QUERY = `
  query GetHierarchyPaths($code: String!, $tenantId: String!) {
    organizationHierarchy(code: $code, tenantId: $tenantId) {
      codePath
      namePath
    }
  }
`;

const mapOrganizationVersions = (
  organizations: OrganizationVersion[],
): TimelineVersion[] =>
  organizations
    .map((org) => ({
      recordId: org.recordId,
      code: org.code,
      name: org.name,
      unitType: org.unitType,
      status: org.status,
      level: org.level,
      effectiveDate: org.effectiveDate,
      endDate: org.endDate,
      isCurrent: org.endDate === null,
      createdAt: org.createdAt,
      updatedAt: org.updatedAt,
      parentCode: org.parentCode,
      description: org.description ?? undefined,
      lifecycleStatus:
        org.endDate === null ? ("CURRENT" as const) : ("HISTORICAL" as const),
      businessStatus: (org.status === "ACTIVE" ? "ACTIVE" : "INACTIVE") as "ACTIVE" | "INACTIVE",
      dataStatus: "NORMAL" as const,
      codePath: org.codePath ?? undefined,
      namePath: org.namePath ?? undefined,
      sortOrder: 1,
      changeReason: "",
    }))
    .sort(
      (a, b) =>
        new Date(b.effectiveDate).getTime() -
        new Date(a.effectiveDate).getTime(),
    );

const mapTimelineItem = (item: TimelineItemResponse): TimelineVersion => ({
  recordId: item.recordId,
  code: item.code,
  name: item.name,
  unitType: item.unitType,
  status: item.status,
  level: item.level,
  effectiveDate: item.effectiveDate,
  endDate: item.endDate,
  isCurrent: item.isCurrent,
  createdAt: item.createdAt,
  updatedAt: item.updatedAt,
  parentCode: item.parentCode ?? undefined,
  description: item.description ?? undefined,
  lifecycleStatus: item.isCurrent ? "CURRENT" : "HISTORICAL",
  businessStatus: (item.status === "ACTIVE" ? "ACTIVE" : "INACTIVE") as "ACTIVE" | "INACTIVE",
  dataStatus: "NORMAL",
  codePath: item.codePath ?? undefined,
  namePath: item.namePath ?? undefined,
  sortOrder: 1,
  changeReason: "",
});

export const fetchOrganizationVersions = async (
  organizationCode: string,
): Promise<FetchVersionsResult> => {
  try {
    const data = await unifiedGraphQLClient.request<OrganizationVersionsResponse>(
      ORGANIZATION_VERSIONS_QUERY,
      { code: organizationCode },
    );
    return { versions: mapOrganizationVersions(data.organizationVersions) };
  } catch (graphqlError) {
    logger.warn(
      "organizationVersions查询失败，回退到单体快照逻辑:",
      graphqlError,
    );

    try {
      const data = await unifiedGraphQLClient.request<OrganizationSnapshotResponse>(
        ORGANIZATION_SNAPSHOT_QUERY,
        { code: organizationCode },
      );

      const snapshotVersions = data.organization
        ? mapOrganizationVersions([data.organization])
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
  const result = await unifiedRESTClient.request<CreateOrganizationResponse>(
    "/organization-units",
    {
      method: "POST",
      body: JSON.stringify(payload),
    },
  );

  if (result.data?.code) {
    return result.data.code;
  }

  if (result.data?.organization?.code) {
    return result.data.organization.code;
  }

  if (result.code) {
    return result.code;
  }

  return result.organization?.code ?? null;
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
