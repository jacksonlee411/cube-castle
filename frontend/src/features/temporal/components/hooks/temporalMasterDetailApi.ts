import {
  unifiedGraphQLClient,
  unifiedRESTClient,
} from "../../../shared/api/unified-client";
import { env } from "../../../shared/config/environment";
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
  path?: string | null;
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
      path
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
      path
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
      description: org.description,
      lifecycleStatus:
        org.endDate === null ? ("CURRENT" as const) : ("HISTORICAL" as const),
      businessStatus: org.status === "ACTIVE" ? "ACTIVE" : "INACTIVE",
      dataStatus: "NORMAL" as const,
      path: org.path ?? undefined,
      sortOrder: 1,
      changeReason: "",
    }))
    .sort(
      (a, b) =>
        new Date(b.effectiveDate).getTime() -
        new Date(a.effectiveDate).getTime(),
    );

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
    console.warn(
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
  const resp = (await unifiedRESTClient.request(
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
  )) as { data?: { timeline?: Record<string, unknown>[] } };

  const timeline = resp?.data?.timeline;
  if (!Array.isArray(timeline)) {
    return null;
  }

  const mapped = timeline.map((item) => {
    const isCurrent =
      (item.endDate as string) === null || item.endDate === undefined;
    return {
      recordId: item.recordId as string,
      code: item.code as string,
      name: item.name as string,
      unitType: item.unitType as string,
      status: item.status as string,
      level: item.level as number,
      effectiveDate: item.effectiveDate as string,
      endDate: (item.endDate as string) || null,
      isCurrent,
      createdAt: item.createdAt as string,
      updatedAt: item.updatedAt as string,
      parentCode: (item.parentCode as string) || undefined,
      description: (item.description as string) || undefined,
      lifecycleStatus: isCurrent ? "CURRENT" : "HISTORICAL",
      businessStatus: item.status === "ACTIVE" ? "ACTIVE" : "INACTIVE",
      dataStatus: "NORMAL",
      path: (item.path as string | undefined) ?? undefined,
      sortOrder: 1,
      changeReason: "",
    } as TimelineVersion;
  });

  return mapped.sort(
    (a, b) =>
      new Date(b.effectiveDate).getTime() -
      new Date(a.effectiveDate).getTime(),
  );
};

export const createOrganizationUnit = async (
  payload: Record<string, unknown>,
): Promise<string | null> => {
  const result = (await unifiedRESTClient.request(
    "/organization-units",
    {
      method: "POST",
      body: JSON.stringify(payload),
    },
  )) as Record<string, unknown>;

  interface CreateResult {
    code?: string;
    organization?: { code?: string };
    data?: { code?: string };
  }

  const typedResult = result as CreateResult;
  return (
    typedResult.data?.code ||
    typedResult.code ||
    typedResult.organization?.code ||
    null
  );
};

export const createTemporalVersion = async (
  organizationCode: string,
  payload: Record<string, unknown>,
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
  payload: Record<string, unknown>,
): Promise<void> => {
  await unifiedRESTClient.request(
    `/organization-units/${organizationCode}/history/${recordId}`,
    {
      method: "PUT",
      body: JSON.stringify(payload),
    },
  );
};
