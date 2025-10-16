import { useMemo } from 'react';
import type { QueryFunctionContext, UseQueryResult } from '@tanstack/react-query';
import { useQuery } from '@tanstack/react-query';
import { graphqlEnterpriseAdapter } from '../api/graphql-enterprise-adapter';
import { createQueryError } from '../api/queryClient';
import type {
  PositionAssignmentRecord,
  PositionDetailResult,
  PositionRecord,
  PositionStatus,
  PositionTimelineEvent,
  PositionTransferRecord,
  PositionsQueryResult,
} from '../types/positions';

const DEFAULT_PAGE = 1;
const DEFAULT_PAGE_SIZE = 50;
const MAX_PAGE_SIZE = 200;

export interface PositionQueryParams {
  page?: number;
  pageSize?: number;
  status?: string;
  jobFamilyGroupCode?: string;
  jobFamilyCode?: string;
  jobRoleCode?: string;
  jobLevelCode?: string;
  positionType?: string;
  employmentType?: string;
}

interface NormalizedPositionQueryParams {
  page: number;
  pageSize: number;
  status?: string;
  jobFamilyGroupCode?: string;
  jobFamilyCode?: string;
  jobRoleCode?: string;
  jobLevelCode?: string;
  positionType?: string;
  employmentType?: string;
}

interface PositionAssignmentGraphQLNode {
  assignmentId: string;
  positionCode: string;
  positionRecordId?: string | null;
  employeeId: string;
  employeeName: string;
  employeeNumber?: string | null;
  assignmentType: string;
  assignmentStatus: string;
  fte: number;
  startDate: string;
  endDate?: string | null;
  isCurrent: boolean;
  notes?: string | null;
  createdAt: string;
  updatedAt: string;
}

interface PositionTransferGraphQLNode {
  transferId: string;
  positionCode: string;
  fromOrganizationCode: string;
  toOrganizationCode: string;
  effectiveDate: string;
  initiatedBy: {
    id: string;
    name: string;
  };
  operationReason?: string | null;
  createdAt: string;
}

interface PositionNodeResponse {
  code: string;
  title: string;
  jobFamilyGroupCode: string;
  jobFamilyGroupName?: string | null;
  jobFamilyCode: string;
  jobFamilyName?: string | null;
  jobRoleCode: string;
  jobRoleName?: string | null;
  jobLevelCode: string;
  jobLevelName?: string | null;
  organizationCode: string;
  organizationName?: string | null;
  positionType: string;
  employmentType: string;
  headcountCapacity: number;
  headcountInUse: number;
  availableHeadcount?: number | null;
  gradeLevel?: string | null;
  reportsToPositionCode?: string | null;
  status: PositionStatus;
  effectiveDate: string;
  endDate?: string | null;
  isCurrent: boolean;
  isFuture: boolean;
  createdAt: string;
  updatedAt: string;
  currentAssignment?: PositionAssignmentGraphQLNode | null;
}

interface PositionTimelineResponse {
  recordId: string;
  status: string;
  title: string;
  effectiveDate: string;
  endDate?: string | null;
  changeReason?: string | null;
  isCurrent?: boolean;
}

interface PositionsGraphQLResponse {
  positions: {
    data: PositionNodeResponse[];
    pagination: {
      total: number;
      page: number;
      pageSize: number;
      hasNext: boolean;
      hasPrevious: boolean;
    };
    totalCount: number;
  };
}

interface PositionDetailGraphQLResponse {
  position: PositionNodeResponse | null;
  positionTimeline: PositionTimelineResponse[];
  positionAssignments: {
    data: PositionAssignmentGraphQLNode[];
  };
  positionTransfers: {
    data: PositionTransferGraphQLNode[];
  };
}

const POSITIONS_QUERY_DOCUMENT = /* GraphQL */ `
  query EnterprisePositions($filter: PositionFilterInput, $pagination: PaginationInput) {
    positions(filter: $filter, pagination: $pagination) {
      data {
        code
        title
        jobFamilyGroupCode
        jobFamilyCode
        jobRoleCode
        jobLevelCode
        organizationCode
        organizationName
        positionType
        employmentType
        headcountCapacity
        headcountInUse
        availableHeadcount
        gradeLevel
        reportsToPositionCode
        status
        effectiveDate
        endDate
        isCurrent
        isFuture
        createdAt
        updatedAt
      }
      pagination {
        total
        page
        pageSize
        hasNext
        hasPrevious
      }
      totalCount
    }
  }
`;

const POSITION_DETAIL_QUERY_DOCUMENT = /* GraphQL */ `
  query PositionDetail($code: PositionCode!) {
    position(code: $code) {
      code
      title
      jobFamilyGroupCode
      jobFamilyCode
      jobRoleCode
      jobLevelCode
      organizationCode
      organizationName
      positionType
      employmentType
      headcountCapacity
      headcountInUse
      availableHeadcount
      gradeLevel
      reportsToPositionCode
      status
      effectiveDate
      endDate
      isCurrent
      isFuture
      createdAt
      updatedAt
      currentAssignment {
        assignmentId
        positionCode
        positionRecordId
        employeeId
        employeeName
        employeeNumber
        assignmentType
        assignmentStatus
        fte
        startDate
        endDate
        isCurrent
        notes
        createdAt
        updatedAt
      }
    }
    positionTimeline(code: $code) {
      recordId
      status
      title
      effectiveDate
      endDate
      changeReason
      isCurrent
    }
    positionAssignments(
      positionCode: $code
      filter: { includeHistorical: true }
      pagination: { page: 1, pageSize: 50 }
      sorting: [{ field: START_DATE, direction: DESC }]
    ) {
      data {
        assignmentId
        positionCode
        positionRecordId
        employeeId
        employeeName
        employeeNumber
        assignmentType
        assignmentStatus
        fte
        startDate
        endDate
        isCurrent
        notes
        createdAt
        updatedAt
      }
    }
    positionTransfers(
      positionCode: $code
      pagination: { page: 1, pageSize: 50 }
    ) {
      data {
        transferId
        positionCode
        fromOrganizationCode
        toOrganizationCode
        effectiveDate
        initiatedBy {
          id
          name
        }
        operationReason
        createdAt
      }
    }
  }
`;

const normalizeString = (value?: string | null): string | undefined => {
  if (typeof value !== 'string') {
    return undefined;
  }
  const trimmed = value.trim();
  return trimmed === '' ? undefined : trimmed;
};

const normalizeUppercase = (value?: string | null): string | undefined => {
  const normalized = normalizeString(value);
  return normalized ? normalized.toUpperCase() : undefined;
};

const normalizePositionParams = (
  params: PositionQueryParams = {},
): NormalizedPositionQueryParams => {
  const page = Math.max(DEFAULT_PAGE, Math.floor(params.page ?? DEFAULT_PAGE));
  const rawPageSize = Math.floor(params.pageSize ?? DEFAULT_PAGE_SIZE);
  const pageSize = Math.max(1, Math.min(rawPageSize, MAX_PAGE_SIZE));

  return {
    page,
    pageSize,
    status: normalizeUppercase(params.status),
    jobFamilyGroupCode: normalizeUppercase(params.jobFamilyGroupCode),
    jobFamilyCode: normalizeUppercase(params.jobFamilyCode),
    jobRoleCode: normalizeUppercase(params.jobRoleCode),
    jobLevelCode: normalizeUppercase(params.jobLevelCode),
    positionType: normalizeUppercase(params.positionType),
    employmentType: normalizeUppercase(params.employmentType),
  };
};

const buildGraphQLVariables = (params: NormalizedPositionQueryParams) => {
  const filter: Record<string, unknown> = {};

  if (params.status) {
    filter.status = params.status;
  }
  if (params.jobFamilyGroupCode) {
    filter.jobFamilyGroupCodes = [params.jobFamilyGroupCode];
  }
  if (params.jobFamilyCode) {
    filter.jobFamilyCodes = [params.jobFamilyCode];
  }
  if (params.jobRoleCode) {
    filter.jobRoleCodes = [params.jobRoleCode];
  }
  if (params.jobLevelCode) {
    filter.jobLevelCodes = [params.jobLevelCode];
  }
  if (params.positionType) {
    filter.positionTypes = [params.positionType];
  }
  if (params.employmentType) {
    filter.employmentTypes = [params.employmentType];
  }

  return {
    filter: Object.keys(filter).length > 0 ? filter : undefined,
    pagination: {
      page: params.page,
      pageSize: params.pageSize,
      sortBy: 'code',
      sortOrder: 'asc',
    },
  };
};

const transformPositionNode = (node: PositionNodeResponse): PositionRecord => {
  const availableHeadcount =
    typeof node.availableHeadcount === 'number'
      ? node.availableHeadcount
      : Math.max(node.headcountCapacity - node.headcountInUse, 0);

  return {
    code: node.code,
    title: node.title,
    jobFamilyGroupCode: node.jobFamilyGroupCode,
    jobFamilyGroupName: node.jobFamilyGroupName ?? undefined,
    jobFamilyCode: node.jobFamilyCode,
    jobFamilyName: node.jobFamilyName ?? undefined,
    jobRoleCode: node.jobRoleCode,
    jobRoleName: node.jobRoleName ?? undefined,
    jobLevelCode: node.jobLevelCode,
    jobLevelName: node.jobLevelName ?? undefined,
    organizationCode: node.organizationCode,
    organizationName: node.organizationName ?? undefined,
    positionType: node.positionType,
    employmentType: node.employmentType,
    headcountCapacity: node.headcountCapacity,
    headcountInUse: node.headcountInUse,
    availableHeadcount,
    gradeLevel: node.gradeLevel ?? undefined,
    reportsToPositionCode: normalizeString(node.reportsToPositionCode),
    status: node.status,
    effectiveDate: node.effectiveDate,
    endDate: node.endDate ?? undefined,
    isCurrent: node.isCurrent,
    isFuture: node.isFuture,
    createdAt: node.createdAt,
    updatedAt: node.updatedAt,
  };
};

const transformAssignmentNode = (node: PositionAssignmentGraphQLNode): PositionAssignmentRecord => ({
  assignmentId: node.assignmentId,
  positionCode: node.positionCode,
  positionRecordId: node.positionRecordId ?? undefined,
  employeeId: node.employeeId,
  employeeName: node.employeeName,
  employeeNumber: node.employeeNumber ?? undefined,
  assignmentType: node.assignmentType,
  assignmentStatus: node.assignmentStatus,
  fte: node.fte,
  startDate: node.startDate,
  endDate: node.endDate ?? undefined,
  isCurrent: node.isCurrent,
  notes: node.notes ?? undefined,
  createdAt: node.createdAt,
  updatedAt: node.updatedAt,
});

const transformTransferNode = (node: PositionTransferGraphQLNode): PositionTransferRecord => ({
  transferId: node.transferId,
  positionCode: node.positionCode,
  fromOrganizationCode: node.fromOrganizationCode,
  toOrganizationCode: node.toOrganizationCode,
  effectiveDate: node.effectiveDate,
  initiatedBy: {
    id: node.initiatedBy?.id ?? '',
    name: node.initiatedBy?.name ?? '',
  },
  operationReason: node.operationReason ?? undefined,
  createdAt: node.createdAt,
});

const transformTimelineEntry = (entry: PositionTimelineResponse): PositionTimelineEvent => ({
  id: entry.recordId,
  status: (entry.status ?? '').toUpperCase(),
  title: entry.title,
  effectiveDate: entry.effectiveDate,
  endDate: entry.endDate ?? undefined,
  changeReason: entry.changeReason ?? undefined,
  isCurrent: entry.isCurrent,
});

const fetchPositionsWithParams = async (
  params: NormalizedPositionQueryParams,
  signal?: AbortSignal,
): Promise<PositionsQueryResult> => {
  const response = await graphqlEnterpriseAdapter.request<PositionsGraphQLResponse>(
    POSITIONS_QUERY_DOCUMENT,
    buildGraphQLVariables(params),
    { signal },
  );

  if (!response.success || !response.data) {
    throw createQueryError(response.error?.message ?? '获取职位列表失败', {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    });
  }

  const payload = response.data.positions;
  const positions = (payload?.data ?? []).map(transformPositionNode);

  return {
    positions,
    pagination: {
      total: payload?.pagination?.total ?? positions.length,
      page: payload?.pagination?.page ?? params.page,
      pageSize: payload?.pagination?.pageSize ?? params.pageSize,
      hasNext:
        payload?.pagination?.hasNext ??
        ((payload?.pagination?.page ?? params.page) *
          (payload?.pagination?.pageSize ?? params.pageSize) <
          (payload?.pagination?.total ?? positions.length)),
      hasPrevious:
        payload?.pagination?.hasPrevious ??
        ((payload?.pagination?.page ?? params.page) > 1),
    },
    totalCount: payload?.totalCount ?? positions.length,
    timestamp: response.timestamp ?? new Date().toISOString(),
  };
};

const fetchPositionDetail = async (
  code: string,
  signal?: AbortSignal,
): Promise<PositionDetailResult> => {
  const response = await graphqlEnterpriseAdapter.request<PositionDetailGraphQLResponse>(
    POSITION_DETAIL_QUERY_DOCUMENT,
    { code },
    { signal },
  );

  if (!response.success || !response.data) {
    throw createQueryError(response.error?.message ?? '获取职位详情失败', {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    });
  }

  if (!response.data.position) {
    throw createQueryError('未找到指定职位', {
      code: 'POSITION_NOT_FOUND',
      requestId: response.requestId,
    });
  }

  const position = transformPositionNode(response.data.position);
  const timeline = (response.data.positionTimeline ?? []).map(transformTimelineEntry);
  const assignments = (response.data.positionAssignments?.data ?? []).map(transformAssignmentNode);
  const transfers = (response.data.positionTransfers?.data ?? []).map(transformTransferNode);

  let currentAssignment: PositionAssignmentRecord | null = null;
  if (response.data.position.currentAssignment) {
    currentAssignment = transformAssignmentNode(response.data.position.currentAssignment);
  } else {
    currentAssignment = assignments.find(item => item.isCurrent) ?? null;
  }

  return {
    position,
    timeline,
    currentAssignment: currentAssignment ?? null,
    assignments,
    transfers,
    fetchedAt: response.timestamp ?? new Date().toISOString(),
  };
};

export const POSITIONS_QUERY_ROOT_KEY = ['enterprise-positions'] as const;
export const POSITION_DETAIL_QUERY_ROOT_KEY = ['enterprise-position-detail'] as const;

export const positionsQueryKey = (params: NormalizedPositionQueryParams) =>
  [...POSITIONS_QUERY_ROOT_KEY, params] as const;

export const positionDetailQueryKey = (code: string) =>
  [...POSITION_DETAIL_QUERY_ROOT_KEY, code] as const;

type PositionsQueryKey = ReturnType<typeof positionsQueryKey>;
type PositionDetailQueryKey = ReturnType<typeof positionDetailQueryKey>;

const positionsQueryFn = async ({
  queryKey,
  signal,
}: QueryFunctionContext<PositionsQueryKey>): Promise<PositionsQueryResult> => {
  const [, params] = queryKey;
  return fetchPositionsWithParams(params, signal);
};

const positionDetailQueryFn = async ({
  queryKey,
  signal,
}: QueryFunctionContext<PositionDetailQueryKey>): Promise<PositionDetailResult> => {
  const [, code] = queryKey;
  return fetchPositionDetail(code, signal);
};

export function useEnterprisePositions(
  params: PositionQueryParams = {},
): UseQueryResult<PositionsQueryResult> {
  const serialized = JSON.stringify(params ?? {});
  const normalizedParams = useMemo(
    () => normalizePositionParams(params),
    [serialized],
  );

  return useQuery({
    queryKey: positionsQueryKey(normalizedParams),
    queryFn: positionsQueryFn,
    staleTime: 60_000,
    keepPreviousData: true,
  });
}

export interface PositionDetailOptions {
  enabled?: boolean;
}

export function usePositionDetail(
  code: string | undefined,
  options?: PositionDetailOptions,
): UseQueryResult<PositionDetailResult> {
  const enabled = Boolean(code) && (options?.enabled ?? true);
  const queryKey = code ? positionDetailQueryKey(code) : positionDetailQueryKey('placeholder');

  return useQuery({
    queryKey,
    queryFn: positionDetailQueryFn,
    enabled,
    staleTime: 60_000,
  });
}

const defaultExport = useEnterprisePositions;
export default defaultExport;

export const __internal = {
  DEFAULT_PAGE,
  DEFAULT_PAGE_SIZE,
  MAX_PAGE_SIZE,
  normalizePositionParams,
  buildGraphQLVariables,
  transformPositionNode,
  transformAssignmentNode,
  transformTransferNode,
  transformTimelineEntry,
  fetchPositionsWithParams,
  fetchPositionDetail,
};
