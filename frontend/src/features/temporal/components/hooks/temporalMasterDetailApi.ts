import { logger } from '@/shared/utils/logger';
import {
  listOrganizationVersions,
  getOrganizationByCode,
  createOrganization as facadeCreateOrganization,
} from '@/shared/api/facade/organization';
import { unifiedGraphQLClient, unifiedRESTClient } from '@/shared/api/unified-client';
import { env } from "@/shared/config/environment";
import type { OrganizationRequest } from "@/shared/types/organization";
import type { TemporalVersionPayload } from "@/shared/types/temporal";
import type { TimelineVersion } from '../TimelineComponent';
import { standardObjectAdapter } from '@/features/temporal/entity/standardObjectAdapter';
import type {
  StandardObject,
  StandardObjectAudit,
  StandardObjectKernel,
  StandardObjectLink,
  StandardObjectVersion,
  Status,
} from '@/generated/graphql-types';
import { StandardObjectLinkType, StandardObjectType } from '@/generated/graphql-types';

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
  isCurrent?: boolean | null;
  changeReason?: string | null;
  suspendedAt?: string | null;
  suspendedBy?: string | null;
  suspensionReason?: string | null;
  deletedAt?: string | null;
  deletedBy?: string | null;
  deletionReason?: string | null;
}

// 版本与快照查询改由 Facade 提供

interface TimelineItemResponse extends OrganizationVersion {
  isCurrent: boolean;
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

interface StandardObjectTimelineResponse {
  standardObjects?: {
    data: StandardObject[];
  };
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

const STANDARD_OBJECT_TIMELINE_QUERY = `
  query StandardObjectTimeline($objectType: StandardObjectType!, $code: String!) {
    standardObjects(objectType: $objectType, filter: { code: $code }) {
      data {
        kernel {
          code
          displayName
          tenantCode
          status
          labels
          schemaVersion
          dataClassification
          retentionPolicy
          createdAt
          updatedAt
        }
        version {
          versionCode
          effectiveDate
          endDate
          isCurrent
          payload
          auditTrail
          createdAt
          updatedAt
        }
        links {
          linkType
          sourceCode
          targetCode
          attributes
        }
      }
    }
  }
`;

const toStandardObjectStatus = (status: string): Status => {
  switch (status) {
    case Status.INACTIVE:
      return Status.INACTIVE;
    case Status.PLANNED:
      return Status.PLANNED;
    case Status.DELETED:
      return Status.DELETED;
    default:
      return Status.ACTIVE;
  }
};

const legacyAggregateToStandardObject = (source: OrganizationVersion): StandardObject => {
  const createdAt = source.createdAt ?? source.effectiveDate;
  const updatedAt = source.updatedAt ?? createdAt;
  const kernel: StandardObjectKernel = {
    __typename: 'StandardObjectKernel',
    objectType: StandardObjectType.ORGANIZATION_UNIT,
    code: source.code,
    displayName: source.name,
    tenantCode: env.defaultTenantId,
    status: toStandardObjectStatus(source.status),
    labels: { unitType: source.unitType ?? 'ORGANIZATION' },
    schemaVersion: 'v1',
    dataClassification: 'INTERNAL',
    retentionPolicy: 'default',
    createdBy: undefined,
    createdAt,
    updatedAt,
  };

  const auditTrail: StandardObjectAudit = {
    __typename: 'StandardObjectAudit',
    createdAt,
    updatedAt,
    suspendedAt: source.suspendedAt ?? undefined,
    suspendedBy: source.suspendedBy ?? undefined,
    suspensionReason: source.suspensionReason ?? undefined,
    deletedAt: source.deletedAt ?? undefined,
    deletedBy: source.deletedBy ?? undefined,
    deletionReason: source.deletionReason ?? undefined,
  };

  const payload = {
    name: source.name,
    unitType: source.unitType,
    status: source.status,
    description: source.description,
    level: source.level,
    parentCode: source.parentCode,
    codePath: source.codePath,
    namePath: source.namePath,
    sortOrder: 0,
    changeReason: source.changeReason,
  };

  const version: StandardObjectVersion = {
    __typename: 'StandardObjectVersion',
    versionCode: source.recordId ?? `${source.code}-${source.effectiveDate}`,
    effectiveDate: source.effectiveDate,
    endDate: source.endDate ?? undefined,
    isCurrent: Boolean(source.isCurrent ?? source.endDate === null),
    payload,
    auditTrail,
    createdAt,
    updatedAt,
    checksum: null,
  };

  const links: StandardObjectLink[] = source.parentCode
    ? [
        {
          __typename: 'StandardObjectLink',
          linkType: StandardObjectLinkType.ORG_HIERARCHY,
          sourceCode: source.code,
          targetCode: source.parentCode,
          attributes: { level: source.level ?? 0 },
        },
      ]
    : [];

  return {
    __typename: 'StandardObject',
    kernel,
    version,
    links,
  };
};

const mapOrganizationVersions = (versions: OrganizationVersion[]): TimelineVersion[] => {
  if (!versions?.length) {
    return [];
  }
  const aggregates = versions.map(legacyAggregateToStandardObject);
  return standardObjectAdapter.toTimelineVersions(aggregates);
};

const mapTimelineItem = (item: TimelineItemResponse): TimelineVersion =>
  standardObjectAdapter.toTimelineVersion(legacyAggregateToStandardObject(item));

export const fetchOrganizationVersions = async (
  organizationCode: string,
): Promise<FetchVersionsResult> => {
  try {
    const somVersions = await fetchStandardObjectTimeline(organizationCode);
    if (somVersions.length > 0) {
      return { versions: somVersions };
    }
  } catch (somError) {
    logger.warn('standardObject timeline 查询失败，回退到 legacy 数据源:', somError);
  }

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

const fetchStandardObjectTimeline = async (code: string): Promise<TimelineVersion[]> => {
  const response = await unifiedGraphQLClient.request<StandardObjectTimelineResponse>(
    STANDARD_OBJECT_TIMELINE_QUERY,
    {
      objectType: StandardObjectType.ORGANIZATION_UNIT,
      code,
    },
  );
  const aggregates = response?.standardObjects?.data ?? [];
  if (!aggregates.length) {
    return [];
  }
  return standardObjectAdapter.toTimelineVersions(aggregates);
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
      (a: TimelineVersion, b: TimelineVersion) =>
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
