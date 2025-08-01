// GraphQL特定类型定义
export interface GraphQLPageInfo {
  hasNextPage: boolean;
  hasPreviousPage: boolean;
  startCursor: string | null;
  endCursor: string | null;
}

export interface GraphQLEdge<T> {
  node: T;
  cursor: string;
}

export interface GraphQLConnection<T> {
  edges: GraphQLEdge<T>[];
  pageInfo: GraphQLPageInfo;
  totalCount: number;
}

// Employee GraphQL Types
export interface GraphQLEmployee {
  id: string;
  employeeId: string;
  legalName: string;
  preferredName?: string;
  email: string;
  status: string;
  hireDate: string;
  terminationDate?: string;
  currentPosition?: {
    positionTitle: string;
    department: string;
    jobLevel?: string;
    location?: string;
    employmentType: string;
    effectiveDate: string;
  };
}

export interface GraphQLEmployeeFilters {
  search?: string;
  department?: string;
  status?: string;
  employmentType?: string;
  managerId?: string;
  hiredAfter?: string;
  hiredBefore?: string;
}

export interface GraphQLEmployeesResponse {
  employees: GraphQLConnection<GraphQLEmployee>;
}

export interface GraphQLEmployeeResponse {
  employee: GraphQLEmployee;
}

// Position Change Types
export interface GraphQLPositionChangeInput {
  employeeId: string;
  positionData: {
    positionTitle: string;
    department: string;
    jobLevel?: string;
    location?: string;
    employmentType: string;
    reportsToEmployeeId?: string;
    minSalary?: number;
    maxSalary?: number;
    currency?: string;
  };
  effectiveDate: string;
  changeReason?: string;
  isRetroactive?: boolean;
}

export interface GraphQLPositionChangeValidation {
  isValid: boolean;
  errors?: Array<{ message: string }>;
  warnings?: Array<{ message: string }>;
}

export interface GraphQLPositionChangeResult {
  workflowId: string;
  positionHistory: any;
  errors?: Array<{ message: string }>;
}

// Subscription Types
export interface GraphQLEmployeePositionChangedPayload {
  employeePositionChanged: GraphQLEmployee;
}

export interface GraphQLWorkflowStatusChangedPayload {
  workflowStatusChanged: {
    workflowId: string;
    status: string;
    error?: string;
  };
}

export interface GraphQLPositionChangeApprovalRequiredPayload {
  positionChangeApprovalRequired: {
    workflowId: string;
    employeeId: string;
  };
}

// Workflow Types
export interface GraphQLWorkflowStatus {
  status: string;
  workflowId: string;
  workflowName?: string;
  progress?: number;
  error?: string;
}

// Position Timeline Types
export interface GraphQLPositionTimelineEntry {
  id: string;
  positionTitle: string;
  department: string;
  effectiveDate: string;
  endDate?: string;
  changeReason?: string;
  workflowId?: string;
}

export interface GraphQLPositionTimelineResponse {
  positionTimeline: GraphQLPositionTimelineEntry[];
}

// Mutation Result Types
export interface GraphQLMutationResult<T = any> {
  success: boolean;
  errors?: Array<{ message: string }>;
  data?: T;
}

export interface GraphQLApprovePositionChangeResult {
  workflowId: string;
  errors?: Array<{ message: string }>;
}

export interface GraphQLRejectPositionChangeResult {
  workflowId: string;
  errors?: Array<{ message: string }>;
}