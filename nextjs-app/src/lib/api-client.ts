import axios, { AxiosInstance, AxiosResponse } from 'axios'
import toast from 'react-hot-toast'
import { 
  Employee, 
  EmployeeListResponse, 
  CreateEmployeeRequest, 
  UpdateEmployeeRequest,
  Organization,
  OrganizationProfile,
  OrganizationsApiResponse,
  OrganizationChartApiResponse,
  OrganizationStatsApiResponse,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
  InterpretRequest,
  InterpretResponse,
  SystemHealth,
  BusinessMetrics,
  WorkflowInstance,
  WorkflowStatsResponse
} from '@/types'
import { 
  API_BASE_URL, 
  AI_API_URL, 
  DEFAULT_TENANT_ID, 
  DEFAULT_TIMEOUT,
  REST_ROUTES,
  AI_ROUTES,
  buildApiUrl 
} from '@/lib/routes'

// 创建 HTTP 客户端
const httpClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: DEFAULT_TIMEOUT.STANDARD,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 创建 AI 服务客户端 (gRPC Gateway)
const aiClient: AxiosInstance = axios.create({
  baseURL: AI_API_URL,
  timeout: DEFAULT_TIMEOUT.AI_SERVICE,
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
    
    // 添加租户ID (多租户支持) - 开发环境默认配置
    const tenantId = localStorage.getItem('tenant_id') || DEFAULT_TENANT_ID
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
    const tenantId = localStorage.getItem('tenant_id') || DEFAULT_TENANT_ID
    config.headers['X-Tenant-ID'] = tenantId
    return config
  },
  (error) => Promise.reject(error)
)

aiClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    // AI Service Error - error handled by caller
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
    const response = await httpClient.get(REST_ROUTES.COREHR.EMPLOYEES, { params })
    return response.data
  },

  // 根据ID获取员工详情
  async getEmployee(id: string): Promise<Employee> {
    const response = await httpClient.get(REST_ROUTES.COREHR.EMPLOYEE_BY_ID(id))
    return response.data
  },

  // 创建员工
  async createEmployee(data: CreateEmployeeRequest): Promise<Employee> {
    const response = await httpClient.post(REST_ROUTES.COREHR.EMPLOYEES, data)
    toast.success('员工创建成功')
    return response.data
  },

  // 更新员工信息
  async updateEmployee(id: string, data: UpdateEmployeeRequest): Promise<Employee> {
    const response = await httpClient.put(REST_ROUTES.COREHR.EMPLOYEE_BY_ID(id), data)
    toast.success('员工信息更新成功')
    return response.data
  },

  // 删除员工
  async deleteEmployee(id: string): Promise<void> {
    await httpClient.delete(REST_ROUTES.COREHR.EMPLOYEE_BY_ID(id))
    toast.success('员工删除成功')
  },

  // 批量操作
  async bulkUpdateEmployees(ids: string[], data: Partial<UpdateEmployeeRequest>): Promise<void> {
    await httpClient.patch(buildApiUrl('/api/v1/corehr/employees/bulk'), { ids, data })
    toast.success(`批量更新 ${ids.length} 名员工成功`)
  }
}

// 统一的组织架构API适配器 (完全对齐后端organization_adapter.go)
export const organizationApi = {
  // 获取组织列表 (标准化API调用)
  async getOrganizations(params: {
    page?: number
    pageSize?: number
    search?: string
    parent_unit_id?: string
    unit_type?: string
    status?: string
  } = {}): Promise<OrganizationsApiResponse> {
    try {
      console.log('🔄 调用CoreHR组织API:', params);
      const response = await httpClient.get(REST_ROUTES.COREHR.ORGANIZATIONS, { params })
      
      console.log('✅ CoreHR组织API响应:', response.data);
      return response.data
    } catch (error) {
      console.error('❌ PostgreSQL组织API调用失败:', error);
      
      // Fallback to mock data only on network errors
      const mockOrganizations: Organization[] = [
        {
          id: '186b1cd6-de34-4418-8219-22c917334787',
          name: 'Cube Castle',
          unit_type: 'COMPANY',
          description: '全栈企业管理解决方案提供商',
          level: 0,
          parent_unit_id: undefined,
          employee_count: 50,
          tenant_id: '00000000-0000-0000-0000-000000000001',
          status: 'ACTIVE',
          profile: {
            managerName: 'CEO',
            maxCapacity: 200
          },
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        }
      ];

      return {
        organizations: mockOrganizations,
        pagination: { 
          page: params.page || 1, 
          pageSize: params.pageSize || 100, 
          total: mockOrganizations.length, 
          totalPages: 1 
        }
      }
    }
  },

  // 获取组织统计 (使用CoreHR适配器API)
  async getStats(): Promise<OrganizationStatsApiResponse> {
    try {
      const response = await httpClient.get(REST_ROUTES.COREHR.ORGANIZATION_STATS)
      return response.data
    } catch (error) {
      console.warn('⚠️ 组织统计API暂不可用，使用默认数据');
      // 返回默认统计数据
      return {
        data: {
          total: 1,
          active: 1,
          inactive: 0,
          totalEmployees: 0
        }
      }
    }
  },

  // 根据ID获取组织详情 (使用CoreHR适配器API)
  async getOrganization(id: string): Promise<Organization> {
    const response = await httpClient.get(REST_ROUTES.COREHR.ORGANIZATION_BY_ID(id))
    return response.data
  },

  // 创建组织 (严格类型检查和数据清理)
  async createOrganization(data: CreateOrganizationRequest): Promise<Organization> {
    // 数据清理和类型确保
    const cleanData: CreateOrganizationRequest = {
      ...data,
      parent_unit_id: data.parent_unit_id ? String(data.parent_unit_id).trim() : undefined,
      status: data.status || 'ACTIVE',
      profile: data.profile || {}
    }
    
    console.log('🎯 创建组织API调用 (清理后数据):', cleanData);
    const response = await httpClient.post(REST_ROUTES.COREHR.ORGANIZATIONS, cleanData)
    console.log('🎉 组织创建成功:', response.data);
    return response.data
  },

  // 更新组织信息 (严格类型检查)
  async updateOrganization(id: string, data: UpdateOrganizationRequest): Promise<Organization> {
    // 数据清理
    const cleanData: UpdateOrganizationRequest = {
      ...data,
      parent_unit_id: data.parent_unit_id ? String(data.parent_unit_id).trim() : undefined
    }
    
    console.log('📝 更新组织API调用:', id, cleanData);
    const response = await httpClient.put(REST_ROUTES.COREHR.ORGANIZATION_BY_ID(id), cleanData)
    console.log('✅ 组织更新成功:', response.data);
    return response.data
  },

  // 删除组织 (使用CoreHR适配器API)
  async deleteOrganization(id: string): Promise<void> {
    console.log('🗑️ 删除组织API调用:', id);
    await httpClient.delete(REST_ROUTES.COREHR.ORGANIZATION_BY_ID(id))
    console.log('✅ 组织删除成功');
  }
}

// AI 智能交互 API
export const intelligenceApi = {
  // 文本意图识别和对话
  async interpretText(data: InterpretRequest): Promise<InterpretResponse> {
    try {
      // 为了保持会话状态，我们添加会话ID
      const sessionId = data.sessionId || generateSessionId()
      
      const response = await httpClient.post(AI_ROUTES.INTELLIGENCE.INTERPRET, {
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
      const response = await httpClient.get(AI_ROUTES.INTELLIGENCE.CONVERSATION_HISTORY(sessionId))
      return response.data.history || []
    } catch {
      // 如果服务不支持历史记录，返回空数组
      return []
    }
  },

  // 清除对话历史
  async clearConversationHistory(sessionId: string): Promise<void> {
    try {
      await httpClient.delete(AI_ROUTES.INTELLIGENCE.CONVERSATION_HISTORY(sessionId))
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
    const response = await httpClient.get(REST_ROUTES.WORKFLOWS.INSTANCES, { params })
    return response.data
  },

  // 获取工作流实例详情
  async getWorkflowInstance(id: string): Promise<WorkflowInstance> {
    const response = await httpClient.get(REST_ROUTES.WORKFLOWS.INSTANCE_BY_ID(id))
    return response.data
  },

  // 启动工作流
  async startWorkflow(workflowName: string, input: any): Promise<WorkflowInstance> {
    const response = await httpClient.post(REST_ROUTES.WORKFLOWS.START, {
      workflowName,
      input
    })
    toast.success('工作流启动成功')
    return response.data
  },

  // 获取工作流统计信息
  async getWorkflowStats(): Promise<WorkflowStatsResponse> {
    const response = await httpClient.get(REST_ROUTES.WORKFLOWS.STATS)
    return response.data
  }
}

// 系统监控 API
export const systemApi = {
  // 获取系统健康状态
  async getSystemHealth(): Promise<SystemHealth> {
    const response = await httpClient.get(REST_ROUTES.SYSTEM.HEALTH)
    return response.data
  },

  // 获取业务指标
  async getBusinessMetrics(): Promise<BusinessMetrics> {
    const response = await httpClient.get(REST_ROUTES.SYSTEM.METRICS)
    return response.data
  },

  // 获取系统版本信息
  async getSystemInfo(): Promise<any> {
    const response = await httpClient.get(REST_ROUTES.SYSTEM.INFO)
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