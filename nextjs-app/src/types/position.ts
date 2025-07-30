// 职位相关类型定义

export interface Position extends BaseEntity {
  positionType: PositionType
  jobProfileId: string
  departmentId: string
  managerPositionId?: string
  status: PositionStatus
  budgetedFte: number
  details?: Record<string, any>
  
  // 关联数据
  department?: Organization
  managerPosition?: Position
  directReports?: Position[]
  occupancyHistory?: PositionOccupancyHistory[]
  
  tenantId: string
}

export enum PositionType {
  FULL_TIME = 'FULL_TIME',
  PART_TIME = 'PART_TIME', 
  CONTINGENT_WORKER = 'CONTINGENT_WORKER',
  INTERN = 'INTERN'
}

export enum PositionStatus {
  OPEN = 'OPEN',
  FILLED = 'FILLED',
  FROZEN = 'FROZEN',
  PENDING_ELIMINATION = 'PENDING_ELIMINATION'
}

export interface PositionOccupancyHistory extends BaseEntity {
  positionId: string
  employeeId: string
  startDate: string
  endDate?: string
  isActive: boolean
  
  // 关联数据
  position?: Position
  employee?: Employee
}

export interface CreatePositionRequest {
  positionType: PositionType
  jobProfileId: string
  departmentId: string
  managerPositionId?: string
  status?: PositionStatus
  budgetedFte?: number
  details?: Record<string, any>
}

export interface UpdatePositionRequest {
  jobProfileId?: string
  departmentId?: string
  managerPositionId?: string
  status?: PositionStatus
  budgetedFte?: number
  details?: Record<string, any>
}

export interface PositionListResponse {
  positions: Position[]
  pagination: PaginationInfo
}

// 职位层级树节点
export interface PositionTreeNode {
  id: string
  positionType: PositionType
  status: PositionStatus
  budgetedFte: number
  departmentName: string
  managerPositionId?: string
  children: PositionTreeNode[]
  occupancy?: {
    isOccupied: boolean
    currentEmployee?: Employee
  }
}

export interface PositionTreeResponse {
  tree: PositionTreeNode[]
}

// 职位统计信息
export interface PositionStats {
  totalPositions: number
  openPositions: number
  filledPositions: number
  frozenPositions: number
  pendingEliminationPositions: number
  totalBudgetedFte: number
  actualFte: number
  utilizationRate: number
  statusDistribution: Record<PositionStatus, number>
  typeDistribution: Record<PositionType, number>
}