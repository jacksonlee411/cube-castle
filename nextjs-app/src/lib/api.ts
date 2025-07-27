import axios from 'axios'

// API 客户端配置
export const apiClient = axios.create({
  baseURL: process.env.CUBE_CASTLE_API_URL || 'http://localhost:8080',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    // 添加认证头
    const token = typeof window !== 'undefined' 
      ? localStorage.getItem('cube_castle_token') 
      : null
    
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // 添加租户ID头
    const tenantId = typeof window !== 'undefined'
      ? localStorage.getItem('cube_castle_tenant_id')
      : null
    
    if (tenantId) {
      config.headers['X-Tenant-ID'] = tenantId
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      // 清除认证信息并跳转到登录页
      if (typeof window !== 'undefined') {
        localStorage.removeItem('cube_castle_token')
        localStorage.removeItem('cube_castle_tenant_id')
        window.location.href = '/login'
      }
    }
    
    return Promise.reject(error)
  }
)

// SWR fetcher 函数
export const fetcher = async (url: string) => {
  const response = await apiClient.get(url)
  return response.data
}

// 通用 API 错误类型
export interface ApiError {
  code: string
  message: string
  details?: Record<string, any>
}

// API 响应包装器
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: ApiError
  pagination?: {
    page: number
    pageSize: number
    total: number
    totalPages: number
  }
}

// 通用 API 方法
export const api = {
  // 获取数据
  get: async <T>(url: string): Promise<T> => {
    const response = await apiClient.get<T>(url)
    return response.data
  },

  // 创建数据
  post: async <T>(url: string, data?: any): Promise<T> => {
    const response = await apiClient.post<T>(url, data)
    return response.data
  },

  // 更新数据
  put: async <T>(url: string, data?: any): Promise<T> => {
    const response = await apiClient.put<T>(url, data)
    return response.data
  },

  // 删除数据
  delete: async <T>(url: string): Promise<T> => {
    const response = await apiClient.delete<T>(url)
    return response.data
  },

  // 上传文件
  upload: async <T>(url: string, file: File): Promise<T> => {
    const formData = new FormData()
    formData.append('file', file)
    
    const response = await apiClient.post<T>(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },
}