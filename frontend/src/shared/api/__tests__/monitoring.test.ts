import { afterEach, describe, expect, it, vi } from 'vitest'
import { monitoringAPI } from '../monitoring'
import { unifiedRESTClient } from '../unified-client'
import type {
  AlertList,
  MonitoringHealthData,
  MonitoringMetrics,
  RateLimitStats,
} from '../monitoring'

describe('monitoringAPI', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('requests health data from /operational/health', async () => {
    const mockHealth: MonitoringHealthData = {
      status: 'HEALTHY',
      healthScore: 96.3,
      issues: {
        duplicateCurrentCount: 0,
        missingCurrentCount: 0,
        timelineOverlapCount: 0,
        inconsistentFlagCount: 0,
        orphanRecordCount: 0,
      },
      summary: {
        totalOrganizations: 13,
        currentRecords: 45,
        futureRecords: 5,
        historicalRecords: 18,
      },
      lastCheckTime: '2025-09-17T02:11:00Z',
    }
    const spy = vi
      .spyOn(unifiedRESTClient, 'request')
      .mockResolvedValue({ success: true, data: mockHealth })

    const result = await monitoringAPI.getHealth()

    expect(spy).toHaveBeenCalledWith('/operational/health', { method: 'GET' })
    expect(result).toEqual(mockHealth)
  })

  it('requests metrics data from /operational/metrics', async () => {
    const mockMetrics: MonitoringMetrics = {
      totalOrganizations: 13,
      currentRecords: 45,
      futureRecords: 6,
      historicalRecords: 18,
      duplicateCurrentCount: 0,
      missingCurrentCount: 0,
      timelineOverlapCount: 0,
      inconsistentFlagCount: 0,
      orphanRecordCount: 0,
      healthScore: 96.3,
      alertLevel: 'HEALTHY',
      lastCheckTime: '2025-09-17T02:11:00Z',
    }
    const spy = vi
      .spyOn(unifiedRESTClient, 'request')
      .mockResolvedValue(mockMetrics as unknown as MonitoringMetrics)

    const result = await monitoringAPI.getMetrics()

    expect(spy).toHaveBeenCalledWith('/operational/metrics', { method: 'GET' })
    expect(result).toEqual(mockMetrics)
  })

  it('requests alerts data from /operational/alerts', async () => {
    const mockAlerts: AlertList = {
      alertCount: 0,
      alerts: [],
    }
    const spy = vi
      .spyOn(unifiedRESTClient, 'request')
      .mockResolvedValue(mockAlerts as unknown as AlertList)

    const result = await monitoringAPI.getAlerts()

    expect(spy).toHaveBeenCalledWith('/operational/alerts', { method: 'GET' })
    expect(result).toEqual(mockAlerts)
  })

  it('requests rate limit stats from /operational/rate-limit/stats', async () => {
    const mockStats: RateLimitStats = {
      totalRequests: 100,
      blockedRequests: 5,
      activeClients: 12,
      lastReset: '2025-09-17T00:00:00Z',
      blockRate: '5%',
    }
    const spy = vi
      .spyOn(unifiedRESTClient, 'request')
      .mockResolvedValue(mockStats as unknown as RateLimitStats)

    const result = await monitoringAPI.getRateLimitStats()

    expect(spy).toHaveBeenCalledWith('/operational/rate-limit/stats', { method: 'GET' })
    expect(result).toEqual(mockStats)
  })
})
