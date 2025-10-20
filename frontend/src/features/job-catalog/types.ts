import { JobCatalogStatus } from '@/generated/graphql-types'

export interface CatalogStatusMeta {
  label: string
  color: string
  background: string
  border: string
}

const STATUS_META: Record<JobCatalogStatus, CatalogStatusMeta> = {
  [JobCatalogStatus.ACTIVE]: {
    label: '启用',
    color: '#056449',
    background: '#D6F0E4',
    border: '#6CD2A4',
  },
  [JobCatalogStatus.INACTIVE]: {
    label: '停用',
    color: '#C43737',
    background: '#F9DAD8',
    border: '#F2A29C',
  },
}

export const getCatalogStatusMeta = (status: JobCatalogStatus): CatalogStatusMeta => STATUS_META[status]

export const jobCatalogStatusOptions: Array<{ value: JobCatalogStatus; label: string }> = [
  { value: JobCatalogStatus.ACTIVE, label: '启用' },
  { value: JobCatalogStatus.INACTIVE, label: '停用' },
]

export const formatISODate = (value?: string | null): string => {
  if (!value) {
    return '—'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}
