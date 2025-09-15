import React from 'react'
import { useScopes } from '../hooks/useScopes'

export interface RequireScopesProps {
  allOf?: string[]
  anyOf?: string[]
  fallback?: React.ReactNode
  children: React.ReactNode
}

/**
 * 基于 scopes 的 UI 门控组件
 * - allOf：需要全部满足
 * - anyOf：满足其一即可
 * 若均未提供，默认直接渲染 children
 */
export const RequireScopes: React.FC<RequireScopesProps> = ({ allOf, anyOf, fallback = null, children }) => {
  const { requireAll, requireAny } = useScopes()

  const passAll = allOf && allOf.length > 0 ? requireAll(...allOf) : true
  const passAny = anyOf && anyOf.length > 0 ? requireAny(...anyOf) : true
  const allowed = passAll && passAny

  if (!allowed) return <>{fallback}</>
  return <>{children}</>
}

export default RequireScopes

