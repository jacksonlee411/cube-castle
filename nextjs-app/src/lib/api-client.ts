import axios, { AxiosInstance, AxiosResponse } from 'axios'
import toast from 'react-hot-toast'
import { 
  Employee, 
  EmployeeListResponse, 
  CreateEmployeeRequest, 
  UpdateEmployeeRequest,
  Organization,
  OrganizationListResponse,
  OrganizationTreeResponse,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
  InterpretRequest,
  InterpretResponse,
  SystemHealth,
  BusinessMetrics,
  WorkflowInstance,
  WorkflowStatsResponse
} from '@/types'

// API 基础配置
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
const AI_API_URL = process.env.NEXT_PUBLIC_AI_API_URL || 'http://localhost:8081'

// 创建 HTTP 客户端
const httpClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 创建 AI 服务客户端 (gRPC Gateway)
const aiClient: AxiosInstance = axios.create({
  baseURL: AI_API_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器 - 添加认证信息
httpClient.interceptors.request.use(
  (config) => {
    // 从localStorage获取token (后续实现认证时使用)
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    
    // 添加租户ID (多租户支持)
    const tenantId = localStorage.getItem('tenant_id') || 'default'
    config.headers['X-Tenant-ID'] = tenantId
    
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器 - 统一错误处理
httpClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    // 检查是否是已知的API未实现错误，不显示toast
    const isKnownUnimplementedAPI = error.config?.url?.includes('/organizations/stats') || 
                                    error.config?.url?.includes('/intelligence/') ||
                                    (error.response?.status === 404 && error.config?.url?.includes('/api/v1/corehr/'))
    
    if (isKnownUnimplementedAPI) {
      // 对于已知的未实现API，静默处理
      return Promise.reject(error)
    }
    
    if (error.response?.status === 401) {
      // 未授权，跳转到登录页
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    } else if (error.response?.status >= 500) {
      // 服务器错误
      toast.error('服务器错误，请稍后重试')
    } else if (error.response?.data?.message) {
      // 业务错误
      toast.error(error.response.data.message)
    } else if (!isKnownUnimplementedAPI) {
      // 其他网络错误（排除已知的未实现API）
      toast.error('网络连接失败，请检查网络设置')
    }
    return Promise.reject(error)
  }
)

// AI 客户端拦截器
aiClient.interceptors.request.use(
  (config) => {
    const tenantId = localStorage.getItem('tenant_id') || 'default'
    config.headers['X-Tenant-ID'] = tenantId
    return config
  },
  (error) => Promise.reject(error)
)

aiClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    console.error('AI Service Error:', error)
    if (error.response?.data?.message) {
      toast.error(`AI服务错误: ${error.response.data.message}`)
    } else {
      toast.error('AI服务暂时不可用')
    }
    return Promise.reject(error)
  }
)

// 员工管理 API
export const employeeApi = {
  // 获取员工列表
  async getEmployees(params: {
    page?: number
    pageSize?: number
    search?: string
    status?: string
    organizationId?: string
  } = {}): Promise<EmployeeListResponse> {
    const response = await httpClient.get('/api/v1/corehr/employees', { params })
    return response.data
  },

  // 根据ID获取员工详情
  async getEmployee(id: string): Promise<Employee> {
    const response = await httpClient.get(`/api/v1/corehr/employees/${id}`)
    return response.data
  },

  // 创建员工
  async createEmployee(data: CreateEmployeeRequest): Promise<Employee> {
    const response = await httpClient.post('/api/v1/corehr/employees', data)
    toast.success('员工创建成功')
    return response.data
  },

  // 更新员工信息
  async updateEmployee(id: string, data: UpdateEmployeeRequest): Promise<Employee> {
    const response = await httpClient.put(`/api/v1/corehr/employees/${id}`, data)
    toast.success('员工信息更新成功')
    return response.data
  },

  // 删除员工
  async deleteEmployee(id: string): Promise<void> {
    await httpClient.delete(`/api/v1/corehr/employees/${id}`)
    toast.success('员工删除成功')
  },

  // 批量操作
  async bulkUpdateEmployees(ids: string[], data: Partial<UpdateEmployeeRequest>): Promise<void> {
    await httpClient.patch('/api/v1/corehr/employees/bulk', { ids, data })
    toast.success(`批量更新 ${ids.length} 名员工成功`)
  }
}

// 组织架构 API
export const organizationApi = {
  // 获取组织列表
  async getOrganizations(params: {
    page?: number
    pageSize?: number
    search?: string
    parentId?: string
  } = {}): Promise<OrganizationListResponse> {
    try {
      const response = await httpClient.get('/api/v1/corehr/organizations', { params })
      
      // 检查后端是否返回未实现状态
      if (response.data?.status === 'not_implemented') {
        // 返回Mock数据
        return {
          organizations: [
            {
              id: '1',
              name: '总公司',
              code: 'HQ',
              description: '公司总部',
              level: 1,
              parentId: undefined,
              employeeCount: 10,
              tenantId: 'default',
              type: 'company' as const,
              status: 'active' as const,
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString(),
            },
            {
              id: '2', 
              name: '技术部',
              code: 'TECH',
              description: '技术研发部门',
              level: 2,
              parentId: '1',
              employeeCount: 8,
              tenantId: 'default',
              type: 'department' as const,
              status: 'active' as const,
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString(),
            }
          ],
          pagination: {
            page: 1,
            pageSize: 20,
            total: 2,
            totalPages: 1
          }
        }
      }
      
      return response.data
    } catch (error) {
      // 网络错误时返回空数据
      return {
        organizations: [],
        pagination: { page: 1, pageSize: 20, total: 0, totalPages: 0 }
      }
    }
  },

  // 获取组织列表 (别名方法)
  async getList(params: {
    page?: number
    pageSize?: number
    search?: string
    parentId?: string
  } = {}): Promise<OrganizationListResponse> {
    return this.getOrganizations(params)
  },

  // 获取组织统计
  async getStats(): Promise<any> {
    try {
      const response = await httpClient.get('/api/v1/corehr/organizations/stats')
      
      // 检查后端是否返回未实现状态
      if (response.data?.status === 'not_implemented') {
        return {
          data: {
            total: 2,
            totalEmployees: 4,
            active: 2,
            inactive: 0
          }
        }
      }
      
      return response.data
    } catch (error) {
      // 返回默认统计数据
      return {
        data: {
          total: 0,
          totalEmployees: 0,
          active: 0,
          inactive: 0
        }
      }
    }
  },

  // 获取组织树结构
  async getOrganizationTree(): Promise<OrganizationTreeResponse> {
    const response = await httpClient.get('/api/v1/corehr/organizations/tree')
    return response.data
  },

  // 根据ID获取组织详情
  async getOrganization(id: string): Promise<Organization> {
    const response = await httpClient.get(`/api/v1/corehr/organizations/${id}`)
    return response.data
  },

  // 创建组织
  async createOrganization(data: CreateOrganizationRequest): Promise<Organization> {
    const response = await httpClient.post('/api/v1/corehr/organizations', data)
    toast.success('组织创建成功')
    return response.data
  },

  // 更新组织信息
  async updateOrganization(id: string, data: UpdateOrganizationRequest): Promise<Organization> {
    const response = await httpClient.put(`/api/v1/corehr/organizations/${id}`, data)
    toast.success('组织信息更新成功')
    return response.data
  },

  // 删除组织
  async deleteOrganization(id: string): Promise<void> {
    await httpClient.delete(`/api/v1/corehr/organizations/${id}`)
    toast.success('组织删除成功')
  }
}

// AI 智能交互 API
export const intelligenceApi = {
  // 文本意图识别和对话
  async interpretText(data: InterpretRequest): Promise<InterpretResponse> {
    try {
      // 为了保持会话状态，我们添加会话ID
      const sessionId = data.sessionId || generateSessionId()
      
      const response = await httpClient.post('/api/v1/intelligence/interpret', {
        ...data,
        sessionId
      })
      
      // 检查后端是否返回未实现状态
      if (response.data?.status === 'not_implemented') {
        // 返回Mock AI响应
        return {
          intent: 'general_query',
          confidence: 0.9,
          response: `我理解您说的是："${data.text}"。这是一个模拟的AI回复，实际AI服务正在开发中。`,
          entities: [],
          sessionId,
          suggestions: [
            '您可以尝试询问员工信息',
            '或者查看组织架构',
            '也可以了解系统功能'
          ]
        }
      }
      
      return {
        ...response.data,
        sessionId
      }
    } catch (error) {
      // 网络错误时返回友好的错误回复
      return {
        intent: 'error',
        confidence: 1.0,
        response: '抱歉，AI服务暂时不可用。请稍后再试或联系管理员。',
        entities: [],
        sessionId: data.sessionId || generateSessionId(),
        suggestions: ['请检查网络连接', '稍后重试', '联系技术支持']
      }
    }
  },

  // 获取对话历史 (如果AI服务支持)
  async getConversationHistory(sessionId: string): Promise<any[]> {
    try {
      const response = await httpClient.get(`/api/v1/intelligence/conversations/${sessionId}`)
      return response.data.history || []
    } catch {
      // 如果服务不支持历史记录，返回空数组
      return []
    }
  },

  // 清除对话历史
  async clearConversationHistory(sessionId: string): Promise<void> {
    try {
      await httpClient.delete(`/api/v1/intelligence/conversations/${sessionId}`)
    } catch {
      // 忽略删除失败的情况
    }
  }
}

// 工作流 API
export const workflowApi = {
  // 获取工作流实例列表
  async getWorkflowInstances(params: {
    page?: number
    pageSize?: number
    status?: string
    workflowName?: string
  } = {}): Promise<{ instances: WorkflowInstance[], pagination: any }> {
    const response = await httpClient.get('/api/v1/workflows/instances', { params })
    return response.data
  },

  // 获取工作流实例详情
  async getWorkflowInstance(id: string): Promise<WorkflowInstance> {
    const response = await httpClient.get(`/api/v1/workflows/instances/${id}`)
    return response.data
  },

  // 启动工作流
  async startWorkflow(workflowName: string, input: any): Promise<WorkflowInstance> {
    const response = await httpClient.post('/api/v1/workflows/start', {
      workflowName,
      input
    })
    toast.success('工作流启动成功')
    return response.data
  },

  // 获取工作流统计信息
  async getWorkflowStats(): Promise<WorkflowStatsResponse> {
    const response = await httpClient.get('/api/v1/workflows/stats')
    return response.data
  }
}

// 系统监控 API
export const systemApi = {
  // 获取系统健康状态
  async getSystemHealth(): Promise<SystemHealth> {
    const response = await httpClient.get('/api/v1/system/health')
    return response.data
  },

  // 获取业务指标
  async getBusinessMetrics(): Promise<BusinessMetrics> {
    const response = await httpClient.get('/api/v1/system/metrics/business')
    return response.data
  },

  // 获取系统版本信息
  async getSystemInfo(): Promise<any> {
    const response = await httpClient.get('/api/v1/system/info')
    return response.data
  }
}

// 辅助函数
function generateSessionId(): string {
  return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
}

// 导出所有API
export const apiClient = {
  employees: employeeApi,
  organizations: organizationApi,
  intelligence: intelligenceApi,
  workflows: workflowApi,
  system: systemApi,
  
  // 直接访问HTTP客户端 (用于自定义请求)
  http: httpClient,
  ai: aiClient
}

export default apiClient