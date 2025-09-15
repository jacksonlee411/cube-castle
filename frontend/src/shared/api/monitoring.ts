import { unifiedRESTClient } from './unified-client'

export interface MonitoringSummary {
  totalOrganizations: number
  currentRecords: number
  futureRecords: number
  historicalRecords: number
}

export interface MonitoringIssues {
  duplicateCurrentCount: number
  missingCurrentCount: number
  timelineOverlapCount: number
  inconsistentFlagCount: number
  orphanRecordCount: number
}

export interface MonitoringHealthData {
  status: 'HEALTHY' | 'WARNING' | 'CRITICAL'
  healthScore: number
  summary: MonitoringSummary
  issues: MonitoringIssues
  lastCheckTime: string
}

export interface MonitoringMetrics extends MonitoringSummary, MonitoringIssues {
  healthScore: number
  alertLevel: 'HEALTHY' | 'WARNING' | 'CRITICAL'
  lastCheckTime: string
}

export interface AlertList {
  alertCount: number
  alerts: string[]
}

export interface RateLimitStats {
  totalRequests: number
  blockedRequests: number
  activeClients: number
  lastReset: string
  blockRate: string
}

export const monitoringAPI = {
  async getHealth(): Promise<MonitoringHealthData> {
    const res = await unifiedRESTClient.request<{ success: boolean; data: MonitoringHealthData }>(
      `/operational/health`,
      { method: 'GET' }
    )
    // unifiedRESTClient unwraps enterprise envelope to data already when success=true
    // but return type safety keeps the shape aligned
    // Here res is the unwrapped data object
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return (res as any) as MonitoringHealthData
  },

  async getMetrics(): Promise<MonitoringMetrics> {
    const res = await unifiedRESTClient.request(`/operational/metrics`, { method: 'GET' })
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return (res as any) as MonitoringMetrics
  },

  async getAlerts(): Promise<AlertList> {
    const res = await unifiedRESTClient.request(`/operational/alerts`, { method: 'GET' })
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return (res as any) as AlertList
  },

  async getRateLimitStats(): Promise<RateLimitStats> {
    const res = await unifiedRESTClient.request(`/operational/rate-limit/stats`, { method: 'GET' })
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return (res as any) as RateLimitStats
  },
}

export default monitoringAPI

