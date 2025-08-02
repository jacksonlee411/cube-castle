import axios, { AxiosInstance } from 'axios'
import { Organization, OrganizationStats } from '@/types'

// CQRS 查询客户端 - 专门处理读操作
class OrganizationQueryService {
  private client: AxiosInstance

  constructor(baseURL: string = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080') {
    this.client = axios.create({
      baseURL: `${baseURL}/api/v1/queries`,
      timeout: 15000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // 请求拦截器
    this.client.interceptors.request.use((config) => {
      const token = localStorage.getItem('auth_token')
      const tenantId = localStorage.getItem('tenant_id') || '00000000-0000-0000-0000-000000000001'
      
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      
      // 查询端点使用查询参数传递租户ID
      if (!config.params) {
        config.params = {}
      }
      config.params.tenant_id = tenantId
      
      return config
    })

    // 响应拦截器
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        console.error('Query API Error:', error)
        throw error
      }
    )
  }

  /**
   * 获取组织架构图 - 层级树形结构
   */
  async getOrganizationChart(params: {
    root_unit_id?: string
    max_depth?: number
    include_inactive?: boolean
  } = {}): Promise<{
    chart: Organization[]
    metadata: {
      total_units: number
      max_depth: number
      total_employees: number
    }
  }> {
    console.log('🔍 CQRS查询: 获取组织架构图', params)
    
    const response = await this.client.get('/organization-chart', { params })
    
    console.log('✅ 组织架构图查询成功:', response.data)
    return response.data
  }

  /**
   * 列出组织单元 - 扁平列表
   */
  async listOrganizationUnits(params: {
    unit_type?: string
    parent_id?: string
    status?: string
    limit?: number
    offset?: number
  } = {}): Promise<{
    units: Organization[]
    pagination: {
      limit: number
      offset: number
      total: number
      has_more: boolean
    }
  }> {
    console.log('🔍 CQRS查询: 列出组织单元', params)
    
    const response = await this.client.get('/organization-units', { params })
    
    console.log('✅ 组织单元列表查询成功:', response.data)
    return response.data
  }

  /**
   * 获取单个组织单元详情
   */
  async getOrganizationUnit(id: string): Promise<Organization> {
    console.log('🔍 CQRS查询: 获取组织单元详情', { id })
    
    const response = await this.client.get(`/organization-units/${id}`)
    
    console.log('✅ 组织单元详情查询成功:', response.data)
    return response.data
  }

  /**
   * 获取汇报层级关系
   */
  async getReportingHierarchy(managerId: string, params: {
    max_depth?: number
    include_positions?: boolean
  } = {}): Promise<{
    manager: Organization
    subordinates: Organization[]
    hierarchy_depth: number
    total_reports: number
  }> {
    console.log('🔍 CQRS查询: 获取汇报层级', { managerId, ...params })
    
    const response = await this.client.get(`/reporting-hierarchy/${managerId}`, { params })
    
    console.log('✅ 汇报层级查询成功:', response.data)
    return response.data
  }

  /**
   * 搜索组织单元
   */
  async searchOrganizationUnits(params: {
    query: string
    unit_type?: string
    status?: string
    include_children?: boolean
    limit?: number
    offset?: number
  }): Promise<{
    results: Organization[]
    total_matches: number
    search_time_ms: number
  }> {
    console.log('🔍 CQRS查询: 搜索组织单元', params)
    
    const response = await this.client.get('/organization-units/search', { params })
    
    console.log('✅ 组织搜索查询成功:', response.data)
    return response.data
  }

  /**
   * 获取部门结构分析
   */
  async getDepartmentStructure(deptId: string, params: {
    include_analytics?: boolean
    include_employee_distribution?: boolean
  } = {}): Promise<{
    department: Organization
    structure: {
      total_levels: number
      total_units: number
      units_by_level: Record<number, number>
      employee_distribution: Record<string, number>
    }
    analytics?: {
      occupancy_rate: number
      span_of_control: number
      organizational_health_score: number
    }
  }> {
    console.log('🔍 CQRS查询: 获取部门结构分析', { deptId, ...params })
    
    const response = await this.client.get(`/department-structure/${deptId}`, { params })
    
    console.log('✅ 部门结构分析查询成功:', response.data)
    return response.data
  }

  /**
   * 查找共同管理者
   */
  async findCommonManager(employeeIds: string[]): Promise<{
    common_manager: Organization | null
    hierarchy_path: Organization[]
    relationship_type: 'direct' | 'indirect' | 'none'
  }> {
    console.log('🔍 CQRS查询: 查找共同管理者', { employeeIds })
    
    const response = await this.client.post('/common-manager', { employee_ids: employeeIds })
    
    console.log('✅ 共同管理者查询成功:', response.data)
    return response.data
  }

  /**
   * 查找员工之间的组织路径
   */
  async findEmployeePath(fromId: string, toId: string): Promise<{
    path: Organization[]
    path_length: number
    relationship_type: 'peer' | 'supervisor' | 'subordinate' | 'cross_department'
  }> {
    console.log('🔍 CQRS查询: 查找员工路径', { fromId, toId })
    
    const response = await this.client.get(`/employee-path/${fromId}/${toId}`)
    
    console.log('✅ 员工路径查询成功:', response.data)
    return response.data
  }

  /**
   * 获取组织统计和分析
   */
  async getOrganizationAnalytics(params: {
    unit_id?: string
    time_range?: 'week' | 'month' | 'quarter' | 'year'
    include_trends?: boolean
  } = {}): Promise<{
    summary: OrganizationStats
    trends?: {
      growth_rate: number
      turnover_rate: number
      organizational_changes: number
    }
    unit_type_distribution: Array<{
      unit_type: string
      count: number
      percentage: number
    }>
    level_distribution: Array<{
      level: number
      count: number
      avg_employees: number
    }>
  }> {
    console.log('🔍 CQRS查询: 获取组织分析', params)
    
    const response = await this.client.get('/organization-analytics', { params })
    
    console.log('✅ 组织分析查询成功:', response.data)
    return response.data
  }

  /**
   * 获取实时组织指标
   */
  async getRealtimeMetrics(): Promise<{
    active_organizations: number
    total_employees: number
    recent_changes: number
    system_health: 'healthy' | 'degraded' | 'critical'
    last_updated: string
  }> {
    console.log('🔍 CQRS查询: 获取实时指标')
    
    const response = await this.client.get('/realtime-metrics')
    
    console.log('✅ 实时指标查询成功:', response.data)
    return response.data
  }
}

// 导出单例实例
export const organizationQueries = new OrganizationQueryService()

// 默认导出
export default organizationQueries