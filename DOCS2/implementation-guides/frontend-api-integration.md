# 前端API集成指南

**版本**: v2.0  
**创建日期**: 2025-08-04  
**更新日期**: 2025-08-06  
**适用范围**: Vite + React + Canvas Kit前端应用  
**目标读者**: 前端开发者

## 📋 概述

本指南提供了在Cube Castle Vite + React + Canvas Kit现代化前端应用中正确集成职位管理API的完整指导，确保开发者使用统一、高效的API调用方式。

## 🎯 核心原则

### 1. 统一路由使用
- ✅ **正确路由**: `/api/v1/positions`
- ❌ **错误路由**: `/api/v1/corehr/positions`
- ❌ **过时路由**: `/api/v1/organization/positions`

### 2. 类型安全
- 使用TypeScript接口定义
- 严格的类型检查
- 运行时类型验证

### 3. 错误处理
- 统一的错误处理机制
- 用户友好的错误提示
- 错误重试策略

### 4. 性能优化
- 请求缓存策略
- 分页和懒加载
- 防抖和节流

## 🔧 API客户端架构

### 文件结构
```
src/lib/api/
├── positions.ts          # 职位API客户端
├── types/                # TypeScript类型定义
│   ├── position.ts      # 职位相关类型
│   └── common.ts        # 通用类型
├── utils/               # 工具函数
│   ├── request.ts       # HTTP请求工具
│   └── error.ts         # 错误处理工具
└── hooks/               # React Hooks
    └── usePositions.ts  # 职位数据hooks
```

### 核心配置
```typescript
// src/lib/api/config.ts
export const API_CONFIG = {
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
  endpoints: {
    positions: '/api/v1/positions',  // ✅ 标准路由
    employees: '/api/v1/corehr/employees',
    organizations: '/api/v1/corehr/organizations'
  },
  timeout: 10000,
  retryAttempts: 3
}
```

## 📊 类型定义

### 核心类型
```typescript
// src/types/position.ts
export interface Position {
  id: string
  tenantId: string
  positionType: PositionType
  jobProfileId: string
  departmentId: string
  managerPositionId?: string
  status: PositionStatus
  budgetedFte: number
  details?: Record<string, any>
  createdAt: string
  updatedAt: string
}

export enum PositionType {
  REGULAR = 'REGULAR',
  CONTRACT = 'CONTRACT',
  INTERN = 'INTERN',
  CONSULTANT = 'CONSULTANT'
}

export enum PositionStatus {
  ACTIVE = 'ACTIVE',
  INACTIVE = 'INACTIVE',
  OPEN = 'OPEN',
  CLOSED = 'CLOSED'
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

export interface PositionListResponse {
  positions: Position[]
  pagination: {
    page: number
    pageSize: number
    total: number
    totalPages: number
  }
}
```

## 🔌 API客户端使用

### 基础使用
```typescript
import { positionsApi } from '@/lib/api/positions'

// 获取职位列表
const getPositions = async () => {
  try {
    const response = await positionsApi.getPositions({
      limit: 20,
      offset: 0,
      departmentId: 'dept-123',
      status: PositionStatus.OPEN
    })
    return response.positions
  } catch (error) {
    console.error('获取职位列表失败:', error)
    throw error
  }
}

// 创建职位
const createPosition = async (positionData: CreatePositionRequest) => {
  try {
    const position = await positionsApi.createPosition(positionData)
    console.log('职位创建成功:', position)
    return position
  } catch (error) {
    console.error('创建职位失败:', error)
    throw error
  }
}
```

### React Hooks集成
```typescript
// src/hooks/usePositions.ts
import { useState, useEffect } from 'react'
import { positionsApi } from '@/lib/api/positions'
import type { Position, PositionListResponse } from '@/types/position'

export function usePositions(params?: {
  departmentId?: string
  status?: PositionStatus
  limit?: number
}) {
  const [positions, setPositions] = useState<Position[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchPositions = async () => {
      try {
        setLoading(true)
        setError(null)
        const response = await positionsApi.getPositions(params)
        setPositions(response.positions)
      } catch (err) {
        setError(err instanceof Error ? err.message : '获取职位数据失败')
      } finally {
        setLoading(false)
      }
    }

    fetchPositions()
  }, [params?.departmentId, params?.status, params?.limit])

  return { positions, loading, error }
}

// 组件中使用
function PositionsList() {
  const { positions, loading, error } = usePositions({
    status: PositionStatus.OPEN,
    limit: 50
  })

  if (loading) return <div>加载中...</div>
  if (error) return <div>错误: {error}</div>

  return (
    <div>
      {positions.map(position => (
        <div key={position.id}>
          {/* 职位卡片内容 */}
        </div>
      ))}
    </div>
  )
}
```

## 🚨 错误处理

### 统一错误处理
```typescript
// src/lib/api/error.ts
export class ApiError extends Error {
  constructor(
    message: string,
    public statusCode: number,
    public code?: string
  ) {
    super(message)
    this.name = 'ApiError'
  }

  static fromResponse(response: Response, data?: any): ApiError {
    const message = data?.error?.message || `API请求失败: ${response.status}`
    return new ApiError(message, response.status, data?.error?.code)
  }
}

export function handleApiError(error: unknown): string {
  if (error instanceof ApiError) {
    switch (error.statusCode) {
      case 400:
        return '请求参数有误，请检查输入信息'
      case 401:
        return '登录已过期，请重新登录'
      case 403:
        return '没有权限执行此操作'
      case 404:
        return '请求的资源不存在'
      case 500:
        return '服务器内部错误，请稍后重试'
      default:
        return error.message
    }
  }
  
  return '网络连接失败，请检查网络设置'
}
```

### 组件级错误处理
```typescript
import { toast } from '@/components/ui/use-toast'
import { handleApiError } from '@/lib/api/error'

function CreatePositionForm() {
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (data: CreatePositionRequest) => {
    try {
      setLoading(true)
      await positionsApi.createPosition(data)
      toast({
        title: '成功',
        description: '职位创建成功'
      })
    } catch (error) {
      toast({
        title: '创建失败',
        description: handleApiError(error),
        variant: 'destructive'
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      {/* 表单内容 */}
    </form>
  )
}
```

## ⚡ 性能优化

### 请求缓存
```typescript
// src/lib/api/cache.ts
const cache = new Map<string, { data: any; timestamp: number }>()
const CACHE_DURATION = 5 * 60 * 1000 // 5分钟

export function getCachedData<T>(key: string): T | null {
  const cached = cache.get(key)
  if (cached && Date.now() - cached.timestamp < CACHE_DURATION) {
    return cached.data
  }
  return null
}

export function setCachedData<T>(key: string, data: T): void {
  cache.set(key, { data, timestamp: Date.now() })
}

// 在API客户端中使用
export class PositionsApi {
  async getPositions(params: GetPositionsParams): Promise<PositionListResponse> {
    const cacheKey = `positions-${JSON.stringify(params)}`
    
    // 尝试从缓存获取
    const cachedData = getCachedData<PositionListResponse>(cacheKey)
    if (cachedData) {
      return cachedData
    }

    // 发起请求
    const response = await this.client.get<ApiResponse>(endpoint)
    const result = this.transformResponse(response)
    
    // 缓存结果
    setCachedData(cacheKey, result)
    
    return result
  }
}
```

### 分页和虚拟滚动
```typescript
// src/hooks/usePaginatedPositions.ts
export function usePaginatedPositions(pageSize = 20) {
  const [positions, setPositions] = useState<Position[]>([])
  const [hasMore, setHasMore] = useState(true)
  const [loading, setLoading] = useState(false)

  const loadMore = useCallback(async () => {
    if (loading || !hasMore) return

    try {
      setLoading(true)
      const response = await positionsApi.getPositions({
        limit: pageSize,
        offset: positions.length
      })
      
      setPositions(prev => [...prev, ...response.positions])
      setHasMore(response.positions.length === pageSize)
    } catch (error) {
      console.error('加载更多职位失败:', error)
    } finally {
      setLoading(false)
    }
  }, [positions.length, pageSize, loading, hasMore])

  return { positions, loadMore, hasMore, loading }
}
```

## 🧪 测试策略

### API客户端测试
```typescript
// src/lib/api/__tests__/positions.test.ts
import { positionsApi } from '../positions'
import { PositionType, PositionStatus } from '@/types/position'

// Mock fetch
global.fetch = jest.fn()

describe('PositionsApi', () => {
  beforeEach(() => {
    (fetch as jest.Mock).mockClear()
  })

  it('应该成功获取职位列表', async () => {
    const mockResponse = {
      data: [
        {
          id: '123',
          position_type: 'REGULAR',
          job_profile_id: '456',
          department_id: '789',
          status: 'OPEN',
          budgeted_fte: 1,
          created_at: '2025-08-04T00:00:00Z',
          updated_at: '2025-08-04T00:00:00Z'
        }
      ],
      total: 1,
      limit: 50,
      offset: 0
    };

    (fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse
    })

    const result = await positionsApi.getPositions()
    
    expect(result.positions).toHaveLength(1)
    expect(result.positions[0].positionType).toBe(PositionType.REGULAR)
    expect(fetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/positions?',
      expect.any(Object)
    )
  })

  it('应该正确处理API错误', async () => {
    (fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: async () => ({ error: { message: '资源不存在' } })
    })

    await expect(positionsApi.getPositions()).rejects.toThrow('资源不存在')
  })
})
```

### 组件测试
```typescript
// src/components/__tests__/PositionsList.test.tsx
import { render, screen, waitFor } from '@testing-library/react'
import { PositionsList } from '../PositionsList'
import { positionsApi } from '@/lib/api/positions'

jest.mock('@/lib/api/positions')

describe('PositionsList', () => {
  it('应该显示加载状态', () => {
    render(<PositionsList />)
    expect(screen.getByText('加载中...')).toBeInTheDocument()
  })

  it('应该显示职位列表', async () => {
    const mockPositions = [
      { id: '1', positionType: 'REGULAR', /* ... */ },
      { id: '2', positionType: 'CONTRACT', /* ... */ }
    ];

    (positionsApi.getPositions as jest.Mock).mockResolvedValue({
      positions: mockPositions,
      pagination: { total: 2 }
    })

    render(<PositionsList />)
    
    await waitFor(() => {
      expect(screen.getByText('职位1')).toBeInTheDocument()
      expect(screen.getByText('职位2')).toBeInTheDocument()
    })
  })
})
```

## 📱 Vite + Canvas Kit集成

### API代理配置
```javascript
// next.config.js
/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: 'http://localhost:8080/api/v1/:path*'
      }
    ]
  }
}

module.exports = nextConfig
```

### 服务端渲染支持
```typescript
// pages/positions/index.tsx
import type { GetServerSideProps } from 'next'
import { positionsApi } from '@/lib/api/positions'

export const getServerSideProps: GetServerSideProps = async () => {
  try {
    const positions = await positionsApi.getPositions({ limit: 20 })
    return {
      props: {
        initialPositions: positions.positions
      }
    }
  } catch (error) {
    console.error('SSR获取职位数据失败:', error)
    return {
      props: {
        initialPositions: [],
        error: '获取数据失败'
      }
    }
  }
}
```

## 🔍 调试和监控

### 开发环境调试
```typescript
// src/lib/api/debug.ts
const DEBUG = process.env.NODE_ENV === 'development'

export function logApiCall(method: string, url: string, data?: any) {
  if (DEBUG) {
    console.group(`🌐 API ${method} ${url}`)
    if (data) console.log('请求数据:', data)
    console.groupEnd()
  }
}

export function logApiResponse(url: string, response: any) {
  if (DEBUG) {
    console.group(`📨 API响应 ${url}`)
    console.log('响应数据:', response)
    console.groupEnd()
  }
}
```

### 错误监控
```typescript
// src/lib/api/monitoring.ts
export function reportApiError(error: ApiError, context: string) {
  // 发送到错误监控服务
  if (process.env.NODE_ENV === 'production') {
    // Sentry, LogRocket 等
    console.error(`API错误 [${context}]:`, error)
  }
}
```

## 📚 最佳实践

### 1. API调用规范
- ✅ 使用TypeScript类型定义
- ✅ 统一的错误处理
- ✅ 合理的缓存策略
- ✅ 适当的加载状态显示

### 2. 性能优化
- ✅ 实施请求缓存
- ✅ 使用分页加载
- ✅ 防抖用户输入
- ✅ 预加载关键数据

### 3. 用户体验
- ✅ 友好的错误提示
- ✅ 及时的加载反馈
- ✅ 优雅的降级方案
- ✅ 离线状态处理

## 📋 检查清单

### 开发阶段
- [ ] 使用正确的API路由 (`/api/v1/positions`)
- [ ] 实现完整的TypeScript类型
- [ ] 添加适当的错误处理
- [ ] 编写单元测试

### 集成阶段
- [ ] 验证API调用正确性
- [ ] 测试错误场景处理
- [ ] 检查性能表现
- [ ] 确认缓存策略有效

### 生产阶段
- [ ] 监控API调用成功率
- [ ] 跟踪性能指标
- [ ] 收集用户反馈
- [ ] 定期更新依赖

---

**维护者**: 前端开发团队  
**审核者**: 技术负责人  
**最后更新**: 2025-08-04