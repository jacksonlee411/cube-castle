import { useMemo } from 'react';
import type { QueryFunctionContext, UseQueryResult } from '@tanstack/react-query';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import { graphqlEnterpriseAdapter } from '../api/graphql-enterprise-adapter';
import { createQueryError } from '../api/queryClient';
import type {
  PositionAssignmentAuditRecord,
  PositionAssignmentRecord,
  PositionDetailResult,
  PositionRecord,
  PositionStatus,
  PositionTimelineEvent,
  PositionTransferRecord,
  PositionsQueryResult,
  PositionHeadcountStats,
  VacantPositionRecord,
  VacantPositionsQueryResult,
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
  effectiveDate: string;
  endDate?: string | null;
  actingUntil?: string | null;
  autoRevert?: boolean | null;
  reminderSentAt?: string | null;
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
  recordId?: string;
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
  timelineCategory?: string | null;
  assignmentType?: string | null;
  assignmentStatus?: string | null;
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
  positionVersions: PositionNodeResponse[];
}

interface PositionAssignmentsGraphQLResponse {
  positionAssignments: {
    data: PositionAssignmentGraphQLNode[];
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

interface PositionAssignmentAuditGraphQLNode {
  assignmentId: string;
  eventType: string;
  effectiveDate: string;
  endDate?: string | null;
  actor: string;
  changes?: Record<string, unknown> | null;
  createdAt: string;
}

interface PositionAssignmentAuditGraphQLResponse {
  positionAssignmentAudit: {
    data: PositionAssignmentAuditGraphQLNode[];
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

interface VacantPositionGraphQLNode {
  positionCode: string;
  organizationCode: string;
  organizationName?: string | null;
  jobFamilyCode: string;
  jobRoleCode: string;
  jobLevelCode: string;
  vacantSince: string;
  headcountCapacity: number;
  headcountAvailable: number;
  totalAssignments: number;
}

interface VacantPositionsGraphQLResponse {
  vacantPositions: {
    data: VacantPositionGraphQLNode[];
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

interface PositionHeadcountStatsGraphQLResponse {
  positionHeadcountStats: {
    organizationCode: string;
    organizationName: string;
    totalCapacity: number;
    totalFilled: number;
    totalAvailable: number;
    fillRate: number;
    byLevel: Array<{
      jobLevelCode: string;
      capacity: number;
      utilized: number;
      available: number;
    }>;
    byType: Array<{
      positionType: string;
      capacity: number;
      filled: number;
      available: number;
    }>;
    byFamily: Array<{
      jobFamilyCode: string;
      jobFamilyName?: string | null;
      capacity: number;
      utilized: number;
      available: number;
    }>;
  };
}

export type VacantPositionSortField = 'VACANT_SINCE' | 'HEADCOUNT_AVAILABLE' | 'HEADCOUNT_CAPACITY';

export interface VacantPositionsQueryParams {
  organizationCodes?: string[];
  jobFamilyCodes?: string[];
  jobRoleCodes?: string[];
  jobLevelCodes?: string[];
  positionTypes?: string[];
  minimumVacantDays?: number;
  asOfDate?: string;
  page?: number;
  pageSize?: number;
  sortField?: VacantPositionSortField;
  sortDirection?: 'ASC' | 'DESC';
}

export interface PositionHeadcountStatsParams {
  organizationCode: string;
  includeSubordinates?: boolean;
}

interface NormalizedVacantPositionsQueryParams {
  page: number;
  pageSize: number;
  organizationCodes?: string[];
  jobFamilyCodes?: string[];
  jobRoleCodes?: string[];
  jobLevelCodes?: string[];
  positionTypes?: string[];
  minimumVacantDays?: number;
  asOfDate?: string;
  sortField?: VacantPositionSortField;
  sortDirection: 'ASC' | 'DESC';
}

interface NormalizedPositionHeadcountParams {
  organizationCode: string;
  includeSubordinates: boolean;
}

export interface PositionAssignmentsQueryParams {
  page?: number;
  pageSize?: number;
  assignmentTypes?: string[];
  status?: string;
  dateFrom?: string;
  dateTo?: string;
  includeHistorical?: boolean;
  includeActingOnly?: boolean;
}

interface NormalizedPositionAssignmentsQueryParams {
  positionCode: string;
  page: number;
  pageSize: number;
  assignmentTypes?: string[];
  status?: string;
  dateFrom?: string;
  dateTo?: string;
  includeHistorical: boolean;
  includeActingOnly: boolean;
}

const DISABLED_ASSIGNMENT_QUERY_PARAMS: NormalizedPositionAssignmentsQueryParams = {
  positionCode: '__disabled__',
  page: 1,
  pageSize: 1,
  includeHistorical: false,
  includeActingOnly: false,
};

export interface PositionAssignmentsQueryResult {
  data: PositionAssignmentRecord[];
  pagination: {
    total: number;
    page: number;
    pageSize: number;
    hasNext: boolean;
    hasPrevious: boolean;
  };
  totalCount: number;
  fetchedAt: string;
}

export interface PositionAssignmentAuditQueryResult {
  records: PositionAssignmentAuditRecord[];
  pagination: {
    total: number;
    page: number;
    pageSize: number;
    hasNext: boolean;
    hasPrevious: boolean;
  };
  totalCount: number;
  fetchedAt: string;
}

type JsonPrimitive = string | number | boolean | null
type JsonValue = JsonPrimitive | JsonValue[] | { [key: string]: JsonValue }

type PositionsQueryVariables = {
  pagination: {
    page: number;
    pageSize: number;
    sortBy?: string;
    sortOrder?: 'asc' | 'desc';
  };
  filter?: Record<string, JsonValue>;
};

type VacantPositionsQueryVariables = {
  pagination: {
    page: number;
    pageSize: number;
  };
  filter?: Record<string, JsonValue>;
  sorting?: Array<{
    field: VacantPositionSortField;
    direction: 'ASC' | 'DESC';
  }>;
};

type PositionAssignmentsQueryVariables = {
  positionCode: string;
  filter?: Record<string, JsonValue>;
  pagination: {
    page: number;
    pageSize: number;
  };
  sorting: Array<{
    field: string;
    direction: 'ASC' | 'DESC';
  }>;
};

type PositionAssignmentAuditQueryVariables = {
  positionCode: string;
  assignmentId?: string;
  dateRange?: {
    from?: string;
    to?: string;
  };
  pagination: {
    page: number;
    pageSize: number;
  };
};

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
          effectiveDate
          endDate
          actingUntil
          autoRevert
          reminderSentAt
          isCurrent
          notes
          createdAt
          updatedAt
        }
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
  query TemporalEntityDetail($code: PositionCode!, $includeDeleted: Boolean!) {
    position(code: $code) {
      code
      recordId
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
        effectiveDate
        endDate
        actingUntil
        autoRevert
        reminderSentAt
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
      timelineCategory
      assignmentType
      assignmentStatus
    }
    positionAssignments(
      positionCode: $code
      filter: { includeHistorical: true }
      pagination: { page: 1, pageSize: 50 }
      sorting: [{ field: EFFECTIVE_DATE, direction: DESC }]
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
        effectiveDate
        endDate
        actingUntil
        autoRevert
        reminderSentAt
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
    positionVersions(
      code: $code
      includeDeleted: $includeDeleted
    ) {
      recordId
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
      gradeLevel
      headcountCapacity
      headcountInUse
      availableHeadcount
      reportsToPositionCode
      status
      effectiveDate
      endDate
      isCurrent
      createdAt
      updatedAt
    }
  }
`;

const POSITION_ASSIGNMENTS_QUERY_DOCUMENT = /* GraphQL */ `
  query PositionAssignments(
    $positionCode: PositionCode!
    $filter: PositionAssignmentFilterInput
    $pagination: PaginationInput
    $sorting: [PositionAssignmentSortInput!]
  ) {
    positionAssignments(
      positionCode: $positionCode
      filter: $filter
      pagination: $pagination
      sorting: $sorting
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
        effectiveDate
        endDate
        actingUntil
        autoRevert
        reminderSentAt
        isCurrent
        notes
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

const POSITION_ASSIGNMENT_AUDIT_QUERY_DOCUMENT = /* GraphQL */ `
  query PositionAssignmentAudit(
    $positionCode: PositionCode!
    $assignmentId: UUID
    $dateRange: DateRangeInput
    $pagination: PaginationInput
  ) {
    positionAssignmentAudit(
      positionCode: $positionCode
      assignmentId: $assignmentId
      dateRange: $dateRange
      pagination: $pagination
    ) {
      data {
        assignmentId
        eventType
        effectiveDate
        endDate
        actor
        changes
        createdAt
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

const VACANT_POSITIONS_QUERY_DOCUMENT = /* GraphQL */ `
  query VacantPositions(
    $filter: VacantPositionFilterInput
    $pagination: PaginationInput
    $sorting: [VacantPositionSortInput!]
  ) {
    vacantPositions(filter: $filter, pagination: $pagination, sorting: $sorting) {
      data {
        positionCode
        organizationCode
        organizationName
        jobFamilyCode
        jobRoleCode
        jobLevelCode
        vacantSince
        headcountCapacity
        headcountAvailable
        totalAssignments
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

const POSITION_HEADCOUNT_STATS_QUERY_DOCUMENT = /* GraphQL */ `
  query PositionHeadcountStats($organizationCode: String!, $includeSubordinates: Boolean) {
    positionHeadcountStats(
      organizationCode: $organizationCode
      includeSubordinates: $includeSubordinates
    ) {
      organizationCode
      organizationName
      totalCapacity
      totalFilled
      totalAvailable
      fillRate
      byLevel {
        jobLevelCode
        capacity
        utilized
        available
      }
      byType {
        positionType
        capacity
        filled
        available
      }
      byFamily {
        jobFamilyCode
        jobFamilyName
        capacity
        utilized
        available
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

const normalizeDateString = (value?: string | null): string | undefined => normalizeString(value);

const normalizePositiveInteger = (value?: number | null): number | undefined => {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    return undefined;
  }
  const integer = Math.floor(value);
  if (integer < 0) {
    return undefined;
  }
  return integer;
};

const normalizeStringArray = (
  values?: readonly (string | null | undefined)[] | null,
): string[] | undefined => {
  if (!Array.isArray(values)) {
    return undefined;
  }

  const mapped = values
    .map(item => normalizeString(item))
    .filter((item): item is string => Boolean(item));

  if (mapped.length === 0) {
    return undefined;
  }

  return Array.from(new Set(mapped));
};

const normalizeUppercaseArray = (
  values?: readonly (string | null | undefined)[] | null,
): string[] | undefined => {
  const normalized = normalizeStringArray(values);
  if (!normalized) {
    return undefined;
  }
  return normalized.map(item => item.toUpperCase());
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

const normalizeVacantPositionsParams = (
  params: VacantPositionsQueryParams = {},
): NormalizedVacantPositionsQueryParams => {
  const page = Math.max(DEFAULT_PAGE, Math.floor(params.page ?? DEFAULT_PAGE));
  const rawPageSize = Math.floor(params.pageSize ?? DEFAULT_PAGE_SIZE);
  const pageSize = Math.max(1, Math.min(rawPageSize, MAX_PAGE_SIZE));

  return {
    page,
    pageSize,
    organizationCodes: normalizeStringArray(params.organizationCodes),
    jobFamilyCodes: normalizeUppercaseArray(params.jobFamilyCodes),
    jobRoleCodes: normalizeUppercaseArray(params.jobRoleCodes),
    jobLevelCodes: normalizeUppercaseArray(params.jobLevelCodes),
    positionTypes: normalizeUppercaseArray(params.positionTypes),
    minimumVacantDays: normalizePositiveInteger(params.minimumVacantDays),
    asOfDate: normalizeDateString(params.asOfDate),
    sortField: params.sortField,
    sortDirection: params.sortDirection === 'ASC' ? 'ASC' : 'DESC',
  };
};

const normalizeHeadcountParams = (
  params: PositionHeadcountStatsParams,
): NormalizedPositionHeadcountParams => ({
  organizationCode: normalizeString(params.organizationCode) ?? '',
  includeSubordinates: params.includeSubordinates !== false,
});

const normalizeAssignmentQueryParams = (
  positionCode: string,
  params: PositionAssignmentsQueryParams = {},
): NormalizedPositionAssignmentsQueryParams => {
  const page = Math.max(DEFAULT_PAGE, Math.floor(params.page ?? DEFAULT_PAGE));
  const rawPageSize = Math.floor(params.pageSize ?? DEFAULT_PAGE_SIZE);
  const pageSize = Math.max(1, Math.min(rawPageSize, MAX_PAGE_SIZE));

  return {
    positionCode,
    page,
    pageSize,
    assignmentTypes: normalizeUppercaseArray(params.assignmentTypes),
    status: normalizeUppercase(params.status),
    dateFrom: normalizeDateString(params.dateFrom),
    dateTo: normalizeDateString(params.dateTo),
    includeHistorical: params.includeHistorical !== false,
    includeActingOnly: params.includeActingOnly === true,
  };
};

export interface PositionAssignmentAuditParams {
  assignmentId?: string;
  dateFrom?: string;
  dateTo?: string;
  page?: number;
  pageSize?: number;
}

interface NormalizedPositionAssignmentAuditParams {
  positionCode: string;
  assignmentId?: string;
  dateFrom?: string;
  dateTo?: string;
  page: number;
  pageSize: number;
}

const normalizeAssignmentAuditParams = (
  positionCode: string,
  params: PositionAssignmentAuditParams = {},
): NormalizedPositionAssignmentAuditParams => {
  const page = Math.max(DEFAULT_PAGE, Math.floor(params.page ?? DEFAULT_PAGE));
  const rawPageSize = Math.floor(params.pageSize ?? DEFAULT_PAGE_SIZE);
  const pageSize = Math.max(1, Math.min(rawPageSize, MAX_PAGE_SIZE));

  return {
    positionCode,
    assignmentId: normalizeString(params.assignmentId),
    dateFrom: normalizeDateString(params.dateFrom),
    dateTo: normalizeDateString(params.dateTo),
    page,
    pageSize,
  };
};

const buildGraphQLVariables = (params: NormalizedPositionQueryParams) => {
  const filter: Record<string, JsonValue> = {};

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

  const variables: PositionsQueryVariables = {
    pagination: {
      page: params.page,
      pageSize: params.pageSize,
      sortBy: 'code',
      sortOrder: 'asc',
    },
  };

  if (Object.keys(filter).length > 0) {
    variables.filter = filter;
  }

  return variables;
};

const buildVacantPositionsVariables = (params: NormalizedVacantPositionsQueryParams) => {
  const filter: Record<string, JsonValue> = {};

  if (params.organizationCodes) {
    filter.organizationCodes = params.organizationCodes;
  }
  if (params.jobFamilyCodes) {
    filter.jobFamilyCodes = params.jobFamilyCodes;
  }
  if (params.jobRoleCodes) {
    filter.jobRoleCodes = params.jobRoleCodes;
  }
  if (params.jobLevelCodes) {
    filter.jobLevelCodes = params.jobLevelCodes;
  }
  if (params.positionTypes) {
    filter.positionTypes = params.positionTypes;
  }
  if (typeof params.minimumVacantDays === 'number') {
    filter.minimumVacantDays = params.minimumVacantDays;
  }
  if (params.asOfDate) {
    filter.asOfDate = params.asOfDate;
  }

  const variables: VacantPositionsQueryVariables = {
    pagination: {
      page: params.page,
      pageSize: params.pageSize,
    },
  };

  if (Object.keys(filter).length > 0) {
    variables.filter = filter;
  }

  if (params.sortField) {
    variables.sorting = [
      {
        field: params.sortField,
        direction: params.sortDirection,
      },
    ];
  }

  return variables;
};

const buildAssignmentVariables = (params: NormalizedPositionAssignmentsQueryParams): PositionAssignmentsQueryVariables => {
  const filter: Record<string, JsonValue> = {};

  if (params.assignmentTypes && params.assignmentTypes.length > 0) {
    filter.assignmentTypes = params.assignmentTypes;
  }
  if (params.status) {
    filter.status = params.status;
  }
  if (params.dateFrom || params.dateTo) {
    filter.dateRange = {
      from: params.dateFrom ?? null,
      to: params.dateTo ?? null,
    };
  }
  if (!params.includeHistorical) {
    filter.includeHistorical = false;
  }
  if (params.includeActingOnly) {
    filter.includeActingOnly = true;
  }

  return {
    positionCode: params.positionCode,
    filter: Object.keys(filter).length > 0 ? filter : undefined,
    pagination: {
      page: params.page,
      pageSize: params.pageSize,
    },
    sorting: [{ field: 'EFFECTIVE_DATE', direction: 'DESC' }],
  };
};

const buildAssignmentAuditVariables = (
  params: NormalizedPositionAssignmentAuditParams,
): PositionAssignmentAuditQueryVariables => {
  const dateRange: { from?: string; to?: string } = {};
  if (params.dateFrom) {
    dateRange.from = params.dateFrom;
  }
  if (params.dateTo) {
    dateRange.to = params.dateTo;
  }

  const hasRange = Object.keys(dateRange).length > 0;

  return {
    positionCode: params.positionCode,
    assignmentId: params.assignmentId,
    dateRange: hasRange ? dateRange : undefined,
    pagination: {
      page: params.page,
      pageSize: params.pageSize,
    },
  };
};

const buildHeadcountVariables = (params: NormalizedPositionHeadcountParams) => {
  if (!params.organizationCode) {
    throw createQueryError('必须提供组织编码以获取编制统计');
  }

  return {
    organizationCode: params.organizationCode,
    includeSubordinates: params.includeSubordinates,
  };
};

const transformPositionNode = (node: PositionNodeResponse): PositionRecord => {
  const availableHeadcount =
    typeof node.availableHeadcount === 'number'
      ? node.availableHeadcount
      : Math.max(node.headcountCapacity - node.headcountInUse, 0);

  return {
    code: node.code,
    recordId: node.recordId ?? undefined,
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
    currentAssignment: node.currentAssignment ? transformAssignmentNode(node.currentAssignment) : undefined,
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
  effectiveDate: node.effectiveDate,
  endDate: node.endDate ?? undefined,
  actingUntil: node.actingUntil ?? undefined,
  autoRevert: Boolean(node.autoRevert),
  reminderSentAt: node.reminderSentAt ?? undefined,
  isCurrent: node.isCurrent,
  notes: node.notes ?? undefined,
  createdAt: node.createdAt,
  updatedAt: node.updatedAt,
});

const transformAssignmentAuditNode = (node: PositionAssignmentAuditGraphQLNode): PositionAssignmentAuditRecord => ({
  assignmentId: node.assignmentId,
  eventType: (node.eventType ?? '').toUpperCase(),
  effectiveDate: node.effectiveDate,
  endDate: node.endDate ?? undefined,
  actor: node.actor,
  changes: node.changes ?? null,
  createdAt: node.createdAt,
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

const transformVacantPositionNode = (node: VacantPositionGraphQLNode): VacantPositionRecord => ({
  positionCode: node.positionCode,
  organizationCode: node.organizationCode,
  organizationName: node.organizationName ?? undefined,
  jobFamilyCode: node.jobFamilyCode,
  jobRoleCode: node.jobRoleCode,
  jobLevelCode: node.jobLevelCode,
  vacantSince: node.vacantSince,
  headcountCapacity: node.headcountCapacity,
  headcountAvailable: node.headcountAvailable,
  totalAssignments: node.totalAssignments,
});

const fetchPositionAssignments = async (
  params: NormalizedPositionAssignmentsQueryParams,
  signal?: AbortSignal,
): Promise<PositionAssignmentsQueryResult> => {
  const response = await graphqlEnterpriseAdapter.request<PositionAssignmentsGraphQLResponse>(
    POSITION_ASSIGNMENTS_QUERY_DOCUMENT,
    buildAssignmentVariables(params),
    { signal },
  );

  if (!response.success || !response.data) {
    throw createQueryError(response.error?.message ?? '获取任职记录失败', {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    });
  }

  const payload = response.data.positionAssignments;
  const assignments = (payload?.data ?? []).map(transformAssignmentNode);
  const pagination = payload?.pagination ?? {
    total: assignments.length,
    page: params.page,
    pageSize: params.pageSize,
    hasNext: false,
    hasPrevious: params.page > 1,
  };

  return {
    data: assignments,
    pagination: {
      total: pagination.total ?? assignments.length,
      page: pagination.page ?? params.page,
      pageSize: pagination.pageSize ?? params.pageSize,
      hasNext:
        pagination.hasNext ??
        ((pagination.page ?? params.page) * (pagination.pageSize ?? params.pageSize) < (pagination.total ?? assignments.length)),
      hasPrevious: pagination.hasPrevious ?? ((pagination.page ?? params.page) > 1),
    },
    totalCount: payload?.totalCount ?? assignments.length,
    fetchedAt: response.timestamp ?? new Date().toISOString(),
  };
};

const fetchPositionAssignmentAuditInternal = async (
  params: NormalizedPositionAssignmentAuditParams,
  signal?: AbortSignal,
): Promise<PositionAssignmentAuditQueryResult> => {
  const response = await graphqlEnterpriseAdapter.request<PositionAssignmentAuditGraphQLResponse>(
    POSITION_ASSIGNMENT_AUDIT_QUERY_DOCUMENT,
    buildAssignmentAuditVariables(params),
    { signal },
  );

  if (!response.success || !response.data) {
    throw createQueryError(response.error?.message ?? '获取任职审计记录失败', {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    });
  }

  const payload = response.data.positionAssignmentAudit;
  const records = (payload?.data ?? []).map(transformAssignmentAuditNode);
  const pagination = payload?.pagination ?? {
    total: records.length,
    page: params.page,
    pageSize: params.pageSize,
    hasNext: false,
    hasPrevious: params.page > 1,
  };

  return {
    records,
    pagination: {
      total: pagination.total ?? records.length,
      page: pagination.page ?? params.page,
      pageSize: pagination.pageSize ?? params.pageSize,
      hasNext:
        pagination.hasNext ??
        ((pagination.page ?? params.page) * (pagination.pageSize ?? params.pageSize) < (pagination.total ?? records.length)),
      hasPrevious: pagination.hasPrevious ?? ((pagination.page ?? params.page) > 1),
    },
    totalCount: payload?.totalCount ?? records.length,
    fetchedAt: response.timestamp ?? new Date().toISOString(),
  };
};

const transformTimelineEntry = (entry: PositionTimelineResponse): PositionTimelineEvent => ({
  id: entry.recordId,
  status: (entry.status ?? '').toUpperCase(),
  title: entry.title,
  effectiveDate: entry.effectiveDate,
  endDate: entry.endDate ?? undefined,
  changeReason: entry.changeReason ?? undefined,
  isCurrent: entry.isCurrent,
  timelineCategory: entry.timelineCategory ?? undefined,
  assignmentType: entry.assignmentType ?? undefined,
  assignmentStatus: entry.assignmentStatus ?? undefined,
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
  includeDeleted: boolean,
  signal?: AbortSignal,
): Promise<PositionDetailResult> => {
  const response = await graphqlEnterpriseAdapter.request<PositionDetailGraphQLResponse>(
    POSITION_DETAIL_QUERY_DOCUMENT,
    { code, includeDeleted },
    { signal },
  );

  if (!response.success || !response.data) {
    // OBS: GraphQL 错误（职位详情）
    try {
      const { obs } = await import('@/shared/observability/obs');
      if (obs.enabled()) {
        const httpStatus =
          (response.error?.details as unknown as { httpStatus?: number })?.httpStatus ??
          (response.error as unknown as { status?: number })?.status ??
          0;
        obs.emit('position.graphql.error', {
          entity: 'position',
          code,
          queryName: 'TemporalEntityDetail',
          status: typeof httpStatus === 'number' ? httpStatus : 0,
        });
      }
    } catch {
      // ignore
    }
    throw createQueryError(response.error?.message ?? '获取职位详情失败', {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    });
  }

  if (!response.data.position) {
    // OBS: 未找到职位也作为失败路径记录
    try {
      const { obs } = await import('@/shared/observability/obs');
      if (obs.enabled()) {
        obs.emit('position.graphql.error', {
          entity: 'position',
          code,
          queryName: 'TemporalEntityDetail',
          status: 404,
        });
      }
    } catch {
      // ignore
    }
    throw createQueryError('未找到指定职位', {
      code: 'POSITION_NOT_FOUND',
      requestId: response.requestId,
    });
  }

  const position = transformPositionNode(response.data.position);
  const timeline = (response.data.positionTimeline ?? []).map(transformTimelineEntry);
  const assignments = (response.data.positionAssignments?.data ?? []).map(transformAssignmentNode);
  const transfers = (response.data.positionTransfers?.data ?? []).map(transformTransferNode);
  const versions = (response.data.positionVersions ?? []).map(transformPositionNode);

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
    versions,
    fetchedAt: response.timestamp ?? new Date().toISOString(),
  };
};

const fetchVacantPositions = async (
  params: NormalizedVacantPositionsQueryParams,
  signal?: AbortSignal,
): Promise<VacantPositionsQueryResult> => {
  const response = await graphqlEnterpriseAdapter.request<VacantPositionsGraphQLResponse>(
    VACANT_POSITIONS_QUERY_DOCUMENT,
    buildVacantPositionsVariables(params),
    { signal },
  );

  if (!response.success || !response.data) {
    throw createQueryError(response.error?.message ?? '获取空缺职位列表失败', {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    });
  }

  const payload = response.data.vacantPositions;
  const records = (payload?.data ?? []).map(transformVacantPositionNode);

  return {
    data: records,
    pagination: {
      total: payload?.pagination?.total ?? records.length,
      page: payload?.pagination?.page ?? params.page,
      pageSize: payload?.pagination?.pageSize ?? params.pageSize,
      hasNext:
        payload?.pagination?.hasNext ??
        ((payload?.pagination?.page ?? params.page) *
          (payload?.pagination?.pageSize ?? params.pageSize) <
          (payload?.pagination?.total ?? records.length)),
      hasPrevious:
        payload?.pagination?.hasPrevious ??
        ((payload?.pagination?.page ?? params.page) > 1),
    },
    totalCount: payload?.totalCount ?? records.length,
    fetchedAt: response.timestamp ?? new Date().toISOString(),
  };
};

const fetchPositionHeadcountStats = async (
  params: NormalizedPositionHeadcountParams,
  signal?: AbortSignal,
): Promise<PositionHeadcountStats> => {
  const response = await graphqlEnterpriseAdapter.request<PositionHeadcountStatsGraphQLResponse>(
    POSITION_HEADCOUNT_STATS_QUERY_DOCUMENT,
    buildHeadcountVariables(params),
    { signal },
  );

  if (!response.success || !response.data) {
    throw createQueryError(response.error?.message ?? '获取编制统计失败', {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    });
  }

  const payload = response.data.positionHeadcountStats;

  return {
    organizationCode: payload.organizationCode,
    organizationName: payload.organizationName,
    totalCapacity: payload.totalCapacity,
    totalFilled: payload.totalFilled,
    totalAvailable: payload.totalAvailable,
    fillRate: payload.fillRate,
    byLevel: payload.byLevel.map(item => ({
      jobLevelCode: item.jobLevelCode,
      capacity: item.capacity,
      utilized: item.utilized,
      available: item.available,
    })),
    byType: payload.byType.map(item => ({
      positionType: item.positionType,
      capacity: item.capacity,
      filled: item.filled,
      available: item.available,
    })),
    byFamily: (payload.byFamily ?? []).map(item => ({
      jobFamilyCode: item.jobFamilyCode,
      jobFamilyName: item.jobFamilyName ?? undefined,
      capacity: item.capacity,
      utilized: item.utilized,
      available: item.available,
    })),
    fetchedAt: response.timestamp ?? new Date().toISOString(),
  };
};

export const POSITIONS_QUERY_ROOT_KEY = ['enterprise-positions'] as const;
export const POSITION_DETAIL_QUERY_ROOT_KEY = ['enterprise-position-detail'] as const;
export const POSITION_ASSIGNMENTS_QUERY_ROOT_KEY = ['enterprise-position-assignments'] as const;
export const VACANT_POSITIONS_QUERY_ROOT_KEY = ['enterprise-vacant-positions'] as const;
export const POSITION_HEADCOUNT_STATS_QUERY_ROOT_KEY = ['enterprise-position-headcount-stats'] as const;

export const positionsQueryKey = (params: NormalizedPositionQueryParams) =>
  [...POSITIONS_QUERY_ROOT_KEY, params] as const;

export const positionDetailQueryKey = (code: string, includeDeleted: boolean) =>
  [...POSITION_DETAIL_QUERY_ROOT_KEY, { code, includeDeleted }] as const;

export const positionAssignmentsQueryKey = (params: NormalizedPositionAssignmentsQueryParams) =>
  [...POSITION_ASSIGNMENTS_QUERY_ROOT_KEY, params] as const;

export const vacantPositionsQueryKey = (params: NormalizedVacantPositionsQueryParams) =>
  [...VACANT_POSITIONS_QUERY_ROOT_KEY, params] as const;

export const positionHeadcountStatsQueryKey = (params: NormalizedPositionHeadcountParams) =>
  [...POSITION_HEADCOUNT_STATS_QUERY_ROOT_KEY, params] as const;

type PositionsQueryKey = ReturnType<typeof positionsQueryKey>;
type PositionDetailQueryKey = ReturnType<typeof positionDetailQueryKey>;
type PositionAssignmentsQueryKey = ReturnType<typeof positionAssignmentsQueryKey>;
type VacantPositionsQueryKey = ReturnType<typeof vacantPositionsQueryKey>;
type PositionHeadcountStatsQueryKey = ReturnType<typeof positionHeadcountStatsQueryKey>;

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
  const [, params] = queryKey;
  return fetchPositionDetail(params.code, params.includeDeleted, signal);
};

const positionAssignmentsQueryFn = async ({
  queryKey,
  signal,
}: QueryFunctionContext<PositionAssignmentsQueryKey>): Promise<PositionAssignmentsQueryResult> => {
  const [, params] = queryKey;
  return fetchPositionAssignments(params, signal);
};

const vacantPositionsQueryFn = async ({
  queryKey,
  signal,
}: QueryFunctionContext<VacantPositionsQueryKey>): Promise<VacantPositionsQueryResult> => {
  const [, params] = queryKey;
  return fetchVacantPositions(params, signal);
};

const positionHeadcountStatsQueryFn = async ({
  queryKey,
  signal,
}: QueryFunctionContext<PositionHeadcountStatsQueryKey>): Promise<PositionHeadcountStats> => {
  const [, params] = queryKey;
  return fetchPositionHeadcountStats(params, signal);
};

export function useEnterprisePositions(
  params: PositionQueryParams = {},
): UseQueryResult<PositionsQueryResult> {
  const normalizedParams = normalizePositionParams(params);

  return useQuery({
    queryKey: positionsQueryKey(normalizedParams),
    queryFn: positionsQueryFn,
    staleTime: 60_000,
  });
}

export interface PositionDetailOptions {
  enabled?: boolean;
  includeDeleted?: boolean;
}

export function usePositionDetail(
  code: string | undefined,
  options?: PositionDetailOptions,
): UseQueryResult<PositionDetailResult> {
  const includeDeleted = options?.includeDeleted ?? false;
  const normalizedCode = code ?? 'placeholder';
  const enabled = Boolean(code) && (options?.enabled ?? true);
  const queryKey = positionDetailQueryKey(normalizedCode, includeDeleted);

  return useQuery({
    queryKey,
    queryFn: positionDetailQueryFn,
    enabled,
    staleTime: 60_000,
  });
}

/**
 * 预热职位详情（用于路由级 Loader 预热）
 * - 作为稳定导出，避免上层直接依赖内部实现细节
 */
export async function prefetchPositionDetail(
  client: QueryClient,
  code: string,
  includeDeleted = false,
): Promise<void> {
  const key = positionDetailQueryKey(code, includeDeleted);
  await client.prefetchQuery({
    queryKey: key,
    queryFn: positionDetailQueryFn,
    staleTime: 60_000,
  });
}

export function usePositionAssignments(
  positionCode: string | undefined,
  params: PositionAssignmentsQueryParams = {},
): UseQueryResult<PositionAssignmentsQueryResult> {
  const normalizedParams = useMemo(() => {
    if (!positionCode) {
      return null;
    }
    return normalizeAssignmentQueryParams(positionCode, params);
  }, [positionCode, params]);

  const effectiveParams = normalizedParams ?? DISABLED_ASSIGNMENT_QUERY_PARAMS;

  return useQuery<
    PositionAssignmentsQueryResult,
    Error,
    PositionAssignmentsQueryResult,
    PositionAssignmentsQueryKey
  >({
    queryKey: positionAssignmentsQueryKey(effectiveParams),
    queryFn: positionAssignmentsQueryFn,
    placeholderData: keepPreviousData,
    staleTime: 30_000,
    enabled: Boolean(normalizedParams),
  });
}

export function useVacantPositions(
  params: VacantPositionsQueryParams = {},
): UseQueryResult<VacantPositionsQueryResult> {
  const normalizedParams = normalizeVacantPositionsParams(params);

  return useQuery({
    queryKey: vacantPositionsQueryKey(normalizedParams),
    queryFn: vacantPositionsQueryFn,
    staleTime: 30_000,
  });
}

export function usePositionHeadcountStats(
  params: PositionHeadcountStatsParams,
): UseQueryResult<PositionHeadcountStats> {
  const normalizedParams = normalizeHeadcountParams(params);

  return useQuery({
    queryKey: positionHeadcountStatsQueryKey(normalizedParams),
    queryFn: positionHeadcountStatsQueryFn,
    enabled: Boolean(normalizedParams.organizationCode),
    staleTime: 60_000,
  });
}

export async function fetchPositionAssignmentAudit(
  positionCode: string,
  params: PositionAssignmentAuditParams = {},
  signal?: AbortSignal,
): Promise<PositionAssignmentAuditQueryResult> {
  const normalizedParams = normalizeAssignmentAuditParams(positionCode, params);
  return fetchPositionAssignmentAuditInternal(normalizedParams, signal);
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
  transformAssignmentAuditNode,
  transformTransferNode,
  transformVacantPositionNode,
  transformTimelineEntry,
  fetchPositionsWithParams,
  fetchPositionDetail,
  fetchVacantPositions,
  normalizeVacantPositionsParams,
  buildVacantPositionsVariables,
  normalizeHeadcountParams,
  buildHeadcountVariables,
  fetchPositionHeadcountStats,
  normalizeAssignmentQueryParams,
  buildAssignmentVariables,
  fetchPositionAssignments,
  normalizeAssignmentAuditParams,
  buildAssignmentAuditVariables,
  fetchPositionAssignmentAuditInternal,
};
