import React from 'react'

function readScopesFromEnv(): string[] {
  // 优先从 window.__SCOPES__ 读取（测试/开发可注入）
  if (typeof window !== 'undefined' && (window as any).__SCOPES__) {
    const v = (window as any).__SCOPES__
    if (Array.isArray(v)) return v.filter(Boolean)
    if (typeof v === 'string') return v.split(/\s+/).filter(Boolean)
  }
  // 兼容测试环境：支持 globalThis.__SCOPES__
  if ((globalThis as any).__SCOPES__) {
    const v = (globalThis as any).__SCOPES__
    if (Array.isArray(v)) return v.filter(Boolean)
    if (typeof v === 'string') return v.split(/\s+/).filter(Boolean)
  }
  // 其次从本地存储中的 OAuth token 读取（开发态）
  try {
    const raw = localStorage.getItem('cube_castle_oauth_token')
    if (raw) {
      const token = JSON.parse(raw) as { scope?: string }
      if (token?.scope && typeof token.scope === 'string') {
        return token.scope.split(/\s+/).filter(Boolean)
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
