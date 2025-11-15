import type { QueryClient } from '@tanstack/react-query'

type TemporalEntity = 'position' | 'organization'

export interface TemporalDetailLoaderConfig {
  entity: TemporalEntity
  // 统一预热函数（由实体侧提供稳定导出）
  prefetch: (client: QueryClient, code: string) => Promise<void>
  // 统一取消依据（根键），使用 react-query 的 cancelQueries(exact:false)
  cancelQueryRootKey: readonly unknown[]
}

export interface TemporalDetailLoader {
  preheat: (client: QueryClient, code: string) => Promise<void>
  cancel: (client: QueryClient) => Promise<void>
}

/**
 * 240B/241 – createTemporalDetailLoader
 * - 提供统一的 Loader 外壳（薄适配），避免在页面分散调用具体实体的 prefetch/cancel 实现
 * - 不引入第二事实来源：prefetch/cancelKey 由实体 Hook 的稳定导出提供
 */
export const createTemporalDetailLoader = (config: TemporalDetailLoaderConfig): TemporalDetailLoader => {
  const { prefetch, cancelQueryRootKey } = config
  return {
    preheat: async (client: QueryClient, code: string) => {
      await prefetch(client, code)
    },
    cancel: async (client: QueryClient) => {
      await client.cancelQueries({ queryKey: [...cancelQueryRootKey], exact: false })
    },
  }
}

export default createTemporalDetailLoader

