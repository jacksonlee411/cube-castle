export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = {
  [K in keyof T]: T[K];
};
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & {
  [SubKey in K]?: Maybe<T[SubKey]>;
};
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & {
  [SubKey in K]: Maybe<T[SubKey]>;
};
export type MakeEmpty<
  T extends { [key: string]: unknown },
  K extends keyof T,
> = { [_ in K]?: never };
export type Incremental<T> =
  | T
  | {
      [P in keyof T]?: P extends " $fragmentName" | "__typename" ? T[P] : never;
    };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string };
  String: { input: string; output: string };
  Boolean: { input: boolean; output: boolean };
  Int: { input: number; output: number };
  Float: { input: number; output: number };
  /** Date scalar for date-only values (YYYY-MM-DD format). */
  Date: { input: string; output: string };
  /** DateTime scalar for timestamp values (ISO 8601 format). */
  DateTime: { input: string; output: string };
  /** JSON scalar for arbitrary JSON data structures. */
  JSON: { input: Record<string, unknown>; output: Record<string, unknown> };
  /** Job family code scalar (group + '-' + 3-6 alphanumeric characters). */
  JobFamilyCode: { input: string; output: string };
  /** Job family group code scalar (4-6 uppercase letters). */
  JobFamilyGroupCode: { input: string; output: string };
  /** Job level code scalar (uppercase letter + 1-2 digits). */
  JobLevelCode: { input: string; output: string };
  /** Job role code scalar (family + '-' + 3-6 alphanumeric characters). */
  JobRoleCode: { input: string; output: string };
  /** Position code scalar (P + 7 digits). */
  PositionCode: { input: string; output: string };
  /** UUID scalar for universally unique identifiers. */
  UUID: { input: string; output: string };
};

/**
 * Comprehensive audit log entry with complete change tracking information.
 * Each audit record tracks changes to a specific organization temporal version (recordId).
 */
export type AuditLogDetail = {
  __typename: "AuditLogDetail";
  afterData?: Maybe<Scalars["String"]["output"]>;
  auditId: Scalars["String"]["output"];
  beforeData?: Maybe<Scalars["String"]["output"]>;
  changes: Array<FieldChange>;
  modifiedFields: Array<Scalars["String"]["output"]>;
  operation: Scalars["String"]["output"];
  operationReason?: Maybe<Scalars["String"]["output"]>;
  recordId: Scalars["String"]["output"];
  timestamp: Scalars["String"]["output"];
};

/** Cache inconsistency detection result. */
export type CacheInconsistency = {
  __typename: "CacheInconsistency";
  cachedValue: Scalars["String"]["output"];
  calculatedValue: Scalars["String"]["output"];
  code: Scalars["String"]["output"];
  fieldName: Scalars["String"]["output"];
  impactLevel: Scalars["String"]["output"];
};

/** Circular reference detection result. */
export type CircularReference = {
  __typename: "CircularReference";
  affectedCodes: Array<Scalars["String"]["output"]>;
  circularPath: Array<Scalars["String"]["output"]>;
  severity: Scalars["String"]["output"];
};

/** Consistency check modes with different performance characteristics. */
export enum ConsistencyCheckMode {
  DEEP = "DEEP",
  FAST = "FAST",
  TARGETED = "TARGETED",
}

/** Consistency check findings with detailed issues. */
export type ConsistencyFindings = {
  __typename: "ConsistencyFindings";
  cacheInconsistencies: Array<CacheInconsistency>;
  circularReferences: Array<CircularReference>;
  depthViolations: Array<DepthViolation>;
  levelInconsistencies: Array<LevelInconsistency>;
  orphanedNodes: Array<OrphanedNode>;
  pathMismatches: Array<PathMismatch>;
};

/** Simplified data changes with before/after comparison. */
export type DataChanges = {
  __typename: "DataChanges";
  afterData?: Maybe<Scalars["JSON"]["output"]>;
  beforeData?: Maybe<Scalars["JSON"]["output"]>;
  modifiedFields: Array<Scalars["String"]["output"]>;
};

/** Date range specification. */
export type DateRange = {
  __typename: "DateRange";
  earliest: Scalars["DateTime"]["output"];
  latest: Scalars["DateTime"]["output"];
};

/** Date range input for filtering. */
export type DateRangeInput = {
  from?: InputMaybe<Scalars["String"]["input"]>;
  to?: InputMaybe<Scalars["String"]["input"]>;
};

/** Hierarchy depth distribution analysis. */
export type DepthDistribution = {
  __typename: "DepthDistribution";
  count: Scalars["Int"]["output"];
  depth: Scalars["Int"]["output"];
};

/** Depth violation detection result. */
export type DepthViolation = {
  __typename: "DepthViolation";
  code: Scalars["String"]["output"];
  currentDepth: Scalars["Int"]["output"];
  maxAllowedDepth: Scalars["Int"]["output"];
  parentChain: Array<Scalars["String"]["output"]>;
};

/** Employment type for position assignments. */
export enum EmploymentType {
  FULL_TIME = "FULL_TIME",
  INTERN = "INTERN",
  PART_TIME = "PART_TIME",
}

export type FamilyHeadcount = {
  __typename: "FamilyHeadcount";
  available: Scalars["Float"]["output"];
  capacity: Scalars["Float"]["output"];
  jobFamilyCode: Scalars["JobFamilyCode"]["output"];
  jobFamilyName?: Maybe<Scalars["String"]["output"]>;
  utilized: Scalars["Float"]["output"];
};

/** Detailed field-level change information for audit tracking. */
export type FieldChange = {
  __typename: "FieldChange";
  dataType: Scalars["String"]["output"];
  field: Scalars["String"]["output"];
  newValue?: Maybe<Scalars["String"]["output"]>;
  oldValue?: Maybe<Scalars["String"]["output"]>;
};

export type HeadcountStats = {
  __typename: "HeadcountStats";
  byFamily: Array<FamilyHeadcount>;
  byLevel: Array<LevelHeadcount>;
  byType: Array<TypeHeadcount>;
  fillRate: Scalars["Float"]["output"];
  organizationCode: Scalars["String"]["output"];
  organizationName: Scalars["String"]["output"];
  totalAvailable: Scalars["Float"]["output"];
  totalCapacity: Scalars["Float"]["output"];
  totalFilled: Scalars["Float"]["output"];
};

/** Hierarchy consistency check report with detailed findings and repair suggestions. */
export type HierarchyConsistencyReport = {
  __typename: "HierarchyConsistencyReport";
  checkId: Scalars["String"]["output"];
  checkMode: ConsistencyCheckMode;
  consistencyReport: ConsistencyFindings;
  executedAt: Scalars["DateTime"]["output"];
  executionTimeMs: Scalars["Int"]["output"];
  healthScore: Scalars["Float"]["output"];
  issuesFound: Scalars["Int"]["output"];
  recommendedActions: Array<Scalars["String"]["output"]>;
  repairSuggestions: Array<RepairSuggestion>;
  tenantId: Scalars["String"]["output"];
  totalChecked: Scalars["Int"]["output"];
};

/** Hierarchy distribution statistics and integrity analysis. */
export type HierarchyStatistics = {
  __typename: "HierarchyStatistics";
  avgDepth: Scalars["Float"]["output"];
  depthDistribution: Array<DepthDistribution>;
  integrityIssues: Array<IntegrityIssue>;
  lastAnalyzed: Scalars["String"]["output"];
  leafOrganizations: Scalars["Int"]["output"];
  maxDepth: Scalars["Int"]["output"];
  rootOrganizations: Scalars["Int"]["output"];
  tenantId: Scalars["String"]["output"];
  totalOrganizations: Scalars["Int"]["output"];
};

/** Hierarchy integrity issue detection. */
export type IntegrityIssue = {
  __typename: "IntegrityIssue";
  affectedCodes: Array<Scalars["String"]["output"]>;
  count: Scalars["Int"]["output"];
  type: Scalars["String"]["output"];
};

/** Status for job catalog entries. */
export enum JobCatalogStatus {
  ACTIVE = "ACTIVE",
  INACTIVE = "INACTIVE",
}

export type JobFamily = {
  __typename: "JobFamily";
  code: Scalars["JobFamilyCode"]["output"];
  description?: Maybe<Scalars["String"]["output"]>;
  effectiveDate: Scalars["Date"]["output"];
  endDate?: Maybe<Scalars["Date"]["output"]>;
  groupCode: Scalars["JobFamilyGroupCode"]["output"];
  name: Scalars["String"]["output"];
  recordId: Scalars["UUID"]["output"];
  status: JobCatalogStatus;
};

export type JobFamilyGroup = {
  __typename: "JobFamilyGroup";
  code: Scalars["JobFamilyGroupCode"]["output"];
  description?: Maybe<Scalars["String"]["output"]>;
  effectiveDate: Scalars["Date"]["output"];
  endDate?: Maybe<Scalars["Date"]["output"]>;
  name: Scalars["String"]["output"];
  recordId: Scalars["UUID"]["output"];
  status: JobCatalogStatus;
};

export type JobLevel = {
  __typename: "JobLevel";
  code: Scalars["JobLevelCode"]["output"];
  description?: Maybe<Scalars["String"]["output"]>;
  effectiveDate: Scalars["Date"]["output"];
  endDate?: Maybe<Scalars["Date"]["output"]>;
  levelRank: Scalars["Int"]["output"];
  name: Scalars["String"]["output"];
  recordId: Scalars["UUID"]["output"];
  roleCode: Scalars["JobRoleCode"]["output"];
  status: JobCatalogStatus;
};

export type JobRole = {
  __typename: "JobRole";
  code: Scalars["JobRoleCode"]["output"];
  description?: Maybe<Scalars["String"]["output"]>;
  effectiveDate: Scalars["Date"]["output"];
  endDate?: Maybe<Scalars["Date"]["output"]>;
  familyCode: Scalars["JobFamilyCode"]["output"];
  name: Scalars["String"]["output"];
  recordId: Scalars["UUID"]["output"];
  status: JobCatalogStatus;
};

export type LevelHeadcount = {
  __typename: "LevelHeadcount";
  available: Scalars["Float"]["output"];
  capacity: Scalars["Float"]["output"];
  jobLevelCode: Scalars["JobLevelCode"]["output"];
  utilized: Scalars["Float"]["output"];
};

/** Level inconsistency detection result. */
export type LevelInconsistency = {
  __typename: "LevelInconsistency";
  actualLevel: Scalars["Int"]["output"];
  code: Scalars["String"]["output"];
  expectedLevel: Scalars["Int"]["output"];
  parentCode: Scalars["String"]["output"];
  reason: Scalars["String"]["output"];
};

/** Statistics by hierarchy level. */
export type LevelStatistic = {
  __typename: "LevelStatistic";
  count: Scalars["Int"]["output"];
  level: Scalars["Int"]["output"];
};

/** Standard operator information with ID and display name. */
export type OperatedBy = {
  __typename: "OperatedBy";
  id: Scalars["String"]["output"];
  name: Scalars["String"]["output"];
};

/** Operation types for audit and temporal tracking. */
export enum OperationType {
  CREATE = "CREATE",
  DEACTIVATE = "DEACTIVATE",
  DELETE = "DELETE",
  REACTIVATE = "REACTIVATE",
  SUSPEND = "SUSPEND",
  UPDATE = "UPDATE",
}

/** Operations summary statistics. */
export type OperationsSummary = {
  __typename: "OperationsSummary";
  create: Scalars["Int"]["output"];
  delete: Scalars["Int"]["output"];
  reactivate: Scalars["Int"]["output"];
  suspend: Scalars["Int"]["output"];
  update: Scalars["Int"]["output"];
};

/**
 * Organization unit entity with complete temporal and audit information.
 * Represents the current state based on asOfDate parameter or latest effective record.
 */
export type Organization = {
  __typename: "Organization";
  changeReason?: Maybe<Scalars["String"]["output"]>;
  childrenCount: Scalars["Int"]["output"];
  code: Scalars["String"]["output"];
  codePath: Scalars["String"]["output"];
  createdAt: Scalars["String"]["output"];
  deletedAt?: Maybe<Scalars["String"]["output"]>;
  deletedBy?: Maybe<Scalars["String"]["output"]>;
  deletionReason?: Maybe<Scalars["String"]["output"]>;
  description?: Maybe<Scalars["String"]["output"]>;
  effectiveDate: Scalars["String"]["output"];
  endDate?: Maybe<Scalars["String"]["output"]>;
  hierarchyDepth: Scalars["Int"]["output"];
  isCurrent: Scalars["Boolean"]["output"];
  isFuture: Scalars["Boolean"]["output"];
  isTemporal: Scalars["Boolean"]["output"];
  level: Scalars["Int"]["output"];
  name: Scalars["String"]["output"];
  namePath: Scalars["String"]["output"];
  parentCode: Scalars["String"]["output"];
  /** @deprecated 使用 codePath/namePath 作为唯一事实来源，path 将在后续版本移除 */
  path?: Maybe<Scalars["String"]["output"]>;
  profile?: Maybe<Scalars["String"]["output"]>;
  recordId: Scalars["String"]["output"];
  sortOrder?: Maybe<Scalars["Int"]["output"]>;
  status: Status;
  suspendedAt?: Maybe<Scalars["String"]["output"]>;
  suspendedBy?: Maybe<Scalars["String"]["output"]>;
  suspensionReason?: Maybe<Scalars["String"]["output"]>;
  tenantId: Scalars["String"]["output"];
  unitType: UnitType;
  updatedAt: Scalars["String"]["output"];
};

/** Connection type for paginated organization results with metadata. */
export type OrganizationConnection = {
  __typename: "OrganizationConnection";
  data: Array<Organization>;
  pagination: PaginationInfo;
  temporal: TemporalInfo;
};

/** Comprehensive filter for organization queries with temporal support. */
export type OrganizationFilter = {
  asOfDate?: InputMaybe<Scalars["String"]["input"]>;
  codes?: InputMaybe<Array<Scalars["String"]["input"]>>;
  excludeCodes?: InputMaybe<Array<Scalars["String"]["input"]>>;
  excludeDescendantsOf?: InputMaybe<Scalars["String"]["input"]>;
  hasChildren?: InputMaybe<Scalars["Boolean"]["input"]>;
  hasProfile?: InputMaybe<Scalars["Boolean"]["input"]>;
  includeDisabledAncestors?: InputMaybe<Scalars["Boolean"]["input"]>;
  includeFuture?: InputMaybe<Scalars["Boolean"]["input"]>;
  leavesOnly?: InputMaybe<Scalars["Boolean"]["input"]>;
  level?: InputMaybe<Scalars["Int"]["input"]>;
  maxLevel?: InputMaybe<Scalars["Int"]["input"]>;
  minLevel?: InputMaybe<Scalars["Int"]["input"]>;
  onlyFuture?: InputMaybe<Scalars["Boolean"]["input"]>;
  operatedBy?: InputMaybe<Scalars["String"]["input"]>;
  operationDateRange?: InputMaybe<DateRangeInput>;
  operationType?: InputMaybe<OperationType>;
  parentCode?: InputMaybe<Scalars["String"]["input"]>;
  profileContains?: InputMaybe<Scalars["JSON"]["input"]>;
  rootsOnly?: InputMaybe<Scalars["Boolean"]["input"]>;
  searchFields?: InputMaybe<Array<SearchField>>;
  searchText?: InputMaybe<Scalars["String"]["input"]>;
  status?: InputMaybe<Status>;
  unitType?: InputMaybe<UnitType>;
};

/** Hierarchy-specific organization information with relationship context. */
export type OrganizationHierarchy = {
  __typename: "OrganizationHierarchy";
  children: Array<OrganizationHierarchy>;
  childrenCount: Scalars["Int"]["output"];
  code: Scalars["String"]["output"];
  codePath: Scalars["String"]["output"];
  hierarchyDepth: Scalars["Int"]["output"];
  isLeaf: Scalars["Boolean"]["output"];
  isRoot: Scalars["Boolean"]["output"];
  level: Scalars["Int"]["output"];
  name: Scalars["String"]["output"];
  namePath: Scalars["String"]["output"];
  parentChain: Array<Scalars["String"]["output"]>;
};

/** Comprehensive organization statistics with temporal breakdown. */
export type OrganizationStats = {
  __typename: "OrganizationStats";
  activeCount: Scalars["Int"]["output"];
  byLevel: Array<LevelStatistic>;
  byStatus: Array<StatusStatistic>;
  byType: Array<TypeStatistic>;
  deletedCount: Scalars["Int"]["output"];
  inactiveCount: Scalars["Int"]["output"];
  plannedCount: Scalars["Int"]["output"];
  temporalStats: TemporalStatistics;
  totalCount: Scalars["Int"]["output"];
};

/** Orphaned node detection result. */
export type OrphanedNode = {
  __typename: "OrphanedNode";
  code: Scalars["String"]["output"];
  name: Scalars["String"]["output"];
  parentCode: Scalars["String"]["output"];
  reason: Scalars["String"]["output"];
};

/** Pagination information for connection types. */
export type PaginationInfo = {
  __typename: "PaginationInfo";
  hasNext: Scalars["Boolean"]["output"];
  hasPrevious: Scalars["Boolean"]["output"];
  page: Scalars["Int"]["output"];
  pageSize: Scalars["Int"]["output"];
  total: Scalars["Int"]["output"];
};

/** Pagination configuration for result sets. */
export type PaginationInput = {
  page?: InputMaybe<Scalars["Int"]["input"]>;
  pageSize?: InputMaybe<Scalars["Int"]["input"]>;
  sortBy?: InputMaybe<Scalars["String"]["input"]>;
  sortOrder?: InputMaybe<Scalars["String"]["input"]>;
};

/** Path mismatch detection result. */
export type PathMismatch = {
  __typename: "PathMismatch";
  actualCodePath: Scalars["String"]["output"];
  actualNamePath: Scalars["String"]["output"];
  code: Scalars["String"]["output"];
  expectedCodePath: Scalars["String"]["output"];
  expectedNamePath: Scalars["String"]["output"];
  severity: Scalars["String"]["output"];
};

/** Position resource exposed via GraphQL. */
export type Position = {
  __typename: "Position";
  assignmentHistory: Array<PositionAssignment>;
  availableHeadcount: Scalars["Float"]["output"];
  code: Scalars["PositionCode"]["output"];
  createdAt: Scalars["DateTime"]["output"];
  currentAssignment?: Maybe<PositionAssignment>;
  effectiveDate: Scalars["Date"]["output"];
  employmentType: EmploymentType;
  endDate?: Maybe<Scalars["Date"]["output"]>;
  gradeLevel?: Maybe<Scalars["String"]["output"]>;
  headcountCapacity: Scalars["Float"]["output"];
  headcountInUse: Scalars["Float"]["output"];
  isCurrent: Scalars["Boolean"]["output"];
  isFuture: Scalars["Boolean"]["output"];
  jobFamilyCode: Scalars["JobFamilyCode"]["output"];
  jobFamilyGroupCode: Scalars["JobFamilyGroupCode"]["output"];
  jobLevelCode: Scalars["JobLevelCode"]["output"];
  jobProfileCode?: Maybe<Scalars["String"]["output"]>;
  jobProfileName?: Maybe<Scalars["String"]["output"]>;
  jobRoleCode: Scalars["JobRoleCode"]["output"];
  organizationCode: Scalars["String"]["output"];
  organizationName?: Maybe<Scalars["String"]["output"]>;
  positionType: PositionType;
  recordId: Scalars["UUID"]["output"];
  reportsToPositionCode?: Maybe<Scalars["PositionCode"]["output"]>;
  status: PositionStatus;
  tenantId: Scalars["UUID"]["output"];
  title: Scalars["String"]["output"];
  updatedAt: Scalars["DateTime"]["output"];
};

export type PositionAssignment = {
  __typename: "PositionAssignment";
  assignmentId: Scalars["UUID"]["output"];
  assignmentStatus: PositionAssignmentStatus;
  assignmentType: PositionAssignmentType;
  createdAt: Scalars["DateTime"]["output"];
  employeeId: Scalars["UUID"]["output"];
  employeeName: Scalars["String"]["output"];
  employeeNumber?: Maybe<Scalars["String"]["output"]>;
  endDate?: Maybe<Scalars["Date"]["output"]>;
  fte: Scalars["Float"]["output"];
  isCurrent: Scalars["Boolean"]["output"];
  notes?: Maybe<Scalars["String"]["output"]>;
  positionCode: Scalars["PositionCode"]["output"];
  positionRecordId: Scalars["UUID"]["output"];
  startDate: Scalars["Date"]["output"];
  updatedAt: Scalars["DateTime"]["output"];
};

export type PositionAssignmentConnection = {
  __typename: "PositionAssignmentConnection";
  data: Array<PositionAssignment>;
  edges: Array<PositionAssignmentEdge>;
  pagination: PaginationInfo;
  totalCount: Scalars["Int"]["output"];
};

export type PositionAssignmentEdge = {
  __typename: "PositionAssignmentEdge";
  cursor: Scalars["String"]["output"];
  node: PositionAssignment;
};

/** Filter options for position assignment queries. */
export type PositionAssignmentFilterInput = {
  asOfDate?: InputMaybe<Scalars["Date"]["input"]>;
  assignmentStatus?: InputMaybe<PositionAssignmentStatus>;
  assignmentType?: InputMaybe<PositionAssignmentType>;
  employeeId?: InputMaybe<Scalars["UUID"]["input"]>;
  includeHistorical?: InputMaybe<Scalars["Boolean"]["input"]>;
};

/** Supported assignment sorting fields. */
export enum PositionAssignmentSortField {
  CREATED_AT = "CREATED_AT",
  END_DATE = "END_DATE",
  START_DATE = "START_DATE",
}

/** Sorting input for assignment queries. */
export type PositionAssignmentSortInput = {
  direction?: InputMaybe<SortOrder>;
  field: PositionAssignmentSortField;
};

/** Lifecycle status for a position assignment. */
export enum PositionAssignmentStatus {
  ACTIVE = "ACTIVE",
  ENDED = "ENDED",
  PENDING = "PENDING",
}

/** Assignment type within a position. */
export enum PositionAssignmentType {
  ACTING = "ACTING",
  PRIMARY = "PRIMARY",
  SECONDARY = "SECONDARY",
}

export type PositionConnection = {
  __typename: "PositionConnection";
  data: Array<Position>;
  edges: Array<PositionEdge>;
  pagination: PaginationInfo;
  totalCount: Scalars["Int"]["output"];
};

export type PositionEdge = {
  __typename: "PositionEdge";
  cursor: Scalars["String"]["output"];
  node: Position;
};

/** Filter options for position queries. */
export type PositionFilterInput = {
  effectiveRange?: InputMaybe<DateRangeInput>;
  employmentTypes?: InputMaybe<Array<EmploymentType>>;
  jobFamilyCodes?: InputMaybe<Array<Scalars["JobFamilyCode"]["input"]>>;
  jobFamilyGroupCodes?: InputMaybe<
    Array<Scalars["JobFamilyGroupCode"]["input"]>
  >;
  jobLevelCodes?: InputMaybe<Array<Scalars["JobLevelCode"]["input"]>>;
  jobRoleCodes?: InputMaybe<Array<Scalars["JobRoleCode"]["input"]>>;
  organizationCode?: InputMaybe<Scalars["String"]["input"]>;
  positionCodes?: InputMaybe<Array<Scalars["PositionCode"]["input"]>>;
  positionTypes?: InputMaybe<Array<PositionType>>;
  status?: InputMaybe<PositionStatus>;
};

/** Supported position sorting fields. */
export enum PositionSortField {
  CODE = "CODE",
  EFFECTIVE_DATE = "EFFECTIVE_DATE",
  STATUS = "STATUS",
  TITLE = "TITLE",
}

/** Sorting input for position queries. */
export type PositionSortInput = {
  direction?: InputMaybe<SortOrder>;
  field: PositionSortField;
};

/** Lifecycle status for positions. */
export enum PositionStatus {
  ACTIVE = "ACTIVE",
  DELETED = "DELETED",
  FILLED = "FILLED",
  INACTIVE = "INACTIVE",
  PLANNED = "PLANNED",
  VACANT = "VACANT",
}

/** Entry describing a specific temporal version of a position. */
export type PositionTimelineEntry = {
  __typename: "PositionTimelineEntry";
  changeReason?: Maybe<Scalars["String"]["output"]>;
  effectiveDate: Scalars["Date"]["output"];
  endDate?: Maybe<Scalars["Date"]["output"]>;
  isCurrent: Scalars["Boolean"]["output"];
  recordId: Scalars["UUID"]["output"];
  status: PositionStatus;
  title: Scalars["String"]["output"];
};

export type PositionTransfer = {
  __typename: "PositionTransfer";
  createdAt: Scalars["DateTime"]["output"];
  effectiveDate: Scalars["Date"]["output"];
  fromOrganizationCode: Scalars["String"]["output"];
  initiatedBy: OperatedBy;
  operationReason?: Maybe<Scalars["String"]["output"]>;
  positionCode: Scalars["PositionCode"]["output"];
  toOrganizationCode: Scalars["String"]["output"];
  transferId: Scalars["UUID"]["output"];
};

export type PositionTransferConnection = {
  __typename: "PositionTransferConnection";
  data: Array<PositionTransfer>;
  edges: Array<PositionTransferEdge>;
  pagination: PaginationInfo;
  totalCount: Scalars["Int"]["output"];
};

export type PositionTransferEdge = {
  __typename: "PositionTransferEdge";
  cursor: Scalars["String"]["output"];
  node: PositionTransfer;
};

/** Contract type for positions. */
export enum PositionType {
  CONTRACTOR = "CONTRACTOR",
  REGULAR = "REGULAR",
  TEMPORARY = "TEMPORARY",
}

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type Query = {
  __typename: "Query";
  /**
   * Get complete audit history for a specific organization record (temporal version).
   * This query returns audit records for a specific recordId, allowing precise tracking
   * of individual temporal version lifecycle changes.
   *
   * Permissions Required: org:read:audit
   * Performance: Record-specific audit queries with optimized indexing < 50ms
   */
  auditHistory: Array<AuditLogDetail>;
  /**
   * Get detailed audit record with before/after data snapshots and field-level changes.
   *
   * Permissions Required: org:read:audit
   * Performance: Single audit record retrieval < 20ms
   */
  auditLog?: Maybe<AuditLogDetail>;
  /**
   * Get hierarchy distribution statistics and integrity analysis.
   *
   * Permissions Required: org:read:hierarchy
   * Performance: Aggregation with hierarchy statistics < 150ms
   */
  hierarchyStatistics: HierarchyStatistics;
  /**
   * Get job families under a specified group.
   *
   * Permissions Required: job-catalog:read
   */
  jobFamilies: Array<JobFamily>;
  /**
   * Get job family groups with optional historical inclusion.
   *
   * Permissions Required: job-catalog:read
   */
  jobFamilyGroups: Array<JobFamilyGroup>;
  /**
   * Get job levels under a specified role.
   *
   * Permissions Required: job-catalog:read
   */
  jobLevels: Array<JobLevel>;
  /**
   * Get job roles under a specified family.
   *
   * Permissions Required: job-catalog:read
   */
  jobRoles: Array<JobRole>;
  /**
   * Get single organization unit by business code with temporal support.
   *
   * Permissions Required: org:read
   * Performance: Index-optimized single record retrieval < 10ms
   */
  organization?: Maybe<Organization>;
  /**
   * Get complete hierarchy information for an organization including paths and relationships.
   *
   * Permissions Required: org:read:hierarchy
   * Performance: Path-optimized queries with hierarchy cache < 30ms
   */
  organizationHierarchy?: Maybe<OrganizationHierarchy>;
  /**
   * Get comprehensive organization statistics with temporal breakdown.
   *
   * Permissions Required: org:read:stats
   * Performance: Cached aggregation queries < 100ms
   */
  organizationStats: OrganizationStats;
  /**
   * Get organization subtree with configurable depth limits and relationship details.
   * Use this for multi-level hierarchy display (depth >= 2).
   * For direct children only, use organizations(filter: {parentCode: "code"}) instead.
   *
   * Permissions Required: org:read:hierarchy
   * Performance: Recursive CTE queries, optimized for display < 200ms
   */
  organizationSubtree: Array<OrganizationHierarchy>;
  /**
   * Return all temporal versions for an organization code (ascending by effectiveDate).
   * Requires scope: org:read:history
   */
  organizationVersions: Array<Organization>;
  /**
   * Get paginated list of organization units with advanced filtering and temporal support.
   *
   * Permissions Required: org:read
   * Performance: Optimized with specialized indexes, typical response < 50ms
   */
  organizations: OrganizationConnection;
  /**
   * Get single position by code with optional temporal perspective.
   *
   * Permissions Required: position:read
   */
  position?: Maybe<Position>;
  /**
   * Get paginated assignment records for a position.
   *
   * Permissions Required: position:read
   */
  positionAssignments: PositionAssignmentConnection;
  /**
   * Get headcount statistics for positions under an organization.
   *
   * Permissions Required: position:read:stats
   */
  positionHeadcountStats: HeadcountStats;
  /**
   * Get temporal timeline for a specific position.
   *
   * Permissions Required: position:read:history
   */
  positionTimeline: Array<PositionTimelineEntry>;
  /**
   * Get transfer history for positions.
   *
   * Permissions Required: position:read:history
   */
  positionTransfers: PositionTransferConnection;
  /**
   * List temporal versions for a position.
   *
   * Permissions Required: position:read:history
   */
  positionVersions: Array<Position>;
  /**
   * Get paginated position list with temporal awareness.
   *
   * Permissions Required: position:read
   */
  positions: PositionConnection;
  /**
   * Get list of vacant positions with optional filters.
   *
   * Permissions Required: position:read
   */
  vacantPositions: VacantPositionConnection;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryAuditHistoryArgs = {
  endDate?: InputMaybe<Scalars["String"]["input"]>;
  limit?: InputMaybe<Scalars["Int"]["input"]>;
  operation?: InputMaybe<OperationType>;
  recordId: Scalars["String"]["input"];
  startDate?: InputMaybe<Scalars["String"]["input"]>;
  userId?: InputMaybe<Scalars["String"]["input"]>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryAuditLogArgs = {
  auditId: Scalars["String"]["input"];
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryHierarchyStatisticsArgs = {
  includeIntegrityCheck?: InputMaybe<Scalars["Boolean"]["input"]>;
  tenantId: Scalars["String"]["input"];
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryJobFamiliesArgs = {
  asOfDate?: InputMaybe<Scalars["Date"]["input"]>;
  groupCode: Scalars["JobFamilyGroupCode"]["input"];
  includeInactive?: InputMaybe<Scalars["Boolean"]["input"]>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryJobFamilyGroupsArgs = {
  asOfDate?: InputMaybe<Scalars["Date"]["input"]>;
  includeInactive?: InputMaybe<Scalars["Boolean"]["input"]>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryJobLevelsArgs = {
  asOfDate?: InputMaybe<Scalars["Date"]["input"]>;
  includeInactive?: InputMaybe<Scalars["Boolean"]["input"]>;
  roleCode: Scalars["JobRoleCode"]["input"];
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryJobRolesArgs = {
  asOfDate?: InputMaybe<Scalars["Date"]["input"]>;
  familyCode: Scalars["JobFamilyCode"]["input"];
  includeInactive?: InputMaybe<Scalars["Boolean"]["input"]>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryOrganizationArgs = {
  asOfDate?: InputMaybe<Scalars["String"]["input"]>;
  code: Scalars["String"]["input"];
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryOrganizationHierarchyArgs = {
  code: Scalars["String"]["input"];
  tenantId: Scalars["String"]["input"];
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryOrganizationStatsArgs = {
  asOfDate?: InputMaybe<Scalars["String"]["input"]>;
  includeHistorical?: InputMaybe<Scalars["Boolean"]["input"]>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryOrganizationSubtreeArgs = {
  code: Scalars["String"]["input"];
  includeInactive?: InputMaybe<Scalars["Boolean"]["input"]>;
  maxDepth?: InputMaybe<Scalars["Int"]["input"]>;
  tenantId: Scalars["String"]["input"];
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryOrganizationVersionsArgs = {
  code: Scalars["String"]["input"];
  includeDeleted?: InputMaybe<Scalars["Boolean"]["input"]>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryOrganizationsArgs = {
  filter?: InputMaybe<OrganizationFilter>;
  pagination?: InputMaybe<PaginationInput>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryPositionArgs = {
  asOfDate?: InputMaybe<Scalars["Date"]["input"]>;
  code: Scalars["PositionCode"]["input"];
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryPositionAssignmentsArgs = {
  filter?: InputMaybe<PositionAssignmentFilterInput>;
  pagination?: InputMaybe<PaginationInput>;
  positionCode: Scalars["PositionCode"]["input"];
  sorting?: InputMaybe<Array<PositionAssignmentSortInput>>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryPositionHeadcountStatsArgs = {
  includeSubordinates?: InputMaybe<Scalars["Boolean"]["input"]>;
  organizationCode: Scalars["String"]["input"];
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryPositionTimelineArgs = {
  code: Scalars["PositionCode"]["input"];
  endDate?: InputMaybe<Scalars["Date"]["input"]>;
  startDate?: InputMaybe<Scalars["Date"]["input"]>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryPositionTransfersArgs = {
  organizationCode?: InputMaybe<Scalars["String"]["input"]>;
  pagination?: InputMaybe<PaginationInput>;
  positionCode?: InputMaybe<Scalars["PositionCode"]["input"]>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryPositionVersionsArgs = {
  code: Scalars["PositionCode"]["input"];
  includeDeleted?: InputMaybe<Scalars["Boolean"]["input"]>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryPositionsArgs = {
  filter?: InputMaybe<PositionFilterInput>;
  pagination?: InputMaybe<PaginationInput>;
  sorting?: InputMaybe<Array<PositionSortInput>>;
};

/**
 * Root Query type providing all organization management query operations.
 * All queries require appropriate OAuth 2.0 permissions and support multi-tenant isolation.
 */
export type QueryVacantPositionsArgs = {
  filter?: InputMaybe<VacantPositionFilterInput>;
  pagination?: InputMaybe<PaginationInput>;
  sorting?: InputMaybe<Array<VacantPositionSortInput>>;
};

/** Repair suggestion with automation capability. */
export type RepairSuggestion = {
  __typename: "RepairSuggestion";
  affectedCodes: Array<Scalars["String"]["output"]>;
  automatable: Scalars["Boolean"]["output"];
  issueType: Scalars["String"]["output"];
  riskLevel: Scalars["String"]["output"];
  suggestedAction: Scalars["String"]["output"];
};

/** Search fields for text-based filtering. */
export enum SearchField {
  CODE_PATH = "CODE_PATH",
  DESCRIPTION = "DESCRIPTION",
  NAME = "NAME",
  NAME_PATH = "NAME_PATH",
}

/** Sorting field options. */
export enum SortField {
  CODE = "CODE",
  CREATED_AT = "CREATED_AT",
  EFFECTIVE_DATE = "EFFECTIVE_DATE",
  LEVEL = "LEVEL",
  NAME = "NAME",
  SORT_ORDER = "SORT_ORDER",
  UPDATED_AT = "UPDATED_AT",
}

/** Sorting order options. */
export enum SortOrder {
  ASC = "ASC",
  DESC = "DESC",
}

/** Organization business status (ADR-008: 一维业务状态模型). */
export enum Status {
  ACTIVE = "ACTIVE",
  DELETED = "DELETED",
  INACTIVE = "INACTIVE",
  PLANNED = "PLANNED",
}

/** Statistics by organization status. */
export type StatusStatistic = {
  __typename: "StatusStatistic";
  count: Scalars["Int"]["output"];
  status: Status;
};

/** Temporal context information for queries. */
export type TemporalInfo = {
  __typename: "TemporalInfo";
  asOfDate: Scalars["String"]["output"];
  currentCount: Scalars["Int"]["output"];
  futureCount: Scalars["Int"]["output"];
  historicalCount: Scalars["Int"]["output"];
};

/** Temporal statistics breakdown. */
export type TemporalStatistics = {
  __typename: "TemporalStatistics";
  averageVersionsPerOrg: Scalars["Float"]["output"];
  newestEffectiveDate: Scalars["String"]["output"];
  oldestEffectiveDate: Scalars["String"]["output"];
  totalVersions: Scalars["Int"]["output"];
};

export type TypeHeadcount = {
  __typename: "TypeHeadcount";
  available: Scalars["Float"]["output"];
  capacity: Scalars["Float"]["output"];
  filled: Scalars["Float"]["output"];
  positionType: PositionType;
};

/** Statistics by organization unit type. */
export type TypeStatistic = {
  __typename: "TypeStatistic";
  count: Scalars["Int"]["output"];
  unitType: UnitType;
};

/** Organization unit types with specific business semantics. */
export enum UnitType {
  COMPANY = "COMPANY",
  DEPARTMENT = "DEPARTMENT",
  ORGANIZATION_UNIT = "ORGANIZATION_UNIT",
  PROJECT_TEAM = "PROJECT_TEAM",
}

/** User information for audit trails. */
export type UserInfo = {
  __typename: "UserInfo";
  role?: Maybe<Scalars["String"]["output"]>;
  userId: Scalars["String"]["output"];
  userName: Scalars["String"]["output"];
};

export type VacantPosition = {
  __typename: "VacantPosition";
  headcountAvailable: Scalars["Float"]["output"];
  headcountCapacity: Scalars["Float"]["output"];
  jobFamilyCode: Scalars["JobFamilyCode"]["output"];
  jobLevelCode: Scalars["JobLevelCode"]["output"];
  jobRoleCode: Scalars["JobRoleCode"]["output"];
  organizationCode: Scalars["String"]["output"];
  organizationName?: Maybe<Scalars["String"]["output"]>;
  positionCode: Scalars["PositionCode"]["output"];
  totalAssignments: Scalars["Int"]["output"];
  vacantSince: Scalars["Date"]["output"];
};

export type VacantPositionConnection = {
  __typename: "VacantPositionConnection";
  data: Array<VacantPosition>;
  edges: Array<VacantPositionEdge>;
  pagination: PaginationInfo;
  totalCount: Scalars["Int"]["output"];
};

export type VacantPositionEdge = {
  __typename: "VacantPositionEdge";
  cursor: Scalars["String"]["output"];
  node: VacantPosition;
};

/** Filter options for vacant position queries. */
export type VacantPositionFilterInput = {
  asOfDate?: InputMaybe<Scalars["Date"]["input"]>;
  jobFamilyCodes?: InputMaybe<Array<Scalars["JobFamilyCode"]["input"]>>;
  jobLevelCodes?: InputMaybe<Array<Scalars["JobLevelCode"]["input"]>>;
  jobRoleCodes?: InputMaybe<Array<Scalars["JobRoleCode"]["input"]>>;
  minimumVacantDays?: InputMaybe<Scalars["Int"]["input"]>;
  organizationCodes?: InputMaybe<Array<Scalars["String"]["input"]>>;
  positionTypes?: InputMaybe<Array<PositionType>>;
};

/** Supported vacant position sorting fields. */
export enum VacantPositionSortField {
  HEADCOUNT_AVAILABLE = "HEADCOUNT_AVAILABLE",
  HEADCOUNT_CAPACITY = "HEADCOUNT_CAPACITY",
  VACANT_SINCE = "VACANT_SINCE",
}

/** Sorting input for vacant position queries. */
export type VacantPositionSortInput = {
  direction?: InputMaybe<SortOrder>;
  field: VacantPositionSortField;
};

export type GetValidParentOrganizationsQueryVariables = Exact<{
  asOfDate: Scalars["String"]["input"];
  currentCode: Scalars["String"]["input"];
  pageSize?: InputMaybe<Scalars["Int"]["input"]>;
}>;

export type GetValidParentOrganizationsQuery = {
  __typename: "Query";
  organizations: {
    __typename: "OrganizationConnection";
    data: Array<{
      __typename: "Organization";
      code: string;
      name: string;
      unitType: UnitType;
      parentCode: string;
      level: number;
      effectiveDate: string;
      endDate?: string | null;
      isFuture: boolean;
      childrenCount: number;
    }>;
    pagination: {
      __typename: "PaginationInfo";
      total: number;
      page: number;
      pageSize: number;
    };
  };
};

export type EnterpriseOrganizationsQueryVariables = Exact<{
  filter?: InputMaybe<OrganizationFilter>;
  pagination?: InputMaybe<PaginationInput>;
  statsAsOfDate?: InputMaybe<Scalars["String"]["input"]>;
  statsIncludeHistorical: Scalars["Boolean"]["input"];
}>;

export type EnterpriseOrganizationsQuery = {
  __typename: "Query";
  organizations: {
    __typename: "OrganizationConnection";
    data: Array<{
      __typename: "Organization";
      code: string;
      parentCode: string;
      tenantId: string;
      name: string;
      unitType: UnitType;
      status: Status;
      level: number;
      codePath: string;
      namePath: string;
      path?: string | null;
      sortOrder?: number | null;
      description?: string | null;
      profile?: string | null;
      effectiveDate: string;
      endDate?: string | null;
      createdAt: string;
      updatedAt: string;
      recordId: string;
      isFuture: boolean;
      hierarchyDepth: number;
      childrenCount: number;
      changeReason?: string | null;
      deletedAt?: string | null;
      deletedBy?: string | null;
      deletionReason?: string | null;
      suspendedAt?: string | null;
      suspendedBy?: string | null;
      suspensionReason?: string | null;
    }>;
    pagination: {
      __typename: "PaginationInfo";
      total: number;
      page: number;
      pageSize: number;
      hasNext: boolean;
      hasPrevious: boolean;
    };
    temporal: {
      __typename: "TemporalInfo";
      asOfDate: string;
      currentCount: number;
      futureCount: number;
      historicalCount: number;
    };
  };
  organizationStats: {
    __typename: "OrganizationStats";
    totalCount: number;
    activeCount: number;
    inactiveCount: number;
    plannedCount: number;
    deletedCount: number;
    byType: Array<{
      __typename: "TypeStatistic";
      unitType: UnitType;
      count: number;
    }>;
    byStatus: Array<{
      __typename: "StatusStatistic";
      status: Status;
      count: number;
    }>;
    byLevel: Array<{
      __typename: "LevelStatistic";
      level: number;
      count: number;
    }>;
    temporalStats: {
      __typename: "TemporalStatistics";
      totalVersions: number;
      averageVersionsPerOrg: number;
      oldestEffectiveDate: string;
      newestEffectiveDate: string;
    };
  };
};

export type OrganizationByCodeQueryVariables = Exact<{
  code: Scalars["String"]["input"];
  asOfDate?: InputMaybe<Scalars["String"]["input"]>;
}>;

export type OrganizationByCodeQuery = {
  __typename: "Query";
  organization?: {
    __typename: "Organization";
    code: string;
    parentCode: string;
    tenantId: string;
    name: string;
    unitType: UnitType;
    status: Status;
    level: number;
    codePath: string;
    namePath: string;
    path?: string | null;
    sortOrder?: number | null;
    description?: string | null;
    profile?: string | null;
    effectiveDate: string;
    endDate?: string | null;
    createdAt: string;
    updatedAt: string;
    recordId: string;
    changeReason?: string | null;
    deletedAt?: string | null;
    deletedBy?: string | null;
    deletionReason?: string | null;
    suspendedAt?: string | null;
    suspendedBy?: string | null;
    suspensionReason?: string | null;
    isFuture: boolean;
    hierarchyDepth: number;
    childrenCount: number;
  } | null;
};

export type EnterprisePositionsQueryVariables = Exact<{
  filter?: InputMaybe<PositionFilterInput>;
  pagination?: InputMaybe<PaginationInput>;
}>;

export type EnterprisePositionsQuery = {
  __typename: "Query";
  positions: {
    __typename: "PositionConnection";
    totalCount: number;
    data: Array<{
      __typename: "Position";
      code: string;
      title: string;
      jobFamilyGroupCode: string;
      jobFamilyCode: string;
      jobRoleCode: string;
      jobLevelCode: string;
      organizationCode: string;
      organizationName?: string | null;
      positionType: PositionType;
      employmentType: EmploymentType;
      headcountCapacity: number;
      headcountInUse: number;
      availableHeadcount: number;
      gradeLevel?: string | null;
      reportsToPositionCode?: string | null;
      status: PositionStatus;
      effectiveDate: string;
      endDate?: string | null;
      isCurrent: boolean;
      isFuture: boolean;
      createdAt: string;
      updatedAt: string;
    }>;
    pagination: {
      __typename: "PaginationInfo";
      total: number;
      page: number;
      pageSize: number;
      hasNext: boolean;
      hasPrevious: boolean;
    };
  };
};

export type PositionDetailQueryVariables = Exact<{
  code: Scalars["PositionCode"]["input"];
}>;

export type PositionDetailQuery = {
  __typename: "Query";
  position?: {
    __typename: "Position";
    code: string;
    recordId: string;
    title: string;
    jobFamilyGroupCode: string;
    jobFamilyCode: string;
    jobRoleCode: string;
    jobLevelCode: string;
    organizationCode: string;
    organizationName?: string | null;
    positionType: PositionType;
    employmentType: EmploymentType;
    headcountCapacity: number;
    headcountInUse: number;
    availableHeadcount: number;
    gradeLevel?: string | null;
    reportsToPositionCode?: string | null;
    status: PositionStatus;
    effectiveDate: string;
    endDate?: string | null;
    isCurrent: boolean;
    isFuture: boolean;
    createdAt: string;
    updatedAt: string;
    currentAssignment?: {
      __typename: "PositionAssignment";
      assignmentId: string;
      positionCode: string;
      positionRecordId: string;
      employeeId: string;
      employeeName: string;
      employeeNumber?: string | null;
      assignmentType: PositionAssignmentType;
      assignmentStatus: PositionAssignmentStatus;
      fte: number;
      startDate: string;
      endDate?: string | null;
      isCurrent: boolean;
      notes?: string | null;
      createdAt: string;
      updatedAt: string;
    } | null;
  } | null;
  positionTimeline: Array<{
    __typename: "PositionTimelineEntry";
    recordId: string;
    status: PositionStatus;
    title: string;
    effectiveDate: string;
    endDate?: string | null;
    changeReason?: string | null;
    isCurrent: boolean;
  }>;
  positionAssignments: {
    __typename: "PositionAssignmentConnection";
    data: Array<{
      __typename: "PositionAssignment";
      assignmentId: string;
      positionCode: string;
      positionRecordId: string;
      employeeId: string;
      employeeName: string;
      employeeNumber?: string | null;
      assignmentType: PositionAssignmentType;
      assignmentStatus: PositionAssignmentStatus;
      fte: number;
      startDate: string;
      endDate?: string | null;
      isCurrent: boolean;
      notes?: string | null;
      createdAt: string;
      updatedAt: string;
    }>;
  };
  positionTransfers: {
    __typename: "PositionTransferConnection";
    data: Array<{
      __typename: "PositionTransfer";
      transferId: string;
      positionCode: string;
      fromOrganizationCode: string;
      toOrganizationCode: string;
      effectiveDate: string;
      operationReason?: string | null;
      createdAt: string;
      initiatedBy: { __typename: "OperatedBy"; id: string; name: string };
    }>;
  };
  positionVersions: Array<{
    __typename: "Position";
    recordId: string;
    code: string;
    title: string;
    jobFamilyGroupCode: string;
    jobFamilyCode: string;
    jobRoleCode: string;
    jobLevelCode: string;
    organizationCode: string;
    organizationName?: string | null;
    positionType: PositionType;
    employmentType: EmploymentType;
    gradeLevel?: string | null;
    headcountCapacity: number;
    headcountInUse: number;
    availableHeadcount: number;
    reportsToPositionCode?: string | null;
    status: PositionStatus;
    effectiveDate: string;
    endDate?: string | null;
    isCurrent: boolean;
    createdAt: string;
    updatedAt: string;
  }>;
};

export type VacantPositionsQueryVariables = Exact<{
  filter?: InputMaybe<VacantPositionFilterInput>;
  pagination?: InputMaybe<PaginationInput>;
  sorting?: InputMaybe<
    Array<VacantPositionSortInput> | VacantPositionSortInput
  >;
}>;

export type VacantPositionsQuery = {
  __typename: "Query";
  vacantPositions: {
    __typename: "VacantPositionConnection";
    totalCount: number;
    data: Array<{
      __typename: "VacantPosition";
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
    }>;
    pagination: {
      __typename: "PaginationInfo";
      total: number;
      page: number;
      pageSize: number;
      hasNext: boolean;
      hasPrevious: boolean;
    };
  };
};

export type PositionHeadcountStatsQueryVariables = Exact<{
  organizationCode: Scalars["String"]["input"];
  includeSubordinates?: InputMaybe<Scalars["Boolean"]["input"]>;
}>;

export type PositionHeadcountStatsQuery = {
  __typename: "Query";
  positionHeadcountStats: {
    __typename: "HeadcountStats";
    organizationCode: string;
    organizationName: string;
    totalCapacity: number;
    totalFilled: number;
    totalAvailable: number;
    fillRate: number;
    byLevel: Array<{
      __typename: "LevelHeadcount";
      jobLevelCode: string;
      capacity: number;
      utilized: number;
      available: number;
    }>;
    byType: Array<{
      __typename: "TypeHeadcount";
      positionType: PositionType;
      capacity: number;
      filled: number;
      available: number;
    }>;
    byFamily: Array<{
      __typename: "FamilyHeadcount";
      jobFamilyCode: string;
      jobFamilyName?: string | null;
      capacity: number;
      utilized: number;
      available: number;
    }>;
  };
};
