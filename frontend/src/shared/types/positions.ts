export type PositionStatus =
  | 'PLANNED'
  | 'ACTIVE'
  | 'FILLED'
  | 'VACANT'
  | 'INACTIVE'
  | 'SUSPENDED'
  | 'DELETED'
  | string;

export interface PositionRecord {
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
}

export interface PositionTimelineEvent {
  id: string;
  status: PositionStatus;
  title: string;
  effectiveDate: string;
  endDate?: string | null;
  changeReason?: string | null;
  isCurrent?: boolean;
}

export interface PositionAssignmentRecord {
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

export interface PositionTransferRecord {
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

export interface PositionsQueryResult {
  positions: PositionRecord[];
  pagination: {
    total: number;
    page: number;
    pageSize: number;
    hasNext: boolean;
    hasPrevious: boolean;
  };
  totalCount: number;
  timestamp: string;
}

export interface PositionDetailResult {
  position: PositionRecord;
  timeline: PositionTimelineEvent[];
  currentAssignment?: PositionAssignmentRecord | null;
  assignments: PositionAssignmentRecord[];
  transfers: PositionTransferRecord[];
  versions: PositionRecord[];
  fetchedAt: string;
}

export interface VacantPositionRecord {
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

export interface VacantPositionsQueryResult {
  data: VacantPositionRecord[];
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

export interface PositionHeadcountLevelBreakdown {
  jobLevelCode: string;
  capacity: number;
  utilized: number;
  available: number;
}

export interface PositionHeadcountTypeBreakdown {
  positionType: string;
  capacity: number;
  filled: number;
  available: number;
}

export interface PositionHeadcountFamilyBreakdown {
  jobFamilyCode: string;
  jobFamilyName?: string | null;
  capacity: number;
  utilized: number;
  available: number;
}

export interface PositionHeadcountStats {
  organizationCode: string;
  organizationName: string;
  totalCapacity: number;
  totalFilled: number;
  totalAvailable: number;
  fillRate: number;
  byLevel: PositionHeadcountLevelBreakdown[];
  byType: PositionHeadcountTypeBreakdown[];
  byFamily: PositionHeadcountFamilyBreakdown[];
  fetchedAt: string;
}

export interface PositionMutationRequest {
  title: string;
  jobFamilyGroupCode: string;
  jobFamilyCode: string;
  jobRoleCode: string;
  jobLevelCode: string;
  organizationCode: string;
  positionType: string;
  employmentType: string;
  headcountCapacity: number;
  effectiveDate: string;
  operationReason: string;
  jobProfileCode?: string | null;
  jobProfileName?: string | null;
  gradeLevel?: string | null;
  reportsToPositionCode?: string | null;
}

export interface UpdatePositionRequest extends PositionMutationRequest {
  code: string;
}

export interface CreatePositionVersionRequest extends PositionMutationRequest {
  code: string;
}

export type CreatePositionRequest = PositionMutationRequest;

export interface PositionResource extends PositionRecord {
  recordId: string;
  currentAssignment?: PositionAssignmentRecord | null;
  assignmentHistory?: PositionAssignmentRecord[];
}
