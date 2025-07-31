// 基础类型定义
export interface BaseEntity {
  id: string
  createdAt: string
  updatedAt: string
}

// 分页信息
export interface PaginationInfo {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

// 导出职位相关类型
export * from './position'

// 员工相关类型
export interface Employee extends BaseEntity {
  employeeNumber: string
  firstName: string
  lastName: string
  fullName: string
  email: string
  phoneNumber?: string
  hireDate: string
  status: EmployeeStatus
  jobTitle?: string
  department?: string
  managerId?: string
  manager?: Employee
  organizationId?: string
  organization?: Organization
  tenantId: string
}

// 扩展Employee类型以支持实际API返回的字段格式
export interface EmployeeApiResponse extends BaseEntity {
  employee_number: string
  first_name: string
  last_name: string
  email: string
  phone_number?: string
  hire_date: string
  status: EmployeeStatus
  job_title?: string
  organization_id?: string
  position_id?: string
  manager_id?: string
  tenant_id: string
}

// 类型转换工具函数类型
export type EmployeeConverter = {
  fromApi: (apiData: EmployeeApiResponse) => Employee
  toApi: (employee: Partial<Employee>) => Partial<EmployeeApiResponse>
}

export enum EmployeeStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  TERMINATED = 'terminated',
  ON_LEAVE = 'on_leave'
}

export interface CreateEmployeeRequest {
  employeeNumber: string
  firstName: string
  lastName: string
  email: string
  phoneNumber?: string
  hireDate: string
  jobTitle?: string
  organizationId?: string
  managerId?: string
}

export interface UpdateEmployeeRequest {
  firstName?: string
  lastName?: string
  email?: string
  phoneNumber?: string
  jobTitle?: string
  organizationId?: string
  managerId?: string
  status?: EmployeeStatus
}

export interface EmployeeListResponse {
  employees: Employee[]
  pagination: PaginationInfo
}

// 组织架构相关类型
export interface Organization extends BaseEntity {
  name: string
  code: string
  description?: string
  level: number
  parentId?: string
  parent?: Organization
  children?: Organization[]
  employeeCount?: number
  tenantId: string
  type?: 'company' | 'department' | 'team'
  status?: 'active' | 'inactive'
  managerName?: string
}

export interface OrganizationTreeNode {
  id: string
  name: string
  code: string
  level: number
  parentId?: string
  children: OrganizationTreeNode[]
  employeeCount: number
}

export interface OrganizationListResponse {
  organizations: Organization[]
  pagination: PaginationInfo
}

export interface OrganizationTreeResponse {
  tree: OrganizationTreeNode[]
}

export interface CreateOrganizationRequest {
  name: string
  code: string
  description?: string
  parentId?: string
}

export interface UpdateOrganizationRequest {
  name?: string
  code?: string
  description?: string
  parentId?: string
}

// 租户相关类型
export interface Tenant extends BaseEntity {
  name: string
  domain: string
  status: TenantStatus
  settings?: TenantSettings
  plan?: string
  expiresAt?: string
}

export enum TenantStatus {
  ACTIVE = 'active',
  SUSPENDED = 'suspended',
  EXPIRED = 'expired'
}

export interface TenantSettings {
  timezone: string
  dateFormat: string
  language: string
  currency: string
  features: string[]
}

// 用户相关类型
export interface User extends BaseEntity {
  username: string
  email: string
  firstName: string
  lastName: string
  fullName: string
  roles: UserRole[]
  status: UserStatus
  lastLoginAt?: string
  tenantId: string
  employeeId?: string
  employee?: Employee
}

export interface UserRole {
  id: string
  name: string
  description?: string
  permissions: Permission[]
}

export interface Permission {
  id: string
  name: string
  resource: string
  action: string
}

export enum UserStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  LOCKED = 'locked'
}

// 工作流相关类型
export interface WorkflowInstance extends BaseEntity {
  workflowId: string
  workflowName: string
  status: WorkflowStatus
  progress: number
  startedAt: string
  completedAt?: string
  createdBy: string
  assignedTo?: string
  data: Record<string, any>
  steps: WorkflowStep[]
  tenantId: string
}

export enum WorkflowStatus {
  PENDING = 'pending',
  RUNNING = 'running',
  COMPLETED = 'completed',
  FAILED = 'failed',
  CANCELLED = 'cancelled'
}

export interface WorkflowStep {
  id: string
  name: string
  status: WorkflowStepStatus
  startedAt?: string
  completedAt?: string
  error?: string
  data?: Record<string, any>
}

export enum WorkflowStepStatus {
  PENDING = 'pending',
  RUNNING = 'running',
  COMPLETED = 'completed',
  FAILED = 'failed',
  SKIPPED = 'skipped'
}

export interface WorkflowStatsResponse {
  totalInstances: number
  activeInstances: number
  completedInstances: number
  failedInstances: number
  averageCompletionTime: number
  statusDistribution: Record<WorkflowStatus, number>
}

// Intelligence Gateway 相关类型
export interface InterpretRequest {
  text: string
  context?: Record<string, any>
  sessionId?: string
}

export interface InterpretResponse {
  intent: string
  confidence: number
  entities: IntentEntity[]
  response: string
  suggestions?: string[]
  sessionId: string
}

export interface IntentEntity {
  type: string
  value: string
  start: number
  end: number
  confidence: number
}

// 系统监控相关类型
export interface SystemHealth {
  status: 'healthy' | 'unhealthy' | 'degraded'
  timestamp: string
  version: string
  environment: string
  services: ServiceHealth[]
  metrics: SystemMetrics
}

export interface ServiceHealth {
  name: string
  status: 'healthy' | 'unhealthy'
  latency: number
  message?: string
}

export interface SystemMetrics {
  memoryUsage: number
  cpuUsage: number
  diskUsage: number
  activeConnections: number
  requestsPerSecond: number
  errorRate: number
}

export interface BusinessMetrics {
  totalEmployees: number
  activeEmployees: number
  totalOrganizations: number
  workflowsCompleted: number
  aiQueriesProcessed: number
  lastUpdated: string
}

// 通用响应类型
export interface GeneralResponse {
  message: string
  success: boolean
}

export interface ErrorResponse {
  code: string
  message: string
  details?: Record<string, any>
}

// 表单相关类型
export interface FormField {
  name: string
  label: string
  type: 'text' | 'email' | 'password' | 'select' | 'textarea' | 'date' | 'number'
  required?: boolean
  placeholder?: string
  options?: Array<{ label: string; value: string }>
  validation?: {
    pattern?: string
    minLength?: number
    maxLength?: number
    min?: number
    max?: number
  }
}

// 表格相关类型
export interface TableColumn<T = any> {
  key: keyof T
  title: string
  width?: string
  sortable?: boolean
  filterable?: boolean
  render?: (value: any, record: T) => React.ReactNode
}

export interface TableProps<T = any> {
  columns: TableColumn<T>[]
  data: T[]
  loading?: boolean
  pagination?: {
    current: number
    pageSize: number
    total: number
    onChange: (page: number, pageSize: number) => void
  }
  onSort?: (key: keyof T, direction: 'asc' | 'desc') => void
  onFilter?: (filters: Record<keyof T, any>) => void
  onRowClick?: (record: T) => void
}

// 导航相关类型
export interface NavigationItem {
  id: string
  title: string
  href: string
  icon?: React.ComponentType<{ className?: string }>
  badge?: string | number
  children?: NavigationItem[]
  permission?: string
}

// 主题相关类型
export type Theme = 'light' | 'dark' | 'system'

// 应用状态类型
export interface AppState {
  user: User | null
  tenant: Tenant | null
  theme: Theme
  sidebarOpen: boolean
  notifications: Notification[]
}

export interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message: string
  timestamp: string
  read: boolean
  action?: {
    label: string
    href: string
  }
}