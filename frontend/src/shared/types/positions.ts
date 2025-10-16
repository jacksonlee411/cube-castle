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
  fetchedAt: string;
}
