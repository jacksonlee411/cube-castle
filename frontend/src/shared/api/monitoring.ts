import { unifiedRESTClient } from './unified-client'

// 监控系统接口定义
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

type EnvelopeError = { message?: string }

type Envelope<T> = {
  success: boolean
  data?: T
  error?: EnvelopeError
}

type MaybeEnvelope<T> = T | Envelope<T>

function unwrapData<T>(res: MaybeEnvelope<T>): T {
  if (typeof res === 'object' && res !== null && 'success' in res) {
    const payload = res as Envelope<T>
    if (!payload.success) {
      throw new Error(payload.error?.message || '监控接口返回失败')
    }
    if (!payload.data) {
      throw new Error('监控接口缺少数据返回')
    }
    return payload.data
  }
  return res as T
}

// 监控API客户端
export const monitoringAPI = {
  async getHealth(): Promise<MonitoringHealthData> {
    const res = await unifiedRESTClient.request<{ success: boolean; data: MonitoringHealthData }>(
      `/operational/health`,
      { method: 'GET' }
    )
    return unwrapData<MonitoringHealthData>(res)
  },

  async getMetrics(): Promise<MonitoringMetrics> {
    const res = await unifiedRESTClient.request<MaybeEnvelope<MonitoringMetrics>>(`/operational/metrics`, { method: 'GET' })
    return unwrapData<MonitoringMetrics>(res)
  },

  async getAlerts(): Promise<AlertList> {
    const res = await unifiedRESTClient.request<MaybeEnvelope<AlertList>>(`/operational/alerts`, { method: 'GET' })
    return unwrapData<AlertList>(res)
  },

  async getRateLimitStats(): Promise<RateLimitStats> {
    const res = await unifiedRESTClient.request<MaybeEnvelope<RateLimitStats>>(`/operational/rate-limit/stats`, { method: 'GET' })
    return unwrapData<RateLimitStats>(res)
  },
}

export default monitoringAPI
