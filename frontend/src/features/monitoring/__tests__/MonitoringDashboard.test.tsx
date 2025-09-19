import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import monitoringAPI, {
  type MonitoringHealthData,
  type MonitoringMetrics,
  type AlertList,
  type RateLimitStats,
} from '../../../shared/api/monitoring'
import { MonitoringDashboard } from '../MonitoringDashboard'

describe('MonitoringDashboard', () => {
afterEach(() => {
  vi.restoreAllMocks()
})

  it('renders monitoring stats after successful load', async () => {
    const health: MonitoringHealthData = {
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
    const metrics: MonitoringMetrics = {
      totalOrganizations: 13,
      currentRecords: 45,
      futureRecords: 6,
      historicalRecords: 18,
      duplicateCurrentCount: 1,
      missingCurrentCount: 0,
      timelineOverlapCount: 0,
      inconsistentFlagCount: 0,
      orphanRecordCount: 0,
      healthScore: 96.3,
      alertLevel: 'HEALTHY',
      lastCheckTime: '2025-09-17T02:11:00Z',
    }
    const alerts: AlertList = {
      alertCount: 0,
      alerts: [],
    }
    const rateLimit: RateLimitStats = {
      totalRequests: 100,
      blockedRequests: 5,
      activeClients: 12,
      lastReset: '2025-09-17T00:00:00Z',
      blockRate: '5%',
    }

    const healthSpy = vi.spyOn(monitoringAPI, 'getHealth').mockResolvedValue(health)
    vi.spyOn(monitoringAPI, 'getMetrics').mockResolvedValue(metrics)
    vi.spyOn(monitoringAPI, 'getAlerts').mockResolvedValue(alerts)
    vi.spyOn(monitoringAPI, 'getRateLimitStats').mockResolvedValue(rateLimit)

    render(<MonitoringDashboard />)

    await waitFor(() => expect(healthSpy).toHaveBeenCalled())

    expect(await screen.findByText('系统监控总览')).toBeInTheDocument()
    expect(await screen.findByText('96.3')).toBeInTheDocument()
    expect(screen.getByText('总组织数')).toBeInTheDocument()
    expect(screen.getByText('13')).toBeInTheDocument()
    expect(screen.getByText('问题总数')).toBeInTheDocument()
    expect(screen.getByText('暂无告警')).toBeInTheDocument()
  })

  it('shows an error card when API fails', async () => {
    const healthSpy = vi.spyOn(monitoringAPI, 'getHealth').mockRejectedValue(new Error('加载失败'))
    vi.spyOn(monitoringAPI, 'getMetrics').mockResolvedValue({} as MonitoringMetrics)
    vi.spyOn(monitoringAPI, 'getAlerts').mockResolvedValue({ alertCount: 0, alerts: [] })
    vi.spyOn(monitoringAPI, 'getRateLimitStats').mockResolvedValue({
      totalRequests: 0,
      blockedRequests: 0,
      activeClients: 0,
      lastReset: '',
      blockRate: '0%'
    } as RateLimitStats)

    render(<MonitoringDashboard />)

    await waitFor(() => expect(healthSpy).toHaveBeenCalled())
    expect(await screen.findByText(/加载失败/)).toBeInTheDocument()
  })
})
