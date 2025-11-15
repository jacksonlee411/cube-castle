import React, { useEffect, type PropsWithChildren } from 'react'
import { obs } from '@/shared/observability/obs'

export type TemporalEntityKind = 'organization' | 'position'

type ShellProps = PropsWithChildren<{
  entity?: TemporalEntityKind
}>

/**
 * TemporalEntityLayout.Shell
 * - 最小合流外壳：不引入额外 DOM（Fragment），不改变现有 testid/布局
 * - 仅在挂载时打性能起始标记，供 E2E/CI 采集；事件发射仍由页面内部负责（避免重复）
 */
const Shell: React.FC<ShellProps> = ({ children }) => {
  useEffect(() => {
    // 仅做性能起始标记；结束与事件发射由页面内控制，避免重复/冲突
    try {
      if (obs.enabled()) {
        obs.markStart('obs:temporal:hydrate')
      }
    } catch {
      // ignore
    }
  }, [])

  return <>{children}</>
}

export const TemporalEntityLayout = {
  Shell,
} as const

export default TemporalEntityLayout

