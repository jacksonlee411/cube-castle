export type PositionLifecycleType =
  | 'CREATE'
  | 'FILL'
  | 'VACATE'
  | 'TRANSFER'
  | 'SUSPEND'
  | 'REACTIVATE'

export interface PositionLifecycleEvent {
  id: string
  type: PositionLifecycleType
  label: string
  operator: string
  occurredAt: string
  summary: string
}
export type {
  PositionStatus,
  PositionRecord,
  PositionTimelineEvent,
  PositionsQueryResult,
  PositionDetailResult,
} from '@/shared/types/positions'
