import React from 'react'
import { getCatalogStatusMeta } from '../types'
import type { JobCatalogStatus } from '@/generated/graphql-types'

export const StatusBadge: React.FC<{ status: JobCatalogStatus }> = ({ status }) => {
  const meta = getCatalogStatusMeta(status)
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '4px 8px',
        borderRadius: 12,
        fontSize: 12,
        fontWeight: 600,
        minWidth: 54,
        color: meta.color,
        backgroundColor: meta.background,
        border: `1px solid ${meta.border}`,
      }}
    >
      {meta.label}
    </span>
  )
}
