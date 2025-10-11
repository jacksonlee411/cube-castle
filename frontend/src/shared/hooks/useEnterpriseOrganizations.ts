import { useCallback, useMemo, useState } from 'react';
import type { QueryFunctionContext } from '@tanstack/react-query';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { graphqlEnterpriseAdapter } from '../api/graphql-enterprise-adapter';
import { createQueryError, type QueryErrorDetail } from '../api/queryClient';
import type { OrganizationQueryParams, OrganizationUnit } from '../types/organization';
import type { APIResponse } from '../types/api';
import type { JsonValue } from '../types/json';
import {
  convertGraphQLToOrganizationUnit,
  type GraphQLOrganizationData,
} from '../types/converters';
import {
  OrganizationUnitTypeEnum,
  OrganizationStatusEnum,
  isOrganizationUnitTypeEnum,
  isOrganizationStatusEnum,
} from '../types/contract_gen';

const DEFAULT_PAGE = 1;
const DEFAULT_PAGE_SIZE = 50;
const MAX_PAGE_SIZE = 1000;

export interface OrganizationStats {
  totalCount: number;
  activeCount: number;
  inactiveCount: number;
  plannedCount: number;
  deletedCount: number;
  byType: Array<{ unitType: OrganizationUnitTypeEnum; count: number }>;
  byStatus: Array<{ status: OrganizationStatusEnum; count: number }>;
  byLevel: Array<{ level: number; count: number }>;
  temporalStats: {
    totalVersions: number;
    averageVersionsPerOrg: number;
    oldestEffectiveDate: string | null;
    newestEffectiveDate: string | null;
  };
}

interface GraphQLOrganizationStats {
  totalCount: number;
  activeCount: number;
  inactiveCount: number;
  plannedCount: number;
  deletedCount: number;
  byType: Array<{ unitType: string; count: number }>;
  byStatus: Array<{ status: string; count: number }>;
  byLevel: Array<{ level: number; count: number }>;
  temporalStats: {
    totalVersions: number;
    averageVersionsPerOrg: number;
    oldestEffectiveDate: string | null;
    newestEffectiveDate: string | null;
  };
}

export interface OrganizationTemporalSummary {
  asOfDate?: string | null;
  currentCount?: number | null;
  futureCount?: number | null;
  historicalCount?: number | null;
}

interface OrganizationsGraphQLResponse {
  organizations: {
    data: GraphQLOrganizationData[];
    pagination: {
      total: number;
      page: number;
      pageSize: number;
      hasNext: boolean;
      hasPrevious: boolean;
    };
    temporal?: OrganizationTemporalSummary;
  };
  organizationStats?: GraphQLOrganizationStats | null;
}

interface OrganizationByCodeResponse {
  organization: GraphQLOrganizationData | null;
}

export interface OrganizationsQueryResult {
  organizations: OrganizationUnit[];
  pagination: {
    total: number;
    page: number;
    pageSize: number;
    totalPages: number;
    hasNext: boolean;
    hasPrevious: boolean;
  };
  stats: OrganizationStats | null;
  temporal?: OrganizationTemporalSummary;
  timestamp: string;
}

export interface NormalizedQueryParams {
  page: number;
  pageSize: number;
  searchText?: string;
  unitType?: OrganizationUnitTypeEnum;
  status?: OrganizationStatusEnum;
  parentCode?: string;
  level?: number;
  asOfDate?: string;
  includeHistorical: boolean;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

const DEFAULT_QUERY_PARAMS: NormalizedQueryParams = {
  page: DEFAULT_PAGE,
  pageSize: DEFAULT_PAGE_SIZE,
  includeHistorical: false,
};

const ORGANIZATIONS_QUERY_DOCUMENT = /* GraphQL */ `
  query EnterpriseOrganizations(
    $filter: OrganizationFilter
    $pagination: PaginationInput
    $statsAsOfDate: String
    $statsIncludeHistorical: Boolean!
  ) {
    organizations(filter: $filter, pagination: $pagination) {
      data {
        code
        parentCode
        tenantId
        name
        unitType
        status
        level
        codePath
        namePath
        path
        sortOrder
        description
        profile
        effectiveDate
        endDate
        createdAt
        updatedAt
        recordId
        isFuture
        hierarchyDepth
        childrenCount
        changeReason
        deletedAt
        deletedBy
        deletionReason
        suspendedAt
        suspendedBy
        suspensionReason
      }
      pagination {
        total
        page
        pageSize
        hasNext
        hasPrevious
      }
      temporal {
        asOfDate
        currentCount
        futureCount
        historicalCount
      }
    }
    organizationStats(
      asOfDate: $statsAsOfDate
      includeHistorical: $statsIncludeHistorical
    ) {
      totalCount
      activeCount
      inactiveCount
      plannedCount
      deletedCount
      byType {
        unitType
        count
      }
      byStatus {
        status
        count
      }
      byLevel {
        level
        count
      }
      temporalStats {
        totalVersions
        averageVersionsPerOrg
        oldestEffectiveDate
        newestEffectiveDate
      }
    }
  }
`;

const ORGANIZATION_BY_CODE_DOCUMENT = /* GraphQL */ `
  query OrganizationByCode($code: String!, $asOfDate: String) {
    organization(code: $code, asOfDate: $asOfDate) {
      code
      parentCode
      tenantId
      name
      unitType
      status
      level
      codePath
      namePath
      path
      sortOrder
      description
      profile
      effectiveDate
      endDate
      createdAt
      updatedAt
      recordId
      changeReason
      deletedAt
      deletedBy
      deletionReason
      suspendedAt
      suspendedBy
      suspensionReason
      isFuture
      hierarchyDepth
      childrenCount
    }
  }
`;

const clamp = (value: number, min: number, max: number): number => {
  return Math.min(Math.max(value, min), max);
};

const normalizeQueryParams = (
  params?: Partial<OrganizationQueryParams | NormalizedQueryParams>,
): NormalizedQueryParams => {
  const base = params ?? {};
  const page = clamp(
    typeof base.page === 'number' ? Math.trunc(base.page) : DEFAULT_PAGE,
    1,
    Number.MAX_SAFE_INTEGER,
  );
  const pageSize = clamp(
    typeof base.pageSize === 'number' ? Math.trunc(base.pageSize) : DEFAULT_PAGE_SIZE,
    1,
    MAX_PAGE_SIZE,
  );

  const normalized: NormalizedQueryParams = {
    ...DEFAULT_QUERY_PARAMS,
    page,
    pageSize,
    includeHistorical: Boolean(base.includeHistorical),
  };

  const searchText = typeof base.searchText === 'string' ? base.searchText.trim() : '';
  if (searchText) {
    normalized.searchText = searchText;
  }

  const unitType = base.unitType as unknown;
  if (typeof unitType === 'string' && isOrganizationUnitTypeEnum(unitType)) {
    normalized.unitType = unitType;
  }

  const status = base.status as unknown;
  if (typeof status === 'string' && isOrganizationStatusEnum(status)) {
    normalized.status = status;
  }

  if (typeof base.parentCode === 'string' && base.parentCode.trim()) {
    normalized.parentCode = base.parentCode.trim();
  }

  if (typeof base.level === 'number' && Number.isFinite(base.level)) {
    normalized.level = Math.trunc(base.level);
  }

  const asOfDate =
    typeof base.asOfDate === 'string'
      ? base.asOfDate
      : typeof (base as OrganizationQueryParams).effectiveDate === 'string'
        ? (base as OrganizationQueryParams).effectiveDate
        : undefined;
  if (asOfDate) {
    normalized.asOfDate = asOfDate;
  }

  if (typeof base.sortBy === 'string' && base.sortBy.trim()) {
    normalized.sortBy = base.sortBy.trim();
  }

  if (typeof base.sortOrder === 'string') {
    const lower = base.sortOrder.toLowerCase();
    normalized.sortOrder = lower === 'desc' ? 'desc' : 'asc';
  }

  return normalized;
};

const mergeQueryParams = (
  current: NormalizedQueryParams,
  patch?: Partial<OrganizationQueryParams>,
): NormalizedQueryParams => normalizeQueryParams({ ...current, ...(patch ?? {}) });

const buildGraphQLVariables = (params: NormalizedQueryParams) => {
  const filter: Record<string, JsonValue> = {};

  if (params.asOfDate) {
    filter.asOfDate = params.asOfDate;
  }
  if (params.searchText) {
    filter.searchText = params.searchText;
  }
  if (params.unitType) {
    filter.unitType = params.unitType;
  }
  if (params.status) {
    filter.status = params.status;
  }
  if (params.parentCode) {
    filter.parentCode = params.parentCode;
  }
  if (typeof params.level === 'number') {
    filter.level = params.level;
  }
  if (params.includeHistorical) {
    filter.includeFuture = true;
  }

  const pagination: {
    page: number;
    pageSize: number;
    sortBy?: string;
    sortOrder?: string;
  } = {
    page: params.page,
    pageSize: params.pageSize,
  };

  if (params.sortBy) {
    pagination.sortBy = params.sortBy;
  }

  if (params.sortOrder) {
    pagination.sortOrder = params.sortOrder;
  }

  const variables: Record<string, JsonValue> = {
    pagination,
    statsAsOfDate: params.asOfDate ?? null,
    statsIncludeHistorical: params.includeHistorical,
  };

  if (Object.keys(filter).length > 0) {
    variables.filter = filter;
  }

  return variables;
};

const mapOrganizationStats = (
  stats?: GraphQLOrganizationStats | null,
): OrganizationStats | null => {
  if (!stats) {
    return null;
  }

  const byType =
    stats.byType?.reduce<Array<{ unitType: OrganizationUnitTypeEnum; count: number }>>(
      (acc, item) => {
        if (item && typeof item.unitType === 'string' && isOrganizationUnitTypeEnum(item.unitType)) {
          acc.push({ unitType: item.unitType, count: item.count ?? 0 });
        }
        return acc;
      },
      [],
    ) ?? [];

  const byStatus =
    stats.byStatus?.reduce<Array<{ status: OrganizationStatusEnum; count: number }>>(
      (acc, item) => {
        if (item && typeof item.status === 'string' && isOrganizationStatusEnum(item.status)) {
          acc.push({ status: item.status, count: item.count ?? 0 });
        }
        return acc;
      },
      [],
    ) ?? [];

  const byLevel =
    stats.byLevel?.map(({ level, count }) => ({
      level: Number.isFinite(level) ? level : 0,
      count: count ?? 0,
    })) ?? [];

  return {
    totalCount: stats.totalCount ?? 0,
    activeCount: stats.activeCount ?? 0,
    inactiveCount: stats.inactiveCount ?? 0,
    plannedCount: stats.plannedCount ?? 0,
    deletedCount: stats.deletedCount ?? 0,
    byType,
    byStatus,
    byLevel,
    temporalStats: {
      totalVersions: stats.temporalStats?.totalVersions ?? 0,
      averageVersionsPerOrg: stats.temporalStats?.averageVersionsPerOrg ?? 0,
      oldestEffectiveDate: stats.temporalStats?.oldestEffectiveDate ?? null,
      newestEffectiveDate: stats.temporalStats?.newestEffectiveDate ?? null,
    },
  };
};

const transformOrganizationsResponse = (
  payload: OrganizationsGraphQLResponse,
  params: NormalizedQueryParams,
  timestamp: string,
): OrganizationsQueryResult => {
  const list = payload.organizations?.data ?? [];
  const mappedOrganizations = list.map(convertGraphQLToOrganizationUnit);

  const pagination = payload.organizations?.pagination;
  const total = pagination?.total ?? mappedOrganizations.length;
  const pageSize = pagination?.pageSize ?? params.pageSize;
  const currentPage = pagination?.page ?? params.page;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return {
    organizations: mappedOrganizations,
    pagination: {
      total,
      page: currentPage,
      pageSize,
      totalPages,
      hasNext: pagination?.hasNext ?? currentPage < totalPages,
      hasPrevious: pagination?.hasPrevious ?? currentPage > 1,
    },
    stats: mapOrganizationStats(payload.organizationStats),
    temporal: payload.organizations?.temporal,
    timestamp,
  };
};

const fetchOrganizationsWithParams = async (
  params: NormalizedQueryParams,
  signal?: AbortSignal,
): Promise<OrganizationsQueryResult> => {
  const variables = buildGraphQLVariables(params);
  const response = await graphqlEnterpriseAdapter.request<OrganizationsGraphQLResponse>(
    ORGANIZATIONS_QUERY_DOCUMENT,
    variables,
    { signal },
  );

  if (!response.success || !response.data) {
    throw createQueryError(response.error?.message ?? '组织数据获取失败', {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    });
  }

  return transformOrganizationsResponse(response.data, params, response.timestamp);
};

const fetchOrganizationDetail = async (
  code: string,
  signal?: AbortSignal,
  asOfDate?: string,
): Promise<OrganizationUnit | null> => {
  const response = await graphqlEnterpriseAdapter.request<OrganizationByCodeResponse>(
    ORGANIZATION_BY_CODE_DOCUMENT,
    {
      code,
      asOfDate: asOfDate ?? null,
    },
    { signal },
  );

  if (!response.success) {
    throw createQueryError(response.error?.message ?? '获取组织详情失败', {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    });
  }

  const organization = response.data?.organization ?? null;
  return organization ? convertGraphQLToOrganizationUnit(organization) : null;
};

export const ORGANIZATIONS_QUERY_ROOT_KEY = ['organizations'] as const;
const ORGANIZATION_DETAIL_KEY = 'organization-detail';

export const organizationsQueryKey = (params: NormalizedQueryParams) =>
  [...ORGANIZATIONS_QUERY_ROOT_KEY, params] as const;

export const organizationByCodeQueryKey = (code: string, asOfDate?: string) =>
  [ORGANIZATION_DETAIL_KEY, code, asOfDate ?? null] as const;

type OrganizationsQueryKey = ReturnType<typeof organizationsQueryKey>;
type OrganizationByCodeQueryKey = ReturnType<typeof organizationByCodeQueryKey>;

const organizationsQueryFn = async ({
  queryKey,
  signal,
}: QueryFunctionContext<OrganizationsQueryKey>): Promise<OrganizationsQueryResult> => {
  const [, params] = queryKey;
  return fetchOrganizationsWithParams(params, signal);
};

const organizationByCodeQueryFn = async ({
  queryKey,
  signal,
}: QueryFunctionContext<OrganizationByCodeQueryKey>): Promise<OrganizationUnit | null> => {
  const [, code, asOfDateValue] = queryKey;
  return fetchOrganizationDetail(code, signal, asOfDateValue ?? undefined);
};

export interface UseEnterpriseOrganizationsResult {
  organizations: OrganizationUnit[];
  totalCount: number;
  page: number;
  pageSize: number;
  totalPages: number;
  hasNext: boolean;
  hasPrevious: boolean;
  stats: OrganizationStats | null;
  temporal?: OrganizationTemporalSummary;
  loading: boolean;
  isFetching: boolean;
  error: string | null;
  lastUpdate: string | null;
  queryParams: NormalizedQueryParams;
  setQueryParams: (patch: Partial<OrganizationQueryParams>) => void;
  fetchOrganizations: (
    patch?: Partial<OrganizationQueryParams>,
  ) => Promise<APIResponse<OrganizationsQueryResult>>;
  fetchOrganizationByCode: (
    code: string,
    options?: { asOfDate?: string },
  ) => Promise<OrganizationUnit | null>;
  fetchStats: () => Promise<APIResponse<OrganizationStats>>;
  refreshData: () => Promise<OrganizationsQueryResult | undefined>;
  clearError: () => void;
}

/* c8 ignore start */
export const useEnterpriseOrganizations = (
  initialParams?: OrganizationQueryParams,
): UseEnterpriseOrganizationsResult => {
  const [queryParams, setQueryParamsState] = useState<NormalizedQueryParams>(() =>
    normalizeQueryParams(initialParams),
  );
  const queryClient = useQueryClient();

  const queryKey = useMemo(() => organizationsQueryKey(queryParams), [queryParams]);

  const queryResult = useQuery<
    OrganizationsQueryResult,
    Error,
    OrganizationsQueryResult,
    OrganizationsQueryKey
  >({
    queryKey,
    queryFn: organizationsQueryFn,
    placeholderData: (previousData) => previousData,
    staleTime: 5 * 60 * 1000,
  });

  const updateQueryParams = useCallback(
    (patch: Partial<OrganizationQueryParams>) => {
      setQueryParamsState((current) => mergeQueryParams(current, patch));
    },
    [],
  );

  const fetchOrganizations = useCallback(
    async (
      patch?: Partial<OrganizationQueryParams>,
    ): Promise<APIResponse<OrganizationsQueryResult>> => {
      const nextParams = mergeQueryParams(queryParams, patch);
      const nextKey = organizationsQueryKey(nextParams);
      const data = await queryClient.fetchQuery<
        OrganizationsQueryResult,
        Error,
        OrganizationsQueryResult,
        OrganizationsQueryKey
      >({
        queryKey: nextKey,
        queryFn: organizationsQueryFn,
      });
      setQueryParamsState(nextParams);
      return {
        success: true,
        data,
        message: '组织数据获取成功',
        timestamp: data.timestamp,
      };
    },
    [queryClient, queryParams],
  );

  const fetchStats = useCallback(async (): Promise<APIResponse<OrganizationStats>> => {
    const data = await queryClient.fetchQuery<
      OrganizationsQueryResult,
      Error,
      OrganizationsQueryResult,
      OrganizationsQueryKey
    >({
      queryKey,
      queryFn: organizationsQueryFn,
    });

    if (!data.stats) {
      return {
        success: false,
        error: {
          code: 'NO_STATS_AVAILABLE',
          message: '暂无组织统计信息',
        },
        timestamp: data.timestamp,
      };
    }

    return {
      success: true,
      data: data.stats,
      message: '组织统计获取成功',
      timestamp: data.timestamp,
    };
  }, [queryClient, queryKey]);

  const fetchOrganizationByCode = useCallback(
    (code: string, options?: { asOfDate?: string }) =>
      queryClient.fetchQuery<
        OrganizationUnit | null,
        Error,
        OrganizationUnit | null,
        OrganizationByCodeQueryKey
      >({
        queryKey: organizationByCodeQueryKey(code, options?.asOfDate ?? queryParams.asOfDate),
        queryFn: organizationByCodeQueryFn,
      }),
    [queryClient, queryParams.asOfDate],
  );

  const refreshData = useCallback(
    async (): Promise<OrganizationsQueryResult | undefined> => {
      const result = await queryResult.refetch({
        throwOnError: false,
        cancelRefetch: false,
      });
      return result.data;
    },
    [queryResult],
  );

  const clearError = useCallback(() => {
    queryClient.resetQueries({ queryKey, exact: true });
  }, [queryClient, queryKey]);

  const organizations = queryResult.data?.organizations ?? [];
  const pagination = queryResult.data?.pagination ?? {
    total: 0,
    page: queryParams.page,
    pageSize: queryParams.pageSize,
    totalPages: 1,
    hasNext: false,
    hasPrevious: false,
  };

  const errorMessage = useMemo(() => {
    if (!queryResult.error) {
      return null;
    }
    const err = queryResult.error as Error & QueryErrorDetail;
    const requestId = err.requestId ? ` (请求ID: ${err.requestId})` : '';
    return `${err.message}${requestId}`;
  }, [queryResult.error]);

  return {
    organizations,
    totalCount: pagination.total,
    page: pagination.page,
    pageSize: pagination.pageSize,
    totalPages: pagination.totalPages,
    hasNext: pagination.hasNext,
    hasPrevious: pagination.hasPrevious,
    stats: queryResult.data?.stats ?? null,
    temporal: queryResult.data?.temporal,
    loading: queryResult.isPending,
    isFetching: queryResult.isFetching,
    error: errorMessage,
    lastUpdate: queryResult.data?.timestamp ?? null,
    queryParams,
    setQueryParams: updateQueryParams,
    fetchOrganizations,
    fetchOrganizationByCode,
    fetchStats,
    refreshData,
    clearError,
  };
};

export default useEnterpriseOrganizations;
/* c8 ignore end */

export const __internal = {
  normalizeQueryParams,
  mergeQueryParams,
  buildGraphQLVariables,
  transformOrganizationsResponse,
  mapOrganizationStats,
  fetchOrganizationsWithParams,
  fetchOrganizationDetail,
  organizationsQueryKey,
  organizationByCodeQueryKey,
};
