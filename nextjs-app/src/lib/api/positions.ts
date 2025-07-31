/**
 * 职位管理API客户端
 * 连接前端Position接口与后端Position API
 */

import { 
  Position, 
  PositionType, 
  PositionStatus, 
  CreatePositionRequest, 
  UpdatePositionRequest, 
  PositionListResponse,
  PositionTreeResponse,
  PositionStats
} from '@/types'

// API基础配置
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'
const POSITIONS_ENDPOINT = '/api/v1/positions'

// 错误类定义
export class PositionApiError extends Error {
  constructor(
    message: string,
    public statusCode: number,
    public code?: string
  ) {
    super(message)
    this.name = 'PositionApiError'
  }
}

// 后端Position API响应类型
interface PositionApiResponse {
  id: string
  tenant_id: string
  position_type: string
  job_profile_id: string
  department_id: string
  manager_position_id?: string
  status: string
  budgeted_fte: number
  details?: Record<string, any>
  created_at: string
  updated_at: string
}

interface PositionCreateRequest {
  position_type: string
  job_profile_id: string
  department_id: string
  manager_position_id?: string
  status: string
  budgeted_fte: number
  details?: Record<string, any>
}

// 数据转换工具类
export class PositionApiAdapter {
  /**
   * 后端Position数据转换为前端Position
   */
  static toPosition(position: PositionApiResponse): Position {
    return {
      id: position.id,
      positionType: position.position_type as PositionType,
      jobProfileId: position.job_profile_id,
      departmentId: position.department_id,
      managerPositionId: position.manager_position_id,
      status: position.status as PositionStatus,
      budgetedFte: position.budgeted_fte,
      details: position.details,
      tenantId: position.tenant_id,
      createdAt: position.created_at,
      updatedAt: position.updated_at,
    }
  }

  /**
   * 前端CreatePositionRequest转换为后端请求
   */
  static toPositionCreateRequest(position: CreatePositionRequest): PositionCreateRequest {
    return {
      position_type: position.positionType,
      job_profile_id: position.jobProfileId,
      department_id: position.departmentId,
      manager_position_id: position.managerPositionId,
      status: position.status || PositionStatus.OPEN,
      budgeted_fte: position.budgetedFte || 1.0,
      details: position.details,
    }
  }

  /**
   * 前端UpdatePositionRequest转换为后端请求
   */
  static toPositionUpdateRequest(position: UpdatePositionRequest): Partial<PositionCreateRequest> {
    const request: Partial<PositionCreateRequest> = {}
    
    if (position.jobProfileId) {
      request.job_profile_id = position.jobProfileId
    }
    
    if (position.departmentId) {
      request.department_id = position.departmentId
    }
    
    if (position.managerPositionId) {
      request.manager_position_id = position.managerPositionId
    }
    
    if (position.status) {
      request.status = position.status
    }
    
    if (position.budgetedFte !== undefined) {
      request.budgeted_fte = position.budgetedFte
    }
    
    if (position.details) {
      request.details = position.details
    }

    return request
  }
}

// HTTP客户端工具
class PositionApiClient {
  private baseUrl: string

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl
  }

  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`
    
    const defaultHeaders = {
      'Content-Type': 'application/json',
      // TODO: 添加认证头
      // 'Authorization': `Bearer ${getAuthToken()}`,
      // 'X-Tenant-ID': getTenantId(),
    }

    const config: RequestInit = {
      ...options,
      headers: {
        ...defaultHeaders,
        ...options.headers,
      },
    }

    try {
      const response = await fetch(url, config)
      
      const contentType = response.headers.get('content-type')
      let responseData: any = {}
      
      if (contentType && contentType.includes('application/json')) {
        responseData = await response.json()
      }
      
      if (!response.ok) {
        const errorMessage = responseData.error?.message || `API请求失败: ${response.status} ${response.statusText}`
        const apiError = new PositionApiError(errorMessage, response.status, responseData.error?.code)
        throw apiError
      }

      return responseData
    } catch (error) {
      if (error instanceof PositionApiError) {
        throw error
      }
      
      // 职位API请求错误 - error handled by throwing specific error
      throw new PositionApiError('网络请求失败，请检查网络连接', 0, 'NETWORK_ERROR')
    }
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' })
  }

  async post<T>(endpoint: string, data: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async put<T>(endpoint: string, data: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' })
  }
}

// 职位API服务类
export class PositionsApi {
  private client: PositionApiClient

  constructor() {
    this.client = new PositionApiClient()
  }

  /**
   * 获取职位列表
   */
  async getPositions(params: {
    limit?: number
    offset?: number
    departmentId?: string
    status?: PositionStatus
    positionType?: PositionType
  } = {}): Promise<PositionListResponse> {
    const queryParams = new URLSearchParams()
    
    if (params.limit) queryParams.set('limit', params.limit.toString())
    if (params.offset) queryParams.set('offset', params.offset.toString())
    if (params.departmentId) queryParams.set('department_id', params.departmentId)
    if (params.status) queryParams.set('status', params.status)
    if (params.positionType) queryParams.set('position_type', params.positionType)

    const endpoint = `${POSITIONS_ENDPOINT}?${queryParams}`
    const response = await this.client.get<{
      data: PositionApiResponse[]
      limit: number
      offset: number
      total: number
    }>(endpoint)
    
    // 转换数据格式
    const positionsData = response.data.map(pos => PositionApiAdapter.toPosition(pos))
    
    return {
      positions: positionsData,
      pagination: {
        page: Math.floor((params.offset || 0) / (params.limit || 50)) + 1,
        pageSize: params.limit || 50,
        total: response.total,
        totalPages: Math.ceil(response.total / (params.limit || 50))
      }
    }
  }

  /**
   * 根据ID获取职位
   */
  async getPosition(id: string): Promise<Position> {
    const position = await this.client.get<PositionApiResponse>(`${POSITIONS_ENDPOINT}/${id}`)
    return PositionApiAdapter.toPosition(position)
  }

  /**
   * 创建职位
   */
  async createPosition(position: CreatePositionRequest): Promise<Position> {
    const positionRequest = PositionApiAdapter.toPositionCreateRequest(position)
    const createdPosition = await this.client.post<PositionApiResponse>(POSITIONS_ENDPOINT, positionRequest)
    return PositionApiAdapter.toPosition(createdPosition)
  }

  /**
   * 更新职位
   */
  async updatePosition(id: string, position: UpdatePositionRequest): Promise<Position> {
    const positionRequest = PositionApiAdapter.toPositionUpdateRequest(position)
    const updatedPosition = await this.client.put<PositionApiResponse>(`${POSITIONS_ENDPOINT}/${id}`, positionRequest)
    return PositionApiAdapter.toPosition(updatedPosition)
  }

  /**
   * 删除职位
   */
  async deletePosition(id: string): Promise<void> {
    await this.client.delete(`${POSITIONS_ENDPOINT}/${id}`)
  }

  /**
   * 获取职位层级树
   */
  async getPositionsTree(departmentId?: string): Promise<PositionTreeResponse> {
    const queryParams = new URLSearchParams()
    if (departmentId) queryParams.set('department_id', departmentId)
    
    const endpoint = `${POSITIONS_ENDPOINT}/tree?${queryParams}`
    return this.client.get<PositionTreeResponse>(endpoint)
  }

  /**
   * 获取职位统计信息
   */
  async getPositionStats(departmentId?: string): Promise<PositionStats> {
    const queryParams = new URLSearchParams()
    if (departmentId) queryParams.set('department_id', departmentId)
    
    const endpoint = `${POSITIONS_ENDPOINT}/stats?${queryParams}`
    return this.client.get<PositionStats>(endpoint)
  }

  /**
   * 批量更新职位
   */
  async batchUpdatePositions(updates: Array<{ id: string; data: UpdatePositionRequest }>): Promise<Position[]> {
    const promises = updates.map(update => 
      this.updatePosition(update.id, update.data)
    )
    return Promise.all(promises)
  }
}

// 导出单例实例
export const positionsApi = new PositionsApi()

// 便捷的钩子函数
export async function fetchPositions(params?: {
  limit?: number
  offset?: number
  departmentId?: string
  status?: PositionStatus
  positionType?: PositionType
}) {
  return positionsApi.getPositions(params)
}

export async function fetchPosition(id: string) {
  return positionsApi.getPosition(id)
}

export async function createPosition(position: CreatePositionRequest) {
  return positionsApi.createPosition(position)
}

export async function updatePosition(id: string, position: UpdatePositionRequest) {
  return positionsApi.updatePosition(id, position)
}

export async function deletePosition(id: string) {
  return positionsApi.deletePosition(id)
}