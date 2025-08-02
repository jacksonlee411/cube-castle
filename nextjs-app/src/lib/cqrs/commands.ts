import axios, { AxiosInstance } from 'axios'
import { CreateOrganizationRequest, UpdateOrganizationRequest } from '@/types'

// CQRS 命令客户端 - 专门处理写操作
class OrganizationCommandService {
  private client: AxiosInstance

  constructor(baseURL: string = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080') {
    this.client = axios.create({
      baseURL: `${baseURL}/api/v1/commands`,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // 请求拦截器 - 添加认证和租户信息
    this.client.interceptors.request.use((config) => {
      const token = localStorage.getItem('auth_token')
      const tenantId = localStorage.getItem('tenant_id') || '00000000-0000-0000-0000-000000000001'
      
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      config.headers['X-Tenant-ID'] = tenantId
      
      return config
    })

    // 响应拦截器 - 错误处理
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        console.error('Command API Error:', error)
        throw error
      }
    )
  }

  /**
   * 创建组织单元命令
   */
  async createOrganizationUnit(data: CreateOrganizationRequest): Promise<{
    unit_id: string
    status: string
    message: string
  }> {
    const tenantId = localStorage.getItem('tenant_id') || '00000000-0000-0000-0000-000000000001'
    
    const payload = {
      tenant_id: tenantId,
      unit_type: data.unit_type,
      name: data.name,
      description: data.description,
      parent_unit_id: data.parent_unit_id,
      profile: data.profile || {}
    }

    console.log('🎯 CQRS命令: 创建组织单元', payload)
    
    const response = await this.client.post('/create-organization-unit', payload)
    
    console.log('✅ 组织单元创建命令成功:', response.data)
    return response.data
  }

  /**
   * 更新组织单元命令
   */
  async updateOrganizationUnit(id: string, data: UpdateOrganizationRequest): Promise<{
    status: string
    message: string
  }> {
    const tenantId = localStorage.getItem('tenant_id') || '00000000-0000-0000-0000-000000000001'
    
    const payload = {
      id,
      tenant_id: tenantId,
      ...data
    }

    console.log('📝 CQRS命令: 更新组织单元', payload)
    
    const response = await this.client.put('/update-organization-unit', payload)
    
    console.log('✅ 组织单元更新命令成功:', response.data)
    return response.data
  }

  /**
   * 删除组织单元命令
   */
  async deleteOrganizationUnit(id: string): Promise<{
    status: string
    message: string
  }> {
    const tenantId = localStorage.getItem('tenant_id') || '00000000-0000-0000-0000-000000000001'
    
    const payload = {
      id,
      tenant_id: tenantId
    }

    console.log('🗑️ CQRS命令: 删除组织单元', payload)
    
    const response = await this.client.delete('/delete-organization-unit', { data: payload })
    
    console.log('✅ 组织单元删除命令成功:', response.data)
    return response.data
  }

  /**
   * 批量更新组织单元命令
   */
  async bulkUpdateOrganizationUnits(updates: Array<{
    id: string
    data: UpdateOrganizationRequest
  }>): Promise<{
    success_count: number
    failed_count: number
    errors: Array<{ id: string, error: string }>
  }> {
    const tenantId = localStorage.getItem('tenant_id') || '00000000-0000-0000-0000-000000000001'
    
    const payload = {
      tenant_id: tenantId,
      updates
    }

    console.log('🔄 CQRS命令: 批量更新组织单元', payload)
    
    const response = await this.client.patch('/bulk-update-organization-units', payload)
    
    console.log('✅ 批量更新命令成功:', response.data)
    return response.data
  }

  /**
   * 重组组织架构命令（移动组织单元）
   */
  async restructureOrganization(moves: Array<{
    unit_id: string
    new_parent_id: string | null
    new_position?: number
  }>): Promise<{
    status: string
    message: string
    affected_units: string[]
  }> {
    const tenantId = localStorage.getItem('tenant_id') || '00000000-0000-0000-0000-000000000001'
    
    const payload = {
      tenant_id: tenantId,
      moves
    }

    console.log('🔄 CQRS命令: 重组组织架构', payload)
    
    const response = await this.client.post('/restructure-organization', payload)
    
    console.log('✅ 组织重组命令成功:', response.data)
    return response.data
  }
}

// 导出单例实例
export const organizationCommands = new OrganizationCommandService()

// 默认导出
export default organizationCommands