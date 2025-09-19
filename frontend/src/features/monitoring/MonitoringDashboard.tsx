import React, { useEffect, useMemo, useState } from 'react'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Text } from '@workday/canvas-kit-react/text'
import { Card } from '@workday/canvas-kit-react/card'
import { SecondaryButton } from '@workday/canvas-kit-react/button'
import { colors } from '@workday/canvas-kit-react/tokens'
import { SystemIcon } from '@workday/canvas-kit-react/icon'
import { dashboardIcon, exclamationCircleIcon } from '@workday/canvas-system-icons-web'
import monitoringAPI from '../../shared/api/monitoring'
import type { MonitoringHealthData, MonitoringMetrics, AlertList, RateLimitStats } from '../../shared/api/monitoring'

const StatCard: React.FC<{
  title: string
  value: string | number
  color?: string
  subtitle?: string
}> = ({ title, value, color, subtitle }) => (
  <Card padding="l">
    <Text typeLevel="body.small" color={colors.licorice500}>{title}</Text>
    <Text typeLevel="heading.large" color={color || colors.licorice600}>{value}</Text>
    {subtitle && (
      <Text color={colors.licorice500}>{subtitle}</Text>
    )}
  </Card>
)

const Section: React.FC<{ title: string; icon?: React.ReactNode; right?: React.ReactNode; children: React.ReactNode }>
  = ({ title, icon, right, children }) => (
  <Box marginBottom="l">
    <Flex alignItems="center" justifyContent="space-between" marginBottom="s">
      <Flex alignItems="center" gap="s">
        {icon}
        <Text typeLevel="heading.small">{title}</Text>
      </Flex>
      {right}
    </Flex>
    {children}
  </Box>
)

function statusColor(status: string) {
  if (status === 'CRITICAL') return colors.cinnamon600
  if (status === 'WARNING') return colors.cantaloupe600
  return colors.greenApple600
}

export const MonitoringDashboard: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [health, setHealth] = useState<MonitoringHealthData | null>(null)
  const [metrics, setMetrics] = useState<MonitoringMetrics | null>(null)
  const [alerts, setAlerts] = useState<AlertList | null>(null)
  const [rate, setRate] = useState<RateLimitStats | null>(null)
  const [lastUpdated, setLastUpdated] = useState<string>('')
  const [error, setError] = useState<string>('')

  const refresh = async () => {
    setLoading(true)
    setError('')
    try {
      const [h, m, a, r] = await Promise.all([
        monitoringAPI.getHealth(),
        monitoringAPI.getMetrics(),
        monitoringAPI.getAlerts(),
        monitoringAPI.getRateLimitStats(),
      ])
      setHealth(h)
      setMetrics(m)
      setAlerts(a)
      setRate(r)
      setLastUpdated(new Date().toLocaleString('zh-CN'))
    } catch (e) {
      setError((e as Error).message || '加载监控数据失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
    const timer = setInterval(refresh, 30000)
    return () => clearInterval(timer)
  }, [])

  const issueTotal = useMemo(() => {
    if (!metrics) return 0
    return (
      (metrics.duplicateCurrentCount || 0) +
      (metrics.missingCurrentCount || 0) +
      (metrics.timelineOverlapCount || 0) +
      (metrics.inconsistentFlagCount || 0) +
      (metrics.orphanRecordCount || 0)
    )
  }, [metrics])

  return (
    <Box>
      <Flex alignItems="center" marginBottom="l" gap="s">
        <SystemIcon icon={dashboardIcon} size={24} />
        <Text typeLevel="heading.large">系统监控总览</Text>
        <Box flex={1} />
        <SecondaryButton onClick={refresh} disabled={loading} size="small">
          {loading ? '刷新中...' : '刷新'}
        </SecondaryButton>
      </Flex>

      {error && (
        <Card padding="m">
          <Flex alignItems="center" gap="s" color={colors.cinnamon600}>
            <SystemIcon icon={exclamationCircleIcon} size={16} color={colors.cinnamon600} />
            <Text>{error}</Text>
          </Flex>
        </Card>
      )}

      <Text color={colors.licorice500} marginBottom="m">最后更新: {lastUpdated || '-'}</Text>

      <Section title="健康概览" icon={<SystemIcon icon={dashboardIcon} size={20} />}>
        <Box style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16 }}>
          <StatCard title="健康分" value={health?.healthScore?.toFixed?.(1) ?? '-'} color={statusColor(health?.status || 'HEALTHY')} subtitle={health?.status ?? '-'} />
          <StatCard title="总组织数" value={health?.summary?.totalOrganizations ?? '-'} />
          <StatCard title="当前记录" value={health?.summary?.currentRecords ?? '-'} />
          <StatCard title="未来记录" value={health?.summary?.futureRecords ?? '-'} />
          <StatCard title="历史记录" value={health?.summary?.historicalRecords ?? '-'} />
        </Box>
      </Section>

      <Section title="一致性问题">
        <Box style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16 }}>
          <StatCard title="重复当前记录" value={metrics?.duplicateCurrentCount ?? '-'} color={metrics && metrics.duplicateCurrentCount > 0 ? colors.cinnamon600 : colors.greenApple600} />
          <StatCard title="缺失当前记录" value={metrics?.missingCurrentCount ?? '-'} color={metrics && metrics.missingCurrentCount > 0 ? colors.cinnamon600 : colors.greenApple600} />
          <StatCard title="时间线重叠" value={metrics?.timelineOverlapCount ?? '-'} color={metrics && metrics.timelineOverlapCount > 0 ? colors.cantaloupe600 : colors.greenApple600} />
          <StatCard title="标志不一致" value={metrics?.inconsistentFlagCount ?? '-'} color={metrics && metrics.inconsistentFlagCount > 0 ? colors.cantaloupe600 : colors.greenApple600} />
          <StatCard title="孤立记录" value={metrics?.orphanRecordCount ?? '-'} color={metrics && metrics.orphanRecordCount > 0 ? colors.cantaloupe600 : colors.greenApple600} />
          <StatCard title="问题总数" value={issueTotal} />
        </Box>
      </Section>

      <Section title="告警">
        <Card padding="m">
          <Text color={colors.licorice500} marginBottom="s">当前告警数：{alerts?.alertCount ?? 0}</Text>
          {alerts && alerts.alerts && alerts.alerts.length > 0 ? (
            <Box as="ul" paddingLeft="m">
              {alerts.alerts.map((a, i) => (
                <Text as="li" key={i}>{a}</Text>
              ))}
            </Box>
          ) : (
            <Text color={colors.greenApple600}>暂无告警</Text>
          )}
        </Card>
      </Section>

      <Section title="限流统计">
        <Box style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16 }}>
          <StatCard title="总请求" value={rate?.totalRequests ?? '-'} />
          <StatCard title="被拦截" value={rate?.blockedRequests ?? '-'} />
          <StatCard title="活跃客户端" value={rate?.activeClients ?? '-'} />
          <StatCard title="拦截率" value={rate?.blockRate ?? '-'} />
        </Box>
      </Section>
    </Box>
  )
}

export default MonitoringDashboard

