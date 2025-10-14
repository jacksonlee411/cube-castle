import React from 'react'
import { TOKEN_STORAGE_KEY } from '@/shared/api/auth'

interface ScopeContainer {
  __SCOPES__?: string | string[]
}

const normalizeScopes = (value: string | string[] | undefined): string[] | undefined => {
  if (Array.isArray(value)) {
    return value
      .filter((item): item is string => typeof item === 'string' && item.trim().length > 0)
      .map((item) => item.trim())
  }

  if (typeof value === 'string') {
    return value.split(/\s+/).filter(Boolean)
  }

  return undefined
}

const readInjectedScopes = (container: ScopeContainer | undefined): string[] | undefined => {
  if (!container) {
    return undefined
  }

  return normalizeScopes(container.__SCOPES__)
}

// JWT roles 到 OAuth scopes 的映射规则
function mapRolesToScopes(roles: string[]): string[] {
  const scopes: string[] = []

  for (const role of roles) {
    switch (role) {
      case 'ADMIN':
        // 管理员拥有所有组织相关权限
        scopes.push(
          'org:read',
          'org:write',
          'org:validate',
          'org:read:hierarchy',
          'org:delete'
        )
        break
      case 'USER':
        // 普通用户拥有基本读取权限
        scopes.push('org:read')
        break
      case 'HR_MANAGER':
        // HR管理员拥有组织管理权限
        scopes.push(
          'org:read',
          'org:write',
          'org:validate',
          'org:read:hierarchy'
        )
        break
      case 'READONLY':
        // 只读用户
        scopes.push('org:read', 'org:read:hierarchy')
        break
      default:
        // 未知角色不映射任何权限
        break
    }
  }

  // 去重并返回
  return [...new Set(scopes)]
}

function readScopesFromEnv(): string[] {
  // 优先从 window.__SCOPES__ 读取（测试/开发可注入）
  const injected = typeof window !== 'undefined'
    ? readInjectedScopes(window as ScopeContainer)
    : undefined
  if (injected && injected.length > 0) {
    return injected
  }
  // 兼容测试环境：支持 globalThis.__SCOPES__
  const gInjected = readInjectedScopes(globalThis as ScopeContainer)
  if (gInjected && gInjected.length > 0) {
    return gInjected
  }

  // 从本地存储中的 OAuth token 读取
  const legacyOauthTokenKey = ['cube', 'castle', 'oauth', 'token'].join('_')

  try {
    const raw =
      localStorage.getItem(TOKEN_STORAGE_KEY) ??
      localStorage.getItem(legacyOauthTokenKey)
    if (raw) {
      const token = JSON.parse(raw) as { scope?: string; accessToken?: string }

      // 优先使用 OAuth scope 字段
      if (token?.scope && typeof token.scope === 'string') {
        return token.scope.split(/\s+/).filter(Boolean)
      }

      // 如果没有 scope 字段，尝试从 JWT accessToken 中解析 roles
      if (token?.accessToken && typeof token.accessToken === 'string') {
        try {
          const parts = token.accessToken.split('.')
          if (parts.length === 3) {
            const payload = JSON.parse(atob(parts[1])) as { roles?: string[] }
            if (payload?.roles && Array.isArray(payload.roles)) {
              // 将 JWT roles 映射为 OAuth scopes
              return mapRolesToScopes(payload.roles)
            }
          }
        } catch {
          /* ignore JWT parsing errors */
        }
      }
    }
  } catch {
    /* ignore */
  }
  return []
}

export function useScopes() {
  const [scopes, setScopes] = React.useState<Set<string>>(() => new Set(readScopesFromEnv()))

  React.useEffect(() => {
    // 简单监听：window.__SCOPES__ 变化不易捕获，这里在mount时读取一次即可
    setScopes(new Set(readScopesFromEnv()))
  }, [])

  const has = React.useCallback((s: string) => scopes.has(s), [scopes])
  const requireAll = React.useCallback((...ss: string[]) => ss.every(scopes.has, scopes), [scopes])
  const requireAny = React.useCallback((...ss: string[]) => ss.some(scopes.has, scopes), [scopes])

  return { scopes, has, requireAll, requireAny }
}

export function useOrgPBAC() {
  const { has } = useScopes()
  const canRead = has('org:read')
  const canReadHierarchy = has('org:read:hierarchy') || canRead
  const canValidate = has('org:validate')
  return { canRead, canReadHierarchy, canValidate }
}
