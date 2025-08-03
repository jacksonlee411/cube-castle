/**
 * 防抖工具函数
 * @param func 要防抖的函数
 * @param wait 等待时间（毫秒）
 * @param immediate 是否立即执行
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number,
  immediate?: boolean
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | undefined
  
  return function executedFunction(...args: Parameters<T>) {
    const later = () => {
      timeout = undefined
      if (!immediate) func(...args)
    }
    
    const callNow = immediate && !timeout
    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
    
    if (callNow) func(...args)
  }
}

/**
 * 节流工具函数
 * @param func 要节流的函数
 * @param limit 时间限制（毫秒）
 */
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean
  
  return function executedFunction(...args: Parameters<T>) {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => inThrottle = false, limit)
    }
  }
}

/**
 * 缓存工具类
 */
export class MemoryCache<T> {
  private cache = new Map<string, { data: T; timestamp: number; ttl: number }>()

  set(key: string, data: T, ttl: number = 5 * 60 * 1000): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    })
  }

  get(key: string): T | null {
    const item = this.cache.get(key)
    if (!item) return null

    if (Date.now() - item.timestamp > item.ttl) {
      this.cache.delete(key)
      return null
    }

    return item.data
  }

  has(key: string): boolean {
    return this.get(key) !== null
  }

  clear(): void {
    this.cache.clear()
  }

  size(): number {
    return this.cache.size
  }

  delete(key: string): void {
    this.cache.delete(key)
  }

  // 清理过期项
  cleanup(): void {
    const now = Date.now()
    for (const [key, item] of this.cache.entries()) {
      if (now - item.timestamp > item.ttl) {
        this.cache.delete(key)
      }
    }
  }
}

/**
 * 请求去重工具
 */
export class RequestDeduplicator {
  private pendingRequests = new Map<string, Promise<any>>()

  async dedupe<T>(key: string, requestFn: () => Promise<T>): Promise<T> {
    // 如果已有相同的请求在进行中，直接返回该Promise
    if (this.pendingRequests.has(key)) {
      return this.pendingRequests.get(key)!
    }

    // 创建新的请求
    const request = requestFn().finally(() => {
      this.pendingRequests.delete(key)
    })

    this.pendingRequests.set(key, request)
    return request
  }

  clear(): void {
    this.pendingRequests.clear()
  }

  size(): number {
    return this.pendingRequests.size
  }
}